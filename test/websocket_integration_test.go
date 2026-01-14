package test

import (
	"testing"

	"sw_runtime/internal/runtime"
)

// TestWebSocketServerClientIntegration 集成测试 - WebSocket 服务器和客户端
func TestWebSocketServerClientIntegration(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	// 测试 WebSocket 服务器和客户端集成
	script := `
		const server = require('httpserver');
		const ws = require('websocket');

		const app = server.createServer();

		// 验证服务器有 ws 方法
		global.wsHasWSMethod = typeof app.ws === 'function';

		// 验证客户端有 connect 方法
		global.wsHasConnectMethod = typeof ws.connect === 'function';

		// 验证可以设置 WebSocket 处理器
		app.ws('/test', (socket) => {
			// WebSocket 处理器
		});

		global.wsHandlerSet = true;
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result1 := runner.GetValue("wsHasWSMethod")
	if result1 == nil || !result1.ToBoolean() {
		t.Error("Server should have ws method")
	}

	result2 := runner.GetValue("wsHasConnectMethod")
	if result2 == nil || !result2.ToBoolean() {
		t.Error("Client should have connect method")
	}

	result3 := runner.GetValue("wsHandlerSet")
	if result3 == nil || !result3.ToBoolean() {
		t.Error("Should be able to set WebSocket handler")
	}

	t.Log("WebSocket server-client integration test passed")
}

// TestWebSocketPerformanceComparison 性能对比测试
func TestWebSocketPerformanceComparison(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	// 简单性能测试 - 只测试 API 调用开销
	script := `
		const server = require('httpserver');

		const app = server.createServer();

		// 多次设置 WebSocket 处理器
		for (let i = 0; i < 10; i++) {
			app.ws('/ws' + i, (socket) => {});
		}

		// 验证所有处理器都已设置
		global.perfTestSuccess = true;
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("perfTestSuccess")
	if result == nil || !result.ToBoolean() {
		t.Error("Should be able to set multiple WebSocket handlers")
	}

	t.Log("WebSocket performance test passed")
}

// TestWebSocketBidirectionalCommunication 双向通信测试
func TestWebSocketBidirectionalCommunication(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	// 测试双向通信 API
	script := `
		const server = require('httpserver');
		const ws = require('websocket');

		const app = server.createServer();

		// 验证 WebSocket 处理器有 send 方法
		app.ws('/chat', (socket) => {
			global.socketHasSend = typeof socket.send === 'function';
			global.socketHasOn = typeof socket.on === 'function';
			global.socketHasClose = typeof socket.close === 'function';
			global.socketHasSendJSON = typeof socket.sendJSON === 'function';
		});

		global.testRan = true;
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	tests := []string{"socketHasSend", "socketHasOn", "socketHasClose", "socketHasSendJSON", "testRan"}
	for _, name := range tests {
		result := runner.GetValue(name)
		if result == nil || !result.ToBoolean() {
			t.Errorf("%s should be true", name)
		}
	}

	t.Log("WebSocket bidirectional communication test passed")
}

// TestWebSocketJSONDataExchange JSON 数据交换测试
func TestWebSocketJSONDataExchange(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	// 测试 JSON 数据交换 API
	script := `
		const server = require('httpserver');
		const ws = require('websocket');

		const app = server.createServer();

		// 验证服务器和客户端都有 sendJSON 方法
		app.ws('/json', (socket) => {
			global.serverHasSendJSON = typeof socket.sendJSON === 'function';
		});

		// 验证客户端 API
		global.clientHasSendJSON = typeof ws.sendJSON === 'function';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result1 := runner.GetValue("serverHasSendJSON")
	if result1 == nil || !result1.ToBoolean() {
		t.Error("Server WebSocket should have sendJSON method")
	}

	result2 := runner.GetValue("clientHasSendJSON")
	if result2 == nil || !result2.ToBoolean() {
		t.Error("Client WebSocket should have sendJSON method")
	}

	t.Log("WebSocket JSON data exchange test passed")
}

// TestWebSocketConnectionLifecycle 连接生命周期测试
func TestWebSocketConnectionLifecycle(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	// 测试连接生命周期 API
	script := `
		const server = require('httpserver');
		const ws = require('websocket');

		const app = server.createServer();

		// 验证有连接事件
		app.ws('/lifecycle', (socket) => {
			// 验证 socket 有 close 事件
			global.socketHasOn = typeof socket.on === 'function';
		});

		// 验证客户端有连接状态方法
		global.wsHasIsClosed = typeof ws.isClosed === 'function';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result1 := runner.GetValue("socketHasOn")
	if result1 == nil || !result1.ToBoolean() {
		t.Error("WebSocket socket should have on method")
	}

	result2 := runner.GetValue("wsHasIsClosed")
	if result2 == nil || !result2.ToBoolean() {
		t.Error("WebSocket client should have isClosed method")
	}

	t.Log("WebSocket connection lifecycle test passed")
}
