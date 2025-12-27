// app.js - 主入口文件
const utils = require('./utils.js');
const mathLib = require('./math-lib.ts');

console.log('=== Bundle Test Application ===\n');

// 测试 utils 模块
console.log('1. Utils Module:');
console.log('   5 + 3 =', utils.add(5, 3));
console.log('   5 * 3 =', utils.multiply(5, 3));
console.log('   ', utils.greet('World'));

// 测试 math-lib 模块 (TypeScript)
console.log('\n2. Math Library (TypeScript):');
console.log('   square(4) =', mathLib.square(4));
console.log('   cube(3) =', mathLib.cube(3));
console.log('   PI =', mathLib.PI);

// 测试 Promise 支持
console.log('\n3. Async Support:');
Promise.resolve(42).then(value => {
    console.log('   Promise resolved with:', value);
});

// 测试 setTimeout
setTimeout(() => {
    console.log('   Timeout executed after 100ms');
}, 100);

console.log('\n=== Bundle Test Complete ===');
