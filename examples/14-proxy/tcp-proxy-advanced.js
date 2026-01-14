// TCP 代理服务器高级示例 - 统计和监控
const { proxy } = require('net');

console.log('=== Advanced TCP Proxy Server Demo ===\n');

// 创建 TCP 代理，转发到 Redis 服务器
const tcpProxy = proxy.createTCPProxy('localhost:6379');

let connectionCount = 0;
let activeConnections = 0;
let totalBytesTransferred = 0;
let errorCount = 0;

// 监听连接事件
tcpProxy.on('connection', (conn) => {
    connectionCount++;
    activeConnections++;
    
    const timestamp = new Date().toISOString();
    console.log(`\n[${connectionCount}] ${timestamp}`);
    console.log(`  New connection from ${conn.remoteAddr}`);
    console.log(`  Target: ${conn.target}`);
    console.log(`  Active connections: ${activeConnections}`);
});

// 监听数据传输事件
tcpProxy.on('data', (data) => {
    totalBytesTransferred += data.bytes;
    
    const direction = data.direction === 'client->target' ? '→' : '←';
    console.log(`  ${direction} ${data.bytes} bytes (Total: ${formatBytes(totalBytesTransferred)})`);
});

// 监听关闭事件
tcpProxy.on('close', () => {
    activeConnections--;
    console.log(`  Connection closed (Active: ${activeConnections})`);
});

// 监听错误事件
tcpProxy.on('error', (err) => {
    errorCount++;
    console.error(`\n[ERROR #${errorCount}] ${err.message}`);
    if (err.direction) {
        console.error(`  Direction: ${err.direction}`);
    }
});

// 格式化字节数
function formatBytes(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
}

// 启动代理服务器
tcpProxy.listen('6380', () => {
    console.log('Advanced TCP Proxy server is ready');
}).then(() => {
    console.log('\n🚀 Advanced TCP Proxy Server is listening on localhost:6380');
    console.log('\n这个代理服务器会:');
    console.log('  ✓ 转发到 Redis 服务器 (localhost:6379)');
    console.log('  ✓ 统计连接数和数据传输量');
    console.log('  ✓ 实时监控活跃连接');
    console.log('  ✓ 记录所有错误');
    
    console.log('\n使用方法:');
    console.log('  redis-cli -p 6380');
    console.log('\n或使用任何 Redis 客户端连接到 localhost:6380');
    
    console.log('\n统计信息会实时显示...\n');
    
    // 定期打印统计信息
    setInterval(() => {
        if (connectionCount > 0) {
            console.log(`\n[Stats] Total Connections: ${connectionCount}, Active: ${activeConnections}`);
            console.log(`[Stats] Total Data: ${formatBytes(totalBytesTransferred)}, Errors: ${errorCount}`);
        }
    }, 30000); // 每30秒
    
}).catch(err => {
    console.error('Failed to start TCP proxy server:', err.message);
});
