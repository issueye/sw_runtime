var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

// examples/comprehensive-demo.ts
var comprehensive_demo_exports = {};
__export(comprehensive_demo_exports, {
  main: () => main
});
module.exports = __toCommonJS(comprehensive_demo_exports);
console.log("=== SW Runtime \u7EFC\u5408\u529F\u80FD\u6F14\u793A ===");
var http = require("http");
var redis = require("redis");
var crypto = require("crypto");
var compression = require("compression");
var fs = require("fs");
var path = require("path");
async function dataProcessingPipeline() {
  console.log("\n1. \u6570\u636E\u5904\u7406\u6D41\u6C34\u7EBF\u6F14\u793A:");
  try {
    console.log("  - \u4ECE API \u83B7\u53D6\u6570\u636E...");
    const response = await http.get("https://jsonplaceholder.typicode.com/posts/1");
    const apiData = response.data;
    console.log("  - API \u6570\u636E\u83B7\u53D6\u6210\u529F:", apiData.title);
    console.log("  - \u52A0\u5BC6\u6570\u636E...");
    const jsonData = JSON.stringify(apiData);
    const encryptionKey = "my-secret-key-32-bytes-long!!!";
    const encryptedData = crypto.aesEncrypt(jsonData, encryptionKey);
    console.log("  - \u6570\u636E\u52A0\u5BC6\u5B8C\u6210\uFF0C\u957F\u5EA6:", encryptedData.length);
    console.log("  - \u538B\u7F29\u52A0\u5BC6\u6570\u636E...");
    const compressedData = compression.gzipCompress(encryptedData);
    console.log(
      "  - \u538B\u7F29\u5B8C\u6210\uFF0C\u538B\u7F29\u7387:",
      ((1 - compressedData.length / encryptedData.length) * 100).toFixed(2) + "%"
    );
    console.log("  - \u4FDD\u5B58\u5230\u6587\u4EF6...");
    const fileName = "processed_data.bin";
    fs.writeFileSync(fileName, compressedData);
    console.log("  - \u6587\u4EF6\u4FDD\u5B58\u6210\u529F:", fileName);
    console.log("  - \u8BFB\u53D6\u548C\u89E3\u5BC6\u6570\u636E...");
    const readData = fs.readFileSync(fileName, "utf8");
    const decompressedData = compression.gzipDecompress(readData);
    const decryptedData = crypto.aesDecrypt(decompressedData, encryptionKey);
    const originalData = JSON.parse(decryptedData);
    console.log("  - \u6570\u636E\u6062\u590D\u6210\u529F:", originalData.title);
    console.log(
      "  - \u6570\u636E\u5B8C\u6574\u6027\u9A8C\u8BC1:",
      JSON.stringify(originalData) === JSON.stringify(apiData) ? "\u2713" : "\u2717"
    );
    fs.unlinkSync(fileName);
    console.log("  - \u4E34\u65F6\u6587\u4EF6\u5DF2\u6E05\u7406");
  } catch (error) {
    console.error("  - \u6570\u636E\u5904\u7406\u6D41\u6C34\u7EBF\u9519\u8BEF:", error.message);
  }
}
async function cacheSystemDemo() {
  console.log("\n2. \u7F13\u5B58\u7CFB\u7EDF\u6F14\u793A:");
  try {
    const client = redis.createClient({
      host: "localhost",
      port: 6379
    });
    console.log("  - Redis \u8FDE\u63A5\u6210\u529F");
    async function fetchUserData(userId) {
      const cacheKey = `user:${userId}`;
      const cached = await client.getJSON(cacheKey);
      if (cached) {
        console.log("  - \u4ECE\u7F13\u5B58\u83B7\u53D6\u7528\u6237\u6570\u636E:", cached.name);
        return cached;
      }
      console.log("  - \u7F13\u5B58\u672A\u547D\u4E2D\uFF0C\u4ECE API \u83B7\u53D6\u6570\u636E...");
      const response = await http.get(`https://jsonplaceholder.typicode.com/users/${userId}`);
      const userData = response.data;
      await client.setJSON(cacheKey, userData, 300);
      console.log("  - \u6570\u636E\u5DF2\u7F13\u5B58:", userData.name);
      return userData;
    }
    const user1 = await fetchUserData(1);
    const user1Cached = await fetchUserData(1);
    console.log("  - \u7F13\u5B58\u7CFB\u7EDF\u6D4B\u8BD5\u5B8C\u6210");
  } catch (error) {
    console.log("  - Redis \u4E0D\u53EF\u7528 (\u8FD9\u662F\u6B63\u5E38\u7684\uFF0C\u5982\u679C\u6CA1\u6709\u8FD0\u884C Redis \u670D\u52A1\u5668)");
    console.log("  - \u9519\u8BEF:", error.message);
  }
}
async function fileProcessingDemo() {
  console.log("\n3. \u6587\u4EF6\u5904\u7406\u6F14\u793A:");
  try {
    const testDir = "test_workspace";
    const dataDir = path.join(testDir, "data");
    const outputDir = path.join(testDir, "output");
    console.log("  - \u521B\u5EFA\u76EE\u5F55\u7ED3\u6784...");
    fs.mkdirSync(testDir, { recursive: true });
    fs.mkdirSync(dataDir, { recursive: true });
    fs.mkdirSync(outputDir, { recursive: true });
    const testData = {
      timestamp: (/* @__PURE__ */ new Date()).toISOString(),
      data: Array.from({ length: 100 }, (_, i) => ({
        id: i + 1,
        value: Math.random() * 1e3,
        category: ["A", "B", "C"][i % 3]
      }))
    };
    const originalFile = path.join(dataDir, "original.json");
    fs.writeFileSync(originalFile, JSON.stringify(testData, null, 2));
    console.log("  - \u539F\u59CB\u6570\u636E\u5DF2\u4FDD\u5B58:", originalFile);
    const compressedData = compression.gzipCompress(JSON.stringify(testData));
    const compressedFile = path.join(outputDir, "compressed.gz");
    fs.writeFileSync(compressedFile, compressedData);
    console.log("  - \u538B\u7F29\u6570\u636E\u5DF2\u4FDD\u5B58:", compressedFile);
    const encryptedData = crypto.aesEncrypt(JSON.stringify(testData), "encryption-key-32-bytes-long!");
    const encryptedFile = path.join(outputDir, "encrypted.bin");
    fs.writeFileSync(encryptedFile, encryptedData);
    console.log("  - \u52A0\u5BC6\u6570\u636E\u5DF2\u4FDD\u5B58:", encryptedFile);
    const originalStat = fs.statSync(originalFile);
    const compressedStat = fs.statSync(compressedFile);
    const encryptedStat = fs.statSync(encryptedFile);
    console.log("  - \u6587\u4EF6\u5927\u5C0F\u5BF9\u6BD4:");
    console.log("    \u539F\u59CB\u6587\u4EF6:", originalStat.size, "\u5B57\u8282");
    console.log(
      "    \u538B\u7F29\u6587\u4EF6:",
      compressedStat.size,
      "\u5B57\u8282",
      `(${((1 - compressedStat.size / originalStat.size) * 100).toFixed(1)}% \u538B\u7F29)`
    );
    console.log("    \u52A0\u5BC6\u6587\u4EF6:", encryptedStat.size, "\u5B57\u8282");
    fs.rmdirSync(testDir, { recursive: true });
    console.log("  - \u6D4B\u8BD5\u6587\u4EF6\u5DF2\u6E05\u7406");
  } catch (error) {
    console.error("  - \u6587\u4EF6\u5904\u7406\u9519\u8BEF:", error.message);
  }
}
async function networkRequestDemo() {
  console.log("\n4. \u7F51\u7EDC\u8BF7\u6C42\u6F14\u793A:");
  const requests = [
    { name: "JSON API", url: "https://httpbin.org/json" },
    { name: "Status 200", url: "https://httpbin.org/status/200" },
    { name: "Status 404", url: "https://httpbin.org/status/404" },
    { name: "Invalid URL", url: "https://nonexistent-domain-12345.com" }
  ];
  for (const req of requests) {
    try {
      console.log(`  - \u8BF7\u6C42 ${req.name}...`);
      const response = await http.get(req.url);
      console.log(`    \u2713 \u6210\u529F: ${response.status} ${response.statusText}`);
      if (response.data && typeof response.data === "object") {
        console.log(`    \u6570\u636E\u952E: ${Object.keys(response.data).slice(0, 3).join(", ")}`);
      }
    } catch (error) {
      console.log(`    \u2717 \u5931\u8D25: ${error.message}`);
    }
  }
}
async function main() {
  console.log("\u5F00\u59CB\u7EFC\u5408\u529F\u80FD\u6F14\u793A...\n");
  await dataProcessingPipeline();
  await cacheSystemDemo();
  await fileProcessingDemo();
  await networkRequestDemo();
  console.log("\n=== \u7EFC\u5408\u6F14\u793A\u5B8C\u6210 ===");
  console.log("SW Runtime \u63D0\u4F9B\u4E86\u5B8C\u6574\u7684\u4F01\u4E1A\u7EA7\u529F\u80FD:");
  console.log("\u2713 HTTP \u5BA2\u6237\u7AEF - \u7F51\u7EDC\u8BF7\u6C42\u548C API \u8C03\u7528");
  console.log("\u2713 Redis \u5BA2\u6237\u7AEF - \u9AD8\u6027\u80FD\u6570\u636E\u7F13\u5B58");
  console.log("\u2713 \u52A0\u5BC6\u6A21\u5757 - \u6570\u636E\u5B89\u5168\u4FDD\u62A4");
  console.log("\u2713 \u538B\u7F29\u6A21\u5757 - \u6570\u636E\u5B58\u50A8\u4F18\u5316");
  console.log("\u2713 \u6587\u4EF6\u7CFB\u7EDF - \u5B8C\u6574\u7684\u6587\u4EF6\u64CD\u4F5C");
  console.log("\u2713 \u8DEF\u5F84\u5904\u7406 - \u8DE8\u5E73\u53F0\u8DEF\u5F84\u64CD\u4F5C");
  console.log("\u2713 \u6A21\u5757\u7CFB\u7EDF - ES6 \u548C CommonJS \u652F\u6301");
  console.log("\u2713 \u5F02\u6B65\u652F\u6301 - Promise \u548C\u4E8B\u4EF6\u5FAA\u73AF");
}
main().catch((error) => {
  console.error("\u6F14\u793A\u8FC7\u7A0B\u4E2D\u53D1\u751F\u9519\u8BEF:", error.message);
});
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  main
});
