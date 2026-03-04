# Channel 并发模式实战示例

## 场景: 多视频源并发搜索优化

### 问题背景
- 原方案: 串行请求 10 个视频源，总耗时 = 10 × 平均耗时 (500ms) = 5s
- 目标: 并发请求，限制最大并发数，总耗时 ≈ 平均耗时 (500ms)

### 方案对比

#### 方案1: 串行 (慢)
```go
func SearchSerial(query string, sites []ApiSite) []Result {
    var results []Result
    for _, site := range sites {
        res := search(site, query) // 500ms
        results = append(results, res...)
    }
    return results
}
// 10个源 = 5秒
```

#### 方案2: 无限制并发 (资源耗尽)
```go
func SearchUnlimited(query string, sites []ApiSite) []Result {
    var results []Result
    var mu sync.Mutex
    
    for _, site := range sites {
        go func(s ApiSite) {
            res := search(s, query)
            mu.Lock()
            results = append(results, res...)
            mu.Unlock()
        }(site)
    }
    // 问题: 同时创建 10 个 goroutine，可能耗尽连接池
    return results
}
```

#### 方案3: Channel Semaphore (推荐 ✅)
```go
func SearchWithChannel(query string, sites []ApiSite) []Result {
    const MaxConcurrent = 5
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // 1. Channel 作为信号量
    semaphore := make(chan struct{}, MaxConcurrent)
    
    // 2. Channel 收集结果
    resultChan := make(chan []Result, len(sites))
    
    // 3. 原子计数器
    var completed int32
    
    for _, site := range sites {
        site := site
        
        select {
        case semaphore <- struct{}{}: // 获取信号量
            go func() {
                defer func() { <-semaphore }() // 释放
                
                res, err := search(ctx, site, query)
                if err != nil {
                    log.Printf("源 %s 失败: %v", site.Name, err)
                }
                
                atomic.AddInt32(&completed, 1)
                resultChan <- res
            }()
            
        case <-ctx.Done():
            break
        }
    }
    
    // 4. 收集结果 (不需要锁)
    var allResults []Result
    for i := 0; i < len(sites); i++ {
        select {
        case res := <-resultChan:
            allResults = append(allResults, res...)
        case <-ctx.Done():
            log.Printf("搜索超时，已完成 %d/%d", atomic.LoadInt32(&completed), len(sites))
            goto DONE
        }
    }

DONE:
    return allResults
}
// 10个源，并发5个 = 约 1秒
```

### 性能对比

| 方案 | 10个源耗时 | 内存占用 | 连接数 | 推荐 |
|------|-----------|---------|--------|------|
| 串行 | 5000ms | 低 | 1 | ❌ |
| 无限制并发 | 500ms | 高 | 10 | ❌ |
| Channel Semaphore | 1000ms | 中 | 5 | ✅ |

### 关键优化点

1. **信号量控制**: 防止同时请求过多，耗尽连接池
2. **Channel 收集**: 避免锁竞争，提高性能
3. **原子计数**: 无锁计数，减少开销
4. **超时控制**: 防止慢源拖累整体响应

### 实际应用

```go
// internal/service/search_service_v2.go
// 基于 Channel 的高性能并发搜索

type searchServiceV2 struct {
    client    *http.Client
    logger    *zap.Logger
    semaphore chan struct{} // 信号量 channel
}

func (s *searchServiceV2) Search(ctx context.Context, query string, sites []ApiSite) ([]Result, error) {
    // 复用上面的 SearchWithChannel 模式
    // ...
}
```

### 注意事项

1. **Channel 缓冲大小**: 根据任务数量设置，避免阻塞
2. **及时释放信号量**: `defer` 确保一定释放
3. **处理 Panic**: 使用 `recover` 防止单个任务崩溃影响整体
4. **Context 传递**: 支持取消和超时

```go
go func() {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("任务 panic: %v", r)
        }
        <-semaphore // 确保释放
    }()
    
    process(task)
}()
```
