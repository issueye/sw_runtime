// server-app.js - 使用内置模块的应用
const { server } = require('http');
const fs = require('fs');
const utils = require('./utils.js');

console.log('=== Server Application ===\n');

// 使用自定义模块
console.log('Testing custom module:');
console.log('  add(10, 20) =', utils.add(10, 20));

// 创建 HTTP 服务器（这些模块应该被排除）
const app = server.createServer();

app.get('/hello', (req, res) => {
    res.send(utils.greet('Server'));
});

app.get('/math', (req, res) => {
    const result = {
        sum: utils.add(5, 10),
        product: utils.multiply(3, 4)
    };
    res.json(result);
});

console.log('Server configured with routes:');
console.log('  GET /hello');
console.log('  GET /math');

app.listen('38200', () => {
    console.log('\n✓ Server ready on http://localhost:38200');
});
