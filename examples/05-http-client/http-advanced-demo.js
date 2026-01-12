// HTTP 高级功能示例 - 请求配置和参数修改
const http = require("http/client");

console.log("=== HTTP Advanced Features Demo ===\n");

// 1. 修改请求头
console.log("--- 测试 1: 自定义请求头 ---");
http
  .get("https://httpbin.org/headers", {
    headers: {
      "User-Agent": "SW-Runtime/1.0",
      Accept: "application/json",
      "X-API-Key": "my-api-key-12345",
      "X-Client-Version": "1.0.0",
    },
  })
  .then((response) => {
    console.log("✅ 请求成功");
    console.log("服务器收到的请求头:");
    console.log(JSON.stringify(response.data.headers, null, 2));
  })
  .catch((err) => {
    console.error("❌ 请求失败:", err.message);
  });

// 2. 修改 URL 查询参数
console.log("\n--- 测试 2: URL 查询参数 ---");
http
  .get("https://httpbin.org/get", {
    params: {
      name: "John Doe",
      age: 30,
      city: "New York",
      tags: "developer,nodejs,javascript",
    },
  })
  .then((response) => {
    console.log("✅ 请求成功");
    console.log("请求 URL:", response.url);
    console.log("查询参数:", response.data.args);
  })
  .catch((err) => {
    console.error("❌ 请求失败:", err.message);
  });

// 3. 修改请求体 - JSON 格式
console.log("\n--- 测试 3: JSON 请求体 ---");
http
  .post("https://httpbin.org/post", {
    headers: {
      "Content-Type": "application/json",
    },
    data: {
      user: {
        name: "Alice",
        email: "alice@example.com",
        age: 25,
      },
      preferences: {
        theme: "dark",
        language: "zh-CN",
      },
      tags: ["developer", "designer"],
    },
  })
  .then((response) => {
    console.log("✅ 请求成功");
    console.log("发送的数据:");
    console.log(JSON.stringify(response.data.json, null, 2));
  })
  .catch((err) => {
    console.error("❌ 请求失败:", err.message);
  });

// 4. 修改请求体 - 字符串格式
console.log("\n--- 测试 4: 文本请求体 ---");
http
  .post("https://httpbin.org/post", {
    headers: {
      "Content-Type": "text/plain",
    },
    data: "This is plain text data\nLine 2\nLine 3",
  })
  .then((response) => {
    console.log("✅ 请求成功");
    console.log("发送的文本:", response.data.data);
  })
  .catch((err) => {
    console.error("❌ 请求失败:", err.message);
  });

// 5. 组合使用请求头和参数
console.log("\n--- 测试 5: 组合请求头和参数 ---");
http
  .post("https://httpbin.org/post", {
    headers: {
      Authorization: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
      "X-Request-ID": "req-" + Date.now(),
      "Content-Type": "application/json",
    },
    params: {
      api_version: "v2",
      format: "json",
    },
    data: {
      action: "create",
      resource: "user",
      payload: {
        username: "bob",
        email: "bob@example.com",
      },
    },
  })
  .then((response) => {
    console.log("✅ 请求成功");
    console.log("请求 URL:", response.url);
    console.log("请求头:", response.data.headers);
    console.log("请求体:", response.data.json);
  })
  .catch((err) => {
    console.error("❌ 请求失败:", err.message);
  });

// 6. PUT 请求 - 更新数据
console.log("\n--- 测试 6: PUT 请求更新数据 ---");
http
  .put("https://httpbin.org/put", {
    headers: {
      "Content-Type": "application/json",
      "If-Match": "etag-12345",
    },
    data: {
      id: 123,
      name: "Updated Name",
      description: "This resource has been updated",
      updatedAt: new Date().toISOString(),
    },
  })
  .then((response) => {
    console.log("✅ 更新成功");
    console.log("更新的数据:", response.data.json);
  })
  .catch((err) => {
    console.error("❌ 更新失败:", err.message);
  });

// 7. PATCH 请求 - 部分更新
console.log("\n--- 测试 7: PATCH 请求部分更新 ---");
http
  .patch("https://httpbin.org/patch", {
    data: {
      status: "active",
      lastLoginAt: new Date().toISOString(),
    },
  })
  .then((response) => {
    console.log("✅ 部分更新成功");
    console.log("更新的字段:", response.data.json);
  })
  .catch((err) => {
    console.error("❌ 更新失败:", err.message);
  });

// 8. DELETE 请求 - 删除资源
console.log("\n--- 测试 8: DELETE 请求 ---");
http
  .delete("https://httpbin.org/delete", {
    headers: {
      Authorization: "Bearer token-12345",
    },
    params: {
      id: 456,
      cascade: "true",
    },
  })
  .then((response) => {
    console.log("✅ 删除成功");
    console.log("删除请求参数:", response.data.args);
  })
  .catch((err) => {
    console.error("❌ 删除失败:", err.message);
  });

// 9. 响应头访问
console.log("\n--- 测试 9: 访问响应头 ---");
http
  .get("https://httpbin.org/response-headers", {
    params: {
      "X-Custom-Header": "CustomValue",
      "Cache-Control": "no-cache",
    },
  })
  .then((response) => {
    console.log("✅ 请求成功");
    console.log("响应状态:", response.status, response.statusText);
    console.log("\n响应头:");
    Object.keys(response.headers).forEach((key) => {
      console.log(`  ${key}: ${response.headers[key]}`);
    });
  })
  .catch((err) => {
    console.error("❌ 请求失败:", err.message);
  });

// 10. Basic 认证
console.log("\n--- 测试 10: Basic 认证 ---");
http
  .get("https://httpbin.org/basic-auth/user/passwd", {
    auth: {
      username: "user",
      password: "passwd",
    },
  })
  .then((response) => {
    console.log("✅ 认证成功");
    console.log("认证结果:", response.data);
  })
  .catch((err) => {
    console.error("❌ 认证失败:", err.message);
  });

// 11. Bearer Token 认证
console.log("\n--- 测试 11: Bearer Token 认证 ---");
http
  .get("https://httpbin.org/bearer", {
    auth: {
      token: "my-secret-token-12345",
    },
  })
  .then((response) => {
    console.log("✅ Token 认证成功");
    console.log("认证结果:", response.data);
  })
  .catch((err) => {
    console.error("❌ Token 认证失败:", err.message);
  });

// 12. 自定义超时
console.log("\n--- 测试 12: 自定义超时（5秒）---");
http
  .get("https://httpbin.org/delay/2", {
    timeout: 5,
    headers: {
      "X-Custom-Timeout": "5s",
    },
  })
  .then((response) => {
    console.log("✅ 请求在超时前完成");
    console.log("延迟:", response.data.args);
  })
  .catch((err) => {
    console.error("❌ 请求超时或失败:", err.message);
  });

// 13. 多个请求并发
console.log("\n--- 测试 13: 并发请求 ---");
Promise.all([
  http.get("https://httpbin.org/uuid"),
  http.get("https://httpbin.org/user-agent"),
  http.get("https://httpbin.org/headers"),
])
  .then((responses) => {
    console.log("✅ 所有请求完成");
    console.log("UUID:", responses[0].data.uuid);
    console.log("User-Agent:", responses[1].data["user-agent"]);
    console.log("请求头数量:", Object.keys(responses[2].data.headers).length);
  })
  .catch((err) => {
    console.error("❌ 某个请求失败:", err.message);
  });

console.log("\n✨ 所有测试已启动，等待响应...\n");
