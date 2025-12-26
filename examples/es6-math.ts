// es6-math.ts - 使用 ES6 import/export 语法的 TypeScript 模块

// 命名导出
export function add(a: number, b: number): number {
    return a + b;
}

export function subtract(a: number, b: number): number {
    return a - b;
}

export function multiply(a: number, b: number): number {
    return a * b;
}

export function divide(a: number, b: number): number {
    if (b === 0) {
        throw new Error("Division by zero");
    }
    return a / b;
}

// 常量导出
export const PI = 3.14159265359;
export const E = 2.71828182846;

// 类型定义
export interface MathOperation {
    name: string;
    operation: (a: number, b: number) => number;
}

// 类导出
export class Calculator {
    private history: string[] = [];

    add(a: number, b: number): number {
        const result = a + b;
        this.history.push(`${a} + ${b} = ${result}`);
        return result;
    }

    multiply(a: number, b: number): number {
        const result = a * b;
        this.history.push(`${a} * ${b} = ${result}`);
        return result;
    }

    getHistory(): string[] {
        return [...this.history];
    }

    clearHistory(): void {
        this.history = [];
    }
}

// 默认导出
const mathUtils = {
    add,
    subtract,
    multiply,
    divide,
    PI,
    E,
    Calculator
};

export default mathUtils;