// internal/handler/auth_handler.go
package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	cookieName   string
	tokenExpire  time.Duration
	jwtSecret    string
	ownerUser    string
	ownerPass    string
	adminStorage model.AdminStorageService
	logger       *zap.Logger
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(
	cookieName string,
	tokenExpire time.Duration,
	jwtSecret string,
	ownerUser string,
	ownerPass string,
	adminStorage model.AdminStorageService,
	logger *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		cookieName:   cookieName,
		tokenExpire:  tokenExpire,
		jwtSecret:    jwtSecret,
		ownerUser:    ownerUser,
		ownerPass:    ownerPass,
		adminStorage: adminStorage,
		logger:       logger,
	}
}

// Login 登录接口
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效"))
		return
	}

	h.logger.Info("登录请求",
		zap.String("username", req.Username),
		zap.String("ip", c.ClientIP()),
	)

	// 验证密码
	ok, role := h.validateCredentials(c.Request.Context(), req.Username, req.Password)
	if !ok {
		h.logger.Warn("登录失败: 密码错误",
			zap.String("username", req.Username),
			zap.String("ip", c.ClientIP()),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 生成认证 Cookie
	cookieValue, err := h.generateAuthCookie(req.Username, req.Password, role)
	if err != nil {
		h.logger.Error("生成认证Cookie失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "登录失败"))
		return
	}

	// 设置 Cookie
	c.SetCookie(
		h.cookieName,
		cookieValue,
		int(h.tokenExpire.Seconds()),
		"/",
		"",
		false,
		false,
	)

	h.logger.Info("登录成功",
		zap.String("username", req.Username),
		zap.String("role", role),
		zap.String("ip", c.ClientIP()),
	)

	c.JSON(http.StatusOK, gin.H{
		"ok":   true,
		"role": role,
	})
}

// Logout 登出接口
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// 清除 Cookie
	c.SetCookie(
		h.cookieName,
		"",
		-1,
		"/",
		"",
		false,
		false,
	)

	h.logger.Info("登出成功",
		zap.String("ip", c.ClientIP()),
	)

	c.JSON(http.StatusOK, model.Success(gin.H{"ok": true}))
}

// ChangePassword 修改密码接口
// PUT /api/v1/auth/password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword      string `json:"old_password"`
		OldPasswordCamel string `json:"oldPassword"`
		NewPassword      string `json:"new_password"`
		NewPasswordCamel string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		if strings.HasPrefix(c.Request.URL.Path, "/api/change-password") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
			return
		}
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效"))
		return
	}

	username := strings.TrimSpace(c.GetString("username"))
	role := strings.TrimSpace(c.GetString("role"))
	newPassword := firstNonEmpty(req.NewPassword, req.NewPasswordCamel)
	oldPassword := firstNonEmpty(req.OldPassword, req.OldPasswordCamel)

	if strings.HasPrefix(c.Request.URL.Path, "/api/change-password") {
		if username == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		if newPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "新密码不得为空"})
			return
		}
		if username == h.ownerUser {
			c.JSON(http.StatusForbidden, gin.H{"error": "站长不能通过此接口修改密码"})
			return
		}
		if h.adminStorage == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "当前模式不支持修改密码"})
			return
		}
		if err := h.adminStorage.ChangePassword(c.Request.Context(), username, hashAuthPassword(newPassword)); err != nil {
			h.logger.Error("修改密码失败", zap.String("username", username), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "修改密码失败"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	if username == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeUnauthorized, "未登录"))
		return
	}
	if newPassword == "" || oldPassword == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "请求参数无效"))
		return
	}

	h.logger.Info("修改密码请求",
		zap.String("username", username),
		zap.String("role", role),
		zap.String("ip", c.ClientIP()),
	)

	// 验证旧密码
	ok, _ := h.validateCredentials(c.Request.Context(), username, oldPassword)
	if !ok {
		h.logger.Warn("修改密码失败: 旧密码错误",
			zap.String("username", username),
			zap.String("ip", c.ClientIP()),
		)
		c.JSON(http.StatusOK, model.Error(model.CodeUnauthorized, "旧密码错误"))
		return
	}

	if username == h.ownerUser || h.adminStorage == nil {
		h.logger.Warn("修改密码功能暂不支持", zap.String("username", username))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "当前模式不支持修改密码"))
		return
	}

	if err := h.adminStorage.ChangePassword(c.Request.Context(), username, hashAuthPassword(newPassword)); err != nil {
		h.logger.Error("修改密码失败", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "修改密码失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}

// GetCurrentUser 获取当前登录用户信息
// GET /api/v1/auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	username, _ := c.Get("username")
	role, _ := c.Get("role")

	c.JSON(http.StatusOK, model.Success(gin.H{
		"username": username,
		"role":     role,
	}))
}

// validateCredentials 验证凭据
func (h *AuthHandler) validateCredentials(ctx context.Context, username, password string) (bool, string) {
	if h.ownerPass == "" {
		return true, model.RoleOwner
	}

	if username == h.ownerUser && password == h.ownerPass {
		return true, model.RoleOwner
	}

	if username == "" || h.adminStorage == nil {
		return false, ""
	}

	user, err := h.adminStorage.GetUser(ctx, username)
	if err != nil || user == nil || user.Banned {
		return false, ""
	}

	if user.PasswordHash == hashAuthPassword(password) {
		role := strings.TrimSpace(user.Role)
		if role == "" {
			role = model.RoleUser
		}
		return true, role
	}

	return false, ""
}

// generateAuthCookie 生成认证 Cookie 值
func (h *AuthHandler) generateAuthCookie(username, password, role string) (string, error) {
	authData := model.AuthInfo{
		Role:      role,
		Timestamp: time.Now().UnixMilli(),
	}

	// 数据库模式包含用户名和签名
	if username != "" && h.ownerPass != "" {
		authData.Username = username
		authData.Signature = generateSignature(username, h.ownerPass)
	}

	// LocalStorage 模式包含密码
	if username == "" && password != "" {
		authData.Password = password
	}

	// 序列化为 JSON
	jsonData, err := json.Marshal(authData)
	if err != nil {
		return "", err
	}

	// URL 编码
	return url.QueryEscape(string(jsonData)), nil
}

func hashAuthPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// generateSignature 生成 HMAC-SHA256 签名
func generateSignature(data, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}
