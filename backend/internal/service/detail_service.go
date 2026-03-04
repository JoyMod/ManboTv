// internal/service/detail_service.go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/JoyMod/ManboTV/backend/internal/config"
	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// DetailService 详情服务接口
type DetailService interface {
	GetDetail(ctx context.Context, site model.ApiSite, vodID string) (*model.SearchResult, error)
	GetDetails(ctx context.Context, sites []model.ApiSite, vodID string) ([]*model.SearchResult, error)
}

// detailService 详情服务实现
type detailService struct {
	client     *http.Client
	config     *config.SearchConfig
	logger     *zap.Logger
	m3u8Regex  *regexp.Regexp
}

// NewDetailService 创建详情服务
func NewDetailService(cfg *config.SearchConfig, httpCfg *config.HTTPClientConfig, logger *zap.Logger) DetailService {
	client := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        httpCfg.MaxIdleConns,
			MaxIdleConnsPerHost: httpCfg.MaxIdleConnsPerHost,
			IdleConnTimeout:     httpCfg.IdleConnTimeout,
		},
	}

	// 预编译正则表达式
	m3u8Regex := regexp.MustCompile(`\$(https?://[^"'\s]+?\.m3u8)`)

	return &detailService{
		client:    client,
		config:    cfg,
		logger:    logger,
		m3u8Regex: m3u8Regex,
	}
}

// GetDetail 获取单个源的详情
func (s *detailService) GetDetail(ctx context.Context, site model.ApiSite, vodID string) (*model.SearchResult, error) {
	// 特殊源处理
	if site.Detail != "" {
		return s.getSpecialDetail(ctx, site, vodID)
	}

	// 标准API获取详情
	detailURL := fmt.Sprintf("%s?ac=videolist&ids=%s", site.API, vodID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, detailURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求详情失败: %w", err)
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

	if len(apiResp.List) == 0 {
		return nil, fmt.Errorf("未找到详情")
	}

	videoDetail := apiResp.List[0]
	episodes, titles := s.parseEpisodes(videoDetail.VodPlayURL)

	// 如果从播放URL没有解析到，尝试从内容中解析
	if len(episodes) == 0 && videoDetail.VodContent != "" {
		matches := s.m3u8Regex.FindAllString(videoDetail.VodContent, -1)
		for _, match := range matches {
			episodes = append(episodes, strings.TrimPrefix(match, "$"))
		}
	}

	return &model.SearchResult{
		ID:             vodID,
		Title:          strings.TrimSpace(videoDetail.VodName),
		Poster:         videoDetail.VodPic,
		Episodes:       episodes,
		EpisodesTitles: titles,
		Source:         site.Key,
		SourceName:     site.Name,
		Class:          videoDetail.VodClass,
		Year:           s.extractYear(videoDetail.VodYear),
		Desc:           s.cleanHTML(videoDetail.VodContent),
		TypeName:       videoDetail.TypeName,
		DoubanID:       videoDetail.VodDoubanID,
	}, nil
}

// GetDetails 从多个源获取详情
func (s *detailService) GetDetails(ctx context.Context, sites []model.ApiSite, vodID string) ([]*model.SearchResult, error) {
	var results []*model.SearchResult
	var mu sync.Mutex

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(5) // 限制并发数

	for _, site := range sites {
		site := site
		g.Go(func() error {
			detail, err := s.GetDetail(ctx, site, vodID)
			if err != nil {
				s.logger.Warn("获取详情失败",
					zap.String("site", site.Name),
					zap.String("vod_id", vodID),
					zap.Error(err),
				)
				return nil // 单个源失败不中断
			}

			mu.Lock()
			results = append(results, detail)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

// getSpecialDetail 处理特殊源详情
func (s *detailService) getSpecialDetail(ctx context.Context, site model.ApiSite, vodID string) (*model.SearchResult, error) {
	detailURL := fmt.Sprintf("%s/index.php/vod/detail/id/%s.html", site.Detail, vodID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, detailURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	html := string(body)

	// 特殊正则处理（ffzy源）
	var pattern *regexp.Regexp
	if site.Key == "ffzy" {
		pattern = regexp.MustCompile(`\$(https?://[^"'\s]+?/\d{8}/\d+_[a-f0-9]+/index\.m3u8)`)
	} else {
		pattern = s.m3u8Regex
	}

	matches := pattern.FindAllString(html, -1)
	var episodes []string
	seen := make(map[string]bool)

	for _, match := range matches {
		url := strings.TrimPrefix(match, "$")
		// 去重
		parenIdx := strings.Index(url, "(")
		if parenIdx > 0 {
			url = url[:parenIdx]
		}
		if !seen[url] {
			seen[url] = true
			episodes = append(episodes, url)
		}
	}

	// 提取标题
	titleRegex := regexp.MustCompile(`<h1[^>]*>([^<]+)</h1>`)
	titleMatch := titleRegex.FindStringSubmatch(html)
	title := ""
	if len(titleMatch) > 1 {
		title = strings.TrimSpace(titleMatch[1])
	}

	// 提取描述
	descRegex := regexp.MustCompile(`<div[^>]*class=["']sketch["'][^>]*>([\s\S]*?)</div>`)
	descMatch := descRegex.FindStringSubmatch(html)
	desc := ""
	if len(descMatch) > 1 {
		desc = s.cleanHTML(descMatch[1])
	}

	// 提取封面
	coverRegex := regexp.MustCompile(`(https?://[^"'\s]+?\.jpg)`)
	coverMatch := coverRegex.FindString(html)

	// 提取年份
	yearRegex := regexp.MustCompile(`>(\d{4})<`)
	yearMatch := yearRegex.FindStringSubmatch(html)
	year := "unknown"
	if len(yearMatch) > 1 {
		year = yearMatch[1]
	}

	return &model.SearchResult{
		ID:             vodID,
		Title:          title,
		Poster:         coverMatch,
		Episodes:       episodes,
		EpisodesTitles: s.generateEpisodeTitles(len(episodes)),
		Source:         site.Key,
		SourceName:     site.Name,
		Year:           year,
		Desc:           desc,
	}, nil
}

// parseEpisodes 解析剧集
func (s *detailService) parseEpisodes(vodPlayURL string) ([]string, []string) {
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

			title := parts[0]
			url := parts[1]

			if strings.HasSuffix(url, model.M3U8Suffix) {
				titles = append(titles, title)
				episodes = append(episodes, url)
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
func (s *detailService) extractYear(year string) string {
	if year == "" {
		return "unknown"
	}

	yearRegex := regexp.MustCompile(`\d{4}`)
	match := yearRegex.FindString(year)
	if match != "" {
		return match
	}

	return "unknown"
}

// cleanHTML 清理HTML标签
func (s *detailService) cleanHTML(html string) string {
	// 移除HTML标签
	re := regexp.MustCompile(`<[^>]+>`)
	text := re.ReplaceAllString(html, "")

	// 解码HTML实体
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&quot;", "\"")

	return strings.TrimSpace(text)
}

// generateEpisodeTitles 生成剧集标题
func (s *detailService) generateEpisodeTitles(count int) []string {
	titles := make([]string, count)
	for i := 0; i < count; i++ {
		titles[i] = fmt.Sprintf("%d", i+1)
	}
	return titles
}
