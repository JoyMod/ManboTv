// internal/handler/skip_config_handler.go
package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// SkipConfigHandler 跳过配置处理器
type SkipConfigHandler struct {
	storage model.StorageService
	logger  *zap.Logger
}

// NewSkipConfigHandler 创建跳过配置处理器
func NewSkipConfigHandler(storage model.StorageService, logger *zap.Logger) *SkipConfigHandler {
	return &SkipConfigHandler{
		storage: storage,
		logger:  logger,
	}
}

// GetConfig 获取跳过配置
// GET /api/v1/skipconfigs?source=xxx&id=xxx
func (h *SkipConfigHandler) GetConfig(c *gin.Context) {
	username := c.GetString("username")
	if username == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeUnauthorized, "未登录"))
		return
	}

	source := c.Query("source")
	id := c.Query("id")

	if source != "" && id != "" {
		// 获取单个配置
		config, err := h.storage.GetSkipConfig(c.Request.Context(), username, source, id)
		if err != nil {
			h.logger.Error("获取跳过配置失败",
				zap.String("username", username),
				zap.String("source", source),
				zap.String("id", id),
				zap.Error(err),
			)
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取配置失败"))
			return
		}
		if config == nil {
			config = &model.SkipConfig{
				Enable:    false,
				IntroTime: 0,
				OutroTime: 0,
			}
		}
		c.JSON(http.StatusOK, model.Success(config))
		return
	}

	// 获取所有配置
	configs, err := h.storage.GetAllSkipConfigs(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("获取所有跳过配置失败",
			zap.String("username", username),
			zap.Error(err),
		)
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取配置失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(configs))
}

// SetConfig 设置跳过配置
// POST /api/v1/skipconfigs
func (h *SkipConfigHandler) SetConfig(c *gin.Context) {
	username := c.GetString("username")
	if username == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeUnauthorized, "未登录"))
		return
	}

	var req struct {
		Key    string           `json:"key" binding:"required"`
		Config model.SkipConfig `json:"config" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效"))
		return
	}

	// 解析 key (source+id)
	parts := strings.Split(req.Key, "+")
	if len(parts) != 2 {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "无效的key格式"))
		return
	}
	source := parts[0]
	id := parts[1]

	// 保存配置
	if err := h.storage.SetSkipConfig(c.Request.Context(), username, source, id, &req.Config); err != nil {
		h.logger.Error("保存跳过配置失败",
			zap.String("username", username),
			zap.String("source", source),
			zap.String("id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "保存配置失败"))
		return
	}

	h.logger.Debug("跳过配置已保存",
		zap.String("username", username),
		zap.String("source", source),
		zap.String("id", id),
		zap.Bool("enable", req.Config.Enable),
	)

	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}

// DeleteConfig 删除跳过配置
// DELETE /api/v1/skipconfigs?key=xxx
func (h *SkipConfigHandler) DeleteConfig(c *gin.Context) {
	username := c.GetString("username")
	if username == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeUnauthorized, "未登录"))
		return
	}

	key := c.Query("key")
	if key == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少key参数"))
		return
	}

	// 解析 key (source+id)
	parts := strings.Split(key, "+")
	if len(parts) != 2 {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "无效的key格式"))
		return
	}
	source := parts[0]
	id := parts[1]

	// 删除配置
	if err := h.storage.DeleteSkipConfig(c.Request.Context(), username, source, id); err != nil {
		h.logger.Error("删除跳过配置失败",
			zap.String("username", username),
			zap.String("source", source),
			zap.String("id", id),
			zap.Error(err),
		)
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "删除配置失败"))
		return
	}

	h.logger.Debug("跳过配置已删除",
		zap.String("username", username),
		zap.String("source", source),
		zap.String("id", id),
	)

	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}

// GetConfigLegacy handles GET /api/skipconfigs.
func (h *SkipConfigHandler) GetConfigLegacy(c *gin.Context) {
	username := resolveUsernameFromContext(c)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	source := strings.TrimSpace(c.Query("source"))
	id := strings.TrimSpace(c.Query("id"))

	if source != "" && id != "" {
		config, err := h.storage.GetSkipConfig(c.Request.Context(), username, source, id)
		if err != nil {
			h.logger.Error("获取跳过配置失败", zap.String("username", username), zap.String("source", source), zap.String("id", id), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取跳过配置失败"})
			return
		}
		if config == nil {
			config = &model.SkipConfig{Enable: false, IntroTime: 0, OutroTime: 0}
		}
		c.JSON(http.StatusOK, config)
		return
	}

	configs, err := h.storage.GetAllSkipConfigs(c.Request.Context(), username)
	if err != nil {
		h.logger.Error("获取所有跳过配置失败", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取跳过配置失败"})
		return
	}
	if configs == nil {
		configs = map[string]*model.SkipConfig{}
	}

	c.JSON(http.StatusOK, configs)
}

// SetConfigLegacy handles POST /api/skipconfigs.
func (h *SkipConfigHandler) SetConfigLegacy(c *gin.Context) {
	username := resolveUsernameFromContext(c)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数"})
		return
	}

	key := strValue(body, "key")
	source, id := splitLegacyKey(key)
	if source == "" || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的key格式"})
		return
	}

	configData, _ := body["config"].(map[string]interface{})
	if configData == nil {
		configData = body
	}

	config := &model.SkipConfig{
		Enable:    boolFromValue(configData["enable"]),
		IntroTime: intValue(configData, "intro_time"),
		OutroTime: intValue(configData, "outro_time"),
	}

	if err := h.storage.SetSkipConfig(c.Request.Context(), username, source, id, config); err != nil {
		h.logger.Error("保存跳过配置失败", zap.String("username", username), zap.String("source", source), zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存跳过配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteConfigLegacy handles DELETE /api/skipconfigs.
func (h *SkipConfigHandler) DeleteConfigLegacy(c *gin.Context) {
	username := resolveUsernameFromContext(c)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	key := strings.TrimSpace(c.Query("key"))
	if key == "" {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err == nil {
			key = strValue(body, "key")
		}
	}

	source, id := splitLegacyKey(key)
	if source == "" || id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数"})
		return
	}

	if err := h.storage.DeleteSkipConfig(c.Request.Context(), username, source, id); err != nil {
		h.logger.Error("删除跳过配置失败", zap.String("username", username), zap.String("source", source), zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除跳过配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func boolFromValue(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case float64:
		return val != 0
	case int:
		return val != 0
	case int64:
		return val != 0
	case string:
		trimmed := strings.TrimSpace(strings.ToLower(val))
		return trimmed == "1" || trimmed == "true" || trimmed == "yes" || trimmed == "on"
	default:
		return false
	}
}
