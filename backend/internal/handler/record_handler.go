// internal/handler/record_handler.go
package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// RecordHandler 播放记录处理器
type RecordHandler struct {
	storage model.StorageService
	logger  *zap.Logger
}

// NewRecordHandler 创建播放记录处理器
func NewRecordHandler(storage model.StorageService, logger *zap.Logger) *RecordHandler {
	return &RecordHandler{
		storage: storage,
		logger:  logger,
	}
}

// GetRecords 获取播放记录列表
func (h *RecordHandler) GetRecords(c *gin.Context) {
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

		record, err := h.storage.GetPlayRecord(c.Request.Context(), username, parts[0], parts[1])
		if err != nil {
			h.logger.Error("获取播放记录失败", zap.Error(err))
			c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取播放记录失败"))
			return
		}

		c.JSON(http.StatusOK, model.Success(record))
		return
	}

	// 查询全部
	page := 1
	pageSize := 100

	records, total, err := h.storage.GetPlayRecords(c.Request.Context(), username, page, pageSize)
	if err != nil {
		h.logger.Error("获取播放记录列表失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取播放记录列表失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{
		"list":  records,
		"total": total,
	}))
}

// SaveRecord 保存播放记录
func (h *RecordHandler) SaveRecord(c *gin.Context) {
	username := resolveUsernameFromContext(c)
	if username == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeUnauthorized, "未登录"))
		return
	}

	var record model.PlayRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "参数错误"))
		return
	}

	if record.Source == "" || record.VodID == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少必要参数"))
		return
	}

	if err := h.storage.SavePlayRecord(c.Request.Context(), username, &record); err != nil {
		h.logger.Error("保存播放记录失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "保存播放记录失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}

// DeleteRecord 删除播放记录
func (h *RecordHandler) DeleteRecord(c *gin.Context) {
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

	if err := h.storage.DeletePlayRecord(c.Request.Context(), username, source, vodID); err != nil {
		h.logger.Error("删除播放记录失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "删除播放记录失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}

// GetRecordsLegacy handles GET /api/playrecords and returns key-value map.
func (h *RecordHandler) GetRecordsLegacy(c *gin.Context) {
	username := resolveUsernameFromContext(c)
	if username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	list, _, err := h.storage.GetPlayRecords(c.Request.Context(), username, 1, 10000)
	if err != nil {
		h.logger.Error("获取播放记录失败", zap.Error(err), zap.String("username", username))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取播放记录失败"})
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
			"index":          item.EpisodeIndex,
			"total_episodes": item.TotalEpisodes,
			"play_time":      item.Progress,
			"total_time":     item.Duration,
			"save_time":      item.SaveTime,
			"last_play_time": item.UpdatedAt,
			"search_title":   item.SearchTitle,
		}
	}

	c.JSON(http.StatusOK, result)
}

// SaveRecordLegacy handles POST /api/playrecords.
func (h *RecordHandler) SaveRecordLegacy(c *gin.Context) {
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

	recordData, _ := body["record"].(map[string]interface{})
	if recordData == nil {
		recordData = body
	}

	record := &model.PlayRecord{
		Source:        source,
		VodID:         vodID,
		SourceName:    strValue(recordData, "source_name"),
		VodName:       strValue(recordData, "title"),
		VodPic:        strValue(recordData, "cover"),
		Year:          strValue(recordData, "year"),
		EpisodeIndex:  intValue(recordData, "index"),
		TotalEpisodes: intValue(recordData, "total_episodes"),
		Progress:      intValue(recordData, "play_time"),
		Duration:      intValue(recordData, "total_time"),
		SaveTime:      int64Value(recordData, "save_time"),
		SearchTitle:   strValue(recordData, "search_title"),
	}

	if record.VodName == "" {
		record.VodName = strValue(recordData, "vod_name")
	}
	if record.VodPic == "" {
		record.VodPic = strValue(recordData, "vod_pic")
	}

	if err := h.storage.SavePlayRecord(c.Request.Context(), username, record); err != nil {
		h.logger.Error("保存播放记录失败", zap.Error(err), zap.String("username", username))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存播放记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteRecordLegacy handles DELETE /api/playrecords.
func (h *RecordHandler) DeleteRecordLegacy(c *gin.Context) {
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
		list, _, err := h.storage.GetPlayRecords(c.Request.Context(), username, 1, 10000)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "清空播放记录失败"})
			return
		}
		for _, item := range list {
			if item == nil {
				continue
			}
			_ = h.storage.DeletePlayRecord(c.Request.Context(), username, item.Source, item.VodID)
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	if err := h.storage.DeletePlayRecord(c.Request.Context(), username, source, vodID); err != nil {
		h.logger.Error("删除播放记录失败", zap.Error(err), zap.String("username", username), zap.String("source", source), zap.String("id", vodID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除播放记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
