// internal/service/search_service.go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/JoyMod/ManboTV/backend/internal/config"
	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// SearchService 搜索服务接口
type SearchService interface {
	Search(ctx context.Context, query string, sites []model.ApiSite) ([]model.SearchResult, error)
	SearchSingle(ctx context.Context, site model.ApiSite, query string) ([]model.SearchResult, error)
}

// searchService 搜索服务实现
type searchService struct {
	client     *http.Client
	config     *config.SearchConfig
	logger     *zap.Logger
	maxRetries int
}

// NewSearchService 创建搜索服务
func NewSearchService(cfg *config.SearchConfig, httpCfg *config.HTTPClientConfig, logger *zap.Logger) SearchService {
	client := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        httpCfg.MaxIdleConns,
			MaxIdleConnsPerHost: httpCfg.MaxIdleConnsPerHost,
			IdleConnTimeout:     httpCfg.IdleConnTimeout,
		},
	}

	return &searchService{
		client:     client,
		config:     cfg,
		logger:     logger,
		maxRetries: 3,
	}
}

// Search 多源聚合搜索
func (s *searchService) Search(ctx context.Context, query string, sites []model.ApiSite) ([]model.SearchResult, error) {
	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(s.config.MaxConcurrent)

	var mu sync.Mutex
	var results []model.SearchResult

	for _, site := range sites {
		site := site // 捕获循环变量
		g.Go(func() error {
			res, err := s.searchSingle(ctx, site, query)
			if err != nil {
				s.logger.Warn("搜索源失败",
					zap.String("source", site.Name),
					zap.Error(err),
				)
				return nil // 单个源失败不影响整体
			}

			mu.Lock()
			results = append(results, res...)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		s.logger.Error("聚合搜索失败", zap.Error(err))
		return nil, fmt.Errorf("聚合搜索失败: %w", err)
	}

	s.logger.Info("搜索完成",
		zap.String("query", query),
		zap.Int("source_count", len(sites)),
		zap.Int("result_count", len(results)),
	)

	return results, nil
}

// SearchSingle 单源搜索
func (s *searchService) SearchSingle(ctx context.Context, site model.ApiSite, query string) ([]model.SearchResult, error) {
	return s.searchSingle(ctx, site, query)
}

// searchSingle 内部单源搜索实现
func (s *searchService) searchSingle(ctx context.Context, site model.ApiSite, query string) ([]model.SearchResult, error) {
	searchURL := fmt.Sprintf("%s?ac=videolist&wd=%s",
		site.API,
		url.QueryEscape(query),
	)

	var results []model.SearchResult
	err := s.withRetry(ctx, func() error {
		res, err := s.doSearchRequest(ctx, searchURL)
		if err != nil {
			return err
		}
		results = res
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("搜索源 %s 失败: %w", site.Name, err)
	}

	// 设置来源信息
	for i := range results {
		results[i].Source = site.Key
		results[i].SourceName = site.Name
	}

	return results, nil
}

// doSearchRequest 执行搜索请求
func (s *searchService) doSearchRequest(ctx context.Context, searchURL string) ([]model.SearchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var apiResp model.ApiSearchResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	if apiResp.Code != 0 && apiResp.Msg != "" {
		return nil, fmt.Errorf("API错误: %s", apiResp.Msg)
	}

	return s.parseResults(apiResp.List), nil
}

// parseResults 解析搜索结果
func (s *searchService) parseResults(items []model.ApiSearchItem) []model.SearchResult {
	results := make([]model.SearchResult, 0, len(items))

	for _, item := range items {
		episodes, titles := s.parseEpisodes(item.VodPlayURL)
		if len(episodes) == 0 {
			continue // 过滤掉无播放链接的结果
		}

		result := model.SearchResult{
			ID:             item.VodID,
			Title:          strings.TrimSpace(item.VodName),
			Poster:         item.VodPic,
			Episodes:       episodes,
			EpisodesTitles: titles,
			Class:          item.VodClass,
			Year:           s.extractYear(item.VodYear),
			Desc:           item.VodContent,
			TypeName:       item.TypeName,
			DoubanID:       item.VodDoubanID,
		}

		results = append(results, result)
	}

	return results
}

// parseEpisodes 解析剧集链接
func (s *searchService) parseEpisodes(vodPlayURL string) ([]string, []string) {
	if vodPlayURL == "" {
		return nil, nil
	}

	var bestEpisodes []string
	var bestTitles []string

	// 按 $$$ 分割不同播放源
	sources := strings.Split(vodPlayURL, model.EpisodeURLSeparator)

	for _, source := range sources {
		var episodes []string
		var titles []string

		// 按 # 分割剧集
		items := strings.Split(source, model.EpisodeItemSeparator)

		for _, item := range items {
			// 按 $ 分割标题和链接
			parts := strings.Split(item, model.EpisodeTitleURLSeparator)
			if len(parts) != 2 {
				continue
			}

			title := parts[0]
			url := parts[1]

			// 只保留 m3u8 链接
			if strings.HasSuffix(url, model.M3U8Suffix) {
				titles = append(titles, title)
				episodes = append(episodes, url)
			}
		}

		// 选择剧集最多的播放源
		if len(episodes) > len(bestEpisodes) {
			bestEpisodes = episodes
			bestTitles = titles
		}
	}

	return bestEpisodes, bestTitles
}

// extractYear 提取年份
func (s *searchService) extractYear(year string) string {
	if year == "" {
		return "unknown"
	}

	// 尝试匹配4位数字年份
	for i := 0; i <= len(year)-4; i++ {
		if year[i] >= '0' && year[i] <= '9' &&
			year[i+1] >= '0' && year[i+1] <= '9' &&
			year[i+2] >= '0' && year[i+2] <= '9' &&
			year[i+3] >= '0' && year[i+3] <= '9' {
			return year[i : i+4]
		}
	}

	return "unknown"
}

// withRetry 带重试的执行
func (s *searchService) withRetry(ctx context.Context, fn func() error) error {
	var err error
	for i := 0; i < s.maxRetries; i++ {
		if err = fn(); err == nil {
			return nil
		}

		if i < s.maxRetries-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(i+1) * time.Second):
				// 指数退避
			}
		}
	}
	return err
}
