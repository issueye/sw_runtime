// test_process_v3.js
console.log("--- Process 模块测试 v3 ---");

// 全局对象测试
console.log("1. 基础属性:");
console.log("   - PID:", process.pid);
console.log("   - Platform:", process.platform);
console.log("   - Arch:", process.arch);
console.log("   - CWD:", process.cwd());

// 命令行参数测试
console.log("\n2. 命令行参数 (process.argv):");
if (process.argv && process.argv.length > 0) {
  process.argv.forEach((arg, index) => {
    console.log(`   - argv[${index}]: ${arg}`);
  });
} else {
  console.log("   - argv 为空或 undefined");
}

// 环境变量测试
console.log("\n3. 环境变量 (process.env):");
console.log("   - PATH 存在:", !!(process.env.PATH || process.env.Path));
console.log("   - USERNAME:", process.env.USERNAME || process.env.USER);

// 标准流测试
console.log("\n4. 标准流测试:");
process.stdout.write("   - process.stdout.write 测试成功\n");
process.stderr.write("   - process.stderr.write 测试成功\n");

// 高精度时间测试
console.log("\n5. 高精度时间 (hrtime):");
const start = process.hrtime();
console.log("   - Start:", JSON.stringify(start));
const diff = process.hrtime(start);
console.log("   - Diff:", JSON.stringify(diff));

// nextTick 测试
console.log("\n6. nextTick 测试:");
let nextTickCalled = false;
process.nextTick(() => {
  console.log("   - nextTick 回调执行成功!");
  nextTickCalled = true;
});

// 运行时长和内存测试
setTimeout(() => {
  console.log("\n7. 运行时长和内存:");
  console.log("   - Uptime:", process.uptime().toFixed(2), "s");
  const mem = process.memoryUsage();
  console.log(
    "   - Memory Usage (RSS):",
    (mem.rss / 1024 / 1024).toFixed(2),
    "MB"
  );

  if (!nextTickCalled) {
    console.error("❌ Error: nextTick 回调未执行!");
  } else {
    console.log("✅ 所有测试完成");
  }
}, 50);

// process/exec 模块测试
console.log("\n8. process/exec 模块测试:");
try {
  const exec = require("process/exec");
  const result = exec.execSync("echo hello process/exec");
  console.log("   - execSync 输出:", result.stdout.trim());
} catch (e) {
  console.error("   - execSync 失败:", e.message);
}
