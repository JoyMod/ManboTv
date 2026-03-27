package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// FetchSubscriptionRequest 订阅请求
type FetchSubscriptionRequest struct {
	URL string `json:"url" binding:"required"`
}

// ConfigFileRequest 配置文件请求
type ConfigFileRequest struct {
	ConfigFile      string `json:"configFile" binding:"required"`
	SubscriptionURL string `json:"subscriptionUrl,omitempty"`
	AutoUpdate      bool   `json:"autoUpdate,omitempty"`
	LastCheckTime   string `json:"lastCheckTime,omitempty"`
}

// HandleDataExport 数据导出
// GET /api/admin/data_migration/export
func (h *AdminLegacyHandler) HandleDataExport(c *gin.Context) {
	h.adminHandler.ExportData(c)
}

// HandleDataImport 数据导入
// POST /api/admin/data_migration/import
func (h *AdminLegacyHandler) HandleDataImport(c *gin.Context) {
	h.adminHandler.ImportData(c)
}

// HandleFetchSubscription 获取订阅
// POST /api/admin/config_subscription/fetch
func (h *AdminLegacyHandler) HandleFetchSubscription(c *gin.Context) {
	var req FetchSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少订阅URL"))
		return
	}

	decoded, err := h.fetchAndDecodeSubscription(c.Request.Context(), req.URL)
	if err != nil {
		h.logger.Error("获取订阅失败", zap.String("url", req.URL), zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "拉取配置失败"))
		return
	}

	h.logger.Info("获取订阅成功", zap.String("url", req.URL))
	c.JSON(http.StatusOK, model.Success(gin.H{
		"ok":            true,
		"configContent": decoded,
		"message":       "配置拉取成功",
	}))
}

// HandleConfigFile 配置文件保存
// POST /api/admin/config_file
func (h *AdminLegacyHandler) HandleConfigFile(c *gin.Context) {
	if !h.adminHandler.requireOwner(c) {
		return
	}

	var req ConfigFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少配置数据"))
		return
	}

	if !json.Valid([]byte(req.ConfigFile)) {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "配置文件格式错误，请检查 JSON 语法"))
		return
	}

	config, err := h.adminStorage.GetAdminConfig(c.Request.Context())
	if err != nil {
		h.logger.Error("获取配置失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取配置失败"))
		return
	}

	config.ConfigFile = req.ConfigFile
	config.ConfigSubscription.URL = req.SubscriptionURL
	config.ConfigSubscription.AutoUpdate = req.AutoUpdate
	config.ConfigSubscription.LastCheck = req.LastCheckTime

	if err := h.adminStorage.SaveAdminConfig(c.Request.Context(), config); err != nil {
		h.logger.Error("保存配置失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "保存失败"))
		return
	}

	h.logger.Info("配置文件已保存")
	c.JSON(http.StatusOK, model.Success(gin.H{
		"success": true,
		"message": "配置文件更新成功",
	}))
}
