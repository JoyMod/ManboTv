---
name: refactor-moontv
description: MoonTV 重构开发指南 - Next.js 后端迁移到 Go + Gin
---

# MoonTV 重构 Skill

## 项目背景
将 MoonTV (影视聚合播放器) 的后端从 Next.js API Routes 迁移到 Go + Gin，解决并发性能问题。

## 核心原则
1. **前端保持原样** - 只改动 API 调用地址
2. **接口兼容** - Go 后端完全兼容原有 API 契约
3. **性能优先** - Goroutine 并发、连接池、缓存
4. **渐进重构** - 先核心功能 (search/proxy)，后边缘功能

## 关键技术决策

### 1. 并发模型
```go
// 多源搜索 - 使用 errgroup 并发
import "golang.org/x/sync/errgroup"

func (s *SearchService) SearchMulti(ctx context.Context, query string, sites []ApiSite) ([]SearchResult, error) {
    g, ctx := errgroup.WithContext(ctx)
    results := make([][]SearchResult, len(sites))
    
    for i, site := range sites {
        i, site := i, site // 闭包捕获
        g.Go(func() error {
            res, err := s.searchSingle(ctx, site, query)
            if err != nil {
                return nil // 忽略单个源错误
            }
            results[i] = res
            return nil
        })
    }
    
    if err := g.Wait(); err != nil {
        return nil, err
    }
    
    return flatten(results), nil
}
```

### 2. HTTP Client 复用
```go
// 全局复用 Client (不要每个请求 new)
var httpClient = &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

### 3. 图片代理优化
```go
// 1. 内存缓存热点图片
// 2. 流式转发 (不全部读入内存)
func (h *ImageHandler) Proxy(c *gin.Context) {
    imageUrl := c.Query("url")
    
    resp, err := httpClient.Get(imageUrl)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    defer resp.Body.Close()
    
    // 流式转发
    c.DataFromReader(resp.StatusCode, resp.ContentLength, 
        resp.Header.Get("Content-Type"), resp.Body, nil)
}
```

## 开发检查清单

### 新增 API 时检查
- [ ] 路由注册在 `internal/handler/xxx_handler.go`
- [ ] 业务逻辑在 `internal/service/xxx_service.go`
- [ ] 接口契约符合统一响应格式
- [ ] 有超时控制 (`context.WithTimeout`)
- [ ] 错误日志记录 (zap)
- [ ] CORS 配置正确

### 代码审查检查
- [ ] 无资源泄漏 (defer close)
- [ ] 无 Goroutine 泄漏 (使用 WaitGroup/errgroup)
- [ ] 敏感信息不打印日志
- [ ] 配置不硬编码

## 常见陷阱

### ❌ 错误: 每个请求新建 HTTP Client
```go
// 错误 - 导致连接不复用
client := &http.Client{}
resp, _ := client.Get(url)
```

### ✅ 正确: 复用全局 Client
```go
// 正确 - 连接复用
resp, _ := httpClient.Get(url)
```

### ❌ 错误: 不控制并发数
```go
// 错误 - 可能创建数千 Goroutine
for _, site := range sites {
    go search(site) // 无限制
}
```

### ✅ 正确: 使用信号量控制
```go
// 正确 - 最多10个并发
sem := make(chan struct{}, 10)
for _, site := range sites {
    sem <- struct{}{}
    go func() {
        defer func() { <-sem }()
        search(site)
    }()
}
```

## 前端适配

### API 地址切换
```typescript
// 原来: 相对路径 (同域)
const API_BASE = '/api'

// 重构后: 可配置后端地址
const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1'
```

### CORS 配置
Go 后端需配置允许前端域名:
```go
config := cors.Config{
    AllowOrigins:     []string{"http://localhost:3000"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
}
```

## 测试策略

1. **单元测试**: `go test ./internal/service/...`
2. **集成测试**: 使用 `httptest` 测试 Handler
3. **性能测试**: 
   ```bash
   # 压测搜索接口
   wrk -t12 -c400 -d30s "http://localhost:8080/api/v1/search?q=电影"
   ```

## 部署检查

```bash
# 1. 构建
make build

# 2. 测试运行
./bin/moontv-server -config configs/config.yaml

# 3. 健康检查
curl http://localhost:8080/health

# 4. 前端构建 (静态导出)
cd frontend && npm run export
```
