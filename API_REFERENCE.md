# SW Runtime API 参考文档

本文档提供 SW Runtime 所有内置模块的完整 API 接口说明，便于 AI 理解和使用。

## 目录
- [模块系统](#模块系统)
- [path - 路径模块](#path---路径模块)
- [fs - 文件系统模块](#fs---文件系统模块)
- [crypto - 加密模块](#crypto---加密模块)
- [compression/zlib - 压缩模块](#compressionzlib---压缩模块)
- [http - HTTP客户端模块](#http---http客户端模块)
- [httpserver/server - HTTP服务器模块](#httpserverserver---http服务器模块)
- [websocket/ws - WebSocket模块](#websocketws---websocket模块)
- [net - 网络模块](#net---网络模块)
- [redis - Redis客户端模块](#redis---redis客户端模块)
- [sqlite - SQLite数据库模块](#sqlite---sqlite数据库模块)
- [exec/child_process - 进程执行模块](#execchild_process---进程执行模块)

---

## 模块系统

### require(id: string): any
**功能**: CommonJS 风格的同步模块加载  
**参数**:
- `id` (string): 模块标识符，支持相对路径、绝对路径或内置模块名
**返回值**: 模块导出的内容  
**示例**:
```javascript
const fs = require('fs');
const utils = require('./utils.js');
const config = require('../config.json');
```

### import(id: string): Promise<any>
**功能**: ES6 风格的异步模块导入  
**参数**:
- `id` (string): 模块标识符
**返回值**: Promise<any> - 解析为模块导出的内容  
**示例**:
```javascript
import('./module.js').then(mod => console.log(mod));
```

---

## path - 路径模块

### join(...paths: string[]): string
**功能**: 连接多个路径片段  
**参数**: 任意数量的路径字符串  
**返回值**: 连接后的路径字符串  

### resolve(...paths: string[]): string
**功能**: 将路径解析为绝对路径  
**参数**: 任意数量的路径字符串  
**返回值**: 绝对路径字符串  

### dirname(path: string): string
**功能**: 获取路径的目录部分  
**参数**: `path` (string) - 文件路径  
**返回值**: 目录路径  

### basename(path: string, ext?: string): string
**功能**: 获取路径的基础文件名  
**参数**:
- `path` (string) - 文件路径
- `ext` (string, 可选) - 要移除的扩展名
**返回值**: 文件名  

### extname(path: string): string
**功能**: 获取路径的扩展名  
**参数**: `path` (string) - 文件路径  
**返回值**: 扩展名（包含点）  

### isAbsolute(path: string): boolean
**功能**: 判断路径是否为绝对路径  
**参数**: `path` (string) - 文件路径  
**返回值**: true/false  

### normalize(path: string): string
**功能**: 规范化路径  
**参数**: `path` (string) - 文件路径  
**返回值**: 规范化后的路径  

### relative(from: string, to: string): string
**功能**: 计算从 from 到 to 的相对路径  
**参数**:
- `from` (string) - 起始路径
- `to` (string) - 目标路径
**返回值**: 相对路径  

### 常量
- `sep`: 路径分隔符
- `delimiter`: 路径定界符

---

## fs - 文件系统模块

### 同步方法

#### readFileSync(path: string, encoding?: string): string
**功能**: 同步读取文件  
**参数**:
- `path` (string) - 文件路径
- `encoding` (string, 可选) - 编码格式，默认 'utf8'
**返回值**: 文件内容字符串  

#### writeFileSync(path: string, data: string, encoding?: string): void
**功能**: 同步写入文件  
**参数**:
- `path` (string) - 文件路径
- `data` (string) - 写入的数据
- `encoding` (string, 可选) - 编码格式，默认 'utf8'

#### existsSync(path: string): boolean
**功能**: 检查文件或目录是否存在  
**参数**: `path` (string) - 文件路径  
**返回值**: true/false  

#### statSync(path: string): object
**功能**: 获取文件或目录信息  
**参数**: `path` (string) - 文件路径  
**返回值**: 包含 `isFile()`, `isDirectory()`, `size`, `modTime` 等方法和属性的对象  

#### mkdirSync(path: string, recursive?: boolean): void
**功能**: 同步创建目录  
**参数**:
- `path` (string) - 目录路径
- `recursive` (boolean, 可选) - 是否递归创建

#### readdirSync(path: string): string[]
**功能**: 同步读取目录内容  
**参数**: `path` (string) - 目录路径  
**返回值**: 文件名数组  

#### unlinkSync(path: string): void
**功能**: 同步删除文件  
**参数**: `path` (string) - 文件路径  

#### rmdirSync(path: string): void
**功能**: 同步删除目录  
**参数**: `path` (string) - 目录路径  

#### copyFileSync(src: string, dest: string): void
**功能**: 同步复制文件  
**参数**:
- `src` (string) - 源文件路径
- `dest` (string) - 目标文件路径

#### renameSync(oldPath: string, newPath: string): void
**功能**: 同步重命名或移动文件  
**参数**:
- `oldPath` (string) - 原路径
- `newPath` (string) - 新路径

### 异步方法（Promise）

所有同步方法都有对应的异步版本，去掉 `Sync` 后缀，返回 Promise：
- `readFile(path, encoding?): Promise<string>`
- `writeFile(path, data, encoding?): Promise<void>`
- `exists(path): Promise<boolean>`
- `stat(path): Promise<object>`
- `mkdir(path, recursive?): Promise<void>`
- `readdir(path): Promise<string[]>`
- `unlink(path): Promise<void>`
- `rmdir(path): Promise<void>`
- `copyFile(src, dest): Promise<void>`
- `rename(oldPath, newPath): Promise<void>`

---

## crypto - 加密模块

### 哈希函数

#### md5(data: string): string
**功能**: 计算 MD5 哈希值  
**参数**: `data` (string) - 输入数据  
**返回值**: 十六进制哈希字符串  

#### sha1(data: string): string
**功能**: 计算 SHA1 哈希值  
**参数**: `data` (string) - 输入数据  
**返回值**: 十六进制哈希字符串  

#### sha256(data: string): string
**功能**: 计算 SHA256 哈希值  
**参数**: `data` (string) - 输入数据  
**返回值**: 十六进制哈希字符串  

#### sha512(data: string): string
**功能**: 计算 SHA512 哈希值  
**参数**: `data` (string) - 输入数据  
**返回值**: 十六进制哈希字符串  

### 编解码

#### base64Encode(data: string): string
**功能**: Base64 编码  
**参数**: `data` (string) - 原始数据  
**返回值**: Base64 编码字符串  

#### base64Decode(data: string): string
**功能**: Base64 解码  
**参数**: `data` (string) - Base64 编码字符串  
**返回值**: 解码后的原始数据  

#### hexEncode(data: string): string
**功能**: 十六进制编码  
**参数**: `data` (string) - 原始数据  
**返回值**: 十六进制字符串  

#### hexDecode(data: string): string
**功能**: 十六进制解码  
**参数**: `data` (string) - 十六进制字符串  
**返回值**: 解码后的原始数据  

### 加密

#### aesEncrypt(data: string, key: string): string
**功能**: AES-256-GCM 加密  
**参数**:
- `data` (string) - 待加密数据
- `key` (string) - 加密密钥
**返回值**: Base64 编码的加密数据  

#### aesDecrypt(data: string, key: string): string
**功能**: AES-256-GCM 解密  
**参数**:
- `data` (string) - Base64 编码的加密数据
- `key` (string) - 解密密钥
**返回值**: 解密后的原始数据  

#### randomBytes(size?: number): string
**功能**: 生成安全随机字节  
**参数**: `size` (number, 可选) - 字节数，默认 16  
**返回值**: 十六进制编码的随机字节  

---

## compression/zlib - 压缩模块

### gzipCompress(data: string): string
**功能**: Gzip 压缩  
**参数**: `data` (string) - 原始数据  
**返回值**: Base64 编码的压缩数据  

### gzipDecompress(data: string): string
**功能**: Gzip 解压  
**参数**: `data` (string) - Base64 编码的压缩数据  
**返回值**: 解压后的原始数据  

### zlibCompress(data: string): string
**功能**: Zlib 压缩  
**参数**: `data` (string) - 原始数据  
**返回值**: Base64 编码的压缩数据  

### zlibDecompress(data: string): string
**功能**: Zlib 解压  
**参数**: `data` (string) - Base64 编码的压缩数据  
**返回值**: 解压后的原始数据  

---

## http - HTTP客户端模块

### HTTP 方法

所有 HTTP 方法返回 Promise<HTTPResponse>

#### get(url: string, config?: RequestConfig): Promise<HTTPResponse>
**功能**: 发送 GET 请求  

#### post(url: string, config?: RequestConfig): Promise<HTTPResponse>
**功能**: 发送 POST 请求  

#### put(url: string, config?: RequestConfig): Promise<HTTPResponse>
**功能**: 发送 PUT 请求  

#### delete(url: string, config?: RequestConfig): Promise<HTTPResponse>
**功能**: 发送 DELETE 请求  

#### patch(url: string, config?: RequestConfig): Promise<HTTPResponse>
**功能**: 发送 PATCH 请求  

#### head(url: string, config?: RequestConfig): Promise<HTTPResponse>
**功能**: 发送 HEAD 请求  

#### options(url: string, config?: RequestConfig): Promise<HTTPResponse>
**功能**: 发送 OPTIONS 请求  

#### request(url: string, config?: RequestConfig): Promise<HTTPResponse>
**功能**: 通用请求方法  

### RequestConfig 对象
```typescript
{
  method?: string,          // HTTP 方法
  headers?: object,         // 请求头
  data?: any,              // 请求体（自动 JSON 序列化）
  params?: object,         // URL 查询参数
  timeout?: number,        // 超时时间（秒），默认 30
  auth?: {                 // 认证信息
    username?: string,
    password?: string,
    token?: string         // Bearer token
  }
}
```

### HTTPResponse 对象
```typescript
{
  status: number,          // HTTP 状态码
  statusText: string,      // 状态文本
  headers: object,         // 响应头
  data: any,              // 响应数据（自动 JSON 解析）
  text: string,           // 原始响应文本
  url: string             // 请求 URL
}
```

### createClient(config?: {timeout?: number}): HTTPClient
**功能**: 创建自定义 HTTP 客户端实例  
**参数**: 可选配置对象  
**返回值**: 具有所有 HTTP 方法的客户端对象  

### STATUS_CODES 常量
```javascript
{
  OK: 200,
  CREATED: 201,
  NO_CONTENT: 204,
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  INTERNAL_SERVER_ERROR: 500
}
```

---

## httpserver/server - HTTP服务器模块

### createServer(): HTTPServer
**功能**: 创建 HTTP 服务器实例  
**返回值**: HTTPServer 对象  

### HTTPServer 对象方法

#### listen(port: string|number, callback?: function): Promise<string>
**功能**: 启动服务器监听指定端口  
**参数**:
- `port` (string|number) - 端口号
- `callback` (function, 可选) - 启动成功回调
**返回值**: Promise - 解析为启动成功消息  

#### use(middleware: function): void
**功能**: 添加中间件  
**参数**: `middleware` (function) - 中间件函数 `(req, res, next) => {}`

#### get(path: string, handler: function): void
**功能**: 添加 GET 路由  
**参数**:
- `path` (string) - 路由路径
- `handler` (function) - 请求处理函数 `(req, res) => {}`

#### post(path: string, handler: function): void
**功能**: 添加 POST 路由  

#### put(path: string, handler: function): void
**功能**: 添加 PUT 路由  

#### delete(path: string, handler: function): void
**功能**: 添加 DELETE 路由  

#### static(directory: string, urlPath?: string): void
**功能**: 设置静态文件服务  
**参数**:
- `directory` (string) - 静态文件目录
- `urlPath` (string, 可选) - URL 路径前缀，默认 '/'

#### ws(path: string, handler: function): void
**功能**: 添加 WebSocket 路由  
**参数**:
- `path` (string) - WebSocket 路由路径
- `handler` (function) - WebSocket 处理函数 `(ws) => {}`

#### close(): Promise<void>
**功能**: 关闭服务器  
**返回值**: Promise  

### Request 对象（req）
```typescript
{
  method: string,          // HTTP 方法
  path: string,            // 请求路径
  url: string,             // 完整 URL
  headers: object,         // 请求头
  query: object,           // 查询参数
  body: string,            // 原始请求体
  json: any               // 自动解析的 JSON 数据
}
```

### Response 对象（res）

#### status(code: number): Response
**功能**: 设置响应状态码  
**参数**: `code` (number) - HTTP 状态码  
**返回值**: Response 对象（链式调用）  

#### header(name: string, value: string): Response
**功能**: 设置响应头  
**参数**:
- `name` (string) - 响应头名称
- `value` (string) - 响应头值
**返回值**: Response 对象（链式调用）  

#### send(data: string): Response
**功能**: 发送文本响应  
**参数**: `data` (string) - 响应内容  

#### json(data: any): Response
**功能**: 发送 JSON 响应  
**参数**: `data` (any) - 响应数据（自动序列化）  

#### html(data: string): Response
**功能**: 发送 HTML 响应  
**参数**: `data` (string) - HTML 内容  

#### sendFile(path: string): Response
**功能**: 发送文件（自动检测 MIME 类型）  
**参数**: `path` (string) - 文件路径  

#### download(path: string, filename?: string): Response
**功能**: 发送文件下载响应  
**参数**:
- `path` (string) - 文件路径
- `filename` (string, 可选) - 下载文件名

#### redirect(url: string, code?: number): Response
**功能**: 重定向  
**参数**:
- `url` (string) - 重定向 URL
- `code` (number, 可选) - 状态码，默认 302

### WebSocket 对象（ws）

#### send(message: string): void
**功能**: 发送文本消息  
**参数**: `message` (string) - 消息内容  

#### sendJSON(data: any): void
**功能**: 发送 JSON 消息  
**参数**: `data` (any) - 数据对象  

#### on(event: string, handler: function): void
**功能**: 监听事件  
**参数**:
- `event` (string) - 事件名称（'message', 'close', 'error'）
- `handler` (function) - 事件处理函数

#### close(): void
**功能**: 关闭连接  

---

## websocket/ws - WebSocket模块

### connect(url: string, options?: ConnectOptions): Promise<WebSocketClient>
**功能**: 连接到 WebSocket 服务器  
**参数**:
- `url` (string) - WebSocket URL（ws:// 或 wss://）
- `options` (object, 可选) - 连接选项
**返回值**: Promise<WebSocketClient>  

### ConnectOptions 对象
```typescript
{
  timeout?: number,        // 连接超时（毫秒），默认 10000
  headers?: object,        // 自定义 HTTP 请求头
  protocols?: string[]     // WebSocket 子协议
}
```

### WebSocketClient 对象方法

#### send(message: string): void
**功能**: 发送文本消息  
**参数**: `message` (string) - 消息内容  

#### sendJSON(data: any): void
**功能**: 发送 JSON 消息  
**参数**: `data` (any) - 数据对象（自动序列化）  

#### sendBinary(data: ArrayBuffer|Uint8Array): void
**功能**: 发送二进制消息  
**参数**: `data` - 二进制数据  

#### ping(data?: string): void
**功能**: 发送 ping 帧  
**参数**: `data` (string, 可选) - ping 数据  

#### close(code?: number, reason?: string): void
**功能**: 关闭连接  
**参数**:
- `code` (number, 可选) - 关闭代码
- `reason` (string, 可选) - 关闭原因

#### isClosed(): boolean
**功能**: 检查连接是否已关闭  
**返回值**: true/false  

#### on(event: string, handler: function): void
**功能**: 监听事件  
**参数**:
- `event` (string) - 事件名称
- `handler` (function) - 事件处理函数

### 支持的事件
- `'message'`: 收到消息 - `handler(data: string)`
- `'close'`: 连接关闭 - `handler()`
- `'error'`: 发生错误 - `handler(error: {message: string})`
- `'pong'`: 收到 pong 响应 - `handler(data: string)`

---

## net - 网络模块

### TCP 功能

#### createTCPServer(): TCPServer
**功能**: 创建 TCP 服务器实例  
**返回值**: TCPServer 对象  

#### connectTCP(address: string, options?: ConnectOptions): Promise<TCPSocket>
**功能**: 连接到 TCP 服务器  
**参数**:
- `address` (string) - 服务器地址，格式为 "host:port"
- `options` (object, 可选) - 连接选项
**返回值**: Promise<TCPSocket>  

**ConnectOptions 对象**:
```typescript
{
  timeout?: number         // 连接超时（毫秒），默认 10000
}
```

### TCPServer 对象方法

#### listen(port: string|number, callback?: function): Promise<string>
**功能**: 启动 TCP 服务器监听指定端口  
**参数**:
- `port` (string|number) - 端口号
- `callback` (function, 可选) - 启动成功回调
**返回值**: Promise<string> - 解析为启动成功消息  

#### on(event: string, handler: function): TCPServer
**功能**: 注册事件处理器  
**参数**:
- `event` (string) - 事件名称
- `handler` (function) - 事件处理函数
**返回值**: TCPServer 对象（链式调用）  

**支持的事件**:
- `'connection'`: 新客户端连接 - `handler(socket: TCPSocket)`

#### close(): Promise<void>
**功能**: 关闭 TCP 服务器  
**返回值**: Promise<void>  

### TCPSocket 对象

#### 属性
- `remoteAddress` (string): 远程地址
- `localAddress` (string): 本地地址

#### write(data: string): Promise<boolean>
**功能**: 发送数据  
**参数**: `data` (string) - 要发送的数据  
**返回值**: Promise<boolean>  

#### on(event: string, handler: function): TCPSocket
**功能**: 注册事件处理器  
**参数**:
- `event` (string) - 事件名称
- `handler` (function) - 事件处理函数

**支持的事件**:
- `'data'`: 收到数据 - `handler(data: string)`
- `'close'`: 连接关闭 - `handler()`
- `'error'`: 发生错误 - `handler(error: {message: string})`

#### close(): void
**功能**: 关闭连接  

#### setTimeout(timeout: number): TCPSocket
**功能**: 设置连接超时  
**参数**: `timeout` (number) - 超时时间（毫秒）  
**返回值**: TCPSocket 对象（链式调用）  

### UDP 功能

#### createUDPSocket(type?: string): UDPSocket
**功能**: 创建 UDP 套接字  
**参数**: `type` (string, 可选) - 套接字类型，'udp4' 或 'udp6'，默认 'udp4'  
**返回值**: UDPSocket 对象  

### UDPSocket 对象方法

#### bind(port: string, host?: string, callback?: function): Promise<string>
**功能**: 绑定 UDP 套接字到指定端口  
**参数**:
- `port` (string) - 端口号
- `host` (string, 可选) - 主机地址，默认 '0.0.0.0'
- `callback` (function, 可选) - 绑定成功回调
**返回值**: Promise<string>  

#### send(data: string, port: string, host: string, callback?: function): Promise<boolean>
**功能**: 发送 UDP 数据包  
**参数**:
- `data` (string) - 要发送的数据
- `port` (string) - 目标端口
- `host` (string) - 目标主机
- `callback` (function, 可选) - 发送成功回调
**返回值**: Promise<boolean>  

#### on(event: string, handler: function): UDPSocket
**功能**: 注册事件处理器  
**参数**:
- `event` (string) - 事件名称
- `handler` (function) - 事件处理函数

**支持的事件**:
- `'message'`: 收到消息 - `handler(msg: string, rinfo: {address: string, port: number, data: string})`
- `'close'`: 套接字关闭 - `handler()`
- `'error'`: 发生错误 - `handler(error: {message: string})`

#### close(): void
**功能**: 关闭 UDP 套接字  

#### address(): object|undefined
**功能**: 获取套接字地址信息  
**返回值**: 地址对象 `{address: string, port: number, family: string}` 或 undefined  

### 示例代码

#### TCP 服务器示例
```javascript
const net = require('net');

const server = net.createTCPServer();

server.on('connection', (socket) => {
  console.log('新客户端连接:', socket.remoteAddress);
  
  socket.write('欢迎使用 TCP 服务器!\n');
  
  socket.on('data', (data) => {
    console.log('收到数据:', data);
    socket.write('回显: ' + data);
  });
  
  socket.on('close', () => {
    console.log('客户端断开连接');
  });
});

server.listen('8080').then(() => {
  console.log('TCP 服务器监听端口 8080');
});
```

#### TCP 客户端示例
```javascript
const net = require('net');

net.connectTCP('localhost:8080', { timeout: 5000 })
  .then(socket => {
    console.log('已连接到服务器');
    
    socket.on('data', (data) => {
      console.log('收到:', data);
    });
    
    socket.write('Hello Server!\n');
  })
  .catch(err => {
    console.error('连接失败:', err.message);
  });
```

#### UDP 服务器示例
```javascript
const net = require('net');

const socket = net.createUDPSocket('udp4');

socket.on('message', (msg, rinfo) => {
  console.log('收到来自', rinfo.address + ':' + rinfo.port, '的消息:', msg);
  
  // 回复客户端
  socket.send('回复: ' + msg, rinfo.port.toString(), rinfo.address);
});

socket.bind('9090', '0.0.0.0').then(() => {
  console.log('UDP 服务器监听端口 9090');
});
```

#### UDP 客户端示例
```javascript
const net = require('net');

const socket = net.createUDPSocket('udp4');

// 发送消息
socket.send('Hello UDP Server!\n', '9090', 'localhost')
  .then(() => {
    console.log('消息已发送');
  })
  .catch(err => {
    console.error('发送失败:', err.message);
  });

// 绑定本地端口以接收回复
socket.bind('0', '0.0.0.0').then(() => {
  socket.on('message', (msg, rinfo) => {
    console.log('收到回复:', msg);
  });
});
```

---

## redis - Redis客户端模块

### createClient(config?: RedisConfig): RedisClient
**功能**: 创建 Redis 客户端  
**参数**: 可选配置对象  
**返回值**: RedisClient 对象  

### RedisConfig 对象
```typescript
{
  host?: string,           // 主机地址，默认 'localhost'
  port?: number,           // 端口，默认 6379
  password?: string,       // 密码
  db?: number             // 数据库编号，默认 0
}
```

### RedisClient 对象方法（所有方法返回 Promise）

#### 字符串操作

##### set(key: string, value: string, expiration?: number): Promise<string>
**功能**: 设置键值  
**参数**:
- `key` (string) - 键名
- `value` (string) - 值
- `expiration` (number, 可选) - 过期时间（秒）
**返回值**: Promise<string> - 'OK'  

##### get(key: string): Promise<string|null>
**功能**: 获取键值  
**参数**: `key` (string) - 键名  
**返回值**: Promise<string|null>  

##### setJSON(key: string, value: any, expiration?: number): Promise<string>
**功能**: 设置 JSON 数据  
**参数**:
- `key` (string) - 键名
- `value` (any) - 数据对象（自动序列化）
- `expiration` (number, 可选) - 过期时间（秒）

##### getJSON(key: string): Promise<any|null>
**功能**: 获取 JSON 数据  
**参数**: `key` (string) - 键名  
**返回值**: Promise<any|null> - 自动反序列化的对象  

##### del(key: string): Promise<number>
**功能**: 删除键  
**参数**: `key` (string) - 键名  
**返回值**: Promise<number> - 删除的键数量  

##### exists(key: string): Promise<boolean>
**功能**: 检查键是否存在  
**参数**: `key` (string) - 键名  
**返回值**: Promise<boolean>  

##### expire(key: string, seconds: number): Promise<boolean>
**功能**: 设置过期时间  
**参数**:
- `key` (string) - 键名
- `seconds` (number) - 过期时间（秒）
**返回值**: Promise<boolean>  

##### ttl(key: string): Promise<number>
**功能**: 获取剩余生存时间  
**参数**: `key` (string) - 键名  
**返回值**: Promise<number> - 剩余秒数，-1 表示永不过期，-2 表示不存在  

#### 哈希操作

##### hset(key: string, field: string, value: string): Promise<number>
**功能**: 设置哈希字段  

##### hget(key: string, field: string): Promise<string|null>
**功能**: 获取哈希字段  

##### hgetall(key: string): Promise<object>
**功能**: 获取所有哈希字段  
**返回值**: Promise<object> - 字段-值对象  

##### hdel(key: string, field: string): Promise<number>
**功能**: 删除哈希字段  

##### hexists(key: string, field: string): Promise<boolean>
**功能**: 检查哈希字段是否存在  

##### hkeys(key: string): Promise<string[]>
**功能**: 获取所有哈希字段名  

##### hvals(key: string): Promise<string[]>
**功能**: 获取所有哈希字段值  

#### 列表操作

##### lpush(key: string, ...values: string[]): Promise<number>
**功能**: 从左侧推入元素  
**返回值**: Promise<number> - 列表长度  

##### rpush(key: string, ...values: string[]): Promise<number>
**功能**: 从右侧推入元素  

##### lpop(key: string): Promise<string|null>
**功能**: 从左侧弹出元素  

##### rpop(key: string): Promise<string|null>
**功能**: 从右侧弹出元素  

##### lrange(key: string, start: number, stop: number): Promise<string[]>
**功能**: 获取列表范围元素  
**参数**:
- `key` (string) - 键名
- `start` (number) - 起始索引
- `stop` (number) - 结束索引（-1 表示到末尾）
**返回值**: Promise<string[]>  

##### llen(key: string): Promise<number>
**功能**: 获取列表长度  

#### 集合操作

##### sadd(key: string, ...members: string[]): Promise<number>
**功能**: 添加集合成员  
**返回值**: Promise<number> - 添加的成员数量  

##### srem(key: string, ...members: string[]): Promise<number>
**功能**: 删除集合成员  

##### smembers(key: string): Promise<string[]>
**功能**: 获取所有集合成员  

##### sismember(key: string, member: string): Promise<boolean>
**功能**: 检查是否为集合成员  

##### scard(key: string): Promise<number>
**功能**: 获取集合大小  

#### 有序集合操作

##### zadd(key: string, score: number, member: string): Promise<number>
**功能**: 添加有序集合成员  
**参数**:
- `key` (string) - 键名
- `score` (number) - 分数
- `member` (string) - 成员

##### zrange(key: string, start: number, stop: number, withScores?: boolean): Promise<any[]>
**功能**: 按索引范围获取成员  
**参数**:
- `key` (string) - 键名
- `start` (number) - 起始索引
- `stop` (number) - 结束索引
- `withScores` (boolean, 可选) - 是否返回分数

##### zscore(key: string, member: string): Promise<number|null>
**功能**: 获取成员分数  

##### zcard(key: string): Promise<number>
**功能**: 获取有序集合大小  

#### 通用操作

##### ping(): Promise<string>
**功能**: 测试连接  
**返回值**: Promise<string> - 'PONG'  

##### close(): Promise<void>
**功能**: 关闭连接  

---

## sqlite - SQLite数据库模块

### open(path: string): Promise<Database>
**功能**: 打开数据库连接  
**参数**: `path` (string) - 数据库文件路径，':memory:' 表示内存数据库  
**返回值**: Promise<Database>  

### Database 对象方法

#### exec(sql: string): Promise<void>
**功能**: 执行 SQL 语句（不返回结果）  
**参数**: `sql` (string) - SQL 语句  
**返回值**: Promise<void>  

#### run(sql: string, params?: any[]): Promise<RunResult>
**功能**: 执行 SQL 语句（INSERT、UPDATE、DELETE）  
**参数**:
- `sql` (string) - SQL 语句，支持 ? 占位符
- `params` (array, 可选) - 参数数组
**返回值**: Promise<RunResult> - `{lastInsertId, rowsAffected}`  

#### get(sql: string, params?: any[]): Promise<object|null>
**功能**: 查询单条记录  
**参数**:
- `sql` (string) - SQL 查询语句
- `params` (array, 可选) - 参数数组
**返回值**: Promise<object|null> - 记录对象  

#### all(sql: string, params?: any[]): Promise<object[]>
**功能**: 查询多条记录  
**参数**:
- `sql` (string) - SQL 查询语句
- `params` (array, 可选) - 参数数组
**返回值**: Promise<object[]> - 记录数组  

#### prepare(sql: string): Promise<Statement>
**功能**: 创建预处理语句  
**参数**: `sql` (string) - SQL 语句  
**返回值**: Promise<Statement>  

#### transaction(callback: function): Promise<void>
**功能**: 执行事务  
**参数**: `callback` (function) - 事务函数 `async (tx) => {}`  
**返回值**: Promise<void>  

#### close(): Promise<void>
**功能**: 关闭数据库连接  

#### tables(): Promise<string[]>
**功能**: 获取所有表名  
**返回值**: Promise<string[]>  

#### schema(tableName: string): Promise<object[]>
**功能**: 获取表结构  
**参数**: `tableName` (string) - 表名  
**返回值**: Promise<object[]> - 列信息数组  

### Statement 对象方法

#### run(params?: any[]): Promise<RunResult>
**功能**: 执行预处理语句  

#### get(params?: any[]): Promise<object|null>
**功能**: 查询单条记录  

#### all(params?: any[]): Promise<object[]>
**功能**: 查询多条记录  

#### close(): Promise<void>
**功能**: 关闭预处理语句  

---

## exec/child_process - 进程执行模块

### execSync(command: string, args?: string[], options?: ExecOptions): object
**功能**: 同步执行命令  
**参数**:
- `command` (string) - 命令名称
- `args` (string[], 可选) - 命令参数
- `options` (object, 可选) - 执行选项
**返回值**: 结果对象  

### exec(command: string, args?: string[], options?: ExecOptions): Promise<object>
**功能**: 异步执行命令  
**参数**:
- `command` (string) - 命令名称
- `args` (string[], 可选) - 命令参数
- `options` (object, 可选) - 执行选项
**返回值**: Promise<object>  

### ExecOptions 对象
```typescript
{
  cwd?: string,            // 工作目录
  env?: object,            // 环境变量
  timeout?: number         // 超时时间（毫秒）
}
```

### 执行结果对象
```typescript
{
  stdout: string,          // 标准输出
  stderr: string,          // 标准错误
  exitCode: number,        // 退出码
  success: boolean,        // 是否成功
  error: string|null,      // 错误信息
  command: string,         // 执行的命令
  args: string[],          // 命令参数
  timedOut?: boolean       // 是否超时（仅异步）
}
```

### getEnv(key?: string, defaultValue?: string): any
**功能**: 获取环境变量  
**参数**:
- `key` (string, 可选) - 环境变量名，省略则返回所有环境变量
- `defaultValue` (string, 可选) - 默认值
**返回值**: 环境变量值或环境变量对象  

### setEnv(key: string, value: string): boolean
**功能**: 设置环境变量  
**参数**:
- `key` (string) - 环境变量名
- `value` (string) - 环境变量值
**返回值**: boolean - 是否成功  

### which(command: string): string|null
**功能**: 查找命令路径  
**参数**: `command` (string) - 命令名称  
**返回值**: string|null - 命令完整路径  

### commandExists(command: string): boolean
**功能**: 检查命令是否存在  
**参数**: `command` (string) - 命令名称  
**返回值**: boolean  

---

## 全局对象

### console
- `console.log(...args)`: 输出日志
- `console.error(...args)`: 输出错误
- `console.warn(...args)`: 输出警告
- `console.info(...args)`: 输出信息

### 定时器
- `setTimeout(callback, delay, ...args)`: 延迟执行
- `clearTimeout(id)`: 取消延迟执行
- `setInterval(callback, interval, ...args)`: 定时执行
- `clearInterval(id)`: 取消定时执行

### Promise
完整的 Promise/A+ 实现，支持 `then`, `catch`, `finally`, `Promise.all`, `Promise.race` 等

---

## 类型约定

### 异步操作
所有异步操作返回 Promise 对象，可使用 `async/await` 或 `.then()` 处理。

### 错误处理
- 同步方法抛出异常
- 异步方法返回 rejected Promise
- 建议使用 try-catch 或 Promise.catch() 捕获错误

### 路径处理
- 支持相对路径和绝对路径
- Windows 和 Unix 路径自动处理
- 建议使用 `path` 模块规范化路径

### 数据序列化
- JSON 数据自动序列化和反序列化
- 支持嵌套对象和数组
- 循环引用会导致错误

---

## 示例代码

### 完整 Web 服务器示例
```javascript
const server = require('httpserver');
const fs = require('fs');
const redis = require('redis');

const app = server.createServer();
const redisClient = redis.createClient({ host: 'localhost' });

// 中间件
app.use((req, res, next) => {
  console.log(`${req.method} ${req.path}`);
  next();
});

// API 路由
app.get('/api/users', async (req, res) => {
  const users = await redisClient.getJSON('users') || [];
  res.json({ users });
});

app.post('/api/users', async (req, res) => {
  const user = req.json;
  const users = await redisClient.getJSON('users') || [];
  users.push(user);
  await redisClient.setJSON('users', users);
  res.status(201).json({ message: 'Created', user });
});

// 静态文件
app.static('./public', '/');

// 启动服务器
app.listen('3000').then(() => {
  console.log('Server running on http://localhost:3000');
});
```

### 数据库操作示例
```javascript
const sqlite = require('sqlite');

async function main() {
  const db = await sqlite.open('./app.db');
  
  // 创建表
  await db.exec(`
    CREATE TABLE IF NOT EXISTS users (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      email TEXT UNIQUE
    )
  `);
  
  // 插入数据
  const result = await db.run(
    'INSERT INTO users (name, email) VALUES (?, ?)',
    ['Alice', 'alice@example.com']
  );
  console.log('Inserted ID:', result.lastInsertId);
  
  // 查询数据
  const users = await db.all('SELECT * FROM users');
  console.log('Users:', users);
  
  // 使用事务
  await db.transaction(async (tx) => {
    await tx.run('INSERT INTO users (name, email) VALUES (?, ?)', 
      ['Bob', 'bob@example.com']);
    await tx.run('UPDATE users SET name = ? WHERE email = ?',
      ['Bobby', 'bob@example.com']);
  });
  
  await db.close();
}

main().catch(console.error);
```

---

本文档涵盖了 SW Runtime 的所有内置模块 API。所有接口均经过测试验证，可直接用于生产环境。
