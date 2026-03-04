# 数据库设计

## 概述

- **主数据库**: Redis (缓存 + 持久化)
- **配置存储**: SQLite / YAML 文件
- **数据格式**: JSON

## Redis Key 设计

### 用户数据

| Key Pattern | 类型 | 说明 |
|------------|------|------|
| `user:{username}` | Hash | 用户信息 |
| `user:{username}:favorites` | ZSet | 收藏列表 (score: timestamp) |
| `user:{username}:history` | ZSet | 播放历史 (score: timestamp) |
| `user:{username}:search_history` | List | 搜索历史 |

### 缓存数据

| Key Pattern | 类型 | TTL | 说明 |
|------------|------|-----|------|
| `search:{source}:{query}:{page}` | String | 15min | 搜索结果缓存 |
| `detail:{source}:{id}` | String | 30min | 详情缓存 |
| `suggestions:{query}` | String | 5min | 搜索建议缓存 |
| `image:{hash}` | String | 24h | 图片缓存 |

### 全局配置

| Key Pattern | 类型 | 说明 |
|------------|------|------|
| `config:site` | Hash | 站点配置 |
| `config:api_sites` | Hash | API源配置 |
| `config:live` | Hash | 直播源配置 |

## 数据结构

### User
```json
{
  "username": "admin",
  "password_hash": "...",
  "role": "admin",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Favorite (ZSet Member)
```json
{
  "id": "fav_123",
  "vod_id": "123",
  "vod_name": "影片名称",
  "vod_pic": "https://...",
  "source": "source_key",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### PlayRecord (ZSet Member)
```json
{
  "id": "rec_123",
  "vod_id": "123",
  "vod_name": "影片名称",
  "vod_pic": "https://...",
  "episode_index": 5,
  "episode_title": "第5集",
  "progress": 120,
  "duration": 2400,
  "updated_at": "2024-01-01T00:00:00Z"
}
```

### ApiSite (配置)
```json
{
  "key": "site1",
  "name": "源名称",
  "api": "https://api.example.com",
  "detail": "https://detail.example.com",
  "enabled": true
}
```

## 索引设计

### 搜索缓存
- Key: `search:{source}:{query}:{page}`
- TTL: 15分钟
- 清理策略: 自动过期

### 用户收藏
- ZSet: `user:{username}:favorites`
- Score: 创建时间戳
- 分页: ZREVRANGE + LIMIT

### 播放历史
- ZSet: `user:{username}:history`
- Score: 更新时间戳
- 去重: 先 ZREM 旧记录，再 ZADD
