// math-lib.ts - TypeScript 数学库
function square(n: number): number {
    return n * n;
}

function cube(n: number): number {
    return n * n * n;
}

const PI: number = 3.14159;

// 使用 CommonJS 导出
exports.square = square;
exports.cube = cube;
exports.PI = PI;
