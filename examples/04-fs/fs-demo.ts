// fs-demo.ts - 文件系统功能演示

import * as fs from 'fs';
import * as path from 'path';

console.log('=== 文件系统功能演示 ===');

// 1. 同步文件操作
console.log('\n1. 同步文件操作:');
try {
    const testFile = 'test-sync.txt';
    const testContent = 'Hello, Sync File System!';
    
    // 写入文件
    fs.writeFileSync(testFile, testContent);
    console.log('✓ 文件写入成功');
    
    // 读取文件
    const content = fs.readFileSync(testFile, 'utf8');
    console.log('✓ 文件读取成功:', content);
    
    // 检查文件是否存在
    console.log('✓ 文件存在:', fs.existsSync(testFile));
    
    // 获取文件信息
    const stats = fs.statSync(testFile);
    console.log('✓ 文件大小:', stats.size, '字节');
    console.log('✓ 是否为文件:', stats.isFile());
    console.log('✓ 是否为目录:', stats.isDirectory());
    
    // 删除文件
    fs.unlinkSync(testFile);
    console.log('✓ 文件删除成功');
    console.log('✓ 文件存在:', fs.existsSync(testFile));
    
} catch (error) {
    console.error('✗ 同步操作错误:', error.message);
}

// 2. 异步文件操作
console.log('\n2. 异步文件操作:');
const testAsyncFile = 'test-async.txt';
const testAsyncContent = 'Hello, Async File System!';

fs.writeFile(testAsyncFile, testAsyncContent)
    .then(() => {
        console.log('✓ 异步文件写入成功');
        return fs.readFile(testAsyncFile);
    })
    .then((content: string) => {
        console.log('✓ 异步文件读取成功:', content);
        return fs.stat(testAsyncFile);
    })
    .then((stats: any) => {
        console.log('✓ 异步获取文件信息成功');
        console.log('  - 文件大小:', stats.size, '字节');
        console.log('  - 是否为文件:', stats.isFile());
        return fs.unlink(testAsyncFile);
    })
    .then(() => {
        console.log('✓ 异步文件删除成功');
    })
    .catch((error: Error) => {
        console.error('✗ 异步操作错误:', error.message);
    });

// 3. 目录操作
console.log('\n3. 目录操作:');
const testDir = 'test-directory';

setTimeout(() => {
    try {
        // 创建目录
        fs.mkdirSync(testDir, { recursive: true });
        console.log('✓ 目录创建成功');
        
        // 在目录中创建文件
        const fileInDir = path.join(testDir, 'file-in-dir.txt');
        fs.writeFileSync(fileInDir, 'File in directory');
        console.log('✓ 目录中文件创建成功');
        
        // 读取目录内容
        const dirContents = fs.readdirSync(testDir);
        console.log('✓ 目录内容:', dirContents);
        
        // 复制文件
        const copiedFile = path.join(testDir, 'copied-file.txt');
        fs.copyFileSync(fileInDir, copiedFile);
        console.log('✓ 文件复制成功');
        
        // 重命名文件
        const renamedFile = path.join(testDir, 'renamed-file.txt');
        fs.renameSync(copiedFile, renamedFile);
        console.log('✓ 文件重命名成功');
        
        // 再次读取目录内容
        const newDirContents = fs.readdirSync(testDir);
        console.log('✓ 更新后目录内容:', newDirContents);
        
        // 清理：删除目录及其内容
        fs.rmdirSync(testDir, { recursive: true });
        console.log('✓ 目录删除成功');
        
    } catch (error) {
        console.error('✗ 目录操作错误:', error.message);
    }
}, 200);

// 4. 异步目录操作
console.log('\n4. 异步目录操作:');
const asyncTestDir = 'async-test-directory';

setTimeout(() => {
    fs.mkdir(asyncTestDir)
        .then(() => {
            console.log('✓ 异步目录创建成功');
            return fs.writeFile(path.join(asyncTestDir, 'async-file.txt'), 'Async file content');
        })
        .then(() => {
            console.log('✓ 异步文件创建成功');
            return fs.readdir(asyncTestDir);
        })
        .then((files: string[]) => {
            console.log('✓ 异步读取目录成功:', files);
            return fs.rmdir(asyncTestDir);
        })
        .then(() => {
            console.log('✓ 异步目录删除成功');
        })
        .catch((error: Error) => {
            console.error('✗ 异步目录操作错误:', error.message);
        });
}, 400);

export { fs, path };

// 如果在 CommonJS 环境中运行
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { fs, path };
}