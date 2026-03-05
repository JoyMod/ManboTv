// internal/handler/image_handler.go
package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
	"github.com/JoyMod/ManboTV/backend/internal/service"
)

// ImageHandler 图片处理器
type ImageHandler struct {
	service service.ImageService
	logger  *zap.Logger
}

// NewImageHandler 创建图片处理器
func NewImageHandler(service service.ImageService, logger *zap.Logger) *ImageHandler {
	return &ImageHandler{
		service: service,
		logger:  logger,
	}
}

// Proxy 图片代理接口
func (h *ImageHandler) Proxy(c *gin.Context) {
	imageURL := c.Query("url")
	if imageURL == "" {
		if isLegacyImagePath(c.FullPath()) {
			c.String(http.StatusBadRequest, "Missing URL parameter")
			return
		}
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少图片URL参数"))
		return
	}

	h.logger.Debug("图片代理请求",
		zap.String("url", imageURL),
		zap.String("client_ip", c.ClientIP()),
	)

	// 执行代理
	resp, err := h.service.Proxy(c.Request.Context(), imageURL)
	if err != nil {
		h.logger.Error("图片代理失败",
			zap.String("url", imageURL),
			zap.Error(err),
		)
		if isLegacyImagePath(c.FullPath()) {
			c.Redirect(http.StatusTemporaryRedirect, "/placeholder-poster.svg")
			return
		}
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取图片失败"))
		return
	}

	// 设置响应头
	c.Header("Content-Type", resp.ContentType)
	c.Header("Content-Length", strconv.Itoa(resp.Size))

	// 缓存控制 (半年)
	cacheSeconds := 15720000
	c.Header("Cache-Control", "public, max-age="+strconv.Itoa(cacheSeconds))
	c.Header("CDN-Cache-Control", "public, s-maxage="+strconv.Itoa(cacheSeconds))
	c.Header("Vercel-CDN-Cache-Control", "public, s-maxage="+strconv.Itoa(cacheSeconds))

	if resp.FromCache {
		c.Header("X-Cache", "HIT")
	} else {
		c.Header("X-Cache", "MISS")
	}

	c.Data(http.StatusOK, resp.ContentType, resp.Data)
}

// ProxyWithHeader 带自定义Header的图片代理 (用于特殊 referer)
func (h *ImageHandler) ProxyWithHeader(c *gin.Context) {
	imageURL := c.Query("url")
	referer := c.Query("referer")
	if referer == "" {
		referer = c.Query("header")
	}

	if imageURL == "" {
		if isLegacyImagePath(c.FullPath()) {
			c.String(http.StatusBadRequest, "Missing URL parameter")
			return
		}
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少图片URL参数"))
		return
	}

	h.logger.Debug("图片代理请求(带Header)",
		zap.String("url", imageURL),
		zap.String("referer", referer),
	)

	// 复用主代理逻辑
	resp, err := h.service.Proxy(c.Request.Context(), imageURL)
	if err != nil {
		h.logger.Error("图片代理失败",
			zap.String("url", imageURL),
			zap.Error(err),
		)
		if isLegacyImagePath(c.FullPath()) {
			c.Redirect(http.StatusTemporaryRedirect, "/placeholder-poster.svg")
			return
		}
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取图片失败"))
		return
	}

	// 设置响应头
	c.Header("Content-Type", resp.ContentType)
	c.Header("Content-Length", strconv.Itoa(resp.Size))
	c.Header("Cache-Control", "public, max-age=15720000")

	c.Data(http.StatusOK, resp.ContentType, resp.Data)
}

func isLegacyImagePath(path string) bool {
	return strings.HasPrefix(path, "/api/image")
}

// GetCacheStats 获取缓存统计 (管理接口)
func (h *ImageHandler) GetCacheStats(c *gin.Context) {
	// TODO: 实现缓存统计
	c.JSON(http.StatusOK, model.Success(gin.H{
		"status": "ok",
		"time":   time.Now().Unix(),
	}))
}
