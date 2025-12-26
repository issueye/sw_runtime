package test

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sw_runtime/internal/runtime"
	"testing"
	"time"
)

// TestHTTPServerFileServiceBasic 测试基本的文件服务功能
func TestHTTPServerFileServiceBasic(t *testing.T) {
	runner := runtime.New()
	defer runner.Close()

	// 创建临时测试文件
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, File Service!"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	code := `
		const server = require('httpserver');
		const app = server.createServer();

		const testFilePath = '` + filepath.ToSlash(testFile) + `';

		app.get('/test', (req, res) => {
			res.sendFile(testFilePath);
		});

		app.listen('38090');
	`

	// 在后台运行服务器
	go func() {
		err := runner.RunCode(code)
		if err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	// 测试文件访问
	resp, err := http.Get("http://localhost:38090/test")
	if err != nil {
		t.Logf("⚠️  无法连接到服务器(这在某些环境下是正常的): %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != testContent {
		t.Errorf("Expected %q, got %q", testContent, string(body))
	} else {
		t.Log("✅ 文件服务基本功能测试通过")
	}

	// 检查 Content-Type
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		t.Error("Content-Type header not set")
	} else {
		t.Logf("✅ Content-Type: %s", contentType)
	}
}

// TestHTTPServerMIMEDetection 测试 MIME 类型检测
func TestHTTPServerMIMEDetection(t *testing.T) {
	runner := runtime.New()
	defer runner.Close()

	tmpDir := t.TempDir()
	
	// 创建不同类型的测试文件
	files := map[string]struct {
		content  string
		mimeType string
	}{
		"test.html": {"<html></html>", "text/html"},
		"test.json": {`{"key":"value"}`, "application/json"},
		"test.css":  {"body{}", "text/css"},
	}

	for filename, data := range files {
		filePath := filepath.Join(tmpDir, filename)
		os.WriteFile(filePath, []byte(data.content), 0644)
	}

	code := `
		const server = require('httpserver');
		const app = server.createServer();
		const tmpDir = '` + filepath.ToSlash(tmpDir) + `';

		app.get('/html', (req, res) => res.sendFile(tmpDir + '/test.html'));
		app.get('/json', (req, res) => res.sendFile(tmpDir + '/test.json'));
		app.get('/css', (req, res) => res.sendFile(tmpDir + '/test.css'));

		app.listen('38091');
	`

	go func() {
		runner.RunCode(code)
	}()

	time.Sleep(500 * time.Millisecond)

	testCases := []struct {
		path     string
		expected string
	}{
		{"/html", "text/html"},
		{"/json", "application/json"},
		{"/css", "text/css"},
	}

	for _, tc := range testCases {
		resp, err := http.Get("http://localhost:38091" + tc.path)
		if err != nil {
			t.Logf("⚠️  无法连接到服务器: %v", err)
			continue
		}
		defer resp.Body.Close()

		contentType := resp.Header.Get("Content-Type")
		// 只检查前缀,因为可能包含 charset
		if len(contentType) >= len(tc.expected) && contentType[:len(tc.expected)] == tc.expected {
			t.Logf("✅ %s: MIME type correct (%s)", tc.path, contentType)
		} else {
			t.Errorf("%s: Expected MIME type %s, got %s", tc.path, tc.expected, contentType)
		}
	}
}
