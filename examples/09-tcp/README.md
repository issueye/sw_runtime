# TCP 网络示例

本目录包含 TCP 服务器和客户端的功能演示。

## 文件说明

- **tcp-server-demo.js** - TCP 服务器演示
- **tcp-client-demo.js** - TCP 客户端演示

## 功能特点

### TCP 服务器
- 监听端口
- 接受连接
- 接收/发送数据
- 事件驱动
- 多客户端支持

### TCP 客户端
- 连接服务器
- 发送/接收数据
- 连接超时设置
- 自动重连

## 运行示例

### 启动服务器
```bash
sw_runtime run examples/09-tcp/tcp-server-demo.js
```

### 启动客户端
在另一个终端中：
```bash
sw_runtime run examples/09-tcp/tcp-client-demo.js
```

或使用 telnet 测试：
```bash
telnet localhost 8080
```

## 示例代码

### 服务器
```javascript
const net = require('net');
const server = net.createTCPServer();

server.on('connection', (socket) => {
  console.log('新客户端连接');
  
  socket.write('Welcome!\n');
  
  socket.on('data', (data) => {
    console.log('收到:', data);
    socket.write('Echo: ' + data);
  });
  
  socket.on('close', () => {
    console.log('客户端断开');
  });
});

server.listen('8080');
```

### 客户端
```javascript
const net = require('net');

net.connectTCP('localhost:8080', { timeout: 5000 })
  .then(socket => {
    socket.on('data', (data) => {
      console.log('收到:', data);
    });
    
    socket.write('Hello Server!\n');
  });
```
