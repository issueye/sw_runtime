// 并发测试 - 验证 VMProcessor 的线程安全
const server = require("http/server");

// 创建服务器
const app = server.createServer({
  readTimeout: 10,
  writeTimeout: 10,
});

console.log("创建服务器成功");

// 添加多个路由测试并发
app.get("/test1", (req, res) => {
  console.log("处理请求: /test1");
  res.json({ route: "test1", message: "Hello from test1" });
});

app.get("/test2", (req, res) => {
  console.log("处理请求: /test2");
  res.json({ route: "test2", message: "Hello from test2" });
});

app.get("/test3", (req, res) => {
  console.log("处理请求: /test3");
  res.json({ route: "test3", message: "Hello from test3" });
});

app.post("/echo", (req, res) => {
  console.log("处理请求: POST /echo");
  res.json({
    method: req.method,
    body: req.body,
    echo: "Echo endpoint",
  });
});

// 启动服务器
app.listen("3300", () => {
  console.log("服务器已启动在 http://localhost:3300");
  console.log("可以使用多个并发请求测试 VMProcessor 的线程安全性");
});
