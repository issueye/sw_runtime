// HTTP 代理服务器示例
const { proxy } = require('net');

console.log('=== HTTP Proxy Server Demo ===\n');

// 创建 HTTP 代理，将请求转发到目标服务器
const httpProxy = proxy.createHTTPProxy('https://httpbin.org');

// 监听请求事件
httpProxy.on('request', (req) => {
    console.log(`[Request] ${req.method} ${req.path}`);
    console.log(`  Host: ${req.host}`);
    console.log(`  Remote: ${req.remoteAddr}`);
});

// 监听响应事件
httpProxy.on('response', (resp) => {
    console.log(`[Response] ${resp.status} ${resp.statusText}`);
});

// 监听错误事件
httpProxy.on('error', (err) => {
    console.error(`[Error] ${err.message}`);
    console.error(`  URL: ${err.url}`);
});

// 启动代理服务器
httpProxy.listen('8888', () => {
    console.log('HTTP Proxy server is ready');
}).then(() => {
    console.log('\n🚀 HTTP Proxy Server is listening on http://localhost:8888');
    console.log('\n使用方法:');
    console.log('  curl -x http://localhost:8888 https://httpbin.org/get');
    console.log('  curl -x http://localhost:8888 https://httpbin.org/post -d "key=value"');
    console.log('\n或在浏览器中设置代理服务器为 localhost:8888');
}).catch(err => {
    console.error('Failed to start HTTP proxy server:', err.message);
});
