# SQLite 数据库示例

本目录包含 SQLite 数据库的功能演示。

## 文件说明

- **sqlite-demo.ts** - SQLite 数据库完整演示
- **config.json** - 数据库配置文件

## 功能特点

### 数据库操作
- 打开/关闭数据库
- 创建表
- 插入数据
- 查询数据
- 更新数据
- 删除数据

### 高级功能
- 事务支持
- 预处理语句
- 参数绑定
- 批量操作
- 数据库信息查询

## 运行示例

```bash
sw_runtime run examples/12-sqlite/sqlite-demo.ts
```

## 示例代码

```javascript
const sqlite = require('sqlite');

// 打开数据库
const db = await sqlite.open('./database.db');
// 或内存数据库
// const db = await sqlite.open(':memory:');

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
const result = await db.run(
  'INSERT INTO users (name, email, age) VALUES (?, ?, ?)', 
  ['张三', 'zhangsan@example.com', 25]
);
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

// 关闭数据库
await db.close();
```

## 数据库文件

运行示例后会生成：
- `database.db` - SQLite 数据库文件
- 可使用 SQLite 客户端工具查看
