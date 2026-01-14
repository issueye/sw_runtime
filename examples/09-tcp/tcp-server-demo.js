// TCP 服务器示例
const { net } = require('net');

console.log('=== TCP Server Example ===\n');

// 创建 TCP 服务器
const server = net.createTCPServer();

server.on('connection', (socket) => {
    console.log('New client connected:', socket.remoteAddress);
    
    // 发送欢迎消息
    socket.write('Welcome to TCP Server!\n');
    
    // 接收数据
    socket.on('data', (data) => {
        console.log('Received from client:', data.trim());
        
        // 回显数据
        socket.write('Echo: ' + data);
        
        // 如果收到 'quit'，关闭连接
        if (data.trim() === 'quit') {
            console.log('Client requested to quit');
            socket.close();
        }
    });
    
    // 连接关闭
    socket.on('close', () => {
        console.log('Client disconnected:', socket.remoteAddress);
    });
    
    // 错误处理
    socket.on('error', (err) => {
        console.error('Socket error:', err.message);
    });
});

// 启动服务器
server.listen('8080', () => {
    console.log('TCP Server is listening on port 8080');
    console.log('Connect using: telnet localhost 8080\n');
}).then(() => {
    console.log('Server started successfully!');
}).catch(err => {
    console.error('Failed to start server:', err.message);
});

// 等待连接（保持服务器运行）
setTimeout(() => {
    console.log('Server is running... Press Ctrl+C to stop');
}, 1000);
