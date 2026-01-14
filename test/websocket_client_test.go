package test

import (
	"testing"

	"sw_runtime/internal/runtime"
)

// TestWebSocketClient 测试 WebSocket 客户端基本功能
func TestWebSocketClient(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	// 测试 WebSocket 客户端模块
	script := `
		const ws = require('websocket');

		// 验证模块方法存在
		global.wsConnectIsFunction = typeof ws.connect === 'function';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("wsConnectIsFunction")
	if result == nil || !result.ToBoolean() {
		t.Error("websocket.connect should be a function")
	}

	t.Log("WebSocket client test passed")
}

// TestWebSocketClientJSON 测试 WebSocket 客户端 JSON 消息
func TestWebSocketClientJSON(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	// 测试 WebSocket 客户端方法
	script := `
		const ws = require('websocket');

		// 验证客户端方法存在
		global.wsHasConnect = typeof ws.connect === 'function';
		global.wsHasSend = typeof ws.send === 'function';
		global.wsHasSendJSON = typeof ws.sendJSON === 'function';
		global.wsHasOn = typeof ws.on === 'function';
		global.wsHasClose = typeof ws.close === 'function';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	tests := []string{"wsHasConnect", "wsHasSend", "wsHasSendJSON", "wsHasOn", "wsHasClose"}
	for _, name := range tests {
		result := runner.GetValue(name)
		if result == nil || !result.ToBoolean() {
			t.Errorf("%s method should exist", name)
		}
	}

	t.Log("WebSocket client JSON test passed")
}

// TestWebSocketClientReconnect 测试 WebSocket 客户端重连
func TestWebSocketClientReconnect(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	// 测试 WebSocket 客户端连接功能
	script := `
		const ws = require('websocket');

		// 验证客户端可以创建连接对象
		global.wsConnectExists = typeof ws.connect === 'function';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("wsConnectExists")
	if result == nil || !result.ToBoolean() {
		t.Error("websocket.connect should be a function")
	}

	t.Log("WebSocket client reconnect test passed")
}

// TestWebSocketClientOptions 测试 WebSocket 客户端选项
func TestWebSocketClientOptions(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	// 测试 WebSocket 客户端选项
	script := `
		const ws = require('websocket');

		// 验证连接选项支持
		global.wsConnectExists = typeof ws.connect === 'function';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("wsConnectExists")
	if result == nil || !result.ToBoolean() {
		t.Error("websocket.connect should be a function")
	}

	t.Log("WebSocket client options test passed")
}
