// internal/handler/favorite_handler.go
package handler

import (
	"fmt"
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
	username := resolveUsernameFromContext(c)
	if username == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeUnauthorized, "未登录"))
		return
	}

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
	username := resolveUsernameFromContext(c)
	if username == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeUnauthorized, "未登录"))
		return
	}

	var req struct {
		Key      string         `json:"key" binding:"required"`
		Favorite model.Favorite `json:"favorite" binding:"required"`
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
	username := resolveUsernameFromContext(c)
	if username == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeUnauthorized, "未登录"))
		return
	}
	key := c.Param("key")

	if key == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少key参数"))
		return
	}

	source, vodID := splitLegacyKey(key)
	if source == "" || vodID == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "无效的key格式"))
		return
	}

	if err := h.storage.DeleteFavorite(c.Request.Context(), username, source, vodID); err != nil {
		h.logger.Error("删除收藏失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "删除收藏失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}

// GetFavoritesLegacy handles GET /api/favorites and returns key-value map.
func (h *FavoriteHandler) GetFavoritesLegacy(c *gin.Context) {
	username := resolveUsernameFromContext(c)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	list, _, err := h.storage.GetFavorites(c.Request.Context(), username, 1, 10000)
	if err != nil {
		h.logger.Error("获取收藏列表失败", zap.Error(err), zap.String("username", username))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取收藏失败"})
		return
	}

	result := make(map[string]gin.H, len(list))
	for _, item := range list {
		if item == nil {
			continue
		}
		key := fmt.Sprintf("%s+%s", item.Source, item.VodID)
		result[key] = gin.H{
			"id":             key,
			"source":         item.Source,
			"source_name":    item.SourceName,
			"title":          item.VodName,
			"cover":          item.VodPic,
			"year":           item.Year,
			"total_episodes": item.TotalEpisode,
			"save_time":      item.SaveTime,
			"search_title":   item.SearchTitle,
		}
	}

	c.JSON(http.StatusOK, result)
}

// AddFavoriteLegacy handles POST /api/favorites.
func (h *FavoriteHandler) AddFavoriteLegacy(c *gin.Context) {
	username := resolveUsernameFromContext(c)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	source, vodID := parseSourceID(body)
	if source == "" || vodID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 source 或 id"})
		return
	}

	favoriteData, _ := body["favorite"].(map[string]interface{})
	if favoriteData == nil {
		favoriteData = body
	}

	favorite := &model.Favorite{
		Source:       source,
		VodID:        vodID,
		SourceName:   strValue(favoriteData, "source_name"),
		VodName:      strValue(favoriteData, "title"),
		VodPic:       strValue(favoriteData, "cover"),
		Year:         strValue(favoriteData, "year"),
		TotalEpisode: intValue(favoriteData, "total_episodes"),
		SaveTime:     int64Value(favoriteData, "save_time"),
		SearchTitle:  strValue(favoriteData, "search_title"),
	}

	if favorite.VodName == "" {
		favorite.VodName = strValue(favoriteData, "vod_name")
	}
	if favorite.VodPic == "" {
		favorite.VodPic = strValue(favoriteData, "vod_pic")
	}

	if err := h.storage.SaveFavorite(c.Request.Context(), username, favorite); err != nil {
		h.logger.Error("保存收藏失败", zap.Error(err), zap.String("username", username))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存收藏失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteFavoriteLegacy handles DELETE /api/favorites.
func (h *FavoriteHandler) DeleteFavoriteLegacy(c *gin.Context) {
	username := resolveUsernameFromContext(c)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	key := strings.TrimSpace(c.Query("key"))
	source := ""
	vodID := ""

	if key != "" {
		source, vodID = splitLegacyKey(key)
	} else {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err == nil {
			source, vodID = parseSourceID(body)
		}
	}

	if source == "" || vodID == "" {
		list, _, err := h.storage.GetFavorites(c.Request.Context(), username, 1, 10000)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "清空收藏失败"})
			return
		}
		for _, item := range list {
			if item == nil {
				continue
			}
			_ = h.storage.DeleteFavorite(c.Request.Context(), username, item.Source, item.VodID)
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	if err := h.storage.DeleteFavorite(c.Request.Context(), username, source, vodID); err != nil {
		h.logger.Error("删除收藏失败", zap.Error(err), zap.String("username", username), zap.String("source", source), zap.String("id", vodID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除收藏失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func resolveUsernameFromContext(c *gin.Context) string {
	if username := c.GetString("username"); strings.TrimSpace(username) != "" {
		return strings.TrimSpace(username)
	}
	if cookie, err := c.Cookie("auth"); err == nil && strings.TrimSpace(cookie) != "" {
		return "admin"
	}
	return ""
}

func parseSourceID(body map[string]interface{}) (string, string) {
	if body == nil {
		return "", ""
	}

	if key := strValue(body, "key"); key != "" {
		return splitLegacyKey(key)
	}

	source := strValue(body, "source")
	vodID := strValue(body, "id")
	if source != "" && vodID != "" {
		if strings.Contains(vodID, "+") || strings.Contains(vodID, "_") {
			parsedSource, parsedID := splitLegacyKey(vodID)
			if parsedSource == source && parsedID != "" {
				return source, parsedID
			}
		}
		return source, vodID
	}

	if favoriteData, ok := body["favorite"].(map[string]interface{}); ok {
		source = strValue(favoriteData, "source")
		vodID = strValue(favoriteData, "id")
		if source != "" && vodID != "" {
			return source, vodID
		}
	}

	return "", ""
}

func splitLegacyKey(key string) (string, string) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return "", ""
	}

	if strings.Contains(trimmed, "+") {
		parts := strings.SplitN(trimmed, "+", 2)
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}

	if strings.Contains(trimmed, "_") {
		parts := strings.SplitN(trimmed, "_", 2)
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}

	return "", ""
}

func strValue(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case string:
			return strings.TrimSpace(val)
		}
	}
	return ""
}

func intValue(m map[string]interface{}, key string) int {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	}
	return 0
}

func int64Value(m map[string]interface{}, key string) int64 {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int64(val)
	case int:
		return int64(val)
	case int64:
		return val
	}
	return 0
}
