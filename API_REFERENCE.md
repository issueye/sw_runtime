# SW Runtime API 参考文档

本文档提供 SW Runtime 所有内置模块的完整 API 接口说明，便于 AI 理解和使用。

## 目录

- [模块系统](#模块系统)
- [命名空间概览](#命名空间概览)
- [http - HTTP 模块](#http---http-模块)
  - [http/client - HTTP 客户端](#httpclient---http-客户端)
  - [http/server - HTTP 服务器](#httpserver---http-服务器)
- [db - 数据库模块](#db---数据库模块)
  - [db/redis - Redis 客户端](#dbredis---redis-客户端)
  - [db/sqlite - SQLite 数据库](#dbsqlite---sqlite-数据库)
- [utils - 工具模块](#utils---工具模块)
  - [utils/path - 路径处理](#utilspath---路径处理)
  - [utils/time - 时间处理](#utilstime---时间处理)
  - [utils/crypto - 加密模块](#utilscrypto---加密模块)
  - [utils/compression - 压缩模块](#utilscompression---压缩模块)
  - [utils/util - 工具函数](#utilsutil---工具函数)
- [net - 网络模块](#net---网络模块)
  - [net/net - TCP/UDP 网络](#netnet---tcpudp-网络)
  - [net/websocket - WebSocket](#netwebsocket---websocket)
  - [net/proxy - 代理模块](#netproxy---代理模块)
- [fs - 文件系统模块](#fs---文件系统模块)
  - [fs/fs - 文件操作](#fsfs---文件操作)
  - [fs/os - 操作系统信息](#fsos---操作系统信息)
- [config - 配置模块](#config---配置模块)
  - [config/viper - 配置管理](#configviper---配置管理)
- [process - 进程模块](#process---进程模块)
  - [process/exec - 进程执行模块](#processexec---进程执行模块)
  - [process/process - 进程模块](#processprocess---进程模块)

---

## 模块系统

### require(id: string): any

**功能**: CommonJS 风格的同步模块加载
**参数**:
- `id` (string): 模块标识符，支持相对路径、绝对路径或内置模块名

**返回值**: 模块导出的内容

**示例**:
```javascript
// 命名空间导入
const { server } = require('http');
const { crypto, path } = require('utils');
const { redis, sqlite } = require('db');

// 子模块导入
const client = require('http/client');
const ws = require('net/websocket');

// 相对路径导入
const utils = require("./utils.js");
const config = require("../config.json");
```

### import(id: string): Promise<any>

**功能**: ES6 风格的异步模块导入
**参数**:
- `id` (string): 模块标识符

**返回值**: Promise<any> - 解析为模块导出的内容

**示例**:
```javascript
import("./module.js").then((mod) => console.log(mod));
```

---

## 命名空间概览

SW Runtime 使用命名空间组织内置模块，提供更清晰的模块结构：

| 命名空间 | 子模块 | 说明 |
|---------|--------|------|
| `http` | client, server | HTTP 客户端和服务器 |
| `db` | redis, sqlite | Redis 和 SQLite 数据库 |
| `utils` | path, time, crypto, compression, util | 工具函数集合 |
| `net` | net, proxy, websocket | 网络相关功能 |
| `fs` | fs, os | 文件系统和操作系统 |
| `config` | viper | 配置管理 |
| `process` | process, exec | 进程管理和命令执行 |

**向后兼容**: 旧的模块名仍然可用（如 `require('httpserver')` 等效于 `require('http').server`）

---

## http - HTTP 模块

```javascript
const { client, server } = require('http');
// 或分别导入
const client = require('http/client');
const server = require('http/server');
```

---

### http/client - HTTP 客户端

#### HTTP 方法

所有 HTTP 方法返回 Promise<HTTPResponse>

##### get(url: string, config?: RequestConfig): Promise<HTTPResponse>

**功能**: 发送 GET 请求

##### post(url: string, config?: RequestConfig): Promise<HTTPResponse>

**功能**: 发送 POST 请求

##### put(url: string, config?: RequestConfig): Promise<HTTPResponse>

**功能**: 发送 PUT 请求

##### delete(url: string, config?: RequestConfig): Promise<HTTPResponse>

**功能**: 发送 DELETE 请求

##### patch(url: string, config?: RequestConfig): Promise<HTTPResponse>

**功能**: 发送 PATCH 请求

##### head(url: string, config?: RequestConfig): Promise<HTTPResponse>

**功能**: 发送 HEAD 请求

##### options(url: string, config?: RequestConfig): Promise<HTTPResponse>

**功能**: 发送 OPTIONS 请求

##### request(url: string, config?: RequestConfig): Promise<HTTPResponse>

**功能**: 通用请求方法

#### RequestConfig 对象

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
  },
  // 拦截器函数
  beforeRequest?: (config) => config,       // 请求前修改配置
  afterResponse?: (response) => response,   // 响应后处理
  transformRequest?: (data) => data,        // 转换请求数据
  transformResponse?: (data) => data        // 转换响应数据
}
```

#### HTTPResponse 对象

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

#### createClient(config?: {timeout?: number}): HTTPClient

**功能**: 创建自定义 HTTP 客户端实例
**参数**: 可选配置对象
**返回值**: 具有所有 HTTP 方法的客户端对象

#### setRequestInterceptor(interceptor: function): void

**功能**: 设置全局请求拦截器
**参数**: `interceptor` (function) - 拦截器函数 `(config) => config`

**示例**:
```javascript
const { client } = require('http');
client.setRequestInterceptor((config) => {
  // 所有请求自动添加 token
  config.headers["Authorization"] = "Bearer " + token;
  return config;
});
```

#### setResponseInterceptor(interceptor: function): void

**功能**: 设置全局响应拦截器
**参数**: `interceptor` (function) - 拦截器函数 `(response) => response`

#### STATUS_CODES 常量

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

### http/server - HTTP 服务器模块

#### createServer(): HTTPServer

**功能**: 创建 HTTP 服务器实例
**返回值**: HTTPServer 对象

#### HTTPServer 对象方法

##### listen(port: string|number, callback?: function): Promise<string>

**功能**: 启动 HTTP 服务器监听指定端口
**参数**:
- `port` (string|number) - 端口号
- `callback` (function, 可选) - 启动成功回调

**返回值**: Promise - 解析为启动成功消息

##### listenTLS(port: string|number, certFile: string, keyFile: string, callback?: function): Promise<string>

**功能**: 启动 HTTPS 服务器监听指定端口
**参数**:
- `port` (string|number) - 端口号
- `certFile` (string) - SSL 证书文件路径（.crt 或 .pem）
- `keyFile` (string) - SSL 私钥文件路径（.key）
- `callback` (function, 可选) - 启动成功回调

**返回值**: Promise - 解析为启动成功消息

**示例**:
```javascript
const { server } = require('http');
const app = server.createServer();

app.listenTLS("8443", "./certs/server.crt", "./certs/server.key")
  .then(() => console.log("HTTPS Server running"));
```

##### use(middleware: function): void

**功能**: 添加中间件
**参数**: `middleware` (function) - 中间件函数 `(req, res, next) => {}`

##### get(path: string, handler: function): void

**功能**: 添加 GET 路由
**参数**:
- `path` (string) - 路由路径
- `handler` (function) - 请求处理函数 `(req, res) => {}`

##### post(path: string, handler: function): void

**功能**: 添加 POST 路由

##### put(path: string, handler: function): void

**功能**: 添加 PUT 路由

##### delete(path: string, handler: function): void

**功能**: 添加 DELETE 路由

##### static(directory: string, urlPath?: string): void

**功能**: 设置静态文件服务
**参数**:
- `directory` (string) - 静态文件目录
- `urlPath` (string, 可选) - URL 路径前缀，默认 '/'

##### ws(path: string, handler: function): void

**功能**: 添加 WebSocket 路由
**参数**:
- `path` (string) - WebSocket 路由路径
- `handler` (function) - WebSocket 处理函数 `(ws) => {}`

##### close(): Promise<void>

**功能**: 关闭服务器
**返回值**: Promise

#### Request 对象（req）

```typescript
{
  method: string,          // HTTP 方法
  path: string,            // 请求路径
  url: string,             // 完整 URL
  headers: object,         // 请求头
  query: object,           // 查询参数
  body: string,            // 原始请求体
  json: any                // 自动解析的 JSON 数据
}
```

#### Response 对象（res）

##### status(code: number): Response

**功能**: 设置响应状态码
**参数**: `code` (number) - HTTP 状态码
**返回值**: Response 对象（链式调用）

##### header(name: string, value: string): Response

**功能**: 设置响应头
**参数**:
- `name` (string) - 响应头名称
- `value` (string) - 响应头值

**返回值**: Response 对象（链式调用）

##### send(data: string): Response

**功能**: 发送文本响应
**参数**: `data` (string) - 响应内容

##### json(data: any): Response

**功能**: 发送 JSON 响应
**参数**: `data` (any) - 响应数据（自动序列化）

##### html(data: string): Response

**功能**: 发送 HTML 响应
**参数**: `data` (string) - HTML 内容

##### sendFile(path: string): Response

**功能**: 发送文件（自动检测 MIME 类型）
**参数**: `path` (string) - 文件路径

##### download(path: string, filename?: string): Response

**功能**: 发送文件下载响应
**参数**:
- `path` (string) - 文件路径
- `filename` (string, 可选) - 下载文件名

##### redirect(url: string, code?: number): Response

**功能**: 重定向
**参数**:
- `url` (string) - 重定向 URL
- `code` (number, 可选) - 状态码，默认 302

#### WebSocket 对象（ws）

##### send(message: string): void

**功能**: 发送文本消息
**参数**: `message` (string) - 消息内容

##### sendJSON(data: any): void

**功能**: 发送 JSON 消息
**参数**: `data` (any) - 数据对象

##### on(event: string, handler: function): void

**功能**: 监听事件
**参数**:
- `event` (string) - 事件名称（'message', 'close', 'error'）
- `handler` (function) - 事件处理函数

##### close(): void

**功能**: 关闭连接

---

## db - 数据库模块

```javascript
const { redis, sqlite } = require('db');
// 或分别导入
const redis = require('db/redis');
const sqlite = require('db/sqlite');
```

---

### db/redis - Redis 客户端

#### createClient(config?: RedisConfig): RedisClient

**功能**: 创建 Redis 客户端
**参数**: 可选配置对象
**返回值**: RedisClient 对象

#### RedisConfig 对象

```typescript
{
  host?: string,           // 主机地址，默认 'localhost'
  port?: number,           // 端口，默认 6379
  password?: string,       // 密码
  db?: number              // 数据库编号，默认 0
}
```

#### RedisClient 对象方法（所有方法返回 Promise）

##### 字符串操作

###### set(key: string, value: string, expiration?: number): Promise<string>

**功能**: 设置键值
**参数**:
- `key` (string) - 键名
- `value` (string) - 值
- `expiration` (number, 可选) - 过期时间（秒）

**返回值**: Promise<string> - 'OK'

###### get(key: string): Promise<string|null>

**功能**: 获取键值
**参数**: `key` (string) - 键名
**返回值**: Promise<string|null>

###### setJSON(key: string, value: any, expiration?: number): Promise<string>

**功能**: 设置 JSON 数据
**参数**:
- `key` (string) - 键名
- `value` (any) - 数据对象（自动序列化）
- `expiration` (number, 可选) - 过期时间（秒）

###### getJSON(key: string): Promise<any|null>

**功能**: 获取 JSON 数据
**参数**: `key` (string) - 键名
**返回值**: Promise<any|null> - 自动反序列化的对象

###### del(key: string): Promise<number>

**功能**: 删除键
**参数**: `key` (string) - 键名
**返回值**: Promise<number> - 删除的键数量

###### exists(key: string): Promise<boolean>

**功能**: 检查键是否存在
**参数**: `key` (string) - 键名
**返回值**: Promise<boolean>

###### expire(key: string, seconds: number): Promise<boolean>

**功能**: 设置过期时间
**参数**:
- `key` (string) - 键名
- `seconds` (number) - 过期时间（秒）

**返回值**: Promise<boolean>

###### ttl(key: string): Promise<number>

**功能**: 获取剩余生存时间
**参数**: `key` (string) - 键名
**返回值**: Promise<number> - 剩余秒数，-1 表示永不过期，-2 表示不存在

##### 哈希操作

###### hset(key: string, field: string, value: string): Promise<number>

**功能**: 设置哈希字段

###### hget(key: string, field: string): Promise<string|null>

**功能**: 获取哈希字段

###### hgetall(key: string): Promise<object>

**功能**: 获取所有哈希字段
**返回值**: Promise<object> - 字段-值对象

###### hdel(key: string, field: string): Promise<number>

**功能**: 删除哈希字段

###### hexists(key: string, field: string): Promise<boolean>

**功能**: 检查哈希字段是否存在

###### hkeys(key: string): Promise<string[]>

**功能**: 获取所有哈希字段名

###### hvals(key: string): Promise<string[]>

**功能**: 获取所有哈希字段值

##### 列表操作

###### lpush(key: string, ...values: string[]): Promise<number>

**功能**: 从左侧推入元素
**返回值**: Promise<number> - 列表长度

###### rpush(key: string, ...values: string[]): Promise<number>

**功能**: 从右侧推入元素

###### lpop(key: string): Promise<string|null>

**功能**: 从左侧弹出元素

###### rpop(key: string): Promise<string|null>

**功能**: 从右侧弹出元素

###### lrange(key: string, start: number, stop: number): Promise<string[]>

**功能**: 获取列表范围元素
**参数**:
- `key` (string) - 键名
- `start` (number) - 起始索引
- `stop` (number) - 结束索引（-1 表示到末尾）

**返回值**: Promise<string[]>

###### llen(key: string): Promise<number>

**功能**: 获取列表长度

##### 集合操作

###### sadd(key: string, ...members: string[]): Promise<number>

**功能**: 添加集合成员
**返回值**: Promise<number> - 添加的成员数量

###### srem(key: string, ...members: string[]): Promise<number>

**功能**: 删除集合成员

###### smembers(key: string): Promise<string[]>

**功能**: 获取所有集合成员

###### sismember(key: string, member: string): Promise<boolean>

**功能**: 检查是否为集合成员

###### scard(key: string): Promise<number>

**功能**: 获取集合大小

##### 有序集合操作

###### zadd(key: string, score: number, member: string): Promise<number>

**功能**: 添加有序集合成员
**参数**:
- `key` (string) - 键名
- `score` (number) - 分数
- `member` (string) - 成员

###### zrange(key: string, start: number, stop: number, withScores?: boolean): Promise<any[]>

**功能**: 按索引范围获取成员
**参数**:
- `key` (string) - 键名
- `start` (number) - 起始索引
- `stop` (number) - 结束索引
- `withScores` (boolean, 可选) - 是否返回分数

###### zscore(key: string, member: string): Promise<number|null>

**功能**: 获取成员分数

###### zcard(key: string): Promise<number>

**功能**: 获取有序集合大小

##### 通用操作

###### ping(): Promise<string>

**功能**: 测试连接
**返回值**: Promise<string> - 'PONG'

###### close(): Promise<void>

**功能**: 关闭连接

---

### db/sqlite - SQLite 数据库

#### open(path: string): Promise<Database>

**功能**: 打开数据库连接
**参数**: `path` (string) - 数据库文件路径，':memory:' 表示内存数据库
**返回值**: Promise<Database>

#### Database 对象方法

##### exec(sql: string): Promise<void>

**功能**: 执行 SQL 语句（不返回结果）
**参数**: `sql` (string) - SQL 语句
**返回值**: Promise<void>

##### run(sql: string, params?: any[]): Promise<RunResult>

**功能**: 执行 SQL 语句（INSERT、UPDATE、DELETE）
**参数**:
- `sql` (string) - SQL 语句，支持 ? 占位符
- `params` (array, 可选) - 参数数组

**返回值**: Promise<RunResult> - `{lastInsertId, rowsAffected}`

##### get(sql: string, params?: any[]): Promise<object|null>

**功能**: 查询单条记录
**参数**:
- `sql` (string) - SQL 查询语句
- `params` (array, 可选) - 参数数组

**返回值**: Promise<object|null> - 记录对象

##### all(sql: string, params?: any[]): Promise<object[]>

**功能**: 查询多条记录
**参数**:
- `sql` (string) - SQL 查询语句
- `params` (array, 可选) - 参数数组

**返回值**: Promise<object[]> - 记录数组

##### prepare(sql: string): Promise<Statement>

**功能**: 创建预处理语句
**参数**: `sql` (string) - SQL 语句
**返回值**: Promise<Statement>

##### transaction(callback: function): Promise<void>

**功能**: 执行事务
**参数**: `callback` (function) - 事务函数 `async (tx) => {}`
**返回值**: Promise<void>

##### close(): Promise<void>

**功能**: 关闭数据库连接

##### tables(): Promise<string[]>

**功能**: 获取所有表名
**返回值**: Promise<string[]>

##### schema(tableName: string): Promise<object[]>

**功能**: 获取表结构
**参数**: `tableName` (string) - 表名
**返回值**: Promise<object[]> - 列信息数组

#### Statement 对象方法

##### run(params?: any[]): Promise<RunResult>

**功能**: 执行预处理语句

##### get(params?: any[]): Promise<object|null>

**功能**: 查询单条记录

##### all(params?: any[]): Promise<object[]>

**功能**: 查询多条记录

##### close(): Promise<void>

**功能**: 关闭预处理语句

---

## utils - 工具模块

```javascript
const { path, time, crypto, compression, util } = require('utils');
// 或分别导入
const path = require('utils/path');
const time = require('utils/time');
const crypto = require('utils/crypto');
const compression = require('utils/compression');
const util = require('utils/util');
```

---

### utils/path - 路径处理

#### join(...paths: string[]): string

**功能**: 连接多个路径片段
**参数**: 任意数量的路径字符串
**返回值**: 连接后的路径字符串

#### resolve(...paths: string[]): string

**功能**: 将路径解析为绝对路径
**参数**: 任意数量的路径字符串
**返回值**: 绝对路径字符串

#### dirname(path: string): string

**功能**: 获取路径的目录部分
**参数**: `path` (string) - 文件路径
**返回值**: 目录路径

#### basename(path: string, ext?: string): string

**功能**: 获取路径的基础文件名
**参数**:
- `path` (string) - 文件路径
- `ext` (string, 可选) - 要移除的扩展名

**返回值**: 文件名

#### extname(path: string): string

**功能**: 获取路径的扩展名
**参数**: `path` (string) - 文件路径
**返回值**: 扩展名（包含点）

#### isAbsolute(path: string): boolean

**功能**: 判断路径是否为绝对路径
**参数**: `path` (string) - 文件路径
**返回值**: true/false

#### normalize(path: string): string

**功能**: 规范化路径
**参数**: `path` (string) - 文件路径
**返回值**: 规范化后的路径

#### relative(from: string, to: string): string

**功能**: 计算从 from 到 to 的相对路径
**参数**:
- `from` (string) - 起始路径
- `to` (string) - 目标路径

**返回值**: 相对路径

#### 常量

- `sep`: 路径分隔符
- `delimiter`: 路径定界符

---

### utils/time - 时间处理

#### now(): string

**功能**: 获取当前时间（ISO 8601 格式）
**返回值**: string - ISO 8601 格式的时间字符串

#### nowUnix(): number

**功能**: 获取当前 Unix 时间戳（秒）
**返回值**: number - Unix 时间戳

#### nowUnixMilli(): number

**功能**: 获取当前 Unix 时间戳（毫秒）
**返回值**: number - 毫秒级 Unix 时间戳

#### nowUnixNano(): number

**功能**: 获取当前 Unix 时间戳（纳秒）
**返回值**: number - 纳秒级 Unix 时间戳

#### parse(timeStr: string, layout?: string): object

**功能**: 解析时间字符串
**参数**:
- `timeStr` (string) - 时间字符串
- `layout` (string, 可选) - 时间格式，默认 RFC3339

**返回值**: object - 包含 unix, iso, year, month, day, hour, minute, second, weekday

#### format(timestamp: number, layout?: string): string

**功能**: 格式化时间戳
**参数**:
- `timestamp` (number) - Unix 时间戳（秒）
- `layout` (string, 可选) - 时间格式，默认 RFC3339

**返回值**: string - 格式化后的时间字符串

#### sleep(seconds: number): Promise<void>

**功能**: 延迟执行（秒）
**参数**: `seconds` (number) - 延迟秒数
**返回值**: Promise<void>

#### sleepMillis(milliseconds: number): Promise<void>

**功能**: 延迟执行（毫秒）
**参数**: `milliseconds` (number) - 延迟毫秒数
**返回值**: Promise<void>

#### addDays(timestamp: number, days: number): number

**功能**: 添加天数
**返回值**: number - 新的 Unix 时间戳

#### addHours(timestamp: number, hours: number): number

**功能**: 添加小时
**返回值**: number - 新的 Unix 时间戳

#### addMinutes(timestamp: number, minutes: number): number

**功能**: 添加分钟
**返回值**: number - 新的 Unix 时间戳

#### addSeconds(timestamp: number, seconds: number): number

**功能**: 添加秒数
**返回值**: number - 新的 Unix 时间戳

#### isBefore(time1: number, time2: number): boolean

**功能**: 判断时间 1 是否在时间 2 之前
**返回值**: boolean

#### isAfter(time1: number, time2: number): boolean

**功能**: 判断时间 1 是否在时间 2 之后
**返回值**: boolean

#### diff(time1: number, time2: number): object

**功能**: 计算时间差
**返回值**: object - 包含 seconds, minutes, hours, days

#### utc(timestamp: number): string

**功能**: 转换为 UTC 时区
**返回值**: string - UTC 时间字符串

#### local(timestamp: number): string

**功能**: 转换为本地时区
**返回值**: string - 本地时间字符串

#### getYear(timestamp: number): number

**功能**: 获取年份

#### getMonth(timestamp: number): number

**功能**: 获取月份（1-12）

#### getDay(timestamp: number): number

**功能**: 获取日期（1-31）

#### getHour(timestamp: number): number

**功能**: 获取小时（0-23）

#### getMinute(timestamp: number): number

**功能**: 获取分钟（0-59）

#### getSecond(timestamp: number): number

**功能**: 获取秒数（0-59）

#### getWeekday(timestamp: number): object

**功能**: 获取星期几
**返回值**: object - 包含 number (0-6) 和 name (英文名称)

#### create(year: number, month: number, day: number, hour?: number, minute?: number, second?: number): number

**功能**: 创建时间
**参数**: 年、月、日、时、分、秒
**返回值**: number - Unix 时间戳

#### fromUnix(timestamp: number): object

**功能**: 从 Unix 时间戳创建时间对象
**返回值**: object - 包含 unix, iso, year, month, day, hour, minute, second, weekday

#### setInterval(callback: function, interval: number): number

**功能**: 设置周期性定时器
**参数**:
- `callback` (function) - 回调函数
- `interval` (number) - 时间间隔（毫秒）

**返回值**: number - 定时器 ID

#### clearInterval(timerId: number): void

**功能**: 清除定时器
**参数**: `timerId` (number) - 定时器 ID

#### createTicker(interval: number): Ticker

**功能**: 创建 Ticker 对象
**参数**: `interval` (number) - 时间间隔（毫秒）
**返回值**: Ticker 对象

#### FORMAT 常量

```javascript
time.FORMAT.RFC3339;     // "2006-01-02T15:04:05Z07:00"
time.FORMAT.RFC1123;     // "Mon, 02 Jan 2006 15:04:05 MST"
time.FORMAT.DateTime;    // "2006-01-02 15:04:05"
time.FORMAT.Date;        // "2006-01-02"
time.FORMAT.Time;        // "15:04:05"
time.FORMAT.Kitchen;     // "3:04PM"
```

#### UNIT 常量

```javascript
time.UNIT.MILLISECOND;   // 毫秒
time.UNIT.SECOND;        // 秒
time.UNIT.MINUTE;        // 分钟
time.UNIT.HOUR;          // 小时
```

---

### utils/crypto - 加密模块

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

#### aesEncrypt(data: string, key: string): string

**功能**: AES-256-GCM 加密
**参数**:
- `data` (string) - 待加密数据
- `key` (string) - 加密密钥（32字节）

**返回值**: Base64 编码的加密数据

#### aesDecrypt(data: string, key: string): string

**功能**: AES-256-GCM 解密
**参数**:
- `data` (string) - Base64 编码的加密数据
- `key` (string) - 解密密钥（32字节）

**返回值**: 解密后的原始数据

#### randomBytes(size?: number): string

**功能**: 生成安全随机字节
**参数**: `size` (number, 可选) - 字节数，默认 16
**返回值**: 十六进制编码的随机字节

---

### utils/compression - 压缩模块

#### gzipCompress(data: string): string

**功能**: Gzip 压缩
**参数**: `data` (string) - 原始数据
**返回值**: Base64 编码的压缩数据

#### gzipDecompress(data: string): string

**功能**: Gzip 解压
**参数**: `data` (string) - Base64 编码的压缩数据
**返回值**: 解压后的原始数据

#### zlibCompress(data: string): string

**功能**: Zlib 压缩
**参数**: `data` (string) - 原始数据
**返回值**: Base64 编码的压缩数据

#### zlibDecompress(data: string): string

**功能**: Zlib 解压
**参数**: `data` (string) - Base64 编码的压缩数据
**返回值**: 解压后的原始数据

---

### utils/util - 工具函数

#### format(...args: any[]): string

**功能**: 格式化字符串（类似 util.format）
**参数**: 任意参数
**返回值**: 格式化后的字符串

支持格式说明符：
- `%s` - 字符串
- `%d` - 数字
- `%j` - JSON
- `%%` - 百分号

#### inspect(value: any, options?: object): string

**功能**: 返回对象的字符串表示
**参数**:
- `value` (any) - 要检查的值
- `options` (object, 可选) - 选项

**返回值**: 字符串表示

#### isDeepStrictEqual(value1: any, value2: any): boolean

**功能**: 深度严格比较两个值
**参数**:
- `value1` (any) - 第一个值
- `value2` (any) - 第二个值

**返回值**: boolean - 是否相等

#### types

##### types.isDate(value: any): boolean

**功能**: 检查是否为 Date 对象

##### types.isRegExp(value: any): boolean

**功能**: 检查是否为 RegExp 对象

##### types.isPromise(value: any): boolean

**功能**: 检查是否为 Promise 对象

##### types.isMap(value: any): boolean

**功能**: 检查是否为 Map 对象

##### types.isSet(value: any): boolean

**功能**: 检查是否为 Set 对象

---

## net - 网络模块

```javascript
const { net, websocket, proxy } = require('net');
// 或分别导入
const net = require('net/net');
const websocket = require('net/websocket');
const proxy = require('net/proxy');
```

---

### net/net - TCP/UDP 网络

#### createTCPServer(): TCPServer

**功能**: 创建 TCP 服务器实例
**返回值**: TCPServer 对象

#### connectTCP(address: string, options?: ConnectOptions): Promise<TCPSocket>

**功能**: 连接到 TCP 服务器
**参数**:
- `address` (string) - 服务器地址，格式为 "host:port"
- `options` (object, 可选) - 连接选项

**返回值**: Promise<TCPSocket>

#### createUDPSocket(type?: string): UDPSocket

**功能**: 创建 UDP 套接字
**参数**: `type` (string, 可选) - 套接字类型，'udp4' 或 'udp6'，默认 'udp4'
**返回值**: UDPSocket 对象

#### TCPServer 对象方法

##### listen(port: string|number, callback?: function): Promise<string>

**功能**: 启动 TCP 服务器监听指定端口

##### on(event: string, handler: function): TCPServer

**功能**: 注册事件处理器
**事件**: `'connection'` - 新客户端连接

##### close(): Promise<void>

**功能**: 关闭 TCP 服务器

#### TCPSocket 对象

##### 属性

- `remoteAddress` (string): 远程地址
- `localAddress` (string): 本地地址

##### write(data: string): Promise<boolean>

**功能**: 发送数据

##### on(event: string, handler: function): TCPSocket

**功能**: 注册事件处理器
**事件**:
- `'data'`: 收到数据
- `'close'`: 连接关闭
- `'error'`: 发生错误

##### close(): void

**功能**: 关闭连接

##### setTimeout(timeout: number): TCPSocket

**功能**: 设置连接超时

#### UDPSocket 对象方法

##### bind(port: string, host?: string, callback?: function): Promise<string>

**功能**: 绑定 UDP 套接字到指定端口

##### send(data: string, port: string, host: string, callback?: function): Promise<boolean>

**功能**: 发送 UDP 数据包

##### on(event: string, handler: function): UDPSocket

**功能**: 注册事件处理器
**事件**:
- `'message'`: 收到消息
- `'close'`: 套接字关闭
- `'error'`: 发生错误

##### close(): void

**功能**: 关闭 UDP 套接字

##### address(): object|undefined

**功能**: 获取套接字地址信息
**返回值**: 地址对象 `{address, port, family}`

---

### net/websocket - WebSocket

#### connect(url: string, options?: ConnectOptions): Promise<WebSocketClient>

**功能**: 连接到 WebSocket 服务器
**参数**:
- `url` (string) - WebSocket URL（ws:// 或 wss://）
- `options` (object, 可选) - 连接选项

**返回值**: Promise<WebSocketClient>

#### ConnectOptions 对象

```typescript
{
  timeout?: number,        // 连接超时（毫秒），默认 10000
  headers?: object,        // 自定义 HTTP 请求头
  protocols?: string[]     // WebSocket 子协议
}
```

#### WebSocketClient 对象方法

##### send(message: string): void

**功能**: 发送文本消息

##### sendJSON(data: any): void

**功能**: 发送 JSON 消息

##### sendBinary(data: ArrayBuffer|Uint8Array): void

**功能**: 发送二进制消息

##### ping(data?: string): void

**功能**: 发送 ping 帧

##### close(code?: number, reason?: string): void

**功能**: 关闭连接

##### isClosed(): boolean

**功能**: 检查连接是否已关闭
**返回值**: true/false

##### on(event: string, handler: function): void

**功能**: 监听事件
**事件**:
- `'message'`: 收到消息
- `'close'`: 连接关闭
- `'error'`: 发生错误
- `'pong'`: 收到 pong 响应

---

### net/proxy - 代理模块

#### createHTTPProxy(targetURL: string): HTTPProxy

**功能**: 创建 HTTP/HTTPS 代理服务器
**参数**: `targetURL` (string) - 目标服务器 URL

#### createTCPProxy(target: string): TCPProxy

**功能**: 创建 TCP 代理服务器
**参数**: `target` (string) - 目标服务器地址

#### HTTPProxy 对象方法

##### on(event: string, handler: function): void

**功能**: 注册事件处理器
**事件**:
- `'request'`: 接收到请求
- `'response'`: 收到响应
- `'error'`: 代理错误

##### listen(port: string|number, callback?: function): Promise<string>

**功能**: 启动 HTTP 代理服务器

##### close(): Promise<string>

**功能**: 关闭 HTTP 代理服务器

#### TCPProxy 对象方法

##### on(event: string, handler: function): void

**功能**: 注册事件处理器
**事件**:
- `'connection'`: 新连接建立
- `'data'`: 数据传输
- `'close'`: 连接关闭
- `'error'`: 代理错误

##### listen(port: string|number, callback?: function): Promise<string>

**功能**: 启动 TCP 代理服务器

##### close(): Promise<string>

**功能**: 关闭 TCP 代理服务器

---

## fs - 文件系统模块

```javascript
const { fs, os } = require('fs');
// 或分别导入
const fs = require('fs/fs');
const os = require('fs/os');
```

---

### fs/fs - 文件操作

#### 同步方法

##### readFileSync(path: string, encoding?: string): string

**功能**: 同步读取文件
**参数**:
- `path` (string) - 文件路径
- `encoding` (string, 可选) - 编码格式，默认 'utf8'

**返回值**: 文件内容字符串

##### writeFileSync(path: string, data: string, encoding?: string): void

**功能**: 同步写入文件

##### existsSync(path: string): boolean

**功能**: 检查文件或目录是否存在

##### statSync(path: string): object

**功能**: 获取文件或目录信息
**返回值**: 包含 `isFile()`, `isDirectory()`, `size`, `modTime` 等的对象

##### mkdirSync(path: string, recursive?: boolean): void

**功能**: 同步创建目录

##### readdirSync(path: string): string[]

**功能**: 同步读取目录内容

##### unlinkSync(path: string): void

**功能**: 同步删除文件

##### rmdirSync(path: string, recursive?: boolean): void

**功能**: 同步删除目录

##### copyFileSync(src: string, dest: string): void

**功能**: 同步复制文件

##### renameSync(oldPath: string, newPath: string): void

**功能**: 同步重命名或移动文件

#### 异步方法（Promise）

所有同步方法都有对应的异步版本（去掉 Sync 后缀）：

- `readFile(path, encoding?): Promise<string>`
- `writeFile(path, data, encoding?): Promise<void>`
- `exists(path): Promise<boolean>`
- `stat(path): Promise<object>`
- `mkdir(path, recursive?): Promise<void>`
- `readdir(path): Promise<string[]>`
- `unlink(path): Promise<void>`
- `rmdir(path, recursive?): Promise<void>`
- `copyFile(src, dest): Promise<void>`
- `rename(oldPath, newPath): Promise<void>`

---

### fs/os - 操作系统信息

#### hostname(): string

**功能**: 获取主机名
**返回值**: 主机名字符串

#### homedir(): string

**功能**: 获取用户主目录
**返回值**: 主目录路径

#### tmpdir(): string

**功能**: 获取临时目录
**返回值**: 临时目录路径

#### arch(): string

**功能**: 获取 CPU 架构
**返回值**: 架构名称（如 'x64'）

#### platform(): string

**功能**: 获取操作系统平台
**返回值**: 平台名称（如 'win32', 'linux', 'darwin'）

#### uptime(): number

**功能**: 获取系统运行时间（秒）

#### totalmem(): number

**功能**: 获取总内存字节数

#### freemem(): number

**功能**: 获取空闲内存字节数

#### cpus(): object[]

**功能**: 获取 CPU 信息
**返回值**: CPU 对象数组，每个包含 `model`, `speed`, `times`

#### networkInterfaces(): object

**功能**: 获取网络接口信息
**返回值**: 网络接口对象

#### userInfo(): object

**功能**: 获取当前用户信息
**返回值**: 包含 `username`, `uid`, `gid`, `homedir` 的对象

#### type(): string

**功能**: 获取操作系统类型
**返回值**: 如 'Windows_NT', 'Linux', 'Darwin'

#### release(): string

**功能**: 获取内核版本
**返回值**: 内核版本字符串

---

## config - 配置模块

```javascript
const { viper } = require('config');
// 或导入
const viper = require('config/viper');
```

---

### config/viper - 配置管理

#### new(name?: string): ViperInstance

**功能**: 创建新的 Viper 实例
**参数**: `name` (string, 可选) - 实例名称，默认 'default'
**返回值**: ViperInstance

#### ViperInstance 对象方法

##### setConfigFile(configFile: string): void

**功能**: 设置配置文件路径

##### setConfigName(name: string): void

**功能**: 设置配置文件名（不含扩展名）

##### addConfigPath(path: string): void

**功能**: 添加搜索配置文件的路径

##### setConfigType(configType: string): void

**功能**: 设置配置类型（如 'yaml', 'json', 'toml' 等）

##### readInConfig(): void

**功能**: 读取配置文件

##### safeWriteConfig(): void

**功能**: 安全写入配置

##### get(key: string): any

**功能**: 获取配置值

##### getString(key: string): string

**功能**: 获取字符串配置值

##### getInt(key: string): number

**功能**: 获取整数配置值

##### getInt64(key: string): number

**功能**: 获取 64 位整数配置值

##### getFloat64(key: string): number

**功能**: 获取浮点数配置值

##### getBool(key: string): boolean

**功能**: 获取布尔配置值

##### getStringSlice(key: string): string[]

**功能**: 获取字符串数组配置值

##### set(key: string, value: any): void

**功能**: 设置配置值

##### setDefault(key: string, value: any): void

**功能**: 设置默认值

##### isSet(key: string): boolean

**功能**: 检查键是否已设置

##### allSettings(): object

**功能**: 获取所有配置

##### keys(): string[]

**功能**: 获取所有配置键

##### bindEnv(names: ...string): void

**功能**: 绑定环境变量

##### setEnvPrefix(prefix: string): void

**功能**: 设置环境变量前缀

##### unmarshal(obj: object): void

**功能**: 将配置解组到对象

##### unmarshalExact(obj: object): void

**功能**: 精确解组配置到对象

---

## process - 进程模块

```javascript
const { process, exec } = require('process');
// 或分别导入
const process = require('process/process');
const exec = require('process/exec');
// 向后兼容
const exec = require('exec');
```

---

### process/process - 进程信息与控制

`process` 对象提供有关当前进程的信息和控制。

#### 属性

- `pid` (number): 当前进程 ID
- `platform` (string): 操作系统平台
- `arch` (string): CPU 架构
- `versions` (object): 版本信息对象
  - `node`: 兼容性版本
  - `sw_runtime`: 运行时版本
  - `go`: Go 语言版本
- `env` (object): 环境变量对象
- `argv` (string[]): 命令行参数数组
- `stdout` (object): 标准输出流
- `stderr` (object): 标准错误流

### 方法

#### cwd(): string

获取当前工作目录。

#### chdir(path: string): void

改变当前工作目录。

#### exit(code?: number): void

退出进程。`code` 默认为 0。

#### uptime(): number

获取进程运行时间（秒）。

#### memoryUsage(): object

获取内存使用情况。返回对象包含 `rss`, `heapTotal`, `heapUsed` 等属性。

#### hrtime(time?: [number, number]): [number, number]

获取高精度时间。

#### kill(pid: number, signal?: string): boolean

发送信号给进程。

---

### process/exec - 进程执行

#### exec(command: string, args?: string[], options?: ExecOptions): Promise<object>

**功能**: 异步执行命令
**参数**:
- `command` (string) - 命令名称
- `args` (string[], 可选) - 命令参数
- `options` (object, 可选) - 执行选项

**返回值**: Promise<object> - 执行结果

#### execSync(command: string, args?: string[], options?: ExecOptions): object

**功能**: 同步执行命令

#### ExecOptions 对象

```typescript
{
  cwd?: string,            // 工作目录
  env?: object,            // 环境变量
  timeout?: number         // 超时时间（毫秒）
}
```

#### 执行结果对象

```typescript
{
  stdout: string,          // 标准输出
  stderr: string,          // 标准错误
  exitCode: number,        // 退出码
  success: boolean,        // 是否成功
  error: string|null,      // 错误信息
  command: string,         // 执行的命令
  args: string[],          // 命令参数
  timedOut?: boolean       // 是否超时
}
```

#### getEnv(key?: string, defaultValue?: string): any

**功能**: 获取环境变量

#### setEnv(key: string, value: string): boolean

**功能**: 设置环境变量

#### which(command: string): string|null

**功能**: 查找命令路径

#### commandExists(command: string): boolean

**功能**: 检查命令是否存在

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
- 建议使用 `utils/path` 模块规范化路径

### 数据序列化

- JSON 数据自动序列化和反序列化
- 支持嵌套对象和数组
- 循环引用会导致错误

---

## 示例代码

### 完整 Web 服务器示例

```javascript
const { server } = require('http');
const fs = require('fs');
const { redis } = require('db');

const app = server.createServer();
const redisClient = redis.createClient({ host: "localhost" });

// 中间件
app.use((req, res, next) => {
  console.log(`${req.method} ${req.path}`);
  next();
});

// API 路由
app.get("/api/users", async (req, res) => {
  const users = (await redisClient.getJSON("users")) || [];
  res.json({ users });
});

app.post("/api/users", async (req, res) => {
  const user = req.json;
  const users = (await redisClient.getJSON("users")) || [];
  users.push(user);
  await redisClient.setJSON("users", users);
  res.status(201).json({ message: "Created", user });
});

// 静态文件
app.static("./public", "/");

// 启动服务器
app.listen("3000").then(() => {
  console.log("Server running on http://localhost:3000");
});
```

### 数据库操作示例

```javascript
const { sqlite } = require('db');

async function main() {
  const db = await sqlite.open("./app.db");

  // 创建表
  await db.exec(`
    CREATE TABLE IF NOT EXISTS users (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL,
      email TEXT UNIQUE
    )
  `);

  // 插入数据
  const result = await db.run("INSERT INTO users (name, email) VALUES (?, ?)", [
    "Alice",
    "alice@example.com",
  ]);
  console.log("Inserted ID:", result.lastInsertId);

  // 查询数据
  const users = await db.all("SELECT * FROM users");
  console.log("Users:", users);

  // 使用事务
  await db.transaction(async (tx) => {
    await tx.run("INSERT INTO users (name, email) VALUES (?, ?)", [
      "Bob",
      "bob@example.com",
    ]);
    await tx.run("UPDATE users SET name = ? WHERE email = ?", [
      "Bobby",
      "bob@example.com",
    ]);
  });

  await db.close();
}

main().catch(console.error);
```

### 加密和压缩示例

```javascript
const { crypto, compression } = require('utils');

const data = "Hello, World!";
const key = "my-super-secret-key-32-bytes!!";

// 压缩数据
const compressed = compression.gzipCompress(data);
console.log('Compressed size:', compressed.length);

// 加密数据
const encrypted = crypto.aesEncrypt(compressed, key);
console.log('Encrypted:', encrypted.substring(0, 50) + '...');

// 解密
const decrypted = crypto.aesDecrypt(encrypted, key);
const original = compression.gzipDecompress(decrypted);
console.log('Original:', original);
```

### WebSocket 示例

```javascript
const { server } = require('http');
const { websocket } = require('net');

const app = server.createServer();

// WebSocket 路由
app.ws('/chat', (socket) => {
  console.log('新客户端连接');

  socket.on('message', (data) => {
    console.log('收到消息:', data);
    socket.send('服务器收到: ' + data);
  });

  socket.on('close', () => {
    console.log('连接关闭');
  });
});

app.listen('3000').then(() => {
  console.log('服务器启动在 http://localhost:3000');
});

// 客户端连接
websocket.connect('ws://localhost:3000/chat').then((client) => {
  console.log('已连接到服务器');

  client.on('message', (data) => {
    console.log('收到:', data);
  });

  client.send('Hello!');
});
```

---

本文档涵盖了 SW Runtime 的所有内置模块 API。所有接口均经过测试验证，可直接用于生产环境。
