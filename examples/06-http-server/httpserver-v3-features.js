/**
 * HTTP Server v3.0 New Features Demo
 * - Path Parameters
 * - Stream File Response
 * - Enhanced Request Object
 */
const server = require("http/server");
const fs = require("fs");
const app = server.createServer();

// 1. Path Parameters Demo
// Match /user/123, /user/abc, etc.
app.get("/user/:id", (req, res) => {
  const userId = req.params.id;
  console.log(`Fetching data for user: ${userId}`);

  res.json({
    id: userId,
    name: "John Doe",
    path: req.path,
    params: req.params, // Path parameters
  });
});

// Multiple parameters
app.get("/posts/:year/:month/:day", (req, res) => {
  res.json({
    date: `${req.params.year}-${req.params.month}-${req.params.day}`,
    params: req.params,
  });
});

// 2. Enhanced Request Object Demo
app.post("/debug-request", (req, res) => {
  res.json({
    protocol: req.protocol, // http or https
    secure: req.secure, // boolean
    hostname: req.hostname, // e.g. localhost
    xhr: req.xhr, // boolean
    userAgent: req.get("user-agent"), // req.get() method
    cookies: req.cookies, // parsed cookies
    isJson: req.is("json"), // req.is() method
    query: req.query, // parsed query string
    body: req.json || req.form, // parsed body
  });
});

// 3. Stream File Response Demo
app.get("/download-test", (req, res) => {
  const tempFile = "temp_download_demo.txt";
  const content = "This is a large file content demo for streaming.\n".repeat(
    100
  );

  // Create a temporary file
  fs.writeFileSync(tempFile, content);

  console.log(`Sending file via stream: ${tempFile}`);

  // res.sendFile uses http.ServeFile internally for efficient streaming,
  // range requests, and automatic MIME detection.
  res.sendFile(tempFile);

  // Clean up hint: In a real app, you might want a timeout or a manual cleanup later.
});

app.listen(3007, () => {
  console.log("Server v3 demo listening on http://localhost:3007");
  console.log("Available routes:");
  console.log("  GET  /user/:id");
  console.log("  GET  /posts/:year/:month/:day");
  console.log("  POST /debug-request");
  console.log("  GET  /download-test");
});
