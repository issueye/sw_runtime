package test

import (
	"testing"
	"time"

	"sw_runtime/internal/builtins"

	"github.com/dop251/goja"
)

// TestTCPServerCreation 测试创建 TCP 服务器
func TestTCPServerCreation(t *testing.T) {
	vm := goja.New()
	netModule := builtins.NewNetModule(vm)
	vm.Set("net", netModule.GetModule())

	script := `
		const server = net.createTCPServer();
		typeof server;
	`

	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	if result.String() != "object" {
		t.Errorf("Expected object, got %s", result.String())
	}

	t.Log("TCP server creation test passed")
}

// TestTCPServerListen 测试 TCP 服务器监听
func TestTCPServerListen(t *testing.T) {
	vm := goja.New()
	netModule := builtins.NewNetModule(vm)
	vm.Set("net", netModule.GetModule())

	script := `
		const server = net.createTCPServer();
		let listenCalled = false;
		
		server.listen('28080', () => {
			listenCalled = true;
		}).then(() => {
			console.log('Server listening');
		});
		
		listenCalled;
	`

	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// 初始状态应该为 false
	if result.ToBoolean() {
		t.Error("Expected listenCalled to be false initially")
	}

	t.Log("TCP server listen test passed")
}

// TestTCPClientConnect 测试 TCP 客户端连接函数
func TestTCPClientConnect(t *testing.T) {
	vm := goja.New()
	netModule := builtins.NewNetModule(vm)
	vm.Set("net", netModule.GetModule())

	script := `
		let connectResult = null;
		
		net.connectTCP('localhost:28080', { timeout: 1000 })
			.then(socket => {
				connectResult = 'connected';
			})
			.catch(err => {
				connectResult = 'failed';
			});
		
		connectResult;
	`

	_, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	time.Sleep(1500 * time.Millisecond)
	t.Log("TCP client connect test passed")
}

// TestUDPSocketCreation 测试创建 UDP 套接字
func TestUDPSocketCreation(t *testing.T) {
	vm := goja.New()
	netModule := builtins.NewNetModule(vm)
	vm.Set("net", netModule.GetModule())

	script := `
		const socket = net.createUDPSocket('udp4');
		typeof socket;
	`

	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	if result.String() != "object" {
		t.Errorf("Expected object, got %s", result.String())
	}

	t.Log("UDP socket creation test passed")
}

// TestUDPSocketBind 测试 UDP 套接字绑定
func TestUDPSocketBind(t *testing.T) {
	vm := goja.New()
	netModule := builtins.NewNetModule(vm)
	vm.Set("net", netModule.GetModule())

	script := `
		const socket = net.createUDPSocket('udp4');
		let bindCalled = false;
		
		socket.bind('29090', '0.0.0.0', () => {
			bindCalled = true;
		}).then(() => {
			console.log('UDP socket bound');
		});
		
		bindCalled;
	`

	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// 初始状态应该为 false
	if result.ToBoolean() {
		t.Error("Expected bindCalled to be false initially")
	}

	t.Log("UDP socket bind test passed")
}

// TestUDPSocketSend 测试 UDP 发送消息
func TestUDPSocketSend(t *testing.T) {
	vm := goja.New()
	netModule := builtins.NewNetModule(vm)
	vm.Set("net", netModule.GetModule())

	script := `
		const socket = net.createUDPSocket('udp4');
		let sendResult = null;
		
		socket.send('Test message\n', '29091', 'localhost')
			.then(() => {
				sendResult = 'sent';
			})
			.catch(err => {
				sendResult = 'failed';
			});
		
		sendResult;
	`

	_, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	time.Sleep(500 * time.Millisecond)
	t.Log("UDP socket send test passed")
}

// TestNetModuleAPI 测试 net 模块 API 存在
func TestNetModuleAPI(t *testing.T) {
	vm := goja.New()
	netModule := builtins.NewNetModule(vm)
	vm.Set("net", netModule.GetModule())

	script := `
		const hasCreateTCPServer = typeof net.createTCPServer === 'function';
		const hasConnectTCP = typeof net.connectTCP === 'function';
		const hasCreateUDPSocket = typeof net.createUDPSocket === 'function';
		
		({ hasCreateTCPServer, hasConnectTCP, hasCreateUDPSocket });
	`

	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	obj := result.ToObject(vm)
	if !obj.Get("hasCreateTCPServer").ToBoolean() {
		t.Error("createTCPServer method not found")
	}
	if !obj.Get("hasConnectTCP").ToBoolean() {
		t.Error("connectTCP method not found")
	}
	if !obj.Get("hasCreateUDPSocket").ToBoolean() {
		t.Error("createUDPSocket method not found")
	}

	t.Log("Net module API test passed")
}

// TestTCPSocketMethods 测试 TCP Socket 方法
func TestTCPSocketMethods(t *testing.T) {
	vm := goja.New()
	netModule := builtins.NewNetModule(vm)
	vm.Set("net", netModule.GetModule())

	script := `
		const server = net.createTCPServer();
		let socketMethods = {};
		
		server.on('connection', (socket) => {
			socketMethods.hasWrite = typeof socket.write === 'function';
			socketMethods.hasOn = typeof socket.on === 'function';
			socketMethods.hasClose = typeof socket.close === 'function';
			socketMethods.hasSetTimeout = typeof socket.setTimeout === 'function';
			socketMethods.hasRemoteAddress = typeof socket.remoteAddress === 'string';
			socketMethods.hasLocalAddress = typeof socket.localAddress === 'string';
		});
		
		socketMethods;
	`

	_, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	t.Log("TCP socket methods test passed")
}

// TestUDPSocketMethods 测试 UDP Socket 方法
func TestUDPSocketMethods(t *testing.T) {
	vm := goja.New()
	netModule := builtins.NewNetModule(vm)
	vm.Set("net", netModule.GetModule())

	script := `
		const socket = net.createUDPSocket('udp4');
		
		const hasBindMethod = typeof socket.bind === 'function';
		const hasSendMethod = typeof socket.send === 'function';
		const hasOnMethod = typeof socket.on === 'function';
		const hasCloseMethod = typeof socket.close === 'function';
		const hasAddressMethod = typeof socket.address === 'function';
		
		({ hasBindMethod, hasSendMethod, hasOnMethod, hasCloseMethod, hasAddressMethod });
	`

	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	obj := result.ToObject(vm)
	if !obj.Get("hasBindMethod").ToBoolean() {
		t.Error("bind method not found")
	}
	if !obj.Get("hasSendMethod").ToBoolean() {
		t.Error("send method not found")
	}
	if !obj.Get("hasOnMethod").ToBoolean() {
		t.Error("on method not found")
	}
	if !obj.Get("hasCloseMethod").ToBoolean() {
		t.Error("close method not found")
	}
	if !obj.Get("hasAddressMethod").ToBoolean() {
		t.Error("address method not found")
	}

	t.Log("UDP socket methods test passed")
}
