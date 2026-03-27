package handler

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/JoyMod/ManboTV/backend/internal/model"
	"github.com/JoyMod/ManboTV/backend/internal/service"
)

const (
	playBootstrapRelatedLimit     = 8
	playBootstrapExactTitleScore  = 60
	playBootstrapFuzzyTitleScore  = 30
	playBootstrapYearMatchScore   = 15
	playBootstrapTypeMatchScore   = 10
	playBootstrapEpisodeBiasScore = 4
	playBootstrapPreferSourceName = "优"
)

type PlayBootstrapHandler struct {
	detailService service.DetailService
	searchService service.SearchService
	storage       model.StorageService
	logger        *zap.Logger
	sites         []model.ApiSite
	adminStorage  model.AdminStorageService
	ownerUser     string
}

type playBootstrapRedirect struct {
	Source string `json:"source"`
	ID     string `json:"id"`
	Title  string `json:"title,omitempty"`
}

type playBootstrapResponse struct {
	Detail           *model.SearchResult    `json:"detail,omitempty"`
	Redirect         *playBootstrapRedirect `json:"redirect,omitempty"`
	IsFavorite       bool                   `json:"is_favorite"`
	AvailableSources []model.SearchResult   `json:"available_sources"`
	RelatedVideos    []model.SearchResult   `json:"related_videos"`
}

type playBootstrapRequest struct {
	Source        string
	ID            string
	FallbackTitle string
	SearchTitle   string
	YearHint      string
	TypeHint      string
	DirectEpisode string
	PreferLine    bool
}

func NewPlayBootstrapHandler(
	detailService service.DetailService,
	searchService service.SearchService,
	storage model.StorageService,
	logger *zap.Logger,
	sites []model.ApiSite,
	adminStorage model.AdminStorageService,
	ownerUser string,
) *PlayBootstrapHandler {
	return &PlayBootstrapHandler{
		detailService: detailService,
		searchService: searchService,
		storage:       storage,
		logger:        logger,
		sites:         sites,
		adminStorage:  adminStorage,
		ownerUser:     ownerUser,
	}
}

func (h *PlayBootstrapHandler) GetBootstrap(c *gin.Context) {
	response, statusCode, code, message := h.buildBootstrapResponse(c)
	if message != "" {
		c.JSON(statusCode, model.Error(code, message))
		return
	}

	c.JSON(http.StatusOK, model.Success(response))
}

func (h *PlayBootstrapHandler) GetBootstrapLegacy(c *gin.Context) {
	response, statusCode, _, message := h.buildBootstrapResponse(c)
	if message != "" {
		c.JSON(statusCode, gin.H{"error": message})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *PlayBootstrapHandler) buildBootstrapResponse(
	c *gin.Context,
) (*playBootstrapResponse, int, int, string) {
	request := parsePlayBootstrapRequest(c)
	sites := resolveVideoSites(
		c.Request.Context(),
		h.adminStorage,
		h.sites,
		c.GetString("username"),
		h.ownerUser,
	)
	policy := resolveContentPolicyFromRequest(c, h.adminStorage)

	if request.ID == "" || request.Source == "" || request.Source == "douban" {
		redirect := h.resolveRedirect(
			c.Request.Context(),
			sites,
			policy,
			request.SearchKeyword(),
			request.FallbackTitle,
			request.YearHint,
			request.TypeHint,
		)
		if redirect == nil {
			return nil, http.StatusNotFound, model.CodeNotFound, "未找到可播放资源"
		}

		return &playBootstrapResponse{
			Redirect: redirect,
		}, http.StatusOK, model.CodeSuccess, ""
	}

	targetSite, ok := findApiSiteByKey(sites, request.Source)
	if !ok {
		redirect := h.resolveRedirect(
			c.Request.Context(),
			sites,
			policy,
			request.SearchKeyword(),
			request.FallbackTitle,
			request.YearHint,
			request.TypeHint,
		)
		if redirect != nil {
			return &playBootstrapResponse{
				Redirect: redirect,
			}, http.StatusOK, model.CodeSuccess, ""
		}

		return nil, http.StatusBadRequest, model.CodeInvalidParams, "无效的播放源"
	}

	var (
		detail       *model.SearchResult
		searchResult []model.SearchResult
		isFavorite   bool
	)

	group, ctx := errgroup.WithContext(c.Request.Context())
	group.Go(func() error {
		result, err := h.detailService.GetDetail(ctx, targetSite, request.ID)
		if err != nil {
			return fmt.Errorf("get detail failed: %w", err)
		}
		detail = result
		return nil
	})

	initialKeyword := request.SearchKeyword()
	if initialKeyword != "" {
		group.Go(func() error {
			result, err := h.searchService.Search(ctx, initialKeyword, sites)
			if err != nil {
				h.logger.Warn(
					"play bootstrap search degraded",
					zap.String("query", initialKeyword),
					zap.Error(err),
				)
				return nil
			}
			searchResult = filterResults(result, policy)
			return nil
		})
	}

	username := resolveUsernameFromContext(c)
	if username != "" {
		group.Go(func() error {
			favorite, err := h.storage.GetFavorite(ctx, username, request.Source, request.ID)
			if err != nil {
				h.logger.Warn(
					"play bootstrap favorite degraded",
					zap.String("username", username),
					zap.String("source", request.Source),
					zap.String("id", request.ID),
					zap.Error(err),
				)
				return nil
			}
			isFavorite = favorite != nil
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		h.logger.Error(
			"play bootstrap failed",
			zap.String("source", request.Source),
			zap.String("id", request.ID),
			zap.Error(err),
		)
		return nil, http.StatusInternalServerError, model.CodeInternalError, "加载播放信息失败"
	}

	if detail == nil {
		return nil, http.StatusNotFound, model.CodeNotFound, "未找到播放详情"
	}

	filteredDetail, ok := filterResult(*detail, policy)
	if !ok {
		return nil, http.StatusNotFound, model.CodeNotFound, "影片不存在或不可访问"
	}

	resolvedDetail := applyDirectEpisodeFallback(filteredDetail, request.DirectEpisode)
	searchKeyword := firstNonEmpty(resolvedDetail.Title, initialKeyword, request.FallbackTitle)
	if shouldRefreshBootstrapSearch(searchKeyword, initialKeyword, searchResult) {
		refreshed, err := h.searchService.Search(c.Request.Context(), searchKeyword, sites)
		if err == nil {
			searchResult = filterResults(refreshed, policy)
		}
	}

	availableSources := buildPlayBootstrapSources(
		searchResult,
		searchKeyword,
		firstNonEmpty(resolvedDetail.Year, request.YearHint),
		request.TypeHint,
	)

	currentCandidate := buildCurrentSourceCandidate(resolvedDetail, request.Source, request.ID)
	availableSources = mergePlayBootstrapSources(currentCandidate, availableSources)
	availableSources = prioritizeBootstrapSources(
		availableSources,
		request.Source,
		request.ID,
		searchKeyword,
		firstNonEmpty(resolvedDetail.Year, request.YearHint),
		request.TypeHint,
	)

	response := &playBootstrapResponse{
		Detail:           &resolvedDetail,
		IsFavorite:       isFavorite,
		AvailableSources: availableSources,
		RelatedVideos:    buildPlayBootstrapRelated(searchResult, request.Source, request.ID),
	}

	if len(resolvedDetail.Episodes) == 0 && request.DirectEpisode == "" {
		redirect := selectRedirectFromSources(
			availableSources,
			request.Source,
			request.ID,
		)
		if redirect != nil {
			response.Redirect = redirect
		}
	}

	return response, http.StatusOK, model.CodeSuccess, ""
}

func parsePlayBootstrapRequest(c *gin.Context) playBootstrapRequest {
	return playBootstrapRequest{
		Source:        strings.TrimSpace(c.Query("source")),
		ID:            strings.TrimSpace(c.Query("id")),
		FallbackTitle: strings.TrimSpace(c.Query("title")),
		SearchTitle:   strings.TrimSpace(c.Query("stitle")),
		YearHint:      strings.TrimSpace(c.Query("year")),
		TypeHint:      strings.TrimSpace(c.Query("stype")),
		DirectEpisode: strings.TrimSpace(c.Query("ep")),
		PreferLine:    parseTruthyFlag(c.Query("prefer")),
	}
}

func (r playBootstrapRequest) SearchKeyword() string {
	return firstNonEmpty(r.SearchTitle, r.FallbackTitle)
}

func (h *PlayBootstrapHandler) resolveRedirect(
	ctx context.Context,
	sites []model.ApiSite,
	policy service.ContentPolicy,
	searchKeyword string,
	fallbackTitle string,
	yearHint string,
	typeHint string,
) *playBootstrapRedirect {
	keyword := firstNonEmpty(searchKeyword, fallbackTitle)
	if keyword == "" {
		return nil
	}

	results, err := h.searchService.Search(ctx, keyword, sites)
	if err != nil {
		h.logger.Warn(
			"play redirect search failed",
			zap.String("query", keyword),
			zap.Error(err),
		)
		return nil
	}

	filtered := filterResults(results, policy)
	candidates := buildPlayBootstrapSources(filtered, keyword, yearHint, typeHint)

	return selectRedirectFromSources(candidates, "", "")
}

func findApiSiteByKey(sites []model.ApiSite, key string) (model.ApiSite, bool) {
	for _, site := range sites {
		if site.Key == key {
			return site, true
		}
	}

	return model.ApiSite{}, false
}

func applyDirectEpisodeFallback(
	detail model.SearchResult,
	directEpisode string,
) model.SearchResult {
	if len(detail.Episodes) > 0 || strings.TrimSpace(directEpisode) == "" {
		return detail
	}

	detail.Episodes = []string{strings.TrimSpace(directEpisode)}
	detail.EpisodesTitles = []string{"1"}
	return detail
}

func buildCurrentSourceCandidate(
	detail model.SearchResult,
	source string,
	id string,
) []model.SearchResult {
	if strings.TrimSpace(source) == "" || strings.TrimSpace(id) == "" {
		return nil
	}

	if len(detail.Episodes) == 0 {
		return nil
	}

	detail.Source = source
	detail.ID = id
	return []model.SearchResult{detail}
}

func shouldRefreshBootstrapSearch(
	searchKeyword string,
	initialKeyword string,
	results []model.SearchResult,
) bool {
	if len(results) > 0 {
		return false
	}

	normalizedSearch := normalizeBootstrapTitle(searchKeyword)
	normalizedInitial := normalizeBootstrapTitle(initialKeyword)
	return normalizedSearch != "" && normalizedSearch != normalizedInitial
}

func buildPlayBootstrapSources(
	results []model.SearchResult,
	expectedTitle string,
	yearHint string,
	typeHint string,
) []model.SearchResult {
	if len(results) == 0 {
		return nil
	}

	normalizedExpected := normalizeBootstrapTitle(expectedTitle)
	candidates := make([]model.SearchResult, 0, len(results))

	for _, result := range results {
		if strings.TrimSpace(result.Source) == "" || strings.TrimSpace(result.ID) == "" {
			continue
		}
		if len(result.Episodes) == 0 {
			continue
		}
		if !bootstrapSourceMatchesTitle(result.Title, normalizedExpected) {
			continue
		}

		candidates = append(candidates, result)
	}

	if len(candidates) == 0 {
		for _, result := range results {
			if strings.TrimSpace(result.Source) == "" || strings.TrimSpace(result.ID) == "" {
				continue
			}
			if len(result.Episodes) == 0 {
				continue
			}
			candidates = append(candidates, result)
		}
	}

	deduped := make(map[string]model.SearchResult, len(candidates))
	for _, candidate := range candidates {
		deduped[candidate.Source+"+"+candidate.ID] = candidate
	}

	merged := make([]model.SearchResult, 0, len(deduped))
	for _, candidate := range deduped {
		merged = append(merged, candidate)
	}

	sort.SliceStable(merged, func(left, right int) bool {
		leftScore := scoreBootstrapSource(
			merged[left],
			normalizedExpected,
			yearHint,
			typeHint,
		)
		rightScore := scoreBootstrapSource(
			merged[right],
			normalizedExpected,
			yearHint,
			typeHint,
		)
		return rightScore < leftScore
	})

	return merged
}

func buildPlayBootstrapRelated(
	results []model.SearchResult,
	currentSource string,
	currentID string,
) []model.SearchResult {
	if len(results) == 0 {
		return nil
	}

	related := make([]model.SearchResult, 0, playBootstrapRelatedLimit)
	for _, item := range results {
		if item.Source == currentSource && item.ID == currentID {
			continue
		}
		related = append(related, item)
		if len(related) >= playBootstrapRelatedLimit {
			break
		}
	}

	return related
}

func mergePlayBootstrapSources(
	primary []model.SearchResult,
	incoming []model.SearchResult,
) []model.SearchResult {
	merged := make(map[string]model.SearchResult, len(primary)+len(incoming))
	for _, item := range primary {
		merged[item.Source+"+"+item.ID] = item
	}
	for _, item := range incoming {
		merged[item.Source+"+"+item.ID] = item
	}

	result := make([]model.SearchResult, 0, len(merged))
	for _, item := range merged {
		result = append(result, item)
	}

	return result
}

func prioritizeBootstrapSources(
	sources []model.SearchResult,
	currentSource string,
	currentID string,
	expectedTitle string,
	yearHint string,
	typeHint string,
) []model.SearchResult {
	if len(sources) <= 1 {
		return sources
	}

	normalizedExpected := normalizeBootstrapTitle(expectedTitle)
	sort.SliceStable(sources, func(left, right int) bool {
		leftCurrent := sources[left].Source == currentSource && sources[left].ID == currentID
		rightCurrent := sources[right].Source == currentSource && sources[right].ID == currentID
		if leftCurrent != rightCurrent {
			return leftCurrent
		}

		leftScore := scoreBootstrapSource(sources[left], normalizedExpected, yearHint, typeHint)
		rightScore := scoreBootstrapSource(sources[right], normalizedExpected, yearHint, typeHint)
		if leftScore != rightScore {
			return leftScore > rightScore
		}

		return sources[left].Title < sources[right].Title
	})

	return sources
}

func selectRedirectFromSources(
	sources []model.SearchResult,
	currentSource string,
	currentID string,
) *playBootstrapRedirect {
	for _, item := range sources {
		if item.Source == currentSource && item.ID == currentID {
			continue
		}
		return &playBootstrapRedirect{
			Source: item.Source,
			ID:     item.ID,
			Title:  item.Title,
		}
	}

	return nil
}

func scoreBootstrapSource(
	item model.SearchResult,
	normalizedExpected string,
	yearHint string,
	typeHint string,
) int {
	score := 0
	normalizedTitle := normalizeBootstrapTitle(item.Title)
	switch {
	case normalizedExpected != "" && normalizedTitle == normalizedExpected:
		score += playBootstrapExactTitleScore
	case normalizedExpected != "" && (strings.Contains(normalizedTitle, normalizedExpected) ||
		strings.Contains(normalizedExpected, normalizedTitle)):
		score += playBootstrapFuzzyTitleScore
	}

	if yearHint != "" && item.Year == yearHint {
		score += playBootstrapYearMatchScore
	}

	if bootstrapSourceMatchesType(item, typeHint) {
		score += playBootstrapTypeMatchScore
	}

	if len(item.Episodes) > 1 {
		score += playBootstrapEpisodeBiasScore
	}

	if strings.Contains(item.SourceName, playBootstrapPreferSourceName) {
		score++
	}

	return score
}

func bootstrapSourceMatchesType(item model.SearchResult, typeHint string) bool {
	switch strings.TrimSpace(typeHint) {
	case "movie":
		return len(item.Episodes) <= 1
	case "tv":
		return len(item.Episodes) > 1
	default:
		return true
	}
}

func bootstrapSourceMatchesTitle(title string, normalizedExpected string) bool {
	if normalizedExpected == "" {
		return true
	}

	normalizedTitle := normalizeBootstrapTitle(title)
	return normalizedTitle == normalizedExpected ||
		strings.Contains(normalizedTitle, normalizedExpected) ||
		strings.Contains(normalizedExpected, normalizedTitle)
}

func normalizeBootstrapTitle(title string) string {
	replacer := strings.NewReplacer(
		" ", "",
		"-", "",
		"_", "",
		".", "",
		":", "",
		"：", "",
		"!", "",
		"！", "",
		"?", "",
		"？", "",
		",", "",
		"，", "",
		"。", "",
		"·", "",
		"'", "",
		"\"", "",
	)

	return strings.ToLower(strings.TrimSpace(replacer.Replace(title)))
}

func parseTruthyFlag(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes":
		return true
	default:
		return false
	}
}
