# HTTP 服务器示例

本目录包含 httpserver 服务器模块的功能演示。

## 文件说明

- **httpserver-demo.ts** - HTTP 服务器基础演示
- **httpserver-file-demo.js** - 文件服务演示

## 功能特点

### 路由系统
- GET, POST, PUT, DELETE 等方法
- 路径参数
- 查询参数

### 中间件
- Express 风格中间件链
- 请求日志
- 自定义处理

### 响应方法
- JSON 响应
- HTML 响应
- 文本响应
- 文件发送
- 文件下载
- 重定向

### 静态文件
- 自动 MIME 类型检测
- 静态目录服务
- 缓存控制

### WebSocket
- 实时双向通信
- 事件驱动

## 运行示例

```bash
# 运行基础服务器
sw_runtime run examples/06-http-server/httpserver-demo.ts

# 运行文件服务器
sw_runtime run examples/06-http-server/httpserver-file-demo.js
```

## 示例代码

```javascript
const server = require('httpserver');
const app = server.createServer();

// 中间件
app.use((req, res, next) => {
  console.log(`${req.method} ${req.path}`);
  next();
});

// 路由
app.get('/', (req, res) => {
  res.html('<h1>Hello!</h1>');
});

app.get('/api/users', (req, res) => {
  res.json({ users: [] });
});

// 静态文件
app.static('./public', '/');

// 启动服务器
app.listen('3000');
```
