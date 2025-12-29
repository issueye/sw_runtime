package test

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/dop251/goja"

	"sw_runtime/internal/builtins"
)

// TestHTTPServerSamePathDifferentMethods 测试相同路径不同 HTTP 方法
func TestHTTPServerSamePathDifferentMethods(t *testing.T) {
	vm := goja.New()
	httpModule := builtins.NewHTTPServerModule(vm)
	vm.Set("httpserver", httpModule.GetModule())

	script := `
		const server = httpserver.createServer();

		// 注册相同路径的不同方法
		server.get('/api/user', (req, res) => {
			res.json({ method: 'GET', message: 'Get user' });
		});

		server.post('/api/user', (req, res) => {
			res.json({ method: 'POST', message: 'Create user' });
		});

		server.put('/api/user', (req, res) => {
			res.json({ method: 'PUT', message: 'Update user' });
		});

		server.delete('/api/user', (req, res) => {
			res.json({ method: 'DELETE', message: 'Delete user' });
		});

		server.patch('/api/user', (req, res) => {
			res.json({ method: 'PATCH', message: 'Patch user' });
		});

		server.listen('38910');
	`

	_, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("创建服务器失败: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	tests := []struct {
		method         string
		expectedStatus int
		expectContent  string
	}{
		{"GET", http.StatusOK, `"method":"GET"`},
		{"POST", http.StatusOK, `"method":"POST"`},
		{"PUT", http.StatusOK, `"method":"PUT"`},
		{"DELETE", http.StatusOK, `"method":"DELETE"`},
		{"PATCH", http.StatusOK, `"method":"PATCH"`},
		{"HEAD", http.StatusMethodNotAllowed, "Method not allowed"},
		{"OPTIONS", http.StatusMethodNotAllowed, "Method not allowed"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "http://localhost:38910/api/user", nil)
			if err != nil {
				t.Fatalf("创建请求失败: %v", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("请求失败: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("状态码错误: 期望 %d, 实际 %d", tt.expectedStatus, resp.StatusCode)
			}

			body, _ := io.ReadAll(resp.Body)
			bodyStr := string(body)

			if tt.expectedStatus == http.StatusOK {
				// HEAD 请求的 body 为空是正常的
				if tt.method != "HEAD" {
					if !contains(bodyStr, tt.expectContent) {
						t.Errorf("响应内容错误: 期望包含 %s, 实际 %s", tt.expectContent, bodyStr)
					}
				}
				t.Logf("✅ %s 请求成功: %s", tt.method, bodyStr)
			} else {
				// 405 的 body 可能为空（HEAD 请求）
				if tt.method != "HEAD" && !contains(bodyStr, tt.expectContent) {
					t.Errorf("响应内容错误: 期望包含 %s, 实际 %s", tt.expectContent, bodyStr)
				}
				t.Logf("✅ %s 请求正确返回 405", tt.method)
			}
		})
	}

	t.Log("✅ 相同路径不同方法测试通过")
}

// TestHTTPServerMultiplePathsMultipleMethods 测试多个路径多个方法
func TestHTTPServerMultiplePathsMultipleMethods(t *testing.T) {
	vm := goja.New()
	httpModule := builtins.NewHTTPServerModule(vm)
	vm.Set("httpserver", httpModule.GetModule())

	script := `
		const server = httpserver.createServer();

		// 路径 1: /api/users
		server.get('/api/users', (req, res) => {
			res.json({ path: '/api/users', method: 'GET' });
		});

		server.post('/api/users', (req, res) => {
			res.json({ path: '/api/users', method: 'POST' });
		});

		// 路径 2: /api/products
		server.get('/api/products', (req, res) => {
			res.json({ path: '/api/products', method: 'GET' });
		});

		server.post('/api/products', (req, res) => {
			res.json({ path: '/api/products', method: 'POST' });
		});

		// 路径 3: /api/orders
		server.get('/api/orders', (req, res) => {
			res.json({ path: '/api/orders', method: 'GET' });
		});

		server.post('/api/orders', (req, res) => {
			res.json({ path: '/api/orders', method: 'POST' });
		});

		server.listen('38911');
	`

	_, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("创建服务器失败: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	tests := []struct {
		method string
		path   string
		expect string
	}{
		{"GET", "/api/users", `"path":"/api/users"`},
		{"POST", "/api/users", `"path":"/api/users"`},
		{"GET", "/api/products", `"path":"/api/products"`},
		{"POST", "/api/products", `"path":"/api/products"`},
		{"GET", "/api/orders", `"path":"/api/orders"`},
		{"POST", "/api/orders", `"path":"/api/orders"`},
	}

	for _, tt := range tests {
		t.Run(tt.method+"_"+tt.path, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "http://localhost:38911"+tt.path, nil)
			if err != nil {
				t.Fatalf("创建请求失败: %v", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("请求失败: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("状态码错误: 期望 200, 实际 %d", resp.StatusCode)
			}

			body, _ := io.ReadAll(resp.Body)
			bodyStr := string(body)

			if !contains(bodyStr, tt.expect) {
				t.Errorf("响应内容错误: 期望包含 %s, 实际 %s", tt.expect, bodyStr)
			}

			t.Logf("✅ %s %s 成功", tt.method, tt.path)
		})
	}

	t.Log("✅ 多路径多方法测试通过")
}

// contains 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
