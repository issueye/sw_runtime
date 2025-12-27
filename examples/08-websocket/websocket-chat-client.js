// WebSocket 聊天客户端示例
// 需要先运行 websocket-demo.js 启动服务器

const ws = require('websocket');

const clientName = 'Client-' + Math.floor(Math.random() * 1000);
console.log(`${clientName} 正在连接到聊天服务器...`);

ws.connect('ws://localhost:3200/chat').then(client => {
    console.log(`✅ ${clientName} 已连接到聊天室`);

    // 监听消息
    client.on('message', (data) => {
        if (typeof data === 'string') {
            console.log('📩 收到:', data);
        } else {
            console.log('📩 收到 JSON:', JSON.stringify(data));
        }
    });

    // 监听连接关闭
    client.on('close', () => {
        console.log('❌ 连接已关闭');
    });

    // 监听错误
    client.on('error', (err) => {
        console.error('⚠️  错误:', err.message);
    });

    // 发送进入聊天室消息
    client.sendJSON({
        type: 'join',
        user: clientName
    });

    // 每 2 秒发送一条消息
    let msgCount = 0;
    const interval = setInterval(() => {
        msgCount++;
        client.sendJSON({
            type: 'message',
            user: clientName,
            text: `Hello ${msgCount} from ${clientName}!`
        });

        if (msgCount >= 5) {
            clearInterval(interval);
            // 发送离开消息
            client.sendJSON({
                type: 'leave',
                user: clientName
            });
            // 关闭连接
            setTimeout(() => {
                client.close();
            }, 500);
        }
    }, 2000);

}).catch(err => {
    console.error('连接失败:', err.message);
});
