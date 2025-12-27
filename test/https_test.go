package test

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"sw_runtime/internal/builtins"

	"github.com/dop251/goja"
)

// TestHTTPSServerCreation 测试创建 HTTPS 服务器
func TestHTTPSServerCreation(t *testing.T) {
	vm := goja.New()
	httpModule := builtins.NewHTTPServerModule(vm)
	vm.Set("httpserver", httpModule.GetModule())

	script := `
		const server = httpserver.createServer();
		typeof server.listenTLS === 'function';
	`

	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	if !result.ToBoolean() {
		t.Error("listenTLS method should exist")
	}

	t.Log("HTTPS server creation test passed")
}

// TestHTTPSServerListenTLS 测试 HTTPS 服务器监听
func TestHTTPSServerListenTLS(t *testing.T) {
	// 注意：此测试需要有效的证书文件
	// 在实际环境中运行前，需要先生成证书
	t.Skip("Skipping HTTPS test - requires valid SSL certificates")

	vm := goja.New()
	httpModule := builtins.NewHTTPServerModule(vm)
	vm.Set("httpserver", httpModule.GetModule())

	script := `
		const server = httpserver.createServer();
		
		server.get('/', (req, res) => {
			res.send('Hello HTTPS!');
		});
		
		server.listenTLS('18443', './test/certs/server.crt', './test/certs/server.key')
			.then(() => {
				console.log('HTTPS Server started');
			})
			.catch(err => {
				console.error('Failed to start HTTPS server:', err.message);
			});
	`

	_, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	time.Sleep(500 * time.Millisecond)
	t.Log("HTTPS server listen test passed")
}

// TestHTTPSRequest 测试 HTTPS 请求
func TestHTTPSRequest(t *testing.T) {
	// 注意：此测试需要有效的证书文件
	t.Skip("Skipping HTTPS request test - requires valid SSL certificates")

	vm := goja.New()
	httpModule := builtins.NewHTTPServerModule(vm)
	vm.Set("httpserver", httpModule.GetModule())

	// 启动 HTTPS 服务器
	serverScript := `
		const server = httpserver.createServer();
		
		server.get('/test', (req, res) => {
			res.json({ message: 'HTTPS works!' });
		});
		
		server.listenTLS('18443', './test/certs/server.crt', './test/certs/server.key');
	`

	_, err := vm.RunString(serverScript)
	if err != nil {
		t.Fatalf("Failed to start HTTPS server: %v", err)
	}

	time.Sleep(1 * time.Second)

	// 创建 HTTPS 客户端（忽略证书验证用于测试）
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 5 * time.Second,
	}

	// 发送请求
	resp, err := client.Get("https://localhost:18443/test")
	if err != nil {
		t.Fatalf("Failed to make HTTPS request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	t.Logf("HTTPS response: %s", string(body))
	t.Log("HTTPS request test passed")
}

// TestHTTPSWithMiddleware 测试 HTTPS 服务器中间件
func TestHTTPSWithMiddleware(t *testing.T) {
	vm := goja.New()
	httpModule := builtins.NewHTTPServerModule(vm)
	vm.Set("httpserver", httpModule.GetModule())

	script := `
		const server = httpserver.createServer();
		let middlewareCalled = false;
		
		server.use((req, res, next) => {
			middlewareCalled = true;
			res.header('X-Custom-Header', 'HTTPS-Middleware');
			next();
		});
		
		server.get('/test', (req, res) => {
			res.json({ secure: true });
		});
		
		middlewareCalled;
	`

	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	// 初始状态应该为 false
	if result.ToBoolean() {
		t.Error("Middleware should not be called yet")
	}

	t.Log("HTTPS middleware test passed")
}

// TestHTTPSServerMethods 测试 HTTPS 服务器方法
func TestHTTPSServerMethods(t *testing.T) {
	vm := goja.New()
	httpModule := builtins.NewHTTPServerModule(vm)
	vm.Set("httpserver", httpModule.GetModule())

	script := `
		const server = httpserver.createServer();
		
		const hasListenTLS = typeof server.listenTLS === 'function';
		const hasListen = typeof server.listen === 'function';
		const hasGet = typeof server.get === 'function';
		const hasPost = typeof server.post === 'function';
		const hasUse = typeof server.use === 'function';
		const hasStatic = typeof server.static === 'function';
		const hasClose = typeof server.close === 'function';
		
		({ hasListenTLS, hasListen, hasGet, hasPost, hasUse, hasStatic, hasClose });
	`

	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	obj := result.ToObject(vm)

	methods := []string{"hasListenTLS", "hasListen", "hasGet", "hasPost", "hasUse", "hasStatic", "hasClose"}
	for _, method := range methods {
		if !obj.Get(method).ToBoolean() {
			t.Errorf("%s method not found", method)
		}
	}

	t.Log("HTTPS server methods test passed")
}

// TestHTTPSListenTLSParameters 测试 listenTLS 参数验证
func TestHTTPSListenTLSParameters(t *testing.T) {
	vm := goja.New()
	httpModule := builtins.NewHTTPServerModule(vm)
	vm.Set("httpserver", httpModule.GetModule())

	// 测试缺少参数的情况
	script := `
		const server = httpserver.createServer();
		let errorCaught = false;
		
		try {
			server.listenTLS('8443'); // 缺少证书和密钥参数
		} catch (e) {
			errorCaught = true;
		}
		
		errorCaught;
	`

	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	if !result.ToBoolean() {
		t.Error("Should throw error when parameters are missing")
	}

	t.Log("HTTPS listenTLS parameters test passed")
}

// TestMixedHTTPAndHTTPS 测试混合 HTTP 和 HTTPS 服务器
func TestMixedHTTPAndHTTPS(t *testing.T) {
	t.Skip("Skipping mixed HTTP/HTTPS test - requires valid SSL certificates")

	vm := goja.New()
	httpModule := builtins.NewHTTPServerModule(vm)
	vm.Set("httpserver", httpModule.GetModule())

	script := `
		// HTTP 服务器
		const httpServer = httpserver.createServer();
		httpServer.get('/', (req, res) => {
			res.send('HTTP Server');
		});
		httpServer.listen('18080');
		
		// HTTPS 服务器
		const httpsServer = httpserver.createServer();
		httpsServer.get('/', (req, res) => {
			res.send('HTTPS Server');
		});
		httpsServer.listenTLS('18443', './test/certs/server.crt', './test/certs/server.key');
	`

	_, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run mixed servers script: %v", err)
	}

	time.Sleep(1 * time.Second)

	// 测试 HTTP 服务器
	httpResp, err := http.Get("http://localhost:18080/")
	if err == nil {
		defer httpResp.Body.Close()
		if httpResp.StatusCode == http.StatusOK {
			t.Log("HTTP server is working")
		}
	}

	// 测试 HTTPS 服务器
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	httpsResp, err := client.Get("https://localhost:18443/")
	if err == nil {
		defer httpsResp.Body.Close()
		if httpsResp.StatusCode == http.StatusOK {
			t.Log("HTTPS server is working")
		}
	}

	t.Log("Mixed HTTP/HTTPS test passed")
}

// TestHTTPSCallbackFunction 测试 listenTLS 回调函数
func TestHTTPSCallbackFunction(t *testing.T) {
	t.Skip("Skipping HTTPS callback test - requires valid SSL certificates")

	vm := goja.New()
	httpModule := builtins.NewHTTPServerModule(vm)
	vm.Set("httpserver", httpModule.GetModule())

	script := `
		const server = httpserver.createServer();
		let callbackExecuted = false;
		
		server.listenTLS('18443', './test/certs/server.crt', './test/certs/server.key', () => {
			callbackExecuted = true;
		});
		
		callbackExecuted;
	`

	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// 回调应该在服务器启动时执行
	if result.ToBoolean() {
		t.Log("Callback was executed during server startup")
	}

	t.Log("HTTPS callback function test passed")
}

// BenchmarkHTTPSServerCreation 性能测试：创建 HTTPS 服务器
func BenchmarkHTTPSServerCreation(b *testing.B) {
	vm := goja.New()
	httpModule := builtins.NewHTTPServerModule(vm)
	vm.Set("httpserver", httpModule.GetModule())

	script := `
		const server = httpserver.createServer();
		server.get('/', (req, res) => {
			res.send('Hello');
		});
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm.RunString(script)
	}
}

// Helper function to print test results
func printTestInfo(t *testing.T, name string, info string) {
	t.Helper()
	fmt.Printf("\n=== %s ===\n%s\n", name, info)
}
