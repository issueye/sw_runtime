# HTTPS 服务器示例

本目录包含 HTTPS 安全服务器的功能演示。

## 文件说明

- **https-server-demo.js** - HTTPS 服务器演示
- **https-mixed-demo.js** - HTTP/HTTPS 混合服务器演示
- **certs/** - SSL 证书目录

## 功能特点

### HTTPS 支持
- SSL/TLS 加密通信
- 自定义证书和私钥
- 安全的 HTTPS 连接

### 证书管理
- 自签名证书生成
- 证书配置说明
- 自动化脚本

## 运行示例

### 1. 生成 SSL 证书

```bash
# Windows
cd examples/07-https/certs
.\generate-cert.ps1

# Linux/macOS
cd examples/07-https/certs
./generate-cert.sh
```

### 2. 运行服务器

```bash
# 纯 HTTPS 服务器
sw_runtime run examples/07-https/https-server-demo.js

# HTTP/HTTPS 混合服务器
sw_runtime run examples/07-https/https-mixed-demo.js
```

### 3. 访问服务器

在浏览器中访问：`https://localhost:8443`

⚠️ **注意**: 由于使用自签名证书，浏览器会显示安全警告，点击"高级"继续访问即可。

## 示例代码

```javascript
const server = require('httpserver');
const app = server.createServer();

app.get('/', (req, res) => {
  res.html('<h1>🔐 Secure HTTPS!</h1>');
});

// 启动 HTTPS 服务器
app.listenTLS('8443', './certs/server.crt', './certs/server.key')
  .then(() => {
    console.log('HTTPS Server running on https://localhost:8443');
  });
```

## 证书说明

- `server.crt` - SSL 证书（公钥）
- `server.key` - SSL 私钥
- 证书有效期：365 天
- 仅用于开发和测试

⚠️ **生产环境请使用 Let's Encrypt 或其他 CA 签发的证书**
