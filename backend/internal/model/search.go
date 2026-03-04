// internal/model/search.go
package model

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
	Query    string `form:"q" binding:"required"`
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=20" binding:"min=1,max=50"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	List       []SearchResult `json:"list"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// ApiSearchItem 第三方API返回的原始数据
type ApiSearchItem struct {
	VodID       string `json:"vod_id"`
	VodName     string `json:"vod_name"`
	VodPic      string `json:"vod_pic"`
	VodRemarks  string `json:"vod_remarks,omitempty"`
	VodPlayURL  string `json:"vod_play_url,omitempty"`
	VodClass    string `json:"vod_class,omitempty"`
	VodYear     string `json:"vod_year,omitempty"`
	VodContent  string `json:"vod_content,omitempty"`
	VodDoubanID int    `json:"vod_douban_id,omitempty"`
	TypeName    string `json:"type_name,omitempty"`
}

// ApiSearchResponse 第三方API返回结构
type ApiSearchResponse struct {
	Code      int             `json:"code"`
	Msg       string          `json:"msg"`
	PageCount int             `json:"pagecount"`
	List      []ApiSearchItem `json:"list"`
}

// 搜索相关常量
const (
	EpisodeURLSeparator       = "$$$"
	EpisodeItemSeparator      = "#"
	EpisodeTitleURLSeparator  = "$"
	M3U8Suffix               = ".m3u8"
)
