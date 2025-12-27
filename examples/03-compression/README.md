# 压缩模块示例

本目录包含 compression/zlib 压缩模块的功能演示。

## 文件说明

- **compression-demo.ts** - 压缩模块完整演示

## 功能特点

### Gzip 压缩
- 数据压缩
- 数据解压
- 高性能压缩算法

### Zlib 压缩
- 数据压缩
- 数据解压

## 运行示例

```bash
sw_runtime run examples/03-compression/compression-demo.ts
```

## 示例代码

```javascript
const compression = require('compression');

// Gzip 压缩
const data = 'This is a long text that needs compression...';
const compressed = compression.gzipCompress(data);
console.log('压缩后大小:', compressed.length);

const decompressed = compression.gzipDecompress(compressed);
console.log('解压后:', decompressed);

// Zlib 压缩
const zlibCompressed = compression.zlibCompress(data);
const zlibDecompressed = compression.zlibDecompress(zlibCompressed);
```
