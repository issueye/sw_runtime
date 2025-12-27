// TCP 代理服务器示例
const proxy = require('proxy');

console.log('=== TCP Proxy Server Demo ===\n');

// 创建 TCP 代理，将连接转发到目标服务器
// 这里以转发到 Google DNS (8.8.8.8:53) 为例
const tcpProxy = proxy.createTCPProxy('8.8.8.8:53');

// 监听连接事件
tcpProxy.on('connection', (conn) => {
    console.log(`[Connection] New connection from ${conn.remoteAddr}`);
    console.log(`  Target: ${conn.target}`);
});

// 监听数据传输事件
tcpProxy.on('data', (data) => {
    console.log(`[Data] ${data.direction}: ${data.bytes} bytes transferred`);
});

// 监听关闭事件
tcpProxy.on('close', () => {
    console.log('[Close] Connection closed');
});

// 监听错误事件
tcpProxy.on('error', (err) => {
    console.error(`[Error] ${err.message}`);
    if (err.direction) {
        console.error(`  Direction: ${err.direction}`);
    }
});

// 启动代理服务器
tcpProxy.listen('5353', () => {
    console.log('TCP Proxy server is ready');
}).then(() => {
    console.log('\n🚀 TCP Proxy Server is listening on localhost:5353');
    console.log('\n使用方法:');
    console.log('  将任何 DNS 客户端配置为使用 localhost:5353');
    console.log('  例如: nslookup google.com localhost 5353');
    console.log('\n或者使用以下命令测试:');
    console.log('  nc localhost 5353');
}).catch(err => {
    console.error('Failed to start TCP proxy server:', err.message);
});
