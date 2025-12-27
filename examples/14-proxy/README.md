# 代理服务器示例

本目录包含 HTTP 和 TCP 代理服务器的功能演示。

## 文件说明

- **http-proxy-demo.js** - HTTP 代理基础演示
- **http-proxy-advanced.js** - HTTP 代理高级示例（统计和监控）
- **tcp-proxy-demo.js** - TCP 代理基础演示
- **tcp-proxy-advanced.js** - TCP 代理高级示例（统计和监控）

## 功能特点

### HTTP 代理
- 反向代理 HTTP/HTTPS 请求
- 请求和响应拦截
- 自定义请求头修改
- 错误处理
- HTTPS 支持（自动处理 SSL）

### TCP 代理
- 透明 TCP 连接转发
- 双向数据传输
- 连接监控
- 数据统计
- 错误处理

## 运行示例

### HTTP 代理服务器

#### 基础示例
```bash
sw_runtime run examples/14-proxy/http-proxy-demo.js
```

#### 高级示例（带统计）
```bash
sw_runtime run examples/14-proxy/http-proxy-advanced.js
```

#### 测试 HTTP 代理
```bash
# 使用 curl 测试
curl -x http://localhost:8888 https://httpbin.org/get
curl -x http://localhost:8888 https://httpbin.org/post -d "key=value"

# 或在浏览器中设置代理
# 代理服务器: localhost
# 端口: 8888
```

### TCP 代理服务器

#### 基础示例
```bash
sw_runtime run examples/14-proxy/tcp-proxy-demo.js
```

#### 高级示例（带统计）
```bash
# 确保 Redis 服务器在 localhost:6379 运行
redis-server

# 启动 TCP 代理
sw_runtime run examples/14-proxy/tcp-proxy-advanced.js
```

#### 测试 TCP 代理
```bash
# 使用 redis-cli 测试（连接到代理端口）
redis-cli -p 6380

# 或使用 netcat
nc localhost 5353
```

## 示例代码

### HTTP 代理
```javascript
const proxy = require('proxy');

// 创建 HTTP 代理
const httpProxy = proxy.createHTTPProxy('https://api.example.com');

// 监听请求
httpProxy.on('request', (req) => {
  console.log(`${req.method} ${req.path}`);
});

// 监听响应
httpProxy.on('response', (resp) => {
  console.log(`Status: ${resp.status}`);
});

// 监听错误
httpProxy.on('error', (err) => {
  console.error('Proxy error:', err.message);
});

// 启动代理
httpProxy.listen('8080').then(() => {
  console.log('HTTP Proxy running on port 8080');
});
```

### TCP 代理
```javascript
const proxy = require('proxy');

// 创建 TCP 代理
const tcpProxy = proxy.createTCPProxy('localhost:6379');

// 监听连接
tcpProxy.on('connection', (conn) => {
  console.log('New connection:', conn.remoteAddr);
});

// 监听数据传输
tcpProxy.on('data', (data) => {
  console.log(`${data.direction}: ${data.bytes} bytes`);
});

// 监听关闭
tcpProxy.on('close', () => {
  console.log('Connection closed');
});

// 启动代理
tcpProxy.listen('6380').then(() => {
  console.log('TCP Proxy running on port 6380');
});
```

## 事件说明

### HTTP 代理事件

- **request** - 接收到客户端请求
  - `req.method` - 请求方法
  - `req.url` - 请求 URL
  - `req.path` - 请求路径
  - `req.host` - 主机名
  - `req.headers` - 请求头

- **response** - 收到目标服务器响应
  - `resp.status` - 状态码
  - `resp.statusText` - 状态文本
  - `resp.headers` - 响应头

- **error** - 代理错误
  - `err.message` - 错误消息
  - `err.url` - 请求 URL

### TCP 代理事件

- **connection** - 新连接建立
  - `conn.remoteAddr` - 客户端地址
  - `conn.target` - 目标服务器地址

- **data** - 数据传输
  - `data.direction` - 传输方向（"client->target" 或 "target->client"）
  - `data.bytes` - 传输字节数

- **close** - 连接关闭

- **error** - 代理错误
  - `err.message` - 错误消息
  - `err.direction` - 错误发生的方向

## 使用场景

### HTTP 代理
- 开发环境 API 转发
- 跨域问题解决
- 请求/响应日志记录
- API 监控和调试
- 负载均衡前端

### TCP 代理
- 端口转发
- 数据库连接代理
- 服务迁移过渡
- 网络流量监控
- 协议分析和调试

## 注意事项

1. **安全性**: 代理服务器应在受信任的网络环境中使用
2. **性能**: 代理会增加额外的延迟
3. **SSL/TLS**: HTTP 代理会自动处理 HTTPS，但会跳过证书验证
4. **连接数**: 注意监控活跃连接数，避免资源耗尽
5. **错误处理**: 始终监听 error 事件并妥善处理

## 高级功能

### 请求修改
HTTP 代理会自动转发所有请求头，你可以在 request 事件中记录或分析它们。

### 响应修改
HTTP 代理支持在 response 事件中检查响应头和状态码。

### 连接池
两种代理都使用连接池来优化性能。

### 超时设置
代理内置了合理的超时设置，确保不会无限期等待。

## 故障排除

### HTTP 代理无法连接
- 检查目标 URL 是否正确
- 确保目标服务器可访问
- 检查防火墙设置

### TCP 代理连接失败
- 确保目标服务器正在运行
- 检查目标地址和端口
- 验证网络连接

### 数据传输中断
- 检查 error 事件日志
- 验证网络稳定性
- 增加超时时间
