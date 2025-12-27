# SW Runtime 优化计划

本文档详细记录了代码审查中发现的问题和对应的优化计划，按优先级分级。

## 问题总览

| 类别 | Critical | High | Medium | Low | 总计 |
|------|----------|------|--------|-----|------|
| 安全问题 | 3 | 2 | 0 | 0 | 5 |
| 并发问题 | 2 | 2 | 0 | 0 | 4 |
| 性能问题 | 0 | 3 | 2 | 0 | 5 |
| 错误处理 | 0 | 3 | 1 | 0 | 4 |
| 代码质量 | 0 | 0 | 2 | 2 | 4 |
| 测试覆盖 | 0 | 6 | 4 | 0 | 10 |
| **总计** | **5** | **16** | **9** | **2** | **32** |

---

## 第一阶段：紧急安全修复 (Critical - 必须立即修复)

### 1.1 WebSocket CORS 漏洞 🔴 Critical
**文件**: `internal/builtins/httpserver.go:105-108`

**问题**: WebSocket 允许所有来源连接，存在严重安全风险
```go
CheckOrigin: func(r *http.Request) bool {
    return true // 允许所有来源,生产环境应该限制
}
```

**修复方案**:
```go
type HTTPServer struct {
    // ... 现有字段
    allowedOrigins []string
    wsAllowAll     bool // 默认 false
}

func (s *HTTPServer) configureWebSocket(options map[string]interface{}) {
    s.upgrader = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            if s.wsAllowAll {
                return true
            }
            origin := r.Header.Get("Origin")
            for _, allowed := range s.allowedOrigins {
                if origin == allowed || allowed == "*" {
                    return true
                }
            }
            return false
        },
    }
}
```

### 1.2 路径遍历漏洞 🔴 Critical
**文件**: `internal/builtins/fs.go`

**问题**: 文件操作没有路径验证，可能导致读取任意系统文件
```go
filename := call.Arguments[0].String()
content, err := os.ReadFile(filename)  // 可能读取 ../../../etc/passwd
```

**修复方案**: 添加路径沙箱
```go
// 添加到 FSModule
type FSModule struct {
    vm       *goja.Runtime
    basePath string // 工作目录，作为沙箱根目录
    mu       sync.RWMutex
}

func (m *FSModule) sanitizePath(path string) (string, error) {
    // 转换为绝对路径
    absPath, err := filepath.Abs(path)
    if err != nil {
        return "", fmt.Errorf("invalid path: %w", err)
    }

    // 检查是否在允许的基础路径内
    relPath, err := filepath.Rel(m.basePath, absPath)
    if err != nil || strings.HasPrefix(relPath, "..") {
        return "", fmt.Errorf("access denied: path outside sandbox")
    }

    return absPath, nil
}

// 在所有文件操作中使用
func (m *FSModule) readFile(call goja.FunctionCall) goja.Value {
    filename := call.Arguments[0].String()
    safePath, err := m.sanitizePath(filename)
    if err != nil {
        panic(m.vm.NewGoError(err))
    }
    content, err := os.ReadFile(safePath)
    // ...
}
```

### 1.3 SQL 注入风险 🔴 Critical
**文件**: `internal/builtins/sqlite.go`

**问题**: 多处使用字符串拼接构造 SQL 语句
```go
sql := "SELECT * FROM " + tableName + " WHERE id = " + id
```

**修复方案**:
```go
// 强制使用预处理语句
func (db *SQLiteDB) querySafe(sql string, params ...interface{}) (*goja.Object, error) {
    // 验证 SQL 不包含多语句（防止 SQL 注入）
    if strings.Contains(sql, ";") && !strings.HasPrefix(strings.TrimSpace(sql), "BEGIN") {
        return nil, fmt.Errorf("multi-statement SQL not allowed")
    }

    stmt, err := db.db.Prepare(sql)
    if err != nil {
        return nil, err
    }
    defer stmt.Close()

    // 执行参数化查询
    rows, err := stmt.Query(params...)
    // ...
}

// 提供 API 级别的查询方法（自动转义）
func (db *SQLiteDB) Select(table string, where string, args []interface{}) (*goja.Object, error) {
    sql := fmt.Sprintf("SELECT * FROM %s WHERE %s", table, where)
    return db.querySafe(sql, args...)
}
```

### 1.4 goja.Runtime 并发访问 🔴 Critical
**文件**: `internal/builtins/httpserver.go`

**问题**: goja.Runtime 不是线程安全的，但在 HTTP 处理器中被并发访问

**修复方案**: 实现请求队列
```go
type HTTPServer struct {
    vm          *goja.Runtime
    vmMutex     sync.Mutex
    requestChan chan func(*goja.Runtime) // 请求处理队列
    wg          sync.WaitGroup
}

func (s *HTTPServer) startVMProcessor() {
    s.wg.Add(1)
    go func() {
        defer s.wg.Done()
        for fn := range s.requestChan {
            s.vmMutex.Lock()
            fn(s.vm)
            s.vmMutex.Unlock()
        }
    }()
}

// 在处理器中使用
func (s *HTTPServer) handleRequest(w http.ResponseWriter, r *http.Request) {
    resultChan := make(chan goja.Value)
    s.requestChan <- func(vm *goja.Runtime) {
        result := s.executeHandler(vm, w, r)
        resultChan <- result
    }
    result := <-resultChan
    // ...
}
```

### 1.5 SSRF 风险 (HTTP 请求) 🔴 Critical
**文件**: `internal/builtins/http.go`

**问题**: 没有验证目标 URL，可能被利用进行内网探测

**修复方案**:
```go
type HTTPClient struct {
    client      *http.Client
    vm          *goja.Runtime
    blockedNets []*net.IPNet // 禁止访问的网段
}

func (c *HTTPClient) validateURL(urlStr string) error {
    u, err := url.Parse(urlStr)
    if err != nil {
        return err
    }

    // 只允许 http/https
    if u.Scheme != "http" && u.Scheme != "https" {
        return fmt.Errorf("unsupported scheme: %s", u.Scheme)
    }

    // 检查是否是内网地址
    host := u.Hostname()
    ip := net.ParseIP(host)
    if ip != nil {
        for _, blocked := range c.blockedNets {
            if blocked.Contains(ip) {
                return fmt.Errorf("access to private network denied")
            }
        }
    }

    return nil
}

// 配置默认阻止的内网网段
func NewHTTPModule(vm *goja.Runtime) *HTTPClient {
    blockedNets := []*net.IPNet{
        parseCIDR("127.0.0.0/8"),
        parseCIDR("10.0.0.0/8"),
        parseCIDR("172.16.0.0/12"),
        parseCIDR("192.168.0.0/16"),
    }
    // ...
}
```

---

## 第二阶段：高优先级修复 (High Priority)

### 2.1 错误处理改进 ⚠️ High

**问题列表**:
1. `runner.go:29` - 忽略 `os.Getwd()` 错误
2. `sqlite.go` - 使用 `panic` 而非返回错误
3. 所有 builtin 模块 - 一致的错误处理模式

**修复方案**:
```go
// 创建统一的错误处理工具包
// internal/errors/errors.go
package errors

import (
    "fmt"
    "runtime"
)

type RuntimeError struct {
    Code    string
    Message string
    File    string
    Line    int
}

func (e *RuntimeError) Error() string {
    return fmt.Sprintf("[%s] %s (%s:%d)", e.Code, e.Message, e.File, e.Line)
}

func Wrap(code, message string) error {
    _, file, line, _ := runtime.Caller(1)
    return &RuntimeError{
        Code:    code,
        Message: message,
        File:    file,
        Line:    line,
    }
}

// 在 builtin 模块中使用
func (m *FSModule) readFile(call goja.FunctionCall) goja.Value {
    filename := call.Arguments[0].String()
    safePath, err := m.sanitizePath(filename)
    if err != nil {
        // 返回 JavaScript Error 对象而非 panic
        errObj := m.vm.NewObject()
        errObj.Set("code", "FS_ACCESS_DENIED")
        errObj.Set("message", err.Error())
        return m.vm.ToValue(errObj)
    }
    // ...
}
```

### 2.2 Goroutine 泄漏修复 ⚠️ High

**文件**: `internal/builtins/httpserver.go:303-332`

**问题**: Interval timer goroutine 没有清理机制

**修复方案**:
```go
type HTTPServer struct {
    // ... 现有字段
    intervals      map[int]*intervalEntry
    intervalsMutex sync.RWMutex
    intervalWg     sync.WaitGroup
    stopChan       chan struct{}
}

func (s *HTTPServer) cleanupIntervals() {
    s.intervalsMutex.Lock()
    defer s.intervalsMutex.Unlock()

    for id, interval := range s.intervals {
        interval.stop()
        delete(s.intervals, id)
    }
}

func (s *HTTPServer) Close() error {
    // 停止接受新连接
    if s.server != nil {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        s.server.Shutdown(ctx)
    }

    // 清理所有 interval
    s.cleanupIntervals()

    // 等待所有 goroutine 完成
    done := make(chan struct{})
    go func() {
        s.intervalWg.Wait()
        close(done)
    }()

    select {
    case <-done:
        return nil
    case <-time.After(10 * time.Second):
        return fmt.Errorf("timeout waiting for goroutines to finish")
    }
}
```

### 2.3 性能优化：缓冲池 ⚠️ High

**文件**: `internal/builtins/fs.go`, `http.go`

**问题**: 频繁的内存分配

**修复方案**:
```go
// internal/pool/buffer.go
package pool

import "bytes"

var (
    // 小缓冲池 (用于读取文本文件)
    SmallBufferPool = sync.Pool{
        New: func() interface{} {
            return make([]byte, 4*1024) // 4KB
        },
    }

    // 大缓冲池 (用于读取二进制文件)
    LargeBufferPool = sync.Pool{
        New: func() interface{} {
            return make([]byte, 64*1024) // 64KB
        },
    }

    // Buffer 池 (用于字符串拼接)
    ByteBufferPool = sync.Pool{
        New: func() interface{} {
            return new(bytes.Buffer)
        },
    }
}

// 使用示例
func (m *FSModule) readFile(call goja.FunctionCall) goja.Value {
    filename := call.Arguments[0].String()

    // 从池中获取缓冲区
    buf := pool.SmallBufferPool.Get().([]byte)
    defer pool.SmallBufferPool.Put(buf)

    f, err := os.Open(filename)
    if err != nil {
        panic(m.vm.NewGoError(err))
    }
    defer f.Close()

    n, err := f.Read(buf)
    if err != nil && err != io.EOF {
        panic(m.vm.NewGoError(err))
    }

    result := make([]byte, n)
    copy(result, buf[:n])
    return m.vm.ToValue(string(result))
}
```

### 2.4 性能优化：定时器改进 ⚠️ High

**文件**: `internal/runtime/eventloop_simple.go`

**问题**: 每个定时器创建一个 goroutine，效率低

**修复方案**: 使用时间轮算法
```go
// internal/runtime/timingwheel.go
package runtime

import (
    "container/heap"
    "time"
)

type timerEntry struct {
    id       int
    deadline time.Time
    callback func()
    index    int
}

type timerHeap []*timerEntry

func (h timerHeap) Len() int           { return len(h) }
func (h timerHeap) Less(i, j int) bool { return h[i].deadline.Before(h[j].deadline) }
func (h timerHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i]; h[i].index = i; h[j].index = j }

func (h *timerHeap) Push(x interface{}) {
    item := x.(*timerEntry)
    item.index = len(*h)
    *h = append(*h, item)
}

func (h *timerHeap) Pop() interface{} {
    old := *h
    n := len(old)
    item := old[n-1]
    *h = old[0 : n-1]
    return item
}

type TimingWheelEventLoop struct {
    vm       *goja.Runtime
    timers   timerHeap
    mu       sync.Mutex
    ticker   *time.Ticker
    stopChan chan struct{}
}

func NewTimingWheelEventLoop(vm *goja.Runtime) *TimingWheelEventLoop {
    el := &TimingWheelEventLoop{
        vm:       vm,
        timers:   make(timerHeap, 0),
        stopChan: make(chan struct{}),
    }
    heap.Init(&el.timers)

    el.ticker = time.NewTicker(10 * time.Millisecond) // 10ms 精度
    go el.run()

    return el
}

func (el *TimingWheelEventLoop) run() {
    for {
        select {
        case <-el.ticker.C:
            el.processTimers()
        case <-el.stopChan:
            el.ticker.Stop()
            return
        }
    }
}

func (el *TimingWheelEventLoop) processTimers() {
    el.mu.Lock()
    defer el.mu.Unlock()

    now := time.Now()
    for el.timers.Len() > 0 {
        entry := el.timers[0]
        if entry.deadline.After(now) {
            break
        }

        heap.Pop(&el.timers)
        go entry.callback()
    }
}
```

### 2.5 模块缓存限制 ⚠️ High

**文件**: `internal/modules/system.go`

**问题**: 模块缓存无限增长

**修复方案**:
```go
type System struct {
    vm             *goja.Runtime
    cache          *lru.Cache  // 使用 LRU 缓存
    builtinManager *builtins.Manager
    mu             sync.RWMutex
    basePath       string
    nodeModules    []string
}

func NewSystem(vm *goja.Runtime, basePath string) *System {
    ms := &System{
        vm:             vm,
        cache:          lru.New(1000), // 最多缓存 1000 个模块
        builtinManager: builtins.NewManager(vm),
        basePath:       basePath,
        nodeModules: []string{
            filepath.Join(basePath, "node_modules"),
        },
    }
    return ms
}

// 需要添加依赖: github.com/hashicorp/golang-lru
```

---

## 第三阶段：代码质量提升 (Medium Priority)

### 3.1 提取魔法数字为常量

```go
// internal/consts/constants.go
package consts

const (
    // 文件系统权限
    FilePermReadWrite = 0644
    DirPermReadWrite  = 0755

    // 网络相关
    DefaultHTTPTimeout  = 30 * time.Second
    DefaultReadTimeout  = 10 * time.Second
    DefaultWriteTimeout = 10 * time.Second
    DefaultIdleTimeout  = 120 * time.Second

    // 缓存大小
    DefaultTimerCacheSize    = 64
    DefaultIntervalCacheSize = 32
    DefaultModuleCacheSize   = 1000

    // WebSocket
    WSReadBufferSize  = 1024
    WSWriteBufferSize = 1024
    WSMaxMessageSize  = 10 * 1024 * 1024 // 10MB

    // 缓冲区大小
    SmallBufferSize  = 4 * 1024   // 4KB
    MediumBufferSize = 64 * 1024  // 64KB
    LargeBufferSize  = 1024 * 1024 // 1MB
)
```

### 3.2 统一错误响应格式

```go
// internal/errors/response.go
package errors

import "github.com/dop251/goja"

// JSError 创建标准化的 JavaScript 错误对象
func JSError(vm *goja.Runtime, code, message string) goja.Value {
    errObj := vm.NewObject()
    errObj.Set("code", code)
    errObj.Set("message", message)
    errObj.Set("name", "RuntimeError")

    // 添加堆栈跟踪
    errObj.Set("stack", getStackTrace(2))

    return vm.ToValue(errObj)
}

// 错误代码常量
const (
    ErrCodeFSAccessDenied    = "FS_ACCESS_DENIED"
    ErrCodeFSNotFound        = "FS_NOT_FOUND"
    ErrCodeFSPermission      = "FS_PERMISSION_DENIED"
    ErrCodeDBQueryFailed     = "DB_QUERY_FAILED"
    ErrCodeDBConnection      = "DB_CONNECTION_FAILED"
    ErrCodeHTTPInvalidURL    = "HTTP_INVALID_URL"
    ErrCodeHTTPSSRF          = "HTTP_SSRF_BLOCKED"
    ErrCodeModuleNotFound    = "MODULE_NOT_FOUND"
    ErrCodeModuleLoadError   = "MODULE_LOAD_ERROR"
    ErrCodeValidationFailed  = "VALIDATION_FAILED"
)
```

### 3.3 消除代码重复

创建公共辅助函数:
```go
// internal/builtins/utils.go
package builtins

import (
    "github.com/dop251/goja"
)

// GetStringArg 安全获取字符串参数
func GetStringArg(call goja.FunctionCall, index int, defaultValue string) string {
    if len(call.Arguments) <= index {
        return defaultValue
    }
    return call.Arguments[index].String()
}

// GetIntArg 安全获取整数参数
func GetIntArg(call goja.FunctionCall, index int, defaultValue int64) int64 {
    if len(call.Arguments) <= index {
        return defaultValue
    }
    return call.Arguments[index].ToInteger()
}

// GetObjectArg 安全获取对象参数
func GetObjectArg(call goja.FunctionCall, index int) map[string]interface{} {
    if len(call.Arguments) <= index {
        return nil
    }
    return call.Arguments[index].Export().(map[string]interface{})
}

// ThrowError 抛出标准化错误
func ThrowError(vm *goja.Runtime, code, message string) {
    panic(JSError(vm, code, message))
}
```

---

## 第四阶段：测试完善 (Testing)

### 4.1 安全测试

```go
// test/security_test.go
package test

import (
    "testing"
)

func TestPathTraversal(t *testing.T) {
    tests := []struct {
        path    string
        allowed bool
    }{
        {"./test.txt", true},
        {"../secret.txt", false},
        {"/etc/passwd", false},
        {"./subdir/file.txt", true},
        {"../../../../../etc/passwd", false},
    }

    for _, tt := range tests {
        // 测试路径遍历防护
    }
}

func TestSQLInjection(t *testing.T) {
    tests := []string{
        "'; DROP TABLE users; --",
        "1' OR '1'='1",
        "'; INSERT INTO users VALUES ('hacker', 'password'); --",
    }

    for _, payload := range tests {
        // 测试 SQL 注入防护
    }
}

func TestSSRF(t *testing.T) {
    tests := []struct {
        url     string
        allowed bool
    }{
        {"https://api.example.com/data", true},
        {"http://127.0.0.1:6379", false},
        {"http://localhost:8080", false},
        {"http://192.168.1.1/admin", false},
        {"https://10.0.0.1/secret", false},
    }

    for _, tt := range tests {
        // 测试 SSRF 防护
    }
}
```

### 4.2 并发测试

```go
// test/concurrency_test.go
package test

import (
    "sync"
    "testing"
)

func TestConcurrentModuleLoading(t *testing.T) {
    const goroutines = 100
    var wg sync.WaitGroup

    for i := 0; i < goroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            // 并发加载相同模块
            runner := runtime.New()
            runner.RunFile("test-module.js")
        }()
    }

    wg.Wait()
}

func TestConcurrentHTTPRequest(t *testing.T) {
    const requests = 1000
    var wg sync.WaitGroup

    for i := 0; i < requests; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            // 发送并发 HTTP 请求
            http.Get("http://localhost:8080/api/data")
        }()
    }

    wg.Wait()
}
```

### 4.3 性能基准测试

```go
// test/benchmark_test.go
package test

import (
    "testing"
)

func BenchmarkModuleLoading(b *testing.B) {
    runner := runtime.New()
    for i := 0; i < b.N; i++ {
        runner.ClearModuleCache()
        runner.RunFile("test-module.js")
    }
}

func BenchmarkTimerOperations(b *testing.B) {
    runner := runtime.New()
    runner.RunCode(`
        let count = 0;
        function callback() { count++; }
    `)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        runner.RunCode("setTimeout(callback, 1);")
    }
}

func BenchmarkHTTPServer(b *testing.B) {
    // 启动服务器
    go startTestServer()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        http.Get("http://localhost:8080/api/test")
    }
}
```

---

## 优化时间表

| 阶段 | 工作内容 | 预计工作量 | 优先级 |
|------|----------|------------|--------|
| **第1阶段** | 安全漏洞修复 | 3-5天 | Critical |
| 第2阶段 | 高优先级问题修复 | 5-7天 | High |
| 第3阶段 | 代码质量提升 | 3-5天 | Medium |
| 第4阶段 | 测试完善 | 5-7天 | High |

---

## 建议的实施顺序

1. **立即修复** (本周内)
   - WebSocket CORS 问题
   - 路径遍历漏洞
   - SQL 注入风险
   - SSRF 风险

2. **下周完成**
   - goja.Runtime 并发访问修复
   - 错误处理改进
   - Goroutine 泄漏修复

3. **两周内完成**
   - 性能优化（缓冲池、定时器改进）
   - 模块缓存限制

4. **持续进行**
   - 添加安全测试
   - 添加并发测试
   - 代码重构和质量提升

---

## 验证清单

完成每个优化后，使用以下清单验证：

- [ ] 代码能正常编译
- [ ] 所有现有测试通过
- [ ] 新增测试覆盖新代码
- [ ] 性能基准测试无退化
- [ ] 文档已更新
- [ ] 安全测试通过
- [ ] 内存泄漏测试通过
- [ ] 并发安全测试通过
