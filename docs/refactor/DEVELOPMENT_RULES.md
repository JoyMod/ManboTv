# 🛠️ 开发规则与约束

## 1. 代码规范 (Code Style)

### Go 代码规范
- **格式**: 使用 `gofmt` 自动格式化
- **命名**: 
  - 导出成员: PascalCase (如 `SearchService`)
  - 私有成员: camelCase (如 `searchCache`)
  - 包名: 小写单数 (如 `service`, `repository`)
- **错误处理**: 必须处理所有错误, 禁止忽略
  ```go
  // ✅ 正确
  result, err := service.Search(ctx, query)
  if err != nil {
      return nil, fmt.Errorf("search failed: %w", err)
  }
  
  // ❌ 错误
  result, _ := service.Search(ctx, query)
  ```
- **注释**: 导出成员必须有注释, 注释以成员名开头
  ```go
  // SearchService 处理搜索相关业务逻辑
  type SearchService struct {}
  ```

### Git 提交规范
- **格式**: `<type>(<scope>): <subject>`
- **类型**:
  - `feat`: 新功能
  - `fix`: 修复
  - `refactor`: 重构
  - `perf`: 性能优化
  - `docs`: 文档
  - `chore`: 构建/工具
- **示例**:
  ```
  feat(search): 实现多源并发搜索
  fix(proxy): 修复图片代理内存泄漏
  refactor(config): 重构配置加载逻辑
  ```

---

## 2. 项目结构 (Project Structure)

```
moontv-backend/              # Go 后端项目
├── cmd/
│   └── server/              # 主程序入口
│       └── main.go
├── internal/                # 私有代码
│   ├── config/              # 配置管理
│   │   ├── config.go        # 配置结构体 & 加载
│   │   └── types.go         # 配置类型定义
│   ├── handler/             # HTTP 处理器 (Controller)
│   │   ├── search_handler.go
│   │   ├── proxy_handler.go
│   │   ├── image_handler.go
│   │   ├── favorite_handler.go
│   │   ├── live_handler.go
│   │   └── admin_handler.go
│   ├── service/             # 业务逻辑层
│   │   ├── search_service.go    # 搜索业务
│   │   ├── proxy_service.go     # 代理业务
│   │   ├── image_service.go     # 图片代理
│   │   ├── storage_service.go   # 存储抽象
│   │   ├── live_service.go      # 直播管理
│   │   └── admin_service.go     # 后台管理
│   ├── repository/          # 数据访问层
│   │   ├── redis/
│   │   │   └── client.go
│   │   └── sqlite/
│   │       └── client.go
│   ├── middleware/          # Gin 中间件
│   │   ├── cors.go
│   │   ├── auth.go
│   │   ├── logger.go
│   │   └── ratelimit.go
│   ├── model/               # 数据模型
│   │   ├── search.go
│   │   ├── favorite.go
│   │   ├── playrecord.go
│   │   └── live.go
│   └── util/                # 工具函数
│       ├── http.go
│       ├── crypto.go
│       └── time.go
├── pkg/                     # 公共库 (可被外部使用)
│   └── cache/
│       └── lru.go
├── api/                     # API 定义
│   └── openapi.yaml
├── configs/                 # 配置文件
│   ├── config.yaml
│   └── config.prod.yaml
├── scripts/                 # 脚本
│   └── build.sh
├── Dockerfile
├── Makefile
├── go.mod
└── README.md

frontend/                    # 前端 (保持原样)
├── src/
├── public/
├── next.config.js
└── package.json
```

---

## 3. 技术约束 (Constraints)

### 依赖管理
- **Go 版本**: >= 1.21
- **核心依赖**:
  ```go
  require (
      github.com/gin-gonic/gin v1.9.1      // Web框架
      github.com/redis/go-redis/v9 v9.3.0  // Redis客户端
      github.com/spf13/viper v1.18.0       // 配置管理
      go.uber.org/zap v1.26.0              // 日志
      golang.org/x/sync v0.6.0             // 并发工具
  )
  ```

### 性能约束
- **超时控制**: 所有外部 HTTP 调用必须有超时
  ```go
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()
  ```
- **并发限制**: 使用 `semaphore` 或 `worker pool` 限制并发数
- **连接池**: HTTP Client 必须复用连接
  ```go
  client := &http.Client{
      Transport: &http.Transport{
          MaxIdleConns:        100,
          MaxIdleConnsPerHost: 10,
          IdleConnTimeout:     90 * time.Second,
      },
  }
  ```

### 安全约束
- **CORS**: 只允许特定域名访问
- **认证**: JWT Token, 有效期 24h
- **速率限制**: IP 级别限流 (100 req/min)
- **输入验证**: 所有参数必须验证

---

## 4. 接口契约 (API Contract)

### 统一响应格式
```go
type Response struct {
    Code    int         `json:"code"`    // 0=成功, 非0=错误码
    Message string      `json:"message"` // 提示信息
    Data    interface{} `json:"data"`    // 数据
}

// 成功响应
{
    "code": 0,
    "message": "success",
    "data": {...}
}

// 错误响应
{
    "code": 1001,
    "message": "参数错误",
    "data": null
}
```

### 分页格式
```go
type Pagination struct {
    Page      int         `json:"page"`
    PageSize  int         `json:"page_size"`
    Total     int         `json:"total"`
    TotalPage int         `json:"total_page"`
    List      interface{} `json:"list"`
}
```

---

## 5. 测试规范

### 单元测试
- 所有 Service 层必须有单元测试
- 覆盖率 >= 60%
- 命名: `xxx_test.go`

### 集成测试
- API Handler 层必须有集成测试
- 使用 `httptest` 模拟 HTTP 请求

---

## 6. 部署规范

### 构建
```bash
# 开发
make dev

# 生产构建 (交叉编译)
make build-linux
make build-darwin
make build-windows
```

### Docker
- 多阶段构建
- 镜像大小 < 50MB (使用 alpine/scratch)
- 非 root 用户运行


---

## 7. 代码质量红线 (Code Quality Red Lines) ⛔

### 7.1 文件行数限制

**规则**: 单个代码文件不得超过 **800 行** (不含空行和注释)

**适用范围**:
- ✅ Go 后端代码 (`.go`)
- ✅ TypeScript/JavaScript 前端代码 (`.ts`, `.tsx`, `.js`, `.jsx`)
- ✅ CSS/SCSS 样式文件
- ❌ HTML 模板文件 (除外)
- ❌ 自动生成的代码 (除外)

**检查命令**:
```bash
# Go 文件行数检查
find . -name "*.go" -not -path "./vendor/*" -exec wc -l {} + | awk '$1 > 800 {print $2 ": " $1 " lines"}'

# TS/TSX 文件行数检查
find . -name "*.ts" -o -name "*.tsx" | grep -v node_modules | xargs wc -l | awk '$1 > 800 {print $2 ": " $1 " lines"}'
```

**拆分策略**:
```go
// ❌ 错误: search_service.go (1200 行)
// 包含: 搜索逻辑 + 缓存逻辑 + 结果排序 + 过滤逻辑

// ✅ 正确: 拆分为多个文件
// search/
// ├── service.go          (300 行) - 主服务入口
// ├── cache.go            (200 行) - 缓存管理
// ├── sorter.go           (150 行) - 结果排序
// └── filter.go           (200 行) - 结果过滤
```

### 7.2 禁止魔法数值 (No Magic Numbers)

**定义**: 魔法数值是指直接在代码中出现的、没有明确含义说明的常量数值。

**规则**: 所有数值常量必须定义为具名常量 (const)，且命名能表达其业务含义。

#### 错误示例 ❌

```go
// Go 后端 - 错误
func Search(ctx context.Context, query string) ([]Result, error) {
    ctx, cancel := context.WithTimeout(ctx, 20000)  // 魔法数值: 20000 是什么?
    defer cancel()
    
    if len(results) > 50 {  // 魔法数值: 50 是什么?
        results = results[:50]
    }
    
    time.Sleep(100 * time.Millisecond)  // 魔法数值: 100
    return results, nil
}
```

```typescript
// TypeScript 前端 - 错误
function fetchData() {
    const timeout = 30000;  // 魔法数值
    const maxRetries = 3;   // 魔法数值
    const pageSize = 20;    // 魔法数值
    
    if (items.length > 100) {  // 魔法数值
        items = items.slice(0, 100);
    }
}
```

#### 正确示例 ✅

```go
// Go 后端 - 正确
package search

// 搜索相关常量
const (
    DefaultSearchTimeout     = 20 * time.Second  // 默认搜索超时时间
    MaxSearchResults         = 50                // 最大搜索结果数
    DefaultRetryDelay        = 100 * time.Millisecond  // 默认重试间隔
    MaxConcurrentSources     = 10                // 最大并发搜索源
    CacheExpirationMinutes   = 15                // 缓存过期时间(分钟)
)

func Search(ctx context.Context, query string) ([]Result, error) {
    ctx, cancel := context.WithTimeout(ctx, DefaultSearchTimeout)
    defer cancel()
    
    if len(results) > MaxSearchResults {
        results = results[:MaxSearchResults]
    }
    
    time.Sleep(DefaultRetryDelay)
    return results, nil
}
```

```typescript
// TypeScript 前端 - 正确
// constants/api.ts
export const API_CONSTANTS = {
    TIMEOUT: {
        DEFAULT: 30000,      // 默认请求超时 (ms)
        SEARCH: 20000,       // 搜索请求超时 (ms)
        UPLOAD: 60000,       // 上传请求超时 (ms)
    },
    RETRY: {
        MAX_ATTEMPTS: 3,     // 最大重试次数
        DELAY_MS: 1000,      // 重试间隔 (ms)
    },
    PAGINATION: {
        DEFAULT_PAGE_SIZE: 20,   // 默认每页条数
        MAX_PAGE_SIZE: 100,      // 最大每页条数
        FIRST_PAGE: 1,           // 起始页码
    },
} as const;

// hooks/useSearch.ts
import { API_CONSTANTS } from '@/constants/api';

function useSearch() {
    const { DEFAULT_PAGE_SIZE, MAX_PAGE_SIZE } = API_CONSTANTS.PAGINATION;
    
    const fetchData = async (pageSize: number = DEFAULT_PAGE_SIZE) => {
        const limit = Math.min(pageSize, MAX_PAGE_SIZE);
        // ...
    };
}
```

#### 常见数值分类

```go
// config/constants.go - Go 后端统一常量定义
package config

import "time"

// HTTP 相关
const (
    HTTPTimeoutDefault       = 10 * time.Second
    HTTPTimeoutSearch        = 20 * time.Second
    HTTPTimeoutImageProxy    = 30 * time.Second
    HTTPMaxIdleConns         = 100
    HTTPMaxConnsPerHost      = 10
)

// 搜索相关
const (
    SearchMaxResults         = 50
    SearchMaxConcurrent      = 10
    SearchCacheMinutes       = 15
    SearchSuggestionLimit    = 10
)

// 业务限制
const (
    MaxFavoritesPerUser      = 1000
    MaxPlayHistoryDays       = 30
    MaxSearchHistoryItems    = 100
)

// 安全相关
const (
    JWTExpirationHours       = 24
    RateLimitRequestsPerMin  = 100
    PasswordMinLength        = 6
    PasswordMaxLength        = 32
)
```

```typescript
// src/constants/index.ts - 前端统一常量定义

// 播放器相关
export const PLAYER = {
    SEEK_STEP_SECONDS: 10,           // 快进/快退步长(秒)
    VOLUME_STEP: 0.1,                // 音量调节步长
    MIN_PLAYBACK_RATE: 0.5,          // 最小播放速度
    MAX_PLAYBACK_RATE: 2.0,          // 最大播放速度
    PLAYBACK_RATE_STEP: 0.25,        // 播放速度调节步长
    SKIP_INTRO_SECONDS: 85,          // 跳过片头默认时长
    SKIP_OUTRO_SECONDS: 120,         // 跳过片尾默认时长
} as const;

// UI 相关
export const UI = {
    DEBOUNCE_DELAY_MS: 300,          // 防抖延迟
    TOAST_DURATION_MS: 3000,         // Toast 显示时长
    MODAL_ANIMATION_MS: 200,         // 弹窗动画时长
    INFINITE_SCROLL_THRESHOLD: 100,  // 无限滚动触发阈值(px)
} as const;

// 缓存相关
export const CACHE = {
    SEARCH_TTL_MINUTES: 15,          // 搜索结果缓存时间
    IMAGE_TTL_HOURS: 24,             // 图片缓存时间
    CONFIG_TTL_MINUTES: 5,           // 配置缓存时间
} as const;
```

### 7.3 代码审查清单

提交 PR 前必须自查:

- [ ] 所有代码文件行数 ≤ 800 行
- [ ] 无常量数值直接出现在业务逻辑中
- [ ] 所有 `const` 都有注释说明用途和单位
- [ ] 相同含义的数值没有在多处重复定义

**自动化检查** (GitHub Actions):
```yaml
name: Code Quality Check
on: [pull_request]
jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Check File Line Count
        run: |
          find . -name "*.go" -exec wc -l {} + | awk '$1 > 800 {exit 1}'
      - name: Check Magic Numbers (Go)
        run: |
          # 使用 go-critic 检查魔法数值
          go install github.com/go-critic/go-critic/cmd/go-critic@latest
          go-critic check -enable=builtinShadow,magicNumbers ./...
```

---

## 8. 文件组织规范

### 8.1 目录命名
```
✅ 正确:
├── search/
├── favorite/
├── play_record/
├── admin/

❌ 错误:
├── searchService/
├── FavoriteUtils/
├── play-record/
```

### 8.2 文件命名
```
✅ 正确:
├── search_service.go
├── image_handler.go
├── user_types.go
├── redis_client.go

❌ 错误:
├── searchService.go
├── ImageHandler.go
├── user-types.go
├── redis.go (太笼统)
```

### 8.3 拆分示例

**原始文件 (1000+ 行) - 需要拆分**:
```go
// admin_service.go - 1000 行
func (s *AdminService) ManageUser() {}
func (s *AdminService) ManageSource() {}
func (s *AdminService) ManageCategory() {}
func (s *AdminService) ManageLive() {}
func (s *AdminService) DataExport() {}
func (s *AdminService) DataImport() {}
```

**拆分后**:
```
admin/
├── service.go           (100 行) - AdminService 定义和初始化
├── user_service.go      (150 行) - 用户管理
├── source_service.go    (200 行) - 源管理
├── category_service.go  (150 行) - 分类管理
├── live_service.go      (150 行) - 直播管理
├── data_export.go       (100 行) - 数据导出
└── data_import.go       (100 行) - 数据导入
```
