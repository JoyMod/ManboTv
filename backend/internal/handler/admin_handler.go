// internal/handler/admin_handler.go
// Admin 管理后台处理器 - 完整实现

package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// AdminHandler 管理后台处理器
type AdminHandler struct {
	storage      model.StorageService
	adminStorage model.AdminStorageService
	logger       *zap.Logger
	ownerUser    string
	ownerPass    string
}

const (
	adminStatsProbePage     = 1
	adminStatsProbePageSize = 1
)

func appendOwnerIfMissing(users []model.UserConfig, ownerUsername string) []model.UserConfig {
	for _, user := range users {
		if user.Username == ownerUsername {
			return users
		}
	}

	return append(users, model.UserConfig{
		Username: ownerUsername,
		Role:     model.RoleOwner,
	})
}

// NewAdminHandler 创建管理后台处理器
func NewAdminHandler(
	storage model.StorageService,
	adminStorage model.AdminStorageService,
	ownerUser string,
	ownerPass string,
	logger *zap.Logger,
) *AdminHandler {
	return &AdminHandler{
		storage:      storage,
		adminStorage: adminStorage,
		logger:       logger,
		ownerUser:    ownerUser,
		ownerPass:    ownerPass,
	}
}

// hashPassword 密码哈希
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// isOwner 检查是否站长
func (h *AdminHandler) isOwner(username string) bool {
	return username == h.ownerUser
}

// requireOwner 要求站长权限
func (h *AdminHandler) requireOwner(c *gin.Context) bool {
	username := c.GetString("username")
	if !h.isOwner(username) {
		c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "仅站长可执行此操作"))
		return false
	}
	return true
}

// requireAdmin 要求管理员权限
func (h *AdminHandler) requireAdmin(c *gin.Context) bool {
	username := c.GetString("username")

	// 站长直接通过
	if h.isOwner(username) {
		return true
	}

	// 检查是否为管理员
	user, err := h.adminStorage.GetUser(c.Request.Context(), username)
	if err != nil || user == nil || user.Role != "admin" || user.Banned {
		c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "权限不足"))
		return false
	}
	return true
}

// canManageUser 检查是否可以管理目标用户
func (h *AdminHandler) canManageUser(operator string, target *model.UserConfig) bool {
	// 站长可以管理所有人（除了自己某些操作）
	if h.isOwner(operator) {
		return true
	}

	// 管理员只能管理普通用户
	if target.Role == "owner" || target.Role == "admin" {
		return false
	}

	return true
}

// ========== 配置管理 ==========

// GetConfig 获取配置
// GET /api/v1/admin/config
func (h *AdminHandler) GetConfig(c *gin.Context) {
	config, err := h.adminStorage.GetAdminConfig(c.Request.Context())
	if err != nil {
		h.logger.Error("获取配置失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取配置失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(config))
}

// UpdateConfig 更新配置
// PUT /api/v1/admin/config
func (h *AdminHandler) UpdateConfig(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}

	var config model.AdminConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效"))
		return
	}

	if err := h.adminStorage.SaveAdminConfig(c.Request.Context(), &config); err != nil {
		h.logger.Error("保存配置失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "保存配置失败"))
		return
	}

	h.logger.Info("配置已更新", zap.String("operator", c.GetString("username")))
	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}

// ========== 用户管理 ==========

// GetUsers 获取用户列表
// GET /api/v1/admin/users
func (h *AdminHandler) GetUsers(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}

	users, err := h.adminStorage.GetAllUsers(c.Request.Context())
	if err != nil {
		h.logger.Error("获取用户列表失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取用户列表失败"))
		return
	}

	// 添加站长到列表开头
	owner := model.UserConfig{
		Username: h.ownerUser,
		Role:     "owner",
	}
	users = append([]model.UserConfig{owner}, users...)

	c.JSON(http.StatusOK, model.Success(users))
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=user admin"`
}

// CreateUser 创建用户
// POST /api/v1/admin/users
func (h *AdminHandler) CreateUser(c *gin.Context) {
	operator := c.GetString("username")

	// 站长和管理员都可以创建用户，但管理员只能创建普通用户
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效: "+err.Error()))
		return
	}

	// 权限检查
	if req.Role == "admin" && !h.isOwner(operator) {
		c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "仅站长可创建管理员"))
		return
	}

	// 检查用户是否已存在
	existing, err := h.adminStorage.GetUser(c.Request.Context(), req.Username)
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

	if err := h.adminStorage.CreateUser(c.Request.Context(), user); err != nil {
		h.logger.Error("创建用户失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "创建用户失败"))
		return
	}

	h.logger.Info("用户已创建",
		zap.String("operator", operator),
		zap.String("username", req.Username),
		zap.String("role", req.Role),
	)
	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Role        string   `json:"role,omitempty"`
	Banned      *bool    `json:"banned,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	EnabledAPIs []string `json:"enabled_apis,omitempty"`
}

// UpdateUser 更新用户
// PUT /api/v1/admin/users/:username
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	operator := c.GetString("username")
	targetUsername := c.Param("username")

	// 不能操作自己（通过此接口）
	if operator == targetUsername {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "不能通过此接口操作自己"))
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效"))
		return
	}

	// 获取目标用户
	target, err := h.adminStorage.GetUser(c.Request.Context(), targetUsername)
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
	if !h.canManageUser(operator, target) {
		c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "权限不足，无法管理该用户"))
		return
	}

	// 更新字段
	if req.Role != "" {
		// 只有站长可以改角色
		if !h.isOwner(operator) {
			c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "仅站长可修改角色"))
			return
		}
		target.Role = req.Role
	}

	if req.Banned != nil {
		target.Banned = *req.Banned
	}

	if req.Tags != nil {
		target.Tags = req.Tags
	}

	if req.EnabledAPIs != nil {
		target.EnabledAPIs = req.EnabledAPIs
	}

	if err := h.adminStorage.UpdateUser(c.Request.Context(), targetUsername, target); err != nil {
		h.logger.Error("更新用户失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "更新用户失败"))
		return
	}

	h.logger.Info("用户已更新",
		zap.String("operator", operator),
		zap.String("target", targetUsername),
	)
	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	Password string `json:"password" binding:"required,min=6"`
}

// ChangeUserPassword 修改用户密码
// PUT /api/v1/admin/users/:username/password
func (h *AdminHandler) ChangeUserPassword(c *gin.Context) {
	operator := c.GetString("username")
	targetUsername := c.Param("username")

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "密码至少6位"))
		return
	}

	// 获取目标用户
	target, err := h.adminStorage.GetUser(c.Request.Context(), targetUsername)
	if err != nil {
		h.logger.Error("获取用户失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取用户失败"))
		return
	}
	if target == nil {
		c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "用户不存在"))
		return
	}

	// 权限检查：只能修改自己的密码，或者管理员修改普通用户
	if operator != targetUsername {
		if !h.canManageUser(operator, target) {
			c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "权限不足"))
			return
		}
	}

	// 修改密码
	hash := hashPassword(req.Password)
	if err := h.adminStorage.ChangePassword(c.Request.Context(), targetUsername, hash); err != nil {
		h.logger.Error("修改密码失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "修改密码失败"))
		return
	}

	h.logger.Info("密码已修改",
		zap.String("operator", operator),
		zap.String("target", targetUsername),
	)
	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}

// DeleteUser 删除用户
// DELETE /api/v1/admin/users/:username
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	operator := c.GetString("username")
	targetUsername := c.Param("username")

	// 不能删除自己
	if operator == targetUsername {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "不能删除自己"))
		return
	}

	// 获取目标用户
	target, err := h.adminStorage.GetUser(c.Request.Context(), targetUsername)
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
	if !h.canManageUser(operator, target) {
		c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "权限不足，无法删除该用户"))
		return
	}

	if err := h.adminStorage.DeleteUser(c.Request.Context(), targetUsername); err != nil {
		h.logger.Error("删除用户失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "删除用户失败"))
		return
	}

	h.logger.Info("用户已删除",
		zap.String("operator", operator),
		zap.String("target", targetUsername),
	)
	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}

// ========== 站点配置（视频源）==========

// GetSites 获取站点（视频源）列表
// GET /api/v1/admin/sites
func (h *AdminHandler) GetSites(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}

	sources, err := h.adminStorage.GetVideoSources(c.Request.Context())
	if err != nil {
		h.logger.Error("获取视频源失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取视频源失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(sources))
}

// UpdateSiteRequest 更新站点请求
type UpdateSiteRequest struct {
	Name     string `json:"name,omitempty"`
	API      string `json:"api,omitempty"`
	Detail   string `json:"detail,omitempty"`
	Disabled *bool  `json:"disabled,omitempty"`
}

// UpdateSite 更新站点
// PUT /api/v1/admin/sites/:key
func (h *AdminHandler) UpdateSite(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}

	key := c.Param("key")

	var req UpdateSiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效"))
		return
	}

	sources, err := h.adminStorage.GetVideoSources(c.Request.Context())
	if err != nil {
		h.logger.Error("获取视频源失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取视频源失败"))
		return
	}

	// 查找并更新
	found := false
	for i := range sources {
		if sources[i].Key == key {
			if req.Name != "" {
				sources[i].Name = req.Name
			}
			if req.API != "" {
				sources[i].API = req.API
			}
			if req.Detail != "" {
				sources[i].Detail = req.Detail
			}
			if req.Disabled != nil {
				sources[i].Disabled = *req.Disabled
			}
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "站点不存在"))
		return
	}

	if err := h.adminStorage.SaveVideoSources(c.Request.Context(), sources); err != nil {
		h.logger.Error("保存视频源失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "保存视频源失败"))
		return
	}

	h.logger.Info("站点已更新", zap.String("key", key))
	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}

// DeleteSite 删除站点
// DELETE /api/v1/admin/sites/:key
func (h *AdminHandler) DeleteSite(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}

	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少站点标识"))
		return
	}

	sources, err := h.adminStorage.GetVideoSources(c.Request.Context())
	if err != nil {
		h.logger.Error("获取视频源失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取视频源失败"))
		return
	}

	index := -1
	for i := range sources {
		if sources[i].Key == key {
			if sources[i].From == "config" {
				c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "不能删除配置文件中的站点"))
				return
			}
			index = i
			break
		}
	}

	if index == -1 {
		c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "站点不存在"))
		return
	}

	sources = append(sources[:index], sources[index+1:]...)
	if err := h.adminStorage.SaveVideoSources(c.Request.Context(), sources); err != nil {
		h.logger.Error("保存视频源失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "删除站点失败"))
		return
	}

	h.logger.Info("站点已删除", zap.String("key", key))
	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}

// ========== 统计数据 ==========

// GetDataStatus 获取数据状态
// GET /api/v1/admin/data-status
func (h *AdminHandler) GetDataStatus(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}

	ctx := c.Request.Context()

	// 获取统计数据
	config, _ := h.adminStorage.GetAdminConfig(ctx)
	users, _ := h.adminStorage.GetAllUsers(ctx)
	users = appendOwnerIfMissing(users, h.ownerUser)

	stats := model.AdminStats{
		TotalUsers:       int64(len(users)),
		VideoSources:     len(config.VideoSources),
		LiveSources:      len(config.LiveSources),
		CustomCategories: len(config.CustomCategories),
	}

	for _, user := range users {
		_, favoriteTotal, err := h.storage.GetFavorites(
			ctx,
			user.Username,
			adminStatsProbePage,
			adminStatsProbePageSize,
		)
		if err != nil {
			h.logger.Warn("获取用户收藏统计失败", zap.String("username", user.Username), zap.Error(err))
		} else {
			stats.TotalFavorites += favoriteTotal
		}

		_, recordTotal, err := h.storage.GetPlayRecords(
			ctx,
			user.Username,
			adminStatsProbePage,
			adminStatsProbePageSize,
		)
		if err != nil {
			h.logger.Warn("获取用户播放记录统计失败", zap.String("username", user.Username), zap.Error(err))
			continue
		}
		stats.TotalRecords += recordTotal
	}

	c.JSON(http.StatusOK, model.Success(stats))
}

// ========== 数据迁移 ==========

// ExportData 导出数据
// GET /api/v1/admin/data/export
func (h *AdminHandler) ExportData(c *gin.Context) {
	if !h.requireAdmin(c) {
		return
	}

	config, err := h.adminStorage.GetAdminConfig(c.Request.Context())
	if err != nil {
		h.logger.Error("获取配置失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "导出失败"))
		return
	}

	data := model.ExportData{
		ExportTime: time.Now().Unix(),
		Version:    "1.0.0",
		Data:       *config,
	}

	h.logger.Info("数据已导出", zap.String("operator", c.GetString("username")))
	c.JSON(http.StatusOK, model.Success(data))
}

// ImportDataRequest 导入数据请求
type ImportDataRequest struct {
	Data model.AdminConfig `json:"data" binding:"required"`
}

// ImportData 导入数据
// POST /api/v1/admin/data/import
func (h *AdminHandler) ImportData(c *gin.Context) {
	if !h.requireOwner(c) {
		return
	}

	var req ImportDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "数据格式无效"))
		return
	}

	// 导入数据
	if err := h.adminStorage.SaveAdminConfig(c.Request.Context(), &req.Data); err != nil {
		h.logger.Error("保存配置失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "导入失败"))
		return
	}

	h.logger.Info("数据已导入", zap.String("operator", c.GetString("username")))
	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}
