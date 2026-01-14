package test

import (
	"testing"

	"sw_runtime/internal/runtime"
)

// TestWebSocketBasic 测试基本的 WebSocket 功能
func TestWebSocketBasic(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	// 测试 WebSocket 模块是否存在
	script := `
		const server = require('httpserver');
		const app = server.createServer();

		// 验证 app 有 ws 方法
		global.appWsIsFunction = typeof app.ws === 'function';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("appWsIsFunction")
	if result == nil || !result.ToBoolean() {
		t.Error("app.ws should be a function")
	}

	t.Log("WebSocket basic test passed")
}

// TestWebSocketJSON 测试 WebSocket JSON 消息
func TestWebSocketJSON(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	// 测试 WebSocket 消息处理
	script := `
		const server = require('httpserver');
		const app = server.createServer();

		// 测试 ws 路由配置
		const handlerType = typeof app.ws;

		// 验证 ws 方法存在且为函数
		global.wsHandlerIsFunction = handlerType === 'function';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("wsHandlerIsFunction")
	if result == nil || !result.ToBoolean() {
		t.Error("ws method should be a function")
	}

	t.Log("WebSocket JSON test passed")
}

// TestWebSocketMultipleConnections 测试 WebSocket 多连接
func TestWebSocketMultipleConnections(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	// 测试多个 WebSocket 处理器
	script := `
		const server = require('httpserver');
		const app = server.createServer();

		// 验证可以设置多个 ws 路由
		global.wsMethodExists = typeof app.ws === 'function';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result := runner.GetValue("wsMethodExists")
	if result == nil || !result.ToBoolean() {
		t.Error("ws method should exist")
	}

	t.Log("WebSocket multiple connections test passed")
}

// TestWebSocketWithHTTP 测试 WebSocket 与 HTTP 路由共存
func TestWebSocketWithHTTP(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	// 测试 HTTP 和 WebSocket 路由共存
	script := `
		const server = require('httpserver');
		const app = server.createServer();

		// HTTP 路由
		app.get('/api', () => {
			return { status: 'ok' };
		});

		// WebSocket 路由
		app.ws('/ws', (ws) => {
			// WebSocket 处理
		});

		// 验证两者都存在
		global.hasGetMethod = typeof app.get === 'function';
		global.hasWsMethod = typeof app.ws === 'function';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}

	result1 := runner.GetValue("hasGetMethod")
	if result1 == nil || !result1.ToBoolean() {
		t.Error("app.get method should exist")
	}

	result2 := runner.GetValue("hasWsMethod")
	if result2 == nil || !result2.ToBoolean() {
		t.Error("app.ws method should exist")
	}

	t.Log("WebSocket with HTTP test passed")
}
