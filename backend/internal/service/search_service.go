// internal/service/search_service.go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/config"
	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// SearchService 搜索服务接口
type SearchService interface {
	Search(ctx context.Context, query string, sites []model.ApiSite) ([]model.SearchResult, error)
	SearchSingle(ctx context.Context, site model.ApiSite, query string) ([]model.SearchResult, error)
}

type searchCacheEntry struct {
	Results   []model.SearchResult
	ExpiresAt time.Time
}

type sourceSearchResult struct {
	Site    model.ApiSite
	Results []model.SearchResult
	Err     error
}

// searchService 搜索服务实现
type searchService struct {
	client          *http.Client
	config          *config.SearchConfig
	logger          *zap.Logger
	maxRetries      int
	sourceTimeout   time.Duration
	fastReturnAfter time.Duration
	cacheTTL        time.Duration
	cacheMaxEntries int

	cacheMu sync.RWMutex
	cache   map[string]searchCacheEntry
}

const (
	defaultSearchTimeout         = 10 * time.Second
	defaultSourceTimeout         = 5 * time.Second
	defaultFastReturnAfter       = 0
	defaultSearchRetryTimes      = 2
	defaultSearchCacheMaxEntries = 512
)

// NewSearchService 创建搜索服务
func NewSearchService(cfg *config.SearchConfig, httpCfg *config.HTTPClientConfig, logger *zap.Logger) SearchService {
	if cfg == nil {
		cfg = &config.SearchConfig{}
	}
	if httpCfg == nil {
		httpCfg = &config.HTTPClientConfig{}
	}

	searchTimeout := cfg.Timeout
	if searchTimeout <= 0 {
		searchTimeout = defaultSearchTimeout
	}

	sourceTimeout := cfg.SourceTimeout
	if sourceTimeout <= 0 {
		sourceTimeout = defaultSourceTimeout
	}
	if sourceTimeout > searchTimeout {
		sourceTimeout = searchTimeout
	}

	fastReturnAfter := cfg.FastReturnAfter
	if fastReturnAfter <= 0 {
		fastReturnAfter = defaultFastReturnAfter
	}
	if fastReturnAfter > 0 && fastReturnAfter > sourceTimeout {
		fastReturnAfter = sourceTimeout
	}

	cacheTTL := time.Duration(cfg.CacheMinutes) * time.Minute
	if cfg.CacheMinutes <= 0 {
		cacheTTL = 0
	}

	client := &http.Client{
		Timeout: searchTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        httpCfg.MaxIdleConns,
			MaxIdleConnsPerHost: httpCfg.MaxIdleConnsPerHost,
			IdleConnTimeout:     httpCfg.IdleConnTimeout,
		},
	}

	return &searchService{
		client:          client,
		config:          cfg,
		logger:          logger,
		maxRetries:      defaultSearchRetryTimes,
		sourceTimeout:   sourceTimeout,
		fastReturnAfter: fastReturnAfter,
		cacheTTL:        cacheTTL,
		cacheMaxEntries: defaultSearchCacheMaxEntries,
		cache:           make(map[string]searchCacheEntry, defaultSearchCacheMaxEntries),
	}
}

// Search 多源聚合搜索
func (s *searchService) Search(ctx context.Context, query string, sites []model.ApiSite) ([]model.SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" || len(sites) == 0 {
		return []model.SearchResult{}, nil
	}

	cacheKey := s.buildCacheKey(query, sites)
	if cached, ok := s.getCachedResults(cacheKey); ok {
		s.logger.Debug("搜索缓存命中",
			zap.String("query", query),
			zap.Int("source_count", len(sites)),
			zap.Int("result_count", len(cached)),
		)
		return cached, nil
	}

	searchCtx, cancel := context.WithTimeout(ctx, s.resolveSearchTimeout())
	defer cancel()

	resultCh := make(chan sourceSearchResult, len(sites))
	sem := make(chan struct{}, s.resolveMaxConcurrent())

	var wg sync.WaitGroup
	for _, site := range sites {
		site := site
		wg.Add(1)
		go func() {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
			case <-searchCtx.Done():
				return
			}
			defer func() { <-sem }()

			sourceCtx := searchCtx
			release := func() {}
			if s.sourceTimeout > 0 {
				sourceCtx, release = context.WithTimeout(searchCtx, s.sourceTimeout)
			}
			defer release()

			res, err := s.searchSingle(sourceCtx, site, query)
			select {
			case resultCh <- sourceSearchResult{Site: site, Results: res, Err: err}:
			case <-searchCtx.Done():
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	var fastTimer *time.Timer
	var fastTimerCh <-chan time.Time
	if s.fastReturnAfter > 0 {
		fastTimer = time.NewTimer(s.fastReturnAfter)
		fastTimerCh = fastTimer.C
		defer fastTimer.Stop()
	}

	results := make([]model.SearchResult, 0)
	completed := 0
	degraded := false

COLLECT:
	for {
		select {
		case item, ok := <-resultCh:
			if !ok {
				break COLLECT
			}
			completed++
			if item.Err != nil {
				s.logger.Warn("搜索源失败",
					zap.String("source", item.Site.Name),
					zap.Error(item.Err),
				)
				continue
			}
			results = append(results, item.Results...)

		case <-fastTimerCh:
			// 快源优先：如果已有部分源完成，直接降级返回，避免被慢源拖尾
			if completed > 0 && completed < len(sites) {
				degraded = true
				cancel()
				break COLLECT
			}

		case <-searchCtx.Done():
			degraded = completed < len(sites)
			break COLLECT
		}
	}

	// 尽量吸收已完成但尚未消费的结果，不阻塞返回
	for {
		select {
		case item, ok := <-resultCh:
			if !ok {
				goto DONE
			}
			completed++
			if item.Err != nil {
				s.logger.Warn("搜索源失败",
					zap.String("source", item.Site.Name),
					zap.Error(item.Err),
				)
				continue
			}
			results = append(results, item.Results...)
		default:
			goto DONE
		}
	}

DONE:
	if s.cacheTTL > 0 {
		s.setCachedResults(cacheKey, results)
	}

	if degraded {
		s.logger.Warn("搜索降级返回",
			zap.String("query", query),
			zap.Int("completed_sources", completed),
			zap.Int("total_sources", len(sites)),
			zap.Int("result_count", len(results)),
		)
	}

	s.logger.Info("搜索完成",
		zap.String("query", query),
		zap.Int("source_count", len(sites)),
		zap.Int("completed_sources", completed),
		zap.Int("result_count", len(results)),
	)

	return results, nil
}

// SearchSingle 单源搜索
func (s *searchService) SearchSingle(ctx context.Context, site model.ApiSite, query string) ([]model.SearchResult, error) {
	if s.sourceTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.sourceTimeout)
		defer cancel()
	}
	return s.searchSingle(ctx, site, query)
}

// searchSingle 内部单源搜索实现
func (s *searchService) searchSingle(ctx context.Context, site model.ApiSite, query string) ([]model.SearchResult, error) {
	var results []model.SearchResult
	queries := []string{query}
	normalized := normalizeSearchKeyword(query)
	if normalized != "" && normalized != query {
		queries = append(queries, normalized)
	}

	var err error
	for _, q := range queries {
		searchURL, buildErr := appendProviderQuery(site.API, url.Values{
			"ac": {"videolist"},
			"wd": {q},
		})
		if buildErr != nil {
			err = fmt.Errorf("构建搜索URL失败: %w", buildErr)
			continue
		}
		err = s.withRetry(ctx, func() error {
			res, reqErr := s.doSearchRequest(ctx, searchURL, site)
			if reqErr != nil {
				return reqErr
			}
			results = res
			return nil
		})
		if err == nil && len(results) > 0 {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("搜索源 %s 失败: %w", site.Name, err)
	}

	// 设置来源信息
	for i := range results {
		results[i].Source = site.Key
		results[i].SourceName = site.Name
	}

	return results, nil
}

func normalizeSearchKeyword(query string) string {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		"！", "",
		"!", "",
		"？", "",
		"?", "",
		"：", "",
		":", "",
		"·", "",
		"，", "",
		",", "",
		"。", "",
		" ", "",
	)
	return replacer.Replace(trimmed)
}

// doSearchRequest 执行搜索请求
func (s *searchService) doSearchRequest(ctx context.Context, searchURL string, site model.ApiSite) ([]model.SearchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var apiResp model.ApiSearchResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	if len(apiResp.List) == 0 {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(body, &raw); err == nil {
			if dataRaw, ok := raw["data"]; ok {
				var nested struct {
					List []model.ApiSearchItem `json:"list"`
				}
				if json.Unmarshal(dataRaw, &nested) == nil && len(nested.List) > 0 {
					apiResp.List = nested.List
				}
			}
		}
	}

	if len(apiResp.List) == 0 && apiResp.Code != 0 && apiResp.Msg != "" {
		return nil, fmt.Errorf("API错误: %s", apiResp.Msg)
	}

	return s.parseResults(apiResp.List, site), nil
}

// parseResults 解析搜索结果
func (s *searchService) parseResults(items []model.ApiSearchItem, site model.ApiSite) []model.SearchResult {
	results := make([]model.SearchResult, 0, len(items))

	for _, item := range items {
		episodes, titles := s.parseEpisodes(item.VodPlayURL)
		if len(episodes) == 0 {
			continue // 过滤掉无播放链接的结果
		}

		result := model.SearchResult{
			ID:             item.VodID.String(),
			Title:          strings.TrimSpace(item.VodName),
			Poster:         normalizeMediaURL(item.VodPic, site.API),
			Episodes:       episodes,
			EpisodesTitles: titles,
			Source:         site.Key,
			SourceName:     site.Name,
			Class:          item.VodClass,
			Year:           s.extractYear(item.VodYear),
			Desc:           item.VodContent,
			TypeName:       item.TypeName,
			DoubanID:       int(item.VodDoubanID),
		}

		results = append(results, EnrichSearchResult(result))
	}

	return results
}

// parseEpisodes 解析剧集链接
func (s *searchService) parseEpisodes(vodPlayURL string) ([]string, []string) {
	if vodPlayURL == "" {
		return nil, nil
	}

	var bestEpisodes []string
	var bestTitles []string

	// 按 $$$ 分割不同播放源
	sources := strings.Split(vodPlayURL, model.EpisodeURLSeparator)

	for _, source := range sources {
		var episodes []string
		var titles []string

		// 按 # 分割剧集
		items := strings.Split(source, model.EpisodeItemSeparator)

		for _, item := range items {
			// 按 $ 分割标题和链接
			parts := strings.Split(item, model.EpisodeTitleURLSeparator)
			if len(parts) != 2 {
				continue
			}

			title := parts[0]
			url := parts[1]

			// 只保留 m3u8 链接
			if strings.HasSuffix(url, model.M3U8Suffix) {
				titles = append(titles, title)
				episodes = append(episodes, url)
			}
		}

		// 选择剧集最多的播放源
		if len(episodes) > len(bestEpisodes) {
			bestEpisodes = episodes
			bestTitles = titles
		}
	}

	return bestEpisodes, bestTitles
}

// extractYear 提取年份
func (s *searchService) extractYear(year string) string {
	if year == "" {
		return "unknown"
	}

	// 尝试匹配4位数字年份
	for i := 0; i <= len(year)-4; i++ {
		if year[i] >= '0' && year[i] <= '9' &&
			year[i+1] >= '0' && year[i+1] <= '9' &&
			year[i+2] >= '0' && year[i+2] <= '9' &&
			year[i+3] >= '0' && year[i+3] <= '9' {
			return year[i : i+4]
		}
	}

	return "unknown"
}

func (s *searchService) resolveSearchTimeout() time.Duration {
	if s.config == nil || s.config.Timeout <= 0 {
		return defaultSearchTimeout
	}
	return s.config.Timeout
}

func (s *searchService) resolveMaxConcurrent() int {
	if s.config == nil || s.config.MaxConcurrent <= 0 {
		return 8
	}
	return s.config.MaxConcurrent
}

func (s *searchService) buildCacheKey(query string, sites []model.ApiSite) string {
	keys := make([]string, 0, len(sites))
	for _, site := range sites {
		key := strings.TrimSpace(site.Key)
		if key != "" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	return strings.ToLower(strings.TrimSpace(query)) + "|" + strings.Join(keys, ",")
}

func (s *searchService) getCachedResults(cacheKey string) ([]model.SearchResult, bool) {
	if s.cacheTTL <= 0 {
		return nil, false
	}

	now := time.Now()

	s.cacheMu.RLock()
	entry, ok := s.cache[cacheKey]
	s.cacheMu.RUnlock()
	if !ok {
		return nil, false
	}

	if now.After(entry.ExpiresAt) {
		s.cacheMu.Lock()
		if stale, exists := s.cache[cacheKey]; exists && now.After(stale.ExpiresAt) {
			delete(s.cache, cacheKey)
		}
		s.cacheMu.Unlock()
		return nil, false
	}

	return cloneSearchResults(entry.Results), true
}

func (s *searchService) setCachedResults(cacheKey string, results []model.SearchResult) {
	if s.cacheTTL <= 0 {
		return
	}

	now := time.Now()

	s.cacheMu.Lock()
	s.cache[cacheKey] = searchCacheEntry{
		Results:   cloneSearchResults(results),
		ExpiresAt: now.Add(s.cacheTTL),
	}
	s.cleanupCacheLocked(now)
	s.cacheMu.Unlock()
}

func (s *searchService) cleanupCacheLocked(now time.Time) {
	for key, entry := range s.cache {
		if now.After(entry.ExpiresAt) {
			delete(s.cache, key)
		}
	}

	if len(s.cache) <= s.cacheMaxEntries {
		return
	}

	overflow := len(s.cache) - s.cacheMaxEntries
	for key := range s.cache {
		delete(s.cache, key)
		overflow--
		if overflow <= 0 {
			break
		}
	}
}

func cloneSearchResults(results []model.SearchResult) []model.SearchResult {
	cloned := make([]model.SearchResult, len(results))
	for i := range results {
		cloned[i] = results[i]
		cloned[i].Episodes = append([]string(nil), results[i].Episodes...)
		cloned[i].EpisodesTitles = append([]string(nil), results[i].EpisodesTitles...)
	}
	return cloned
}

// withRetry 带重试的执行
func (s *searchService) withRetry(ctx context.Context, fn func() error) error {
	var err error
	for i := 0; i < s.maxRetries; i++ {
		if err = fn(); err == nil {
			return nil
		}

		if !s.shouldRetry(err) {
			return err
		}

		if i < s.maxRetries-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(i+1) * 200 * time.Millisecond):
			}
		}
	}
	return err
}

func (s *searchService) shouldRetry(err error) bool {
	if err == nil {
		return false
	}

	message := err.Error()
	if strings.Contains(message, "解析JSON失败") || strings.Contains(message, "API错误") {
		return false
	}

	if strings.HasPrefix(message, "HTTP ") {
		statusCode := 0
		if _, scanErr := fmt.Sscanf(message, "HTTP %d", &statusCode); scanErr == nil {
			if statusCode >= 400 && statusCode < 500 && statusCode != http.StatusRequestTimeout && statusCode != http.StatusTooManyRequests {
				return false
			}
		}
	}

	return true
}
