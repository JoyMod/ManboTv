// internal/service/segment_proxy_service.go
package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/config"
)

// SegmentProxyService 视频片段代理服务接口
type SegmentProxyService interface {
	ProxySegment(ctx context.Context, targetURL string, rangeHeader string) (*SegmentResponse, error)
}

// SegmentResponse 片段响应
type SegmentResponse struct {
	Body          io.ReadCloser
	StatusCode    int
	ContentType   string
	ContentLength int64
	ContentRange  string
	AcceptRanges  string
}

// segmentProxyService 视频片段代理服务实现
type segmentProxyService struct {
	client         *http.Client
	insecureClient *http.Client
	logger         *zap.Logger
	timeout        time.Duration
	userAgent      string
}

// NewSegmentProxyService 创建视频片段代理服务
func NewSegmentProxyService(cfg *config.HTTPClientConfig, logger *zap.Logger) SegmentProxyService {
	client, insecureClient := buildHTTPClients(cfg)

	return &segmentProxyService{
		client:         client,
		insecureClient: insecureClient,
		logger:         logger,
		timeout:        client.Timeout,
		userAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
	}
}

// ProxySegment 代理视频片段
func (s *segmentProxyService) ProxySegment(ctx context.Context, targetURL string, rangeHeader string) (*SegmentResponse, error) {
	// 解码URL
	decodedURL, err := url.QueryUnescape(targetURL)
	if err != nil {
		return nil, fmt.Errorf("解码URL失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, decodedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Accept", "*/*")
	if rangeHeader != "" {
		req.Header.Set("Range", rangeHeader)
	}

	// 执行请求
	resp, err := s.client.Do(req)
	if err != nil && shouldRetryWithoutTLSVerify(err) {
		resp, err = s.insecureClient.Do(req)
	}
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		resp.Body.Close()
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// 提取响应头信息
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "video/mp2t" // 默认TS片段类型
	}

	var contentLength int64
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		contentLength, _ = strconv.ParseInt(cl, 10, 64)
	}

	contentRange := resp.Header.Get("Content-Range")
	acceptRanges := resp.Header.Get("Accept-Ranges")

	s.logger.Debug("片段代理成功",
		zap.String("url", decodedURL),
		zap.Int("status_code", resp.StatusCode),
		zap.String("content_type", contentType),
		zap.Int64("content_length", contentLength),
	)

	return &SegmentResponse{
		Body:          resp.Body,
		StatusCode:    resp.StatusCode,
		ContentType:   contentType,
		ContentLength: contentLength,
		ContentRange:  contentRange,
		AcceptRanges:  acceptRanges,
	}, nil
}

// SegmentProxyStream 流式代理片段（用于支持流式传输）
type SegmentProxyStream struct {
	Reader        io.Reader
	ContentType   string
	ContentLength int64
	AcceptRanges  string
}

// ProxySegmentStream 流式代理视频片段
func (s *segmentProxyService) ProxySegmentStream(ctx context.Context, targetURL string) (*SegmentProxyStream, error) {
	resp, err := s.ProxySegment(ctx, targetURL, "")
	if err != nil {
		return nil, err
	}

	return &SegmentProxyStream{
		Reader:        resp.Body,
		ContentType:   resp.ContentType,
		ContentLength: resp.ContentLength,
		AcceptRanges:  resp.AcceptRanges,
	}, nil
}

// DefaultSegmentTimeout 默认片段代理超时
const DefaultSegmentTimeout = 30 * time.Second

// MaxSegmentSize 最大片段大小 (100MB)
const MaxSegmentSize = 100 * 1024 * 1024
