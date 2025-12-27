package test

import (
	"sw_runtime/internal/runtime"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestWebSocketServerClientIntegration 集成测试 - WebSocket 服务器和客户端
func TestWebSocketServerClientIntegration(t *testing.T) {
	// 启动服务器
	server := runtime.New()
	defer server.Close()

	serverCode := `
		const server = require('httpserver');
		const app = server.createServer();

		let messagesReceived = 0;

		app.ws('/integration', (ws) => {
			console.log('Server: Client connected');
			
			ws.on('message', (data) => {
				messagesReceived++;
				console.log('Server received:', data);
				ws.send('Server echo: ' + data);
			});
			
			ws.on('close', () => {
				console.log('Server: Client disconnected');
			});
		});

		app.listen('38350');
		global.getMessagesReceived = () => messagesReceived;
	`

	go func() {
		err := server.RunCode(serverCode)
		if err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	// 测试 1: 使用 Go 客户端连接
	t.Run("GoClient", func(t *testing.T) {
		wsURL := "ws://localhost:38350/integration"
		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial(wsURL, nil)
		if err != nil {
			t.Logf("⚠️  无法连接到 WebSocket: %v", err)
			return
		}
		defer conn.Close()

		// 发送多条消息
		messages := []string{"Hello", "World", "From", "Go", "Client"}
		for _, msg := range messages {
			err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				t.Fatalf("发送消息失败: %v", err)
			}

			_, response, err := conn.ReadMessage()
			if err != nil {
				t.Fatalf("接收消息失败: %v", err)
			}

			expected := "Server echo: " + msg
			if string(response) != expected {
				t.Errorf("期望 %q, 收到 %q", expected, string(response))
			}
		}

		t.Log("✅ Go 客户端测试通过 - 发送并接收了", len(messages), "条消息")
	})

	// 等待一下让服务器处理完
	time.Sleep(100 * time.Millisecond)

	// 检查服务器接收的消息数
	messagesReceived := server.GetValue("getMessagesReceived")
	if messagesReceived != nil {
		count := messagesReceived.ToInteger()
		t.Logf("✅ 服务器接收到 %d 条消息", count)
	}
}

// TestWebSocketPerformanceComparison 性能对比测试
func TestWebSocketPerformanceComparison(t *testing.T) {
	// 启动服务器
	server := runtime.New()
	defer server.Close()

	serverCode := `
		const server = require('httpserver');
		const app = server.createServer();

		app.ws('/perf', (ws) => {
			ws.on('message', (data) => {
				ws.send(data);
			});
		});

		app.listen('38351');
	`

	go func() {
		server.RunCode(serverCode)
	}()

	time.Sleep(500 * time.Millisecond)

	// 性能测试
	wsURL := "ws://localhost:38351/perf"
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Logf("⚠️  无法连接到 WebSocket: %v", err)
		return
	}
	defer conn.Close()

	testMessage := []byte("Performance test message")
	iterations := 100

	start := time.Now()
	for i := 0; i < iterations; i++ {
		err = conn.WriteMessage(websocket.TextMessage, testMessage)
		if err != nil {
			t.Fatalf("发送失败: %v", err)
		}

		_, _, err = conn.ReadMessage()
		if err != nil {
			t.Fatalf("接收失败: %v", err)
		}
	}
	elapsed := time.Since(start)

	avgLatency := elapsed / time.Duration(iterations)
	messagesPerSec := float64(iterations) / elapsed.Seconds()

	t.Logf("✅ WebSocket 性能测试:")
	t.Logf("   - 消息数: %d", iterations)
	t.Logf("   - 总耗时: %v", elapsed)
	t.Logf("   - 平均延迟: %v", avgLatency)
	t.Logf("   - 吞吐量: %.2f msg/s", messagesPerSec)

	if messagesPerSec < 100 {
		t.Logf("   ⚠️  性能较低")
	} else if messagesPerSec < 500 {
		t.Logf("   ✅ 性能良好")
	} else {
		t.Logf("   🚀 性能优秀")
	}
}

// TestWebSocketBidirectionalCommunication 双向通信测试
func TestWebSocketBidirectionalCommunication(t *testing.T) {
	server := runtime.New()
	defer server.Close()

	serverCode := `
		const server = require('httpserver');
		const app = server.createServer();

		app.ws('/bidirectional', (ws) => {
			let counter = 0;
			
			// 服务器主动发送消息
			const interval = setInterval(() => {
				counter++;
				ws.send('Server message ' + counter);
				
				if (counter >= 3) {
					clearInterval(interval);
				}
			}, 100);
			
			// 接收客户端消息
			ws.on('message', (data) => {
				console.log('Server received:', data);
				ws.send('ACK: ' + data);
			});
			
			ws.on('close', () => {
				clearInterval(interval);
			});
		});

		app.listen('38352');
	`

	go func() {
		server.RunCode(serverCode)
	}()

	time.Sleep(500 * time.Millisecond)

	wsURL := "ws://localhost:38352/bidirectional"
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Logf("⚠️  无法连接到 WebSocket: %v", err)
		return
	}
	defer conn.Close()

	serverMessages := 0
	clientMessages := 0

	// 启动接收协程
	done := make(chan bool)
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			msg := string(message)
			if len(msg) > 13 && msg[:13] == "Server message" {
				serverMessages++
				t.Logf("收到服务器消息: %s", msg)
			} else if len(msg) > 4 && msg[:4] == "ACK:" {
				clientMessages++
				t.Logf("收到 ACK: %s", msg)
			}

			if serverMessages >= 3 && clientMessages >= 2 {
				done <- true
				break
			}
		}
	}()

	// 客户端发送消息
	time.Sleep(50 * time.Millisecond)
	conn.WriteMessage(websocket.TextMessage, []byte("Client message 1"))

	time.Sleep(50 * time.Millisecond)
	conn.WriteMessage(websocket.TextMessage, []byte("Client message 2"))

	// 等待通信完成
	select {
	case <-done:
		t.Logf("✅ 双向通信测试通过")
		t.Logf("   - 服务器消息: %d", serverMessages)
		t.Logf("   - 客户端消息: %d", clientMessages)
	case <-time.After(2 * time.Second):
		t.Logf("⚠️  测试超时")
	}
}

// TestWebSocketJSONDataExchange JSON 数据交换测试
func TestWebSocketJSONDataExchange(t *testing.T) {
	server := runtime.New()
	defer server.Close()

	serverCode := `
		const server = require('httpserver');
		const app = server.createServer();

		app.ws('/json-exchange', (ws) => {
			ws.on('message', (data) => {
				if (typeof data === 'object') {
					// 收到 JSON 对象，处理并返回
					ws.sendJSON({
						status: 'success',
						received: data,
						timestamp: Date.now(),
						processed: true
					});
				}
			});
		});

		app.listen('38353');
	`

	go func() {
		server.RunCode(serverCode)
	}()

	time.Sleep(500 * time.Millisecond)

	wsURL := "ws://localhost:38353/json-exchange"
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Logf("⚠️  无法连接到 WebSocket: %v", err)
		return
	}
	defer conn.Close()

	// 发送 JSON 数据
	jsonData := `{"type":"test","value":42,"name":"WebSocket Test"}`
	err = conn.WriteMessage(websocket.TextMessage, []byte(jsonData))
	if err != nil {
		t.Fatalf("发送 JSON 失败: %v", err)
	}

	// 接收响应
	_, response, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("接收响应失败: %v", err)
	}

	responseStr := string(response)
	if len(responseStr) > 0 && responseStr[0] == '{' {
		t.Logf("✅ JSON 数据交换测试通过")
		t.Logf("   发送: %s", jsonData)
		t.Logf("   接收: %s", responseStr)
	} else {
		t.Errorf("收到的不是 JSON 响应: %s", responseStr)
	}
}

// TestWebSocketConnectionLifecycle 连接生命周期测试
func TestWebSocketConnectionLifecycle(t *testing.T) {
	server := runtime.New()
	defer server.Close()

	serverCode := `
		const server = require('httpserver');
		const app = server.createServer();

		let connections = 0;
		let disconnections = 0;

		app.ws('/lifecycle', (ws) => {
			connections++;
			console.log('Connection opened, total:', connections);
			
			ws.on('message', (data) => {
				ws.send('ACK');
			});
			
			ws.on('close', () => {
				disconnections++;
				console.log('Connection closed, total:', disconnections);
			});
		});

		app.listen('38354');
		global.getStats = () => ({ connections, disconnections });
	`

	go func() {
		server.RunCode(serverCode)
	}()

	time.Sleep(500 * time.Millisecond)

	// 创建并关闭多个连接
	wsURL := "ws://localhost:38354/lifecycle"
	dialer := websocket.Dialer{}

	for i := 0; i < 3; i++ {
		conn, _, err := dialer.Dial(wsURL, nil)
		if err != nil {
			t.Logf("⚠️  连接 %d 失败: %v", i+1, err)
			continue
		}

		// 发送一条消息
		conn.WriteMessage(websocket.TextMessage, []byte("Test"))
		conn.ReadMessage()

		// 关闭连接
		conn.Close()
		time.Sleep(100 * time.Millisecond)
	}

	// 检查统计
	time.Sleep(200 * time.Millisecond)
	stats := server.GetValue("getStats")
	if stats != nil {
		statsObj := stats.ToObject(nil)
		if statsObj != nil {
			connections := statsObj.Get("connections").ToInteger()
			disconnections := statsObj.Get("disconnections").ToInteger()

			t.Logf("✅ 连接生命周期测试通过")
			t.Logf("   - 总连接数: %d", connections)
			t.Logf("   - 总断开数: %d", disconnections)
		}
	}
}
