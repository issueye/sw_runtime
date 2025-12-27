// math.ts - TypeScript 测试模块
function multiply(a: number, b: number): number {
    return a * b;
}

function divide(a: number, b: number): number {
    if (b === 0) {
        throw new Error("Division by zero");
    }
    return a / b;
}

const PI = 3.14159;

// CommonJS 导出
module.exports = {
    multiply,
    divide,
    PI
};