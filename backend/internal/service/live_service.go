// internal/service/live_service.go
package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
)

// LiveService 直播服务接口
type LiveService interface {
	GetSources(ctx context.Context) ([]LiveSource, error)
	GetChannels(ctx context.Context, sourceKey string) (*LiveChannels, error)
	GetEPG(ctx context.Context, sourceKey string, tvgID string) ([]EPGItem, error)
	Precheck(ctx context.Context, url string) (bool, error)
}

// LiveSource 直播源
type LiveSource struct {
	Key           string `json:"key"`
	Name          string `json:"name"`
	URL           string `json:"url"`
	UA            string `json:"ua"`
	EPG           string `json:"epg,omitempty"`
	From          string `json:"from"`
	ChannelNumber int    `json:"channel_number,omitempty"`
	Disabled      bool   `json:"disabled"`
}

// LiveChannel 直播频道
type LiveChannel struct {
	ID    string `json:"id"`
	TvgID string `json:"tvg_id"`
	Name  string `json:"name"`
	Logo  string `json:"logo"`
	Group string `json:"group"`
	URL   string `json:"url"`
}

// LiveChannels 直播频道列表
type LiveChannels struct {
	ChannelNumber int           `json:"channel_number"`
	Channels      []LiveChannel `json:"channels"`
	EPGUrl        string        `json:"epg_url"`
	EPGs          map[string][]EPGItem
}

// EPGItem 节目单条目
type EPGItem struct {
	Start string `json:"start"`
	End   string `json:"end"`
	Title string `json:"title"`
}

// liveService 直播服务实现
type liveService struct {
	logger       *zap.Logger
	httpClient   *http.Client
	adminStorage model.AdminStorageService
	cache        map[string]*LiveChannels
}

// NewLiveService 创建直播服务
func NewLiveService(adminStorage model.AdminStorageService, logger *zap.Logger) LiveService {
	return &liveService{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		adminStorage: adminStorage,
		cache:        make(map[string]*LiveChannels),
	}
}

// GetSources 获取直播源列表
func (s *liveService) GetSources(ctx context.Context) ([]LiveSource, error) {
	return s.loadSources(ctx)
}

// GetChannels 获取频道列表
func (s *liveService) GetChannels(ctx context.Context, sourceKey string) (*LiveChannels, error) {
	// 检查缓存
	if cached, ok := s.cache[sourceKey]; ok {
		return cached, nil
	}

	allSources, err := s.loadSources(ctx)
	if err != nil {
		return nil, err
	}

	// 查找源配置
	var source *LiveSource
	for i := range allSources {
		if allSources[i].Key == sourceKey {
			source = &allSources[i]
			break
		}
	}
	if source == nil {
		return nil, fmt.Errorf("直播源不存在: %s", sourceKey)
	}

	// 获取 M3U 数据
	channels, err := s.fetchChannels(source)
	if err != nil {
		return nil, err
	}

	s.cache[sourceKey] = channels
	return channels, nil
}

// GetEPG 获取节目单
func (s *liveService) GetEPG(ctx context.Context, sourceKey string, tvgID string) ([]EPGItem, error) {
	channels, err := s.GetChannels(ctx, sourceKey)
	if err != nil {
		return nil, err
	}

	if channels.EPGs == nil {
		return []EPGItem{}, nil
	}

	if epg, ok := channels.EPGs[tvgID]; ok {
		return epg, nil
	}

	return []EPGItem{}, nil
}

// Precheck 预检查直播源
func (s *liveService) Precheck(ctx context.Context, url string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("User-Agent", "AptvPlayer/1.4.10")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// fetchChannels 获取频道列表
func (s *liveService) fetchChannels(source *LiveSource) (*LiveChannels, error) {
	req, err := http.NewRequest(http.MethodGet, source.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	ua := source.UA
	if ua == "" {
		ua = "AptvPlayer/1.4.10"
	}
	req.Header.Set("User-Agent", ua)

	resp, err := s.httpClient.Do(req)
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

	// 解析 M3U
	channels := s.parseM3U(source.Key, string(body))

	// 获取 EPG
	epgURL := source.EPG
	if epgURL == "" {
		epgURL = channels.TvgUrl
	}

	var epgs map[string][]EPGItem
	if epgURL != "" {
		epgs = s.parseEPG(epgURL, ua, channels)
	}

	return &LiveChannels{
		ChannelNumber: len(channels.Channels),
		Channels:      channels.Channels,
		EPGUrl:        epgURL,
		EPGs:          epgs,
	}, nil
}

// parseM3U 解析 M3U 文件
type parsedM3U struct {
	TvgUrl   string
	Channels []LiveChannel
}

func (s *liveService) parseM3U(sourceKey string, content string) *parsedM3U {
	var channels []LiveChannel
	var tvgURL string

	scanner := bufio.NewScanner(strings.NewReader(content))
	channelIndex := 0
	var currentChannel LiveChannel

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 解析 #EXTM3U 行
		if strings.HasPrefix(line, "#EXTM3U") {
			// 提取 tvg-url
			re := regexp.MustCompile(`(?:x-tvg-url|url-tvg)="([^"]*)"`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				// 取第一个 URL
				tvgURL = strings.Split(matches[1], ",")[0]
				tvgURL = strings.TrimSpace(tvgURL)
			}
			continue
		}

		// 解析 #EXTINF 行
		if strings.HasPrefix(line, "#EXTINF:") {
			// 提取 tvg-id
			re := regexp.MustCompile(`tvg-id="([^"]*)"`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentChannel.TvgID = matches[1]
			}

			// 提取 tvg-name
			re = regexp.MustCompile(`tvg-name="([^"]*)"`)
			matches = re.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentChannel.Name = matches[1]
			}

			// 提取 tvg-logo
			re = regexp.MustCompile(`tvg-logo="([^"]*)"`)
			matches = re.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentChannel.Logo = matches[1]
			}

			// 提取 group-title
			re = regexp.MustCompile(`group-title="([^"]*)"`)
			matches = re.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentChannel.Group = matches[1]
			} else {
				currentChannel.Group = "无分组"
			}

			// 提取标题 (逗号后的内容)
			if idx := strings.LastIndex(line, ","); idx != -1 {
				title := strings.TrimSpace(line[idx+1:])
				if currentChannel.Name == "" {
					currentChannel.Name = title
				}
			}

			continue
		}

		// 非注释行，认为是 URL
		if !strings.HasPrefix(line, "#") && currentChannel.Name != "" {
			currentChannel.URL = line
			currentChannel.ID = fmt.Sprintf("%s-%d", sourceKey, channelIndex)
			channels = append(channels, currentChannel)
			channelIndex++
			currentChannel = LiveChannel{}
		}
	}

	return &parsedM3U{
		TvgUrl:   tvgURL,
		Channels: channels,
	}
}

// parseEPG 解析 EPG 文件
func (s *liveService) parseEPG(epgURL string, ua string, m3u *parsedM3U) map[string][]EPGItem {
	result := make(map[string][]EPGItem)

	// 构建 tvgID 集合
	tvgSet := make(map[string]bool)
	for _, ch := range m3u.Channels {
		if ch.TvgID != "" {
			tvgSet[ch.TvgID] = true
		}
	}

	req, err := http.NewRequest(http.MethodGet, epgURL, nil)
	if err != nil {
		s.logger.Warn("创建EPG请求失败", zap.Error(err))
		return result
	}

	req.Header.Set("User-Agent", ua)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Warn("获取EPG失败", zap.Error(err))
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Warn("EPG请求失败", zap.Int("status", resp.StatusCode))
		return result
	}

	scanner := bufio.NewScanner(resp.Body)
	var currentTvgID string
	var currentEPG *EPGItem

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 解析 <programme> 标签
		if strings.HasPrefix(line, "<programme") {
			// 提取 channel
			re := regexp.MustCompile(`channel="([^"]*)"`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentTvgID = matches[1]
			}

			// 只处理我们关心的频道
			if !tvgSet[currentTvgID] {
				currentTvgID = ""
				continue
			}

			// 提取 start
			re = regexp.MustCompile(`start="([^"]*)"`)
			matches = re.FindStringSubmatch(line)
			start := ""
			if len(matches) > 1 {
				start = matches[1]
			}

			// 提取 stop
			re = regexp.MustCompile(`stop="([^"]*)"`)
			matches = re.FindStringSubmatch(line)
			stop := ""
			if len(matches) > 1 {
				stop = matches[1]
			}

			if start != "" && stop != "" {
				currentEPG = &EPGItem{
					Start: start,
					End:   stop,
				}
			}
			continue
		}

		// 解析 </programme>
		if line == "</programme>" {
			currentTvgID = ""
			currentEPG = nil
			continue
		}

		// 解析 <title>
		if strings.HasPrefix(line, "<title") && currentEPG != nil && currentTvgID != "" {
			re := regexp.MustCompile(`>([^<]*)<`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentEPG.Title = matches[1]
				result[currentTvgID] = append(result[currentTvgID], *currentEPG)
			}
		}
	}

	return result
}

func (s *liveService) loadSources(ctx context.Context) ([]LiveSource, error) {
	if s.adminStorage == nil {
		return []LiveSource{}, nil
	}

	legacySources, err := s.adminStorage.GetLiveSources(ctx)
	if err != nil {
		return nil, err
	}

	sources := make([]LiveSource, 0, len(legacySources))
	for _, source := range legacySources {
		sources = append(sources, LiveSource{
			Key:           source.Key,
			Name:          source.Name,
			URL:           source.URL,
			UA:            source.UA,
			EPG:           source.EPG,
			From:          source.From,
			ChannelNumber: source.ChannelNumber,
			Disabled:      source.Disabled,
		})
	}

	return sources, nil
}
