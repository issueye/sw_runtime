# HTTP 客户端示例

本目录包含 http 客户端模块的功能演示。

## 文件说明

- **http-demo.ts** - HTTP 客户端完整演示

## 功能特点

### HTTP 方法
- GET 请求
- POST 请求
- PUT 请求
- DELETE 请求
- PATCH 请求
- HEAD 请求
- OPTIONS 请求

### 请求配置
- 自定义请求头
- URL 查询参数
- 请求超时设置
- 基本认证
- Bearer Token 认证

### 响应处理
- 自动 JSON 解析
- 状态码和状态文本
- 响应头访问
- Promise 支持

## 运行示例

```bash
sw_runtime run examples/05-http-client/http-demo.ts
```

## 示例代码

```javascript
const http = require('http');

// GET 请求
http.get('https://api.example.com/users')
  .then(response => {
    console.log('状态码:', response.status);
    console.log('数据:', response.data);
  });

// POST 请求
http.post('https://api.example.com/users', {
  data: { name: 'John', email: 'john@example.com' },
  headers: { 'Content-Type': 'application/json' }
}).then(response => console.log(response.data));

// 自定义客户端
const client = http.createClient({ timeout: 10 });
client.get('https://api.example.com/data');
```
