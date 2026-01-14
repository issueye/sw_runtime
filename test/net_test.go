package test

import (
	"testing"
	"time"

	"sw_runtime/internal/runtime"
)

// TestTCPServerCreation 测试创建 TCP 服务器
func TestTCPServerCreation(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const net = require('net');
		const server = net.createTCPServer();
		global.serverCreated = server !== null && typeof server === 'object';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	result := runner.GetValue("serverCreated")
	if result == nil || !result.ToBoolean() {
		t.Error("Expected server to be created")
	}

	t.Log("TCP server creation test passed")
}

// TestTCPServerListen 测试 TCP 服务器监听
func TestTCPServerListen(t *testing.T) {
	t.Skip("Skipping - requires actual network connection")

	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const net = require('net');
		const server = net.createTCPServer();

		server.listen('28080', () => {
			console.log('Server listening');
		}).then(() => {
			global.listenSuccess = true;
		});
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	t.Log("TCP server listen test passed")
}

// TestTCPClientConnect 测试 TCP 客户端连接函数
func TestTCPClientConnect(t *testing.T) {
	t.Skip("Skipping - requires actual network connection")

	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const net = require('net');

		net.connectTCP('localhost:28080', { timeout: 1000 })
			.then(socket => {
				global.tcpConnected = socket !== null;
			})
			.catch(err => {
				global.tcpConnected = false;
			});
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	time.Sleep(500 * time.Millisecond)
	t.Log("TCP client connect test passed")
}

// TestUDPSocketCreation 测试创建 UDP 套接字
func TestUDPSocketCreation(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const net = require('net');
		const socket = net.createUDPSocket('udp4');
		global.udpCreated = socket !== null && typeof socket === 'object';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	result := runner.GetValue("udpCreated")
	if result == nil || !result.ToBoolean() {
		t.Error("Expected socket to be created")
	}

	t.Log("UDP socket creation test passed")
}

// TestUDPSocketBind 测试 UDP 套接字绑定
func TestUDPSocketBind(t *testing.T) {
	t.Skip("Skipping - requires actual network connection")

	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const net = require('net');
		const socket = net.createUDPSocket('udp4');

		socket.bind('29090', '0.0.0.0', () => {
			console.log('UDP socket bound');
		}).then(() => {
			global.bindSuccess = true;
		});
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	t.Log("UDP socket bind test passed")
}

// TestUDPSocketSend 测试 UDP 发送消息
func TestUDPSocketSend(t *testing.T) {
	t.Skip("Skipping - requires actual network connection")

	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const net = require('net');
		const socket = net.createUDPSocket('udp4');

		socket.send('Test message', '29091', 'localhost')
			.then(() => {
				global.udpSendSuccess = true;
			})
			.catch(err => {
				global.udpSendSuccess = false;
			});
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	time.Sleep(200 * time.Millisecond)
	t.Log("UDP socket send test passed")
}

// TestNetModuleAPI 测试 net 模块 API 存在
func TestNetModuleAPI(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const net = require('net');
		global.hasCreateTCPServer = typeof net.createTCPServer === 'function';
		global.hasConnectTCP = typeof net.connectTCP === 'function';
		global.hasCreateUDPSocket = typeof net.createUDPSocket === 'function';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	result1 := runner.GetValue("hasCreateTCPServer")
	if result1 == nil || !result1.ToBoolean() {
		t.Error("createTCPServer method not found")
	}

	result2 := runner.GetValue("hasConnectTCP")
	if result2 == nil || !result2.ToBoolean() {
		t.Error("connectTCP method not found")
	}

	result3 := runner.GetValue("hasCreateUDPSocket")
	if result3 == nil || !result3.ToBoolean() {
		t.Error("createUDPSocket method not found")
	}

	t.Log("Net module API test passed")
}

// TestTCPSocketMethods 测试 TCP Socket 方法
func TestTCPSocketMethods(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const net = require('net');
		const server = net.createTCPServer();
		// TCPServer 对象有 listen, on, close 方法（write 方法在连接对象上）
		global.tcpHasListen = typeof server.listen === 'function';
		global.tcpHasOn = typeof server.on === 'function';
		global.tcpHasClose = typeof server.close === 'function';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	result1 := runner.GetValue("tcpHasListen")
	if result1 == nil || !result1.ToBoolean() {
		t.Error("listen method not found")
	}

	result2 := runner.GetValue("tcpHasOn")
	if result2 == nil || !result2.ToBoolean() {
		t.Error("on method not found")
	}

	result3 := runner.GetValue("tcpHasClose")
	if result3 == nil || !result3.ToBoolean() {
		t.Error("close method not found")
	}

	t.Log("TCP socket methods test passed")
}

// TestUDPSocketMethods 测试 UDP Socket 方法
func TestUDPSocketMethods(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	script := `
		const net = require('net');
		const socket = net.createUDPSocket('udp4');
		global.udpHasBind = typeof socket.bind === 'function';
		global.udpHasSend = typeof socket.send === 'function';
		global.udpHasOn = typeof socket.on === 'function';
		global.udpHasClose = typeof socket.close === 'function';
		global.udpHasAddress = typeof socket.address === 'function';
	`

	err := runner.RunCode(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	tests := []string{"udpHasBind", "udpHasSend", "udpHasOn", "udpHasClose", "udpHasAddress"}
	for _, name := range tests {
		result := runner.GetValue(name)
		if result == nil || !result.ToBoolean() {
			t.Errorf("%s method not found", name)
		}
	}

	t.Log("UDP socket methods test passed")
}
