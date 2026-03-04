// internal/handler/record_handler.go
package handler

import (
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
	username := "admin" // TODO: 从认证中间件获取

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
	username := "admin" // TODO: 从认证中间件获取

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

	if err := h.storage.DeletePlayRecord(c.Request.Context(), username, parts[0], parts[1]); err != nil {
		h.logger.Error("删除播放记录失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "删除播放记录失败"))
		return
	}

	c.JSON(http.StatusOK, model.Success(gin.H{"success": true}))
}
