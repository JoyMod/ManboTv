package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// LiveSourceActionRequest 直播源操作请求
type LiveSourceActionRequest struct {
	Action string `json:"action" binding:"required"`
	Key    string `json:"key,omitempty"`
	Name   string `json:"name,omitempty"`
	URL    string `json:"url,omitempty"`
	UA     string `json:"ua,omitempty"`
	EPG    string `json:"epg,omitempty"`
}

// HandleLive 直播源管理
// POST /api/admin/live
func (h *AdminLegacyHandler) HandleLive(c *gin.Context) {
	var req LiveSourceActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效"))
		return
	}

	ctx := c.Request.Context()
	sources, err := h.adminStorage.GetLiveSources(ctx)
	if err != nil {
		h.logger.Error("获取直播源失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "操作失败"))
		return
	}

	switch req.Action {
	case "add":
		if req.Key == "" || req.Name == "" || req.URL == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少必要参数"))
			return
		}
		for _, source := range sources {
			if source.Key == req.Key {
				c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "该直播源已存在"))
				return
			}
		}

		sources = append(sources, model.LiveSource{
			Key:      req.Key,
			Name:     req.Name,
			URL:      req.URL,
			UA:       req.UA,
			EPG:      req.EPG,
			Disabled: false,
			From:     "custom",
		})
		if err := h.adminStorage.SaveLiveSources(ctx, sources); err != nil {
			h.logger.Error("保存直播源失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "添加失败"))
			return
		}
		h.logger.Info("直播源已添加", zap.String("key", req.Key))

	case "delete":
		if req.Key == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少 key 参数"))
			return
		}
		for index, source := range sources {
			if source.Key != req.Key {
				continue
			}
			if source.From == "config" {
				c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "不能删除配置文件中的直播源"))
				return
			}

			sources = append(sources[:index], sources[index+1:]...)
			if err := h.adminStorage.SaveLiveSources(ctx, sources); err != nil {
				h.logger.Error("保存直播源失败", zap.Error(err))
				c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "删除失败"))
				return
			}
			h.logger.Info("直播源已删除", zap.String("key", req.Key))
			c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
			return
		}

		c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "直播源不存在"))
		return

	case "disable", "enable":
		if req.Key == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少 key 参数"))
			return
		}
		for index := range sources {
			if sources[index].Key != req.Key {
				continue
			}
			sources[index].Disabled = req.Action == "disable"
			if err := h.adminStorage.SaveLiveSources(ctx, sources); err != nil {
				h.logger.Error("保存直播源失败", zap.Error(err))
				c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "操作失败"))
				return
			}
			h.logger.Info("直播源状态已切换", zap.String("key", req.Key))
			c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
			return
		}

		c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "直播源不存在"))
		return

	case "edit":
		if req.Key == "" || req.Name == "" || req.URL == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少必要参数"))
			return
		}
		for index := range sources {
			if sources[index].Key != req.Key {
				continue
			}
			if sources[index].From == "config" {
				c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "不能编辑配置文件中的直播源"))
				return
			}

			sources[index].Name = req.Name
			sources[index].URL = req.URL
			sources[index].UA = req.UA
			sources[index].EPG = req.EPG
			if err := h.adminStorage.SaveLiveSources(ctx, sources); err != nil {
				h.logger.Error("保存直播源失败", zap.Error(err))
				c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "编辑失败"))
				return
			}
			h.logger.Info("直播源已编辑", zap.String("key", req.Key))
			c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
			return
		}

		c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "直播源不存在"))
		return

	default:
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "未知的操作类型"))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
}

// HandleLiveRefresh 刷新直播源
// POST /api/admin/live/refresh
func (h *AdminLegacyHandler) HandleLiveRefresh(c *gin.Context) {
	var req struct {
		Key string `json:"key"`
	}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效"))
			return
		}
	}

	count, err := h.refreshLiveSourceChannels(c.Request.Context(), req.Key)
	if err != nil {
		h.logger.Error("刷新直播源失败", zap.String("key", req.Key), zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "刷新直播源失败"))
		return
	}

	h.logger.Info("刷新直播源完成", zap.String("key", req.Key), zap.Int("channel_number", count))
	c.JSON(http.StatusOK, model.Success(gin.H{"ok": true, "channelNumber": count}))
}
