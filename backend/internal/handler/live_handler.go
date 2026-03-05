// internal/handler/live_handler.go
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
	"github.com/JoyMod/ManboTV/backend/internal/service"
)

// LiveHandler 直播处理器
type LiveHandler struct {
	service service.LiveService
	logger  *zap.Logger
}

// NewLiveHandler 创建直播处理器
func NewLiveHandler(service service.LiveService, logger *zap.Logger) *LiveHandler {
	return &LiveHandler{
		service: service,
		logger:  logger,
	}
}

// GetSources 获取直播源列表
// GET /api/v1/live/sources
func (h *LiveHandler) GetSources(c *gin.Context) {
	sources, err := h.service.GetSources(c.Request.Context())
	if err != nil {
		h.logger.Error("获取直播源列表失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取直播源失败"))
		return
	}

	h.logger.Debug("获取直播源列表", zap.Int("count", len(sources)))
	c.JSON(http.StatusOK, model.Success(sources))
}

// GetChannels 获取频道列表
// GET /api/v1/live/channels?source=xxx
func (h *LiveHandler) GetChannels(c *gin.Context) {
	source := c.Query("source")
	if source == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少source参数"))
		return
	}

	channels, err := h.service.GetChannels(c.Request.Context(), source)
	if err != nil {
		h.logger.Error("获取频道列表失败",
			zap.String("source", source),
			zap.Error(err),
		)
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取频道列表失败"))
		return
	}

	h.logger.Debug("获取频道列表",
		zap.String("source", source),
		zap.Int("count", channels.ChannelNumber),
	)

	c.JSON(http.StatusOK, model.Success(channels))
}

// GetEPG 获取节目单
// GET /api/v1/live/epg?source=xxx&tvgId=xxx
func (h *LiveHandler) GetEPG(c *gin.Context) {
	source := c.Query("source")
	tvgID := c.Query("tvgId")

	if source == "" || tvgID == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少source或tvgId参数"))
		return
	}

	epg, err := h.service.GetEPG(c.Request.Context(), source, tvgID)
	if err != nil {
		h.logger.Error("获取节目单失败",
			zap.String("source", source),
			zap.String("tvg_id", tvgID),
			zap.Error(err),
		)
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取节目单失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(epg))
}

// Precheck 直播预检
// POST /api/v1/live/precheck
func (h *LiveHandler) Precheck(c *gin.Context) {
	var req struct {
		URL string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "URL不能为空"))
		return
	}

	h.logger.Debug("直播预检", zap.String("url", req.URL))

	valid, err := h.service.Precheck(c.Request.Context(), req.URL)
	if err != nil {
		h.logger.Error("直播预检失败",
			zap.String("url", req.URL),
			zap.Error(err),
		)
	}

	c.JSON(http.StatusOK, model.Success(gin.H{
		"valid": valid,
		"url":   req.URL,
	}))
}
