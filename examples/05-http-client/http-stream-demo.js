// http-stream-demo.js - HTTP 客户端流式处理演示
//
// 本示例演示如何使用 HTTP 客户端进行流式响应处理和文件上传

const http = require('http/client');
const fs = require('fs');

console.log('=== HTTP 客户端流式处理演示 ===');

// ============================================
// 1. 流式响应下载大文件
// ============================================
console.log('\n--- 1. 流式响应下载大文件 ---');

// 流式下载示例：适合下载大型文件，避免一次性加载到内存
async function streamDownloadExample() {
    // 使用 responseType: 'stream' 启用流式响应
    const response = await http.get('https://raw.githubusercontent.com/example/large-file.zip', {
        responseType: 'stream',
        headers: {
            'Accept-Encoding': 'identity' // 避免压缩，便于处理
        }
    });

    console.log('连接状态:', response.status);
    console.log('内容类型:', response.headers['Content-Type']);
    console.log('文件大小:', response.headers['Content-Length'], 'bytes');

    // 方式1：直接写入文件
    console.log('\n方式1: 使用 pipeToFile 直接写入文件...');
    await response.data.pipeToFile('./download_pipe.zip');
    console.log('文件已保存到: ./download_pipe.zip');
    response.data.close();

    // 方式2：分块读取处理
    console.log('\n方式2: 分块读取处理...');
    const response2 = await http.get('https://raw.githubusercontent.com/example/large-file.zip', {
        responseType: 'stream'
    });

    let totalBytes = 0;
    let chunks = 0;

    while (true) {
        // 读取 8KB 数据块
        const chunk = response2.data.read(8192);
        if (!chunk || chunk.length === 0) break;

        totalBytes += chunk.length;
        chunks++;

        // 可以在这里处理每个数据块，例如：
        // - 写入另一个流
        // - 进行数据转换
        // - 计算哈希值
        // - 发送到其他服务
    }

    console.log(`读取了 ${chunks} 个数据块, 共 ${totalBytes} bytes`);
    response2.data.close();
}

// ============================================
// 2. 使用 copy 方法自定义处理
// ============================================
console.log('\n--- 2. 使用 copy 方法自定义处理 ---');

// copy 方法允许将流复制到任意支持 write 方法的对象
async function copyExample() {
    const response = await http.get('https://example.com/stream-data', {
        responseType: 'stream'
    });

    // 创建自定义 writer 对象
    const myWriter = {
        // 统计写入的字节数
        totalWritten: 0,

        write(chunk) {
            // 在这里可以自定义处理每个数据块
            this.totalWritten += chunk.length;

            // 示例：转换数据（转为大写）
            const upperCaseChunk = chunk.toUpperCase();

            // 写入到文件（追加模式）
            fs.writeFileSync('./output_stream.txt', upperCaseChunk, { flag: 'a' });

            return chunk.length;
        }
    };

    console.log('开始复制流...');
    const written = response.data.copy(myWriter);
    console.log(`复制完成，共 ${written} bytes`);

    // 可以继续使用 myWriter.totalWritten
    console.log('Writer 统计:', myWriter.totalWritten, 'bytes');

    response.data.close();
}

// ============================================
// 3. 文件上传
// ============================================
console.log('\n--- 3. 文件上传 ---');

// 使用 filePath 配置上传本地文件
async function fileUploadExample() {
    // 首先创建一个测试文件
    const testContent = '这是要上传的测试文件内容\n'.repeat(1000);
    fs.writeFileSync('./test_upload.txt', testContent);
    console.log('测试文件已创建: ./test_upload.txt');

    // 上传文件
    const response = await http.post('https://httpbin.org/post', {
        filePath: './test_upload.txt'
        // 会自动设置 Content-Type（根据文件扩展名）
        // .txt -> text/plain, .json -> application/json, 等
    });

    console.log('上传响应状态:', response.status);
    console.log('服务器接收的文件大小:', response.data.files?.length, 'bytes');

    // 清理测试文件
    fs.unlinkSync('./test_upload.txt');
}

// 文件上传（带自定义 Content-Type）
async function fileUploadWithTypeExample() {
    // 创建二进制测试文件（模拟视频片段）
    const buffer = Buffer.alloc(1024 * 100, 0xAA); // 100KB 随机数据
    fs.writeFileSync('./video_fragment.ts', buffer);
    console.log('视频片段已创建: ./video_fragment.ts');

    // 上传 .ts 文件会自动设置 Content-Type: video/mp2t
    const response = await http.post('https://httpbin.org/post', {
        filePath: './video_fragment.ts'
    });

    console.log('上传响应状态:', response.status);
    console.log('Content-Type 检测:', response.headers['Content-Type']);

    // 清理
    fs.unlinkSync('./video_fragment.ts');
}

// ============================================
// 4. 下载 TS 视频流切片
// ============================================
console.log('\n--- 4. 下载 TS 视频流切片 ---');

async function downloadTsStream() {
    // 模拟下载 HLS TS 切片
    const response = await http.get('https://example.com/video/segment1.ts', {
        responseType: 'stream'
    });

    console.log('TS 切片信息:');
    console.log('  状态:', response.status);
    console.log('  Content-Type:', response.headers['Content-Type']);

    // 直接保存为 .ts 文件
    response.data.pipeToFile('./segment1.ts');
    response.data.close();
    console.log('TS 切片已保存: ./segment1.ts');
}

// ============================================
// 运行所有示例
// ============================================
async function main() {
    try {
        // 注意：由于这些是示例，实际运行时需要有效的 URL
        // 注释掉需要网络请求的示例，使用 httpbin.org 进行测试

        // 1. 文件上传测试（使用 httpbin.org）
        await fileUploadExample();
        console.log('\n文件上传示例完成');

        // 2. 文件上传（带类型检测）
        await fileUploadWithTypeExample();
        console.log('\n带类型检测的上传示例完成');

        // 流式下载示例（需要有效的 URL）
        console.log('\n流式下载示例需要有效的文件 URL，暂跳过');
        console.log('如需测试，请将 URL 替换为有效的下载链接');

    } catch (error) {
        console.error('示例运行错误:', error.message);
    }
}

main();

export { };
