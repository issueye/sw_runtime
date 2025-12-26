// HTTP 服务器演示
const server = require('httpserver');

async function httpServerDemo() {
    console.log('=== HTTP 服务器演示 ===');

    try {
        // 1. 创建 HTTP 服务器
        console.log('\n1. 创建 HTTP 服务器:');
        const app = server.createServer();
        console.log('✓ HTTP 服务器创建成功');

        // 2. 添加中间件
        console.log('\n2. 添加中间件:');
        app.use((req, res, next) => {
            console.log(`[${new Date().toISOString()}] ${req.method} ${req.path}`);
            res.header('X-Powered-By', 'SW-Runtime');
            next();
        });
        console.log('✓ 日志中间件添加成功');

        // 3. 添加路由
        console.log('\n3. 添加路由:');

        // GET 根路径
        app.get('/', (req, res) => {
            res.html(`
                <!DOCTYPE html>
                <html>
                <head>
                    <title>SW Runtime HTTP Server</title>
                    <style>
                        body { font-family: Arial, sans-serif; margin: 40px; }
                        .container { max-width: 800px; margin: 0 auto; }
                        .endpoint { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 5px; }
                        .method { font-weight: bold; color: #007acc; }
                    </style>
                </head>
                <body>
                    <div class="container">
                        <h1>🚀 SW Runtime HTTP Server</h1>
                        <p>欢迎使用 SW Runtime 内置的 HTTP 服务器！</p>
                        
                        <h2>可用的 API 端点：</h2>
                        <div class="endpoint">
                            <span class="method">GET</span> /api/hello - 简单的问候接口
                        </div>
                        <div class="endpoint">
                            <span class="method">GET</span> /api/time - 获取服务器时间
                        </div>
                        <div class="endpoint">
                            <span class="method">POST</span> /api/echo - 回显请求数据
                        </div>
                        <div class="endpoint">
                            <span class="method">GET</span> /api/users/:id - 获取用户信息（示例）
                        </div>
                        <div class="endpoint">
                            <span class="method">GET</span> /api/status - 服务器状态
                        </div>
                        
                        <h2>测试命令：</h2>
                        <pre>
curl http://localhost:3000/api/hello
curl http://localhost:3000/api/time
curl -X POST http://localhost:3000/api/echo -H "Content-Type: application/json" -d '{"message":"Hello World"}'
curl http://localhost:3000/api/users/123
curl http://localhost:3000/api/status
                        </pre>
                    </div>
                </body>
                </html>
            `);
        });

        // API 路由
        app.get('/api/hello', (req, res) => {
            res.json({
                message: 'Hello from SW Runtime HTTP Server!',
                timestamp: new Date().toISOString(),
                method: req.method,
                path: req.path
            });
        });

        app.get('/api/time', (req, res) => {
            res.json({
                serverTime: new Date().toISOString(),
                timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
                timestamp: Date.now()
            });
        });

        app.post('/api/echo', (req, res) => {
            res.json({
                echo: req.json || req.body,
                headers: req.headers,
                method: req.method,
                receivedAt: new Date().toISOString()
            });
        });

        app.get('/api/users/:id', (req, res) => {
            const userId = req.params.id || 'unknown';
            res.json({
                id: userId,
                name: `User ${userId}`,
                email: `user${userId}@example.com`,
                createdAt: new Date().toISOString(),
                status: 'active'
            });
        });

        app.get('/api/status', (req, res) => {
            res.json({
                status: 'running',
                uptime: process.uptime ? process.uptime() : 'unknown',
                memory: process.memoryUsage ? process.memoryUsage() : 'unknown',
                version: '1.0.0',
                runtime: 'SW Runtime',
                timestamp: new Date().toISOString()
            });
        });

        // 错误处理路由
        app.get('/api/error', (req, res) => {
            res.status(500).json({
                error: 'Internal Server Error',
                message: 'This is a test error endpoint',
                timestamp: new Date().toISOString()
            });
        });

        // 重定向示例
        app.get('/redirect', (req, res) => {
            res.redirect('/');
        });

        // 404 处理（通用路由）
        app.get('*', (req, res) => {
            res.status(404).json({
                error: 'Not Found',
                message: `Path ${req.path} not found`,
                timestamp: new Date().toISOString()
            });
        });

        console.log('✓ 所有路由添加成功');

        // 4. 启动服务器
        console.log('\n4. 启动服务器:');
        await app.listen('3000', () => {
            console.log('✓ 服务器启动回调执行');
        });

        console.log('✓ HTTP 服务器已启动在 http://localhost:3000');
        console.log('✓ 可以通过浏览器或 curl 访问以下端点：');
        console.log('  - http://localhost:3000/ (主页)');
        console.log('  - http://localhost:3000/api/hello');
        console.log('  - http://localhost:3000/api/time');
        console.log('  - http://localhost:3000/api/status');
        console.log('  - http://localhost:3000/api/users/123');

        console.log('\n=== HTTP 服务器演示完成 ===');
        console.log('服务器将继续运行，按 Ctrl+C 停止...');

    } catch (error) {
        console.error('HTTP 服务器演示出错:', error);
    }
}

// Express 风格的服务器演示
async function expressStyleDemo() {
    console.log('\n=== Express 风格服务器演示 ===');

    try {
        const app = server.createServer();

        // 中间件：CORS
        app.use((req, res, next) => {
            res.header('Access-Control-Allow-Origin', '*');
            res.header('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
            res.header('Access-Control-Allow-Headers', 'Content-Type, Authorization');
            
            if (req.method === 'OPTIONS') {
                res.status(200).send('OK');
                return;
            }
            next();
        });

        // 中间件：JSON 解析（模拟）
        app.use((req, res, next) => {
            if (req.headers['content-type'] && req.headers['content-type'].includes('application/json')) {
                try {
                    if (req.body) {
                        req.json = JSON.parse(req.body);
                    }
                } catch (e) {
                    console.log('JSON 解析失败:', e.message);
                }
            }
            next();
        });

        // RESTful API 示例
        const users = [
            { id: 1, name: '张三', email: 'zhangsan@example.com' },
            { id: 2, name: '李四', email: 'lisi@example.com' },
            { id: 3, name: '王五', email: 'wangwu@example.com' }
        ];

        // 获取所有用户
        app.get('/users', (req, res) => {
            res.json({
                success: true,
                data: users,
                total: users.length
            });
        });

        // 获取单个用户
        app.get('/users/:id', (req, res) => {
            const id = parseInt(req.params.id);
            const user = users.find(u => u.id === id);
            
            if (user) {
                res.json({ success: true, data: user });
            } else {
                res.status(404).json({ success: false, error: 'User not found' });
            }
        });

        // 创建用户
        app.post('/users', (req, res) => {
            const newUser = {
                id: users.length + 1,
                name: req.json?.name || 'Unknown',
                email: req.json?.email || 'unknown@example.com'
            };
            users.push(newUser);
            
            res.status(201).json({
                success: true,
                data: newUser,
                message: 'User created successfully'
            });
        });

        // 更新用户
        app.put('/users/:id', (req, res) => {
            const id = parseInt(req.params.id);
            const userIndex = users.findIndex(u => u.id === id);
            
            if (userIndex !== -1) {
                users[userIndex] = { ...users[userIndex], ...req.json };
                res.json({
                    success: true,
                    data: users[userIndex],
                    message: 'User updated successfully'
                });
            } else {
                res.status(404).json({ success: false, error: 'User not found' });
            }
        });

        // 删除用户
        app.delete('/users/:id', (req, res) => {
            const id = parseInt(req.params.id);
            const userIndex = users.findIndex(u => u.id === id);
            
            if (userIndex !== -1) {
                const deletedUser = users.splice(userIndex, 1)[0];
                res.json({
                    success: true,
                    data: deletedUser,
                    message: 'User deleted successfully'
                });
            } else {
                res.status(404).json({ success: false, error: 'User not found' });
            }
        });

        await app.listen('3001');
        console.log('✓ Express 风格服务器已启动在 http://localhost:3001');

    } catch (error) {
        console.error('Express 风格服务器演示出错:', error);
    }
}

// 运行演示
httpServerDemo().then(() => {
    // 可以同时运行多个服务器
    // return expressStyleDemo();
});