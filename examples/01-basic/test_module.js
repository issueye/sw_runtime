// test_module.js - 测试模块
const message = "Hello from test module!";

function greet(name) {
    return `${message} Hello, ${name}!`;
}

function add(a, b) {
    return a + b;
}

// CommonJS 导出
module.exports = {
    greet,
    add,
    message
};