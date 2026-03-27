// internal/handler/admin_legacy_handler.go
// Admin 旧版 API 兼容层 - 供前端现有代码调用

package handler

import (
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
