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
	client      *http.Client
	cache       *lru.Cache[string, *CacheItem]
	logger      *zap.Logger
	userAgent   string
	cacheMaxSize int64
}

// CacheItem 缓存项
type CacheItem struct {
	Data        []byte
	ContentType string
	Timestamp   time.Time
}

// NewImageService 创建图片代理服务
func NewImageService(cfg *config.ImageProxyConfig, httpCfg *config.HTTPClientConfig, logger *zap.Logger) (ImageService, error) {
	client := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        httpCfg.MaxIdleConns,
			MaxIdleConnsPerHost: httpCfg.MaxIdleConnsPerHost,
			IdleConnTimeout:     httpCfg.IdleConnTimeout,
		},
	}

	// 创建 LRU 缓存
	cache, err := lru.New[string, *CacheItem](cfg.CacheSize)
	if err != nil {
		return nil, fmt.Errorf("创建缓存失败: %w", err)
	}

	return &imageService{
		client:       client,
		cache:        cache,
		logger:       logger,
		userAgent:    cfg.UserAgent,
		cacheMaxSize: cfg.CacheMaxItemSize,
	}, nil
}

// Proxy 代理图片请求
func (s *imageService) Proxy(ctx context.Context, imageURL string) (*ImageResponse, error) {
	// 检查缓存
	if cached, ok := s.cache.Get(imageURL); ok {
		s.logger.Debug("图片缓存命中",
			zap.String("url", imageURL),
			zap.Int("size", len(cached.Data)),
		)
		return &ImageResponse{
			Data:        cached.Data,
			ContentType: cached.ContentType,
			Size:        len(cached.Data),
			FromCache:   true,
		}, nil
	}

	// 发起请求
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

	// 读取响应
	contentType := resp.Header.Get("Content-Type")
	
	// 限制大小，防止内存溢出
	var data []byte
	if resp.ContentLength > 0 && resp.ContentLength > s.cacheMaxSize {
		// 大图片直接流式返回，不缓存
		data, err = io.ReadAll(io.LimitReader(resp.Body, s.cacheMaxSize))
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %w", err)
		}
		s.logger.Info("大图片直接返回(不缓存)",
			zap.String("url", imageURL),
			zap.Int64("content_length", resp.ContentLength),
		)
		return &ImageResponse{
			Data:        data,
			ContentType: contentType,
			Size:        len(data),
			FromCache:   false,
		}, nil
	}

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	elapsed := time.Since(start)

	// 缓存小图片
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
