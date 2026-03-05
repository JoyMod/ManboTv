// internal/service/admin_storage_service.go
// Admin 配置存储服务实现

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
	"github.com/JoyMod/ManboTV/backend/internal/repository/redis"
)

// adminStorageService Admin 存储服务实现
type adminStorageService struct {
	redis  *redis.Client
	logger *zap.Logger
}

// NewAdminStorageService 创建 Admin 存储服务
func NewAdminStorageService(redisClient *redis.Client, logger *zap.Logger) model.AdminStorageService {
	return &adminStorageService{
		redis:  redisClient,
		logger: logger,
	}
}

// adminConfigKey Admin 配置存储 key
const adminConfigKey = "manbotv:admin:config"

// usersKey 用户列表 key
const usersKey = "manbotv:admin:users"

// ---------- 配置管理 ----------

// GetAdminConfig 获取 Admin 配置
func (s *adminStorageService) GetAdminConfig(ctx context.Context) (*model.AdminConfig, error) {
	data, err := s.redis.Get(ctx, adminConfigKey)
	if err != nil {
		return nil, fmt.Errorf("获取配置失败: %w", err)
	}

	if data == "" {
		// 返回默认配置
		return &model.AdminConfig{
			SiteConfig: model.SiteConfig{
				SiteName:                "ManboTV",
				SearchDownstreamMaxPage: 5,
				SiteInterfaceCacheTime:  3600,
				FluidSearch:             true,
			},
			VideoSources:     make([]model.VideoSource, 0),
			LiveSources:      make([]model.LiveSource, 0),
			CustomCategories: make([]model.CustomCategory, 0),
			UserConfig:       make([]model.UserConfig, 0),
			UserGroups:       make([]model.UserGroup, 0),
		}, nil
	}

	var config model.AdminConfig
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &config, nil
}

// SaveAdminConfig 保存 Admin 配置
func (s *adminStorageService) SaveAdminConfig(ctx context.Context, config *model.AdminConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := s.redis.Set(ctx, adminConfigKey, string(data), 0); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	s.logger.Debug("Admin 配置已保存")
	return nil
}

// ---------- 用户管理 ----------

// GetUser 获取单个用户
func (s *adminStorageService) GetUser(ctx context.Context, username string) (*model.UserConfig, error) {
	key := generateKey("admin", "user", username)

	data, err := s.redis.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}

	if data == "" {
		return nil, nil
	}

	var user model.UserConfig
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, fmt.Errorf("解析用户数据失败: %w", err)
	}

	return &user, nil
}

// GetAllUsers 获取所有用户
func (s *adminStorageService) GetAllUsers(ctx context.Context) ([]model.UserConfig, error) {
	// 获取用户列表（从 Set 中）
	usernames, err := s.redis.SMembers(ctx, usersKey)
	if err != nil {
		return nil, fmt.Errorf("获取用户列表失败: %w", err)
	}

	users := make([]model.UserConfig, 0, len(usernames))
	for _, username := range usernames {
		user, err := s.GetUser(ctx, username)
		if err != nil {
			s.logger.Warn("获取用户失败", zap.String("username", username), zap.Error(err))
			continue
		}
		if user != nil {
			// 不返回密码哈希
			user.PasswordHash = ""
			users = append(users, *user)
		}
	}

	return users, nil
}

// CreateUser 创建用户
func (s *adminStorageService) CreateUser(ctx context.Context, user *model.UserConfig) error {
	// 检查用户是否已存在
	existing, err := s.GetUser(ctx, user.Username)
	if err != nil {
		return err
	}
	if existing != nil {
		return fmt.Errorf("用户已存在")
	}

	// 保存用户
	key := generateKey("admin", "user", user.Username)
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("序列化用户失败: %w", err)
	}

	if err := s.redis.Set(ctx, key, string(data), 0); err != nil {
		return fmt.Errorf("保存用户失败: %w", err)
	}

	// 添加到用户列表
	if err := s.redis.SAdd(ctx, usersKey, user.Username); err != nil {
		s.logger.Warn("添加用户到列表失败", zap.Error(err))
	}

	s.logger.Info("用户已创建", zap.String("username", user.Username))
	return nil
}

// UpdateUser 更新用户
func (s *adminStorageService) UpdateUser(ctx context.Context, username string, user *model.UserConfig) error {
	// 检查用户是否存在
	existing, err := s.GetUser(ctx, username)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("用户不存在")
	}

	// 保留密码哈希（如果不提供新密码）
	if user.PasswordHash == "" {
		user.PasswordHash = existing.PasswordHash
	}

	// 如果用户名改变，需要删除旧用户并更新列表
	if user.Username != username {
		// 检查新用户名是否已存在
		newUser, err := s.GetUser(ctx, user.Username)
		if err != nil {
			return err
		}
		if newUser != nil {
			return fmt.Errorf("新用户名已存在")
		}

		// 删除旧用户
		oldKey := generateKey("admin", "user", username)
		if err := s.redis.Del(ctx, oldKey); err != nil {
			return fmt.Errorf("删除旧用户失败: %w", err)
		}

		// 从列表中移除旧用户名
		if err := s.redis.SRem(ctx, usersKey, username); err != nil {
			s.logger.Warn("从列表移除旧用户名失败", zap.Error(err))
		}

		// 添加新用户名到列表
		if err := s.redis.SAdd(ctx, usersKey, user.Username); err != nil {
			s.logger.Warn("添加新用户名到列表失败", zap.Error(err))
		}
	}

	// 保存用户
	key := generateKey("admin", "user", user.Username)
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("序列化用户失败: %w", err)
	}

	if err := s.redis.Set(ctx, key, string(data), 0); err != nil {
		return fmt.Errorf("保存用户失败: %w", err)
	}

	s.logger.Info("用户已更新", zap.String("username", user.Username))
	return nil
}

// DeleteUser 删除用户
func (s *adminStorageService) DeleteUser(ctx context.Context, username string) error {
	key := generateKey("admin", "user", username)

	if err := s.redis.Del(ctx, key); err != nil {
		return fmt.Errorf("删除用户失败: %w", err)
	}

	// 从列表中移除
	if err := s.redis.SRem(ctx, usersKey, username); err != nil {
		s.logger.Warn("从列表移除用户名失败", zap.Error(err))
	}

	s.logger.Info("用户已删除", zap.String("username", username))
	return nil
}

// ChangePassword 修改密码
func (s *adminStorageService) ChangePassword(ctx context.Context, username string, passwordHash string) error {
	user, err := s.GetUser(ctx, username)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("用户不存在")
	}

	user.PasswordHash = passwordHash
	return s.UpdateUser(ctx, username, user)
}

// ---------- 视频源管理 ----------

// GetVideoSources 获取视频源列表
func (s *adminStorageService) GetVideoSources(ctx context.Context) ([]model.VideoSource, error) {
	config, err := s.GetAdminConfig(ctx)
	if err != nil {
		return nil, err
	}

	cfgFile := parseAdminConfigFile(config.ConfigFile)
	return mergeVideoSources(config.VideoSources, cfgFile.APISite), nil
}

// SaveVideoSources 保存视频源列表
func (s *adminStorageService) SaveVideoSources(ctx context.Context, sources []model.VideoSource) error {
	config, err := s.GetAdminConfig(ctx)
	if err != nil {
		return err
	}
	config.VideoSources = sources
	return s.SaveAdminConfig(ctx, config)
}

// ---------- 直播源管理 ----------

// GetLiveSources 获取直播源列表
func (s *adminStorageService) GetLiveSources(ctx context.Context) ([]model.LiveSource, error) {
	config, err := s.GetAdminConfig(ctx)
	if err != nil {
		return nil, err
	}

	cfgFile := parseAdminConfigFile(config.ConfigFile)
	return mergeLiveSources(config.LiveSources, cfgFile.Lives), nil
}

// SaveLiveSources 保存直播源列表
func (s *adminStorageService) SaveLiveSources(ctx context.Context, sources []model.LiveSource) error {
	config, err := s.GetAdminConfig(ctx)
	if err != nil {
		return err
	}
	config.LiveSources = sources
	return s.SaveAdminConfig(ctx, config)
}

// ---------- 分类管理 ----------

// GetCustomCategories 获取自定义分类
func (s *adminStorageService) GetCustomCategories(ctx context.Context) ([]model.CustomCategory, error) {
	config, err := s.GetAdminConfig(ctx)
	if err != nil {
		return nil, err
	}
	return config.CustomCategories, nil
}

// SaveCustomCategories 保存自定义分类
func (s *adminStorageService) SaveCustomCategories(ctx context.Context, categories []model.CustomCategory) error {
	config, err := s.GetAdminConfig(ctx)
	if err != nil {
		return err
	}
	config.CustomCategories = categories
	return s.SaveAdminConfig(ctx, config)
}

type adminConfigFile struct {
	APISite map[string]apiSiteConfig  `json:"api_site"`
	Lives   map[string]liveSiteConfig `json:"lives"`
}

type apiSiteConfig struct {
	Name   string `json:"name"`
	API    string `json:"api"`
	Detail string `json:"detail"`
}

type liveSiteConfig struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	UA   string `json:"ua"`
	EPG  string `json:"epg"`
}

func parseAdminConfigFile(raw string) adminConfigFile {
	if strings.TrimSpace(raw) == "" {
		return adminConfigFile{}
	}

	var cfg adminConfigFile
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return adminConfigFile{}
	}

	return cfg
}

func mergeVideoSources(base []model.VideoSource, fileSources map[string]apiSiteConfig) []model.VideoSource {
	merged := make([]model.VideoSource, 0, len(base)+len(fileSources))
	index := make(map[string]int, len(base)+len(fileSources))

	for _, source := range base {
		source.Key = strings.TrimSpace(source.Key)
		source.Name = strings.TrimSpace(source.Name)
		source.API = strings.TrimSpace(source.API)
		source.Detail = strings.TrimSpace(source.Detail)
		if source.Key == "" || source.API == "" {
			continue
		}
		if _, ok := fileSources[source.Key]; ok {
			source.From = "config"
		} else if strings.TrimSpace(source.From) == "" {
			source.From = "custom"
		}
		index[source.Key] = len(merged)
		merged = append(merged, source)
	}

	for key, source := range fileSources {
		trimmedKey := strings.TrimSpace(key)
		api := strings.TrimSpace(source.API)
		if trimmedKey == "" || api == "" {
			continue
		}

		if i, ok := index[trimmedKey]; ok {
			merged[i].Name = strings.TrimSpace(source.Name)
			merged[i].API = api
			merged[i].Detail = strings.TrimSpace(source.Detail)
			merged[i].From = "config"
			continue
		}

		index[trimmedKey] = len(merged)
		merged = append(merged, model.VideoSource{
			Key:      trimmedKey,
			Name:     strings.TrimSpace(source.Name),
			API:      api,
			Detail:   strings.TrimSpace(source.Detail),
			From:     "config",
			Disabled: false,
		})
	}

	return merged
}

func mergeLiveSources(base []model.LiveSource, fileSources map[string]liveSiteConfig) []model.LiveSource {
	merged := make([]model.LiveSource, 0, len(base)+len(fileSources))
	index := make(map[string]int, len(base)+len(fileSources))

	for _, source := range base {
		source.Key = strings.TrimSpace(source.Key)
		source.Name = strings.TrimSpace(source.Name)
		source.URL = strings.TrimSpace(source.URL)
		source.UA = strings.TrimSpace(source.UA)
		source.EPG = strings.TrimSpace(source.EPG)
		if source.Key == "" || source.URL == "" {
			continue
		}
		if _, ok := fileSources[source.Key]; ok {
			source.From = "config"
		} else if strings.TrimSpace(source.From) == "" {
			source.From = "custom"
		}
		index[source.Key] = len(merged)
		merged = append(merged, source)
	}

	for key, source := range fileSources {
		trimmedKey := strings.TrimSpace(key)
		url := strings.TrimSpace(source.URL)
		if trimmedKey == "" || url == "" {
			continue
		}

		if i, ok := index[trimmedKey]; ok {
			merged[i].Name = strings.TrimSpace(source.Name)
			merged[i].URL = url
			merged[i].UA = strings.TrimSpace(source.UA)
			merged[i].EPG = strings.TrimSpace(source.EPG)
			merged[i].From = "config"
			continue
		}

		index[trimmedKey] = len(merged)
		merged = append(merged, model.LiveSource{
			Key:           trimmedKey,
			Name:          strings.TrimSpace(source.Name),
			URL:           url,
			UA:            strings.TrimSpace(source.UA),
			EPG:           strings.TrimSpace(source.EPG),
			From:          "config",
			Disabled:      false,
			ChannelNumber: 0,
		})
	}

	return merged
}
