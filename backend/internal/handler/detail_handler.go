// internal/handler/detail_handler.go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
	"github.com/JoyMod/ManboTV/backend/internal/service"
)

// DetailHandler 详情处理器
type DetailHandler struct {
	service service.DetailService
	logger  *zap.Logger
	sites   []model.ApiSite
}

// NewDetailHandler 创建详情处理器
func NewDetailHandler(service service.DetailService, logger *zap.Logger, sites []model.ApiSite) *DetailHandler {
	return &DetailHandler{
		service: service,
		logger:  logger,
		sites:   sites,
	}
}

// GetDetail 获取详情
func (h *DetailHandler) GetDetail(c *gin.Context) {
	id := c.Query("id")
	sourceKey := c.Query("source")

	if id == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少影片ID"))
		return
	}

	// 查找源配置
	var targetSite *model.ApiSite
	for _, site := range h.sites {
		if site.Key == sourceKey {
			targetSite = &site
			break
		}
	}

	if targetSite == nil {
		c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "源不存在"))
		return
	}

	h.logger.Info("获取详情",
		zap.String("id", id),
		zap.String("source", sourceKey),
	)

	detail, err := h.service.GetDetail(c.Request.Context(), *targetSite, id)
	if err != nil {
		h.logger.Error("获取详情失败",
			zap.String("id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取详情失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(detail))
}

// GetDetails 从多个源获取详情
func (h *DetailHandler) GetDetails(c *gin.Context) {
	id := c.Query("id")

	if id == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少影片ID"))
		return
	}

	h.logger.Info("多源获取详情", zap.String("id", id))

	details, err := h.service.GetDetails(c.Request.Context(), h.sites, id)
	if err != nil {
		h.logger.Error("获取详情失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取详情失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(details))
}
