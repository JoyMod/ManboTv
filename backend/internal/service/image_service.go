// internal/service/image_service.go
package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/golang-lru/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	"github.com/JoyMod/ManboTV/backend/internal/config"
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Referer", "https://movie.douban.com/")
	req.Header.Set("Accept", "image/webp,image/apng,image/*,*/*;q=0.8")

	start := time.Now()
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求图片失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Referer", "https://movie.douban.com/")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求图片失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// 流式复制
	_, err = io.Copy(w, resp.Body)
	return err
}
