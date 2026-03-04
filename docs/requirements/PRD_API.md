# API 接口规范

## 基础信息

- **Base URL**: `http://localhost:8080/api/v1`
- **协议**: HTTP/1.1 或 HTTP/2
- **编码**: UTF-8
- **Content-Type**: `application/json`

## 响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### 错误码

| Code | 含义 | HTTP Status |
|-----|------|------------|
| 0 | 成功 | 200 |
| 400 | 参数错误 | 400 |
| 401 | 未认证 | 401 |
| 403 | 无权限 | 403 |
| 404 | 资源不存在 | 404 |
| 429 | 请求过于频繁 | 429 |
| 500 | 服务器错误 | 500 |
| 503 | 服务不可用 | 503 |

---

## 接口列表

### 搜索

#### GET /search
多源聚合搜索

**Query Parameters:**
- `q` (string, required): 搜索关键词
- `page` (int, optional): 页码, 默认 1
- `page_size` (int, optional): 每页数量, 默认 20, 最大 50

**Response:**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "vod_id": "123",
        "vod_name": "影片名称",
        "vod_pic": "https://...",
        "vod_remarks": "更新至10集",
        "type_name": "国产剧",
        "vod_year": "2024",
        "source": "source_key",
        "site_name": "源名称"
      }
    ],
    "total": 100,
    "page": 1,
    "page_size": 20
  }
}
```

#### GET /search/suggestions
搜索建议

**Query Parameters:**
- `q` (string, required): 关键词

**Response:**
```json
{
  "code": 0,
  "data": ["建议1", "建议2", "建议3"]
}
```

### 详情

#### GET /detail
获取影片详情

**Query Parameters:**
- `id` (string, required): 影片ID
- `source` (string, required): 来源标识

**Response:**
```json
{
  "code": 0,
  "data": {
    "vod_id": "123",
    "vod_name": "影片名称",
    "vod_pic": "https://...",
    "vod_content": "剧情简介...",
    "vod_year": "2024",
    "vod_area": "中国大陆",
    "vod_director": "导演",
    "vod_actor": "演员1/演员2",
    "type_name": "国产剧",
    "episodes": [
      {"title": "第1集", "url": "https://..."}
    ]
  }
}
```

### 代理

#### GET /image
图片代理

**Query Parameters:**
- `url` (string, required): 原图URL

**Response:** 图片二进制

#### GET /proxy/m3u8
M3U8代理

**Query Parameters:**
- `url` (string, required): M3U8地址

### 收藏

#### GET /favorites
获取收藏列表

**Headers:**
- `Authorization`: Bearer {token}

**Response:**
```json
{
  "code": 0,
  "data": [
    {
      "id": "fav_123",
      "vod_id": "123",
      "vod_name": "影片名称",
      "vod_pic": "https://...",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

#### POST /favorites
添加收藏

**Request Body:**
```json
{
  "vod_id": "123",
  "vod_name": "影片名称",
  "vod_pic": "https://...",
  "source": "source_key"
}
```

#### DELETE /favorites/:id
取消收藏

### 播放记录

#### GET /playrecords
获取播放记录

#### POST /playrecords
保存播放记录

**Request Body:**
```json
{
  "vod_id": "123",
  "vod_name": "影片名称",
  "episode_index": 5,
  "progress": 120,
  "duration": 2400
}
```

### 认证

#### POST /auth/login
登录

**Request Body:**
```json
{
  "username": "admin",
  "password": "password"
}
```

**Response:**
```json
{
  "code": 0,
  "data": {
    "token": "jwt_token",
    "expires_in": 86400
  }
}
```

#### POST /auth/logout
登出
