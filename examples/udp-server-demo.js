// UDP 服务器示例
const net = require('net');

console.log('=== UDP Server Example ===\n');

// 创建 UDP 套接字
const socket = net.createUDPSocket('udp4');

// 接收消息
socket.on('message', (msg, rinfo) => {
    console.log('Received message from', rinfo.address + ':' + rinfo.port);
    console.log('Message:', msg.trim());
    console.log();
    
    // 回复客户端
    socket.send('Echo: ' + msg, rinfo.port.toString(), rinfo.address)
        .then(() => {
            console.log('Reply sent to client');
        })
        .catch(err => {
            console.error('Failed to send reply:', err.message);
        });
});

// 错误处理
socket.on('error', (err) => {
    console.error('Socket error:', err.message);
});

// 关闭事件
socket.on('close', () => {
    console.log('Socket closed');
});

// 绑定端口
socket.bind('9090', '0.0.0.0', () => {
    console.log('UDP Socket bound to port 9090');
}).then(result => {
    console.log(result);
    const addr = socket.address();
    console.log('Listening on', addr.address + ':' + addr.port);
    console.log('Send messages using: echo "Hello" | nc -u localhost 9090\n');
}).catch(err => {
    console.error('Failed to bind socket:', err.message);
});

// 保持服务器运行
setTimeout(() => {
    console.log('UDP Server is running... Press Ctrl+C to stop');
}, 1000);
