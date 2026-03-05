// internal/model/storage.go
package model

import "context"

// Favorite 收藏
type Favorite struct {
	ID           string `json:"id"`
	VodID        string `json:"vod_id"`
	VodName      string `json:"vod_name"`
	VodPic       string `json:"vod_pic"`
	Year         string `json:"year,omitempty"`
	Source       string `json:"source"`
	SourceName   string `json:"source_name"`
	TotalEpisode int    `json:"total_episode"`
	SaveTime     int64  `json:"save_time"`
	SearchTitle  string `json:"search_title,omitempty"`
}

// PlayRecord 播放记录
type PlayRecord struct {
	ID            string `json:"id"`
	VodID         string `json:"vod_id"`
	VodName       string `json:"vod_name"`
	VodPic        string `json:"vod_pic"`
	Year          string `json:"year,omitempty"`
	Source        string `json:"source"`
	SourceName    string `json:"source_name,omitempty"`
	EpisodeIndex  int    `json:"episode_index"`
	EpisodeTitle  string `json:"episode_title,omitempty"`
	TotalEpisodes int    `json:"total_episodes,omitempty"`
	Progress      int    `json:"progress"` // 秒
	Duration      int    `json:"duration"` // 秒
	UpdatedAt     int64  `json:"updated_at"`
	SaveTime      int64  `json:"save_time,omitempty"`
	SearchTitle   string `json:"search_title,omitempty"`
}

// User 用户
type User struct {
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	Role         string `json:"role"`
	Banned       bool   `json:"banned"`
	CreatedAt    int64  `json:"created_at"`
}

// SkipConfig 跳过片头片尾配置
type SkipConfig struct {
	Enable    bool `json:"enable"`
	IntroTime int  `json:"intro_time"`
	OutroTime int  `json:"outro_time"`
}

// StorageService 存储服务接口
type StorageService interface {
	// 收藏相关
	GetFavorite(ctx context.Context, username string, source string, vodID string) (*Favorite, error)
	SaveFavorite(ctx context.Context, username string, favorite *Favorite) error
	DeleteFavorite(ctx context.Context, username string, source string, vodID string) error
	GetFavorites(ctx context.Context, username string, page, pageSize int) ([]*Favorite, int64, error)

	// 播放记录相关
	GetPlayRecord(ctx context.Context, username string, source string, vodID string) (*PlayRecord, error)
	SavePlayRecord(ctx context.Context, username string, record *PlayRecord) error
	DeletePlayRecord(ctx context.Context, username string, source string, vodID string) error
	GetPlayRecords(ctx context.Context, username string, page, pageSize int) ([]*PlayRecord, int64, error)

	// 搜索历史相关
	GetSearchHistory(ctx context.Context, username string, limit int) ([]string, error)
	AddSearchHistory(ctx context.Context, username string, keyword string) error
	RemoveSearchHistory(ctx context.Context, username string, keyword string) error
	ClearSearchHistory(ctx context.Context, username string) error

	// 跳过配置相关
	GetSkipConfig(ctx context.Context, username string, source string, vodID string) (*SkipConfig, error)
	SetSkipConfig(ctx context.Context, username string, source string, vodID string, config *SkipConfig) error
	DeleteSkipConfig(ctx context.Context, username string, source string, vodID string) error
	GetAllSkipConfigs(ctx context.Context, username string) (map[string]*SkipConfig, error)

	// 用户相关
	GetUser(ctx context.Context, username string) (*User, error)
	SaveUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, username string) error
}
