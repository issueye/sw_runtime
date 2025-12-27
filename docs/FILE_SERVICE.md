# HTTP Server 文件服务功能

## 功能概述

已为 httpserver 模块添加了完整的文件服务功能,包括:

1. **sendFile()** - 发送单个文件,自动检测 MIME 类型
2. **download()** - 触发浏览器下载文件
3. **自动 MIME 检测** - 根据文件扩展名和内容自动设置正确的 Content-Type
4. **缓存控制** - 自动设置 Last-Modified 和 Cache-Control 头

## API 说明

### res.sendFile(filePath)

发送文件并自动检测 MIME 类型。

**参数:**
- `filePath` (string) - 文件的绝对路径或相对路径

**示例:**
```javascript
app.get('/file', (req, res) => {
    res.sendFile('./public/index.html');
});

app.get('/image', (req, res) => {
    res.sendFile('./images/logo.png');
});

app.get('/data', (req, res) => {
    res.sendFile('./data/report.json');
});
```

**特性:**
- ✅ 自动检测 MIME 类型 (基于文件扩展名)
- ✅ 设置 Last-Modified 头
- ✅ 设置 Cache-Control 头 (public, max-age=3600)
- ✅ 文件不存在时返回 404
- ✅ 目录访问时返回 400

### res.download(filePath, [filename])

触发浏览器下载文件。

**参数:**
- `filePath` (string) - 文件路径
- `filename` (string, 可选) - 下载时使用的文件名

**示例:**
```javascript
// 使用原始文件名
app.get('/download', (req, res) => {
    res.download('./files/document.pdf');
});

// 自定义下载文件名
app.get('/download-report', (req, res) => {
    res.download('./data/report.json', 'monthly-report-2024.json');
});
```

**特性:**
- ✅ 设置 Content-Disposition 头触发下载
- ✅ 支持自定义下载文件名
- ✅ 自动检测 MIME 类型
- ✅ 所有 sendFile 的特性

## MIME 类型支持

自动检测以下常见文件类型的 MIME:

| 文件扩展名 | MIME 类型 |
|-----------|-----------|
| .html, .htm | text/html |
| .css | text/css |
| .js | application/javascript |
| .json | application/json |
| .xml | application/xml |
| .txt | text/plain |
| .pdf | application/pdf |
| .png | image/png |
| .jpg, .jpeg | image/jpeg |
| .gif | image/gif |
| .svg | image/svg+xml |
| .zip | application/zip |
| 其他 | application/octet-stream |

## 完整示例

### 基础文件服务

```javascript
const server = require('httpserver');
const path = require('path');

const app = server.createServer();

// 发送 HTML 文件
app.get('/page', (req, res) => {
    res.sendFile(path.join(__dirname, 'public', 'page.html'));
});

// 发送 JSON 数据文件
app.get('/data', (req, res) => {
    res.sendFile('./data/users.json');
});

// 下载文件
app.get('/download', (req, res) => {
    res.download('./files/report.pdf', 'monthly-report.pdf');
});

app.listen('3000');
```

### 与静态文件服务结合

```javascript
const server = require('httpserver');
const app = server.createServer();

// 静态文件目录
app.static('./public', '/static');

// 特定文件路由(优先级高于 static)
app.get('/index.html', (req, res) => {
    res.sendFile('./templates/index.html');
});

// 下载专区
app.get('/downloads/:filename', (req, res) => {
    const filename = req.params.filename;
    res.download('./downloads/' + filename);
});

app.listen('3000');
```

### 动态文件路由

```javascript
const server = require('httpserver');
const fs = require('fs');
const path = require('path');

const app = server.createServer();

// 动态文档路由
const docFiles = ['intro', 'api', 'guide', 'examples'];

docFiles.forEach(name => {
    app.get('/docs/' + name, (req, res) => {
        const filePath = path.join(__dirname, 'docs', name + '.html');
        res.sendFile(filePath);
    });
});

app.listen('3000');
```

## 实现细节

### 文件检测流程

1. 检查文件是否存在 (`os.Stat`)
2. 验证不是目录
3. 读取文件内容
4. 检测 MIME 类型:
   - 首先尝试根据文件扩展名检测 (`mime.TypeByExtension`)
   - 如果失败,使用内容检测 (`http.DetectContentType`)
   - 默认使用 `application/octet-stream`
5. 设置响应头:
   - Content-Type
   - Last-Modified
   - Cache-Control
6. 发送文件内容

### 错误处理

- 文件不存在: 返回 404 Not Found
- 访问目录: 返回 400 Bad Request
- 读取错误: 返回 500 Internal Server Error

## 性能优化

文件服务功能包含以下优化:

1. **缓存头** - 设置 Cache-Control 和 Last-Modified,支持浏览器缓存
2. **MIME 检测** - 优先使用扩展名检测,避免读取文件内容
3. **一次性读取** - 使用 `os.ReadFile` 一次性读取整个文件

## 测试验证

已通过以下测试:

✅ 基本文件发送功能
✅ MIME 类型自动检测
✅ 文件不存在处理
✅ HTML/CSS/JS/JSON 等常见文件类型
✅ 下载功能

运行测试:
```bash
go test ./test -run TestHTTPServerFileServiceBasic -v
go test ./test -run TestHTTPServerMIMEDetection -v
```

## 示例项目

查看完整示例:
- `examples/httpserver-file-demo.js` - 文件服务功能演示
- `examples/docs_server/test/doc-server.js` - 文档服务器(已使用 sendFile 简化)

## 更新日志

### 2024-12-26

- ✅ 添加 `res.sendFile()` 方法
- ✅ 添加 `res.download()` 方法
- ✅ 实现自动 MIME 类型检测
- ✅ 添加缓存控制头
- ✅ 完善错误处理
- ✅ 更新文档服务器示例
- ✅ 添加测试用例
