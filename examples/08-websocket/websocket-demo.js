// WebSocket 聊天室示例
console.log('=== WebSocket 聊天室示例 ===\n');

const server = require('httpserver');

const app = server.createServer();

// 存储所有连接的客户端
const clients = [];

// HTTP 路由 - 提供聊天界面
app.get('/', (req, res) => {
    res.html(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>WebSocket 聊天室</title>
            <style>
                body {
                    font-family: Arial, sans-serif;
                    max-width: 800px;
                    margin: 50px auto;
                    padding: 20px;
                }
                h1 { color: #333; }
                #messages {
                    height: 400px;
                    border: 1px solid #ddd;
                    padding: 10px;
                    overflow-y: auto;
                    background: #f9f9f9;
                    margin-bottom: 20px;
                }
                .message {
                    padding: 8px;
                    margin: 5px 0;
                    border-radius: 5px;
                    background: white;
                }
                .message.system {
                    background: #e3f2fd;
                    color: #1976d2;
                    font-style: italic;
                }
                .message.user {
                    background: #e8f5e9;
                }
                .message-time {
                    font-size: 0.8em;
                    color: #666;
                    margin-right: 8px;
                }
                .input-area {
                    display: flex;
                    gap: 10px;
                }
                #messageInput {
                    flex: 1;
                    padding: 10px;
                    border: 1px solid #ddd;
                    border-radius: 5px;
                    font-size: 14px;
                }
                button {
                    padding: 10px 20px;
                    background: #4CAF50;
                    color: white;
                    border: none;
                    border-radius: 5px;
                    cursor: pointer;
                    font-size: 14px;
                }
                button:hover {
                    background: #45a049;
                }
                .status {
                    padding: 10px;
                    margin-bottom: 20px;
                    border-radius: 5px;
                    text-align: center;
                }
                .status.connected {
                    background: #d4edda;
                    color: #155724;
                }
                .status.disconnected {
                    background: #f8d7da;
                    color: #721c24;
                }
            </style>
        </head>
        <body>
            <h1>🚀 WebSocket 聊天室</h1>
            <div id="status" class="status disconnected">未连接</div>
            <div id="messages"></div>
            <div class="input-area">
                <input type="text" id="messageInput" placeholder="输入消息..." disabled>
                <button id="sendBtn" onclick="sendMessage()" disabled>发送</button>
            </div>

            <script>
                let ws;
                const messagesDiv = document.getElementById('messages');
                const messageInput = document.getElementById('messageInput');
                const sendBtn = document.getElementById('sendBtn');
                const statusDiv = document.getElementById('status');

                function connect() {
                    ws = new WebSocket('ws://' + window.location.host + '/chat');

                    ws.onopen = function() {
                        console.log('WebSocket 连接已建立');
                        statusDiv.textContent = '✅ 已连接';
                        statusDiv.className = 'status connected';
                        messageInput.disabled = false;
                        sendBtn.disabled = false;
                        addMessage('系统消息: 已连接到聊天室', 'system');
                    };

                    ws.onmessage = function(event) {
                        try {
                            const data = JSON.parse(event.data);
                            if (data.type === 'message') {
                                addMessage(data.user + ': ' + data.text, 'user');
                            } else if (data.type === 'system') {
                                addMessage('系统消息: ' + data.text, 'system');
                            }
                        } catch (e) {
                            addMessage(event.data, 'user');
                        }
                    };

                    ws.onerror = function(error) {
                        console.error('WebSocket 错误:', error);
                        addMessage('系统消息: 连接错误', 'system');
                    };

                    ws.onclose = function() {
                        console.log('WebSocket 连接已关闭');
                        statusDiv.textContent = '❌ 未连接';
                        statusDiv.className = 'status disconnected';
                        messageInput.disabled = true;
                        sendBtn.disabled = true;
                        addMessage('系统消息: 已断开连接', 'system');
                    };
                }

                function addMessage(text, type = 'user') {
                    const messageDiv = document.createElement('div');
                    messageDiv.className = 'message ' + type;
                    
                    const time = new Date().toLocaleTimeString();
                    messageDiv.innerHTML = '<span class="message-time">' + time + '</span>' + text;
                    
                    messagesDiv.appendChild(messageDiv);
                    messagesDiv.scrollTop = messagesDiv.scrollHeight;
                }

                function sendMessage() {
                    const message = messageInput.value.trim();
                    if (message && ws && ws.readyState === WebSocket.OPEN) {
                        ws.send(JSON.stringify({
                            type: 'message',
                            text: message,
                            timestamp: new Date().toISOString()
                        }));
                        messageInput.value = '';
                    }
                }

                messageInput.addEventListener('keypress', function(e) {
                    if (e.key === 'Enter') {
                        sendMessage();
                    }
                });

                // 自动连接
                connect();
            </script>
        </body>
        </html>
    `);
});

// WebSocket 路由 - 聊天功能
app.ws('/chat', (ws) => {
    console.log('新客户端连接');
    
    // 添加到客户端列表
    clients.push(ws);
    
    // 广播用户加入消息
    broadcastMessage({
        type: 'system',
        text: '新用户加入聊天室 (当前在线: ' + clients.length + '人)',
        timestamp: new Date().toISOString()
    }, ws);
    
    // 监听消息
    ws.on('message', (data) => {
        console.log('收到消息:', data);
        
        // 广播消息给所有客户端
        broadcastMessage({
            type: 'message',
            user: '用户' + clients.indexOf(ws),
            text: data.text || data,
            timestamp: new Date().toISOString()
        });
    });
    
    // 监听错误
    ws.on('error', (error) => {
        console.log('WebSocket 错误:', error.message);
    });
    
    // 监听关闭
    ws.on('close', () => {
        console.log('客户端断开连接');
        
        // 从客户端列表移除
        const index = clients.indexOf(ws);
        if (index > -1) {
            clients.splice(index, 1);
        }
        
        // 广播用户离开消息
        broadcastMessage({
            type: 'system',
            text: '用户离开聊天室 (当前在线: ' + clients.length + '人)',
            timestamp: new Date().toISOString()
        });
    });
});

// 广播消息给所有客户端
function broadcastMessage(message, exclude) {
    clients.forEach(client => {
        if (client !== exclude) {
            client.sendJSON(message);
        }
    });
}

// API 端点 - 获取在线人数
app.get('/api/stats', (req, res) => {
    res.json({
        online: clients.length,
        timestamp: new Date().toISOString()
    });
});

// 启动服务器
const PORT = 3200;
app.listen(PORT.toString(), () => {
    console.log('');
    console.log('🚀 WebSocket 聊天室已启动！');
    console.log('📖 访问地址: http://localhost:' + PORT);
    console.log('🔌 WebSocket: ws://localhost:' + PORT + '/chat');
    console.log('');
    console.log('📋 功能说明:');
    console.log('   - 多用户实时聊天');
    console.log('   - 自动广播消息');
    console.log('   - 显示在线人数');
    console.log('   - 系统通知');
    console.log('');
    console.log('按 Ctrl+C 停止服务器');
});
