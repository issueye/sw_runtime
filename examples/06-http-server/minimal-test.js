// 最简单的HTTP服务器测试 - 不带回调
const server = require("http/server");

const app = server.createServer();

app.get("/hello", (req, res) => {
  res.send("Hello, World!");
});

console.log("Starting server...");
app.listen("3300"); // 不带回调
console.log("Server started on http://localhost:3300");
