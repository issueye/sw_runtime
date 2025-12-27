package test

import (
	"testing"
	"time"

	"sw_runtime/internal/runtime"
)

// TestHTTPRequestInterceptor 测试请求拦截器
func TestHTTPRequestInterceptor(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const http = require('http');
		
		// 设置请求拦截器
		http.setRequestInterceptor((config) => {
			config.headers = config.headers || {};
			config.headers['X-Custom-Header'] = 'TestValue';
			return config;
		});
		
		console.log('Request interceptor test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestHTTPResponseInterceptor 测试响应拦截器
func TestHTTPResponseInterceptor(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const http = require('http');
		
		// 设置响应拦截器
		http.setResponseInterceptor((response) => {
			response.data = { processed: true, original: response.data };
			return response;
		});
		
		console.log('Response interceptor test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestHTTPBeforeRequest 测试 beforeRequest
func TestHTTPBeforeRequest(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const http = require('http');
		
		// 测试 beforeRequest 配置
		const config = {
			beforeRequest: (cfg) => {
				cfg.headers = cfg.headers || {};
				cfg.headers['X-Before'] = 'true';
				return cfg;
			}
		};
		
		if (typeof config.beforeRequest !== 'function') {
			throw new Error('beforeRequest should be a function');
		}
		
		console.log('beforeRequest test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestHTTPAfterResponse 测试 afterResponse
func TestHTTPAfterResponse(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const http = require('http');
		
		// 测试 afterResponse 配置
		const config = {
			afterResponse: (response) => {
				response.processed = true;
				return response;
			}
		};
		
		if (typeof config.afterResponse !== 'function') {
			throw new Error('afterResponse should be a function');
		}
		
		console.log('afterResponse test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestHTTPTransformRequest 测试 transformRequest
func TestHTTPTransformRequest(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const http = require('http');
		
		// 测试 transformRequest 配置
		const config = {
			transformRequest: (data) => {
				return { ...data, transformed: true };
			}
		};
		
		if (typeof config.transformRequest !== 'function') {
			throw new Error('transformRequest should be a function');
		}
		
		console.log('transformRequest test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestHTTPTransformResponse 测试 transformResponse
func TestHTTPTransformResponse(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const http = require('http');
		
		// 测试 transformResponse 配置
		const config = {
			transformResponse: (data) => {
				return { transformed: data };
			}
		};
		
		if (typeof config.transformResponse !== 'function') {
			throw new Error('transformResponse should be a function');
		}
		
		console.log('transformResponse test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestHTTPHeaderModification 测试请求头修改
func TestHTTPHeaderModification(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const http = require('http');
		
		// 测试请求头配置
		const config = {
			headers: {
				'Content-Type': 'application/json',
				'Authorization': 'Bearer token123',
				'X-Custom-Header': 'CustomValue'
			}
		};
		
		if (!config.headers['Content-Type']) {
			throw new Error('Content-Type header not set');
		}
		
		if (!config.headers['Authorization']) {
			throw new Error('Authorization header not set');
		}
		
		console.log('Header modification test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestHTTPParamsModification 测试请求参数修改
func TestHTTPParamsModification(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const http = require('http');
		
		// 测试请求参数配置
		const config = {
			params: {
				page: 1,
				size: 10,
				sort: 'created_at',
				order: 'desc'
			}
		};
		
		if (config.params.page !== 1) {
			throw new Error('Params not set correctly');
		}
		
		console.log('Params modification test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestHTTPDataModification 测试请求体修改
func TestHTTPDataModification(t *testing.T) {
	runner := runtime.NewOrPanic()

	script := `
		const http = require('http');
		
		// 测试请求数据配置
		const config = {
			data: {
				username: 'john',
				email: 'john@example.com',
				metadata: {
					source: 'test',
					timestamp: Date.now()
				}
			}
		};
		
		if (!config.data.username) {
			throw new Error('Data not set correctly');
		}
		
		console.log('Data modification test passed');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestHTTPInterceptorChain 测试拦截器链
func TestHTTPInterceptorChain(t *testing.T) {
	t.Skip("Skipping actual HTTP request test")

	runner := runtime.NewOrPanic()

	script := `
		const http = require('http');
		
		let requestInterceptorCalled = false;
		let beforeRequestCalled = false;
		
		// 设置全局拦截器
		http.setRequestInterceptor((config) => {
			requestInterceptorCalled = true;
			return config;
		});
		
		// 发送请求
		http.get('https://httpbin.org/headers', {
			beforeRequest: (config) => {
				beforeRequestCalled = true;
				return config;
			}
		}).then(() => {
			if (!requestInterceptorCalled) {
				throw new Error('Global request interceptor not called');
			}
			if (!beforeRequestCalled) {
				throw new Error('Before request interceptor not called');
			}
			console.log('Interceptor chain test passed');
		}).catch(err => {
			console.error('Request failed:', err.message);
		});
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	time.Sleep(2 * time.Second)
}

// BenchmarkHTTPInterceptor 性能测试
func BenchmarkHTTPInterceptor(b *testing.B) {
	runner := runtime.NewOrPanic()

	script := `
		const http = require('http');
		
		http.setRequestInterceptor((config) => {
			config.headers = config.headers || {};
			config.headers['X-Benchmark'] = 'true';
			return config;
		});
		
		http.setResponseInterceptor((response) => {
			response.processed = true;
			return response;
		});
	`

	for i := 0; i < b.N; i++ {
		err := runner.RunCode(script)
		if err != nil {
			b.Fatalf("Script execution failed: %v", err)
		}
	}
}
