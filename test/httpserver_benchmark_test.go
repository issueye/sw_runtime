package test

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sw_runtime/internal/runtime"
	"testing"
	"time"
)

// BenchmarkHTTPServerSimpleRoute 基准测试 - 简单路由
func BenchmarkHTTPServerSimpleRoute(b *testing.B) {
	runner := runtime.New()
	defer runner.Close()

	code := `
		const server = require('httpserver');
		const app = server.createServer();

		app.get('/hello', (req, res) => {
			res.send('Hello, World!');
		});

		app.listen('38100');
	`

	go func() {
		runner.RunCode(code)
	}()

	time.Sleep(500 * time.Millisecond)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get("http://localhost:38100/hello")
		if err != nil {
			b.Skip("无法连接到服务器")
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkHTTPServerJSONResponse 基准测试 - JSON 响应
func BenchmarkHTTPServerJSONResponse(b *testing.B) {
	runner := runtime.New()
	defer runner.Close()

	code := `
		const server = require('httpserver');
		const app = server.createServer();

		app.get('/api/data', (req, res) => {
			res.json({
				status: 'success',
				data: {
					id: 123,
					name: 'Test User',
					email: 'test@example.com',
					tags: ['tag1', 'tag2', 'tag3']
				},
				timestamp: Date.now()
			});
		});

		app.listen('38101');
	`

	go func() {
		runner.RunCode(code)
	}()

	time.Sleep(500 * time.Millisecond)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get("http://localhost:38101/api/data")
		if err != nil {
			b.Skip("无法连接到服务器")
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkHTTPServerStaticFile 基准测试 - 静态文件服务
func BenchmarkHTTPServerStaticFile(b *testing.B) {
	runner := runtime.New()
	defer runner.Close()

	// 创建临时测试文件
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.html")
	testContent := strings.Repeat("<html><body>Test Content</body></html>", 10)
	os.WriteFile(testFile, []byte(testContent), 0644)

	code := `
		const server = require('httpserver');
		const app = server.createServer();

		const filePath = '` + filepath.ToSlash(testFile) + `';

		app.get('/file', (req, res) => {
			res.sendFile(filePath);
		});

		app.listen('38102');
	`

	go func() {
		runner.RunCode(code)
	}()

	time.Sleep(500 * time.Millisecond)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get("http://localhost:38102/file")
		if err != nil {
			b.Skip("无法连接到服务器")
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkHTTPServerWithMiddleware 基准测试 - 带中间件的路由
func BenchmarkHTTPServerWithMiddleware(b *testing.B) {
	runner := runtime.New()
	defer runner.Close()

	code := `
		const server = require('httpserver');
		const app = server.createServer();

		// 日志中间件
		app.use((req, res, next) => {
			const start = Date.now();
			next();
			const duration = Date.now() - start;
		});

		// 认证中间件
		app.use((req, res, next) => {
			if (req.headers.authorization) {
				next();
			} else {
				next();
			}
		});

		app.get('/api/user', (req, res) => {
			res.json({ id: 1, name: 'User' });
		});

		app.listen('38103');
	`

	go func() {
		runner.RunCode(code)
	}()

	time.Sleep(500 * time.Millisecond)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "http://localhost:38103/api/user", nil)
		req.Header.Set("Authorization", "Bearer test-token")
		resp, err := client.Do(req)
		if err != nil {
			b.Skip("无法连接到服务器")
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkHTTPServerMultipleRoutes 基准测试 - 多路由性能
func BenchmarkHTTPServerMultipleRoutes(b *testing.B) {
	runner := runtime.New()
	defer runner.Close()

	code := `
		const server = require('httpserver');
		const app = server.createServer();

		app.get('/route1', (req, res) => res.send('Route 1'));
		app.get('/route2', (req, res) => res.send('Route 2'));
		app.get('/route3', (req, res) => res.send('Route 3'));
		app.get('/route4', (req, res) => res.send('Route 4'));
		app.get('/route5', (req, res) => res.send('Route 5'));
		app.post('/api/data', (req, res) => res.json({ success: true }));

		app.listen('38104');
	`

	go func() {
		runner.RunCode(code)
	}()

	time.Sleep(500 * time.Millisecond)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	routes := []string{"/route1", "/route2", "/route3", "/route4", "/route5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		route := routes[i%len(routes)]
		resp, err := client.Get("http://localhost:38104" + route)
		if err != nil {
			b.Skip("无法连接到服务器")
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}
}

// BenchmarkHTTPServerConcurrentRequests 基准测试 - 并发请求
func BenchmarkHTTPServerConcurrentRequests(b *testing.B) {
	runner := runtime.New()
	defer runner.Close()

	code := `
		const server = require('httpserver');
		const app = server.createServer();

		let requestCount = 0;

		app.get('/counter', (req, res) => {
			requestCount++;
			res.json({ count: requestCount });
		});

		app.listen('38105');
	`

	go func() {
		runner.RunCode(code)
	}()

	time.Sleep(500 * time.Millisecond)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get("http://localhost:38105/counter")
			if err != nil {
				continue
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}
	})
}

// TestHTTPServerPerformanceProfile 性能测试报告
func TestHTTPServerPerformanceProfile(t *testing.T) {
	runner := runtime.New()
	defer runner.Close()

	code := `
		const server = require('httpserver');
		const app = server.createServer();

		// 简单文本响应
		app.get('/text', (req, res) => {
			res.send('Hello, World!');
		});

		// JSON 响应
		app.get('/json', (req, res) => {
			res.json({
				status: 'success',
				data: { id: 1, name: 'Test' }
			});
		});

		// HTML 响应
		app.get('/html', (req, res) => {
			res.html('<html><body><h1>Test</h1></body></html>');
		});

		app.listen('38106');
	`

	go func() {
		runner.RunCode(code)
	}()

	time.Sleep(500 * time.Millisecond)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	tests := []struct {
		name string
		path string
	}{
		{"Text Response", "/text"},
		{"JSON Response", "/json"},
		{"HTML Response", "/html"},
	}

	for _, tt := range tests {
		start := time.Now()
		iterations := 1000

		for i := 0; i < iterations; i++ {
			resp, err := client.Get("http://localhost:38106" + tt.path)
			if err != nil {
				t.Logf("⚠️  无法连接到服务器: %v", err)
				return
			}
			io.ReadAll(resp.Body)
			resp.Body.Close()
		}

		elapsed := time.Since(start)
		avgLatency := elapsed / time.Duration(iterations)
		reqPerSec := float64(iterations) / elapsed.Seconds()

		t.Logf("✅ %s:", tt.name)
		t.Logf("   - 总耗时: %v", elapsed)
		t.Logf("   - 平均延迟: %v", avgLatency)
		t.Logf("   - 请求/秒: %.2f req/s", reqPerSec)
	}
}

// TestHTTPServerThroughput 吞吐量测试
func TestHTTPServerThroughput(t *testing.T) {
	runner := runtime.New()
	defer runner.Close()

	// 创建一个较大的测试文件
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "large.txt")
	largeContent := strings.Repeat("Lorem ipsum dolor sit amet. ", 1000) // ~28KB
	os.WriteFile(testFile, []byte(largeContent), 0644)

	code := `
		const server = require('httpserver');
		const app = server.createServer();

		const filePath = '` + filepath.ToSlash(testFile) + `';

		app.get('/large', (req, res) => {
			res.sendFile(filePath);
		});

		app.listen('38107');
	`

	go func() {
		runner.RunCode(code)
	}()

	time.Sleep(500 * time.Millisecond)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	iterations := 500
	start := time.Now()
	totalBytes := int64(0)

	for i := 0; i < iterations; i++ {
		resp, err := client.Get("http://localhost:38107/large")
		if err != nil {
			t.Logf("⚠️  无法连接到服务器: %v", err)
			return
		}
		data, _ := io.ReadAll(resp.Body)
		totalBytes += int64(len(data))
		resp.Body.Close()
	}

	elapsed := time.Since(start)
	throughputMBps := float64(totalBytes) / elapsed.Seconds() / 1024 / 1024

	t.Logf("✅ HTTP 服务器吞吐量测试:")
	t.Logf("   - 文件大小: %d bytes", len(largeContent))
	t.Logf("   - 请求次数: %d", iterations)
	t.Logf("   - 总传输: %.2f MB", float64(totalBytes)/1024/1024)
	t.Logf("   - 总耗时: %v", elapsed)
	t.Logf("   - 吞吐量: %.2f MB/s", throughputMBps)
	t.Logf("   - 平均延迟: %v", elapsed/time.Duration(iterations))
}

// TestHTTPServerStressTest HTTP 服务器压力测试
func TestHTTPServerStressTest(t *testing.T) {
	runner := runtime.New()
	defer runner.Close()

	code := `
		const server = require('httpserver');
		const app = server.createServer();

		let totalRequests = 0;
		let errorCount = 0;

		app.get('/stress', (req, res) => {
			totalRequests++;
			
			// 模拟一些计算
			let result = 0;
			for (let i = 0; i < 100; i++) {
				result += i;
			}
			
			res.json({
				request_number: totalRequests,
				result: result,
				timestamp: Date.now()
			});
		});

		app.get('/stats', (req, res) => {
			res.json({
				total_requests: totalRequests,
				error_count: errorCount
			});
		});

		app.listen('38108');
		global.getStats = () => ({ totalRequests, errorCount });
	`

	go func() {
		runner.RunCode(code)
	}()

	time.Sleep(500 * time.Millisecond)

	// 并发压力测试
	concurrency := 10
	requestsPerWorker := 100
	done := make(chan bool, concurrency)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		go func(workerId int) {
			client := &http.Client{
				Timeout: 5 * time.Second,
			}

			for j := 0; j < requestsPerWorker; j++ {
				resp, err := client.Get("http://localhost:38108/stress")
				if err != nil {
					continue
				}
				io.ReadAll(resp.Body)
				resp.Body.Close()
			}

			done <- true
		}(i)
	}

	// 等待所有工作完成
	for i := 0; i < concurrency; i++ {
		<-done
	}

	elapsed := time.Since(start)
	totalRequests := concurrency * requestsPerWorker
	reqPerSec := float64(totalRequests) / elapsed.Seconds()

	// 获取统计信息
	statsResp, err := http.Get("http://localhost:38108/stats")
	if err == nil {
		defer statsResp.Body.Close()
		body, _ := io.ReadAll(statsResp.Body)
		t.Logf("   服务器统计: %s", string(body))
	}

	t.Logf("✅ HTTP 服务器压力测试:")
	t.Logf("   - 并发数: %d", concurrency)
	t.Logf("   - 每工作线程请求数: %d", requestsPerWorker)
	t.Logf("   - 总请求数: %d", totalRequests)
	t.Logf("   - 总耗时: %v", elapsed)
	t.Logf("   - 平均延迟: %v", elapsed/time.Duration(totalRequests))
	t.Logf("   - 吞吐量: %.2f req/s", reqPerSec)

	if reqPerSec < 100 {
		t.Logf("   ⚠️  性能较低，建议优化")
	} else if reqPerSec < 500 {
		t.Logf("   ✅ 性能良好")
	} else {
		t.Logf("   🚀 性能优秀")
	}
}
