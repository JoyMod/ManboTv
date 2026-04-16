// internal/handler/search_history_handler.go
package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// SearchHistoryHandler 搜索历史处理器
type SearchHistoryHandler struct {
	storage model.StorageService
	logger  *zap.Logger
}

// NewSearchHistoryHandler 创建搜索历史处理器
func NewSearchHistoryHandler(storage model.StorageService, logger *zap.Logger) *SearchHistoryHandler {
	return &SearchHistoryHandler{
		storage: storage,
		logger:  logger,
	}
}

// GetHistory 获取搜索历史
// GET /api/v1/searchhistory
func (h *SearchHistoryHandler) GetHistory(c *gin.Context) {
	username := resolveUsernameFromContext(c)
	if username == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeUnauthorized, "未登录"))
		return
	}

	history, err := h.storage.GetSearchHistory(c.Request.Context(), username, 20)
	if err != nil {
		h.logger.Error("获取搜索历史失败",
			zap.String("username", username),
			zap.Error(err),
		)
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取搜索历史失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(history))
}

// AddHistory 添加搜索历史
// POST /api/v1/searchhistory
func (h *SearchHistoryHandler) AddHistory(c *gin.Context) {
	username := resolveUsernameFromContext(c)
	if username == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeUnauthorized, "未登录"))
		return
	}

	var req struct {
		Keyword string `json:"keyword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "关键词不能为空"))
		return
	}

	// 添加搜索历史
	if err := h.storage.AddSearchHistory(c.Request.Context(), username, req.Keyword); err != nil {
		h.logger.Error("添加搜索历史失败",
			zap.String("username", username),
			zap.String("keyword", req.Keyword),
			zap.Error(err),
		)
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "添加搜索历史失败"))
		return
	}

	h.logger.Debug("搜索历史已添加",
		zap.String("username", username),
		zap.String("keyword", req.Keyword),
	)

	// 返回最新的搜索历史
	history, err := h.storage.GetSearchHistory(c.Request.Context(), username, 20)
	if err != nil {
		c.JSON(http.StatusOK, model.Success([]string{}))
		return
	}

	c.JSON(http.StatusOK, model.Success(history))
}

// DeleteHistory 删除搜索历史
// DELETE /api/v1/searchhistory?keyword=xxx
func (h *SearchHistoryHandler) DeleteHistory(c *gin.Context) {
	username := resolveUsernameFromContext(c)
	if username == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeUnauthorized, "未登录"))
		return
	}

	keyword := c.Query("keyword")

	if keyword == "" {
		// 清空所有搜索历史
		if err := h.storage.ClearSearchHistory(c.Request.Context(), username); err != nil {
			h.logger.Error("清空搜索历史失败",
				zap.String("username", username),
				zap.Error(err),
			)
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "清空搜索历史失败"))
			return
		}

		h.logger.Info("搜索历史已清空",
			zap.String("username", username),
		)
		c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
		return
	}

	if err := h.storage.RemoveSearchHistory(c.Request.Context(), username, keyword); err != nil {
		h.logger.Error("删除单条搜索历史失败",
			zap.String("username", username),
			zap.String("keyword", keyword),
			zap.Error(err),
		)
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "删除搜索历史失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}

// GetHistoryLegacy handles GET /api/searchhistory.
func (h *SearchHistoryHandler) GetHistoryLegacy(c *gin.Context) {
	username := resolveUsernameFromContext(c)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	history, err := h.storage.GetSearchHistory(c.Request.Context(), username, 20)
	if err != nil {
		h.logger.Error("获取搜索历史失败", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, history)
}

// AddHistoryLegacy handles POST /api/searchhistory.
func (h *SearchHistoryHandler) AddHistoryLegacy(c *gin.Context) {
	username := resolveUsernameFromContext(c)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req struct {
		Keyword string `json:"keyword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Keyword) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Keyword is required"})
		return
	}

	if err := h.storage.AddSearchHistory(c.Request.Context(), username, strings.TrimSpace(req.Keyword)); err != nil {
		h.logger.Error("添加搜索历史失败", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	history, err := h.storage.GetSearchHistory(c.Request.Context(), username, 20)
	if err != nil {
		c.JSON(http.StatusOK, []string{})
		return
	}

	c.JSON(http.StatusOK, history)
}

// DeleteHistoryLegacy handles DELETE /api/searchhistory.
func (h *SearchHistoryHandler) DeleteHistoryLegacy(c *gin.Context) {
	username := resolveUsernameFromContext(c)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	keyword := strings.TrimSpace(c.Query("keyword"))
	if keyword == "" {
		if err := h.storage.ClearSearchHistory(c.Request.Context(), username); err != nil {
			h.logger.Error("清空搜索历史失败", zap.String("username", username), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	if err := h.storage.RemoveSearchHistory(c.Request.Context(), username, keyword); err != nil {
		h.logger.Error("删除搜索历史失败", zap.String("username", username), zap.String("keyword", keyword), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
