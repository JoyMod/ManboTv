package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/JoyMod/ManboTV/backend/internal/model"
	"github.com/JoyMod/ManboTV/backend/internal/service"
)

const searchBootstrapHistoryLimit = 20

type SearchBootstrapHandler struct {
	searchService     service.SearchService
	suggestionService service.SuggestionService
	storage           model.StorageService
	logger            *zap.Logger
	sites             []model.ApiSite
	adminStorage      model.AdminStorageService
	ownerUser         string
}

type searchBootstrapResponse struct {
	Query            string                        `json:"query"`
	NormalizedQuery  string                        `json:"normalized_query"`
	Results          []model.SearchResult          `json:"results"`
	Aggregates       []model.SearchAggregateResult `json:"aggregates"`
	Facets           model.SearchFacets            `json:"facets"`
	History          []string                      `json:"history"`
	Suggestions      []string                      `json:"suggestions"`
	SourceStatus     map[string]string             `json:"source_status"`
	SourceStatusList []model.SearchSourceStatus    `json:"source_status_items"`
	PageInfo         model.PageInfo                `json:"page_info"`
	Execution        model.SearchExecutionInfo     `json:"execution"`
	SelectedTypes    []string                      `json:"selected_types,omitempty"`
	SelectedSources  []string                      `json:"selected_sources,omitempty"`
	SelectedSort     string                        `json:"selected_sort,omitempty"`
	SelectedView     string                        `json:"selected_view,omitempty"`
	SelectedYearFrom int                           `json:"selected_year_from,omitempty"`
	SelectedYearTo   int                           `json:"selected_year_to,omitempty"`
	SelectedMode     string                        `json:"selected_source_mode,omitempty"`
}

func NewSearchBootstrapHandler(
	searchService service.SearchService,
	suggestionService service.SuggestionService,
	storage model.StorageService,
	logger *zap.Logger,
	sites []model.ApiSite,
	adminStorage model.AdminStorageService,
	ownerUser string,
) *SearchBootstrapHandler {
	return &SearchBootstrapHandler{
		searchService:     searchService,
		suggestionService: suggestionService,
		storage:           storage,
		logger:            logger,
		sites:             sites,
		adminStorage:      adminStorage,
		ownerUser:         ownerUser,
	}
}

func (h *SearchBootstrapHandler) GetBootstrap(c *gin.Context) {
	request := parseSearchRequest(c)
	response, statusCode, code, message := h.buildBootstrapResponseWithRequest(c, request, false)
	if message != "" {
		c.JSON(statusCode, model.Error(code, message))
		return
	}

	c.JSON(http.StatusOK, model.Success(response))
}

func (h *SearchBootstrapHandler) GetBootstrapLegacy(c *gin.Context) {
	request := parseSearchRequest(c)
	request.Page = searchDefaultPage
	request.PageSize = searchMaximumPageSize

	response, statusCode, _, message := h.buildBootstrapResponseWithRequest(c, request, true)
	if message != "" {
		c.JSON(statusCode, gin.H{"error": message})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *SearchBootstrapHandler) buildBootstrapResponse(
	c *gin.Context,
) (*searchBootstrapResponse, int, int, string) {
	request := parseSearchRequest(c)
	return h.buildBootstrapResponseWithRequest(c, request, false)
}

func (h *SearchBootstrapHandler) buildBootstrapResponseWithRequest(
	c *gin.Context,
	request model.SearchRequest,
	forceReturnAll bool,
) (*searchBootstrapResponse, int, int, string) {
	query := strings.TrimSpace(request.Query)
	username := resolveUsernameFromContext(c)
	policy := resolveContentPolicyFromRequest(c, h.adminStorage)
	sites := resolveVideoSites(
		c.Request.Context(),
		h.adminStorage,
		h.sites,
		c.GetString("username"),
		h.ownerUser,
	)

	response := &searchBootstrapResponse{
		Query:            query,
		NormalizedQuery:  "",
		Results:          []model.SearchResult{},
		Aggregates:       []model.SearchAggregateResult{},
		Facets:           model.SearchFacets{},
		History:          []string{},
		Suggestions:      []string{},
		SourceStatus:     map[string]string{},
		SourceStatusList: []model.SearchSourceStatus{},
		PageInfo:         model.PageInfo{Page: request.Page, PageSize: request.PageSize, TotalPages: 1},
		Execution: model.SearchExecutionInfo{
			Query:            query,
			NormalizedQuery:  "",
			CompletedSources: 0,
			TotalSources:     len(sites),
		},
	}

	group, ctx := errgroup.WithContext(c.Request.Context())
	if username != "" {
		group.Go(func() error {
			history, err := h.storage.GetSearchHistory(ctx, username, searchBootstrapHistoryLimit)
			if err != nil {
				h.logger.Warn("search bootstrap history degraded", zap.String("username", username), zap.Error(err))
				return nil
			}
			response.History = history
			return nil
		})
	}

	if query != "" {
		group.Go(func() error {
			params := buildSearchParams(ctx, request, h.adminStorage)
			if forceReturnAll {
				params = buildLegacySearchParams(ctx, request, h.adminStorage)
			}
			envelope, err := h.searchService.SearchAdvanced(
				ctx,
				params,
				sites,
				policy,
			)
			if err != nil {
				return err
			}
			response.Query = envelope.Query
			response.NormalizedQuery = envelope.NormalizedQuery
			response.Results = envelope.Results
			response.Aggregates = envelope.Aggregates
			response.Facets = envelope.Facets
			response.SourceStatus = envelope.LegacySourceMap
			response.SourceStatusList = envelope.SourceStatus
			response.PageInfo = envelope.PageInfo
			response.Execution = envelope.Execution
			response.SelectedTypes = envelope.SelectedTypes
			response.SelectedSources = envelope.SelectedSources
			response.SelectedSort = envelope.SelectedSort
			response.SelectedView = envelope.SelectedView
			response.SelectedYearFrom = envelope.SelectedYearFrom
			response.SelectedYearTo = envelope.SelectedYearTo
			response.SelectedMode = envelope.SelectedMode
			return nil
		})

		group.Go(func() error {
			suggestions, err := h.suggestionService.GetSuggestions(ctx, query)
			if err != nil {
				h.logger.Warn("search bootstrap suggestions degraded", zap.String("query", query), zap.Error(err))
				return nil
			}
			if suggestions == nil {
				response.Suggestions = []string{}
				return nil
			}
			response.Suggestions = suggestions
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		h.logger.Error("search bootstrap failed", zap.String("query", query), zap.Error(err))
		return nil, http.StatusInternalServerError, model.CodeInternalError, "搜索结果加载失败"
	}

	if username != "" && query != "" {
		if err := h.storage.AddSearchHistory(c.Request.Context(), username, query); err == nil {
			if history, err := h.storage.GetSearchHistory(
				c.Request.Context(),
				username,
				searchBootstrapHistoryLimit,
			); err == nil {
				response.History = history
			}
		}
		h.suggestionService.AddKeyword(c.Request.Context(), query)
	}

	return response, http.StatusOK, model.CodeSuccess, ""
}
