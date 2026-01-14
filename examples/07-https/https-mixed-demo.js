// 混合 HTTP/HTTPS 服务器示例
const server = require("http/server");

console.log("=== Mixed HTTP/HTTPS Server Example ===\n");

// 创建 HTTP 服务器
const httpApp = server.createServer();

httpApp.get("/", (req, res) => {
  res.html(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>HTTP Server</title>
        </head>
        <body>
            <h1>HTTP Server (Insecure)</h1>
            <p>This is running on HTTP port 8080</p>
            <p><a href="https://localhost:8443">Switch to HTTPS</a></p>
        </body>
        </html>
    `);
});

httpApp.get("/redirect-to-https", (req, res) => {
  res.redirect("https://localhost:8443", 301);
});

// 创建 HTTPS 服务器
const httpsApp = server.createServer();

httpsApp.get("/", (req, res) => {
  res.html(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>HTTPS Server</title>
            <style>
                body {
                    font-family: Arial, sans-serif;
                    max-width: 600px;
                    margin: 50px auto;
                    padding: 20px;
                    background: #2c3e50;
                    color: white;
                }
            </style>
        </head>
        <body>
            <h1>🔐 HTTPS Server (Secure)</h1>
            <p>This is running on HTTPS port 8443</p>
            <p>Your connection is secure!</p>
            <p><a href="http://localhost:8080" style="color: #3498db;">Switch to HTTP</a></p>
        </body>
        </html>
    `);
});

httpsApp.get("/api/secure-data", (req, res) => {
  res.json({
    message: "This is secure data",
    encrypted: true,
    data: {
      secret: "Only visible over HTTPS",
      timestamp: Date.now(),
    },
  });
});

// 启动 HTTP 服务器
httpApp.listen("8080").then(() => {
  console.log("✓ HTTP  Server listening on http://localhost:8080");
});

// 启动 HTTPS 服务器
httpsApp
  .listenTLS(
    "8443",
    "./examples/certs/server.crt",
    "./examples/certs/server.key",
  )
  .then(() => {
    console.log("✓ HTTPS Server listening on https://localhost:8443");
    console.log("\n📝 Both servers are running:");
    console.log("   HTTP:  http://localhost:8080");
    console.log("   HTTPS: https://localhost:8443\n");
  })
  .catch((err) => {
    console.error("Failed to start HTTPS server:", err.message);
  });

setTimeout(() => {
  console.log("Servers are running... Press Ctrl+C to stop");
}, 2000);
