package builtins

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/dop251/goja"
)

// NetModule 网络模块
type NetModule struct {
	vm          *goja.Runtime
	connections map[string]net.Conn
	listeners   map[string]net.Listener
	mutex       sync.RWMutex
	connID      int
}

// NewNetModule 创建网络模块
func NewNetModule(vm *goja.Runtime) *NetModule {
	return &NetModule{
		vm:          vm,
		connections: make(map[string]net.Conn),
		listeners:   make(map[string]net.Listener),
	}
}

// GetModule 获取网络模块对象
func (n *NetModule) GetModule() *goja.Object {
	obj := n.vm.NewObject()

	// TCP 方法
	obj.Set("createTCPServer", n.createTCPServer)
	obj.Set("connectTCP", n.connectTCP)

	// UDP 方法
	obj.Set("createUDPSocket", n.createUDPSocket)

	return obj
}

// TCPServer TCP 服务器
type TCPServer struct {
	vm       *goja.Runtime
	listener net.Listener
	module   *NetModule
	handlers map[string]goja.Value // 事件处理器
	mutex    sync.RWMutex
}

// createTCPServer 创建 TCP 服务器
func (n *NetModule) createTCPServer(call goja.FunctionCall) goja.Value {
	server := &TCPServer{
		vm:       n.vm,
		module:   n,
		handlers: make(map[string]goja.Value),
	}

	obj := n.vm.NewObject()

	// 监听端口
	obj.Set("listen", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(n.vm.NewTypeError("listen requires port"))
		}

		port := call.Arguments[0].String()
		if port[0] != ':' {
			port = ":" + port
		}

		var callback goja.Value
		if len(call.Arguments) > 1 {
			if _, ok := goja.AssertFunction(call.Arguments[1]); ok {
				callback = call.Arguments[1]
			}
		}

		promise, resolve, reject := n.vm.NewPromise()

		go func() {
			listener, err := net.Listen("tcp", port)
			if err != nil {
				reject(n.vm.NewGoError(err))
				return
			}

			server.listener = listener
			listenerID := fmt.Sprintf("tcp_listener_%d", n.getNextConnID())
			n.mutex.Lock()
			n.listeners[listenerID] = listener
			n.mutex.Unlock()

			// 调用回调
			if callback != nil {
				if fn, ok := goja.AssertFunction(callback); ok {
					fn(goja.Undefined())
				}
			}

			resolve(n.vm.ToValue(fmt.Sprintf("TCP Server listening on %s", port)))

			// 开始接受连接
			for {
				conn, err := listener.Accept()
				if err != nil {
					// 检查是否是因为关闭
					if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
						break
					}
					continue
				}

				// 触发 connection 事件
				go server.handleConnection(conn)
			}
		}()

		return n.vm.ToValue(promise)
	})

	// 注册事件处理器
	obj.Set("on", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(n.vm.NewTypeError("on requires event and handler"))
		}

		event := call.Arguments[0].String()
		handler := call.Arguments[1]

		server.mutex.Lock()
		server.handlers[event] = handler
		server.mutex.Unlock()

		return obj
	})

	// 关闭服务器
	obj.Set("close", func(call goja.FunctionCall) goja.Value {
		promise, resolve, reject := n.vm.NewPromise()

		go func() {
			if server.listener != nil {
				err := server.listener.Close()
				if err != nil {
					reject(n.vm.NewGoError(err))
					return
				}
			}
			resolve(n.vm.ToValue("Server closed"))
		}()

		return n.vm.ToValue(promise)
	})

	return obj
}

// handleConnection 处理新连接
func (s *TCPServer) handleConnection(conn net.Conn) {
	// 创建连接对象
	connObj := s.createTCPConnection(conn)

	// 触发 connection 事件
	s.mutex.RLock()
	handler := s.handlers["connection"]
	s.mutex.RUnlock()

	if handler != nil {
		if fn, ok := goja.AssertFunction(handler); ok {
			fn(goja.Undefined(), connObj)
		}
	}
}

// createTCPConnection 创建 TCP 连接对象
func (s *TCPServer) createTCPConnection(conn net.Conn) goja.Value {
	obj := s.vm.NewObject()
	handlers := make(map[string]goja.Value)
	var handlersMutex sync.RWMutex

	connID := fmt.Sprintf("tcp_conn_%d", s.module.getNextConnID())
	s.module.mutex.Lock()
	s.module.connections[connID] = conn
	s.module.mutex.Unlock()

	// 连接信息
	obj.Set("remoteAddress", conn.RemoteAddr().String())
	obj.Set("localAddress", conn.LocalAddr().String())

	// 发送数据
	obj.Set("write", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(s.vm.NewTypeError("write requires data"))
		}

		data := call.Arguments[0].String()
		promise, resolve, reject := s.vm.NewPromise()

		go func() {
			_, err := conn.Write([]byte(data))
			if err != nil {
				reject(s.vm.NewGoError(err))
				return
			}
			resolve(s.vm.ToValue(true))
		}()

		return s.vm.ToValue(promise)
	})

	// 注册事件
	obj.Set("on", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(s.vm.NewTypeError("on requires event and handler"))
		}

		event := call.Arguments[0].String()
		handler := call.Arguments[1]

		handlersMutex.Lock()
		handlers[event] = handler
		handlersMutex.Unlock()

		return obj
	})

	// 关闭连接
	obj.Set("close", func(call goja.FunctionCall) goja.Value {
		conn.Close()
		s.module.mutex.Lock()
		delete(s.module.connections, connID)
		s.module.mutex.Unlock()

		// 触发 close 事件
		handlersMutex.RLock()
		closeHandler := handlers["close"]
		handlersMutex.RUnlock()

		if closeHandler != nil {
			if fn, ok := goja.AssertFunction(closeHandler); ok {
				fn(goja.Undefined())
			}
		}

		return goja.Undefined()
	})

	// 设置超时
	obj.Set("setTimeout", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(s.vm.NewTypeError("setTimeout requires timeout"))
		}

		timeout := time.Duration(call.Arguments[0].ToInteger()) * time.Millisecond
		conn.SetDeadline(time.Now().Add(timeout))
		return obj
	})

	// 开始读取数据
	go func() {
		reader := bufio.NewReader(conn)
		for {
			data, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					// 触发 error 事件
					handlersMutex.RLock()
					errorHandler := handlers["error"]
					handlersMutex.RUnlock()

					if errorHandler != nil {
						if fn, ok := goja.AssertFunction(errorHandler); ok {
							errObj := s.vm.NewObject()
							errObj.Set("message", err.Error())
							fn(goja.Undefined(), errObj)
						}
					}
				}
				conn.Close()
				break
			}

			// 触发 data 事件
			handlersMutex.RLock()
			dataHandler := handlers["data"]
			handlersMutex.RUnlock()

			if dataHandler != nil {
				if fn, ok := goja.AssertFunction(dataHandler); ok {
					fn(goja.Undefined(), s.vm.ToValue(string(data)))
				}
			}
		}

		// 触发 close 事件
		handlersMutex.RLock()
		closeHandler := handlers["close"]
		handlersMutex.RUnlock()

		if closeHandler != nil {
			if fn, ok := goja.AssertFunction(closeHandler); ok {
				fn(goja.Undefined())
			}
		}
	}()

	return obj
}

// connectTCP 连接到 TCP 服务器
func (n *NetModule) connectTCP(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(n.vm.NewTypeError("connectTCP requires host:port"))
	}

	address := call.Arguments[0].String()

	var options map[string]interface{}
	timeout := 10 * time.Second

	if len(call.Arguments) > 1 && call.Arguments[1] != goja.Undefined() {
		optObj := call.Arguments[1].Export()
		if opt, ok := optObj.(map[string]interface{}); ok {
			options = opt
			if t, ok := options["timeout"].(int64); ok {
				timeout = time.Duration(t) * time.Millisecond
			}
		}
	}

	promise, resolve, reject := n.vm.NewPromise()

	go func() {
		conn, err := net.DialTimeout("tcp", address, timeout)
		if err != nil {
			reject(n.vm.NewGoError(err))
			return
		}

		// 创建连接对象
		server := &TCPServer{vm: n.vm, module: n}
		connObj := server.createTCPConnection(conn)

		resolve(connObj)
	}()

	return n.vm.ToValue(promise)
}

// UDPSocket UDP 套接字
type UDPSocket struct {
	vm       *goja.Runtime
	conn     *net.UDPConn
	module   *NetModule
	handlers map[string]goja.Value
	mutex    sync.RWMutex
}

// createUDPSocket 创建 UDP 套接字
func (n *NetModule) createUDPSocket(call goja.FunctionCall) goja.Value {
	socketType := "udp4"
	if len(call.Arguments) > 0 {
		socketType = call.Arguments[0].String()
	}

	socket := &UDPSocket{
		vm:       n.vm,
		module:   n,
		handlers: make(map[string]goja.Value),
	}

	obj := n.vm.NewObject()

	// 绑定端口
	obj.Set("bind", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(n.vm.NewTypeError("bind requires port"))
		}

		port := call.Arguments[0].String()
		host := "0.0.0.0"
		if len(call.Arguments) > 1 {
			host = call.Arguments[1].String()
		}

		var callback goja.Value
		if len(call.Arguments) > 2 {
			if _, ok := goja.AssertFunction(call.Arguments[2]); ok {
				callback = call.Arguments[2]
			}
		}

		promise, resolve, reject := n.vm.NewPromise()

		go func() {
			addr, err := net.ResolveUDPAddr(socketType, host+":"+port)
			if err != nil {
				reject(n.vm.NewGoError(err))
				return
			}

			conn, err := net.ListenUDP(socketType, addr)
			if err != nil {
				reject(n.vm.NewGoError(err))
				return
			}

			socket.conn = conn

			// 调用回调
			if callback != nil {
				if fn, ok := goja.AssertFunction(callback); ok {
					fn(goja.Undefined())
				}
			}

			resolve(n.vm.ToValue(fmt.Sprintf("UDP Socket bound to %s:%s", host, port)))

			// 开始接收数据
			socket.startReceiving()
		}()

		return n.vm.ToValue(promise)
	})

	// 发送数据
	obj.Set("send", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			panic(n.vm.NewTypeError("send requires data, port, host"))
		}

		data := call.Arguments[0].String()
		port := call.Arguments[1].String()
		host := call.Arguments[2].String()

		var callback goja.Value
		if len(call.Arguments) > 3 {
			if _, ok := goja.AssertFunction(call.Arguments[3]); ok {
				callback = call.Arguments[3]
			}
		}

		promise, resolve, reject := n.vm.NewPromise()

		go func() {
			addr, err := net.ResolveUDPAddr(socketType, host+":"+port)
			if err != nil {
				reject(n.vm.NewGoError(err))
				return
			}

			if socket.conn == nil {
				// 如果没有绑定，创建临时连接
				conn, err := net.DialUDP(socketType, nil, addr)
				if err != nil {
					reject(n.vm.NewGoError(err))
					return
				}
				defer conn.Close()

				_, err = conn.Write([]byte(data))
				if err != nil {
					reject(n.vm.NewGoError(err))
					return
				}
			} else {
				_, err = socket.conn.WriteToUDP([]byte(data), addr)
				if err != nil {
					reject(n.vm.NewGoError(err))
					return
				}
			}

			// 调用回调
			if callback != nil {
				if fn, ok := goja.AssertFunction(callback); ok {
					fn(goja.Undefined())
				}
			}

			resolve(n.vm.ToValue(true))
		}()

		return n.vm.ToValue(promise)
	})

	// 注册事件
	obj.Set("on", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(n.vm.NewTypeError("on requires event and handler"))
		}

		event := call.Arguments[0].String()
		handler := call.Arguments[1]

		socket.mutex.Lock()
		socket.handlers[event] = handler
		socket.mutex.Unlock()

		return obj
	})

	// 关闭套接字
	obj.Set("close", func(call goja.FunctionCall) goja.Value {
		if socket.conn != nil {
			socket.conn.Close()
			socket.conn = nil
		}

		// 触发 close 事件
		socket.mutex.RLock()
		closeHandler := socket.handlers["close"]
		socket.mutex.RUnlock()

		if closeHandler != nil {
			if fn, ok := goja.AssertFunction(closeHandler); ok {
				fn(goja.Undefined())
			}
		}

		return goja.Undefined()
	})

	// 获取地址信息
	obj.Set("address", func(call goja.FunctionCall) goja.Value {
		if socket.conn == nil {
			return goja.Undefined()
		}

		addr := socket.conn.LocalAddr().(*net.UDPAddr)
		addrObj := n.vm.NewObject()
		addrObj.Set("address", addr.IP.String())
		addrObj.Set("port", addr.Port)
		addrObj.Set("family", "IPv4")

		return addrObj
	})

	return obj
}

// startReceiving 开始接收 UDP 数据
func (s *UDPSocket) startReceiving() {
	go func() {
		buffer := make([]byte, 65535)
		for {
			n, addr, err := s.conn.ReadFromUDP(buffer)
			if err != nil {
				// 触发 error 事件
				s.mutex.RLock()
				errorHandler := s.handlers["error"]
				s.mutex.RUnlock()

				if errorHandler != nil {
					if fn, ok := goja.AssertFunction(errorHandler); ok {
						errObj := s.vm.NewObject()
						errObj.Set("message", err.Error())
						fn(goja.Undefined(), errObj)
					}
				}
				break
			}

			// 触发 message 事件
			s.mutex.RLock()
			messageHandler := s.handlers["message"]
			s.mutex.RUnlock()

			if messageHandler != nil {
				if fn, ok := goja.AssertFunction(messageHandler); ok {
					msgObj := s.vm.NewObject()
					msgObj.Set("data", string(buffer[:n]))
					msgObj.Set("address", addr.IP.String())
					msgObj.Set("port", addr.Port)

					fn(goja.Undefined(), s.vm.ToValue(string(buffer[:n])), msgObj)
				}
			}
		}
	}()
}

// getNextConnID 获取下一个连接 ID
func (n *NetModule) getNextConnID() int {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.connID++
	return n.connID
}
