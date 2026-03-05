package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

type legacySiteConfig struct {
	SiteName                string `json:"SiteName"`
	Announcement            string `json:"Announcement"`
	SearchDownstreamMaxPage int    `json:"SearchDownstreamMaxPage"`
	SiteInterfaceCacheTime  int    `json:"SiteInterfaceCacheTime"`
	DoubanProxyType         string `json:"DoubanProxyType"`
	DoubanProxy             string `json:"DoubanProxy"`
	DoubanImageProxyType    string `json:"DoubanImageProxyType"`
	DoubanImageProxy        string `json:"DoubanImageProxy"`
	DisableYellowFilter     bool   `json:"DisableYellowFilter"`
	FluidSearch             bool   `json:"FluidSearch"`
}

type legacyConfigSubscription struct {
	URL        string `json:"URL"`
	AutoUpdate bool   `json:"AutoUpdate"`
	LastCheck  string `json:"LastCheck"`
}

// HandleConfigGet 兼容 GET /api/admin/config。
func (h *AdminLegacyHandler) HandleConfigGet(c *gin.Context) {
	if !h.adminHandler.requireAdmin(c) {
		return
	}

	config, err := h.adminStorage.GetAdminConfig(c.Request.Context())
	if err != nil {
		h.logger.Error("获取管理员配置失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取配置失败"))
		return
	}

	users, _ := h.adminStorage.GetAllUsers(c.Request.Context())
	hasOwner := false
	legacyUsers := make([]gin.H, 0, len(users)+1)
	for _, user := range users {
		if user.Username == h.adminHandler.ownerUser {
			hasOwner = true
		}
		legacyUsers = append(legacyUsers, gin.H{
			"username":    user.Username,
			"role":        user.Role,
			"banned":      user.Banned,
			"tags":        user.Tags,
			"enabledApis": user.EnabledAPIs,
		})
	}
	if !hasOwner {
		legacyUsers = append([]gin.H{{
			"username": h.adminHandler.ownerUser,
			"role":     "owner",
			"banned":   false,
		}}, legacyUsers...)
	}

	videoSources, _ := h.adminStorage.GetVideoSources(c.Request.Context())
	liveSources, _ := h.adminStorage.GetLiveSources(c.Request.Context())
	legacyGroups := make([]gin.H, 0, len(config.UserGroups))
	for _, group := range config.UserGroups {
		legacyGroups = append(legacyGroups, gin.H{
			"name":        group.Name,
			"enabledApis": group.EnabledAPIs,
		})
	}

	configPayload := gin.H{
		"SiteConfig": gin.H{
			"SiteName":                config.SiteConfig.SiteName,
			"Announcement":            config.SiteConfig.Announcement,
			"SearchDownstreamMaxPage": config.SiteConfig.SearchDownstreamMaxPage,
			"SiteInterfaceCacheTime":  config.SiteConfig.SiteInterfaceCacheTime,
			"DoubanProxyType":         config.SiteConfig.DoubanProxyType,
			"DoubanProxy":             config.SiteConfig.DoubanProxy,
			"DoubanImageProxyType":    config.SiteConfig.DoubanImageProxyType,
			"DoubanImageProxy":        config.SiteConfig.DoubanImageProxy,
			"DisableYellowFilter":     config.SiteConfig.DisableYellowFilter,
			"FluidSearch":             config.SiteConfig.FluidSearch,
		},
		"UserConfig": gin.H{
			"Users": legacyUsers,
			"Tags":  legacyGroups,
		},
		"DataSourceConfig": gin.H{
			"DataSources": videoSources,
		},
		"LiveConfig": gin.H{
			"DataSources": liveSources,
		},
		"CustomCategories": config.CustomCategories,
		"ConfigFile":       config.ConfigFile,
		"ConfigSubscribtion": gin.H{
			"URL":        config.ConfigSubscription.URL,
			"AutoUpdate": config.ConfigSubscription.AutoUpdate,
			"LastCheck":  config.ConfigSubscription.LastCheck,
		},
	}

	role := "admin"
	if h.adminHandler.isOwner(c.GetString("username")) {
		role = "owner"
	}

	response := gin.H{
		"Role":   role,
		"Config": configPayload,
	}
	for key, value := range configPayload {
		response[key] = value
	}

	c.JSON(http.StatusOK, response)
}

// HandleConfigPut 兼容 PUT /api/admin/config。
func (h *AdminLegacyHandler) HandleConfigPut(c *gin.Context) {
	if !h.adminHandler.requireAdmin(c) {
		return
	}

	var body map[string]json.RawMessage
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效"))
		return
	}

	if rawConfig, ok := body["Config"]; ok {
		var nested map[string]json.RawMessage
		if err := json.Unmarshal(rawConfig, &nested); err == nil {
			for key, value := range nested {
				if _, exists := body[key]; !exists {
					body[key] = value
				}
			}
		}
	}

	config, err := h.adminStorage.GetAdminConfig(c.Request.Context())
	if err != nil {
		h.logger.Error("获取管理员配置失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取配置失败"))
		return
	}

	if rawSiteConfig, ok := body["SiteConfig"]; ok {
		var payload legacySiteConfig
		if err := json.Unmarshal(rawSiteConfig, &payload); err != nil {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "站点配置格式无效"))
			return
		}
		config.SiteConfig = model.SiteConfig{
			SiteName:                payload.SiteName,
			Announcement:            payload.Announcement,
			SearchDownstreamMaxPage: payload.SearchDownstreamMaxPage,
			SiteInterfaceCacheTime:  payload.SiteInterfaceCacheTime,
			DoubanProxyType:         payload.DoubanProxyType,
			DoubanProxy:             payload.DoubanProxy,
			DoubanImageProxyType:    payload.DoubanImageProxyType,
			DoubanImageProxy:        payload.DoubanImageProxy,
			DisableYellowFilter:     payload.DisableYellowFilter,
			FluidSearch:             payload.FluidSearch,
		}
	}

	if rawConfigFile, ok := body["ConfigFile"]; ok {
		var configFile string
		if err := json.Unmarshal(rawConfigFile, &configFile); err != nil {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "ConfigFile 格式无效"))
			return
		}
		config.ConfigFile = configFile
	}

	for _, key := range []string{"ConfigSubscribtion", "ConfigSubscription"} {
		rawSub, ok := body[key]
		if !ok {
			continue
		}
		var sub legacyConfigSubscription
		if err := json.Unmarshal(rawSub, &sub); err != nil {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "订阅配置格式无效"))
			return
		}
		config.ConfigSubscription.URL = strings.TrimSpace(sub.URL)
		config.ConfigSubscription.AutoUpdate = sub.AutoUpdate
		config.ConfigSubscription.LastCheck = strings.TrimSpace(sub.LastCheck)
	}

	if err := h.adminStorage.SaveAdminConfig(c.Request.Context(), config); err != nil {
		h.logger.Error("保存管理员配置失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "保存配置失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
}

// HandleUsersGet 兼容 GET /api/admin/users。
func (h *AdminLegacyHandler) HandleUsersGet(c *gin.Context) {
	h.adminHandler.GetUsers(c)
}

// HandleUsersCreate 兼容 POST /api/admin/users。
func (h *AdminLegacyHandler) HandleUsersCreate(c *gin.Context) {
	h.adminHandler.CreateUser(c)
}

// HandleUsersUpdate 兼容 PUT /api/admin/users/:username。
func (h *AdminLegacyHandler) HandleUsersUpdate(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求体读取失败"))
		return
	}

	if len(body) == 0 {
		h.adminHandler.UpdateUser(c)
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		h.adminHandler.UpdateUser(c)
		return
	}

	if rawPassword, ok := payload["password"]; ok {
		password, _ := rawPassword.(string)
		if strings.TrimSpace(password) != "" {
			rewritten, _ := json.Marshal(gin.H{"password": password})
			c.Request.Body = io.NopCloser(bytes.NewReader(rewritten))
			c.Request.ContentLength = int64(len(rewritten))
			h.adminHandler.ChangeUserPassword(c)
			return
		}
	}

	if enabledApis, ok := payload["enabledApis"]; ok {
		if _, exists := payload["enabled_apis"]; !exists {
			payload["enabled_apis"] = enabledApis
		}
	}

	rewritten, _ := json.Marshal(payload)
	c.Request.Body = io.NopCloser(bytes.NewReader(rewritten))
	c.Request.ContentLength = int64(len(rewritten))
	h.adminHandler.UpdateUser(c)
}

// HandleUsersDelete 兼容 DELETE /api/admin/users/:username。
func (h *AdminLegacyHandler) HandleUsersDelete(c *gin.Context) {
	h.adminHandler.DeleteUser(c)
}

// HandleSitesGet 兼容 GET /api/admin/sites。
func (h *AdminLegacyHandler) HandleSitesGet(c *gin.Context) {
	h.adminHandler.GetSites(c)
}

// HandleSitesUpdate 兼容 PUT /api/admin/sites/:key。
func (h *AdminLegacyHandler) HandleSitesUpdate(c *gin.Context) {
	h.adminHandler.UpdateSite(c)
}

// HandleSitesDelete 兼容 DELETE /api/admin/sites/:key。
func (h *AdminLegacyHandler) HandleSitesDelete(c *gin.Context) {
	h.adminHandler.DeleteSite(c)
}
