# UDP 网络示例

本目录包含 UDP 服务器和客户端的功能演示。

## 文件说明

- **udp-server-demo.js** - UDP 服务器演示
- **udp-client-demo.js** - UDP 客户端演示

## 功能特点

### UDP 服务器
- 绑定端口
- 接收数据包
- 发送数据包
- 获取发送方信息
- 事件驱动

### UDP 客户端
- 发送数据包
- 接收回复
- 广播支持

## 运行示例

### 启动服务器
```bash
sw_runtime run examples/10-udp/udp-server-demo.js
```

### 启动客户端
在另一个终端中：
```bash
sw_runtime run examples/10-udp/udp-client-demo.js
```

或使用 netcat 测试：
```bash
echo "Hello" | nc -u localhost 9090
```

## 示例代码

### 服务器
```javascript
const net = require('net');
const socket = net.createUDPSocket('udp4');

socket.on('message', (msg, rinfo) => {
  console.log('收到来自', rinfo.address + ':' + rinfo.port);
  console.log('消息:', msg);
  
  // 回复客户端
  socket.send('Echo: ' + msg, rinfo.port.toString(), rinfo.address);
});

socket.bind('9090', '0.0.0.0')
  .then(() => {
    console.log('UDP 服务器监听端口 9090');
  });
```

### 客户端
```javascript
const net = require('net');
const socket = net.createUDPSocket('udp4');

// 发送消息
socket.send('Hello UDP!\n', '9090', 'localhost')
  .then(() => {
    console.log('消息已发送');
  });

// 接收回复
socket.bind('0', '0.0.0.0').then(() => {
  socket.on('message', (msg, rinfo) => {
    console.log('收到回复:', msg);
  });
});
```
