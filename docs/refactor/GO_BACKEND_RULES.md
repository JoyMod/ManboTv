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

## 4. 并发安全规范

### Mutex 使用
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
