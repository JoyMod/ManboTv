// internal/handler/search_handler.go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
	"github.com/JoyMod/ManboTV/backend/internal/service"
)

// SearchHandler 搜索处理器
type SearchHandler struct {
	service service.SearchService
	logger  *zap.Logger
	sites   []model.ApiSite
}

// NewSearchHandler 创建搜索处理器
func NewSearchHandler(service service.SearchService, logger *zap.Logger, sites []model.ApiSite) *SearchHandler {
	return &SearchHandler{
		service: service,
		logger:  logger,
		sites:   sites,
	}
}

// Search 搜索接口
func (h *SearchHandler) Search(c *gin.Context) {
	var req model.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("搜索参数无效", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "搜索关键词不能为空"))
		return
	}

	// 记录请求日志
	h.logger.Info("搜索请求",
		zap.String("query", req.Query),
		zap.Int("page", req.Page),
		zap.String("client_ip", c.ClientIP()),
	)

	// 执行搜索
	results, err := h.service.Search(c.Request.Context(), req.Query, h.sites)
	if err != nil {
		h.logger.Error("搜索失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "搜索服务暂时不可用"))
		return
	}

	// 分页处理
	paginatedResults := h.paginate(results, req.Page, req.PageSize)

	h.logger.Info("搜索完成",
		zap.String("query", req.Query),
		zap.Int("total_results", len(results)),
		zap.Int("returned_results", len(paginatedResults)),
	)

	c.JSON(http.StatusOK, model.Success(paginatedResults))
}

// SearchSingle 单源搜索接口
func (h *SearchHandler) SearchSingle(c *gin.Context) {
	var req model.SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "搜索关键词不能为空"))
		return
	}

	siteKey := c.Query("site")
	if siteKey == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "站点标识不能为空"))
		return
	}

	// 查找站点配置
	var targetSite *model.ApiSite
	for _, site := range h.sites {
		if site.Key == siteKey {
			targetSite = &site
			break
		}
	}

	if targetSite == nil {
		c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "站点不存在"))
		return
	}

	results, err := h.service.SearchSingle(c.Request.Context(), *targetSite, req.Query)
	if err != nil {
		h.logger.Error("单源搜索失败",
			zap.String("site", siteKey),
			zap.Error(err),
		)
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "搜索失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(results))
}

// GetSites 获取可用站点列表
func (h *SearchHandler) GetSites(c *gin.Context) {
	c.JSON(http.StatusOK, model.Success(h.sites))
}

// paginate 分页处理
func (h *SearchHandler) paginate(results []model.SearchResult, page, pageSize int) *model.PaginatedResponse {
	total := int64(len(results))

	// 计算起始和结束位置
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	if start > len(results) {
		start = len(results)
	}

	end := start + pageSize
	if end > len(results) {
		end = len(results)
	}

	var pageData []model.SearchResult
	if start < len(results) {
		pageData = results[start:end]
	} else {
		pageData = []model.SearchResult{}
	}

	return model.NewPaginatedResponse(pageData, page, pageSize, total)
}
