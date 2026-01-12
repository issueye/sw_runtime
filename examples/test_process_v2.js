// test_process_v2.js
console.log("--- Process 模块测试 ---");

// 全局对象测试
console.log("1. 全局 process 对象:", typeof process);
console.log("   - PID:", process.pid);
console.log("   - Platform:", process.platform);
console.log("   - Arch:", process.arch);
console.log("   - CWD:", process.cwd());
console.log("   - Versions:", JSON.stringify(process.versions, null, 2));

// 环境变量测试
console.log("\n2. 环境变量 (process.env):");
console.log("   - PATH 存在:", !!process.env.PATH || !!process.env.Path);
console.log("   - USER/USERNAME:", process.env.USER || process.env.USERNAME);

// 命令行参数测试
console.log("\n3. 命令行参数 (process.argv):");
process.argv.forEach((arg, index) => {
  console.log(`   - argv[${index}]: ${arg}`);
});

// nextTick 测试
console.log("\n4. nextTick 测试:");
let nextTickCalled = false;
process.nextTick(() => {
  console.log("   - nextTick 回调执行成功!");
  nextTickCalled = true;
});
console.log("   - nextTick 已调度 (应该是异步的)");

// uptime 和 memoryUsage 测试
setTimeout(() => {
  console.log("\n5. 运行时长和内存:");
  console.log("   - Uptime:", process.uptime().toFixed(2), "s");
  const mem = process.memoryUsage();
  console.log(
    "   - Memory Usage:",
    JSON.stringify(
      {
        rss: (mem.rss / 1024 / 1024).toFixed(2) + " MB",
        heapUsed: (mem.heapUsed / 1024 / 1024).toFixed(2) + " MB",
      },
      null,
      2
    )
  );

  if (!nextTickCalled) {
    console.error("❌ Error: nextTick 回调未执行!");
  } else {
    console.log("✅ 测试完成");
  }
}, 100);

// process/exec 模块测试
console.log("\n6. process/exec 模块测试:");
try {
  const exec = require("process/exec");
  console.log("   - 模块加载成功:", typeof exec.execSync);
  const result = exec.execSync("echo hello sw_runtime");
  console.log("   - 执行命令成功:", result.stdout.trim());
} catch (e) {
  console.error("   - 模块加载或执行失败:", e.message);
}
