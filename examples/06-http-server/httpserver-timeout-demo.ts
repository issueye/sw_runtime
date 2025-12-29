// HTTP 服务器超时配置演示
const server = require('httpserver');

async function timeoutConfigDemo() {
    console.log('=== HTTP 服务器超时配置演示 ===\n');

    try {
        // 1. 创建带有自定义超时配置的服务器
        console.log('1. 创建带有自定义超时配置的服务器:');
        const app = server.createServer({
            readTimeout: 15,        // 读取超时：15秒（默认10秒）
            writeTimeout: 15,       // 写入超时：15秒（默认10秒）
            idleTimeout: 60,        // 空闲超时：60秒（默认120秒）
            readHeaderTimeout: 5,   // 读取请求头超时：5秒（默认10秒）
            maxHeaderBytes: 16384   // 最大请求头大小：16KB（默认8KB）
        });
        console.log('✓ 服务器创建成功，配置了自定义超时参数');

        // 2. 添加日志中间件
        console.log('\n2. 添加中间件:');
        app.use((req, res, next) => {
            console.log(`[${new Date().toISOString()}] ${req.method} ${req.path}`);
            next();
        });

        // 3. 添加测试路由
        console.log('\n3. 添加测试路由:');

        // 主页
        app.get('/', (req, res) => {
            res.html(`
                <!DOCTYPE html>
                <html>
                <head>
                    <title>HTTP 服务器超时配置演示</title>
                    <style>
                        body { font-family: Arial, sans-serif; margin: 40px; }
                        .container { max-width: 800px; margin: 0 auto; }
                        h1 { color: #007acc; }
                        .config { background: #f0f0f0; padding: 15px; border-radius: 5px; margin: 20px 0; }
                        .endpoint { background: #e8f4f8; padding: 10px; margin: 10px 0; border-radius: 3px; }
                        code { background: #f5f5f5; padding: 2px 6px; border-radius: 3px; }
                    </style>
                </head>
                <body>
                    <div class="container">
                        <h1>🚀 HTTP 服务器超时配置演示</h1>
                        
                        <h2>服务器配置：</h2>
                        <div class="config">
                            <p><strong>读取超时 (readTimeout):</strong> 15秒</p>
                            <p><strong>写入超时 (writeTimeout):</strong> 15秒</p>
                            <p><strong>空闲超时 (idleTimeout):</strong> 60秒</p>
                            <p><strong>读取请求头超时 (readHeaderTimeout):</strong> 5秒</p>
                            <p><strong>最大请求头大小 (maxHeaderBytes):</strong> 16KB</p>
                        </div>

                        <h2>测试端点：</h2>
                        <div class="endpoint">
                            <strong>GET /api/quick</strong> - 快速响应（立即返回）
                        </div>
                        <div class="endpoint">
                            <strong>GET /api/slow</strong> - 慢响应（延迟3秒）
                        </div>
                        <div class="endpoint">
                            <strong>GET /api/very-slow</strong> - 非常慢的响应（延迟10秒）
                        </div>
                        <div class="endpoint">
                            <strong>POST /api/echo</strong> - 回显请求数据
                        </div>

                        <h2>测试命令：</h2>
                        <pre>
# 快速响应
curl http://localhost:3100/api/quick

# 慢响应（3秒）
curl http://localhost:3100/api/slow

# 非常慢的响应（10秒）
curl http://localhost:3100/api/very-slow

# POST 请求
curl -X POST http://localhost:3100/api/echo \\
  -H "Content-Type: application/json" \\
  -d '{"message":"Hello World"}'
                        </pre>

                        <h2>超时说明：</h2>
                        <ul>
                            <li><code>readTimeout</code>: 读取整个请求的最大时间</li>
                            <li><code>writeTimeout</code>: 写入响应的最大时间</li>
                            <li><code>idleTimeout</code>: 保持连接空闲的最大时间（启用 keep-alive 时）</li>
                            <li><code>readHeaderTimeout</code>: 读取请求头的最大时间</li>
                            <li><code>maxHeaderBytes</code>: 请求头的最大字节数</li>
                        </ul>
                    </div>
                </body>
                </html>
            `);
        });

        // 快速响应端点
        app.get('/api/quick', (req, res) => {
            res.json({
                message: '快速响应',
                delay: 0,
                timestamp: new Date().toISOString()
            });
        });

        // 慢响应端点（3秒延迟）
        app.get('/api/slow', (req, res) => {
            console.log('开始处理慢响应请求...');
            setTimeout(() => {
                res.json({
                    message: '慢响应（3秒延迟）',
                    delay: 3,
                    timestamp: new Date().toISOString()
                });
                console.log('慢响应已发送');
            }, 3000);
        });

        // 非常慢的响应端点（10秒延迟）
        app.get('/api/very-slow', (req, res) => {
            console.log('开始处理非常慢的响应请求...');
            setTimeout(() => {
                res.json({
                    message: '非常慢的响应（10秒延迟）',
                    delay: 10,
                    timestamp: new Date().toISOString()
                });
                console.log('非常慢的响应已发送');
            }, 10000);
        });

        // 回显端点
        app.post('/api/echo', (req, res) => {
            res.json({
                echo: req.json || req.body,
                headers: req.headers,
                method: req.method,
                receivedAt: new Date().toISOString()
            });
        });

        // 状态端点
        app.get('/api/status', (req, res) => {
            res.json({
                status: 'running',
                config: {
                    readTimeout: '15s',
                    writeTimeout: '15s',
                    idleTimeout: '60s',
                    readHeaderTimeout: '5s',
                    maxHeaderBytes: '16KB'
                },
                timestamp: new Date().toISOString()
            });
        });

        console.log('✓ 所有路由添加成功');

        // 4. 启动服务器
        console.log('\n4. 启动服务器:');
        await app.listen('3100', () => {
            console.log('✓ 服务器启动回调执行');
        });

        console.log('✓ HTTP 服务器已启动在 http://localhost:3100');
        console.log('\n服务器配置信息:');
        console.log('  - 读取超时 (readTimeout): 15秒');
        console.log('  - 写入超时 (writeTimeout): 15秒');
        console.log('  - 空闲超时 (idleTimeout): 60秒');
        console.log('  - 读取请求头超时 (readHeaderTimeout): 5秒');
        console.log('  - 最大请求头大小 (maxHeaderBytes): 16KB');
        
        console.log('\n可访问的端点:');
        console.log('  - http://localhost:3100/ (主页)');
        console.log('  - http://localhost:3100/api/quick (快速响应)');
        console.log('  - http://localhost:3100/api/slow (3秒延迟)');
        console.log('  - http://localhost:3100/api/very-slow (10秒延迟)');
        console.log('  - http://localhost:3100/api/status (状态信息)');

        console.log('\n=== 超时配置演示完成 ===');
        console.log('服务器将继续运行，按 Ctrl+C 停止...');

    } catch (error) {
        console.error('超时配置演示出错:', error);
    }
}

// 默认配置服务器演示（对比）
async function defaultConfigDemo() {
    console.log('\n=== 默认配置服务器演示（对比） ===\n');

    try {
        // 创建使用默认配置的服务器
        const app = server.createServer();  // 不传入配置参数
        console.log('✓ 使用默认配置创建服务器');
        console.log('  默认配置:');
        console.log('  - readTimeout: 10秒');
        console.log('  - writeTimeout: 10秒');
        console.log('  - idleTimeout: 120秒');
        console.log('  - readHeaderTimeout: 10秒');
        console.log('  - maxHeaderBytes: 8KB');

        app.get('/', (req, res) => {
            res.json({
                message: '这是使用默认配置的服务器',
                config: 'default',
                timestamp: new Date().toISOString()
            });
        });

        await app.listen('3101');
        console.log('✓ 默认配置服务器已启动在 http://localhost:3101');

    } catch (error) {
        console.error('默认配置服务器演示出错:', error);
    }
}

// 运行演示
timeoutConfigDemo().then(() => {
    // 可选：同时运行默认配置服务器进行对比
    // return defaultConfigDemo();
});
