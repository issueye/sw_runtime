// TCP 客户端示例
const { net } = require('net');

console.log('=== TCP Client Example ===\n');

// 连接到 TCP 服务器
net.connectTCP('localhost:8080', {
    timeout: 5000  // 5 秒超时
}).then(socket => {
    console.log('Connected to TCP server!');
    console.log('Local address:', socket.localAddress);
    console.log('Remote address:', socket.remoteAddress);
    console.log();
    
    // 接收数据
    socket.on('data', (data) => {
        console.log('Received:', data.trim());
    });
    
    // 连接关闭
    socket.on('close', () => {
        console.log('Connection closed');
    });
    
    // 错误处理
    socket.on('error', (err) => {
        console.error('Socket error:', err.message);
    });
    
    // 发送消息
    console.log('Sending messages...\n');
    
    setTimeout(() => {
        socket.write('Hello from client!\n');
    }, 500);
    
    setTimeout(() => {
        socket.write('This is a test message\n');
    }, 1000);
    
    setTimeout(() => {
        socket.write('How are you?\n');
    }, 1500);
    
    // 5 秒后关闭连接
    setTimeout(() => {
        console.log('\nSending quit command...');
        socket.write('quit\n');
        
        setTimeout(() => {
            socket.close();
        }, 500);
    }, 3000);
    
}).catch(err => {
    console.error('Failed to connect:', err.message);
});
