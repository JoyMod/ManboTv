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
	service      service.DetailService
	logger       *zap.Logger
	sites        []model.ApiSite
	adminStorage model.AdminStorageService
	ownerUser    string
}

// NewDetailHandler 创建详情处理器
func NewDetailHandler(
	service service.DetailService,
	logger *zap.Logger,
	sites []model.ApiSite,
	adminStorage model.AdminStorageService,
	ownerUser string,
) *DetailHandler {
	return &DetailHandler{
		service:      service,
		logger:       logger,
		sites:        sites,
		adminStorage: adminStorage,
		ownerUser:    ownerUser,
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

	sites := resolveVideoSites(c.Request.Context(), h.adminStorage, h.sites, c.GetString("username"), h.ownerUser)

	// 查找源配置
	var targetSite *model.ApiSite
	for _, site := range sites {
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
	filteredDetail, ok := filterResult(*detail, resolveContentPolicyFromRequest(c, h.adminStorage))
	if !ok {
		c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "影片不存在或不可访问"))
		return
	}

	c.JSON(http.StatusOK, model.Success(filteredDetail))
}

// GetDetails 从多个源获取详情
func (h *DetailHandler) GetDetails(c *gin.Context) {
	id := c.Query("id")

	if id == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少影片ID"))
		return
	}

	sites := resolveVideoSites(c.Request.Context(), h.adminStorage, h.sites, c.GetString("username"), h.ownerUser)

	h.logger.Info("多源获取详情", zap.String("id", id))

	details, err := h.service.GetDetails(c.Request.Context(), sites, id)
	if err != nil {
		h.logger.Error("获取详情失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取详情失败"))
		return
	}
	details = filterDetails(details, resolveContentPolicyFromRequest(c, h.adminStorage))

	c.JSON(http.StatusOK, model.Success(details))
}
