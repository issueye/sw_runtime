// 热加载示例脚本
// 使用命令: sw_runtime run --watch example_hotreload.js

const { server } = require('http');

let requestCount = 0;
const startTime = new Date();

// 创建HTTP服务器
const app = server.createServer((req, res) => {
  requestCount++;

  res.writeHead(200, { 'Content-Type': 'application/json' });
  res.end(JSON.stringify({
    message: 'Hello from SW Runtime!',
    requestCount: requestCount,
    uptime: Math.floor((new Date() - startTime) / 1000) + ' seconds',
    timestamp: new Date().toISOString()
  }));
});

app.listen(3000, () => {
  console.log(`🚀 Server started at http://localhost:3000`);
  console.log(`👀 Watching for file changes... (修改此文件并保存以触发热重载)`);
  console.log(`📊 Initial request count: ${requestCount}`);
});

console.log('✅ 示例脚本已启动');