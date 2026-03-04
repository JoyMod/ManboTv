// internal/config/config.go
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 全局配置实例
var Config *AppConfig

// Load 加载配置
func Load(configPath string) (*AppConfig, error) {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 读取配置文件
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	// 读取环境变量
	v.SetEnvPrefix("MANBO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
		// 配置文件不存在时使用默认值
	}

	// 解析配置
	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	Config = &cfg
	return &cfg, nil
}

// setDefaults 设置默认值
func setDefaults(v *viper.Viper) {
	// 服务器默认值
	v.SetDefault("server.host", DefaultServerHost)
	v.SetDefault("server.port", DefaultServerPort)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.read_timeout", "10s")
	v.SetDefault("server.write_timeout", "10s")

	// 日志默认值
	v.SetDefault("log.level", DefaultLogLevel)
	v.SetDefault("log.format", "json")
	v.SetDefault("log.output", "stdout")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_backups", 30)
	v.SetDefault("log.max_age", 7)
	v.SetDefault("log.compress", true)

	// Redis默认值
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", DefaultRedisDB)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 5)
	v.SetDefault("redis.max_retries", 3)

	// 搜索默认值
	v.SetDefault("search.timeout", "10s")
	v.SetDefault("search.max_concurrent", 10)
	v.SetDefault("search.cache_minutes", 15)
	v.SetDefault("search.max_results_per_source", 50)
	v.SetDefault("search.max_pages", 3)

	// 图片代理默认值
	v.SetDefault("image_proxy.timeout", "30s")
	v.SetDefault("image_proxy.cache_size", 1000)
	v.SetDefault("image_proxy.cache_max_item_size", 1048576)
	v.SetDefault("image_proxy.user_agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	// HTTP客户端默认值
	v.SetDefault("http_client.max_idle_conns", 100)
	v.SetDefault("http_client.max_idle_conns_per_host", 10)
	v.SetDefault("http_client.idle_conn_timeout", "90s")
	v.SetDefault("http_client.timeout", "10s")

	// 认证默认值
	v.SetDefault("auth.cookie_name", "auth_token")
	v.SetDefault("auth.token_expire_hours", 24)

	// CORS默认值
	v.SetDefault("cors.allow_origins", []string{"http://localhost:3000"})
	v.SetDefault("cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization"})
	v.SetDefault("cors.allow_credentials", true)
}

// GetConfig 获取全局配置
func GetConfig() *AppConfig {
	if Config == nil {
		panic("配置未初始化")
	}
	return Config
}
