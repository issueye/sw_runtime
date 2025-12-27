# SW Runtime 示例代码

本目录包含 SW Runtime 所有功能模块的示例代码，每个功能分类为一个独立的文件夹。

## 📁 目录结构

```
examples/
├── 01-basic/           # 基础示例（TypeScript、ES6、模块系统）
├── 02-crypto/          # 加密模块（哈希、编解码、AES）
├── 03-compression/     # 压缩模块（Gzip、Zlib）
├── 04-fs/              # 文件系统（读写、目录操作）
├── 05-http-client/     # HTTP 客户端（GET、POST、REST API）
├── 06-http-server/     # HTTP 服务器（路由、中间件、静态文件）
├── 07-https/           # HTTPS 服务器（SSL/TLS、证书）
├── 08-websocket/       # WebSocket（实时通信）
├── 09-tcp/             # TCP 网络（服务器、客户端）
├── 10-udp/             # UDP 网络（数据包收发）
├── 11-redis/           # Redis 客户端（缓存、数据库）
├── 12-sqlite/          # SQLite 数据库（SQL、事务）
├── 13-exec/            # 进程执行（命令、环境变量）
├── bundle-test/        # 打包测试示例
├── docs_server/        # 文档服务器
├── certs/              # SSL 证书工具（用于 HTTPS）
├── config.json         # 配置文件示例
├── package.json        # Node.js 包配置
└── comprehensive.bundle.js  # 综合打包示例
```

## 🚀 快速开始

### 1. 基础示例
```bash
# TypeScript 和 ES6 语法
sw_runtime run examples/01-basic/example-es6.ts

# 计算器应用
sw_runtime run examples/01-basic/calculator-app.ts

# 综合功能演示
sw_runtime run examples/01-basic/comprehensive-demo.ts
```

### 2. 加密和压缩
```bash
# 加密模块（MD5、SHA256、AES）
sw_runtime run examples/02-crypto/crypto-demo.ts

# 压缩模块（Gzip、Zlib）
sw_runtime run examples/03-compression/compression-demo.ts
```

### 3. 文件操作
```bash
# 文件系统操作
sw_runtime run examples/04-fs/fs-demo.ts
```

### 4. HTTP 客户端
```bash
# HTTP 请求示例
sw_runtime run examples/05-http-client/http-demo.ts
```

### 5. HTTP 服务器
```bash
# 基础 HTTP 服务器
sw_runtime run examples/06-http-server/httpserver-demo.ts

# 文件服务示例
sw_runtime run examples/06-http-server/httpserver-file-demo.js
```

### 6. HTTPS 服务器
```bash
# 首先生成 SSL 证书
cd examples/07-https/certs
.\generate-cert.ps1  # Windows
# 或
./generate-cert.sh   # Linux/macOS

# 运行 HTTPS 服务器
cd ../..
sw_runtime run examples/07-https/https-server-demo.js

# 混合 HTTP/HTTPS 服务器
sw_runtime run examples/07-https/https-mixed-demo.js
```

### 7. WebSocket
```bash
# 启动 WebSocket 服务器
sw_runtime run examples/08-websocket/websocket-demo.js

# 在另一个终端启动客户端
sw_runtime run examples/08-websocket/websocket-client-demo.js
```

### 8. TCP 网络
```bash
# 启动 TCP 服务器
sw_runtime run examples/09-tcp/tcp-server-demo.js

# 在另一个终端启动客户端
sw_runtime run examples/09-tcp/tcp-client-demo.js
```

### 9. UDP 网络
```bash
# 启动 UDP 服务器
sw_runtime run examples/10-udp/udp-server-demo.js

# 在另一个终端启动客户端
sw_runtime run examples/10-udp/udp-client-demo.js
```

### 10. Redis
```bash
# 确保 Redis 服务器正在运行
redis-server

# 运行 Redis 示例
sw_runtime run examples/11-redis/redis-demo.ts
```

### 11. SQLite
```bash
# SQLite 数据库示例
sw_runtime run examples/12-sqlite/sqlite-demo.ts
```

### 12. 进程执行
```bash
# 命令执行示例
sw_runtime run examples/13-exec/exec-demo.js
```

## 📖 详细说明

每个子目录都包含：
- 📄 **README.md** - 详细的功能说明和使用指南
- 📝 **demo 文件** - 可直接运行的示例代码
- ⚙️ **配置文件**（如需要）- 相关的配置文件

## 🔧 前置条件

### 基础运行
- SW Runtime 已安装并在 PATH 中

### 特定模块要求
- **Redis 示例**：需要 Redis 服务器运行
- **HTTPS 示例**：需要生成 SSL 证书
- **WebSocket/TCP/UDP**：需要开放相应端口

## 💡 使用技巧

### 1. TypeScript 支持
所有 `.ts` 文件都会自动编译，无需额外配置：
```bash
sw_runtime run examples/01-basic/calculator-app.ts
```

### 2. 模块导入
支持 CommonJS 和 ES6 模块：
```javascript
// CommonJS
const fs = require('fs');

// ES6 (在 .ts 文件中)
import { something } from './module';
```

### 3. JSON 配置
直接导入 JSON 文件：
```javascript
const config = require('./config.json');
```

### 4. 打包和压缩
```bash
# 打包 JavaScript 文件
sw_runtime bundle app.js -o app.bundle.js

# 压缩代码
sw_runtime bundle app.js -o app.min.js --minify

# 生成 source map
sw_runtime bundle app.js --sourcemap
```

## 🎯 学习路径

### 初学者
1. 从 [01-basic](./01-basic/) 开始学习基础语法
2. 了解 [04-fs](./04-fs/) 文件操作
3. 尝试 [05-http-client](./05-http-client/) HTTP 请求

### 中级
4. 学习 [06-http-server](./06-http-server/) 构建 Web 服务器
5. 探索 [08-websocket](./08-websocket/) 实时通信
6. 使用 [11-redis](./11-redis/) 和 [12-sqlite](./12-sqlite/) 数据库

### 高级
7. 掌握 [07-https](./07-https/) 安全通信
8. 学习 [09-tcp](./09-tcp/) 和 [10-udp](./10-udp/) 底层网络
9. 使用 [02-crypto](./02-crypto/) 加密和 [03-compression](./03-compression/) 压缩

## 🐛 常见问题

### 端口被占用
如果端口被占用，修改示例中的端口号：
```javascript
app.listen('3001'); // 改为其他端口
```

### Redis 连接失败
确保 Redis 服务器正在运行：
```bash
redis-server
```

### HTTPS 证书错误
重新生成证书：
```bash
cd examples/07-https/certs
.\generate-cert.ps1  # Windows
./generate-cert.sh   # Linux/macOS
```

### TypeScript 编译错误
检查 TypeScript 语法，SW Runtime 会自动处理 TypeScript 编译。

## 📚 更多资源

- [主 README](../README.md) - 项目总览
- [API 参考](../API_REFERENCE.md) - 完整 API 文档
- [打包指南](../docs/BUNDLE_GUIDE.md) - 代码打包说明

## 🤝 贡献

欢迎贡献更多示例！请确保：
- 代码简洁易懂
- 包含注释说明
- 提供 README 文档
- 可以独立运行

## 📝 许可

这些示例代码遵循项目的开源许可证。
