// internal/config/types.go
package config

import (
	"time"
)

// AppConfig 应用总配置
type AppConfig struct {
	Server     ServerConfig     `mapstructure:"server"`
	Log        LogConfig        `mapstructure:"log"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Search     SearchConfig     `mapstructure:"search"`
	ImageProxy ImageProxyConfig `mapstructure:"image_proxy"`
	HTTPClient HTTPClientConfig `mapstructure:"http_client"`
	Auth       AuthConfig       `mapstructure:"auth"`
	CORS       CORSConfig       `mapstructure:"cors"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr         string `mapstructure:"addr"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	MaxRetries   int    `mapstructure:"max_retries"`
}

// SearchConfig 搜索配置
type SearchConfig struct {
	Timeout             time.Duration `mapstructure:"timeout"`
	MaxConcurrent       int           `mapstructure:"max_concurrent"`
	SourceTimeout       time.Duration `mapstructure:"source_timeout"`
	FastReturnAfter     time.Duration `mapstructure:"fast_return_after"`
	CacheMinutes        int           `mapstructure:"cache_minutes"`
	MaxResultsPerSource int           `mapstructure:"max_results_per_source"`
	MaxPages            int           `mapstructure:"max_pages"`
}

// ImageProxyConfig 图片代理配置
type ImageProxyConfig struct {
	Timeout          time.Duration `mapstructure:"timeout"`
	CacheSize        int           `mapstructure:"cache_size"`
	CacheMaxItemSize int64         `mapstructure:"cache_max_item_size"`
	UserAgent        string        `mapstructure:"user_agent"`
}

// HTTPClientConfig HTTP客户端配置
type HTTPClientConfig struct {
	MaxIdleConns        int           `mapstructure:"max_idle_conns"`
	MaxIdleConnsPerHost int           `mapstructure:"max_idle_conns_per_host"`
	IdleConnTimeout     time.Duration `mapstructure:"idle_conn_timeout"`
	Timeout             time.Duration `mapstructure:"timeout"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	CookieName       string        `mapstructure:"cookie_name"`
	TokenExpireHours time.Duration `mapstructure:"token_expire_hours"`
	JWTSecret        string        `mapstructure:"jwt_secret"`
}

// CORSConfig 跨域配置
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

// 默认配置常量
const (
	DefaultServerHost = "0.0.0.0"
	DefaultServerPort = 8080
	DefaultLogLevel   = "info"
	DefaultRedisDB    = 0
)
