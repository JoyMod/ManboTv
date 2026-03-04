// internal/service/storage_service.go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/JoyMod/ManboTV/backend/internal/model"
	"github.com/JoyMod/ManboTV/backend/internal/repository/redis"
)

// storageService Redis存储服务实现
type storageService struct {
	redis  *redis.Client
	logger *zap.Logger
}

// NewStorageService 创建存储服务
func NewStorageService(redisClient *redis.Client, logger *zap.Logger) model.StorageService {
	return &storageService{
		redis:  redisClient,
		logger: logger,
	}
}

// 生成存储key
func generateKey(parts ...string) string {
	key := "manbotv"
	for _, part := range parts {
		key += ":" + part
	}
	return key
}

// ---------- 收藏相关 ----------

// GetFavorite 获取单条收藏
func (s *storageService) GetFavorite(ctx context.Context, username string, source string, vodID string) (*model.Favorite, error) {
	key := generateKey("fav", username)
	field := source + "+" + vodID

	data, err := s.redis.HGet(ctx, key, field)
	if err != nil {
		return nil, fmt.Errorf("获取收藏失败: %w", err)
	}
	if data == "" {
		return nil, nil
	}

	var fav model.Favorite
	if err := json.Unmarshal([]byte(data), &fav); err != nil {
		return nil, fmt.Errorf("解析收藏数据失败: %w", err)
	}

	return &fav, nil
}

// SaveFavorite 保存收藏
func (s *storageService) SaveFavorite(ctx context.Context, username string, favorite *model.Favorite) error {
	key := generateKey("fav", username)
	field := favorite.Source + "+" + favorite.VodID

	if favorite.SaveTime == 0 {
		favorite.SaveTime = time.Now().Unix()
	}

	data, err := json.Marshal(favorite)
	if err != nil {
		return fmt.Errorf("序列化收藏失败: %w", err)
	}

	if err := s.redis.HSet(ctx, key, field, string(data)); err != nil {
		return fmt.Errorf("保存收藏失败: %w", err)
	}

	s.logger.Debug("收藏已保存",
		zap.String("user", username),
		zap.String("vod", favorite.VodName),
	)

	return nil
}

// DeleteFavorite 删除收藏
func (s *storageService) DeleteFavorite(ctx context.Context, username string, source string, vodID string) error {
	key := generateKey("fav", username)
	field := source + "+" + vodID

	if err := s.redis.HDel(ctx, key, field); err != nil {
		return fmt.Errorf("删除收藏失败: %w", err)
	}

	s.logger.Debug("收藏已删除",
		zap.String("user", username),
		zap.String("source", source),
		zap.String("vod_id", vodID),
	)

	return nil
}

// GetFavorites 获取收藏列表 (分页)
func (s *storageService) GetFavorites(ctx context.Context, username string, page, pageSize int) ([]*model.Favorite, int64, error) {
	key := generateKey("fav", username)

	// 获取所有收藏
	data, err := s.redis.HGetAll(ctx, key)
	if err != nil {
		return nil, 0, fmt.Errorf("获取收藏列表失败: %w", err)
	}

	total := int64(len(data))

	// 转换为列表
	var favorites []*model.Favorite
	for _, v := range data {
		var fav model.Favorite
		if err := json.Unmarshal([]byte(v), &fav); err != nil {
			s.logger.Warn("解析收藏数据失败", zap.Error(err))
			continue
		}
		favorites = append(favorites, &fav)
	}

	// 按时间倒序排序
	// 这里简化处理，实际应该使用 ZSet

	// 分页
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	if start > len(favorites) {
		start = len(favorites)
	}

	end := start + pageSize
	if end > len(favorites) {
		end = len(favorites)
	}

	return favorites[start:end], total, nil
}

// ---------- 播放记录相关 ----------

// GetPlayRecord 获取播放记录
func (s *storageService) GetPlayRecord(ctx context.Context, username string, source string, vodID string) (*model.PlayRecord, error) {
	key := generateKey("rec", username)
	field := source + "+" + vodID

	data, err := s.redis.HGet(ctx, key, field)
	if err != nil {
		return nil, fmt.Errorf("获取播放记录失败: %w", err)
	}
	if data == "" {
		return nil, nil
	}

	var record model.PlayRecord
	if err := json.Unmarshal([]byte(data), &record); err != nil {
		return nil, fmt.Errorf("解析播放记录失败: %w", err)
	}

	return &record, nil
}

// SavePlayRecord 保存播放记录
func (s *storageService) SavePlayRecord(ctx context.Context, username string, record *model.PlayRecord) error {
	key := generateKey("rec", username)
	field := record.Source + "+" + record.VodID

	record.UpdatedAt = time.Now().Unix()

	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("序列化播放记录失败: %w", err)
	}

	if err := s.redis.HSet(ctx, key, field, string(data)); err != nil {
		return fmt.Errorf("保存播放记录失败: %w", err)
	}

	return nil
}

// DeletePlayRecord 删除播放记录
func (s *storageService) DeletePlayRecord(ctx context.Context, username string, source string, vodID string) error {
	key := generateKey("rec", username)
	field := source + "+" + vodID

	return s.redis.HDel(ctx, key, field)
}

// GetPlayRecords 获取播放记录列表
func (s *storageService) GetPlayRecords(ctx context.Context, username string, page, pageSize int) ([]*model.PlayRecord, int64, error) {
	key := generateKey("rec", username)

	data, err := s.redis.HGetAll(ctx, key)
	if err != nil {
		return nil, 0, fmt.Errorf("获取播放记录失败: %w", err)
	}

	total := int64(len(data))

	var records []*model.PlayRecord
	for _, v := range data {
		var rec model.PlayRecord
		if err := json.Unmarshal([]byte(v), &rec); err != nil {
			continue
		}
		records = append(records, &rec)
	}

	// 分页
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	if start > len(records) {
		start = len(records)
	}

	end := start + pageSize
	if end > len(records) {
		end = len(records)
	}

	return records[start:end], total, nil
}

// ---------- 搜索历史相关 ----------

// GetSearchHistory 获取搜索历史
func (s *storageService) GetSearchHistory(ctx context.Context, username string, limit int) ([]string, error) {
	key := generateKey("hist", username)

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// 从列表右侧获取最近搜索 (LPush 是从左侧插入，所以最新的在左侧)
	// 这里我们从右侧取是错误的，应该从左侧取
	// 修正：LPush 后，最新的在索引 0，所以从 0 开始取
	return s.redis.LRange(ctx, key, 0, int64(limit-1))
}

// AddSearchHistory 添加搜索历史
func (s *storageService) AddSearchHistory(ctx context.Context, username string, keyword string) error {
	if keyword == "" {
		return nil
	}

	key := generateKey("hist", username)

	// 从左侧推入
	if err := s.redis.LPush(ctx, key, keyword); err != nil {
		return fmt.Errorf("添加搜索历史失败: %w", err)
	}

	// 修剪列表，只保留最近100条
	return s.redis.LTrim(ctx, key, 0, 99)
}

// ClearSearchHistory 清空搜索历史
func (s *storageService) ClearSearchHistory(ctx context.Context, username string) error {
	key := generateKey("hist", username)
	return s.redis.Del(ctx, key)
}

// ---------- 用户相关 ----------

// GetUser 获取用户
func (s *storageService) GetUser(ctx context.Context, username string) (*model.User, error) {
	key := generateKey("user", username)

	found, err := s.redis.GetJSON(ctx, key, &model.User{})
	if err != nil {
		return nil, fmt.Errorf("获取用户失败: %w", err)
	}
	if !found {
		return nil, nil
	}

	return &model.User{}, nil
}

// SaveUser 保存用户
func (s *storageService) SaveUser(ctx context.Context, user *model.User) error {
	key := generateKey("user", user.Username)
	return s.redis.SetJSON(ctx, key, user, 0) // 不过期
}

// DeleteUser 删除用户
func (s *storageService) DeleteUser(ctx context.Context, username string) error {
	key := generateKey("user", username)
	return s.redis.Del(ctx, key)
}


