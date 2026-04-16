// internal/model/search.go
package model

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type FlexibleString string

func (s *FlexibleString) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		*s = ""
		return nil
	}

	if trimmed[0] == '"' {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		*s = FlexibleString(strings.TrimSpace(value))
		return nil
	}

	*s = FlexibleString(strings.Trim(trimmed, "\""))
	return nil
}

func (s FlexibleString) String() string {
	return string(s)
}

type FlexibleInt int

func (v *FlexibleInt) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" || trimmed == "null" {
		*v = 0
		return nil
	}

	if trimmed[0] == '"' {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		value = strings.TrimSpace(value)
		if value == "" {
			*v = 0
			return nil
		}
		number, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid int string: %s", value)
		}
		*v = FlexibleInt(number)
		return nil
	}

	number, err := strconv.Atoi(trimmed)
	if err != nil {
		return err
	}
	*v = FlexibleInt(number)
	return nil
}

// SearchResult 搜索结果
type SearchResult struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Poster         string   `json:"poster"`
	Episodes       []string `json:"episodes"`
	EpisodesTitles []string `json:"episodes_titles"`
	Source         string   `json:"source"`
	SourceName     string   `json:"source_name"`
	Class          string   `json:"class,omitempty"`
	Year           string   `json:"year"`
	Desc           string   `json:"desc,omitempty"`
	TypeName       string   `json:"type_name,omitempty"`
	DoubanID       int      `json:"douban_id,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	IsAdult        bool     `json:"is_adult,omitempty"`
	Remarks        string   `json:"remarks,omitempty"`
	MatchScore     float64  `json:"match_score,omitempty"`
	MatchReasons   []string `json:"match_reasons,omitempty"`
}

// ApiSite API站点配置
type ApiSite struct {
	Key    string `json:"key"`
	API    string `json:"api"`
	Name   string `json:"name"`
	Detail string `json:"detail,omitempty"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query       string `form:"q" binding:"required"`
	Page        int    `form:"page,default=1" binding:"min=1"`
	PageSize    int    `form:"page_size,default=20" binding:"min=1,max=120"`
	View        string `form:"view"`
	Sort        string `form:"sort"`
	Types       string `form:"types"`
	Sources     string `form:"sources"`
	YearFrom    int    `form:"year_from"`
	YearTo      int    `form:"year_to"`
	SourceMode  string `form:"source_mode"`
	PreferExact bool   `form:"prefer_exact"`
	Stream      bool   `form:"stream"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	List       []SearchResult `json:"list"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

type SearchFacetBucket struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type SearchFacets struct {
	Types   []SearchFacetBucket `json:"types"`
	Sources []SearchFacetBucket `json:"sources"`
	Years   []SearchFacetBucket `json:"years"`
}

type SearchSourceStatus struct {
	Source      string `json:"source"`
	SourceName  string `json:"source_name"`
	Status      string `json:"status"`
	ResultCount int    `json:"result_count"`
	PageCount   int    `json:"page_count"`
	ElapsedMs   int64  `json:"elapsed_ms"`
	Error       string `json:"error,omitempty"`
}

type SearchAggregateResult struct {
	Key            string         `json:"key"`
	Title          string         `json:"title"`
	Year           string         `json:"year"`
	Type           string         `json:"type"`
	Cover          string         `json:"cover"`
	Rating         string         `json:"rating,omitempty"`
	SourceCount    int            `json:"source_count"`
	ResultCount    int            `json:"result_count"`
	BestSource     string         `json:"best_source,omitempty"`
	BestSourceName string         `json:"best_source_name,omitempty"`
	Tags           []string       `json:"tags,omitempty"`
	Items          []SearchResult `json:"items"`
}

type SearchExecutionInfo struct {
	Query            string `json:"query"`
	NormalizedQuery  string `json:"normalized_query"`
	CompletedSources int    `json:"completed_sources"`
	TotalSources     int    `json:"total_sources"`
	ElapsedMs        int64  `json:"elapsed_ms"`
	Degraded         bool   `json:"degraded"`
	StreamingEnabled bool   `json:"streaming_enabled"`
}

type SearchEnvelope struct {
	Query            string                  `json:"query"`
	NormalizedQuery  string                  `json:"normalized_query"`
	Results          []SearchResult          `json:"results"`
	Aggregates       []SearchAggregateResult `json:"aggregates"`
	Facets           SearchFacets            `json:"facets"`
	SourceStatus     []SearchSourceStatus    `json:"source_status_items"`
	LegacySourceMap  map[string]string       `json:"source_status"`
	PageInfo         PageInfo                `json:"page_info"`
	Execution        SearchExecutionInfo     `json:"execution"`
	SelectedTypes    []string                `json:"selected_types,omitempty"`
	SelectedSources  []string                `json:"selected_sources,omitempty"`
	SelectedSort     string                  `json:"selected_sort,omitempty"`
	SelectedView     string                  `json:"selected_view,omitempty"`
	SelectedYearFrom int                     `json:"selected_year_from,omitempty"`
	SelectedYearTo   int                     `json:"selected_year_to,omitempty"`
	SelectedMode     string                  `json:"selected_source_mode,omitempty"`
}

// ApiSearchItem 第三方API返回的原始数据
type ApiSearchItem struct {
	VodID       FlexibleString `json:"vod_id"`
	VodName     string         `json:"vod_name"`
	VodPic      string         `json:"vod_pic"`
	VodRemarks  string         `json:"vod_remarks,omitempty"`
	VodPlayURL  string         `json:"vod_play_url,omitempty"`
	VodClass    string         `json:"vod_class,omitempty"`
	VodYear     string         `json:"vod_year,omitempty"`
	VodContent  string         `json:"vod_content,omitempty"`
	VodDoubanID FlexibleInt    `json:"vod_douban_id,omitempty"`
	TypeName    string         `json:"type_name,omitempty"`
}

// ApiSearchResponse 第三方API返回结构
type ApiSearchResponse struct {
	Code      FlexibleInt     `json:"code"`
	Msg       string          `json:"msg"`
	PageCount FlexibleInt     `json:"pagecount"`
	List      []ApiSearchItem `json:"list"`
}

// 搜索相关常量
const (
	EpisodeURLSeparator      = "$$$"
	EpisodeItemSeparator     = "#"
	EpisodeTitleURLSeparator = "$"
	M3U8Suffix               = ".m3u8"
)
