package test

import (
	"sw_runtime/internal/runtime"
	"testing"
	"time"
)

// TestWebSocketClient 测试 WebSocket 客户端基本功能
func TestWebSocketClient(t *testing.T) {
	// 启动服务器
	server := runtime.NewOrPanic()
	defer server.Close()

	serverCode := `
		const server = require('httpserver');
		const app = server.createServer();

		app.ws('/echo', (ws) => {
			ws.on('message', (data) => {
				ws.send('Echo: ' + data);
			});
		});

		app.listen('38300');
	`

	go func() {
		err := server.RunCode(serverCode)
		if err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	// 启动客户端
	client := runtime.NewOrPanic()
	defer client.Close()

	clientCode := `
		const ws = require('websocket');
		let messageReceived = false;

		ws.connect('ws://localhost:38300/echo').then(client => {
			console.log('Client connected');
			
			client.on('message', (data) => {
				console.log('Received:', data);
				if (data.includes('Echo:')) {
					messageReceived = true;
				}
				client.close();
			});

			client.send('Hello Server');
			
			// 超时保护
			setTimeout(() => {
				if (!client.isClosed()) {
					client.close();
				}
			}, 2000);
		}).catch(err => {
			console.error('Connection error:', err.message);
		});
	`

	err := client.RunCode(clientCode)
	if err != nil {
		t.Fatalf("客户端运行失败: %v", err)
	}

	// 检查消息是否接收
	messageReceived := client.GetValue("messageReceived")
	if messageReceived != nil && messageReceived.ToBoolean() {
		t.Log("✅ WebSocket 客户端测试通过")
	}
}

// TestWebSocketClientJSON 测试 WebSocket 客户端 JSON 消息
func TestWebSocketClientJSON(t *testing.T) {
	// 启动服务器
	server := runtime.NewOrPanic()
	defer server.Close()

	serverCode := `
		const server = require('httpserver');
		const app = server.createServer();

		app.ws('/json', (ws) => {
			ws.on('message', (data) => {
				if (typeof data === 'object') {
					ws.sendJSON({
						echo: data,
						timestamp: Date.now()
					});
				}
			});
		});

		app.listen('38301');
	`

	go func() {
		err := server.RunCode(serverCode)
		if err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	// 启动客户端
	client := runtime.NewOrPanic()
	defer client.Close()

	clientCode := `
		const ws = require('websocket');
		let jsonReceived = false;

		ws.connect('ws://localhost:38301/json').then(client => {
			console.log('Client connected');
			
			client.on('message', (data) => {
				console.log('Received:', JSON.stringify(data));
				if (typeof data === 'object' && data.echo) {
					jsonReceived = true;
				}
				client.close();
			});

			client.sendJSON({
				type: 'test',
				message: 'Hello JSON'
			});
			
			// 超时保护
			setTimeout(() => {
				if (!client.isClosed()) {
					client.close();
				}
			}, 2000);
		}).catch(err => {
			console.error('Connection error:', err.message);
		});
	`

	err := client.RunCode(clientCode)
	if err != nil {
		t.Fatalf("客户端运行失败: %v", err)
	}

	// 检查 JSON 消息是否接收
	jsonReceived := client.GetValue("jsonReceived")
	if jsonReceived != nil && jsonReceived.ToBoolean() {
		t.Log("✅ WebSocket 客户端 JSON 测试通过")
	}
}

// TestWebSocketClientReconnect 测试 WebSocket 客户端重连
func TestWebSocketClientReconnect(t *testing.T) {
	// 启动服务器
	server := runtime.NewOrPanic()
	defer server.Close()

	serverCode := `
		const server = require('httpserver');
		const app = server.createServer();

		app.ws('/reconnect', (ws) => {
			ws.on('message', (data) => {
				ws.send('OK');
			});
		});

		app.listen('38302');
	`

	go func() {
		err := server.RunCode(serverCode)
		if err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	// 启动客户端
	client := runtime.NewOrPanic()
	defer client.Close()

	clientCode := `
		const ws = require('websocket');
		let connectionCount = 0;

		function connect() {
			ws.connect('ws://localhost:38302/reconnect').then(client => {
				connectionCount++;
				console.log('Connected:', connectionCount);
				
				client.on('message', (data) => {
					console.log('Received:', data);
					client.close();
				});

				client.send('Test');
			}).catch(err => {
				console.error('Connection error:', err.message);
			});
		}

		// 连接两次
		connect();
		setTimeout(() => {
			connect();
		}, 1000);
	`

	err := client.RunCode(clientCode)
	if err != nil {
		t.Fatalf("客户端运行失败: %v", err)
	}

	// 检查连接次数
	connectionCount := client.GetValue("connectionCount")
	if connectionCount != nil && connectionCount.ToInteger() == 2 {
		t.Log("✅ WebSocket 客户端重连测试通过")
	}
}

// TestWebSocketClientOptions 测试 WebSocket 客户端选项
func TestWebSocketClientOptions(t *testing.T) {
	// 启动服务器
	server := runtime.NewOrPanic()
	defer server.Close()

	serverCode := `
		const server = require('httpserver');
		const app = server.createServer();

		app.ws('/options', (ws) => {
			ws.on('message', (data) => {
				ws.send('Received');
			});
		});

		app.listen('38303');
	`

	go func() {
		err := server.RunCode(serverCode)
		if err != nil {
			t.Logf("Server error: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	// 启动客户端
	client := runtime.NewOrPanic()
	defer client.Close()

	clientCode := `
		const ws = require('websocket');
		let connected = false;

		ws.connect('ws://localhost:38303/options', {
			timeout: 5000,
			headers: {
				'User-Agent': 'SW-Runtime-Test'
			}
		}).then(client => {
			connected = true;
			console.log('Connected with options');
			
			client.on('message', (data) => {
				console.log('Received:', data);
				client.close();
			});

			client.send('Test with options');
			
			// 超时保护
			setTimeout(() => {
				if (!client.isClosed()) {
					client.close();
				}
			}, 2000);
		}).catch(err => {
			console.error('Connection error:', err.message);
		});
	`

	err := client.RunCode(clientCode)
	if err != nil {
		t.Fatalf("客户端运行失败: %v", err)
	}

	// 检查连接
	connected := client.GetValue("connected")
	if connected != nil && connected.ToBoolean() {
		t.Log("✅ WebSocket 客户端选项测试通过")
	}
}

// BenchmarkWebSocketClient WebSocket 客户端性能测试
func BenchmarkWebSocketClient(b *testing.B) {
	// 启动服务器
	server := runtime.NewOrPanic()
	defer server.Close()

	serverCode := `
		const server = require('httpserver');
		const app = server.createServer();

		app.ws('/bench', (ws) => {
			ws.on('message', (data) => {
				ws.send(data);
			});
		});

		app.listen('38304');
	`

	go func() {
		server.RunCode(serverCode)
	}()

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client := runtime.NewOrPanic()

		clientCode := `
			const ws = require('websocket');

			ws.connect('ws://localhost:38304/bench').then(client => {
				client.on('message', (data) => {
					client.close();
				});
				client.send('Benchmark');
			});
		`

		client.RunCode(clientCode)
		client.Close()
	}
}
