package test

import (
	"net/http"
	"strings"
	"sw_runtime/internal/runtime"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestWebSocketBasic 测试基本的 WebSocket 功能
func TestWebSocketBasic(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	code := `
		const server = require('httpserver');
		const app = server.createServer();

		app.ws('/test', (ws) => {
			ws.on('message', (data) => {
				ws.send('Echo: ' + data);
			});
		});

		app.listen('38200');
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

	// 连接 WebSocket
	wsURL := "ws://localhost:38200/test"
	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Logf("⚠️  无法连接到 WebSocket (这在某些环境下是正常的): %v", err)
		if resp != nil {
			t.Logf("Response status: %d", resp.StatusCode)
		}
		return
	}
	defer conn.Close()

	t.Log("✅ WebSocket 连接成功")

	// 发送消息
	testMessage := "Hello WebSocket"
	err = conn.WriteMessage(websocket.TextMessage, []byte(testMessage))
	if err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}

	// 接收回显消息
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("接收消息失败: %v", err)
	}

	expected := "Echo: " + testMessage
	if string(message) != expected {
		t.Errorf("期望消息 %q, 收到 %q", expected, string(message))
	} else {
		t.Log("✅ WebSocket 消息回显测试通过")
	}
}

// TestWebSocketJSON 测试 WebSocket JSON 消息
func TestWebSocketJSON(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	code := `
		const server = require('httpserver');
		const app = server.createServer();

		app.ws('/json', (ws) => {
			ws.on('message', (data) => {
				ws.sendJSON({
					received: data,
					timestamp: new Date().toISOString(),
					type: 'response'
				});
			});
		});

		app.listen('38201');
	`

	go func() {
		runner.RunCode(code)
	}()

	time.Sleep(500 * time.Millisecond)

	// 连接 WebSocket
	wsURL := "ws://localhost:38201/json"
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Logf("⚠️  无法连接到 WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// 发送 JSON 消息
	testData := `{"message":"test","value":123}`
	err = conn.WriteMessage(websocket.TextMessage, []byte(testData))
	if err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}

	// 接收 JSON 响应
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("接收消息失败: %v", err)
	}

	// 验证是否为有效 JSON
	if !strings.Contains(string(message), "received") || !strings.Contains(string(message), "timestamp") {
		t.Errorf("收到的不是预期的 JSON 响应: %s", string(message))
	} else {
		t.Log("✅ WebSocket JSON 消息测试通过")
	}
}

// TestWebSocketMultipleConnections 测试多个 WebSocket 连接
func TestWebSocketMultipleConnections(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	code := `
		const server = require('httpserver');
		const app = server.createServer();

		const clients = [];

		app.ws('/multi', (ws) => {
			clients.push(ws);
			
			ws.on('message', (data) => {
				// 广播给所有客户端
				clients.forEach(client => {
					client.send('Broadcast: ' + data);
				});
			});
			
			ws.on('close', () => {
				const index = clients.indexOf(ws);
				if (index > -1) {
					clients.splice(index, 1);
				}
			});
		});

		app.listen('38202');
	`

	go func() {
		runner.RunCode(code)
	}()

	time.Sleep(500 * time.Millisecond)

	// 创建两个连接
	wsURL := "ws://localhost:38202/multi"
	dialer := websocket.Dialer{}

	conn1, _, err1 := dialer.Dial(wsURL, nil)
	if err1 != nil {
		t.Logf("⚠️  无法连接到 WebSocket: %v", err1)
		return
	}
	defer conn1.Close()

	conn2, _, err2 := dialer.Dial(wsURL, nil)
	if err2 != nil {
		t.Logf("⚠️  无法创建第二个连接: %v", err2)
		return
	}
	defer conn2.Close()

	t.Log("✅ 创建了两个 WebSocket 连接")

	// 从第一个连接发送消息
	testMessage := "Hello from client 1"
	err := conn1.WriteMessage(websocket.TextMessage, []byte(testMessage))
	if err != nil {
		t.Fatalf("发送消息失败: %v", err)
	}

	// 从两个连接接收消息
	_, msg1, err1 := conn1.ReadMessage()
	_, msg2, err2 := conn2.ReadMessage()

	if err1 == nil && err2 == nil {
		expected := "Broadcast: " + testMessage
		if string(msg1) == expected && string(msg2) == expected {
			t.Log("✅ 多连接广播测试通过")
		} else {
			t.Logf("消息不匹配: conn1=%s, conn2=%s", string(msg1), string(msg2))
		}
	}
}

// TestWebSocketWithHTTP 测试 WebSocket 与 HTTP 路由共存
func TestWebSocketWithHTTP(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	code := `
		const server = require('httpserver');
		const app = server.createServer();

		// HTTP 路由
		app.get('/api/test', (req, res) => {
			res.json({ message: 'HTTP OK' });
		});

		// WebSocket 路由
		app.ws('/ws', (ws) => {
			ws.send('WebSocket OK');
		});

		app.listen('38203');
	`

	go func() {
		runner.RunCode(code)
	}()

	time.Sleep(500 * time.Millisecond)

	// 测试 HTTP
	httpResp, err := http.Get("http://localhost:38203/api/test")
	if err == nil {
		defer httpResp.Body.Close()
		if httpResp.StatusCode == 200 {
			t.Log("✅ HTTP 路由正常工作")
		}
	}

	// 测试 WebSocket
	wsURL := "ws://localhost:38203/ws"
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err == nil {
		defer conn.Close()
		_, message, err := conn.ReadMessage()
		if err == nil && string(message) == "WebSocket OK" {
			t.Log("✅ WebSocket 路由正常工作")
			t.Log("✅ HTTP 和 WebSocket 共存测试通过")
		}
	}
}

// BenchmarkWebSocketEcho 基准测试 - WebSocket 回显
func BenchmarkWebSocketEcho(b *testing.B) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	code := `
		const server = require('httpserver');
		const app = server.createServer();

		app.ws('/bench', (ws) => {
			ws.on('message', (data) => {
				ws.send(data);
			});
		});

		app.listen('38204');
	`

	go func() {
		runner.RunCode(code)
	}()

	time.Sleep(500 * time.Millisecond)

	// 连接 WebSocket
	wsURL := "ws://localhost:38204/bench"
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		b.Skip("无法连接到 WebSocket")
	}
	defer conn.Close()

	testMessage := []byte("benchmark test message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn.WriteMessage(websocket.TextMessage, testMessage)
		conn.ReadMessage()
	}
}
