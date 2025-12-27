// WebSocket 简单示例
console.log('=== WebSocket 简单示例 ===\n');

const server = require('httpserver');
const app = server.createServer();

// 提供测试页面
app.get('/', (req, res) => {
    res.html(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>WebSocket 测试</title>
            <style>
                body { font-family: Arial; padding: 20px; }
                #messages { 
                    border: 1px solid #ccc; 
                    padding: 10px; 
                    height: 300px; 
                    overflow-y: auto; 
                    margin: 10px 0;
                }
            </style>
        </head>
        <body>
            <h1>WebSocket 测试</h1>
            <div id="messages"></div>
            <input type="text" id="input" placeholder="输入消息...">
            <button onclick="send()">发送</button>

            <script>
                const ws = new WebSocket('ws://' + window.location.host + '/ws');
                const messages = document.getElementById('messages');
                const input = document.getElementById('input');

                ws.onopen = () => {
                    addMessage('✅ 连接已建立');
                };

                ws.onmessage = (event) => {
                    addMessage('📨 收到: ' + event.data);
                };

                ws.onclose = () => {
                    addMessage('❌ 连接已关闭');
                };

                function send() {
                    const message = input.value;
                    if (message) {
                        ws.send(message);
                        addMessage('📤 发送: ' + message);
                        input.value = '';
                    }
                }

                function addMessage(text) {
                    messages.innerHTML += '<div>' + text + '</div>';
                    messages.scrollTop = messages.scrollHeight;
                }

                input.addEventListener('keypress', (e) => {
                    if (e.key === 'Enter') send();
                });
            </script>
        </body>
        </html>
    `);
});

// WebSocket 路由
app.ws('/ws', (ws) => {
    console.log('✅ 新的 WebSocket 连接');
    
    // 发送欢迎消息
    ws.send('欢迎连接到 WebSocket 服务器!');
    
    // 监听消息
    ws.on('message', (data) => {
        console.log('📨 收到消息:', data);
        
        // 回显消息
        ws.send('服务器收到: ' + data);
    });
    
    // 监听关闭
    ws.on('close', () => {
        console.log('❌ WebSocket 连接关闭');
    });
    
    // 监听错误
    ws.on('error', (error) => {
        console.log('⚠️  错误:', error.message);
    });
});

// 启动服务器
const PORT = 3201;
app.listen(PORT.toString(), () => {
    console.log('');
    console.log('🚀 WebSocket 服务器已启动');
    console.log('📖 访问: http://localhost:' + PORT);
    console.log('');
});
