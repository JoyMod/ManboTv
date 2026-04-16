package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

const (
	searchDefaultView          = "aggregate"
	searchDefaultSort          = "smart"
	searchDefaultSourceMode    = "all"
	searchStatusDone           = "done"
	searchStatusEmpty          = "empty"
	searchStatusPartial        = "partial"
	searchStatusTimeout        = "timeout"
	searchStatusError          = "error"
	searchMaximumPageSize      = 120
	searchMaximumFacetYearSize = 24
	searchExactMatchScore      = 140
	searchContainsMatchScore   = 80
	searchTokenMatchScore      = 16
	searchExactYearScore       = 18
	searchRangeYearScore       = 10
	searchExplicitTypeScore    = 16
	searchQueryTypeScore       = 10
	searchMultiSourceScore     = 12
	searchEpisodeBonusCap      = 24
	searchAdvancedTimeoutSlack = 2 * time.Second
	searchAdvancedTimeoutCap   = 45 * time.Second
)

type SearchParams struct {
	Query               string
	Page                int
	PageSize            int
	View                string
	Sort                string
	Types               []string
	Sources             []string
	YearFrom            int
	YearTo              int
	SourceMode          string
	PreferExact         bool
	MaxPages            int
	MaxConcurrent       int
	MaxResultsPerSource int
	SourceTimeout       time.Duration
	EnableFastReturn    bool
	EnableStream        bool
	ReturnAll           bool
}

type searchQueryProfile struct {
	RawQuery         string
	NormalizedQuery  string
	TokenSet         []string
	DetectedYear     int
	DetectedType     string
	CandidateQueries []string
}

type searchSourceResult struct {
	Site      model.ApiSite
	Results   []model.SearchResult
	Status    string
	PageCount int
	ElapsedMs int64
	Err       error
}

type searchAggregateState struct {
	Key     string
	Title   string
	Year    string
	Type    string
	Cover   string
	Tags    []string
	Items   []model.SearchResult
	Sources map[string]string
}

type searchRuntimeOptions struct {
	Page                int
	PageSize            int
	View                string
	Sort                string
	Types               []string
	Sources             []string
	YearFrom            int
	YearTo              int
	SourceMode          string
	PreferExact         bool
	MaxPages            int
	MaxConcurrent       int
	MaxResultsPerSource int
	SourceTimeout       time.Duration
	EnableFastReturn    bool
	EnableStream        bool
	ReturnAll           bool
}

func (s *searchService) SearchAdvanced(
	ctx context.Context,
	params SearchParams,
	sites []model.ApiSite,
	policy ContentPolicy,
) (*model.SearchEnvelope, error) {
	options := s.resolveSearchOptions(params)
	profile := buildSearchQueryProfile(params.Query)
	filteredSites := filterSitesBySelection(sites, options.Sources)
	pageInfo := buildSearchPageInfo(options.Page, options.PageSize, 0, options.ReturnAll)
	envelope := &model.SearchEnvelope{
		Query:            strings.TrimSpace(params.Query),
		NormalizedQuery:  profile.NormalizedQuery,
		Results:          []model.SearchResult{},
		Aggregates:       []model.SearchAggregateResult{},
		Facets:           model.SearchFacets{},
		SourceStatus:     []model.SearchSourceStatus{},
		LegacySourceMap:  map[string]string{},
		PageInfo:         pageInfo,
		Execution:        model.SearchExecutionInfo{Query: strings.TrimSpace(params.Query), NormalizedQuery: profile.NormalizedQuery},
		SelectedTypes:    options.Types,
		SelectedSources:  options.Sources,
		SelectedSort:     options.Sort,
		SelectedView:     options.View,
		SelectedYearFrom: options.YearFrom,
		SelectedYearTo:   options.YearTo,
		SelectedMode:     options.SourceMode,
	}

	if profile.RawQuery == "" || len(filteredSites) == 0 {
		envelope.Execution.TotalSources = len(filteredSites)
		return envelope, nil
	}

	startedAt := time.Now()
	results, statuses, completedSources, degraded := s.collectAdvancedResults(ctx, profile, options, filteredSites)
	visibleResults := filterVisibleSearchResults(results, policy)
	rankedResults := rankSearchResults(visibleResults, profile, options)
	aggregates := buildSearchAggregates(rankedResults)
	rankedResults, aggregates = applySourceModeFilter(rankedResults, aggregates, options.SourceMode)
	facets := buildSearchFacets(rankedResults, aggregates)
	pageInfo = buildSearchPageInfo(options.Page, options.PageSize, len(rankedResults), options.ReturnAll)

	envelope.Results = paginateSearchResults(rankedResults, options.Page, options.PageSize, options.ReturnAll)
	envelope.Aggregates = aggregates
	envelope.Facets = facets
	envelope.SourceStatus = statuses
	envelope.LegacySourceMap = buildLegacySourceMap(statuses)
	envelope.PageInfo = pageInfo
	envelope.Execution = model.SearchExecutionInfo{
		Query:            profile.RawQuery,
		NormalizedQuery:  profile.NormalizedQuery,
		CompletedSources: completedSources,
		TotalSources:     len(filteredSites),
		ElapsedMs:        time.Since(startedAt).Milliseconds(),
		Degraded:         degraded,
		StreamingEnabled: options.EnableStream,
	}

	return envelope, nil
}

func (s *searchService) resolveSearchOptions(params SearchParams) searchRuntimeOptions {
	page := params.Page
	if page <= 0 {
		page = 1
	}

	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > searchMaximumPageSize {
		pageSize = searchMaximumPageSize
	}

	maxPages := params.MaxPages
	if maxPages <= 0 {
		maxPages = 1
		if s.config != nil && s.config.MaxPages > 0 {
			maxPages = s.config.MaxPages
		}
	}

	maxConcurrent := params.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = s.resolveMaxConcurrent()
	}

	maxResultsPerSource := params.MaxResultsPerSource
	if maxResultsPerSource <= 0 {
		maxResultsPerSource = 50
		if s.config != nil && s.config.MaxResultsPerSource > 0 {
			maxResultsPerSource = s.config.MaxResultsPerSource
		}
	}

	sourceTimeout := params.SourceTimeout
	if sourceTimeout <= 0 {
		sourceTimeout = s.sourceTimeout
	}
	if sourceTimeout <= 0 {
		sourceTimeout = defaultSourceTimeout
	}

	return searchRuntimeOptions{
		Page:                page,
		PageSize:            pageSize,
		View:                firstNonEmptyString(strings.TrimSpace(params.View), searchDefaultView),
		Sort:                firstNonEmptyString(strings.TrimSpace(params.Sort), searchDefaultSort),
		Types:               normalizeStringList(params.Types),
		Sources:             normalizeStringList(params.Sources),
		YearFrom:            params.YearFrom,
		YearTo:              params.YearTo,
		SourceMode:          firstNonEmptyString(strings.TrimSpace(params.SourceMode), searchDefaultSourceMode),
		PreferExact:         params.PreferExact,
		MaxPages:            maxPages,
		MaxConcurrent:       maxConcurrent,
		MaxResultsPerSource: maxResultsPerSource,
		SourceTimeout:       sourceTimeout,
		EnableFastReturn:    params.EnableFastReturn,
		EnableStream:        params.EnableStream,
		ReturnAll:           params.ReturnAll,
	}
}

func (s *searchService) resolveAdvancedSearchTimeout(
	options searchRuntimeOptions,
	siteCount int,
) time.Duration {
	baseTimeout := s.resolveSearchTimeout()
	if siteCount <= 0 {
		return baseTimeout
	}

	sourceTimeout := options.SourceTimeout
	if sourceTimeout <= 0 {
		sourceTimeout = defaultSourceTimeout
	}

	maxConcurrent := options.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}

	queueBatches := (siteCount + maxConcurrent - 1) / maxConcurrent
	pageFactor := options.MaxPages
	if pageFactor <= 0 {
		pageFactor = 1
	}

	estimatedTimeout := time.Duration(queueBatches+pageFactor-1)*sourceTimeout + searchAdvancedTimeoutSlack
	if estimatedTimeout < baseTimeout {
		return baseTimeout
	}
	if estimatedTimeout > searchAdvancedTimeoutCap {
		return searchAdvancedTimeoutCap
	}
	return estimatedTimeout
}

func buildSearchQueryProfile(query string) searchQueryProfile {
	rawQuery := strings.TrimSpace(query)
	normalizedQuery := normalizeComparableText(rawQuery)
	segments := strings.Fields(strings.ToLower(strings.TrimSpace(query)))
	tokenSet := make([]string, 0, len(segments))
	seen := make(map[string]struct{}, len(segments))
	for _, segment := range segments {
		token := normalizeComparableText(segment)
		if token == "" {
			continue
		}
		if _, exists := seen[token]; exists {
			continue
		}
		seen[token] = struct{}{}
		tokenSet = append(tokenSet, token)
	}

	if normalizedQuery != "" {
		if _, exists := seen[normalizedQuery]; !exists {
			tokenSet = append(tokenSet, normalizedQuery)
		}
	}

	year := extractYearHint(rawQuery)
	detectedType := detectSearchType(rawQuery)
	candidateQueries := uniqueStrings([]string{rawQuery, normalizeSearchKeyword(rawQuery)})

	return searchQueryProfile{
		RawQuery:         rawQuery,
		NormalizedQuery:  normalizedQuery,
		TokenSet:         tokenSet,
		DetectedYear:     year,
		DetectedType:     detectedType,
		CandidateQueries: candidateQueries,
	}
}

func (s *searchService) collectAdvancedResults(
	ctx context.Context,
	profile searchQueryProfile,
	options searchRuntimeOptions,
	sites []model.ApiSite,
) ([]model.SearchResult, []model.SearchSourceStatus, int, bool) {
	searchCtx, cancel := context.WithTimeout(ctx, s.resolveAdvancedSearchTimeout(options, len(sites)))
	defer cancel()

	resultCh := make(chan searchSourceResult, len(sites))
	sem := make(chan struct{}, options.MaxConcurrent)
	var waitGroup sync.WaitGroup

	for _, site := range sites {
		site := site
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			select {
			case sem <- struct{}{}:
			case <-searchCtx.Done():
				return
			}
			defer func() { <-sem }()

			result := s.searchSiteAdvanced(searchCtx, site, profile, options)
			select {
			case resultCh <- result:
			case <-searchCtx.Done():
			}
		}()
	}

	go func() {
		waitGroup.Wait()
		close(resultCh)
	}()

	var fastTimer *time.Timer
	var fastTimerCh <-chan time.Time
	if options.EnableFastReturn && s.fastReturnAfter > 0 {
		fastTimer = time.NewTimer(s.fastReturnAfter)
		fastTimerCh = fastTimer.C
		defer fastTimer.Stop()
	}

	results := make([]model.SearchResult, 0)
	statuses := make([]model.SearchSourceStatus, 0, len(sites))
	completed := 0
	degraded := false

collectLoop:
	for {
		select {
		case item, ok := <-resultCh:
			if !ok {
				break collectLoop
			}
			completed++
			statuses = append(statuses, toSearchSourceStatus(item))
			results = append(results, item.Results...)

		case <-fastTimerCh:
			if completed > 0 && completed < len(sites) {
				degraded = true
				cancel()
			}

		case <-searchCtx.Done():
			if completed < len(sites) {
				degraded = true
			}
			break collectLoop
		}
	}

	for len(statuses) < len(sites) {
		select {
		case item, ok := <-resultCh:
			if !ok {
				goto finalize
			}
			statuses = append(statuses, toSearchSourceStatus(item))
			results = append(results, item.Results...)
		default:
			goto finalize
		}
	}

finalize:
	completedSources := len(statuses)
	statuses = fillMissingSearchStatuses(statuses, sites, degraded)
	return dedupeSearchResults(results), sortSearchStatuses(statuses, sites), completedSources, degraded
}

func (s *searchService) searchSiteAdvanced(
	ctx context.Context,
	site model.ApiSite,
	profile searchQueryProfile,
	options searchRuntimeOptions,
) searchSourceResult {
	startedAt := time.Now()
	results := make([]model.SearchResult, 0, options.MaxResultsPerSource)
	pageCount := 0
	lastErr := error(nil)

	for page := 1; page <= options.MaxPages; page++ {
		pageCtx := ctx
		cancel := func() {}
		if options.SourceTimeout > 0 {
			pageCtx, cancel = context.WithTimeout(ctx, options.SourceTimeout)
		}

		pageResults, totalPages, err := s.searchPageAcrossQueries(pageCtx, site, profile.CandidateQueries, page)
		cancel()
		if err != nil {
			lastErr = err
			if len(results) == 0 {
				break
			}
			return searchSourceResult{
				Site:      site,
				Results:   results,
				Status:    searchStatusPartial,
				PageCount: pageCount,
				ElapsedMs: time.Since(startedAt).Milliseconds(),
				Err:       err,
			}
		}

		if len(pageResults) == 0 {
			break
		}

		pageCount++
		results = append(results, pageResults...)
		if len(results) >= options.MaxResultsPerSource {
			results = results[:options.MaxResultsPerSource]
			break
		}
		if totalPages > 0 && page >= totalPages {
			break
		}
	}

	status := searchStatusDone
	switch {
	case len(results) == 0 && lastErr == nil:
		status = searchStatusEmpty
	case len(results) == 0 && isSearchTimeoutError(lastErr):
		status = searchStatusTimeout
	case len(results) == 0 && lastErr != nil:
		status = searchStatusError
	case len(results) > 0 && lastErr != nil:
		status = searchStatusPartial
	}

	return searchSourceResult{
		Site:      site,
		Results:   dedupeSearchResults(results),
		Status:    status,
		PageCount: pageCount,
		ElapsedMs: time.Since(startedAt).Milliseconds(),
		Err:       lastErr,
	}
}

func (s *searchService) searchPageAcrossQueries(
	ctx context.Context,
	site model.ApiSite,
	queries []string,
	page int,
) ([]model.SearchResult, int, error) {
	var lastErr error
	for _, query := range queries {
		searchURL, err := appendProviderQuery(site.API, buildSearchPageQuery(query, page))
		if err != nil {
			lastErr = fmt.Errorf("构建搜索URL失败: %w", err)
			continue
		}

		var (
			results    []model.SearchResult
			totalPages int
		)
		lastErr = s.withRetry(ctx, func() error {
			pageResults, pageCount, err := s.doSearchRequestPage(ctx, searchURL, site)
			if err != nil {
				return err
			}
			results = pageResults
			totalPages = pageCount
			return nil
		})
		if lastErr == nil {
			return results, totalPages, nil
		}
	}

	return nil, 0, lastErr
}

func buildSearchPageQuery(query string, page int) url.Values {
	values := url.Values{
		"ac": {"videolist"},
		"wd": {query},
	}
	if page > 1 {
		values.Set("pg", strconv.Itoa(page))
	}
	return values
}

func (s *searchService) doSearchRequestPage(
	ctx context.Context,
	searchURL string,
	site model.ApiSite,
) ([]model.SearchResult, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("读取响应失败: %w", err)
	}

	var payload model.ApiSearchResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, 0, fmt.Errorf("解析JSON失败: %w", err)
	}

	if len(payload.List) == 0 {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(body, &raw); err == nil {
			if dataRaw, ok := raw["data"]; ok {
				var nested struct {
					PageCount model.FlexibleInt     `json:"pagecount"`
					List      []model.ApiSearchItem `json:"list"`
				}
				if json.Unmarshal(dataRaw, &nested) == nil && len(nested.List) > 0 {
					payload.List = nested.List
					if nested.PageCount > 0 {
						payload.PageCount = nested.PageCount
					}
				}
			}
		}
	}

	if len(payload.List) == 0 && payload.Code != 0 && payload.Msg != "" {
		return nil, 0, fmt.Errorf("API错误: %s", payload.Msg)
	}

	return s.parseResults(payload.List, site), int(payload.PageCount), nil
}
