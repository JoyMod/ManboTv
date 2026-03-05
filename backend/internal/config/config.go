// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Load 加载配置
func Load(configPath string) (*AppConfig, error) {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 从环境变量读取
	v.SetEnvPrefix("MANBO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 如果指定了配置文件路径
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}

	// 从环境变量覆盖特定配置
	overrideFromEnv(v)

	var cfg AppConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &cfg, nil
}

// setDefaults 设置默认配置
func setDefaults(v *viper.Viper) {
	// 服务器默认值
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")
	v.SetDefault("server.read_timeout", "10s")
	v.SetDefault("server.write_timeout", "10s")

	// 日志默认值
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.output", "stdout")

	// Redis 默认值
	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)

	// 搜索默认值
	v.SetDefault("search.timeout", "10s")
	v.SetDefault("search.max_concurrent", 10)
	v.SetDefault("search.source_timeout", "2s")
	v.SetDefault("search.fast_return_after", "1200ms")
	v.SetDefault("search.cache_minutes", 15)
	v.SetDefault("search.max_results_per_source", 50)
	v.SetDefault("search.max_pages", 3)

	// 认证默认值
	v.SetDefault("auth.cookie_name", "auth")
	v.SetDefault("auth.token_expire_hours", 24*7) // 7天
}

// overrideFromEnv 从环境变量覆盖配置
func overrideFromEnv(v *viper.Viper) {
	// 服务器配置
	if val := os.Getenv("SERVER_HOST"); val != "" {
		v.Set("server.host", val)
	}
	if val := os.Getenv("SERVER_PORT"); val != "" {
		if port, err := parseInt(val); err == nil {
			v.Set("server.port", port)
		}
	}
	if val := os.Getenv("SERVER_MODE"); val != "" {
		v.Set("server.mode", val)
	}

	// Redis 配置
	if val := os.Getenv("REDIS_ADDR"); val != "" {
		v.Set("redis.addr", val)
	}
	if val := os.Getenv("REDIS_PASSWORD"); val != "" {
		v.Set("redis.password", val)
	}
	if val := os.Getenv("REDIS_DB"); val != "" {
		if db, err := parseInt(val); err == nil {
			v.Set("redis.db", db)
		}
	}

	// 日志配置
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		v.Set("log.level", val)
	}

	// 认证配置
	if val := os.Getenv("AUTH_COOKIE_NAME"); val != "" {
		v.Set("auth.cookie_name", val)
	}
	if val := os.Getenv("AUTH_TOKEN_EXPIRE_HOURS"); val != "" {
		if hours, err := parseInt(val); err == nil {
			v.Set("auth.token_expire_hours", time.Duration(hours))
		}
	}
	if val := os.Getenv("AUTH_JWT_SECRET"); val != "" {
		v.Set("auth.jwt_secret", val)
	}
}

// parseInt 解析整数
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
