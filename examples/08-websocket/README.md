# WebSocket 示例

本目录包含 WebSocket 客户端和服务器的功能演示。

## 文件说明

- **websocket-demo.js** - WebSocket 服务器演示
- **websocket-client-demo.js** - WebSocket 客户端演示
- **websocket-chat-client.js** - 聊天客户端示例
- **websocket-simple.js** - 简单 WebSocket 示例
- **websocket-test-simple.js** - 简单测试
- **websocket-e2e-test.js** - 端到端测试

## 功能特点

### 服务器端
- WebSocket 路由
- 消息广播
- 连接管理
- 事件驱动

### 客户端
- 连接到 WebSocket 服务器
- 发送/接收消息
- JSON 消息支持
- 二进制消息
- 自动重连

## 运行示例

### 启动服务器
```bash
sw_runtime run examples/08-websocket/websocket-demo.js
```

### 启动客户端
在另一个终端中：
```bash
sw_runtime run examples/08-websocket/websocket-client-demo.js
```

## 示例代码

### 服务器
```javascript
const server = require('httpserver');
const app = server.createServer();

app.ws('/chat', (ws) => {
  ws.on('message', (data) => {
    console.log('收到消息:', data);
    ws.send('回复: ' + data);
  });
  
  ws.on('close', () => {
    console.log('连接关闭');
  });
});

app.listen('8080');
```

### 客户端
```javascript
const ws = require('websocket');

ws.connect('ws://localhost:8080/chat')
  .then(client => {
    client.on('message', (data) => {
      console.log('收到:', data);
    });
    
    client.send('Hello!');
  });
```
