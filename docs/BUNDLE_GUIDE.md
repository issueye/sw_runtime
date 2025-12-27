# SW Runtime 打包指南

## 概述

SW Runtime 提供了强大的脚本打包功能，可以将多个 JavaScript/TypeScript 文件及其依赖打包成单个可执行文件。这对于部署、分发和优化应用程序非常有用。

## 核心特性

- ✅ **自动依赖解析** - 从入口文件递归分析所有 `require()` 依赖
- ✅ **TypeScript 支持** - 自动编译 `.ts` 文件为 JavaScript
- ✅ **内置模块排除** - 智能排除运行时可用的内置模块（fs, http, crypto 等）
- ✅ **代码压缩** - 可选的代码压缩优化，减少文件大小 70%+
- ✅ **Source Map** - 支持生成 source map 用于调试
- ✅ **文件排除** - 可以排除特定文件不参与打包

## 基本用法

### 简单打包

```bash
# 打包单个文件（自动生成输出文件名）
sw_runtime bundle app.js

# 生成 app.bundle.js
```

### 指定输出文件

```bash
# 使用 -o 或 --output 指定输出文件
sw_runtime bundle app.js -o dist/application.js
sw_runtime bundle server.ts --output build/server.bundle.js
```

### 代码压缩

```bash
# 使用 --minify 或 -m 压缩代码
sw_runtime bundle app.js -o app.min.js --minify

# 通常可以减少 70-80% 的文件大小
```

### 生成 Source Map

```bash
# 使用 --sourcemap 生成调试映射
sw_runtime bundle app.ts -o dist.js --sourcemap

# 会生成 dist.js 和 dist.js.map
```

### 详细输出

```bash
# 使用 -v 查看详细信息
sw_runtime bundle app.js -v

# 输出示例：
# 📦 正在打包: app.js
# 
# ✅ 打包完成!
# 
# 📄 输出文件: app.bundle.js
# 📊 文件大小: 2.41 KB
# 📦 包含模块: 3 个
# 
# 包含的模块:
#   • D:\project\app.js
#   • D:\project\utils.js
#   • D:\project\lib.ts
```

## 高级用法

### 排除特定文件

```bash
# 排除某些文件不参与打包
sw_runtime bundle app.js --exclude utils.js,helpers.js
```

### 组合多个选项

```bash
# 完整示例：压缩、sourcemap、详细输出
sw_runtime bundle src/main.ts \
  -o dist/app.min.js \
  --minify \
  --sourcemap \
  --verbose
```

## 工作原理

### 1. 依赖分析

打包器从入口文件开始，递归分析所有依赖：

```javascript
// app.js (入口文件)
const utils = require('./utils.js');
const lib = require('./lib.ts');
const config = require('./config.json');

// 打包器会自动找到并包含：
// - utils.js
// - lib.ts (并自动编译)
// - config.json
```

### 2. 内置模块处理

SW Runtime 的内置模块会被自动排除，因为它们在运行时可用：

```javascript
// 这些模块会被排除（不会打包到输出文件）
const fs = require('fs');
const http = require('http');
const crypto = require('crypto');
const httpserver = require('httpserver');
const websocket = require('websocket');

// 这些自定义模块会被打包
const myUtils = require('./my-utils.js');
const myLib = require('./my-lib.ts');
```

**完整的内置模块列表：**
- `server`, `sqlite`, `websocket`, `ws`
- `fs`, `crypto`, `zlib`, `compression`
- `http`, `redis`, `exec`, `child_process`
- `path`, `httpserver`

### 3. TypeScript 编译

TypeScript 文件会被自动编译为 JavaScript：

```typescript
// greeter.ts
function greet(name: string): string {
    return `Hello, ${name}!`;
}

exports.greet = greet;
```

打包时自动编译，无需额外配置。

### 4. 模块格式

打包器使用 CommonJS 格式，生成的代码结构：

```javascript
// 生成的 bundle.js
var __commonJS = (cb, mod) => function __require() {
  return mod || (0, cb[...])((...)), mod.exports;
};

// utils.js
var require_utils = __commonJS({
  "utils.js"(exports) {
    exports.add = function(a, b) { return a + b; };
  }
});

// app.js
var utils = require_utils();
console.log(utils.add(5, 3));
```

## 实际示例

### 示例 1：简单的多模块应用

**项目结构：**
```
my-app/
├── app.js       (入口)
├── utils.js     (工具函数)
└── math.ts      (数学库)
```

**utils.js:**
```javascript
exports.greet = function(name) {
    return `Hello, ${name}!`;
};
```

**math.ts:**
```typescript
function square(n: number): number {
    return n * n;
}
exports.square = square;
```

**app.js:**
```javascript
const utils = require('./utils.js');
const math = require('./math.ts');

console.log(utils.greet('World'));
console.log('5 squared =', math.square(5));
```

**打包命令：**
```bash
sw_runtime bundle app.js -o app.bundle.js -v
```

**输出：**
```
📦 正在打包: app.js

✅ 打包完成!

📄 输出文件: app.bundle.js
📊 文件大小: 1.48 KB
📦 包含模块: 3 个
```

**运行打包后的文件：**
```bash
sw_runtime run app.bundle.js
# Hello, World!
# 5 squared = 25
```

### 示例 2：使用内置模块的 HTTP 服务器

**server.js:**
```javascript
const httpserver = require('httpserver');  // 内置模块，会被排除
const utils = require('./utils.js');        // 自定义模块，会被打包

const app = httpserver.createServer();

app.get('/greet', (req, res) => {
    res.send(utils.greet('Server'));
});

app.listen('8080', () => {
    console.log('Server running on port 8080');
});
```

**打包：**
```bash
sw_runtime bundle server.js -o server.bundle.js -v

# 📦 包含模块: 2 个
#   • server.js
#   • utils.js
# (httpserver 被排除)
```

**运行：**
```bash
sw_runtime run server.bundle.js
# Server running on port 8080
```

### 示例 3：生产环境打包（压缩 + Sourcemap）

```bash
# 开发版本（便于调试）
sw_runtime bundle src/app.ts -o dist/app.dev.js --sourcemap

# 生产版本（最小化体积）
sw_runtime bundle src/app.ts -o dist/app.prod.js --minify

# 比较文件大小
# app.dev.js:  2.41 KB
# app.prod.js: 0.86 KB (减少 64%)
```

## 性能数据

基于实际测试的性能数据：

| 场景 | 原始大小 | 压缩后大小 | 减少比例 | 打包时间 |
|------|----------|------------|----------|----------|
| 简单应用 (3个模块) | 1.48 KB | 0.86 KB | 42% | ~20ms |
| TypeScript应用 | 2.41 KB | - | - | ~170ms |
| 复杂模块 | 754 bytes | 168 bytes | 77.7% | ~10ms |

## 命令行选项参考

```
sw_runtime bundle <entry-file> [flags]

选项:
  -o, --output string      输出文件路径 (默认: <entry>.bundle.js)
  -m, --minify            压缩输出代码
      --sourcemap         生成 source map
      --exclude strings   排除指定文件（逗号分隔）
  -v, --verbose           详细输出模式
  -q, --quiet             静默模式
  -h, --help              显示帮助信息
```

## 最佳实践

### 1. 开发环境

开发时使用详细输出和 sourcemap：

```bash
sw_runtime bundle app.ts -o dist.js --sourcemap -v
```

### 2. 生产环境

生产部署时使用压缩：

```bash
sw_runtime bundle app.ts -o app.min.js --minify -q
```

### 3. 自动化构建

在构建脚本中使用：

```bash
#!/bin/bash
# build.sh

echo "Building development version..."
sw_runtime bundle src/main.ts -o dist/app.dev.js --sourcemap

echo "Building production version..."
sw_runtime bundle src/main.ts -o dist/app.min.js --minify

echo "Build complete!"
```

### 4. TypeScript 项目

确保使用 CommonJS 导出风格：

```typescript
// ✅ 推荐 - CommonJS 风格
function myFunc() { }
exports.myFunc = myFunc;

// ❌ 避免 - ES6 module.exports 赋值
module.exports = { myFunc };  // 可能导致问题
```

### 5. 模块组织

将相关功能组织到独立模块中：

```
src/
├── main.ts          # 入口文件
├── utils/
│   ├── string.js    # 字符串工具
│   ├── math.js      # 数学工具
│   └── index.js     # 导出所有工具
├── services/
│   ├── api.ts       # API 服务
│   └── db.ts        # 数据库服务
└── config.json      # 配置文件
```

## 故障排查

### 问题：打包后运行失败

**解决方法：**
1. 检查是否使用了 CommonJS 导出格式（`exports.xxx` 而不是 `export`）
2. 验证所有依赖路径是否正确
3. 使用 `-v` 查看包含的模块列表

### 问题：文件大小过大

**解决方法：**
1. 使用 `--minify` 压缩代码
2. 使用 `--exclude` 排除不必要的文件
3. 检查是否意外包含了测试文件或示例

### 问题：内置模块未被排除

**解决方法：**
确保使用标准的模块名称：
```javascript
// ✅ 正确
const fs = require('fs');

// ❌ 错误（自定义路径）
const fs = require('./node_modules/fs');
```

## 总结

SW Runtime 的打包功能提供了：

- 🚀 **快速打包** - 毫秒级打包速度
- 📦 **智能优化** - 自动排除内置模块
- 🔧 **TypeScript 支持** - 无缝 TS 编译
- 📉 **体积优化** - 70%+ 的压缩率
- 🎯 **零配置** - 开箱即用

适用场景：
- 单文件部署
- 代码分发
- 性能优化
- 简化依赖管理

立即开始使用打包功能，让您的 JavaScript/TypeScript 应用更加高效！
