// compression-demo.ts - 压缩功能演示

const compression = require('compression');

console.log('=== 压缩功能演示 ===');

// 创建测试数据
const smallText = 'Hello, World!';
const mediumText = 'This is a medium-sized text for compression testing. '.repeat(10);
const largeText = 'This is a large text that should compress well due to repetition. '.repeat(100);

function testCompression(name: string, data: string) {
    console.log(`\n${name}:`);
    console.log('原始大小:', data.length, '字符');
    
    // Gzip 压缩测试
    const gzipStart = Date.now();
    const gzipCompressed = compression.gzipCompress(data);
    const gzipCompressTime = Date.now() - gzipStart;
    
    const gzipDecompressStart = Date.now();
    const gzipDecompressed = compression.gzipDecompress(gzipCompressed);
    const gzipDecompressTime = Date.now() - gzipDecompressStart;
    
    console.log('Gzip 压缩后:', gzipCompressed.length, '字符');
    console.log('Gzip 压缩率:', ((1 - gzipCompressed.length / data.length) * 100).toFixed(2) + '%');
    console.log('Gzip 压缩时间:', gzipCompressTime + 'ms');
    console.log('Gzip 解压时间:', gzipDecompressTime + 'ms');
    console.log('Gzip 数据完整性:', data === gzipDecompressed ? '✓' : '✗');
    
    // Zlib 压缩测试
    const zlibStart = Date.now();
    const zlibCompressed = compression.zlibCompress(data);
    const zlibCompressTime = Date.now() - zlibStart;
    
    const zlibDecompressStart = Date.now();
    const zlibDecompressed = compression.zlibDecompress(zlibCompressed);
    const zlibDecompressTime = Date.now() - zlibDecompressStart;
    
    console.log('Zlib 压缩后:', zlibCompressed.length, '字符');
    console.log('Zlib 压缩率:', ((1 - zlibCompressed.length / data.length) * 100).toFixed(2) + '%');
    console.log('Zlib 压缩时间:', zlibCompressTime + 'ms');
    console.log('Zlib 解压时间:', zlibDecompressTime + 'ms');
    console.log('Zlib 数据完整性:', data === zlibDecompressed ? '✓' : '✗');
}

// 测试不同大小的数据
testCompression('小文本测试', smallText);
testCompression('中等文本测试', mediumText);
testCompression('大文本测试', largeText);

// JSON 数据压缩测试
console.log('\nJSON 数据压缩测试:');
const jsonData = JSON.stringify({
    users: Array.from({length: 100}, (_, i) => ({
        id: i + 1,
        name: `User ${i + 1}`,
        email: `user${i + 1}@example.com`,
        active: i % 2 === 0,
        metadata: {
            created: new Date().toISOString(),
            tags: ['user', 'active', 'test']
        }
    }))
});

testCompression('JSON 数据', jsonData);

export { compression };

// 如果在 CommonJS 环境中运行
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { compression };
}