// 测试相同路由不同方法的问题
const server = require('httpserver');

const app = server.createServer();

console.log('创建服务器成功');

// 注册相同路由的不同方法
app.get('/api/user', (req, res) => {
    console.log('GET /api/user');
    res.json({ method: 'GET', message: 'Get user info' });
});

app.post('/api/user', (req, res) => {
    console.log('POST /api/user');
    res.json({ method: 'POST', message: 'Create user' });
});

app.put('/api/user', (req, res) => {
    console.log('PUT /api/user');
    res.json({ method: 'PUT', message: 'Update user' });
});

app.delete('/api/user', (req, res) => {
    console.log('DELETE /api/user');
    res.json({ method: 'DELETE', message: 'Delete user' });
});

app.listen('3400', () => {
    console.log('服务器已启动在 http://localhost:3400');
    console.log('测试路由:');
    console.log('  GET    /api/user');
    console.log('  POST   /api/user');
    console.log('  PUT    /api/user');
    console.log('  DELETE /api/user');
});
