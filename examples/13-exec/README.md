# 进程执行示例

本目录包含 exec 进程执行模块的功能演示。

## 文件说明

- **exec-demo.js** - 进程执行完整演示

## 功能特点

### 命令执行
- 同步执行（execSync）
- 异步执行（exec）
- 命令超时控制
- 工作目录设置
- 环境变量设置

### 环境变量
- 获取环境变量
- 设置环境变量
- 查询所有环境变量

### 工具函数
- 查找命令路径（which）
- 检查命令是否存在（commandExists）

## 运行示例

```bash
sw_runtime run examples/13-exec/exec-demo.js
```

## 示例代码

### 同步执行
```javascript
const exec = require('exec');

// 执行命令
const result = exec.execSync('ls', ['-la'], {
  cwd: '/tmp',
  env: { PATH: process.env.PATH }
});

console.log('输出:', result.stdout);
console.log('错误:', result.stderr);
console.log('退出码:', result.exitCode);
console.log('成功:', result.success);
```

### 异步执行
```javascript
const exec = require('exec');

exec.exec('node', ['--version'], {
  timeout: 5000  // 5秒超时
}).then(result => {
  console.log('Node 版本:', result.stdout);
}).catch(err => {
  console.error('执行失败:', err.message);
});
```

### 环境变量
```javascript
const exec = require('exec');

// 获取环境变量
const path = exec.getEnv('PATH');
console.log('PATH:', path);

// 设置环境变量
exec.setEnv('MY_VAR', 'my_value');

// 获取所有环境变量
const allEnv = exec.getEnv();
console.log('所有环境变量:', allEnv);
```

### 工具函数
```javascript
const exec = require('exec');

// 查找命令路径
const nodePath = exec.which('node');
console.log('Node 路径:', nodePath);

// 检查命令是否存在
const hasGit = exec.commandExists('git');
console.log('Git 已安装:', hasGit);
```

## 注意事项

- 命令执行有安全风险，请谨慎使用
- 不要执行不受信任的命令
- 建议设置超时时间
- 检查命令执行结果的 `success` 字段
