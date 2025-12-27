// comprehensive-demo.ts - 综合功能演示

console.log('=== SW Runtime 综合功能演示 ===');

// 导入所有模块
const http = require('http');
const redis = require('redis');
const crypto = require('crypto');
const compression = require('compression');
const fs = require('fs');
const path = require('path');

// 1. 数据获取和处理流水线
async function dataProcessingPipeline() {
    console.log('\n1. 数据处理流水线演示:');
    
    try {
        // 步骤1: 从 API 获取数据
        console.log('  - 从 API 获取数据...');
        const response = await http.get('https://jsonplaceholder.typicode.com/posts/1');
        const apiData = response.data;
        console.log('  - API 数据获取成功:', apiData.title);
        
        // 步骤2: 数据加密
        console.log('  - 加密数据...');
        const jsonData = JSON.stringify(apiData);
        const encryptionKey = 'my-secret-key-32-bytes-long!!!';
        const encryptedData = crypto.aesEncrypt(jsonData, encryptionKey);
        console.log('  - 数据加密完成，长度:', encryptedData.length);
        
        // 步骤3: 数据压缩
        console.log('  - 压缩加密数据...');
        const compressedData = compression.gzipCompress(encryptedData);
        console.log('  - 压缩完成，压缩率:', 
            ((1 - compressedData.length / encryptedData.length) * 100).toFixed(2) + '%');
        
        // 步骤4: 存储到文件
        console.log('  - 保存到文件...');
        const fileName = 'processed_data.bin';
        fs.writeFileSync(fileName, compressedData);
        console.log('  - 文件保存成功:', fileName);
        
        // 步骤5: 反向处理 - 读取和解密
        console.log('  - 读取和解密数据...');
        const readData = fs.readFileSync(fileName, 'utf8');
        const decompressedData = compression.gzipDecompress(readData);
        const decryptedData = crypto.aesDecrypt(decompressedData, encryptionKey);
        const originalData = JSON.parse(decryptedData);
        
        console.log('  - 数据恢复成功:', originalData.title);
        console.log('  - 数据完整性验证:', 
            JSON.stringify(originalData) === JSON.stringify(apiData) ? '✓' : '✗');
        
        // 清理文件
        fs.unlinkSync(fileName);
        console.log('  - 临时文件已清理');
        
    } catch (error) {
        console.error('  - 数据处理流水线错误:', error.message);
    }
}

// 2. 缓存系统演示
async function cacheSystemDemo() {
    console.log('\n2. 缓存系统演示:');
    
    try {
        // 创建 Redis 连接
        const client = redis.createClient({
            host: 'localhost',
            port: 6379
        });
        
        console.log('  - Redis 连接成功');
        
        // 模拟数据获取函数
        async function fetchUserData(userId: number) {
            const cacheKey = `user:${userId}`;
            
            // 先检查缓存
            const cached = await client.getJSON(cacheKey);
            if (cached) {
                console.log('  - 从缓存获取用户数据:', cached.name);
                return cached;
            }
            
            // 缓存未命中，从 API 获取
            console.log('  - 缓存未命中，从 API 获取数据...');
            const response = await http.get(`https://jsonplaceholder.typicode.com/users/${userId}`);
            const userData = response.data;
            
            // 存储到缓存，5分钟过期
            await client.setJSON(cacheKey, userData, 300);
            console.log('  - 数据已缓存:', userData.name);
            
            return userData;
        }
        
        // 测试缓存系统
        const user1 = await fetchUserData(1);
        const user1Cached = await fetchUserData(1); // 应该从缓存获取
        
        console.log('  - 缓存系统测试完成');
        
    } catch (error) {
        console.log('  - Redis 不可用 (这是正常的，如果没有运行 Redis 服务器)');
        console.log('  - 错误:', error.message);
    }
}

// 3. 文件处理和路径操作
async function fileProcessingDemo() {
    console.log('\n3. 文件处理演示:');
    
    try {
        // 创建测试目录结构
        const testDir = 'test_workspace';
        const dataDir = path.join(testDir, 'data');
        const outputDir = path.join(testDir, 'output');
        
        console.log('  - 创建目录结构...');
        fs.mkdirSync(testDir, { recursive: true });
        fs.mkdirSync(dataDir, { recursive: true });
        fs.mkdirSync(outputDir, { recursive: true });
        
        // 创建测试数据
        const testData = {
            timestamp: new Date().toISOString(),
            data: Array.from({length: 100}, (_, i) => ({
                id: i + 1,
                value: Math.random() * 1000,
                category: ['A', 'B', 'C'][i % 3]
            }))
        };
        
        // 保存原始数据
        const originalFile = path.join(dataDir, 'original.json');
        fs.writeFileSync(originalFile, JSON.stringify(testData, null, 2));
        console.log('  - 原始数据已保存:', originalFile);
        
        // 压缩数据
        const compressedData = compression.gzipCompress(JSON.stringify(testData));
        const compressedFile = path.join(outputDir, 'compressed.gz');
        fs.writeFileSync(compressedFile, compressedData);
        console.log('  - 压缩数据已保存:', compressedFile);
        
        // 加密数据
        const encryptedData = crypto.aesEncrypt(JSON.stringify(testData), 'encryption-key-32-bytes-long!');
        const encryptedFile = path.join(outputDir, 'encrypted.bin');
        fs.writeFileSync(encryptedFile, encryptedData);
        console.log('  - 加密数据已保存:', encryptedFile);
        
        // 文件信息统计
        const originalStat = fs.statSync(originalFile);
        const compressedStat = fs.statSync(compressedFile);
        const encryptedStat = fs.statSync(encryptedFile);
        
        console.log('  - 文件大小对比:');
        console.log('    原始文件:', originalStat.size, '字节');
        console.log('    压缩文件:', compressedStat.size, '字节', 
            `(${((1 - compressedStat.size / originalStat.size) * 100).toFixed(1)}% 压缩)`);
        console.log('    加密文件:', encryptedStat.size, '字节');
        
        // 清理测试文件
        fs.rmdirSync(testDir, { recursive: true });
        console.log('  - 测试文件已清理');
        
    } catch (error) {
        console.error('  - 文件处理错误:', error.message);
    }
}

// 4. 网络请求和错误处理
async function networkRequestDemo() {
    console.log('\n4. 网络请求演示:');
    
    const requests = [
        { name: 'JSON API', url: 'https://httpbin.org/json' },
        { name: 'Status 200', url: 'https://httpbin.org/status/200' },
        { name: 'Status 404', url: 'https://httpbin.org/status/404' },
        { name: 'Invalid URL', url: 'https://nonexistent-domain-12345.com' }
    ];
    
    for (const req of requests) {
        try {
            console.log(`  - 请求 ${req.name}...`);
            const response = await http.get(req.url);
            console.log(`    ✓ 成功: ${response.status} ${response.statusText}`);
            if (response.data && typeof response.data === 'object') {
                console.log(`    数据键: ${Object.keys(response.data).slice(0, 3).join(', ')}`);
            }
        } catch (error) {
            console.log(`    ✗ 失败: ${error.message}`);
        }
    }
}

// 主函数
async function main() {
    console.log('开始综合功能演示...\n');
    
    // 按顺序执行各个演示
    await dataProcessingPipeline();
    await cacheSystemDemo();
    await fileProcessingDemo();
    await networkRequestDemo();
    
    console.log('\n=== 综合演示完成 ===');
    console.log('SW Runtime 提供了完整的企业级功能:');
    console.log('✓ HTTP 客户端 - 网络请求和 API 调用');
    console.log('✓ Redis 客户端 - 高性能数据缓存');
    console.log('✓ 加密模块 - 数据安全保护');
    console.log('✓ 压缩模块 - 数据存储优化');
    console.log('✓ 文件系统 - 完整的文件操作');
    console.log('✓ 路径处理 - 跨平台路径操作');
    console.log('✓ 模块系统 - ES6 和 CommonJS 支持');
    console.log('✓ 异步支持 - Promise 和事件循环');
}

// 启动演示
main().catch(error => {
    console.error('演示过程中发生错误:', error.message);
});

// 导出（如果需要）
export { main };