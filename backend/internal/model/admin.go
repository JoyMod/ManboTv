// internal/model/admin.go
// Admin 模块数据模型

package model

import "context"

const (
	ContentAccessModeSafe      = "safe"
	ContentAccessModeMixed     = "mixed"
	ContentAccessModeAdultOnly = "adult_only"
)

// VideoSource 视频源配置
type VideoSource struct {
	Key      string `json:"key" binding:"required"`
	Name     string `json:"name" binding:"required"`
	API      string `json:"api" binding:"required"`
	Detail   string `json:"detail,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
	From     string `json:"from,omitempty"` // config 或 custom
}

// LiveSource 直播源配置
type LiveSource struct {
	Key           string `json:"key" binding:"required"`
	Name          string `json:"name" binding:"required"`
	URL           string `json:"url" binding:"required"`
	UA            string `json:"ua,omitempty"`
	EPG           string `json:"epg,omitempty"`
	Disabled      bool   `json:"disabled,omitempty"`
	From          string `json:"from,omitempty"`
	ChannelNumber int    `json:"channel_number,omitempty"`
}

// CustomCategory 自定义分类
type CustomCategory struct {
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required"` // movie 或 tv
	Query    string `json:"query" binding:"required"`
	Disabled bool   `json:"disabled,omitempty"`
	From     string `json:"from,omitempty"`
}

// UserGroup 用户组配置
type UserGroup struct {
	Name        string   `json:"name" binding:"required"`
	EnabledAPIs []string `json:"enabled_apis,omitempty"`
}

// UserConfig 用户配置
type UserConfig struct {
	Username     string   `json:"username" binding:"required"`
	PasswordHash string   `json:"password_hash,omitempty"`
	Role         string   `json:"role" binding:"required"` // owner, admin, user
	Banned       bool     `json:"banned,omitempty"`
	Tags         []string `json:"tags,omitempty"`
	EnabledAPIs  []string `json:"enabled_apis,omitempty"`
	CreatedAt    int64    `json:"created_at,omitempty"`
}

// SiteConfig 站点配置
type SiteConfig struct {
	SiteName                string   `json:"site_name"`
	Announcement            string   `json:"announcement"`
	SearchDownstreamMaxPage int      `json:"search_downstream_max_page"`
	SearchSourceTimeoutMs   int      `json:"search_source_timeout_ms,omitempty"`
	SearchMaxConcurrent     int      `json:"search_max_concurrent,omitempty"`
	SearchDefaultSort       string   `json:"search_default_sort,omitempty"`
	SearchEnableStream      bool     `json:"search_enable_stream,omitempty"`
	SiteInterfaceCacheTime  int      `json:"site_interface_cache_time"`
	DoubanProxyType         string   `json:"douban_proxy_type"`
	DoubanProxy             string   `json:"douban_proxy"`
	DoubanImageProxyType    string   `json:"douban_image_proxy_type"`
	DoubanImageProxy        string   `json:"douban_image_proxy"`
	DisableYellowFilter     bool     `json:"disable_yellow_filter"`
	FluidSearch             bool     `json:"fluid_search"`
	ContentAccessMode       string   `json:"content_access_mode,omitempty"`
	BlockedContentTags      []string `json:"blocked_content_tags,omitempty"`
}

// ConfigSubscription 配置订阅
type ConfigSubscription struct {
	URL        string `json:"url,omitempty"`
	AutoUpdate bool   `json:"auto_update,omitempty"`
	LastCheck  string `json:"last_check,omitempty"`
}

// AdminConfig 管理员配置总结构
type AdminConfig struct {
	SiteConfig         SiteConfig         `json:"site_config"`
	UserConfig         []UserConfig       `json:"user_config,omitempty"`
	VideoSources       []VideoSource      `json:"video_sources,omitempty"`
	LiveSources        []LiveSource       `json:"live_sources,omitempty"`
	CustomCategories   []CustomCategory   `json:"custom_categories,omitempty"`
	UserGroups         []UserGroup        `json:"user_groups,omitempty"`
	ConfigSubscription ConfigSubscription `json:"config_subscription,omitempty"`
	ConfigFile         string             `json:"config_file,omitempty"`
}

// AdminStorageService Admin 存储服务接口
type AdminStorageService interface {
	// 配置管理
	GetAdminConfig(ctx context.Context) (*AdminConfig, error)
	SaveAdminConfig(ctx context.Context, config *AdminConfig) error

	// 用户管理
	GetUser(ctx context.Context, username string) (*UserConfig, error)
	GetAllUsers(ctx context.Context) ([]UserConfig, error)
	CreateUser(ctx context.Context, user *UserConfig) error
	UpdateUser(ctx context.Context, username string, user *UserConfig) error
	DeleteUser(ctx context.Context, username string) error
	ChangePassword(ctx context.Context, username string, passwordHash string) error

	// 视频源管理
	GetVideoSources(ctx context.Context) ([]VideoSource, error)
	SaveVideoSources(ctx context.Context, sources []VideoSource) error

	// 直播源管理
	GetLiveSources(ctx context.Context) ([]LiveSource, error)
	SaveLiveSources(ctx context.Context, sources []LiveSource) error

	// 分类管理
	GetCustomCategories(ctx context.Context) ([]CustomCategory, error)
	SaveCustomCategories(ctx context.Context, categories []CustomCategory) error
}

// AdminStats 管理员统计数据
type AdminStats struct {
	TotalUsers       int64 `json:"total_users"`
	TotalFavorites   int64 `json:"total_favorites"`
	TotalRecords     int64 `json:"total_records"`
	VideoSources     int   `json:"video_sources"`
	LiveSources      int   `json:"live_sources"`
	CustomCategories int   `json:"custom_categories"`
}

// ExportData 导出数据结构
type ExportData struct {
	ExportTime int64       `json:"export_time"`
	Version    string      `json:"version"`
	Data       AdminConfig `json:"data"`
}
