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
	Query        string               `json:"query"`
	Results      []model.SearchResult `json:"results"`
	History      []string             `json:"history"`
	Suggestions  []string             `json:"suggestions"`
	SourceStatus map[string]string    `json:"source_status"`
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
	response, statusCode, code, message := h.buildBootstrapResponse(c)
	if message != "" {
		c.JSON(statusCode, model.Error(code, message))
		return
	}

	c.JSON(http.StatusOK, model.Success(response))
}

func (h *SearchBootstrapHandler) GetBootstrapLegacy(c *gin.Context) {
	response, statusCode, _, message := h.buildBootstrapResponse(c)
	if message != "" {
		c.JSON(statusCode, gin.H{"error": message})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *SearchBootstrapHandler) buildBootstrapResponse(
	c *gin.Context,
) (*searchBootstrapResponse, int, int, string) {
	query := strings.TrimSpace(c.Query("q"))
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
		Query:        query,
		Results:      []model.SearchResult{},
		History:      []string{},
		Suggestions:  []string{},
		SourceStatus: map[string]string{},
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
			results, err := h.searchService.Search(ctx, query, sites)
			if err != nil {
				return err
			}
			response.Results = filterResults(results, policy)
			response.SourceStatus = buildSearchBootstrapSourceStatus(response.Results)
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

func buildSearchBootstrapSourceStatus(results []model.SearchResult) map[string]string {
	status := make(map[string]string)
	for _, item := range results {
		source := strings.TrimSpace(item.Source)
		if source == "" {
			continue
		}
		status[source] = "done"
	}

	return status
}
