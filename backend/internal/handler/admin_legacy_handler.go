// internal/handler/admin_legacy_handler.go
// Admin 旧版 API 兼容层 - 供前端现有代码调用

package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// AdminLegacyHandler 旧版 Admin API 处理器
type AdminLegacyHandler struct {
	adminHandler *AdminHandler
	adminStorage model.AdminStorageService
	logger       *zap.Logger
}

// NewAdminLegacyHandler 创建旧版 Admin 处理器
func NewAdminLegacyHandler(
	adminHandler *AdminHandler,
	adminStorage model.AdminStorageService,
	logger *zap.Logger,
) *AdminLegacyHandler {
	return &AdminLegacyHandler{
		adminHandler: adminHandler,
		adminStorage: adminStorage,
		logger:       logger,
	}
}

// ========== 视频源管理 /api/admin/source ==========

// SourceActionRequest 资源站操作请求
type SourceActionRequest struct {
	Action string   `json:"action" binding:"required"`
	Key    string   `json:"key,omitempty"`
	Name   string   `json:"name,omitempty"`
	API    string   `json:"api,omitempty"`
	Detail string   `json:"detail,omitempty"`
	Keys   []string `json:"keys,omitempty"`   // 批量操作
	Orders []string `json:"orders,omitempty"` // 排序
}

// HandleSource 资源站管理
// POST /api/admin/source
func (h *AdminLegacyHandler) HandleSource(c *gin.Context) {
	var req SourceActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效"))
		return
	}

	ctx := c.Request.Context()
	sources, err := h.adminStorage.GetVideoSources(ctx)
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

func (h *AdminLegacyHandler) handleSourceAdd(c *gin.Context, req *SourceActionRequest, sources []model.VideoSource) {
	if req.Key == "" || req.Name == "" || req.API == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少必要参数"))
		return
	}

	// 检查是否已存在
	for _, s := range sources {
		if s.Key == req.Key {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "该源已存在"))
			return
		}
	}

	// 添加新源
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

func (h *AdminLegacyHandler) handleSourceToggle(c *gin.Context, req *SourceActionRequest, sources []model.VideoSource) {
	for i := range sources {
		if sources[i].Key == req.Key {
			sources[i].Disabled = (req.Action == "disable")
			if err := h.adminStorage.SaveVideoSources(c.Request.Context(), sources); err != nil {
				h.logger.Error("保存视频源失败", zap.Error(err))
				c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "保存失败"))
				return
			}
			h.logger.Info("视频源状态已切换", zap.String("key", req.Key), zap.String("action", req.Action))
			c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
			return
		}
	}
	c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "源不存在"))
}

func (h *AdminLegacyHandler) handleSourceDelete(c *gin.Context, req *SourceActionRequest, sources []model.VideoSource) {
	for i, s := range sources {
		if s.Key == req.Key {
			// 检查是否为配置文件的源
			if s.From == "config" {
				c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "不能删除配置文件中的源"))
				return
			}
			// 删除
			sources = append(sources[:i], sources[i+1:]...)
			if err := h.adminStorage.SaveVideoSources(c.Request.Context(), sources); err != nil {
				h.logger.Error("保存视频源失败", zap.Error(err))
				c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "删除失败"))
				return
			}
			h.logger.Info("视频源已删除", zap.String("key", req.Key))
			c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
			return
		}
	}
	c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "源不存在"))
}

func (h *AdminLegacyHandler) handleSourceSort(c *gin.Context, req *SourceActionRequest, sources []model.VideoSource) {
	if len(req.Orders) == 0 {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "排序列表为空"))
		return
	}

	// 创建新排序的列表
	newSources := make([]model.VideoSource, 0, len(sources))
	sourceMap := make(map[string]model.VideoSource)
	for _, s := range sources {
		sourceMap[s.Key] = s
	}

	// 按顺序添加
	for _, key := range req.Orders {
		if s, ok := sourceMap[key]; ok {
			newSources = append(newSources, s)
			delete(sourceMap, key)
		}
	}

	// 添加剩余的（未在排序列表中的）
	for _, s := range sources {
		if _, ok := sourceMap[s.Key]; ok {
			newSources = append(newSources, s)
		}
	}

	if err := h.adminStorage.SaveVideoSources(c.Request.Context(), newSources); err != nil {
		h.logger.Error("保存视频源失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "排序失败"))
		return
	}

	h.logger.Info("视频源已排序")
	c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
}

func (h *AdminLegacyHandler) handleSourceBatch(c *gin.Context, req *SourceActionRequest, sources []model.VideoSource) {
	if len(req.Keys) == 0 {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "未选择任何源"))
		return
	}

	switch req.Action {
	case "batch_disable", "batch_enable":
		disabled := (req.Action == "batch_disable")
		for i := range sources {
			for _, key := range req.Keys {
				if sources[i].Key == key {
					sources[i].Disabled = disabled
					break
				}
			}
		}
	case "batch_delete":
		// 过滤掉 config 来源的
		newSources := make([]model.VideoSource, 0)
		for _, s := range sources {
			shouldDelete := false
			for _, key := range req.Keys {
				if s.Key == key {
					shouldDelete = true
					break
				}
			}
			if !shouldDelete || s.From == "config" {
				newSources = append(newSources, s)
			}
		}
		sources = newSources
	}

	if err := h.adminStorage.SaveVideoSources(c.Request.Context(), sources); err != nil {
		h.logger.Error("保存视频源失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "批量操作失败"))
		return
	}

	h.logger.Info("视频源批量操作完成", zap.String("action", req.Action))
	c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
}

// ========== 分类管理 /api/admin/category ==========

// CategoryActionRequest 分类操作请求
type CategoryActionRequest struct {
	Action string `json:"action" binding:"required"`
	Name   string `json:"name,omitempty"`
	Type   string `json:"type,omitempty"` // movie/tv
	Query  string `json:"query,omitempty"`
}

// HandleCategory 分类管理
// POST /api/admin/category
func (h *AdminLegacyHandler) HandleCategory(c *gin.Context) {
	var req CategoryActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效"))
		return
	}

	ctx := c.Request.Context()
	categories, err := h.adminStorage.GetCustomCategories(ctx)
	if err != nil {
		h.logger.Error("获取分类失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "操作失败"))
		return
	}

	switch req.Action {
	case "add":
		if req.Name == "" || req.Type == "" || req.Query == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少必要参数"))
			return
		}
		// 检查是否已存在相同的 query 和 type
		for _, cat := range categories {
			if cat.Query == req.Query && cat.Type == req.Type {
				c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "该分类已存在"))
				return
			}
		}
		categories = append(categories, model.CustomCategory{
			Name:     req.Name,
			Type:     req.Type,
			Query:    req.Query,
			Disabled: false,
			From:     "custom",
		})
		if err := h.adminStorage.SaveCustomCategories(ctx, categories); err != nil {
			h.logger.Error("保存分类失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "添加失败"))
			return
		}
		h.logger.Info("分类已添加", zap.String("name", req.Name))

	case "delete":
		if req.Query == "" || req.Type == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少 query 或 type 参数"))
			return
		}
		for i, cat := range categories {
			if cat.Query == req.Query && cat.Type == req.Type {
				if cat.From == "config" {
					c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "不能删除配置文件中的分类"))
					return
				}
				categories = append(categories[:i], categories[i+1:]...)
				if err := h.adminStorage.SaveCustomCategories(ctx, categories); err != nil {
					h.logger.Error("保存分类失败", zap.Error(err))
					c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "删除失败"))
					return
				}
				h.logger.Info("分类已删除", zap.String("name", cat.Name))
				c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
				return
			}
		}
		c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "分类不存在"))
		return

	case "disable", "enable":
		if req.Query == "" || req.Type == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少 query 或 type 参数"))
			return
		}
		for i := range categories {
			if categories[i].Query == req.Query && categories[i].Type == req.Type {
				categories[i].Disabled = (req.Action == "disable")
				if err := h.adminStorage.SaveCustomCategories(ctx, categories); err != nil {
					h.logger.Error("保存分类失败", zap.Error(err))
					c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "操作失败"))
					return
				}
				h.logger.Info("分类状态已切换", zap.String("name", categories[i].Name))
				c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
				return
			}
		}
		c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "分类不存在"))
		return

	default:
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "未知的操作类型"))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
}

// ========== 站点配置 /api/admin/site ==========

// SiteConfigRequest 站点配置请求
type SiteConfigRequest struct {
	SiteConfig model.SiteConfig `json:"site_config" binding:"required"`
}

// HandleSite 站点配置
// POST /api/admin/site
func (h *AdminLegacyHandler) HandleSite(c *gin.Context) {
	var req SiteConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效"))
		return
	}

	config, err := h.adminStorage.GetAdminConfig(c.Request.Context())
	if err != nil {
		h.logger.Error("获取配置失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取配置失败"))
		return
	}

	config.SiteConfig = req.SiteConfig
	if err := h.adminStorage.SaveAdminConfig(c.Request.Context(), config); err != nil {
		h.logger.Error("保存配置失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "保存失败"))
		return
	}

	h.logger.Info("站点配置已更新")
	c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
}

// HandleGetSiteConfig 获取站点配置
// GET /api/admin/site
func (h *AdminLegacyHandler) HandleGetSiteConfig(c *gin.Context) {
	config, err := h.adminStorage.GetAdminConfig(c.Request.Context())
	if err != nil {
		h.logger.Error("获取配置失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取配置失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(config.SiteConfig))
}

// ========== 直播源管理 /api/admin/live ==========

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
		// 检查是否已存在
		for _, s := range sources {
			if s.Key == req.Key {
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
		for i, s := range sources {
			if s.Key == req.Key {
				if s.From == "config" {
					c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "不能删除配置文件中的直播源"))
					return
				}
				sources = append(sources[:i], sources[i+1:]...)
				if err := h.adminStorage.SaveLiveSources(ctx, sources); err != nil {
					h.logger.Error("保存直播源失败", zap.Error(err))
					c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "删除失败"))
					return
				}
				h.logger.Info("直播源已删除", zap.String("key", req.Key))
				c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
				return
			}
		}
		c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "直播源不存在"))
		return

	case "disable", "enable":
		if req.Key == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少 key 参数"))
			return
		}
		for i := range sources {
			if sources[i].Key == req.Key {
				sources[i].Disabled = (req.Action == "disable")
				if err := h.adminStorage.SaveLiveSources(ctx, sources); err != nil {
					h.logger.Error("保存直播源失败", zap.Error(err))
					c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "操作失败"))
					return
				}
				h.logger.Info("直播源状态已切换", zap.String("key", req.Key))
				c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
				return
			}
		}
		c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "直播源不存在"))
		return

	case "edit":
		if req.Key == "" || req.Name == "" || req.URL == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少必要参数"))
			return
		}
		for i := range sources {
			if sources[i].Key == req.Key {
				if sources[i].From == "config" {
					c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "不能编辑配置文件中的直播源"))
					return
				}
				sources[i].Name = req.Name
				sources[i].URL = req.URL
				sources[i].UA = req.UA
				sources[i].EPG = req.EPG
				if err := h.adminStorage.SaveLiveSources(ctx, sources); err != nil {
					h.logger.Error("保存直播源失败", zap.Error(err))
					c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "编辑失败"))
					return
				}
				h.logger.Info("直播源已编辑", zap.String("key", req.Key))
				c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
				return
			}
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

// ========== 用户管理 /api/admin/user ==========

// UserActionRequest 用户操作请求
type UserActionRequest struct {
	Action      string   `json:"action" binding:"required"`
	Username    string   `json:"username,omitempty"`
	Password    string   `json:"password,omitempty"`
	TargetUser  string   `json:"targetUsername,omitempty"`
	TargetPass  string   `json:"targetPassword,omitempty"`
	Role        string   `json:"role,omitempty"`
	UserGroup   string   `json:"userGroup,omitempty"`
	UserGroups  []string `json:"userGroups,omitempty"`
	EnabledAPIs []string `json:"enabledApis,omitempty"`
	GroupName   string   `json:"groupName,omitempty"`
	GroupAction string   `json:"groupAction,omitempty"` // add/edit/delete
	Usernames   []string `json:"usernames,omitempty"`   // 批量操作用
}

// HandleUser 用户管理
// POST /api/admin/user
func (h *AdminLegacyHandler) HandleUser(c *gin.Context) {
	operator := c.GetString("username")

	var req UserActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效"))
		return
	}

	if strings.TrimSpace(req.Username) == "" {
		req.Username = strings.TrimSpace(req.TargetUser)
	}
	if strings.TrimSpace(req.Password) == "" {
		req.Password = strings.TrimSpace(req.TargetPass)
	}
	if req.Action == "deleteUser" {
		req.Action = "delete"
	}

	ctx := c.Request.Context()

	switch req.Action {
	case "add":
		if req.Username == "" || req.Password == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "用户名和密码不能为空"))
			return
		}
		if req.Role == "" {
			req.Role = "user"
		}

		// 权限检查
		if req.Role == "admin" && !h.adminHandler.isOwner(operator) {
			c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "仅站长可创建管理员"))
			return
		}

		// 检查用户是否已存在
		existing, err := h.adminStorage.GetUser(ctx, req.Username)
		if err != nil {
			h.logger.Error("检查用户失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "检查用户失败"))
			return
		}
		if existing != nil {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "用户已存在"))
			return
		}

		// 创建用户
		user := &model.UserConfig{
			Username:     req.Username,
			PasswordHash: hashPassword(req.Password),
			Role:         req.Role,
			CreatedAt:    time.Now().Unix(),
		}

		if req.UserGroup != "" {
			user.Tags = []string{req.UserGroup}
		}

		if err := h.adminStorage.CreateUser(ctx, user); err != nil {
			h.logger.Error("创建用户失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "创建用户失败"))
			return
		}

		h.logger.Info("用户已创建", zap.String("operator", operator), zap.String("username", req.Username))
		c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))

	case "delete":
		if req.Username == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少用户名"))
			return
		}
		if req.Username == operator {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "不能删除自己"))
			return
		}

		target, err := h.adminStorage.GetUser(ctx, req.Username)
		if err != nil {
			h.logger.Error("获取用户失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取用户失败"))
			return
		}
		if target == nil {
			c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "用户不存在"))
			return
		}

		// 权限检查
		if target.Role == "admin" && !h.adminHandler.isOwner(operator) {
			c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "仅站长可删除管理员"))
			return
		}

		if err := h.adminStorage.DeleteUser(ctx, req.Username); err != nil {
			h.logger.Error("删除用户失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "删除失败"))
			return
		}

		h.logger.Info("用户已删除", zap.String("operator", operator), zap.String("username", req.Username))
		c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))

	case "ban", "unban":
		if req.Username == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少用户名"))
			return
		}

		target, err := h.adminStorage.GetUser(ctx, req.Username)
		if err != nil {
			h.logger.Error("获取用户失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取用户失败"))
			return
		}
		if target == nil {
			c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "用户不存在"))
			return
		}

		// 权限检查
		if target.Role == "admin" && !h.adminHandler.isOwner(operator) {
			c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "仅站长可操作管理员"))
			return
		}

		target.Banned = (req.Action == "ban")
		if err := h.adminStorage.UpdateUser(ctx, req.Username, target); err != nil {
			h.logger.Error("更新用户失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "操作失败"))
			return
		}

		h.logger.Info("用户状态已变更", zap.String("operator", operator), zap.String("username", req.Username), zap.String("action", req.Action))
		c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))

	case "setAdmin":
		if req.Username == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少用户名"))
			return
		}
		if !h.adminHandler.isOwner(operator) {
			c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "仅站长可设置管理员"))
			return
		}

		target, err := h.adminStorage.GetUser(ctx, req.Username)
		if err != nil {
			h.logger.Error("获取用户失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取用户失败"))
			return
		}
		if target == nil {
			c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "用户不存在"))
			return
		}

		target.Role = "admin"
		if err := h.adminStorage.UpdateUser(ctx, req.Username, target); err != nil {
			h.logger.Error("更新用户失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "设置失败"))
			return
		}

		h.logger.Info("用户已设为管理员", zap.String("operator", operator), zap.String("username", req.Username))
		c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))

	case "cancelAdmin":
		if req.Username == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少用户名"))
			return
		}
		if !h.adminHandler.isOwner(operator) {
			c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "仅站长可取消管理员"))
			return
		}

		target, err := h.adminStorage.GetUser(ctx, req.Username)
		if err != nil {
			h.logger.Error("获取用户失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取用户失败"))
			return
		}
		if target == nil {
			c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "用户不存在"))
			return
		}

		target.Role = "user"
		if err := h.adminStorage.UpdateUser(ctx, req.Username, target); err != nil {
			h.logger.Error("更新用户失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "取消失败"))
			return
		}

		h.logger.Info("管理员权限已取消", zap.String("operator", operator), zap.String("username", req.Username))
		c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))

	case "changePassword":
		if req.Username == "" || req.Password == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "用户名和密码不能为空"))
			return
		}

		target, err := h.adminStorage.GetUser(ctx, req.Username)
		if err != nil {
			h.logger.Error("获取用户失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取用户失败"))
			return
		}
		if target == nil {
			c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "用户不存在"))
			return
		}

		// 权限检查
		if operator != req.Username {
			if target.Role == "admin" && !h.adminHandler.isOwner(operator) {
				c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "仅站长可修改其他管理员密码"))
				return
			}
		}

		if err := h.adminStorage.ChangePassword(ctx, req.Username, hashPassword(req.Password)); err != nil {
			h.logger.Error("修改密码失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "修改密码失败"))
			return
		}

		h.logger.Info("密码已修改", zap.String("operator", operator), zap.String("username", req.Username))
		c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))

	case "updateUserApis":
		if req.Username == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少用户名"))
			return
		}

		target, err := h.adminStorage.GetUser(ctx, req.Username)
		if err != nil {
			h.logger.Error("获取用户失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取用户失败"))
			return
		}
		if target == nil {
			c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "用户不存在"))
			return
		}

		target.EnabledAPIs = req.EnabledAPIs
		if err := h.adminStorage.UpdateUser(ctx, req.Username, target); err != nil {
			h.logger.Error("更新用户失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "更新失败"))
			return
		}

		h.logger.Info("用户API权限已更新", zap.String("operator", operator), zap.String("username", req.Username))
		c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))

	case "updateUserGroups":
		if req.Username == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少用户名"))
			return
		}

		target, err := h.adminStorage.GetUser(ctx, req.Username)
		if err != nil {
			h.logger.Error("获取用户失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取用户失败"))
			return
		}
		if target == nil {
			c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "用户不存在"))
			return
		}

		target.Tags = req.UserGroups
		if err := h.adminStorage.UpdateUser(ctx, req.Username, target); err != nil {
			h.logger.Error("更新用户失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "更新失败"))
			return
		}

		h.logger.Info("用户组已更新", zap.String("operator", operator), zap.String("username", req.Username))
		c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))

	case "batchUpdateUserGroups":
		if len(req.Usernames) == 0 {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "未选择任何用户"))
			return
		}

		// 批量更新
		for _, username := range req.Usernames {
			target, err := h.adminStorage.GetUser(ctx, username)
			if err != nil || target == nil {
				continue
			}
			target.Tags = req.UserGroups
			h.adminStorage.UpdateUser(ctx, username, target)
		}

		h.logger.Info("用户组批量更新", zap.String("operator", operator), zap.Int("count", len(req.Usernames)))
		c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))

	case "userGroup":
		// 用户组管理
		if req.GroupAction == "" || req.GroupName == "" {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少用户组操作或名称"))
			return
		}

		config, err := h.adminStorage.GetAdminConfig(ctx)
		if err != nil {
			h.logger.Error("获取配置失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取配置失败"))
			return
		}

		switch req.GroupAction {
		case "add":
			// 检查是否已存在
			for _, g := range config.UserGroups {
				if g.Name == req.GroupName {
					c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "用户组已存在"))
					return
				}
			}
			config.UserGroups = append(config.UserGroups, model.UserGroup{
				Name:        req.GroupName,
				EnabledAPIs: req.EnabledAPIs,
			})
		case "edit":
			found := false
			for i := range config.UserGroups {
				if config.UserGroups[i].Name == req.GroupName {
					config.UserGroups[i].EnabledAPIs = req.EnabledAPIs
					found = true
					break
				}
			}
			if !found {
				c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "用户组不存在"))
				return
			}
		case "delete":
			for i, g := range config.UserGroups {
				if g.Name == req.GroupName {
					config.UserGroups = append(config.UserGroups[:i], config.UserGroups[i+1:]...)
					break
				}
			}
		}

		if err := h.adminStorage.SaveAdminConfig(ctx, config); err != nil {
			h.logger.Error("保存配置失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "保存失败"))
			return
		}

		h.logger.Info("用户组已更新", zap.String("operator", operator), zap.String("action", req.GroupAction))
		c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))

	default:
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "未知的操作类型"))
	}
}

// ========== 数据迁移 /api/admin/data_migration/* ==========

// HandleDataExport 数据导出
// GET /api/admin/data_migration/export
func (h *AdminLegacyHandler) HandleDataExport(c *gin.Context) {
	// 复用新版 handler 的逻辑
	h.adminHandler.ExportData(c)
}

// HandleDataImport 数据导入
// POST /api/admin/data_migration/import
func (h *AdminLegacyHandler) HandleDataImport(c *gin.Context) {
	// 复用新版 handler 的逻辑
	h.adminHandler.ImportData(c)
}

// ========== 配置订阅 /api/admin/config_subscription/fetch ==========

// FetchSubscriptionRequest 订阅请求
type FetchSubscriptionRequest struct {
	URL string `json:"url" binding:"required"`
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

// ========== 配置文件 /api/admin/config_file ==========

// ConfigFileRequest 配置文件请求
type ConfigFileRequest struct {
	ConfigFile      string `json:"configFile" binding:"required"`
	SubscriptionURL string `json:"subscriptionUrl,omitempty"`
	AutoUpdate      bool   `json:"autoUpdate,omitempty"`
	LastCheckTime   string `json:"lastCheckTime,omitempty"`
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
