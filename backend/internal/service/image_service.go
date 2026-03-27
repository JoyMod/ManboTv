// internal/service/image_service.go
package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/golang-lru/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	"github.com/JoyMod/ManboTV/backend/internal/config"
)

const (
	defaultImageReferer = "https://movie.douban.com/"
	imageAcceptHeader   = "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8"
	htmlContentType     = "text/html"
	doubanImageHostMin  = 1
	doubanImageHostMax  = 9
)

// ImageService 图片代理服务接口
type ImageService interface {
	Proxy(ctx context.Context, imageURL string) (*ImageResponse, error)
	WarmCache(imageURLs []string)
}

// ImageResponse 图片响应
type ImageResponse struct {
	Data        []byte
	ContentType string
	Size        int
	FromCache   bool
}

// imageService 图片代理服务实现
type imageService struct {
	client       *http.Client
	cache        *lru.Cache[string, *CacheItem]
	logger       *zap.Logger
	userAgent    string
	timeout      time.Duration
	cacheMaxSize int64
	inflight     singleflight.Group
}

// CacheItem 缓存项
type CacheItem struct {
	Data        []byte
	ContentType string
	Timestamp   time.Time
}

// NewImageService 创建图片代理服务
func NewImageService(cfg *config.ImageProxyConfig, httpCfg *config.HTTPClientConfig, logger *zap.Logger) (ImageService, error) {
	if cfg == nil {
		cfg = &config.ImageProxyConfig{}
	}
	if httpCfg == nil {
		httpCfg = &config.HTTPClientConfig{}
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	cacheSize := cfg.CacheSize
	if cacheSize <= 0 {
		cacheSize = 1000
	}

	cacheMaxItemSize := cfg.CacheMaxItemSize
	if cacheMaxItemSize <= 0 {
		cacheMaxItemSize = 1 << 20
	}

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        httpCfg.MaxIdleConns,
			MaxIdleConnsPerHost: httpCfg.MaxIdleConnsPerHost,
			IdleConnTimeout:     httpCfg.IdleConnTimeout,
		},
	}

	// 创建 LRU 缓存
	cache, err := lru.New[string, *CacheItem](cacheSize)
	if err != nil {
		return nil, fmt.Errorf("创建缓存失败: %w", err)
	}

	return &imageService{
		client:       client,
		cache:        cache,
		logger:       logger,
		userAgent:    cfg.UserAgent,
		timeout:      timeout,
		cacheMaxSize: cacheMaxItemSize,
	}, nil
}

// Proxy 代理图片请求
func (s *imageService) Proxy(ctx context.Context, imageURL string) (*ImageResponse, error) {
	if cached, ok := s.getCachedImage(imageURL); ok {
		return cached, nil
	}

	result, err, _ := s.inflight.Do(imageURL, func() (interface{}, error) {
		// 双重检查，防止并发期间已被其他请求写入缓存
		if cached, ok := s.getCachedImage(imageURL); ok {
			return cached, nil
		}

		fetchCtx, cancel := context.WithTimeout(context.Background(), s.timeout)
		defer cancel()

		return s.fetchImage(fetchCtx, imageURL)
	})
	if err != nil {
		return nil, err
	}

	resp, ok := result.(*ImageResponse)
	if !ok {
		return nil, fmt.Errorf("图片响应类型错误")
	}

	return resp, nil
}

func (s *imageService) getCachedImage(imageURL string) (*ImageResponse, bool) {
	cached, ok := s.cache.Get(imageURL)
	if !ok {
		return nil, false
	}

	s.logger.Debug("图片缓存命中",
		zap.String("url", imageURL),
		zap.Int("size", len(cached.Data)),
	)

	return &ImageResponse{
		Data:        cached.Data,
		ContentType: cached.ContentType,
		Size:        len(cached.Data),
		FromCache:   true,
	}, true
}

func (s *imageService) fetchImage(ctx context.Context, imageURL string) (*ImageResponse, error) {
	attemptURLs, referers, err := buildImageRequestMeta(imageURL)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	resp, err := s.doImageRequest(ctx, attemptURLs, referers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(strings.ToLower(contentType), htmlContentType) {
		return nil, fmt.Errorf("返回了HTML内容")
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	elapsed := time.Since(start)

	if int64(len(data)) <= s.cacheMaxSize {
		s.cache.Add(imageURL, &CacheItem{
			Data:        data,
			ContentType: contentType,
			Timestamp:   time.Now(),
		})
		s.logger.Debug("图片已缓存",
			zap.String("url", imageURL),
			zap.Int("size", len(data)),
		)
	} else {
		s.logger.Info("大图片直出(不缓存)",
			zap.String("url", imageURL),
			zap.Int("size", len(data)),
			zap.Int64("cache_limit", s.cacheMaxSize),
		)
	}

	s.logger.Info("图片代理成功",
		zap.String("url", imageURL),
		zap.Int("size", len(data)),
		zap.Duration("elapsed", elapsed),
	)

	return &ImageResponse{
		Data:        data,
		ContentType: contentType,
		Size:        len(data),
		FromCache:   false,
	}, nil
}

// WarmCache 预热缓存
func (s *imageService) WarmCache(imageURLs []string) {
	for _, url := range imageURLs {
		if _, ok := s.cache.Get(url); !ok {
			// 异步预热
			go func(u string) {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				_, err := s.Proxy(ctx, u)
				if err != nil {
					s.logger.Warn("缓存预热失败", zap.String("url", u), zap.Error(err))
				}
			}(url)
		}
	}
}

// StreamProxy 流式代理（用于大图片）
func (s *imageService) StreamProxy(ctx context.Context, imageURL string, w io.Writer) error {
	// 检查缓存
	if cached, ok := s.cache.Get(imageURL); ok {
		_, err := w.Write(cached.Data)
		return err
	}

	attemptURLs, referers, err := buildImageRequestMeta(imageURL)
	if err != nil {
		return err
	}

	resp, err := s.doImageRequest(ctx, attemptURLs, referers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), htmlContentType) {
		return fmt.Errorf("返回了HTML内容")
	}

	// 流式复制
	_, err = io.Copy(w, resp.Body)
	return err
}

func (s *imageService) doImageRequest(
	ctx context.Context,
	attemptURLs []string,
	referers []string,
) (*http.Response, error) {
	var lastErr error
	for _, attemptURL := range attemptURLs {
		for _, referer := range referers {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, attemptURL, nil)
			if err != nil {
				lastErr = fmt.Errorf("创建请求失败: %w", err)
				continue
			}

			req.Header.Set("User-Agent", s.userAgent)
			req.Header.Set("Accept", imageAcceptHeader)
			req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
			if referer != "" {
				req.Header.Set("Referer", referer)
				req.Header.Set("Origin", strings.TrimSuffix(referer, "/"))
			}

			resp, doErr := s.client.Do(req)
			if doErr != nil {
				lastErr = fmt.Errorf("请求图片失败: %w", doErr)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
				resp.Body.Close()
				continue
			}

			if strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), htmlContentType) {
				lastErr = fmt.Errorf("返回了HTML内容")
				resp.Body.Close()
				continue
			}

			return resp, nil
		}
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("请求图片失败: 无可用响应")
	}
	return nil, lastErr
}

func buildImageRequestMeta(imageURL string) ([]string, []string, error) {
	parsedURL, err := url.Parse(strings.TrimSpace(imageURL))
	if err != nil || parsedURL.Host == "" {
		return nil, nil, fmt.Errorf("图片URL无效")
	}

	primaryURL := parsedURL.String()
	attemptURLs := buildAttemptImageURLs(parsedURL, primaryURL)
	switch parsedURL.Scheme {
	case "https":
		attemptURLs = append(attemptURLs, strings.Replace(primaryURL, "https://", "http://", 1))
	case "http":
		attemptURLs = append(attemptURLs, strings.Replace(primaryURL, "http://", "https://", 1))
	}

	hostReferer := fmt.Sprintf("%s://%s/", parsedURL.Scheme, parsedURL.Host)
	referers := []string{hostReferer}
	if hostReferer != defaultImageReferer {
		referers = append(referers, defaultImageReferer)
	}
	referers = append(referers, "")

	return dedupeStrings(attemptURLs), dedupeStrings(referers), nil
}

func buildAttemptImageURLs(parsedURL *url.URL, primaryURL string) []string {
	attemptURLs := []string{primaryURL}
	if !strings.HasSuffix(parsedURL.Host, ".doubanio.com") {
		return attemptURLs
	}

	for hostIndex := doubanImageHostMin; hostIndex <= doubanImageHostMax; hostIndex++ {
		candidate := *parsedURL
		candidate.Host = fmt.Sprintf("img%d.doubanio.com", hostIndex)
		attemptURLs = append(attemptURLs, candidate.String())
	}

	return attemptURLs
}

func dedupeStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			if _, exists := seen[trimmed]; exists {
				continue
			}
			seen[trimmed] = struct{}{}
			result = append(result, trimmed)
			continue
		}

		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}
