// 简单的HTTP服务器测试
const server = require('httpserver');

// 创建服务器
const app = server.createServer({
    readTimeout: 10,
    writeTimeout: 10
});

console.log('创建服务器成功');

// 添加简单路由
app.get('/test', (req, res) => {
    console.log('收到请求:', req.method, req.path);
    res.json({ message: 'Hello, World!', timestamp: new Date().toISOString() });
});

app.get('/ping', (req, res) => {
    res.send('pong');
});

// 启动服务器
app.listen('3200', () => {
    console.log('服务器已启动');
});

console.log('等待服务器启动在 http://localhost:3200');
