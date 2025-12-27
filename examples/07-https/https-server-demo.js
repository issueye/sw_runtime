// HTTPS 服务器示例
const server = require('httpserver');

console.log('=== HTTPS Server Example ===\n');

// 创建服务器
const app = server.createServer();

// 添加中间件
app.use((req, res, next) => {
    console.log(`${req.method} ${req.path}`);
    res.header('X-Powered-By', 'SW-Runtime');
    next();
});

// 添加路由
app.get('/', (req, res) => {
    res.html(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>HTTPS Server</title>
            <style>
                body {
                    font-family: Arial, sans-serif;
                    max-width: 800px;
                    margin: 50px auto;
                    padding: 20px;
                    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                    color: white;
                }
                h1 { color: #fff; }
                .info { background: rgba(255,255,255,0.1); padding: 20px; border-radius: 10px; }
            </style>
        </head>
        <body>
            <h1>🔐 Welcome to HTTPS Server!</h1>
            <div class="info">
                <h2>Server Information</h2>
                <p><strong>Protocol:</strong> HTTPS (Secure)</p>
                <p><strong>Port:</strong> 8443</p>
                <p><strong>Runtime:</strong> SW Runtime</p>
            </div>
        </body>
        </html>
    `);
});

app.get('/api/data', (req, res) => {
    res.json({
        message: 'This is a secure HTTPS response',
        timestamp: Date.now(),
        secure: true,
        data: {
            users: [
                { id: 1, name: 'Alice' },
                { id: 2, name: 'Bob' }
            ]
        }
    });
});

app.post('/api/login', (req, res) => {
    const credentials = req.json;
    console.log('Login attempt:', credentials);
    
    res.status(200).json({
        success: true,
        message: 'Login successful (demo)',
        token: 'secure_token_' + Date.now()
    });
});

// 启动 HTTPS 服务器
// 需要提供 SSL 证书和密钥文件
app.listenTLS('8443', './examples/certs/server.crt', './examples/certs/server.key', () => {
    console.log('HTTPS Server is ready');
}).then(() => {
    console.log('🔐 HTTPS Server is listening on https://localhost:8443');
    console.log('\nAvailable routes:');
    console.log('  - https://localhost:8443/');
    console.log('  - https://localhost:8443/api/data');
    console.log('  - https://localhost:8443/api/login (POST)');
    console.log('\n⚠️  Note: This uses a self-signed certificate.');
    console.log('    Your browser will show a security warning.');
    console.log('    Click "Advanced" and "Proceed to localhost" to continue.\n');
}).catch(err => {
    console.error('Failed to start HTTPS server:', err.message);
    console.error('\n💡 Make sure you have generated SSL certificates:');
    console.error('   See examples/certs/README.md for instructions\n');
});

// 保持服务器运行
setTimeout(() => {
    console.log('Server is running... Press Ctrl+C to stop');
}, 1000);
