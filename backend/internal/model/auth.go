// internal/model/auth.go
package model

// AuthInfo 认证信息
type AuthInfo struct {
	Username  string    `json:"username,omitempty"`
	Password  string    `json:"password,omitempty"`
	Role      string    `json:"role"`
	Signature string    `json:"signature,omitempty"`
	Timestamp int64     `json:"timestamp"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Ok      bool   `json:"ok"`
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// Role 用户角色
const (
	RoleOwner = "owner"
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// IsValidRole 检查角色是否有效
func IsValidRole(role string) bool {
	switch role {
	case RoleOwner, RoleAdmin, RoleUser:
		return true
	}
	return false
}

// HasPermission 检查是否有权限
func HasPermission(userRole, requiredRole string) bool {
	roleLevels := map[string]int{
		RoleUser:  1,
		RoleAdmin: 2,
		RoleOwner: 3,
	}
	
	userLevel := roleLevels[userRole]
	requiredLevel := roleLevels[requiredRole]
	
	return userLevel >= requiredLevel
}
