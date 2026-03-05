// internal/service/m3u8_proxy_service.go
package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/config"
)

// M3U8ProxyService M3U8代理服务接口
type M3U8ProxyService interface {
	ProxyM3U8(ctx context.Context, targetURL string, allowCORS bool, proxyBase string) ([]byte, error)
}

// m3u8ProxyService M3U8代理服务实现
type m3u8ProxyService struct {
	client    *http.Client
	logger    *zap.Logger
	timeout   time.Duration
	userAgent string
}

// NewM3U8ProxyService 创建M3U8代理服务
func NewM3U8ProxyService(cfg *config.HTTPClientConfig, logger *zap.Logger) M3U8ProxyService {
	client := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        cfg.MaxIdleConns,
			MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
			IdleConnTimeout:     cfg.IdleConnTimeout,
		},
	}

	return &m3u8ProxyService{
		client:    client,
		logger:    logger,
		timeout:   cfg.Timeout,
		userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
	}
}

// ProxyM3U8 代理并重写M3U8内容
func (s *m3u8ProxyService) ProxyM3U8(ctx context.Context, targetURL string, allowCORS bool, proxyBase string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// 解码URL
	decodedURL, err := url.QueryUnescape(targetURL)
	if err != nil {
		return nil, fmt.Errorf("解码URL失败: %w", err)
	}

	// 获取M3U8内容
	content, finalURL, err := s.fetchM3U8(ctx, decodedURL)
	if err != nil {
		return nil, fmt.Errorf("获取M3U8失败: %w", err)
	}

	// 重写M3U8内容
	baseURL := s.getBaseURL(finalURL)
	rewritten := s.rewriteM3U8Content(content, baseURL, proxyBase, allowCORS)

	s.logger.Info("M3U8代理完成",
		zap.String("url", decodedURL),
		zap.Int("content_length", len(rewritten)),
	)

	return []byte(rewritten), nil
}

// fetchM3U8 获取M3U8内容，返回内容和最终URL（处理重定向）
func (s *m3u8ProxyService) fetchM3U8(ctx context.Context, targetURL string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "*/*")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("读取响应失败: %w", err)
	}

	return string(body), resp.Request.URL.String(), nil
}

// getBaseURL 获取基础URL
func (s *m3u8ProxyService) getBaseURL(targetURL string) string {
	u, err := url.Parse(targetURL)
	if err != nil {
		return targetURL
	}

	// 移除文件名，保留目录路径
	path := u.Path
	if idx := strings.LastIndex(path, "/"); idx != -1 {
		path = path[:idx+1]
	}

	return fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, path)
}

// rewriteM3U8Content 重写M3U8内容
func (s *m3u8ProxyService) rewriteM3U8Content(content, baseURL, proxyBase string, allowCORS bool) string {
	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(content))

	var nextLineIsStreamURL bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 空行保留
		if line == "" {
			result.WriteString("\n")
			continue
		}

		// 处理 EXT-X-STREAM-INF（嵌套M3U8）
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			result.WriteString(line + "\n")
			nextLineIsStreamURL = true
			continue
		}

		// 处理嵌套M3U8的URL
		if nextLineIsStreamURL {
			nextLineIsStreamURL = false
			if !strings.HasPrefix(line, "#") {
				rewrittenURL := s.rewriteURL(line, baseURL, proxyBase, "m3u8", allowCORS)
				result.WriteString(rewrittenURL + "\n")
				continue
			}
		}

		// 处理 EXT-X-MAP URI
		if strings.HasPrefix(line, "#EXT-X-MAP:") {
			rewritten := s.rewriteMapURI(line, baseURL, proxyBase, allowCORS)
			result.WriteString(rewritten + "\n")
			continue
		}

		// 处理 EXT-X-KEY URI
		if strings.HasPrefix(line, "#EXT-X-KEY:") {
			rewritten := s.rewriteKeyURI(line, baseURL, proxyBase, allowCORS)
			result.WriteString(rewritten + "\n")
			continue
		}

		// 处理普通媒体片段URL
		if !strings.HasPrefix(line, "#") {
			rewrittenURL := s.rewriteURL(line, baseURL, proxyBase, "segment", allowCORS)
			result.WriteString(rewrittenURL + "\n")
			continue
		}

		// 其他行原样输出
		result.WriteString(line + "\n")
	}

	return result.String()
}

// rewriteURL 重写URL为代理URL
func (s *m3u8ProxyService) rewriteURL(originalURL, baseURL, proxyBase string, urlType string, allowCORS bool) string {
	// 如果是绝对URL，直接使用；否则解析为绝对URL
	resolvedURL := originalURL
	if !strings.HasPrefix(originalURL, "http://") && !strings.HasPrefix(originalURL, "https://") {
		resolvedURL = s.resolveURL(baseURL, originalURL)
	}

	// 如果允许CORS，直接返回原始URL
	if allowCORS {
		return resolvedURL
	}

	// 否则返回代理URL
	encodedURL := url.QueryEscape(resolvedURL)
	return fmt.Sprintf("%s/%s?url=%s", proxyBase, urlType, encodedURL)
}

// rewriteMapURI 重写 EXT-X-MAP 中的 URI
func (s *m3u8ProxyService) rewriteMapURI(line, baseURL, proxyBase string, allowCORS bool) string {
	re := regexp.MustCompile(`URI="([^"]+)"`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return line
	}

	originalURI := matches[1]
	rewrittenURL := s.rewriteURL(originalURI, baseURL, proxyBase, "segment", allowCORS)
	return re.ReplaceAllString(line, fmt.Sprintf(`URI="%s"`, rewrittenURL))
}

// rewriteKeyURI 重写 EXT-X-KEY 中的 URI
func (s *m3u8ProxyService) rewriteKeyURI(line, baseURL, proxyBase string, allowCORS bool) string {
	re := regexp.MustCompile(`URI="([^"]+)"`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 2 {
		return line
	}

	originalURI := matches[1]
	rewrittenURL := s.rewriteURL(originalURI, baseURL, proxyBase, "key", allowCORS)
	return re.ReplaceAllString(line, fmt.Sprintf(`URI="%s"`, rewrittenURL))
}

// resolveURL 将相对URL解析为绝对URL
func (s *m3u8ProxyService) resolveURL(baseURL, relativeURL string) string {
	// 处理以 / 开头的绝对路径
	if strings.HasPrefix(relativeURL, "/") {
		u, err := url.Parse(baseURL)
		if err != nil {
			return baseURL + relativeURL
		}
		return fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, relativeURL)
	}

	// 处理相对路径
	return baseURL + relativeURL
}
