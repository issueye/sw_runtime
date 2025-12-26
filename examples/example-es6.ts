// example-es6.ts - ES6 模块使用示例

// 使用动态 import 加载模块
async function demonstrateES6Modules() {
    console.log('=== ES6 模块演示 ===');
    
    try {
        // 动态导入 ES6 模块
        const mathModule = await global.import('./es6-math.ts');
        
        console.log('\n1. 命名导出使用:');
        console.log('   add(15, 25) =', mathModule.add(15, 25));
        console.log('   multiply(6, 7) =', mathModule.multiply(6, 7));
        console.log('   PI =', mathModule.PI);
        console.log('   E =', mathModule.E);
        
        console.log('\n2. 默认导出使用:');
        const defaultExport = mathModule.default;
        console.log('   default.add(100, 200) =', defaultExport.add(100, 200));
        console.log('   default.PI =', defaultExport.PI);
        
        console.log('\n3. 类的使用:');
        const Calculator = mathModule.Calculator;
        const calc = new Calculator();
        
        console.log('   创建计算器实例');
        console.log('   calc.add(50, 30) =', calc.add(50, 30));
        console.log('   calc.multiply(8, 9) =', calc.multiply(8, 9));
        console.log('   计算历史:', calc.getHistory());
        
        console.log('\n4. 链式导入:');
        const appModule = await global.import('./calculator-app.ts');
        console.log('   应用模块已加载');
        
        if (appModule.runCalculations) {
            console.log('   执行应用计算:');
            appModule.runCalculations();
        }
        
    } catch (error) {
        console.error('ES6 模块演示失败:', error.message);
    }
}

// 导出演示函数
export { demonstrateES6Modules };
export default demonstrateES6Modules;