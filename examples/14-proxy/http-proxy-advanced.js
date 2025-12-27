// HTTP 代理服务器高级示例 - 自定义请求和响应
const proxy = require('proxy');

console.log('=== Advanced HTTP Proxy Server Demo ===\n');

// 创建 HTTP 代理
const httpProxy = proxy.createHTTPProxy('https://api.github.com');

let requestCount = 0;
let errorCount = 0;

// 监听请求事件 - 记录和统计
httpProxy.on('request', (req) => {
    requestCount++;
    const timestamp = new Date().toISOString();
    
    console.log(`\n[${requestCount}] ${timestamp}`);
    console.log(`  ${req.method} ${req.path}`);
    console.log(`  Host: ${req.host}`);
    console.log(`  Remote: ${req.remoteAddr}`);
    
    // 记录请求头
    if (req.headers) {
        console.log('  Headers:');
        const headerKeys = Object.keys(req.headers);
        headerKeys.forEach(key => {
            if (key.toLowerCase() !== 'cookie') { // 不打印敏感信息
                console.log(`    ${key}: ${req.headers[key]}`);
            }
        });
    }
});

// 监听响应事件 - 记录状态和性能
httpProxy.on('response', (resp) => {
    console.log(`  ← ${resp.status} ${resp.statusText}`);
    
    // 记录响应头
    if (resp.headers) {
        const contentType = resp.headers['content-type'] || resp.headers['Content-Type'];
        const contentLength = resp.headers['content-length'] || resp.headers['Content-Length'];
        
        if (contentType) {
            console.log(`  Content-Type: ${contentType}`);
        }
        if (contentLength) {
            console.log(`  Content-Length: ${contentLength} bytes`);
        }
    }
});

// 监听错误事件 - 记录和统计错误
httpProxy.on('error', (err) => {
    errorCount++;
    console.error(`\n[ERROR #${errorCount}] ${err.message}`);
    console.error(`  URL: ${err.url}`);
});

// 启动代理服务器
httpProxy.listen('8080', () => {
    console.log('Advanced HTTP Proxy server is ready');
}).then(() => {
    console.log('\n🚀 Advanced HTTP Proxy Server is listening on http://localhost:8080');
    console.log('\n这个代理服务器会:');
    console.log('  ✓ 记录所有请求和响应');
    console.log('  ✓ 统计请求数量和错误数量');
    console.log('  ✓ 打印请求头和响应头');
    console.log('  ✓ 转发到 GitHub API');
    
    console.log('\n测试命令:');
    console.log('  curl -x http://localhost:8080 https://api.github.com');
    console.log('  curl -x http://localhost:8080 https://api.github.com/users/github');
    
    console.log('\n统计信息会实时显示...\n');
    
    // 定期打印统计信息
    setInterval(() => {
        if (requestCount > 0 || errorCount > 0) {
            console.log(`\n[Stats] Requests: ${requestCount}, Errors: ${errorCount}`);
        }
    }, 30000); // 每30秒
    
}).catch(err => {
    console.error('Failed to start HTTP proxy server:', err.message);
});
