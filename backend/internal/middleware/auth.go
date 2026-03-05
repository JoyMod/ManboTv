// internal/middleware/auth.go
package middleware

import (
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

// AuthConfig 认证中间件配置
type AuthConfig struct {
	JWTSecret    string
	CookieName   string
	OwnerUser    string
	OwnerPass    string
	SkipPaths    []string
	Logger       *zap.Logger
}

// AuthMiddleware 创建认证中间件
func AuthMiddleware(config *AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否在跳过路径中
		for _, path := range config.SkipPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		// 从 Cookie 中获取认证信息
		authInfo := extractAuthInfo(c, config)
		if authInfo == nil {
			config.Logger.Warn("认证失败: 无效的认证信息",
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
			)
			c.JSON(http.StatusOK, model.Error(model.CodeUnauthorized, "未登录或登录已过期"))
			c.Abort()
			return
		}

		// 检查是否过期 (7天)
		if time.Now().UnixMilli()-authInfo.Timestamp > 7*24*60*60*1000 {
			config.Logger.Warn("认证失败: 认证已过期",
				zap.String("username", authInfo.Username),
				zap.String("ip", c.ClientIP()),
			)
			c.JSON(http.StatusOK, model.Error(model.CodeUnauthorized, "登录已过期，请重新登录"))
			c.Abort()
			return
		}

		// 验证签名 (数据库模式)
		if authInfo.Username != "" && config.OwnerPass != "" {
			expectedSig := generateSignature(authInfo.Username, config.OwnerPass)
			if !hmac.Equal([]byte(authInfo.Signature), []byte(expectedSig)) {
				config.Logger.Warn("认证失败: 签名无效",
					zap.String("username", authInfo.Username),
					zap.String("ip", c.ClientIP()),
				)
				c.JSON(http.StatusOK, model.Error(model.CodeUnauthorized, "认证信息无效"))
				c.Abort()
				return
			}
		}

		// 将认证信息存入上下文
		c.Set("auth_info", authInfo)
		c.Set("username", authInfo.Username)
		c.Set("role", authInfo.Role)

		c.Next()
	}
}

// extractAuthInfo 从请求中提取认证信息
func extractAuthInfo(c *gin.Context, config *AuthConfig) *model.AuthInfo {
	// 从 Cookie 中获取
	cookie, err := c.Cookie(config.CookieName)
	if err != nil {
		// 尝试从 Header 获取
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			cookie = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			return nil
		}
	}

	// 解码 URL
	decoded, err := url.QueryUnescape(cookie)
	if err != nil {
		return nil
	}

	// 解析 JSON
	var authInfo model.AuthInfo
	if err := json.Unmarshal([]byte(decoded), &authInfo); err != nil {
		return nil
	}

	return &authInfo
}

// generateSignature 生成 HMAC-SHA256 签名
func generateSignature(data, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// GetAuthInfo 从上下文中获取认证信息
func GetAuthInfo(c *gin.Context) *model.AuthInfo {
	info, exists := c.Get("auth_info")
	if !exists {
		return nil
	}
	return info.(*model.AuthInfo)
}

// RequireRole 角色要求中间件
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusOK, model.Error(model.CodeUnauthorized, "未登录"))
			c.Abort()
			return
		}

		userRole := role.(string)
		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusOK, model.Error(model.CodeForbidden, "权限不足"))
		c.Abort()
	}
}

// RequireOwner 要求站长权限
func RequireOwner() gin.HandlerFunc {
	return RequireRole(model.RoleOwner)
}

// RequireAdmin 要求管理员权限
func RequireAdmin() gin.HandlerFunc {
	return RequireRole(model.RoleOwner, model.RoleAdmin)
}
