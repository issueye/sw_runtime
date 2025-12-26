# SW Runtime 文档

这是 SW Runtime 的完整接口文档，采用模块化设计。

## 如何查看文档

### 方法 1: 使用 SW Runtime 文档服务器（推荐）

我们提供了专门为此文档开发的 HTTP 服务器，完美解决模块加载问题：

```bash
# Windows 用户
start-server.bat

# Linux/macOS 用户
./start-server.sh

# 或者直接运行
sw_runtime enhanced-doc-server.ts
```

**服务器特性：**
- ✅ 完美支持动态模块加载
- ✅ CORS 跨域支持
- ✅ 请求日志记录
- ✅ 安全防护（路径遍历、文件类型限制）
- ✅ API 端点支持
- ✅ 配置文件支持
- ✅ 优雅关闭处理

### 方法 2: 使用其他 HTTP 服务器

```bash
# Python 内置服务器
python -m http.server 8000

# Node.js http-server
npx http-server

# PHP 内置服务器
php -S localhost:8000
```

### 方法 3: 使用 VS Code Live Server

1. 安装 Live Server 扩展
2. 右键点击 `index.html`
3. 选择 "Open with Live Server"

## 服务器版本说明

### 基础版本 (doc-server.ts)
- 轻量级实现
- 核心功能完整
- 适合简单使用场景

### 增强版本 (enhanced-doc-server.ts) 🌟
- 支持配置文件 (`server-config.json`)
- 更丰富的 API 端点
- 增强的安全特性
- 开发工具支持
- 详细的日志记录

## 配置文件

增强版服务器支持通过 `server-config.json` 进行配置：

```json
{
    "server": {
        "port": 3000,
        "host": "localhost"
    },
    "features": {
        "cors": true,
        "logging": true
    },
    "security": {
        "preventPathTraversal": true,
        "allowedExtensions": [".html", ".css", ".js", ".json"]
    }
}
```

## API 端点

增强版服务器提供以下 API 端点：

- `GET /` - 文档首页
- `GET /modules/:name` - 动态加载模块
- `GET /api/modules` - 获取模块列表
- `GET /api/status` - 服务器状态
- `GET /api/config` - 配置信息
- `GET /health` - 健康检查

## 文档结构

```
docs/
├── index.html                    # 主文档页面
├── assets/
│   ├── css/styles.css           # 样式文件
│   └── js/app.js                # JavaScript 功能
├── modules/                     # 模块文档
│   ├── overview.html            # 运行时概述
│   ├── modules.html             # 模块系统
│   ├── crypto.html              # 加密模块
│   ├── compression.html         # 压缩模块
│   ├── fs.html                  # 文件系统
│   ├── http.html                # HTTP 客户端
│   ├── httpserver.html          # HTTP 服务器
│   ├── redis.html               # Redis 客户端
│   ├── sqlite.html              # SQLite 数据库
│   ├── path.html                # 路径操作
│   └── examples.html            # 完整示例
├── doc-server.ts                # 基础版文档服务器
├── enhanced-doc-server.ts       # 增强版文档服务器
├── server-config.json           # 服务器配置文件
├── start-server.bat             # Windows 启动脚本
├── start-server.sh              # Linux/macOS 启动脚本
└── README.md                    # 本文件
```

## 功能特性

- 📱 响应式设计，支持各种屏幕尺寸
- 🎨 专业的界面设计和语法高亮
- 🔍 左侧导航菜单，右侧内容展示
- ⚡ 动态模块加载，提高性能
- 📚 完整的 API 文档和示例代码
- 🛠️ 最佳实践和项目结构指南
- 🚀 专用文档服务器，完美支持所有功能

## 模块说明

- **运行时概述**: SW Runtime 的基本介绍和核心特性
- **模块系统**: CommonJS require() 和 ES6 import() 支持
- **加密模块**: AES 加密、哈希函数、随机数生成
- **压缩模块**: Gzip 和 Zlib 压缩/解压
- **文件系统**: 完整的文件和目录操作
- **HTTP 客户端**: GET/POST 请求、JSON 支持
- **HTTP 服务器**: 创建 Web 服务器、路由、中间件
- **Redis 客户端**: 连接 Redis、缓存操作
- **SQLite 数据库**: 数据库连接、SQL 查询
- **路径操作**: 跨平台路径处理
- **完整示例**: 综合应用示例和最佳实践

## 技术栈

- HTML5 + CSS3
- 原生 JavaScript (ES6+)
- SW Runtime HTTP 服务器
- 模块化架构
- 语法高亮
- 响应式设计