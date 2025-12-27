# 打包功能快速开始

## 5 分钟上手 SW Runtime 打包

### 场景 1：简单项目打包

假设你有一个简单的项目：

```
my-app/
├── app.js
├── utils.js
└── config.json
```

**一键打包：**

```bash
sw_runtime bundle app.js
```

生成 `app.bundle.js`，包含所有依赖！

### 场景 2：TypeScript 项目打包

你的 TypeScript 项目：

```
my-ts-app/
├── main.ts
├── lib.ts
└── helpers.ts
```

**自动编译并打包：**

```bash
sw_runtime bundle main.ts -o dist/app.js
```

TypeScript 自动编译为 JavaScript，无需配置！

### 场景 3：生产环境打包

准备部署到生产环境？

```bash
# 压缩代码，减少 70%+ 体积
sw_runtime bundle app.js -o app.min.js --minify
```

**对比：**
- 开发版：`app.bundle.js` (2.41 KB)
- 生产版：`app.min.js` (0.86 KB) ⚡

### 场景 4：使用内置模块的应用

你的应用使用了 SW Runtime 的内置模块：

```javascript
// server.js
const httpserver = require('httpserver');  // 内置模块
const utils = require('./utils.js');        // 自定义模块

const app = httpserver.createServer();
app.get('/', (req, res) => {
    res.send(utils.greet('World'));
});

app.listen('3000');
```

**智能打包：**

```bash
sw_runtime bundle server.js -v
```

输出：
```
📦 包含模块: 2 个
  • server.js
  • utils.js
(httpserver 被自动排除)
```

内置模块不会被打包，保持文件小巧！

### 场景 5：调试打包后的代码

需要调试？生成 source map：

```bash
sw_runtime bundle app.ts --sourcemap
```

生成：
- `app.bundle.js` - 打包后的代码
- `app.bundle.js.map` - 调试映射文件

### 常用命令速查

```bash
# 基本打包
sw_runtime bundle <entry-file>

# 指定输出
sw_runtime bundle app.js -o dist/bundle.js

# 压缩代码
sw_runtime bundle app.js --minify

# 详细输出
sw_runtime bundle app.js -v

# 生成 source map
sw_runtime bundle app.ts --sourcemap

# 排除文件
sw_runtime bundle app.js --exclude test.js,debug.js
```

### 完整工作流示例

**开发阶段：**

```bash
# 直接运行原始文件
sw_runtime run src/app.ts

# 需要打包时使用详细模式
sw_runtime bundle src/app.ts -o dist/app.dev.js --sourcemap -v
sw_runtime run dist/app.dev.js
```

**生产部署：**

```bash
# 生成压缩版本
sw_runtime bundle src/app.ts -o dist/app.min.js --minify

# 部署 dist/app.min.js
# 运行方式不变
sw_runtime run dist/app.min.js
```

### 下一步

查看完整文档了解更多功能：

- [打包完整指南](BUNDLE_GUIDE.md) - 详细功能介绍
- [项目 README](../README.md) - 所有模块和 API

开始打包你的项目吧！🚀
