# HTTP 服务器示例

本目录包含 httpserver 服务器模块的功能演示。

## 文件说明

- **httpserver-demo.ts** - HTTP 服务器基础演示
- **httpserver-v3-features.js** - v3.0 新特性演示 (路径参数、流式传输、Request 增强)
- **httpserver-file-demo.js** - 文件服务演示
- **httpserver-timeout-demo.ts** - 超时配置演示

## 功能特点

### 路由系统

- GET, POST, PUT, DELETE 等方法
- **路径参数** (例如 `/:id`, `/:postId`)
- 查询参数

### 中间件

- Express 风格中间件链
- 请求日志
- 自定义处理

### 响应方法

- JSON 响应
- HTML 响应
- 文本响应
- **流式文件发送** (使用 `res.sendFile`, `res.download`)
- 重定向

### Request 对象增强 (v3.0)

- `req.cookies`: 自动解析 Cookie
- `req.get(header)`: 获取请求头
- `req.protocol`, `req.secure`: 协议识别
- `req.hostname`, `req.xhr`: 客户端信息
- `req.form`: 自动解析表单数据

### 静态文件

- 自动 MIME 类型检测
- 静态目录服务
- 缓存控制

### 超时配置

- 自定义读取超时
- 自定义写入超时
- 自定义空闲超时
- 自定义请求头超时
- 自定义最大请求头大小

### WebSocket

- 实时双向通信
- 事件驱动

## 运行示例

```bash
# 运行 v3.0 特性演示 (新)
sw_runtime run examples/06-http-server/httpserver-v3-features.js

# 运行基础服务器
sw_runtime run examples/06-http-server/httpserver-demo.ts

# 运行文件服务器
sw_runtime run examples/06-http-server/httpserver-file-demo.js
```

## 示例代码

### 基本使用

```javascript
const server = require("http/server");
const app = server.createServer();

// 中间件
app.use((req, res, next) => {
  console.log(`${req.method} ${req.path}`);
  next();
});

// 路由 (支持路径参数)
app.get("/user/:id", (req, res) => {
  res.json({ id: req.params.id, name: "John" });
});

app.get("/", (req, res) => {
  res.html("<h1>Hello!</h1>");
});

// 静态文件
app.static("./public", "/");

// 启动服务器
app.listen("3000");
```

### 超时配置

```javascript
const server = require("http/server");

// 创建带有自定义超时配置的服务器
const app = server.createServer({
  readTimeout: 15, // 读取超时：15秒（默认10秒）
  writeTimeout: 15, // 写入超时：15秒（默认10秒）
  idleTimeout: 60, // 空闲超时：60秒（默认120秒）
  readHeaderTimeout: 5, // 读取请求头超时：5秒（默认10秒）
  maxHeaderBytes: 16384, // 最大请求头大小：16KB（默认8KB）
});
```

app.get('/', (req, res) => {
res.json({ message: 'Server with custom timeouts' });
});

app.listen('3100');

```

```
