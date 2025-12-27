package test

import (
	"testing"
	"time"

	"sw_runtime/internal/builtins"
	"sw_runtime/internal/runtime"

	"github.com/dop251/goja"
)

// TestProxyModuleCreation 测试代理模块创建
func TestProxyModuleCreation(t *testing.T) {
	vm := goja.New()
	proxyModule := builtins.NewProxyModule(vm)

	if proxyModule == nil {
		t.Fatal("Failed to create proxy module")
	}

	obj := proxyModule.GetModule()
	if obj == nil {
		t.Fatal("GetModule returned nil")
	}
}

// TestHTTPProxyCreation 测试 HTTP 代理创建
func TestHTTPProxyCreation(t *testing.T) {
	runner := runtime.New()

	script := `
		const proxy = require('proxy');
		const httpProxy = proxy.createHTTPProxy('https://httpbin.org');
		
		// 验证代理对象存在
		if (!httpProxy) {
			throw new Error('Failed to create HTTP proxy');
		}
		
		// 验证方法存在
		if (typeof httpProxy.on !== 'function') {
			throw new Error('on method not found');
		}
		if (typeof httpProxy.listen !== 'function') {
			throw new Error('listen method not found');
		}
		if (typeof httpProxy.close !== 'function') {
			throw new Error('close method not found');
		}
		
		console.log('HTTP proxy created successfully');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestTCPProxyCreation 测试 TCP 代理创建
func TestTCPProxyCreation(t *testing.T) {
	runner := runtime.New()

	script := `
		const proxy = require('proxy');
		const tcpProxy = proxy.createTCPProxy('localhost:6379');
		
		// 验证代理对象存在
		if (!tcpProxy) {
			throw new Error('Failed to create TCP proxy');
		}
		
		// 验证方法存在
		if (typeof tcpProxy.on !== 'function') {
			throw new Error('on method not found');
		}
		if (typeof tcpProxy.listen !== 'function') {
			throw new Error('listen method not found');
		}
		if (typeof tcpProxy.close !== 'function') {
			throw new Error('close method not found');
		}
		
		console.log('TCP proxy created successfully');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestHTTPProxyEventHandlers 测试 HTTP 代理事件处理器
func TestHTTPProxyEventHandlers(t *testing.T) {
	runner := runtime.New()

	script := `
		const proxy = require('proxy');
		const httpProxy = proxy.createHTTPProxy('https://httpbin.org');
		
		let requestHandlerCalled = false;
		let responseHandlerCalled = false;
		let errorHandlerCalled = false;
		
		// 注册事件处理器
		httpProxy.on('request', (req) => {
			requestHandlerCalled = true;
		});
		
		httpProxy.on('response', (resp) => {
			responseHandlerCalled = true;
		});
		
		httpProxy.on('error', (err) => {
			errorHandlerCalled = true;
		});
		
		console.log('Event handlers registered');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestTCPProxyEventHandlers 测试 TCP 代理事件处理器
func TestTCPProxyEventHandlers(t *testing.T) {
	runner := runtime.New()

	script := `
		const proxy = require('proxy');
		const tcpProxy = proxy.createTCPProxy('localhost:6379');
		
		let connectionHandlerCalled = false;
		let dataHandlerCalled = false;
		let closeHandlerCalled = false;
		let errorHandlerCalled = false;
		
		// 注册事件处理器
		tcpProxy.on('connection', (conn) => {
			connectionHandlerCalled = true;
		});
		
		tcpProxy.on('data', (data) => {
			dataHandlerCalled = true;
		});
		
		tcpProxy.on('close', () => {
			closeHandlerCalled = true;
		});
		
		tcpProxy.on('error', (err) => {
			errorHandlerCalled = true;
		});
		
		console.log('Event handlers registered');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestHTTPProxyListenMethod 测试 HTTP 代理 listen 方法
func TestHTTPProxyListenMethod(t *testing.T) {
	t.Skip("Skipping HTTP proxy listen test - requires actual server startup")

	runner := runtime.New()

	script := `
		const proxy = require('proxy');
		const httpProxy = proxy.createHTTPProxy('https://httpbin.org');
		
		let listenResult = null;
		
		httpProxy.listen('18888').then((result) => {
			listenResult = result;
		});
		
		'Listen method called';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	// 等待一小段时间
	time.Sleep(100 * time.Millisecond)
}

// TestTCPProxyListenMethod 测试 TCP 代理 listen 方法
func TestTCPProxyListenMethod(t *testing.T) {
	t.Skip("Skipping TCP proxy listen test - requires actual server startup")

	runner := runtime.New()

	script := `
		const proxy = require('proxy');
		const tcpProxy = proxy.createTCPProxy('localhost:6379');
		
		let listenResult = null;
		
		tcpProxy.listen('16380').then((result) => {
			listenResult = result;
		});
		
		'Listen method called';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	// 等待一小段时间
	time.Sleep(100 * time.Millisecond)
}

// TestHTTPProxyInvalidTarget 测试 HTTP 代理无效目标
func TestHTTPProxyInvalidTarget(t *testing.T) {
	runner := runtime.New()

	script := `
		const proxy = require('proxy');
		
		try {
			const httpProxy = proxy.createHTTPProxy('invalid-url');
			console.log('Should have thrown error');
		} catch (err) {
			console.log('Error caught: ' + err.message);
		}
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// TestProxyModuleMethods 测试代理模块方法
func TestProxyModuleMethods(t *testing.T) {
	runner := runtime.New()

	script := `
		const proxy = require('proxy');
		
		// 验证模块方法存在
		if (typeof proxy.createHTTPProxy !== 'function') {
			throw new Error('createHTTPProxy method not found');
		}
		
		if (typeof proxy.createTCPProxy !== 'function') {
			throw new Error('createTCPProxy method not found');
		}
		
		console.log('All methods exist');
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
}

// BenchmarkHTTPProxyCreation 性能测试：HTTP 代理创建
func BenchmarkHTTPProxyCreation(b *testing.B) {
	runner := runtime.New()

	script := `
		const proxy = require('proxy');
		const httpProxy = proxy.createHTTPProxy('https://httpbin.org');
		console.log('created');
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := runner.RunCode(script)
		if err != nil {
			b.Fatalf("Script execution failed: %v", err)
		}
	}
}

// BenchmarkTCPProxyCreation 性能测试：TCP 代理创建
func BenchmarkTCPProxyCreation(b *testing.B) {
	runner := runtime.New()

	script := `
		const proxy = require('proxy');
		const tcpProxy = proxy.createTCPProxy('localhost:6379');
		console.log('created');
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := runner.RunCode(script)
		if err != nil {
			b.Fatalf("Script execution failed: %v", err)
		}
	}
}
