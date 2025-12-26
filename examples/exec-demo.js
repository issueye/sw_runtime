// exec-demo.js - 命令执行模块演示
console.log('=== SW Runtime exec 模块演示 ===\n');

const exec = require('exec');

// 1. 获取系统信息
console.log('1. 系统信息:');
console.log('   平台:', exec.platform);
console.log('   架构:', exec.arch);
console.log('   当前目录:', exec.cwd());

// 2. 环境变量操作
console.log('\n2. 环境变量:');
console.log('   PATH:', exec.getEnv('PATH') ? '已设置' : '未设置');
console.log('   HOME/USERPROFILE:', exec.getEnv('HOME') || exec.getEnv('USERPROFILE'));

// 设置自定义环境变量
exec.setEnv('MY_TEST_VAR', 'Hello from SW Runtime');
console.log('   MY_TEST_VAR:', exec.getEnv('MY_TEST_VAR'));

// 3. 检查命令是否存在
console.log('\n3. 命令检查:');
const commands = ['node', 'go', 'python', 'git', 'curl'];
commands.forEach(cmd => {
    const exists = exec.commandExists(cmd);
    const path = exec.which(cmd);
    console.log(`   ${cmd}: ${exists ? '✓ 存在' : '✗ 不存在'}${path ? ' (' + path + ')' : ''}`);
});

// 4. 执行简单命令
console.log('\n4. 执行命令:');

// 使用 shell 执行命令
if (exec.platform === 'windows') {
    const result = exec.shell('echo Hello from SW Runtime');
    console.log('   shell echo:', result.stdout.trim());
    
    const dirResult = exec.shell('dir /b');
    console.log('   目录列表 (前3项):');
    const files = dirResult.stdout.trim().split('\n').slice(0, 3);
    files.forEach(f => console.log('     -', f.trim()));
} else {
    const result = exec.shell('echo "Hello from SW Runtime"');
    console.log('   shell echo:', result.stdout.trim());
    
    const lsResult = exec.shell('ls -la | head -5');
    console.log('   目录列表:');
    console.log(lsResult.stdout);
}

// 5. 使用 run 执行命令
console.log('\n5. 使用 run 执行命令:');
const runResult = exec.run('echo "Run command test"');
console.log('   输出:', runResult.output.trim());
console.log('   成功:', runResult.success);
console.log('   退出码:', runResult.exitCode);

// 6. 使用 exec 执行命令（带参数）
console.log('\n6. 使用 exec 执行命令:');
if (exec.platform === 'windows') {
    const execResult = exec.exec('cmd', ['/C', 'echo', 'Exec with args']);
    console.log('   stdout:', execResult.stdout.trim());
    console.log('   成功:', execResult.success);
} else {
    const execResult = exec.exec('echo', ['Exec', 'with', 'args']);
    console.log('   stdout:', execResult.stdout.trim());
    console.log('   成功:', execResult.success);
}

// 7. 带超时的命令执行
console.log('\n7. 带超时的命令执行:');
if (exec.platform === 'windows') {
    const timeoutResult = exec.execWithTimeout('ping', 1000, ['-n', '1', '127.0.0.1']);
    console.log('   命令:', timeoutResult.command);
    console.log('   超时设置:', timeoutResult.timeout, 'ms');
    console.log('   是否超时:', timeoutResult.timedOut);
    console.log('   成功:', timeoutResult.success);
} else {
    const timeoutResult = exec.execWithTimeout('sleep', 500, ['0.1']);
    console.log('   命令:', timeoutResult.command);
    console.log('   超时设置:', timeoutResult.timeout, 'ms');
    console.log('   是否超时:', timeoutResult.timedOut);
    console.log('   成功:', timeoutResult.success);
}

// 8. 异步执行命令
console.log('\n8. 异步执行命令:');
console.log('   开始异步执行...');

exec.execAsync('echo', ['Async command']).then(result => {
    console.log('   异步结果:', result.stdout.trim());
    console.log('   成功:', result.success);
});

// 9. 带选项的命令执行
console.log('\n9. 带选项的命令执行:');
const optResult = exec.shell('echo %CD%', {
    cwd: '.',
    env: {
        MY_CUSTOM_VAR: 'custom_value'
    }
});
console.log('   工作目录输出:', optResult.stdout.trim());

// 10. 错误处理示例
console.log('\n10. 错误处理:');
const errorResult = exec.exec('nonexistent_command_12345', []);
console.log('    命令:', errorResult.command);
console.log('    成功:', errorResult.success);
console.log('    错误:', errorResult.error ? '有错误' : '无错误');

console.log('\n=== exec 模块演示完成 ===');
console.log('\n可用方法:');
console.log('  exec.exec(cmd, args, options)     - 同步执行命令');
console.log('  exec.execAsync(cmd, args, options) - 异步执行命令');
console.log('  exec.run(command, options)        - 执行 shell 命令并返回输出');
console.log('  exec.shell(command, options)      - 执行 shell 命令');
console.log('  exec.execWithTimeout(cmd, ms, args) - 带超时执行');
console.log('  exec.getEnv(key)                  - 获取环境变量');
console.log('  exec.setEnv(key, value)           - 设置环境变量');
console.log('  exec.cwd()                        - 获取当前目录');
console.log('  exec.chdir(path)                  - 改变当前目录');
console.log('  exec.which(cmd)                   - 查找命令路径');
console.log('  exec.commandExists(cmd)           - 检查命令是否存在');
console.log('  exec.platform                     - 当前平台');
console.log('  exec.arch                         - 当前架构');