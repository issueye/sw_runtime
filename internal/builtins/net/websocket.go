package net

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dop251/goja"
	"github.com/gorilla/websocket"
)

// WebSocketModule WebSocket 客户端模块
type WebSocketModule struct {
	vm    *goja.Runtime
	conns map[int]*websocket.Conn
	mutex sync.RWMutex
	id    int
}

// NewWebSocketModule 创建 WebSocket 客户端模块
func NewWebSocketModule(vm *goja.Runtime) *WebSocketModule {
	return &WebSocketModule{
		vm:    vm,
		conns: make(map[int]*websocket.Conn),
	}
}

// GetModule 获取模块对象
func (w *WebSocketModule) GetModule() *goja.Object {
	obj := w.vm.NewObject()

	// connect 方法 - 连接到 WebSocket 服务器
	obj.Set("connect", w.connect)

	// createClient 别名
	obj.Set("createClient", w.connect)

	return obj
}

// connect 连接到 WebSocket 服务器
func (w *WebSocketModule) connect(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(w.vm.NewTypeError("connect requires URL"))
	}

	url := call.Arguments[0].String()

	// 解析选项
	var options *websocket.Dialer
	headers := http.Header{}

	if len(call.Arguments) > 1 && !goja.IsUndefined(call.Arguments[1]) && !goja.IsNull(call.Arguments[1]) {
		optObj := call.Arguments[1].ToObject(w.vm)
		if optObj != nil {
			// 获取超时设置
			timeoutVal := optObj.Get("timeout")
			if timeoutVal != nil && !goja.IsUndefined(timeoutVal) && !goja.IsNull(timeoutVal) {
				timeout := timeoutVal.ToInteger()
				options = &websocket.Dialer{
					HandshakeTimeout: time.Duration(timeout) * time.Millisecond,
				}
			}

			// 获取请求头
			headersVal := optObj.Get("headers")
			if headersVal != nil && !goja.IsUndefined(headersVal) && !goja.IsNull(headersVal) {
				if headersObj := headersVal.Export(); headersObj != nil {
					if h, ok := headersObj.(map[string]interface{}); ok {
						for key, val := range h {
							headers.Set(key, fmt.Sprint(val))
						}
					}
				}
			}

			// 获取协议
			protocolsVal := optObj.Get("protocols")
			if protocolsVal != nil && !goja.IsUndefined(protocolsVal) && !goja.IsNull(protocolsVal) {
				if protocols := protocolsVal.Export(); protocols != nil {
					if protoList, ok := protocols.([]interface{}); ok {
						subprotocols := make([]string, len(protoList))
						for i, p := range protoList {
							subprotocols[i] = fmt.Sprint(p)
						}
						if options == nil {
							options = &websocket.Dialer{}
						}
						options.Subprotocols = subprotocols
					}
				}
			}
		}
	}

	if options == nil {
		options = &websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		}
	}

	// 创建 Promise
	promise, resolve, reject := w.vm.NewPromise()

	// 异步连接
	go func() {
		conn, _, err := options.Dial(url, headers)
		if err != nil {
			reject(w.vm.NewGoError(fmt.Errorf("WebSocket connection failed: %v", err)))
			return
		}

		// 生成连接 ID
		w.mutex.Lock()
		w.id++
		connID := w.id
		w.conns[connID] = conn
		w.mutex.Unlock()

		// 创建客户端对象
		clientObj := w.createClientObject(conn, connID)
		resolve(clientObj)
	}()

	return w.vm.ToValue(promise)
}

// createClientObject 创建客户端对象
func (w *WebSocketModule) createClientObject(conn *websocket.Conn, connID int) goja.Value {
	obj := w.vm.NewObject()

	// 事件监听器
	listeners := make(map[string][]goja.Value)
	var listenersMutex sync.RWMutex

	// 连接状态
	var closed bool
	var closeMutex sync.RWMutex

	// 设置事件监听器
	obj.Set("on", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}

		eventName := call.Arguments[0].String()
		handler := call.Arguments[1]

		if _, ok := goja.AssertFunction(handler); !ok {
			return goja.Undefined()
		}

		listenersMutex.Lock()
		listeners[eventName] = append(listeners[eventName], handler)
		listenersMutex.Unlock()

		return goja.Undefined()
	})

	// 发送文本消息
	obj.Set("send", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return w.vm.ToValue(false)
		}

		closeMutex.RLock()
		isClosed := closed
		closeMutex.RUnlock()

		if isClosed {
			return w.vm.ToValue(false)
		}

		data := call.Arguments[0].Export()
		var message []byte
		var messageType int

		// 判断消息类型
		switch v := data.(type) {
		case string:
			message = []byte(v)
			messageType = websocket.TextMessage
		case []byte:
			message = v
			messageType = websocket.BinaryMessage
		default:
			// 其他类型转为 JSON
			jsonData, err := json.Marshal(data)
			if err != nil {
				fmt.Printf("WebSocket JSON marshal error: %v\n", err)
				return w.vm.ToValue(false)
			}
			message = jsonData
			messageType = websocket.TextMessage
		}

		err := conn.WriteMessage(messageType, message)
		if err != nil {
			fmt.Printf("WebSocket send error: %v\n", err)
			return w.vm.ToValue(false)
		}

		return w.vm.ToValue(true)
	})

	// 发送 JSON 消息
	obj.Set("sendJSON", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return w.vm.ToValue(false)
		}

		closeMutex.RLock()
		isClosed := closed
		closeMutex.RUnlock()

		if isClosed {
			return w.vm.ToValue(false)
		}

		data := call.Arguments[0].Export()
		message, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("WebSocket JSON marshal error: %v\n", err)
			return w.vm.ToValue(false)
		}

		err = conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			fmt.Printf("WebSocket send error: %v\n", err)
			return w.vm.ToValue(false)
		}

		return w.vm.ToValue(true)
	})

	// 发送二进制消息
	obj.Set("sendBinary", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return w.vm.ToValue(false)
		}

		closeMutex.RLock()
		isClosed := closed
		closeMutex.RUnlock()

		if isClosed {
			return w.vm.ToValue(false)
		}

		data := call.Arguments[0].Export()
		var message []byte

		switch v := data.(type) {
		case []byte:
			message = v
		case string:
			message = []byte(v)
		default:
			return w.vm.ToValue(false)
		}

		err := conn.WriteMessage(websocket.BinaryMessage, message)
		if err != nil {
			fmt.Printf("WebSocket send error: %v\n", err)
			return w.vm.ToValue(false)
		}

		return w.vm.ToValue(true)
	})

	// 发送 ping
	obj.Set("ping", func(call goja.FunctionCall) goja.Value {
		closeMutex.RLock()
		isClosed := closed
		closeMutex.RUnlock()

		if isClosed {
			return w.vm.ToValue(false)
		}

		data := []byte{}
		if len(call.Arguments) > 0 {
			data = []byte(call.Arguments[0].String())
		}

		err := conn.WriteMessage(websocket.PingMessage, data)
		if err != nil {
			fmt.Printf("WebSocket ping error: %v\n", err)
			return w.vm.ToValue(false)
		}

		return w.vm.ToValue(true)
	})

	// 关闭连接
	obj.Set("close", func(call goja.FunctionCall) goja.Value {
		closeMutex.Lock()
		if closed {
			closeMutex.Unlock()
			return goja.Undefined()
		}
		closed = true
		closeMutex.Unlock()

		code := websocket.CloseNormalClosure
		reason := ""

		if len(call.Arguments) > 0 {
			if c, ok := call.Arguments[0].Export().(int64); ok {
				code = int(c)
			}
		}

		if len(call.Arguments) > 1 {
			reason = call.Arguments[1].String()
		}

		message := websocket.FormatCloseMessage(code, reason)
		conn.WriteMessage(websocket.CloseMessage, message)
		conn.Close()

		// 从管理器中移除
		w.mutex.Lock()
		delete(w.conns, connID)
		w.mutex.Unlock()

		return goja.Undefined()
	})

	// 获取连接状态
	obj.Set("isClosed", func(call goja.FunctionCall) goja.Value {
		closeMutex.RLock()
		defer closeMutex.RUnlock()
		return w.vm.ToValue(closed)
	})

	// 启动消息接收循环
	go func() {
		defer func() {
			closeMutex.Lock()
			closed = true
			closeMutex.Unlock()

			conn.Close()

			// 从管理器中移除
			w.mutex.Lock()
			delete(w.conns, connID)
			w.mutex.Unlock()
		}()

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				// 触发 error 或 close 事件
				listenersMutex.RLock()
				handlers := listeners["close"]
				listenersMutex.RUnlock()

				for _, handler := range handlers {
					if fn, ok := goja.AssertFunction(handler); ok {
						func() {
							defer func() {
								if r := recover(); r != nil {
									fmt.Printf("WebSocket close handler panic: %v\n", r)
								}
							}()
							fn(goja.Undefined())
						}()
					}
				}

				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					listenersMutex.RLock()
					errorHandlers := listeners["error"]
					listenersMutex.RUnlock()

					for _, handler := range errorHandlers {
						if fn, ok := goja.AssertFunction(handler); ok {
							func() {
								defer func() {
									if r := recover(); r != nil {
										fmt.Printf("WebSocket error handler panic: %v\n", r)
									}
								}()
								errObj := w.vm.NewObject()
								errObj.Set("message", err.Error())
								fn(goja.Undefined(), w.vm.ToValue(errObj))
							}()
						}
					}
				}
				break
			}

			// 处理不同类型的消息
			var data interface{}
			switch messageType {
			case websocket.TextMessage:
				// 尝试解析为 JSON
				var jsonData interface{}
				if err := json.Unmarshal(message, &jsonData); err == nil {
					data = jsonData
				} else {
					data = string(message)
				}
			case websocket.BinaryMessage:
				data = message
			case websocket.PongMessage:
				// 触发 pong 事件
				listenersMutex.RLock()
				pongHandlers := listeners["pong"]
				listenersMutex.RUnlock()

				for _, handler := range pongHandlers {
					if fn, ok := goja.AssertFunction(handler); ok {
						func() {
							defer func() {
								if r := recover(); r != nil {
									fmt.Printf("WebSocket pong handler panic: %v\n", r)
								}
							}()
							fn(goja.Undefined(), w.vm.ToValue(string(message)))
						}()
					}
				}
				continue
			default:
				continue
			}

			// 触发 message 事件
			listenersMutex.RLock()
			handlers := listeners["message"]
			listenersMutex.RUnlock()

			for _, handler := range handlers {
				if fn, ok := goja.AssertFunction(handler); ok {
					func() {
						defer func() {
							if r := recover(); r != nil {
								fmt.Printf("WebSocket message handler panic: %v\n", r)
							}
						}()
						fn(goja.Undefined(), w.vm.ToValue(data))
					}()
				}
			}
		}
	}()

	return obj
}
