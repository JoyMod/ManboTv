package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// GetDetailLegacy handles GET /api/detail.
func (h *DetailHandler) GetDetailLegacy(c *gin.Context) {
	id := strings.TrimSpace(c.Query("id"))
	sourceKey := strings.TrimSpace(c.Query("source"))

	if id == "" || sourceKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数"})
		return
	}

	sites := resolveVideoSites(c.Request.Context(), h.adminStorage, h.sites, c.GetString("username"), h.ownerUser)

	var targetSite *model.ApiSite
	for _, site := range sites {
		if site.Key == sourceKey {
			targetSite = &site
			break
		}
	}

	if targetSite == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的API来源"})
		return
	}

	detail, err := h.service.GetDetail(c.Request.Context(), *targetSite, id)
	if err != nil {
		h.logger.Error("legacy detail failed", zap.Error(err), zap.String("id", id), zap.String("source", sourceKey))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	filteredDetail, ok := filterResult(*detail, resolveContentPolicyFromRequest(c, h.adminStorage))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "资源已被内容策略拦截"})
		return
	}

	c.JSON(http.StatusOK, filteredDetail)
}
