// internal/handler/favorite_handler.go
package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// FavoriteHandler 收藏处理器
type FavoriteHandler struct {
	storage model.StorageService
	logger  *zap.Logger
}

// NewFavoriteHandler 创建收藏处理器
func NewFavoriteHandler(storage model.StorageService, logger *zap.Logger) *FavoriteHandler {
	return &FavoriteHandler{
		storage: storage,
		logger:  logger,
	}
}

// GetFavorites 获取收藏列表
func (h *FavoriteHandler) GetFavorites(c *gin.Context) {
	// TODO: 从认证中间件获取用户名
	username := "admin" // 临时硬编码

	// 检查是否有指定key
	key := c.Query("key")
	if key != "" {
		// 查询单条
		parts := strings.Split(key, "+")
		if len(parts) != 2 {
			c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "无效的key格式"))
			return
		}

		fav, err := h.storage.GetFavorite(c.Request.Context(), username, parts[0], parts[1])
		if err != nil {
			h.logger.Error("获取收藏失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取收藏失败"))
			return
		}

		c.JSON(http.StatusOK, model.Success(fav))
		return
	}

	// 查询全部
	page := 1
	pageSize := 100

	favorites, total, err := h.storage.GetFavorites(c.Request.Context(), username, page, pageSize)
	if err != nil {
		h.logger.Error("获取收藏列表失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取收藏列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{
		"list":  favorites,
		"total": total,
	}))
}

// AddFavorite 添加收藏
func (h *FavoriteHandler) AddFavorite(c *gin.Context) {
	username := "admin" // TODO: 从认证中间件获取

	var req struct {
		Key       string         `json:"key" binding:"required"`
		Favorite  model.Favorite `json:"favorite" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "参数错误"))
		return
	}

	// 解析key
	parts := strings.Split(req.Key, "+")
	if len(parts) != 2 {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "无效的key格式"))
		return
	}

	req.Favorite.Source = parts[0]
	req.Favorite.VodID = parts[1]

	if err := h.storage.SaveFavorite(c.Request.Context(), username, &req.Favorite); err != nil {
		h.logger.Error("保存收藏失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "保存收藏失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}

// DeleteFavorite 删除收藏
func (h *FavoriteHandler) DeleteFavorite(c *gin.Context) {
	username := "admin" // TODO: 从认证中间件获取
	key := c.Param("key")

	if key == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少key参数"))
		return
	}

	parts := strings.Split(key, "+")
	if len(parts) != 2 {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "无效的key格式"))
		return
	}

	if err := h.storage.DeleteFavorite(c.Request.Context(), username, parts[0], parts[1]); err != nil {
		h.logger.Error("删除收藏失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "删除收藏失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}
