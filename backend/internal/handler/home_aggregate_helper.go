package handler

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

const (
	homeSectionPageSize     = 18
	homeSectionPageStart    = 0
	homeFeaturedBannerLimit = 3
)

type homeBannerPayload struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle,omitempty"`
	Description string   `json:"description"`
	Backdrop    string   `json:"backdrop"`
	Rating      string   `json:"rating,omitempty"`
	Year        string   `json:"year,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type homeContentItem struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Cover  string `json:"cover"`
	Rating string `json:"rating,omitempty"`
	Year   string `json:"year,omitempty"`
	Type   string `json:"type,omitempty"`
	Source string `json:"source,omitempty"`
}

type homeSectionPayload struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Href        string            `json:"href,omitempty"`
	Items       []homeContentItem `json:"items"`
	ShowRanking bool              `json:"showRanking,omitempty"`
}

type homeSectionConfig struct {
	ID          string
	Title       string
	Description string
	Href        string
	APIType     string
	Tag         string
	ContentType string
	ShowRanking bool
}

var defaultHomeSections = []homeSectionConfig{
	{
		ID:          "hot-movie",
		Title:       "热门电影",
		Description: "先看当下讨论度最高、封面质量最稳定的一批电影。",
		Href:        "/movie",
		APIType:     "movie",
		Tag:         "热门",
		ContentType: "movie",
	},
	{
		ID:          "high-score-movie",
		Title:       "高分电影",
		Description: "偏口碑向，适合你不知道看什么时直接开。",
		Href:        "/movie?feature=%E8%B1%86%E7%93%A3%E9%AB%98%E5%88%86",
		APIType:     "movie",
		Tag:         "豆瓣高分",
		ContentType: "movie",
	},
	{
		ID:          "hot-tv",
		Title:       "热播剧集",
		Description: "以剧集资源为主，直接落到电视剧频道而不是泛搜索。",
		Href:        "/tv",
		APIType:     "tv",
		Tag:         "国产剧",
		ContentType: "tv",
		ShowRanking: true,
	},
	{
		ID:          "anime-picks",
		Title:       "动漫精选",
		Description: "优先聚合日本动画与新番内容，避免频道空转。",
		Href:        "/anime",
		APIType:     "tv",
		Tag:         "日本动画",
		ContentType: "anime",
	},
	{
		ID:          "variety-picks",
		Title:       "热门综艺",
		Description: "综艺单独成区，不再混在剧集和电影里。",
		Href:        "/variety",
		APIType:     "tv",
		Tag:         "综艺",
		ContentType: "variety",
	},
}

func buildHomeSections(
	ctx context.Context,
	client *http.Client,
	logger *zap.Logger,
) []homeSectionPayload {
	sections := make([]homeSectionPayload, len(defaultHomeSections))
	group, groupCtx := errgroup.WithContext(ctx)

	for index, config := range defaultHomeSections {
		index := index
		config := config
		group.Go(func() error {
			items, err := fetchDoubanAggregatePage(
				groupCtx,
				client,
				config.APIType,
				config.Tag,
				browseDefaultSort,
				homeSectionPageSize,
				homeSectionPageStart,
			)
			if err != nil {
				logger.Warn(
					"home section degraded",
					zap.String("section", config.ID),
					zap.String("tag", config.Tag),
					zap.Error(err),
				)
				return nil
			}

			sections[index] = homeSectionPayload{
				ID:          config.ID,
				Title:       config.Title,
				Description: config.Description,
				Href:        config.Href,
				Items:       mapHomeItems(items, config.ContentType),
				ShowRanking: config.ShowRanking,
			}
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		logger.Warn("build home sections interrupted", zap.Error(err))
	}

	filtered := make([]homeSectionPayload, 0, len(sections))
	for _, section := range sections {
		if len(section.Items) == 0 {
			continue
		}
		filtered = append(filtered, section)
	}

	return filtered
}

func buildHomeBanner(sections []homeSectionPayload) []homeBannerPayload {
	featured := make([]homeContentItem, 0, homeFeaturedBannerLimit)
	for _, section := range sections {
		if len(section.Items) == 0 {
			continue
		}
		featured = append(featured, section.Items[0])
		if len(featured) >= homeFeaturedBannerLimit {
			break
		}
	}

	if len(featured) == 0 {
		return []homeBannerPayload{
			{
				ID:          "1",
				Title:       "曼波TV 独家首播",
				Subtitle:    "热门大片 · 全网同步",
				Description: "汇聚全网最新最热影视资源，高清画质流畅播放。",
				Backdrop:    "https://images.unsplash.com/photo-1536440136628-849c177e76a1?w=1920&q=80",
				Rating:      "9.8分",
				Year:        "2026",
				Tags:        []string{"热门", "独家"},
			},
			{
				ID:          "2",
				Title:       "经典动漫剧场",
				Subtitle:    "热血回忆 · 青春不败",
				Description: "经典动漫作品全收录，重温那些年追过的热血与感动。",
				Backdrop:    "https://images.unsplash.com/photo-1578632767115-351597cf2477?w=1920&q=80",
				Rating:      "9.6分",
				Year:        "经典",
				Tags:        []string{"动漫", "经典"},
			},
		}
	}

	banners := make([]homeBannerPayload, 0, len(featured))
	for index, item := range featured {
		subtitle := "今日推荐"
		if index > 0 {
			subtitle = "频道精选"
		}
		tags := []string{"推荐", mapBannerTypeLabel(item.Type)}
		banners = append(banners, homeBannerPayload{
			ID:          item.ID,
			Title:       item.Title,
			Subtitle:    subtitle,
			Description: fmt.Sprintf("%s 已进入首页精选分区，适合直接开始浏览与播放。", item.Title),
			Backdrop:    item.Cover,
			Rating:      item.Rating,
			Year:        item.Year,
			Tags:        compactStrings(tags),
		})
	}

	return banners
}

func mapHomeItems(items []DoubanItem, contentType string) []homeContentItem {
	mapped := make([]homeContentItem, 0, len(items))
	for index, item := range items {
		mapped = append(mapped, homeContentItem{
			ID:     resolveAggregateID(item.ID, contentType, index),
			Title:  firstNonEmptyString(item.Title, "未知标题"),
			Cover:  firstNonEmptyString(item.Poster, "/placeholder-poster.svg"),
			Rating: item.Rate,
			Year:   item.Year,
			Type:   contentType,
			Source: "douban",
		})
	}
	return mapped
}

func mapBannerTypeLabel(kind string) string {
	switch kind {
	case "movie":
		return "电影"
	case "tv":
		return "剧集"
	case "anime":
		return "动漫"
	case "variety":
		return "综艺"
	default:
		return ""
	}
}

func compactStrings(values []string) []string {
	compacted := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		compacted = append(compacted, value)
	}
	return compacted
}
