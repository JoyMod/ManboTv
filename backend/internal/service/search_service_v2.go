// internal/service/search_service_v2.go
// 基于 Channel 的高性能并发搜索服务

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/config"
	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// SearchServiceV2 基于Channel的搜索服务
type SearchServiceV2 interface {
	Search(ctx context.Context, query string, sites []model.ApiSite) ([]model.SearchResult, error)
}

// searchServiceV2 实现
type searchServiceV2 struct {
	client      *http.Client
	config      *config.SearchConfig
	logger      *zap.Logger
	semaphore   chan struct{} // 信号量channel控制并发
}

// NewSearchServiceV2 创建基于Channel的搜索服务
func NewSearchServiceV2(cfg *config.SearchConfig, httpCfg *config.HTTPClientConfig, logger *zap.Logger) SearchServiceV2 {
	client := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        httpCfg.MaxIdleConns,
			MaxIdleConnsPerHost: httpCfg.MaxIdleConnsPerHost,
			IdleConnTimeout:     httpCfg.IdleConnTimeout,
		},
	}

	return &searchServiceV2{
		client:    client,
		config:    cfg,
		logger:    logger,
		semaphore: make(chan struct{}, cfg.MaxConcurrent), // Channel作为信号量
	}
}

// Search 多源聚合搜索 - Channel版
func (s *searchServiceV2) Search(ctx context.Context, query string, sites []model.ApiSite) ([]model.SearchResult, error) {
	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	// Channel收集结果
	resultChan := make(chan []model.SearchResult, len(sites))
	errChan := make(chan error, len(sites))
	
	// 原子计数器
	var completed int32
	total := int32(len(sites))

	// 启动worker
	for _, site := range sites {
		site := site // 捕获
		
		// 使用semaphore控制并发数
		select {
		case s.semaphore <- struct{}{}: // 获取信号量
			go func() {
				defer func() { <-s.semaphore }() // 释放信号量
				
				results, err := s.searchSingle(ctx, site, query)
				
				atomic.AddInt32(&completed, 1)
				
				if err != nil {
					s.logger.Warn("搜索源失败",
						zap.String("source", site.Name),
						zap.Error(err),
					)
					errChan <- err
					return
				}
				
				select {
				case resultChan <- results:
				case <-ctx.Done():
				}
			}()
			
		case <-ctx.Done():
			// 上下文取消，不再启动新worker
			continue
		}
	}

	// 收集结果
	var allResults []model.SearchResult
	var errors []error
	
	// 等待所有worker完成或使用channel收集
	for atomic.LoadInt32(&completed) < total {
		select {
		case results := <-resultChan:
			allResults = append(allResults, results...)
			
		case err := <-errChan:
			errors = append(errors, err)
			
		case <-ctx.Done():
			s.logger.Warn("搜索超时",
				zap.String("query", query),
				zap.Int("completed", int(atomic.LoadInt32(&completed))),
				zap.Int("total", int(total)),
			)
			goto DONE
		}
	}

DONE:
	// 最后收集剩余结果
	for {
		select {
		case results := <-resultChan:
			allResults = append(allResults, results...)
		case <-errChan:
		default:
			goto FINISH
		}
	}

FINISH:
	s.logger.Info("搜索完成",
		zap.String("query", query),
		zap.Int("total_results", len(allResults)),
		zap.Int("completed_sources", int(atomic.LoadInt32(&completed))),
		zap.Int("errors", len(errors)),
	)

	return allResults, nil
}

// searchSingle 单源搜索
func (s *searchServiceV2) searchSingle(ctx context.Context, site model.ApiSite, query string) ([]model.SearchResult, error) {
	apiURL := fmt.Sprintf("%s?ac=videolist&wd=%s",
		site.API,
		url.QueryEscape(query),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")
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

	results := s.parseResults(apiResp.List, site)
	return results, nil
}

// parseResults 解析结果
func (s *searchServiceV2) parseResults(items []model.ApiSearchItem, site model.ApiSite) []model.SearchResult {
	results := make([]model.SearchResult, 0, len(items))

	for _, item := range items {
		episodes, titles := s.parseEpisodes(item.VodPlayURL)
		if len(episodes) == 0 {
			continue
		}

		results = append(results, model.SearchResult{
			ID:             item.VodID,
			Title:          strings.TrimSpace(item.VodName),
			Poster:         item.VodPic,
			Episodes:       episodes,
			EpisodesTitles: titles,
			Source:         site.Key,
			SourceName:     site.Name,
			Class:          item.VodClass,
			Year:           s.extractYear(item.VodYear),
			Desc:           item.VodContent,
			TypeName:       item.TypeName,
		})
	}

	return results
}

// parseEpisodes 解析剧集
func (s *searchServiceV2) parseEpisodes(vodPlayURL string) ([]string, []string) {
	if vodPlayURL == "" {
		return nil, nil
	}

	var bestEpisodes []string
	var bestTitles []string

	sources := strings.Split(vodPlayURL, model.EpisodeURLSeparator)

	for _, source := range sources {
		var episodes []string
		var titles []string

		items := strings.Split(source, model.EpisodeItemSeparator)

		for _, item := range items {
			parts := strings.Split(item, model.EpisodeTitleURLSeparator)
			if len(parts) != 2 {
				continue
			}

			if strings.HasSuffix(parts[1], model.M3U8Suffix) {
				titles = append(titles, parts[0])
				episodes = append(episodes, parts[1])
			}
		}

		if len(episodes) > len(bestEpisodes) {
			bestEpisodes = episodes
			bestTitles = titles
		}
	}

	return bestEpisodes, bestTitles
}

// extractYear 提取年份
func (s *searchServiceV2) extractYear(year string) string {
	if year == "" {
		return "unknown"
	}
	
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
