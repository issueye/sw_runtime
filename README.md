# SW Runtime - 企业级 JavaScript/TypeScript 运行时

一个基于 Go 和 goja 的高性能 JavaScript/TypeScript 运行时，支持模块导入、加解密、压缩、文件系统操作等企业级功能。

## 🏗️ 架构设计

### 科学的包结构

```
sw_runtime/
├── main.go                    # 主程序入口
├── go.mod                     # Go 模块定义
├── internal/                  # 内部包
│   ├── runtime/              # 运行时核心
│   │   ├── runner.go         # 主运行器
│   │   ├── eventloop.go      # 事件循环
│   │   └── transpiler.go     # TypeScript 编译器
│   ├── modules/              # 模块系统
│   │   ├── system.go         # 模块系统核心
│   │   └── transpiler.go     # 模块编译器
│   └── builtins/             # 内置模块
│       ├── manager.go        # 模块管理器
│       ├── path.go           # 路径操作
│       ├── fs.go             # 文件系统
│       ├── crypto.go         # 加密功能
│       └── compression.go    # 压缩功能
├── examples/                  # 示例文件
│   ├── 01-basic/            # 基础示例（TypeScript、ES6、模块）
│   ├── 02-crypto/           # 加密功能演示
│   ├── 03-compression/      # 压缩功能演示
│   ├── 04-fs/               # 文件系统演示
│   ├── 05-http-client/      # HTTP 客户端演示
│   ├── 06-http-server/      # HTTP 服务器演示
│   ├── 07-https/            # HTTPS 服务器演示
│   ├── 08-websocket/        # WebSocket 示例
│   ├── 09-tcp/              # TCP 网络示例
│   ├── 10-udp/              # UDP 网络示例
│   ├── 11-redis/            # Redis 客户端演示
│   ├── 12-sqlite/           # SQLite 数据库演示
│   └── 13-exec/             # 进程执行演示
└── [测试文件...]
```

## ✨ 功能特性

### 🔧 核心功能

1. **模块系统**
   - CommonJS 风格的 `require()` 函数
   - ES6 动态 `import()` 函数
   - 支持相对路径、绝对路径导入
   - 模块缓存机制
   - 内置模块管理

2. **文件类型支持**
   - JavaScript (`.js`) 文件
   - TypeScript (`.ts`) 文件 - 自动编译，支持 ES6 import/export
   - JSON (`.json`) 文件 - 直接解析

3. **异步支持**
   - 事件循环
   - `setTimeout` / `clearTimeout`
   - `setInterval` / `clearInterval`
   - Promise 支持
   - 异步模块加载

### 🔐 加密模块 (`crypto`)

- **哈希函数**: MD5, SHA1, SHA256, SHA512
- **编解码**: Base64, Hex
- **对称加密**: AES-256-GCM
- **随机数生成**: 安全随机字节

```javascript
const crypto = require('crypto');

// 哈希
console.log(crypto.sha256('hello')); // 哈希值

// Base64 编解码
const encoded = crypto.base64Encode('hello');
const decoded = crypto.base64Decode(encoded);

// AES 加解密
const encrypted = crypto.aesEncrypt('secret', 'key');
const decrypted = crypto.aesDecrypt(encrypted, 'key');

// 随机数
const random = crypto.randomBytes(16);
```

### 🗜️ 压缩模块 (`compression` / `zlib`)

- **Gzip 压缩/解压**
- **Zlib 压缩/解压**
- **高性能压缩算法**

```javascript
const compression = require('compression');

// Gzip 压缩
const compressed = compression.gzipCompress(data);
const decompressed = compression.gzipDecompress(compressed);

// Zlib 压缩
const zlibCompressed = compression.zlibCompress(data);
const zlibDecompressed = compression.zlibDecompress(zlibCompressed);
```

### 🌐 HTTP 客户端模块 (`http`)

- **HTTP 方法**: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS
- **请求配置**: 请求头、参数、超时、认证
- **响应处理**: 自动 JSON 解析、状态码、响应头
- **Promise 支持**: 所有请求返回 Promise

```javascript
const http = require('http');

// GET 请求
http.get('https://api.example.com/users')
  .then(response => {
    console.log('状态码:', response.status);
    console.log('数据:', response.data);
  });

// POST 请求
http.post('https://api.example.com/users', {
  data: { name: 'John', email: 'john@example.com' },
  headers: { 'Content-Type': 'application/json' }
})
  .then(response => console.log('创建成功:', response.data));

// 自定义客户端
const client = http.createClient({ timeout: 10 });
client.get('https://api.example.com/data')
  .then(response => console.log(response.data));
```

### 🚀 HTTP/HTTPS 服务器模块 (`httpserver` / `server`)

- **路由系统**: 支持 GET, POST, PUT, DELETE 等 HTTP 方法
- **中间件支持**: Express 风格的中间件链
- **请求处理**: 自动解析请求体、查询参数、请求头
- **响应方法**: JSON、HTML、文本、重定向等响应类型
- **文件服务**: sendFile、download 方法,自动 MIME 类型检测
- **静态文件**: 内置静态文件服务器
- **WebSocket**: 实时双向通信支持
- **HTTPS 支持**: 内置 SSL/TLS 支持，安全加密通信
- **Promise 支持**: 异步启动和关闭

```javascript
const server = require('httpserver');

// 创建服务器
const app = server.createServer();

// 添加中间件
app.use((req, res, next) => {
  console.log(`${req.method} ${req.path}`);
  res.header('X-Powered-By', 'SW-Runtime');
  next();
});

// 添加路由
app.get('/', (req, res) => {
  res.html('<h1>Hello SW Runtime!</h1>');
});

app.get('/api/users', (req, res) => {
  res.json({
    users: [
      { id: 1, name: 'Alice' },
      { id: 2, name: 'Bob' }
    ]
  });
});

app.post('/api/users', (req, res) => {
  const user = req.json; // 自动解析的 JSON 数据
  res.status(201).json({
    message: 'User created',
    user: user
  });
});

// 文件服务
app.get('/file', (req, res) => {
  res.sendFile('./path/to/file.html'); // 自动检测 MIME 类型
});

app.get('/download', (req, res) => {
  res.download('./file.pdf', 'custom-name.pdf'); // 下载文件
});

// 静态文件服务
app.static('./public', '/static');

// WebSocket 服务器支持
app.ws('/chat', (ws) => {
  ws.on('message', (data) => {
    console.log('收到消息:', data);
    ws.send('回复: ' + data);
  });
  
  ws.on('close', () => {
    console.log('连接关闭');
  });
});

// 启动服务器
app.listen('3000')
  .then(result => {
    console.log('服务器启动成功:', result);
  });

// 或者启动 HTTPS 服务器
app.listenTLS('8443', './certs/server.crt', './certs/server.key')
  .then(() => {
    console.log('HTTPS 服务器启动在 https://localhost:8443');
  });
```

### 🔌 WebSocket 客户端模块 (`websocket`/`ws`)

- **连接管理**: 支持 ws:// 和 wss:// 协议
- **消息发送**: 文本、JSON、二进制消息
- **事件支持**: message、close、error 事件
- **自动重连**: 支持自定义连接选项
- **Promise API**: 异步连接支持

```javascript
const ws = require('websocket');

// 连接到 WebSocket 服务器
ws.connect('ws://localhost:8080/chat', {
  timeout: 5000,  // 连接超时
  headers: {      // 自定义请求头
    'User-Agent': 'SW-Runtime-Client'
  }
}).then(client => {
  console.log('已连接到服务器');
  
  // 监听消息
  client.on('message', (data) => {
    console.log('收到消息:', data);
  });
  
  // 监听关闭事件
  client.on('close', () => {
    console.log('连接已关闭');
  });
  
  // 监听错误事件
  client.on('error', (err) => {
    console.error('WebSocket 错误:', err.message);
  });
  
  // 发送文本消息
  client.send('Hello Server!');
  
  // 发送 JSON 消息
  client.sendJSON({
    type: 'greeting',
    message: 'Hello from client!',
    timestamp: Date.now()
  });
  
  // 发送二进制消息
  client.sendBinary(new Uint8Array([1, 2, 3, 4]));
  
  // 发送 ping
  client.ping('heartbeat');
  
  // 检查连接状态
  if (!client.isClosed()) {
    console.log('连接正常');
  }
  
  // 关闭连接
  setTimeout(() => {
    client.close();
  }, 5000);
  
}).catch(err => {
  console.error('连接失败:', err.message);
});
```

**客户端 API 详解**:

```javascript
// connect(url, options) - 连接到服务器
ws.connect(url, {
  timeout: 10000,      // 连接超时（毫秒）
  headers: {},         // 自定义 HTTP 请求头
  protocols: []        // WebSocket 子协议
})

// 客户端对象方法
client.send(message)           // 发送文本消息
client.sendJSON(object)        // 发送 JSON 消息
client.sendBinary(data)        // 发送二进制消息
client.ping(data)              // 发送 ping 帧
client.close(code, reason)     // 关闭连接
client.isClosed()              // 检查连接状态
client.on(event, handler)      // 注册事件监听器

// 支持的事件
- 'message': 收到消息
- 'close': 连接关闭
- 'error': 发生错误
- 'pong': 收到 pong 响应
```

### 🌐 网络模块 (`net`)

- **TCP 服务器/客户端**: 支持 TCP 连接和通信
- **UDP 套接字**: 支持 UDP 数据包收发
- **事件驱动**: 基于事件的异步编程模式
- **Promise 支持**: 所有异步操作返回 Promise

```javascript
const net = require('net');

// TCP 服务器
const tcpServer = net.createTCPServer();

tcpServer.on('connection', (socket) => {
  console.log('新客户端连接:', socket.remoteAddress);
  
  socket.on('data', (data) => {
    console.log('收到:', data);
    socket.write('回显: ' + data);
  });
  
  socket.on('close', () => {
    console.log('客户端断开');
  });
});

tcpServer.listen('8080').then(() => {
  console.log('TCP 服务器启动在端口 8080');
});

// TCP 客户端
net.connectTCP('localhost:8080', { timeout: 5000 })
  .then(socket => {
    console.log('已连接到服务器');
    
    socket.on('data', (data) => {
      console.log('收到:', data);
    });
    
    socket.write('Hello Server!\n');
  });

// UDP 服务器
const udpSocket = net.createUDPSocket('udp4');

udpSocket.on('message', (msg, rinfo) => {
  console.log('收到来自', rinfo.address + ':' + rinfo.port, '的消息:', msg);
  
  // 回复客户端
  udpSocket.send('回复: ' + msg, rinfo.port.toString(), rinfo.address);
});

udpSocket.bind('9090', '0.0.0.0').then(() => {
  console.log('UDP 服务器监听端口 9090');
});

// UDP 客户端
const udpClient = net.createUDPSocket('udp4');
udpClient.send('Hello UDP!\n', '9090', 'localhost')
  .then(() => console.log('消息已发送'));
```

### 🔄 代理模块 (`proxy`)

- **HTTP 代理**: 反向代理 HTTP/HTTPS 请求
- **TCP 代理**: 透明 TCP 连接转发
- **事件驱动**: 基于事件的异步编程模式
- **自动处理**: HTTPS 自动处理、连接池管理
- **监控统计**: 请求/响应拦截、数据传输统计

```javascript
const proxy = require('proxy');

// HTTP 代理服务器
const httpProxy = proxy.createHTTPProxy('https://api.github.com');

httpProxy.on('request', (req) => {
  console.log(`请求: ${req.method} ${req.path}`);
});

httpProxy.on('response', (resp) => {
  console.log(`响应: ${resp.status}`);
});

httpProxy.on('error', (err) => {
  console.error('代理错误:', err.message);
});

httpProxy.listen('8080').then(() => {
  console.log('HTTP 代理启动在端口 8080');
});

// TCP 代理服务器
const tcpProxy = proxy.createTCPProxy('localhost:6379');

tcpProxy.on('connection', (conn) => {
  console.log('新连接:', conn.remoteAddr);
});

tcpProxy.on('data', (data) => {
  console.log(`${data.direction}: ${data.bytes} 字节`);
});

tcpProxy.on('close', () => {
  console.log('连接关闭');
});

tcpProxy.listen('6380').then(() => {
  console.log('TCP 代理启动在端口 6380');
});
```

### 🔴 Redis 客户端模块 (`redis`)

- **连接管理**: 支持连接配置、认证、数据库选择
- **数据类型**: 字符串、哈希、列表、集合、有序集合
- **JSON 支持**: 自动序列化/反序列化 JSON 数据
- **Promise 支持**: 所有操作返回 Promise

```javascript
const redis = require('redis');

// 创建连接
const client = redis.createClient({
  host: 'localhost',
  port: 6379,
  db: 0
});

// 字符串操作
await client.set('key', 'value', 60); // 60秒过期
const value = await client.get('key');

// JSON 数据
await client.setJSON('user:1', { name: 'John', age: 30 });
const user = await client.getJSON('user:1');

// 哈希操作
await client.hset('user:profile', 'name', 'Alice');
const profile = await client.hgetall('user:profile');

// 列表操作
await client.lpush('tasks', 'task1', 'task2');
const tasks = await client.lrange('tasks', 0, -1);

// 集合操作
await client.sadd('tags', 'javascript', 'redis');
const tags = await client.smembers('tags');
```

### 🗄️ SQLite 数据库模块 (`sqlite`)

- **数据库连接**: 内存数据库、文件数据库
- **SQL 操作**: 查询、插入、更新、删除
- **事务支持**: 自动事务、手动事务控制
- **预处理语句**: 提高性能和安全性
- **Promise 支持**: 所有操作返回 Promise

```javascript
const sqlite = require('sqlite');

// 打开数据库
const db = await sqlite.open('./database.db');
// 或内存数据库
const memDb = await sqlite.open(':memory:');

// 创建表
await db.exec(`
  CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE,
    age INTEGER
  )
`);

// 插入数据
const result = await db.run('INSERT INTO users (name, email, age) VALUES (?, ?, ?)', 
  ['张三', 'zhangsan@example.com', 25]);
console.log('插入ID:', result.lastInsertId);

// 查询单条记录
const user = await db.get('SELECT * FROM users WHERE id = ?', [1]);
console.log('用户:', user);

// 查询多条记录
const users = await db.all('SELECT * FROM users WHERE age > ?', [20]);
console.log('用户列表:', users);

// 使用事务
await db.transaction(async (tx) => {
  await tx.run('INSERT INTO users (name, email, age) VALUES (?, ?, ?)', 
    ['李四', 'lisi@example.com', 30]);
  await tx.run('UPDATE users SET age = ? WHERE name = ?', [26, '张三']);
});

// 预处理语句
const stmt = await db.prepare('SELECT * FROM users WHERE age > ?');
const olderUsers = await stmt.all(25);
await stmt.close();

// 获取数据库信息
const tables = await db.tables();
const schema = await db.schema('users');

// 关闭数据库
await db.close();
```

### 📁 文件系统模块 (`fs`)

- **同步操作**: `readFileSync`, `writeFileSync`, `existsSync`, `statSync`, `mkdirSync`, `readdirSync`, `unlinkSync`, `rmdirSync`, `copyFileSync`, `renameSync`
- **异步操作**: `readFile`, `writeFile`, `stat`, `mkdir`, `readdir`, `unlink`, `rmdir`, `copyFile`, `rename`
- **Promise 支持**: 所有异步操作返回 Promise

```javascript
const fs = require('fs');

// 同步操作
fs.writeFileSync('file.txt', 'content');
const content = fs.readFileSync('file.txt', 'utf8');

// 异步操作
fs.writeFile('file.txt', 'content')
  .then(() => fs.readFile('file.txt'))
  .then(content => console.log(content));
```

### 🛤️ 路径模块 (`path`)

- **路径操作**: `join`, `resolve`, `dirname`, `basename`, `extname`
- **路径判断**: `isAbsolute`, `relative`, `normalize`
- **跨平台支持**

```javascript
const path = require('path');

console.log(path.join('a', 'b', 'c'));        // a/b/c
console.log(path.resolve('./test'));          // 绝对路径
console.log(path.dirname('/a/b/c.js'));       // /a/b
console.log(path.basename('/a/b/c.js'));      // c.js
console.log(path.extname('test.js'));         // .js
```

## 🚀 使用示例

### 命令行工具

SW Runtime 提供完整的 CLI 工具，支持多种操作：

#### 运行脚本

```bash
# 运行 JavaScript 文件
sw_runtime run app.js

# 运行 TypeScript 文件
sw_runtime run app.ts

# 使用选项
sw_runtime run app.ts --clear-cache  # 清除模块缓存
```

#### 执行代码片段

```bash
# 执行 JavaScript 代码
sw_runtime eval "console.log('Hello, World!')"

# 执行复杂代码
sw_runtime eval "const x = 10; const y = 20; console.log(x + y)"

# 使用 Promise
sw_runtime eval "Promise.resolve(42).then(v => console.log(v))"
```

#### 打包脚本 🆕

```bash
# 基本打包
sw_runtime bundle app.js

# 指定输出文件
sw_runtime bundle app.js -o dist/bundle.js

# 压缩代码（70%+ 体积减少）
sw_runtime bundle app.js -o app.min.js --minify

# 生成 source map
sw_runtime bundle app.ts --sourcemap

# 详细输出
sw_runtime bundle app.js -v

# 排除特定文件
sw_runtime bundle app.js --exclude utils.js,test.js
```

详细文档请参阅：[docs/BUNDLE_GUIDE.md](docs/BUNDLE_GUIDE.md)

**打包功能特性：**
- ✅ 自动依赖解析 - 递归分析所有 `require()` 依赖
- ✅ TypeScript 支持 - 自动编译 `.ts` 文件
- ✅ 内置模块排除 - 智能排除运行时可用的内置模块
- ✅ 代码压缩 - 70%+ 的压缩率
- ✅ Source Map - 支持生成调试映射

#### 查看信息

```bash
# 显示版本
sw_runtime version

# 显示运行时信息
sw_runtime info

# 查看帮助
sw_runtime --help
sw_runtime bundle --help
```

### HTTP 客户端示例

```javascript
const http = require('http');

// 获取用户数据
http.get('https://jsonplaceholder.typicode.com/users/1')
  .then(response => {
    console.log('用户信息:', response.data);
    console.log('状态码:', response.status);
  })
  .catch(error => {
    console.error('请求失败:', error.message);
  });

// 创建新用户
http.post('https://jsonplaceholder.typicode.com/users', {
  data: {
    name: 'John Doe',
    email: 'john@example.com'
  },
  headers: {
    'Content-Type': 'application/json'
  }
})
  .then(response => {
    console.log('用户创建成功:', response.data);
  });
```

### Redis 客户端示例

```javascript
const redis = require('redis');

// 连接 Redis
const client = redis.createClient({
  host: 'localhost',
  port: 6379
});

// 基本操作
async function redisExample() {
  // 设置和获取字符串
  await client.set('username', 'john_doe');
  const username = await client.get('username');
  console.log('用户名:', username);

  // JSON 数据操作
  const userData = {
    id: 1,
    name: 'John Doe',
    email: 'john@example.com'
  };
  
  await client.setJSON('user:1', userData);
  const user = await client.getJSON('user:1');
  console.log('用户数据:', user);

  // 列表操作
  await client.lpush('notifications', 'Welcome!', 'New message');
  const notifications = await client.lrange('notifications', 0, -1);
  console.log('通知列表:', notifications);
}

redisExample().catch(console.error);
```

## 🔧 技术实现

- **Go 语言**: 高性能系统级编程
- **goja**: 纯 Go 实现的 JavaScript 引擎
- **esbuild**: 快速 TypeScript 编译
- **模块化设计**: 清晰的包结构和职责分离
- **并发安全**: 线程安全的模块缓存和异步操作

## 📊 性能特点

- **快速启动**: 无需 Node.js 环境
- **低内存占用**: 精简的运行时设计
- **高并发**: Go 协程支持异步操作
- **模块缓存**: 避免重复加载提升性能

## 🎯 适用场景

- **API 服务**: 内置 HTTP 客户端，轻松调用外部 API
- **网络服务**: TCP/UDP 服务器和客户端，支持实时通信
- **代理服务**: HTTP/TCP 代理服务器，请求转发和监控
- **数据缓存**: Redis 客户端支持高性能数据缓存
- **数据库应用**: SQLite 支持轻量级数据存储和查询
- **服务端脚本**: 替代 Node.js 的轻量级方案
- **配置脚本**: 动态配置和规则引擎
- **数据处理**: 支持加解密和压缩的数据管道
- **微服务**: 嵌入式 JavaScript 执行环境
- **自动化工具**: 跨平台脚本执行
- **爬虫和数据采集**: HTTP 客户端 + 数据处理
- **实时数据处理**: Redis + SQLite + 压缩 + 加密
- **网络通信**: TCP/UDP 协议应用、自定义协议实现
- **反向代理**: API 网关、负载均衡前端、服务路由

## 🔄 扩展性

系统采用插件化设计，可以轻松添加新的内置模块：

```go
// 添加自定义模块
manager.RegisterModule("mymodule", NewMyModule(vm))
```

这是一个企业级的 JavaScript/TypeScript 运行时，提供了完整的模块系统、HTTP/HTTPS/WebSocket/TCP/UDP/Proxy 网络功能、Redis/SQLite 客户端、加解密、压缩、文件操作等功能，适合各种服务端应用场景。