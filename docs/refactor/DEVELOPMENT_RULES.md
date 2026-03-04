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

