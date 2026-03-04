# 后端需求规格文档

## 技术栈

- **语言**: Go 1.21+
- **Web框架**: Gin
- **ORM**: GORM
- **缓存**: Redis (go-redis)
- **配置**: Viper
- **日志**: Zap
- **验证**: validator/v10

---

## 项目结构

```
backend/
├── cmd/server/           # 入口
├── internal/
│   ├── config/          # 配置
│   ├── handler/         # HTTP处理器
│   ├── service/         # 业务逻辑
│   ├── repository/      # 数据访问
│   ├── middleware/      # 中间件
│   ├── model/           # 数据模型
│   └── util/            # 工具
├── pkg/                 # 公共库
├── configs/             # 配置文件
└── scripts/             # 脚本
```

---

## 核心服务

### SearchService (搜索服务)
```go
type SearchService interface {
    Search(ctx context.Context, query string, sites []ApiSite) ([]SearchResult, error)
    SearchSingle(ctx context.Context, site ApiSite, query string) ([]SearchResult, error)
    GetSuggestions(ctx context.Context, query string) ([]string, error)
}
```

### ProxyService (代理服务)
```go
type ProxyService interface {
    ProxyImage(ctx context.Context, url string) (io.Reader, error)
    ProxyM3U8(ctx context.Context, url string) (string, error)
    ProxySegment(ctx context.Context, url string) (io.Reader, error)
}
```

### StorageService (存储服务)
```go
type StorageService interface {
    GetFavorites(ctx context.Context, userID string) ([]Favorite, error)
    AddFavorite(ctx context.Context, userID string, fav Favorite) error
    DeleteFavorite(ctx context.Context, userID string, id string) error
    
    GetPlayRecords(ctx context.Context, userID string) ([]PlayRecord, error)
    SavePlayRecord(ctx context.Context, userID string, record PlayRecord) error
}
```

---

## 中间件

```go
// 注册顺序
r.Use(
    middleware.Recovery(),      //  panic 恢复
    middleware.Logger(),        //  请求日志
    middleware.CORS(),          //  跨域
    middleware.RateLimiter(),   //  限流
    middleware.Auth(),          //  认证 (部分路由)
)
```

---

## 配置项

```yaml
server:
  port: 8080
  mode: release  # debug/release

redis:
  addr: localhost:6379
  password: ""
  db: 0

search:
  timeout: 20s
  max_concurrent: 10
  cache_minutes: 15

log:
  level: info
  format: json
```
