# 加密模块示例

本目录包含 crypto 加密模块的功能演示。

## 文件说明

- **crypto-demo.ts** - 加密模块完整演示

## 功能特点

### 哈希函数
- MD5 哈希
- SHA1 哈希
- SHA256 哈希
- SHA512 哈希

### 编解码
- Base64 编码/解码
- Hex 编码/解码

### 加密
- AES-256-GCM 加密/解密
- 安全随机数生成

## 运行示例

```bash
sw_runtime run examples/02-crypto/crypto-demo.ts
```

## 示例代码

```javascript
const crypto = require('crypto');

// 哈希
console.log(crypto.sha256('hello'));

// Base64
const encoded = crypto.base64Encode('hello');
const decoded = crypto.base64Decode(encoded);

// AES 加密
const encrypted = crypto.aesEncrypt('secret', 'key');
const decrypted = crypto.aesDecrypt(encrypted, 'key');

// 随机数
const random = crypto.randomBytes(16);
```
