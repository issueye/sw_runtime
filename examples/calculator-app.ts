// calculator-app.ts - 使用 ES6 import 语法的应用
import { add, multiply, PI, Calculator } from './es6-math';
import mathUtils from './es6-math';

// 使用命名导入
console.log('Using named imports:');
console.log('add(5, 3):', add(5, 3));
console.log('multiply(4, 7):', multiply(4, 7));
console.log('PI:', PI);

// 使用默认导入
console.log('\nUsing default import:');
console.log('mathUtils.add(10, 20):', mathUtils.add(10, 20));
console.log('mathUtils.PI:', mathUtils.PI);

// 使用类
console.log('\nUsing Calculator class:');
const calc = new Calculator();
console.log('calc.add(15, 25):', calc.add(15, 25));
console.log('calc.multiply(6, 8):', calc.multiply(6, 8));
console.log('History:', calc.getHistory());

// 导出一些功能供其他模块使用
export function runCalculations(): void {
    console.log('Running calculations...');
    const results = {
        sum: add(100, 200),
        product: multiply(12, 12),
        pi: PI
    };
    console.log('Results:', results);
}

export { Calculator } from './es6-math';
export default { runCalculations };