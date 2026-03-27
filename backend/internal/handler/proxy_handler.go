// internal/handler/proxy_handler.go
package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
	"github.com/JoyMod/ManboTV/backend/internal/service"
)

// ProxyHandler 视频代理处理器
type ProxyHandler struct {
	m3u8Service    service.M3U8ProxyService
	segmentService service.SegmentProxyService
	logger         *zap.Logger
	httpClient     *http.Client
}

// NewProxyHandler 创建视频代理处理器
func NewProxyHandler(
	m3u8Service service.M3U8ProxyService,
	segmentService service.SegmentProxyService,
	logger *zap.Logger,
) *ProxyHandler {
	return &ProxyHandler{
		m3u8Service:    m3u8Service,
		segmentService: segmentService,
		logger:         logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyM3U8 M3U8 代理接口
// GET /api/v1/proxy/m3u8?url=xxx&allowCORS=true
func (h *ProxyHandler) ProxyM3U8(c *gin.Context) {
	targetURL := c.Query("url")
	if targetURL == "" {
		c.JSON(http.StatusBadRequest, model.Error(model.CodeInvalidParams, "缺少URL参数"))
		return
	}

	allowCORS := c.Query("allowCORS") == "true"
	proxyBase := resolveLegacyProxyBase(c)

	h.logger.Debug("M3U8代理请求",
		zap.String("url", targetURL),
		zap.Bool("allowCORS", allowCORS),
		zap.String("proxy_base", proxyBase),
		zap.String("client_ip", c.ClientIP()),
	)

	// 执行代理
	content, err := h.m3u8Service.ProxyM3U8(c.Request.Context(), targetURL, allowCORS, proxyBase)
	if err != nil {
		h.logger.Error("M3U8代理失败",
			zap.String("url", targetURL),
			zap.Error(err),
		)
		c.JSON(http.StatusBadGateway, model.Error(model.CodeInternalError, "获取M3U8失败"))
		return
	}

	// 设置响应头
	c.Header("Content-Type", "application/vnd.apple.mpegurl")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Range, Origin, Accept")
	c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Range")
	c.Header("Cache-Control", "no-cache")

	c.Data(http.StatusOK, "application/vnd.apple.mpegurl", content)
}

// ProxySegment 视频片段代理接口
// GET /api/v1/proxy/segment?url=xxx
func (h *ProxyHandler) ProxySegment(c *gin.Context) {
	targetURL := c.Query("url")
	if targetURL == "" {
		c.JSON(http.StatusBadRequest, model.Error(model.CodeInvalidParams, "缺少URL参数"))
		return
	}

	h.logger.Debug("片段代理请求",
		zap.String("url", targetURL),
		zap.String("client_ip", c.ClientIP()),
	)

	// 执行代理
	rangeHeader := c.GetHeader("Range")
	resp, err := h.segmentService.ProxySegment(c.Request.Context(), targetURL, rangeHeader)
	if err != nil {
		h.logger.Error("片段代理失败",
			zap.String("url", targetURL),
			zap.String("range", rangeHeader),
			zap.Error(err),
		)
		c.JSON(http.StatusBadGateway, model.Error(model.CodeInternalError, "获取视频片段失败"))
		return
	}
	defer resp.Body.Close()

	// 设置响应头
	c.Header("Content-Type", resp.ContentType)
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Range, Origin, Accept")
	c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Range")
	c.Header("Cache-Control", "no-cache")

	if resp.ContentLength > 0 {
		c.Header("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
	}
	if resp.ContentRange != "" {
		c.Header("Content-Range", resp.ContentRange)
	}
	if resp.AcceptRanges != "" {
		c.Header("Accept-Ranges", resp.AcceptRanges)
	}

	// 流式传输
	statusCode := resp.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusOK
	}
	c.Status(statusCode)
	if _, copyErr := io.Copy(c.Writer, resp.Body); copyErr != nil {
		h.logger.Warn("片段流式传输中断",
			zap.String("url", targetURL),
			zap.String("range", rangeHeader),
			zap.Error(copyErr),
		)
	}
}

func resolveLegacyProxyBase(c *gin.Context) string {
	host := firstHeaderValue(c.GetHeader("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(c.Request.Host)
	}

	proto := strings.ToLower(firstHeaderValue(c.GetHeader("X-Forwarded-Proto")))
	if proto == "" {
		if c.Request.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}

	if host == "" {
		return "/api/proxy"
	}

	return proto + "://" + host + "/api/proxy"
}

func firstHeaderValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if idx := strings.Index(value, ","); idx >= 0 {
		return strings.TrimSpace(value[:idx])
	}
	return value
}

// ProxyKey 密钥代理接口
// GET /api/v1/proxy/key?url=xxx
func (h *ProxyHandler) ProxyKey(c *gin.Context) {
	targetURL := c.Query("url")
	if targetURL == "" {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "缺少URL参数"))
		return
	}

	h.logger.Debug("密钥代理请求",
		zap.String("url", targetURL),
		zap.String("client_ip", c.ClientIP()),
	)

	// 解码URL
	decodedURL, err := url.QueryUnescape(targetURL)
	if err != nil {
		c.JSON(http.StatusOK, model.Error(model.CodeInvalidParams, "URL解码失败"))
		return
	}

	// 创建请求
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, decodedURL, nil)
	if err != nil {
		h.logger.Error("创建请求失败", zap.Error(err))
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "创建请求失败"))
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Accept", "*/*")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		h.logger.Error("密钥请求失败",
			zap.String("url", decodedURL),
			zap.Error(err),
		)
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "获取密钥失败"))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusOK, model.Error(model.CodeInternalError, "密钥请求失败"))
		return
	}

	// 设置响应头
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Cache-Control", "no-cache")

	// 流式传输
	c.Status(http.StatusOK)
	io.Copy(c.Writer, resp.Body)
}

// ProxyLogo Logo代理接口
// GET /api/v1/proxy/logo?url=xxx
func (h *ProxyHandler) ProxyLogo(c *gin.Context) {
	targetURL := c.Query("url")
	if targetURL == "" {
		c.Redirect(http.StatusFound, "/placeholder-poster.svg")
		return
	}

	h.logger.Debug("Logo代理请求",
		zap.String("url", targetURL),
		zap.String("client_ip", c.ClientIP()),
	)

	// 解码URL
	decodedURL, err := url.QueryUnescape(targetURL)
	if err != nil {
		c.Redirect(http.StatusFound, "/placeholder-poster.svg")
		return
	}

	decodedURL = strings.TrimSpace(decodedURL)
	if decodedURL == "" {
		c.Redirect(http.StatusFound, "/placeholder-poster.svg")
		return
	}
	if strings.HasPrefix(decodedURL, "//") {
		decodedURL = "https:" + decodedURL
	}
	if !strings.HasPrefix(decodedURL, "http://") && !strings.HasPrefix(decodedURL, "https://") {
		c.Redirect(http.StatusFound, "/placeholder-poster.svg")
		return
	}

	parsedURL, err := url.Parse(decodedURL)
	if err != nil || parsedURL.Host == "" {
		c.Redirect(http.StatusFound, "/placeholder-poster.svg")
		return
	}

	primaryReferer := fmt.Sprintf("%s://%s/", parsedURL.Scheme, parsedURL.Host)
	attempts := []string{decodedURL}
	if parsedURL.Scheme == "https" {
		attempts = append(attempts, strings.Replace(decodedURL, "https://", "http://", 1))
	} else if parsedURL.Scheme == "http" {
		attempts = append(attempts, strings.Replace(decodedURL, "http://", "https://", 1))
	}

	var resp *http.Response
	for _, attemptURL := range attempts {
		req, reqErr := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, attemptURL, nil)
		if reqErr != nil {
			continue
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/124.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8")
		req.Header.Set("Referer", primaryReferer)
		req.Header.Set("Origin", primaryReferer)
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")

		candidate, doErr := h.httpClient.Do(req)
		if doErr != nil {
			h.logger.Debug("Logo请求尝试失败", zap.String("url", attemptURL), zap.Error(doErr))
			continue
		}

		if candidate.StatusCode != http.StatusOK {
			candidate.Body.Close()
			continue
		}

		resp = candidate
		break
	}

	if resp == nil {
		c.Redirect(http.StatusFound, "/placeholder-poster.svg")
		return
	}
	defer resp.Body.Close()

	// 设置响应头
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/png"
	}
	if strings.Contains(strings.ToLower(contentType), "text/html") {
		c.Redirect(http.StatusFound, "/placeholder-poster.svg")
		return
	}

	c.Header("Content-Type", contentType)
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Cache-Control", "public, max-age=86400") // Logo缓存1天

	// 流式传输
	c.Status(http.StatusOK)
	io.Copy(c.Writer, resp.Body)
}
