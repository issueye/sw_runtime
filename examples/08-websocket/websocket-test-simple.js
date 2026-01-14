// 简单的 WebSocket 客户端测试
// 连接到公共 WebSocket 测试服务器

const { websocket } = require('net');

console.log('正在连接到 WebSocket 测试服务器...');

// 连接到 echo.websocket.org (公共测试服务器)
websocket.connect('wss://echo.websocket.org/', {
    timeout: 10000
}).then(client => {
    console.log('✅ 已连接到服务器');

    // 监听消息
    client.on('message', (data) => {
        console.log('📩 收到回显:', data);
        // 收到消息后关闭连接
        setTimeout(() => {
            client.close();
            console.log('连接已关闭');
        }, 100);
    });

    // 监听关闭事件
    client.on('close', () => {
        console.log('❌ 连接已断开');
    });

    // 监听错误事件
    client.on('error', (err) => {
        console.error('⚠️  错误:', err.message);
    });

    // 发送测试消息
    console.log('发送测试消息...');
    client.send('Hello from SW Runtime!');

}).catch(err => {
    console.error('❌ 连接失败:', err.message);
});
