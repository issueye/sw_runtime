// HTTP 服务器文件服务示例
console.log("=== HTTP 服务器文件服务示例 ===\n");

const server = require("http/server");
const path = require("path");

const app = server.createServer();

// 中间件：请求日志
app.use((req, res, next) => {
  console.log(`[${new Date().toISOString()}] ${req.method} ${req.path}`);
  next();
});

// 1. 使用 sendFile 发送单个文件
app.get("/file", (req, res) => {
  const filePath = path.join(__dirname, "examples", "httpserver-demo.ts");
  res.sendFile(filePath);
});

// 2. 使用 download 下载文件
app.get("/download", (req, res) => {
  const filePath = path.join(__dirname, "go.mod");
  res.download(filePath);
});

// 3. 自定义下载文件名
app.get("/download-custom", (req, res) => {
  const filePath = path.join(__dirname, "go.mod");
  res.download(filePath, "project-dependencies.mod");
});

// 4. 根据不同文件类型自动设置 MIME
app.get("/html", (req, res) => {
  const filePath = path.join(
    __dirname,
    "examples",
    "docs_server",
    "index.html"
  );
  res.sendFile(filePath);
});

app.get("/css", (req, res) => {
  const filePath = path.join(
    __dirname,
    "examples",
    "docs_server",
    "assets",
    "css",
    "styles.css"
  );
  res.sendFile(filePath);
});

app.get("/js", (req, res) => {
  const filePath = path.join(
    __dirname,
    "examples",
    "docs_server",
    "assets",
    "js",
    "app.js"
  );
  res.sendFile(filePath);
});

// 5. 文件不存在时的处理
app.get("/nonexistent", (req, res) => {
  res.sendFile("/path/to/nonexistent/file.txt");
  // 会自动返回 404
});

// 6. 使用 static 提供静态文件目录
app.static("./examples", "/static");

// 首页
app.get("/", (req, res) => {
  res.html(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>HTTP 服务器文件服务示例</title>
            <style>
                body { font-family: Arial, sans-serif; margin: 40px; line-height: 1.6; }
                .container { max-width: 800px; margin: 0 auto; }
                .endpoint { background: #f5f5f5; padding: 15px; margin: 10px 0; border-radius: 5px; }
                .method { font-weight: bold; color: #007acc; }
                code { background: #e0e0e0; padding: 2px 5px; border-radius: 3px; }
                h2 { color: #333; border-bottom: 2px solid #007acc; padding-bottom: 10px; }
            </style>
        </head>
        <body>
            <div class="container">
                <h1>🚀 HTTP 服务器文件服务示例</h1>
                
                <h2>📁 文件服务功能</h2>
                
                <div class="endpoint">
                    <span class="method">GET</span> <code>/file</code>
                    <p>使用 sendFile 发送单个文件</p>
                    <a href="/file" target="_blank">访问示例</a>
                </div>
                
                <div class="endpoint">
                    <span class="method">GET</span> <code>/download</code>
                    <p>下载文件(使用原始文件名)</p>
                    <a href="/download" target="_blank">下载文件</a>
                </div>
                
                <div class="endpoint">
                    <span class="method">GET</span> <code>/download-custom</code>
                    <p>下载文件(自定义文件名)</p>
                    <a href="/download-custom" target="_blank">下载文件</a>
                </div>
                
                <h2>🎨 不同文件类型</h2>
                
                <div class="endpoint">
                    <span class="method">GET</span> <code>/html</code>
                    <p>HTML 文件 (自动设置 MIME 为 text/html)</p>
                    <a href="/html" target="_blank">查看 HTML</a>
                </div>
                
                <div class="endpoint">
                    <span class="method">GET</span> <code>/css</code>
                    <p>CSS 文件 (自动设置 MIME 为 text/css)</p>
                    <a href="/css" target="_blank">查看 CSS</a>
                </div>
                
                <div class="endpoint">
                    <span class="method">GET</span> <code>/js</code>
                    <p>JavaScript 文件 (自动设置 MIME 为 application/javascript)</p>
                    <a href="/js" target="_blank">查看 JS</a>
                </div>
                
                <h2>📂 静态文件服务</h2>
                
                <div class="endpoint">
                    <span class="method">GET</span> <code>/static/*</code>
                    <p>访问 examples 目录下的任意文件</p>
                    <a href="/static/httpserver-demo.ts" target="_blank">示例文件</a>
                </div>
                
                <h2>⚠️ 错误处理</h2>
                
                <div class="endpoint">
                    <span class="method">GET</span> <code>/nonexistent</code>
                    <p>文件不存在时自动返回 404</p>
                    <a href="/nonexistent" target="_blank">测试 404</a>
                </div>
                
                <h2>💡 使用说明</h2>
                
                <div class="endpoint">
                    <h3>res.sendFile(filePath)</h3>
                    <p>发送文件并自动检测 MIME 类型:</p>
                    <pre><code>app.get('/file', (req, res) => {
    res.sendFile('./path/to/file.html');
});</code></pre>
                </div>
                
                <div class="endpoint">
                    <h3>res.download(filePath, [filename])</h3>
                    <p>触发浏览器下载文件:</p>
                    <pre><code>app.get('/download', (req, res) => {
    res.download('./file.pdf', 'custom-name.pdf');
});</code></pre>
                </div>
                
                <div class="endpoint">
                    <h3>app.static(directory, prefix)</h3>
                    <p>提供静态文件目录:</p>
                    <pre><code>app.static('./public', '/static');</code></pre>
                </div>
            </div>
        </body>
        </html>
    `);
});

// 启动服务器
const PORT = 3100;
app.listen(PORT.toString(), () => {
  console.log("");
  console.log("🚀 HTTP 服务器启动成功！");
  console.log("📖 访问地址: http://localhost:" + PORT);
  console.log("");
  console.log("📋 可用端点:");
  console.log("   GET  /              - 功能首页");
  console.log("   GET  /file          - sendFile 示例");
  console.log("   GET  /download      - 下载文件");
  console.log("   GET  /html          - HTML 文件");
  console.log("   GET  /css           - CSS 文件");
  console.log("   GET  /js            - JavaScript 文件");
  console.log("   GET  /static/*      - 静态文件目录");
  console.log("");
  console.log("按 Ctrl+C 停止服务器");
});
