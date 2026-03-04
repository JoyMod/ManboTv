# Go 后端工程实践规则

## 1. 错误处理规范 (Error Handling)

### 错误包装 (Error Wrapping)
```go
// ✅ 正确: 使用 fmt.Errorf + %w 包装错误
if err != nil {
    return fmt.Errorf("search from %s failed: %w", site.Name, err)
}

// ❌ 错误: 丢失原始错误信息
if err != nil {
    return errors.New("search failed")
}
```

### 错误分类定义
```go
package errors

var (
    ErrNotFound     = errors.New("resource not found")
    ErrUnauthorized = errors.New("unauthorized")
    ErrForbidden    = errors.New("forbidden")
    ErrValidation   = errors.New("validation failed")
    ErrRateLimited  = errors.New("rate limited")
    ErrTimeout      = errors.New("request timeout")
)
```

---

## 2. Context 使用规范

### 强制使用场景
```go
// ✅ 必须传入 Context
func (s *SearchService) Search(ctx context.Context, query string) ([]Result, error)

// ❌ 禁止: 不接收 Context
func (s *SearchService) Search(query string) ([]Result, error)
```

### Context 超时
```go
// ✅ 在服务入口设置超时
func (h *SearchHandler) Handle(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), DefaultSearchTimeout)
    defer cancel()
    
    results, err := h.service.Search(ctx, query)
}
```

---

## 3. 资源管理规范

### defer 使用
```go
// ✅ 立即 defer，紧跟资源获取
f, err := os.Open(file)
if err != nil {
    return err
}
defer f.Close()

// ✅ 成对资源管理
func process() error {
    conn, err := pool.Get()
    if err != nil {
        return err
    }
    defer conn.Close()
    
    tx, err := conn.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    return tx.Commit()
}
```

### HTTP Client 复用
```go
// ✅ 全局复用 Client
var httpClient = &http.Client{
    Timeout: 10 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
}
```

---

## 4. 并发控制规范 (强制使用 Channel)

### 4.1 Channel 作为信号量 (Semaphore)
```go
// ✅ 正确: 使用 Channel 控制并发数 (推荐)
const MaxConcurrency = 10

semaphore := make(chan struct{}, MaxConcurrency)

for _, task := range tasks {
    task := task // 捕获
    
    select {
    case semaphore <- struct{}{}: // 获取信号量
        go func() {
            defer func() { <-semaphore }() // 释放信号量
            process(task)
        }()
    case <-ctx.Done():
        return // 上下文取消
    }
}

// ❌ 错误: 使用 WaitGroup + Mutex 计数
var wg sync.WaitGroup
var mu sync.Mutex
var count int

for _, task := range tasks {
    wg.Add(1)
    go func() {
        defer wg.Done()
        mu.Lock()
        count++
        mu.Unlock()
        process(task)
        mu.Lock()
        count--
        mu.Unlock()
    }()
}
```

### 4.2 Channel 收集结果 (替代锁)
```go
// ✅ 正确: 使用 Channel 收集结果，避免锁竞争
resultChan := make(chan Result, len(tasks))

for _, task := range tasks {
    go func(t Task) {
        resultChan <- process(t)
    }(task)
}

// 收集结果
for i := 0; i < len(tasks); i++ {
    result := <-resultChan
    allResults = append(allResults, result)
}

// ❌ 错误: 使用 Mutex 保护切片
var mu sync.Mutex
var results []Result

for _, task := range tasks {
    go func(t Task) {
        r := process(t)
        mu.Lock()
        results = append(results, r)
        mu.Unlock()
    }(task)
}
```

### 4.3 超时控制 (Select + Channel)
```go
// ✅ 正确: 使用 select 实现超时
type Result struct {
    data []Item
    err  error
}

resultChan := make(chan Result, 1)

go func() {
    data, err := fetchData()
    resultChan <- Result{data, err}
}()

select {
case result := <-resultChan:
    return result.data, result.err
case <-time.After(5 * time.Second):
    return nil, errors.New("timeout")
case <-ctx.Done():
    return nil, ctx.Err()
}
```

### 4.4 原子操作 (Atomic) 替代 Mutex
```go
// ✅ 正确: 使用 atomic 替代 Mutex 做简单计数
var completed int32

go func() {
    atomic.AddInt32(&completed, 1)
}()

// 读取
count := atomic.LoadInt32(&completed)

// ❌ 错误: 用 Mutex 做简单计数
var mu sync.Mutex
var count int

mu.Lock()
count++
mu.Unlock()
```

### 4.5 扇出-扇入模式 (Fan-out/Fan-in)
```go
// Producer (扇出)
func producer(jobs []Job) <-chan Job {
    out := make(chan Job, len(jobs))
    for _, job := range jobs {
        out <- job
    }
    close(out)
    return out
}

// Worker (处理)
func worker(in <-chan Job) <-chan Result {
    out := make(chan Result)
    go func() {
        defer close(out)
        for job := range in {
            out <- process(job)
        }
    }()
    return out
}

// Consumer (扇入)
func consumer(channels ...<-chan Result) <-chan Result {
    var wg sync.WaitGroup
    out := make(chan Result)
    
    output := func(c <-chan Result) {
        defer wg.Done()
        for result := range c {
            out <- result
        }
    }
    
    wg.Add(len(channels))
    for _, c := range channels {
        go output(c)
    }
    
    go func() {
        wg.Wait()
        close(out)
    }()
    
    return out
}
```

### 4.6 Mutex 使用 (仅限复杂状态)
```go
// ✅ 字段紧邻 mutex
type Cache struct {
    mu     sync.RWMutex
    items  map[string]Item
    ttl    time.Duration
}

// ✅ 最小化加锁范围
func (c *Cache) Get(key string) (Item, bool) {
    c.mu.RLock()
    item, exists := c.items[key]
    c.mu.RUnlock()
    
    if !exists {
        return Item{}, false
    }
    
    // 过期检查不需要加锁
    if time.Now().After(item.ExpireAt) {
        c.Delete(key)
        return Item{}, false
    }
    
    return item, true
}
```

### Channel 使用
```go
// ✅ 带缓冲 channel
results := make(chan Result, 10)

// ✅ 发送方关闭
go func() {
    defer close(results)
    for _, site := range sites {
        results <- search(site)
    }
}()
```

---

## 5. 接口设计规范

### RESTful URL
```
GET    /api/v1/movies              # 列表
GET    /api/v1/movies/:id          # 详情
POST   /api/v1/movies/:id/favorite # 收藏
DELETE /api/v1/movies/:id/favorite # 取消收藏
```

### 统一响应
```go
type Response struct {
    Code    int         `json:"code"`    // 0=成功
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}
```

---

## 6. 日志规范

### 结构化日志
```go
// ✅ 使用字段
logger.Info("request completed",
    zap.String("method", c.Request.Method),
    zap.String("path", c.Request.URL.Path),
    zap.Int("status", c.Writer.Status()),
    zap.Duration("latency", time.Since(start)),
)

// ❌ 禁止拼接
logger.Info(fmt.Sprintf("request %s completed", path))
```

---

## 7. 数据库/SQL 规范

### 参数化查询
```go
// ✅ 防止 SQL 注入
db.Query("SELECT * FROM users WHERE id = ?", userID)

// ❌ 禁止字符串拼接
db.Query("SELECT * FROM users WHERE id = " + userID)
```

### 连接池配置
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(5 * time.Minute)
```

---

## 8. 测试规范

### 单元测试
```go
func TestSearchService_Search(t *testing.T) {
    // Arrange
    mockCache := NewMockCache()
    service := NewSearchService(mockCache)
    
    // Act
    results, err := service.Search(context.Background(), "test")
    
    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, results)
}
```

### 覆盖率要求
- 单元测试覆盖率 ≥ 60%
- 核心业务逻辑覆盖率 ≥ 80%

---

## 9. 性能优化检查项

```go
// ✅ 使用 sync.Pool
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 1024)
    },
}

// ✅ 预分配切片
results := make([]Result, 0, expectedSize)

// ✅ strings.Builder
var b strings.Builder
b.WriteString("prefix")
result := b.String()

// ✅ strconv 替代 fmt
strconv.Itoa(num)      // 快
fmt.Sprintf("%d", num) // 慢
```

---

## 10. 日志规范 (强制使用 Zap)

### 10.1 必须使用 Uber Zap

**规则**: 所有日志必须使用 `go.uber.org/zap`，禁止标准库 `log` 或其他日志库。

#### 初始化
```go
// config/logger.go
package config

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func NewLogger(env string) (*zap.Logger, error) {
    var cfg zap.Config
    
    if env == "production" {
        cfg = zap.NewProductionConfig()
        cfg.EncoderConfig.TimeKey = "timestamp"
        cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    } else {
        cfg = zap.NewDevelopmentConfig()
        cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
    }
    
    return cfg.Build()
}

// 全局日志实例
var Log *zap.Logger

func InitLogger(env string) error {
    var err error
    Log, err = NewLogger(env)
    return err
}
```

#### 使用示例
```go
// ✅ 正确: 结构化日志
logger.Info("搜索完成",
    zap.String("query", query),
    zap.Int("results_count", len(results)),
    zap.Duration("elapsed", time.Since(start)),
    zap.String("client_ip", c.ClientIP()),
)

logger.Error("搜索失败",
    zap.String("query", query),
    zap.Error(err),
    zap.Strings("sources", sourceNames),
)

logger.Debug("缓存命中",
    zap.String("key", cacheKey),
    zap.Duration("ttl", ttl),
)

// ❌ 禁止: 字符串拼接
logger.Info(fmt.Sprintf("搜索 %s 完成，找到 %d 条结果", query, count))

// ❌ 禁止: 使用标准库 log
log.Printf("搜索完成: %s", query)
```

#### 字段命名规范
```go
// 通用字段
zap.String("trace_id", traceID)       // 追踪ID
zap.String("user_id", userID)         // 用户ID
zap.String("client_ip", clientIP)     // 客户端IP
zap.Duration("latency", latency)      // 延迟
zap.Int("status_code", statusCode)    // HTTP状态码

// 业务字段
zap.String("query", query)            // 搜索关键词
zap.String("source", source)          // 数据来源
zap.Int("result_count", count)        // 结果数量
zap.String("vod_id", vodID)           // 影片ID
zap.Error(err)                        // 错误对象
```

#### 日志级别使用
```go
// DEBUG: 开发调试信息
logger.Debug("进入函数", zap.String("func", "Search"))

// INFO: 关键业务流程
logger.Info("搜索完成", zap.String("query", query))

// WARN: 可恢复的错误或异常
logger.Warn("搜索源超时", zap.String("source", site.Name))

// ERROR: 需要处理的错误
logger.Error("数据库连接失败", zap.Error(err))

// FATAL: 系统级错误 (极少使用)
logger.Fatal("无法初始化配置", zap.Error(err))
```

#### 性能优化 (SugaredLogger vs Logger)
```go
// 性能敏感场景使用 Logger (强类型)
logger.Info("search",
    zap.String("q", query),
    zap.Int("n", count),
)

// 便利性场景使用 SugaredLogger (弱类型)
sugar := logger.Sugar()
sugar.Infow("搜索完成",
    "query", query,
    "count", count,
)

// 格式化输出 (仅开发使用)
sugar.Infof("搜索 %s 完成", query)
```

### 10.2 日志切割与归档

```go
// 使用 lumberjack 进行日志切割
import "gopkg.in/natefinch/lumberjack.v2"

func NewFileLogger(path string) (*zap.Logger, error) {
    writer := &lumberjack.Logger{
        Filename:   path,
        MaxSize:    100,  // MB
        MaxBackups: 30,
        MaxAge:     7,    // days
        Compress:   true,
    }
    
    core := zapcore.NewCore(
        zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
        zapcore.AddSync(writer),
        zap.InfoLevel,
    )
    
    return zap.New(core), nil
}
```

