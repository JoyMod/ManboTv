package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// SourceActionRequest 资源站操作请求
type SourceActionRequest struct {
	Action string   `json:"action" binding:"required"`
	Key    string   `json:"key,omitempty"`
	Name   string   `json:"name,omitempty"`
	API    string   `json:"api,omitempty"`
	Detail string   `json:"detail,omitempty"`
	Keys   []string `json:"keys,omitempty"`
	Orders []string `json:"orders,omitempty"`
}

// HandleSource 资源站管理
// POST /api/admin/source
func (h *AdminLegacyHandler) HandleSource(c *gin.Context) {
	var req SourceActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效"))
		return
	}

	sources, err := h.adminStorage.GetVideoSources(c.Request.Context())
	if err != nil {
		h.logger.Error("获取视频源失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "操作失败"))
		return
	}

	switch req.Action {
	case "add":
		h.handleSourceAdd(c, &req, sources)
	case "disable", "enable":
		h.handleSourceToggle(c, &req, sources)
	case "delete":
		h.handleSourceDelete(c, &req, sources)
	case "sort":
		h.handleSourceSort(c, &req, sources)
	case "batch_disable", "batch_enable", "batch_delete":
		h.handleSourceBatch(c, &req, sources)
	default:
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "未知的操作类型"))
	}
}

func (h *AdminLegacyHandler) handleSourceAdd(
	c *gin.Context,
	req *SourceActionRequest,
	sources []model.VideoSource,
) {
	if req.Key == "" || req.Name == "" || req.API == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少必要参数"))
		return
	}

	for _, source := range sources {
		if source.Key == req.Key {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "该源已存在"))
			return
		}
	}

	sources = append(sources, model.VideoSource{
		Key:      req.Key,
		Name:     req.Name,
		API:      req.API,
		Detail:   req.Detail,
		Disabled: false,
		From:     "custom",
	})

	if err := h.adminStorage.SaveVideoSources(c.Request.Context(), sources); err != nil {
		h.logger.Error("保存视频源失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "保存失败"))
		return
	}

	h.logger.Info("视频源已添加", zap.String("key", req.Key))
	c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
}

func (h *AdminLegacyHandler) handleSourceToggle(
	c *gin.Context,
	req *SourceActionRequest,
	sources []model.VideoSource,
) {
	for index := range sources {
		if sources[index].Key != req.Key {
			continue
		}
		sources[index].Disabled = req.Action == "disable"
		if err := h.adminStorage.SaveVideoSources(c.Request.Context(), sources); err != nil {
			h.logger.Error("保存视频源失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "保存失败"))
			return
		}
		h.logger.Info("视频源状态已切换", zap.String("key", req.Key), zap.String("action", req.Action))
		c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
		return
	}

	c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "源不存在"))
}

func (h *AdminLegacyHandler) handleSourceDelete(
	c *gin.Context,
	req *SourceActionRequest,
	sources []model.VideoSource,
) {
	for index, source := range sources {
		if source.Key != req.Key {
			continue
		}
		if source.From == "config" {
			c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "不能删除配置文件中的源"))
			return
		}

		sources = append(sources[:index], sources[index+1:]...)
		if err := h.adminStorage.SaveVideoSources(c.Request.Context(), sources); err != nil {
			h.logger.Error("保存视频源失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "删除失败"))
			return
		}

		h.logger.Info("视频源已删除", zap.String("key", req.Key))
		c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
		return
	}

	c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "源不存在"))
}

func (h *AdminLegacyHandler) handleSourceSort(
	c *gin.Context,
	req *SourceActionRequest,
	sources []model.VideoSource,
) {
	if len(req.Orders) == 0 {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "排序列表为空"))
		return
	}

	newSources := make([]model.VideoSource, 0, len(sources))
	sourceMap := make(map[string]model.VideoSource, len(sources))
	for _, source := range sources {
		sourceMap[source.Key] = source
	}

	for _, key := range req.Orders {
		source, ok := sourceMap[key]
		if !ok {
			continue
		}
		newSources = append(newSources, source)
		delete(sourceMap, key)
	}

	for _, source := range sources {
		if _, ok := sourceMap[source.Key]; !ok {
			continue
		}
		newSources = append(newSources, source)
	}

	if err := h.adminStorage.SaveVideoSources(c.Request.Context(), newSources); err != nil {
		h.logger.Error("保存视频源失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "排序失败"))
		return
	}

	h.logger.Info("视频源已排序")
	c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
}

func (h *AdminLegacyHandler) handleSourceBatch(
	c *gin.Context,
	req *SourceActionRequest,
	sources []model.VideoSource,
) {
	if len(req.Keys) == 0 {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "未选择任何源"))
		return
	}

	switch req.Action {
	case "batch_disable", "batch_enable":
		disabled := req.Action == "batch_disable"
		for index := range sources {
			for _, key := range req.Keys {
				if sources[index].Key != key {
					continue
				}
				sources[index].Disabled = disabled
				break
			}
		}
	case "batch_delete":
		nextSources := make([]model.VideoSource, 0, len(sources))
		for _, source := range sources {
			shouldDelete := false
			for _, key := range req.Keys {
				if source.Key != key {
					continue
				}
				shouldDelete = true
				break
			}
			if !shouldDelete || source.From == "config" {
				nextSources = append(nextSources, source)
			}
		}
		sources = nextSources
	}

	if err := h.adminStorage.SaveVideoSources(c.Request.Context(), sources); err != nil {
		h.logger.Error("保存视频源失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "批量操作失败"))
		return
	}

	h.logger.Info("视频源批量操作完成", zap.String("action", req.Action))
	c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
}
