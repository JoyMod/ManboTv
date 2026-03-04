# API 路由映射: Next.js → Go

## 路由对照表

| 原路由 (Next.js) | 新路由 (Go) | 方法 | 说明 |
|-----------------|------------|------|------|
| `/api/search` | `/api/v1/search` | GET | 多源聚合搜索 |
| `/api/search/one` | `/api/v1/search/one` | GET | 单源搜索 |
| `/api/search/suggestions` | `/api/v1/search/suggestions` | GET | 搜索建议 |
| `/api/search/resources` | `/api/v1/search/resources` | GET | 资源列表 |
| `/api/search/ws` | `/api/v1/search/ws` | WS | WebSocket搜索 |
| `/api/detail` | `/api/v1/detail` | GET | 详情页 |
| `/api/image-proxy` | `/api/v1/image` | GET | 图片代理 |
| `/api/proxy/m3u8` | `/api/v1/proxy/m3u8` | GET | M3U8代理 |
| `/api/proxy/segment` | `/api/v1/proxy/segment` | GET | 视频分片代理 |
| `/api/proxy/key` | `/api/v1/proxy/key` | GET | 密钥代理 |
| `/api/proxy/logo` | `/api/v1/proxy/logo` | GET | Logo代理 |
| `/api/favorites` | `/api/v1/favorites` | GET/POST/DELETE | 收藏管理 |
| `/api/playrecords` | `/api/v1/playrecords` | GET/POST/DELETE | 播放记录 |
| `/api/searchhistory` | `/api/v1/searchhistory` | GET/POST | 搜索历史 |
| `/api/skipconfigs` | `/api/v1/skipconfigs` | GET/POST | 跳过配置 |
| `/api/live/sources` | `/api/v1/live/sources` | GET | 直播源 |
| `/api/live/channels` | `/api/v1/live/channels` | GET | 直播频道 |
| `/api/live/epg` | `/api/v1/live/epg` | GET | 节目单 |
| `/api/live/precheck` | `/api/v1/live/precheck` | POST | 直播预检 |
| `/api/douban` | `/api/v1/douban` | GET | 豆瓣搜索 |
| `/api/douban/recommends` | `/api/v1/douban/recommends` | GET | 豆瓣推荐 |
| `/api/douban/categories` | `/api/v1/douban/categories` | GET | 豆瓣分类 |
| `/api/admin/*` | `/api/v1/admin/*` | * | 管理后台 |
| `/api/login` | `/api/v1/auth/login` | POST | 登录 |
| `/api/logout` | `/api/v1/auth/logout` | POST | 登出 |
| `/api/change-password` | `/api/v1/auth/password` | PUT | 修改密码 |
| `/api/cron` | `/api/v1/cron` | POST | 定时任务 |
| `/api/server-config` | `/api/v1/config` | GET | 服务配置 |

## 核心数据结构

### SearchResult (搜索结果)
```go
type SearchResult struct {
    VodID       string   `json:"vod_id"`
    VodName     string   `json:"vod_name"`
    VodPic      string   `json:"vod_pic"`
    VodRemarks  string   `json:"vod_remarks"`
    VodClass    string   `json:"vod_class"`
    VodYear     string   `json:"vod_year"`
    VodContent  string   `json:"vod_content"`
    VodDoubanID int      `json:"vod_douban_id"`
    TypeName    string   `json:"type_name"`
    Episodes    []string `json:"episodes"`
    Titles      []string `json:"titles"`
    Source      string   `json:"source"`
    SiteName    string   `json:"site_name"`
}
```

### ApiSite (API站点配置)
```go
type ApiSite struct {
    Key    string `json:"key"`
    API    string `json:"api"`
    Name   string `json:"name"`
    Detail string `json:"detail,omitempty"`
}
```

### Favorite (收藏)
```go
type Favorite struct {
    ID        string    `json:"id"`
    Source    string    `json:"source"`
    VodID     string    `json:"vod_id"`
    VodName   string    `json:"vod_name"`
    VodPic    string    `json:"vod_pic"`
    CreatedAt time.Time `json:"created_at"`
}
```

### PlayRecord (播放记录)
```go
type PlayRecord struct {
    ID           string    `json:"id"`
    Source       string    `json:"source"`
    VodID        string    `json:"vod_id"`
    VodName      string    `json:"vod_name"`
    EpisodeIndex int       `json:"episode_index"`
    Progress     int       `json:"progress"` // 秒
    Duration     int       `json:"duration"` // 秒
    UpdatedAt    time.Time `json:"updated_at"`
}
```
