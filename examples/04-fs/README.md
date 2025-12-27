# 文件系统模块示例

本目录包含 fs 文件系统模块的功能演示。

## 文件说明

- **fs-demo.ts** - 文件系统模块完整演示
- **config.json** - 测试配置文件

## 功能特点

### 同步操作
- readFileSync - 读取文件
- writeFileSync - 写入文件
- existsSync - 检查存在
- statSync - 获取文件信息
- mkdirSync - 创建目录
- readdirSync - 读取目录
- unlinkSync - 删除文件
- rmdirSync - 删除目录
- copyFileSync - 复制文件
- renameSync - 重命名文件

### 异步操作
所有同步方法都有对应的异步版本（返回 Promise）

## 运行示例

```bash
sw_runtime run examples/04-fs/fs-demo.ts
```

## 示例代码

```javascript
const fs = require('fs');

// 同步操作
fs.writeFileSync('test.txt', 'Hello World');
const content = fs.readFileSync('test.txt', 'utf8');
console.log(content);

// 异步操作
fs.writeFile('test.txt', 'Hello Async')
  .then(() => fs.readFile('test.txt'))
  .then(content => console.log(content));

// 目录操作
fs.mkdirSync('mydir');
const files = fs.readdirSync('mydir');
```
