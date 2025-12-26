// doc-server.js - SW Runtime 文档服务器
console.log('=== SW Runtime 文档服务器 ===');

const server = require('httpserver');
const fs = require('fs');
const path = require('path');

const app = server.createServer();

// 中间件：请求日志
app.use((req, res, next) => {
    console.log(`${new Date().toISOString()} ${req.method} ${req.path}`);
    next();
});

// 中间件：CORS
app.use((req, res, next) => {
    res.header('Access-Control-Allow-Origin', '*');
    res.header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
    res.header('Access-Control-Allow-Headers', 'Content-Type');
    if (req.method === 'OPTIONS') {
        res.status(200).send('OK');
        return;
    }
    next();
});

// 根路径
app.get('/', (req, res) => {
    res.redirect('/index.html');
});

// 主文档页面
app.get('/index.html', (req, res) => {
    try {
        const filePath = path.join(__dirname, 'index.html');
        console.log('filePath', filePath);
        if (fs.existsSync(filePath)) {
            const content = fs.readFileSync(filePath, 'utf8');
            res.header('Content-Type', 'text/html; charset=utf-8');
            res.send(content);
        } else {
            res.html(getDefaultIndexHtml());
        }
    } catch (error) {
        res.html(getDefaultIndexHtml());
    }
});

// CSS 文件
app.get('/assets/css/styles.css', (req, res) => {
    try {
        const filePath = path.join(__dirname, 'assets', 'css', 'styles.css');
        if (fs.existsSync(filePath)) {
            const content = fs.readFileSync(filePath, 'utf8');
            res.header('Content-Type', 'text/css; charset=utf-8');
            res.send(content);
        } else {
            res.status(404).send('CSS file not found');
        }
    } catch (error) {
        res.status(500).send('Error reading CSS file');
    }
});

// JavaScript 文件
app.get('/assets/js/app.js', (req, res) => {
    try {
        const filePath = path.join(__dirname, 'assets', 'js', 'app.js');
        if (fs.existsSync(filePath)) {
            const content = fs.readFileSync(filePath, 'utf8');
            res.header('Content-Type', 'application/javascript; charset=utf-8');
            res.send(content);
        } else {
            res.status(404).send('JS file not found');
        }
    } catch (error) {
        res.status(500).send('Error reading JS file');
    }
});

// 模块文件 - 为每个模块创建单独的路由
const moduleNames = ['overview', 'modules', 'crypto', 'compression', 'fs', 'http', 'httpserver', 'redis', 'sqlite', 'path', 'examples'];

moduleNames.forEach(moduleName => {
    app.get('/modules/' + moduleName + '.html', (req, res) => {
        try {
            const filePath = path.join(__dirname, 'modules', moduleName + '.html');
            if (fs.existsSync(filePath)) {
                const content = fs.readFileSync(filePath, 'utf8');
                res.header('Content-Type', 'text/html; charset=utf-8');
                res.send(content);
                console.log('✓ 成功加载模块:', moduleName);
            } else {
                console.log('✗ 模块未找到:', moduleName);
                res.status(404).send('Module not found: ' + moduleName);
            }
        } catch (error) {
            res.status(500).send('Error loading module');
        }
    });
});

// API: 模块列表
app.get('/api/modules', (req, res) => {
    const modules = [
        { name: 'overview', title: '运行时概述' },
        { name: 'modules', title: '模块系统' },
        { name: 'crypto', title: '加密模块' },
        { name: 'compression', title: '压缩模块' },
        { name: 'fs', title: '文件系统' },
        { name: 'http', title: 'HTTP 客户端' },
        { name: 'httpserver', title: 'HTTP 服务器' },
        { name: 'redis', title: 'Redis 客户端' },
        { name: 'sqlite', title: 'SQLite 数据库' },
        { name: 'path', title: '路径操作' },
        { name: 'examples', title: '完整示例' }
    ];
    res.json({ success: true, modules: modules, count: modules.length });
});

// API: 服务器状态
app.get('/api/status', (req, res) => {
    res.json({
        status: 'running',
        name: 'SW Runtime 文档服务器',
        version: '1.0.0',
        timestamp: new Date().toISOString()
    });
});

// 健康检查
app.get('/health', (req, res) => {
    res.json({ status: 'healthy', timestamp: new Date().toISOString() });
});

// 默认首页 HTML
function getDefaultIndexHtml() {
    return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>SW Runtime - 文档服务器</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; line-height: 1.6; }
        .container { max-width: 800px; margin: 0 auto; }
        .success { background: #d4edda; color: #155724; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .api-list { background: #f8f9fa; padding: 20px; border-radius: 5px; }
        .api-item { padding: 10px; margin: 5px 0; background: white; border-radius: 3px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 SW Runtime 文档服务器</h1>
        <div class="success">✅ 服务器运行正常！</div>
        <h2>可用端点</h2>
        <div class="api-list">
            <div class="api-item"><strong>GET</strong> / - 文档首页</div>
            <div class="api-item"><strong>GET</strong> /modules/:name.html - 模块文档</div>
            <div class="api-item"><strong>GET</strong> /api/modules - 模块列表</div>
            <div class="api-item"><strong>GET</strong> /api/status - 服务器状态</div>
            <div class="api-item"><strong>GET</strong> /health - 健康检查</div>
        </div>
    </div>
</body>
</html>`;
}

// 启动服务器
const PORT = 3000;
console.log('正在启动服务器...');

app.listen(PORT.toString(), () => {
    console.log('');
    console.log('🚀 SW Runtime 文档服务器启动成功！');
    console.log('📖 访问地址: http://localhost:' + PORT);
    console.log('📁 文档根目录:', __dirname);
    console.log('');
    console.log('📋 可用端点:');
    console.log('   GET  /              - 文档首页');
    console.log('   GET  /modules/:name - 模块文档');
    console.log('   GET  /api/modules   - 模块列表');
    console.log('   GET  /api/status    - 服务器状态');
    console.log('   GET  /health        - 健康检查');
    console.log('');
    console.log('按 Ctrl+C 停止服务器');
});