// WebSocket 客户端示例
const ws = require('websocket');

console.log('正在连接到 WebSocket 服务器...');

// 连接到 WebSocket 服务器
ws.connect('ws://localhost:8080/chat', {
    timeout: 5000,
    headers: {
        'User-Agent': 'SW-Runtime-WebSocket-Client'
    }
}).then(client => {
    console.log('✅ 已连接到服务器');

    // 监听消息
    client.on('message', (data) => {
        console.log('收到消息:', data);
    });

    // 监听关闭事件
    client.on('close', () => {
        console.log('连接已关闭');
    });

    // 监听错误事件
    client.on('error', (err) => {
        console.error('WebSocket 错误:', err.message);
    });

    // 发送文本消息
    client.send('Hello from client!');

    // 发送 JSON 消息
    setTimeout(() => {
        client.sendJSON({
            type: 'greeting',
            message: 'Hello from SW Runtime!',
            timestamp: Date.now()
        });
    }, 1000);

    // 5 秒后关闭连接
    setTimeout(() => {
        console.log('准备关闭连接...');
        client.close();
    }, 5000);

}).catch(err => {
    console.error('连接失败:', err.message);
});
