# API 与 Service 代码分析

## 1. 核心 API 路由分析

### 1.1 /api/search (聚合搜索)

**当前实现分析:**
```typescript
// 核心逻辑
const searchPromises = apiSites.map((site) =>
  Promise.race([
    searchFromApi(site, query),
    new Promise((_, reject) =>
      setTimeout(() => reject(new Error(`${site.name} timeout`)), 20000)
    ),
  ]).catch((err) => {
    console.warn(`搜索失败 ${site.name}:`, err.message);
    return [];
  })
);

const results = await Promise.allSettled(searchPromises);
```

**痛点:**
1. 使用 `Promise.race` 实现超时，但 20s 超时时间过长
2. 无并发控制，同时请求所有源
3. 无连接复用，每次 fetch 新建连接
4. 错误处理简单，只是返回空数组
5. 使用 `console.warn` 而非结构化日志

**Go 重构方案:**
```go
// 使用 errgroup + context 控制并发和超时
func (s *SearchService) Search(ctx context.Context, query string, sites []ApiSite) ([]SearchResult, error) {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(10) // 最多10个并发
    
    var mu sync.Mutex
    var results []SearchResult
    
    for _, site := range sites {
        site := site // 捕获
        g.Go(func() error {
            res, err := s.searchSingle(ctx, site, query)
            if err != nil {
                s.logger.Warn("搜索源失败", zap.String("site", site.Name), zap.Error(err))
                return nil // 不中断其他请求
            }
            mu.Lock()
            results = append(results, res...)
            mu.Unlock()
            return nil
        })
    }
    
    if err := g.Wait(); err != nil {
        return nil, err
    }
    return results, nil
}
```

---

### 1.2 /api/image-proxy (图片代理)

**当前实现分析:**
```typescript
const imageResponse = await fetch(imageUrl, {
  headers: { Referer: '...', 'User-Agent': '...' },
});

return new Response(imageResponse.body, { status: 200, headers });
```

**痛点:**
1. 无连接复用，每次请求新建 TCP 连接
2. 无本地缓存，重复请求相同图片
3. 流式转发但无缓冲优化
4. 错误处理简单

**Go 重构方案:**
```go
type ImageService struct {
    client  *http.Client  // 复用连接池
    cache   *lru.Cache     // 本地内存缓存
    logger  *zap.Logger
}

func (s *ImageService) Proxy(ctx context.Context, imageUrl string) (io.Reader, error) {
    // 1. 查本地缓存
    if cached, ok := s.cache.Get(imageUrl); ok {
        s.logger.Debug("图片缓存命中", zap.String("url", imageUrl))
        return bytes.NewReader(cached.([]byte)), nil
    }
    
    // 2. 流式请求
    req, _ := http.NewRequestWithContext(ctx, "GET", imageUrl, nil)
    req.Header.Set("Referer", "https://movie.douban.com/")
    req.Header.Set("User-Agent", "Mozilla/5.0...")
    
    resp, err := s.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("fetch image failed: %w", err)
    }
    defer resp.Body.Close()
    
    // 3. 读取并缓存 (小图片)
    if resp.ContentLength < 1024*1024 { // < 1MB
        data, _ := io.ReadAll(resp.Body)
        s.cache.Add(imageUrl, data)
        return bytes.NewReader(data), nil
    }
    
    // 4. 大图片直接流式转发
    return resp.Body, nil
}
```

---

### 1.3 /api/favorites (收藏管理)

**当前实现分析:**
```typescript
// GET - 查询单条或全部
const fav = await db.getFavorite(authInfo.username, source, id);
const favorites = await db.getAllFavorites(authInfo.username);

// POST - 保存
await db.saveFavorite(authInfo.username, source, id, finalFavorite);

// DELETE - 删除单条或清空
await db.deleteFavorite(username, source, id);
await Promise.all(Object.keys(all).map(...));
```

**痛点:**
1. 认证逻辑重复 (每个 API 都检查)
2. 无事务保证
3. 批量删除使用 Promise.all，无并发控制

**Go 重构方案:**
```go
// 使用中间件统一认证
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authInfo := getAuthInfoFromCookie(c)
        if authInfo == nil {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        c.Set("username", authInfo.Username)
        c.Next()
    }
}

// Handler
func (h *FavoriteHandler) Get(c *gin.Context) {
    username := c.GetString("username")
    key := c.Query("key")
    
    if key != "" {
        // 查询单条
        fav, err := h.service.Get(c.Request.Context(), username, key)
        if err != nil {
            h.logger.Error("获取收藏失败", zap.Error(err))
            c.JSON(500, gin.H{"error": "Internal Server Error"})
            return
        }
        c.JSON(200, fav)
        return
    }
    
    // 查询全部
    favorites, err := h.service.GetAll(c.Request.Context(), username)
    if err != nil {
        h.logger.Error("获取收藏列表失败", zap.Error(err))
        c.JSON(500, gin.H{"error": "Internal Server Error"})
        return
    }
    c.JSON(200, favorites)
}
```

---

## 2. 核心 Service 分析

### 2.1 SearchService (搜索)

**功能职责:**
- 多源并发搜索
- 结果聚合与排序
- 缓存管理
- 超时控制

**关键方法:**
| 方法 | 输入 | 输出 | 复杂度 |
|-----|------|------|-------|
| Search | query, sites | []SearchResult | O(n*m) |
| searchSingle | site, query | []SearchResult | O(1) |
| searchWithCache | site, query, page | cached/remote | O(1)/O(network) |

**重构要点:**
1. 使用 `errgroup` 替代 `Promise.allSettled`
2. 使用 `context.WithTimeout` 替代 `Promise.race`
3. 使用 `sync.Pool` 复用请求对象
4. 使用 `lru.Cache` 替代 Map 缓存

### 2.2 Downstream (下游接口)

**功能职责:**
- 调用第三方影视源 API
- 解析 M3U8 播放链接
- 数据格式转换

**关键逻辑:**
```typescript
// 解析 vod_play_url
const vod_play_url_array = item.vod_play_url.split('$$$');
vod_play_url_array.forEach((url: string) => {
    const title_url_array = url.split('#');
    title_url_array.forEach((title_url: string) => {
        const episode_title_url = title_url.split('$');
        // 提取 m3u8 链接
    });
});
```

**重构要点:**
1. 预编译正则表达式
2. 使用 `strings.Split` 替代 JS split
3. 流式处理大结果集

### 2.3 StorageService (存储)

**当前架构:**
```
DbManager (统一入口)
├── RedisStorage (Redis)
├── UpstashRedisStorage (Upstash)
├── KvrocksStorage (Kvrocks)
└── null (LocalStorage)
```

**存储结构 (Redis):**
```
// 收藏
ZADD user:{username}:favorites {timestamp} {json}
KEYS user:{username}:favorites

// 播放记录
ZADD user:{username}:history {timestamp} {json}

// 搜索历史
LPUSH user:{username}:search_history {keyword}
LTRIM user:{username}:search_history 0 99
```

**重构要点:**
1. 使用 `go-redis` 客户端
2. 使用 Pipeline 批量操作
3. 使用 Lua 脚本保证原子性
4. 添加连接池配置

---

## 3. 性能瓶颈分析

### 3.1 当前性能数据 (估算)

| 场景 | 当前 (Node.js) | 瓶颈 |
|-----|---------------|------|
| 单源搜索 | 500-2000ms | 网络IO |
| 10源聚合 | 3000-8000ms | 并发控制差 |
| 图片代理 | 200-500ms | 无连接复用 |
| 收藏查询 | 50-100ms | Redis 延迟 |

### 3.2 Go 优化目标

| 场景 | 目标 | 优化手段 |
|-----|------|---------|
| 单源搜索 | 300-1000ms | 连接复用 + 超时控制 |
| 10源聚合 | 1000-2000ms | Goroutine并发 + 连接池 |
| 图片代理 | 50-100ms | 内存缓存 + 流式转发 |
| 收藏查询 | 10-30ms | Pipeline + 连接池 |

---

## 4. Go 服务设计

### 4.1 接口定义

```go
// 搜索服务
type SearchService interface {
    Search(ctx context.Context, query string, sites []ApiSite) ([]SearchResult, error)
    SearchSingle(ctx context.Context, site ApiSite, query string) ([]SearchResult, error)
    GetSuggestions(ctx context.Context, query string) ([]string, error)
}

// 图片代理服务
type ImageService interface {
    Proxy(ctx context.Context, url string) (io.ReadCloser, error)
    WarmCache(urls []string)
}

// 存储服务
type StorageService interface {
    GetFavorites(ctx context.Context, username string) ([]Favorite, error)
    AddFavorite(ctx context.Context, username string, fav Favorite) error
    DeleteFavorite(ctx context.Context, username string, id string) error
    
    GetPlayRecords(ctx context.Context, username string) ([]PlayRecord, error)
    SavePlayRecord(ctx context.Context, username string, record PlayRecord) error
    
    GetSearchHistory(ctx context.Context, username string) ([]string, error)
    AddSearchHistory(ctx context.Context, username string, keyword string) error
}
```

### 4.2 依赖注入

```go
type Server struct {
    searchService   SearchService
    imageService    ImageService
    storageService  StorageService
    config          *Config
    logger          *zap.Logger
}

func NewServer(cfg *Config, logger *zap.Logger) (*Server, error) {
    // 初始化存储
    redisClient := redis.NewClient(&redis.Options{
        Addr: cfg.Redis.Addr,
    })
    storageService := NewRedisStorageService(redisClient)
    
    // 初始化搜索
    httpClient := &http.Client{
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 10,
        },
    }
    searchService := NewSearchService(httpClient, storageService, logger)
    
    // 初始化图片代理
    imageService := NewImageService(httpClient, logger)
    
    return &Server{
        searchService:  searchService,
        imageService:   imageService,
        storageService: storageService,
        config:         cfg,
        logger:         logger,
    }, nil
}
```

---

## 5. 迁移检查清单

### 5.1 API 兼容检查
- [ ] 所有路由路径保持不变
- [ ] 请求/响应格式保持一致
- [ ] Cookie/认证机制保持一致
- [ ] 缓存头 (Cache-Control) 保持一致

### 5.2 功能对等检查
- [ ] 多源搜索 + 超时控制
- [ ] 搜索结果过滤 (yellow words)
- [ ] 图片代理 + 缓存头
- [ ] 收藏 CRUD
- [ ] 播放记录
- [ ] 搜索历史
- [ ] 管理员配置
- [ ] 直播源管理

### 5.3 性能检查
- [ ] 并发压力测试 (wrk)
- [ ] 内存占用监控
- [ ] 图片加载速度对比
- [ ] 搜索响应时间对比
