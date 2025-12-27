// WebSocket 端到端集成测试
// 同时启动服务器和客户端

const server = require('httpserver');
const ws = require('websocket');

// 创建服务器
const app = server.createServer();

// 存储所有连接的客户端
const clients = [];

// WebSocket 服务器路由
app.ws('/echo', (wsServer) => {
    console.log('[服务器] 新客户端连接');
    clients.push(wsServer);
    
    wsServer.on('message', (data) => {
        console.log('[服务器] 收到消息:', data);
        // 回显消息
        wsServer.send('服务器回显: ' + data);
    });
    
    wsServer.on('close', () => {
        console.log('[服务器] 客户端断开连接');
        const index = clients.indexOf(wsServer);
        if (index > -1) {
            clients.splice(index, 1);
        }
    });
});

// 启动服务器
console.log('正在启动 WebSocket 服务器...');
app.listen('38400').then(() => {
    console.log('✅ 服务器已启动在端口 38400');
    
    // 等待服务器完全启动后,启动客户端
    setTimeout(() => {
        console.log('\n正在连接客户端...');
        
        // 连接到服务器
        ws.connect('ws://localhost:38400/echo').then(client => {
            console.log('✅ 客户端已连接');
            
            // 监听消息
            client.on('message', (data) => {
                console.log('[客户端] 收到回复:', data);
            });
            
            // 监听关闭
            client.on('close', () => {
                console.log('[客户端] 连接已关闭');
            });
            
            // 发送测试消息
            console.log('[客户端] 发送消息: Hello Server!');
            client.send('Hello Server!');
            
            // 1秒后发送第二条消息
            setTimeout(() => {
                console.log('[客户端] 发送消息: 这是第二条消息');
                client.send('这是第二条消息');
            }, 1000);
            
            // 2秒后发送 JSON 消息
            setTimeout(() => {
                console.log('[客户端] 发送 JSON 消息');
                client.sendJSON({
                    type: 'test',
                    message: 'JSON 数据',
                    timestamp: Date.now()
                });
            }, 2000);
            
            // 3秒后关闭连接
            setTimeout(() => {
                console.log('[客户端] 准备关闭连接');
                client.close();
            }, 3000);
            
        }).catch(err => {
            console.error('[客户端] 连接失败:', err.message);
        });
        
    }, 500);
});

console.log('\nWebSocket 端到端测试运行中...');
console.log('按 Ctrl+C 退出');
