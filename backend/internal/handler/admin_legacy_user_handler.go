package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

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
	GroupAction string   `json:"groupAction,omitempty"`
	Usernames   []string `json:"usernames,omitempty"`
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
		if req.Role == "admin" && !h.adminHandler.isOwner(operator) {
			c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "仅站长可创建管理员"))
			return
		}

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
		if target.Role == "admin" && !h.adminHandler.isOwner(operator) {
			c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "仅站长可操作管理员"))
			return
		}

		target.Banned = req.Action == "ban"
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
		if operator != req.Username && target.Role == "admin" && !h.adminHandler.isOwner(operator) {
			c.JSON(http.StatusOK, model.Error(model.CodePermissionDenied, "仅站长可修改其他管理员密码"))
			return
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

		for _, username := range req.Usernames {
			target, err := h.adminStorage.GetUser(ctx, username)
			if err != nil || target == nil {
				continue
			}
			target.Tags = req.UserGroups
			_ = h.adminStorage.UpdateUser(ctx, username, target)
		}

		h.logger.Info("用户组批量更新", zap.String("operator", operator), zap.Int("count", len(req.Usernames)))
		c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))

	case "userGroup":
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
			for _, group := range config.UserGroups {
				if group.Name == req.GroupName {
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
			for index := range config.UserGroups {
				if config.UserGroups[index].Name != req.GroupName {
					continue
				}
				config.UserGroups[index].EnabledAPIs = req.EnabledAPIs
				found = true
				break
			}
			if !found {
				c.JSON(http.StatusOK, model.Error(model.CodeNotFound, "用户组不存在"))
				return
			}
		case "delete":
			for index, group := range config.UserGroups {
				if group.Name != req.GroupName {
					continue
				}
				config.UserGroups = append(config.UserGroups[:index], config.UserGroups[index+1:]...)
				break
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
