# SSL 证书生成指南

本目录包含用于 HTTPS 服务器演示的自签名 SSL 证书。

## 快速生成证书

在此目录下运行以下命令生成自签名证书：

### Windows (PowerShell)

```powershell
# 生成私钥
openssl genrsa -out server.key 2048

# 生成证书签名请求
openssl req -new -key server.key -out server.csr -subj "/C=CN/ST=Beijing/L=Beijing/O=SW-Runtime/OU=Dev/CN=localhost"

# 生成自签名证书（有效期 365 天）
openssl x509 -req -days 365 -in server.csr -signkey server.key -out server.crt

# 清理临时文件
Remove-Item server.csr
```

### Linux / macOS

```bash
# 一键生成证书和私钥
openssl req -x509 -newkey rsa:2048 -nodes -keyout server.key -out server.crt -days 365 \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=SW-Runtime/OU=Dev/CN=localhost"
```

## 证书文件说明

生成后，你将得到以下文件：

- `server.crt` - SSL 证书（公钥）
- `server.key` - 私钥
- `server.csr` - 证书签名请求（可删除）

## 使用证书

在你的 JavaScript 代码中这样使用：

```javascript
const server = require('httpserver');
const app = server.createServer();

app.get('/', (req, res) => {
    res.send('Hello HTTPS!');
});

// 使用 listenTLS 启动 HTTPS 服务器
app.listenTLS('8443', './examples/certs/server.crt', './examples/certs/server.key')
    .then(() => {
        console.log('HTTPS Server listening on https://localhost:8443');
    });
```

## 浏览器访问提示

由于这是自签名证书，浏览器会显示安全警告。这是正常的：

1. Chrome/Edge: 点击"高级" → "继续前往 localhost（不安全）"
2. Firefox: 点击"高级" → "接受风险并继续"
3. Safari: 点击"显示详细信息" → "访问此网站"

## 生产环境注意事项

⚠️ **重要**: 自签名证书仅用于开发和测试！

在生产环境中，应该使用由受信任的证书颁发机构（CA）签发的证书，例如：

- [Let's Encrypt](https://letsencrypt.org/) - 免费 SSL 证书
- [ZeroSSL](https://zerossl.com/) - 免费 SSL 证书
- 商业 CA（如 DigiCert、GlobalSign 等）

## 自动化脚本

如果没有安装 OpenSSL，可以下载：

- **Windows**: https://slproweb.com/products/Win32OpenSSL.html
- **Linux**: `sudo apt-get install openssl` 或 `sudo yum install openssl`
- **macOS**: 通常已预装，或使用 `brew install openssl`

## 证书信息查看

查看证书详细信息：

```bash
openssl x509 -in server.crt -text -noout
```

验证私钥和证书是否匹配：

```bash
openssl x509 -noout -modulus -in server.crt | openssl md5
openssl rsa -noout -modulus -in server.key | openssl md5
```

两个命令输出的 MD5 值应该相同。

## 故障排除

### 证书已过期

重新生成证书即可：

```bash
openssl req -x509 -newkey rsa:2048 -nodes -keyout server.key -out server.crt -days 365 \
  -subj "/C=CN/ST=Beijing/L=Beijing/O=SW-Runtime/OU=Dev/CN=localhost"
```

### 证书不匹配

确保使用同一对密钥和证书文件。

### 端口被占用

如果 8443 端口被占用，可以使用其他端口：

```javascript
app.listenTLS('9443', './examples/certs/server.crt', './examples/certs/server.key');
```
