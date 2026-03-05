---
name: refactor-moontv
description: ManboTV (MoonTV) 重构开发完整指南 - 前后端分离重构
---

# ManboTV 重构开发 Skill

## 项目概述

将 MoonTV 影视聚合播放器从 Next.js 全栈重构为前后端分离架构：

- **前端**: Next.js 14 (保留)
- **后端**: Go 1.21+ + Gin (新建)

## 核心目标

1. 性能提升: 并发 10x, 内存 5x
2. 代码质量: 文件 ≤800 行, 无魔法数值
3. 可维护性: 清晰分层, 完善文档

---

## 代码质量红线 (不可违反)

### 1. 文件行数限制

- **规则**: 任何代码文件不得超过 **800 行**
- **范围**: Go/TS/JS/CSS (HTML 除外)
- **检查**: `find . -name "*.go" -exec wc -l {} + | awk '$1 > 800'`

### 2. 禁止魔法数值

- **规则**: 所有数值必须定义为具名常量
- **示例**:

  ```go
  // ❌ 禁止
  time.Sleep(100 * time.Millisecond)

  // ✅ 正确
  const DefaultRetryDelay = 100 * time.Millisecond
  time.Sleep(DefaultRetryDelay)
  ```

### 3. 模块重构后必须即时验证

- **规则**: 每完成一个重构模块，必须立即执行该模块的 Docker + curl 验证，验证通过后才能继续下一个模块。
- **最低要求**:
  1. `docker compose up -d --build` 启动最新服务
  2. 用 `curl` 覆盖该模块最少一组 `GET/POST/DELETE`（按模块接口实际情况）
  3. 记录请求与响应结果，失败必须先修复再进入下一模块

### 4. 沟通语言规则

- 默认使用中文进行沟通与汇报
- 仅在用户明确指定时使用其他语言

---

## Go 后端工程规范

### 错误处理

```go
// 包装错误保留上下文
return fmt.Errorf("search from %s failed: %w", site.Name, err)

// 错误分类
var ErrNotFound = errors.New("resource not found")
if errors.Is(err, ErrNotFound) { ... }
```

### Context 使用

```go
// 所有阻塞操作必须接收 Context
func Search(ctx context.Context, query string) ([]Result, error)

// 设置超时
ctx, cancel := context.WithTimeout(parentCtx, 20*time.Second)
defer cancel()
```

### 资源管理

```go
// defer 紧跟资源获取
f, err := os.Open(file)
if err != nil { return err }
defer f.Close()

// HTTP Client 复用 (全局)
var httpClient = &http.Client{
    Transport: &http.Transport{
        MaxIdleConns: 100,
        MaxIdleConnsPerHost: 10,
    },
}
```

### 并发安全

```go
// Mutex 字段紧邻
type Cache struct {
    mu    sync.RWMutex
    items map[string]Item
}

// 最小化加锁范围
func (c *Cache) Get(key string) {
    c.mu.RLock()
    item := c.items[key]
    c.mu.RUnlock()
    // 其他操作...
}

// 使用 errgroup 控制并发
import "golang.org/x/sync/errgroup"

g, ctx := errgroup.WithContext(ctx)
for _, site := range sites {
    site := site // 捕获
    g.Go(func() error {
        return search(site)
    })
}
if err := g.Wait(); err != nil { ... }
```

### 日志规范

```go
// 结构化日志
logger.Info("search completed",
    zap.String("query", query),
    zap.Int("results", len(results)),
    zap.Duration("duration", elapsed),
)

// ❌ 禁止字符串拼接
logger.Info(fmt.Sprintf("search %s completed", query))
```

---

## 项目结构

```
ManboTv/
├── frontend/              # 前端 (原项目保留)
│   ├── src/
│   └── package.json
│
├── backend/               # Go 后端 (新建)
│   ├── cmd/server/
│   │   └── main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── handler/       # HTTP处理器
│   │   ├── service/       # 业务逻辑
│   │   ├── repository/    # 数据访问
│   │   ├── middleware/
│   │   ├── model/
│   │   └── util/
│   ├── pkg/
│   ├── configs/
│   ├── go.mod
│   └── Dockerfile
│
└── docs/                  # 文档
    ├── requirements/      # PRD需求文档
    └── refactor/          # 重构规则
```

---

## API 契约

### 统一响应

```go
type Response struct {
    Code    int         `json:"code"`    // 0=成功
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}
```

### 核心接口

- `GET /api/v1/search?q=xxx` - 聚合搜索
- `GET /api/v1/image?url=xxx` - 图片代理
- `GET /api/v1/detail?id=xxx` - 详情
- `POST /api/v1/favorites` - 收藏
- `POST /api/v1/playrecords` - 播放记录

---

## 开发检查清单

### 新增功能时检查

- [ ] 文件行数 ≤ 800
- [ ] 无常量数值直接出现在代码中
- [ ] 函数接收 Context 参数
- [ ] 错误使用 fmt.Errorf + %w 包装
- [ ] HTTP Client 使用全局复用
- [ ] 资源使用 defer 释放
- [ ] 并发使用 errgroup/semaphore 控制
- [ ] 日志使用结构化字段
- [ ] 有单元测试覆盖

### 代码审查检查

- [ ] 无资源泄漏 (defer close)
- [ ] 无 Goroutine 泄漏
- [ ] 敏感信息不打印日志
- [ ] 配置不硬编码
- [ ] SQL 使用参数化查询

---

## 参考文档

| 文档        | 路径                                 |
| ----------- | ------------------------------------ |
| 架构分析    | `docs/refactor/README.md`            |
| API 映射    | `docs/refactor/API_MAPPING.md`       |
| 开发规则    | `docs/refactor/DEVELOPMENT_RULES.md` |
| Go 后端规则 | `docs/refactor/GO_BACKEND_RULES.md`  |
| PRD 总览    | `docs/requirements/PRD_OVERVIEW.md`  |
| 前端 PRD    | `docs/requirements/PRD_FRONTEND.md`  |
| 后端 PRD    | `docs/requirements/PRD_BACKEND.md`   |
| API 规范    | `docs/requirements/PRD_API.md`       |
| 数据库设计  | `docs/requirements/PRD_DATABASE.md`  |

---

## 常用命令

```bash
# 检查文件行数
find . -name "*.go" -exec wc -l {} + | awk '$1 > 800'

# 运行测试
go test -cover ./...

# 构建
go build -o bin/server cmd/server/main.go

# 运行
./bin/server -config configs/config.yaml
```
