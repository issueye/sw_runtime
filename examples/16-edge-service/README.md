# 16-edge-service - Go + SW Runtime 边缘服务示例

本示例展示如何在 **Go HTTP 服务** 中嵌入 SW Runtime 的 JavaScript/TypeScript 运行时，
用于实现简单的“边缘服务”（Edge Functions）能力：

- Go 负责 HTTP 服务、路由与基础设施
- 每个请求由 JS/TS 脚本决定具体业务逻辑
- 使用 `internal/runtime.RunnerPool` 复用 Runner，减少 VM 创建开销
- 使用 `SafeRunFile` 捕获 goja 层 panic，避免整个服务崩溃

## 目录结构

```text
examples/16-edge-service/
├── main.go           # Go HTTP 服务入口，使用 RunnerPool + SafeRunFile
└── scripts/
    └── hello-edge.ts # 边缘脚本示例
```

## 运行示例

在项目根目录下执行：

```bash
# 运行 Go HTTP 服务
go run ./examples/16-edge-service/main.go
```

启动后，你会看到类似日志：

```text
[edge-service] listening on :8080
try: curl http://localhost:8080/edge/hello?name=SW-Runtime
```

### 调用示例

```bash
# 默认脚本 hello-edge.ts
curl "http://localhost:8080/edge/hello?name=SW-Runtime"
```

返回类似 JSON：

```json
{
  "message": "Hello, SW-Runtime!",
  "method": "GET",
  "path": "/edge/hello",
  "time": "2026-01-04T07:30:00.000Z"
}
```

## 实现要点

### 1. Runner 池

`main.go` 中使用全局 Runner 池：

```go
var runnerPool = rt.NewRunnerPool()
```

在每个 HTTP 请求中：

```go
runner := rt.AcquireRunner()
defer rt.ReleaseRunner(runner)
```

这样可以：

- 避免为每个请求都新建一个 JS VM
- 通过 `ClearModuleCache()` 清理模块缓存，减少跨请求干扰

> 注意：Runner 池会复用同一个 VM 实例，全局状态不会自动重置，
> 所以脚本应避免在 `global` 上保留跨请求的共享可变状态。

### 2. SafeRunFile 捕获 panic

在执行脚本时使用 `SafeRunFile`：

```go
if err := runner.SafeRunFile(scriptPath); err != nil {
    log.Printf("edge script panic (%s): %v", scriptPath, err)
    http.Error(w, "edge script failed", http.StatusInternalServerError)
    return
}
```

这样可以把 goja 内部的 panic（例如极端 Promise/定时器场景）
转换为普通错误，而不会让整个 HTTP 服务进程崩溃。

### 3. 请求上下文传递

Go 侧把当前请求信息打包成 `EdgeRequest` 结构，并注入到 JS 全局：

```go
edgeReq := EdgeRequest{ Method: r.Method, Path: r.URL.Path, Query: query, Body: string(bodyBytes) }
runner.SetValue("request", edgeReq)
runner.SetValue("response", nil) // 清理上一次可能遗留的响应
```

脚本中通过全局 `request` 使用这些信息：

```ts
declare const request: EdgeRequest;

const name = request.query["name"] || "World";
```

### 4. 响应约定

脚本通过设置全局 `response` 来告诉 Go 如何返回结果：

```ts
globalThis.response = {
  status: 200,
  headers: {
    "X-Edge-Service": "sw-runtime",
  },
  json: {
    message: `Hello, ${name}!`,
    method: request.method,
    path: request.path,
    time: new Date().toISOString(),
  },
};
```

Go 侧按约定解析：

- 如果 `response.json` 不为空，则以 `application/json` 返回
- 否则使用 `response.body` 作为纯文本响应
- 未设置 `status` 时默认使用 `200 OK`

## 适用场景

这个示例可以作为以下场景的起点：

- API Gateway / BFF：在边缘用 JS/TS 做轻量处理、聚合与转发
- 灰度逻辑 / 实验逻辑：不改 Go 主服务，通过脚本快速上线试验
- "Edge Functions"：根据路由或租户动态加载不同脚本

你可以在 `scripts/` 目录下添加更多 `*-edge.ts` 文件，
并通过 `/edge/{name}` 的形式进行访问。
