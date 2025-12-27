# HTTP 客户端示例

本目录包含 http 客户端模块的功能演示。

## 文件说明

- **http-demo.ts** - HTTP 客户端完整演示
- **http-advanced-demo.js** - HTTP 高级功能演示（请求配置、参数修改）
- **http-interceptors-demo.js** - HTTP 拦截器演示（请求/响应拦截）

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

### 拦截器功能
- 全局请求拦截器（setRequestInterceptor）
- 全局响应拦截器（setResponseInterceptor）
- 单个请求拦截器（beforeRequest, afterResponse）
- 请求数据转换（transformRequest）
- 响应数据转换（transformResponse）

### 动态修改
- 修改请求头
- 修改请求参数
- 修改请求体
- 访问响应头
- 处理响应数据

### 响应处理
- 自动 JSON 解析
- 状态码和状态文本
- 响应头访问
- Promise 支持

## 运行示例

```bash
# 基础示例
sw_runtime run examples/05-http-client/http-demo.ts

# 高级功能示例
sw_runtime run examples/05-http-client/http-advanced-demo.js

# 拦截器示例
sw_runtime run examples/05-http-client/http-interceptors-demo.js
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

// 使用拦截器
http.setRequestInterceptor((config) => {
  // 所有请求自动添加 token
  config.headers['Authorization'] = 'Bearer ' + getToken();
  return config;
});

http.setResponseInterceptor((response) => {
  // 统一处理响应数据
  if (response.data.code === 0) {
    response.data = response.data.data;
  }
  return response;
});

// 单个请求的拦截器
http.post('https://api.example.com/users', {
  data: { username: 'alice' },
  beforeRequest: (config) => {
    config.headers['X-Request-ID'] = generateRequestId();
    return config;
  },
  transformRequest: (data) => {
    // 转换请求数据
    return { ...data, timestamp: Date.now() };
  },
  transformResponse: (data) => {
    // 转换响应数据
    return data.result;
  },
  afterResponse: (response) => {
    console.log('请求完成:', response.status);
    return response;
  }
});
```
