# Redis 客户端示例

本目录包含 Redis 数据库客户端的功能演示。

## 文件说明

- **redis-demo.ts** - Redis 客户端完整演示
- **config.json** - Redis 配置文件

## 功能特点

### 数据类型
- 字符串（String）
- 哈希（Hash）
- 列表（List）
- 集合（Set）
- 有序集合（Sorted Set）

### 操作
- 基本 CRUD 操作
- JSON 数据存储
- 过期时间设置
- 批量操作

## 运行示例

### 1. 启动 Redis 服务器
```bash
# 确保 Redis 服务器正在运行
redis-server
```

### 2. 运行示例
```bash
sw_runtime run examples/11-redis/redis-demo.ts
```

## 示例代码

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

## 配置说明

编辑 `config.json` 修改 Redis 连接配置：
```json
{
  "redis": {
    "host": "localhost",
    "port": 6379,
    "password": "",
    "db": 0
  }
}
```
