package builtins

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dop251/goja"
	"github.com/gorilla/websocket"
)

// 全局变量，标记是否有 HTTP 服务器在运行
var httpServerRunning int32

// IsHTTPServerRunning 检查是否有 HTTP 服务器在运行
func IsHTTPServerRunning() bool {
	return atomic.LoadInt32(&httpServerRunning) > 0
}

// HTTPServerModule HTTP 服务器模块
type HTTPServerModule struct {
	vm      *goja.Runtime
	servers map[string]*HTTPServer
	mutex   sync.RWMutex
}

// HTTPServer HTTP 服务器实例
type HTTPServer struct {
	server     *http.Server
	mux        *http.ServeMux
	vm         *goja.Runtime
	routes     map[string]map[string]goja.Value // method -> path -> handler
	middleware []goja.Value
	ws         map[string]goja.Value // WebSocket 路由
	upgrader   websocket.Upgrader    // WebSocket 升级器
	mutex      sync.RWMutex
}

// NewHTTPServerModule 创建 HTTP 服务器模块
func NewHTTPServerModule(vm *goja.Runtime) *HTTPServerModule {
	return &HTTPServerModule{
		vm:      vm,
		servers: make(map[string]*HTTPServer),
	}
}

// GetModule 获取 HTTP 服务器模块对象
func (h *HTTPServerModule) GetModule() *goja.Object {
	obj := h.vm.NewObject()

	// 创建服务器
	obj.Set("createServer", h.createServer)
	obj.Set("Server", h.createServer) // 别名

	// 状态码常量
	statusCodes := h.vm.NewObject()
	statusCodes.Set("OK", 200)
	statusCodes.Set("CREATED", 201)
	statusCodes.Set("NO_CONTENT", 204)
	statusCodes.Set("BAD_REQUEST", 400)
	statusCodes.Set("UNAUTHORIZED", 401)
	statusCodes.Set("FORBIDDEN", 403)
	statusCodes.Set("NOT_FOUND", 404)
	statusCodes.Set("METHOD_NOT_ALLOWED", 405)
	statusCodes.Set("INTERNAL_SERVER_ERROR", 500)
	statusCodes.Set("BAD_GATEWAY", 502)
	statusCodes.Set("SERVICE_UNAVAILABLE", 503)
	obj.Set("STATUS_CODES", statusCodes)

	return obj
}

// createServer 创建 HTTP 服务器
func (h *HTTPServerModule) createServer(call goja.FunctionCall) goja.Value {
	server := &HTTPServer{
		mux:        http.NewServeMux(),
		vm:         h.vm,
		routes:     make(map[string]map[string]goja.Value),
		middleware: make([]goja.Value, 0),
	}

	// 初始化路由映射
	server.routes["GET"] = make(map[string]goja.Value)
	server.routes["POST"] = make(map[string]goja.Value)
	server.routes["PUT"] = make(map[string]goja.Value)
	server.routes["DELETE"] = make(map[string]goja.Value)
	server.routes["PATCH"] = make(map[string]goja.Value)
	server.routes["HEAD"] = make(map[string]goja.Value)
	server.routes["OPTIONS"] = make(map[string]goja.Value)

	// 初始化 WebSocket
	server.ws = make(map[string]goja.Value)
	server.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源,生产环境应该限制
		},
	}

	return h.createServerObject(server)
}

// createServerObject 创建服务器对象
func (h *HTTPServerModule) createServerObject(server *HTTPServer) goja.Value {
	obj := h.vm.NewObject()

	// 路由方法
	obj.Set("get", h.createRouteHandler(server, "GET"))
	obj.Set("post", h.createRouteHandler(server, "POST"))
	obj.Set("put", h.createRouteHandler(server, "PUT"))
	obj.Set("delete", h.createRouteHandler(server, "DELETE"))
	obj.Set("patch", h.createRouteHandler(server, "PATCH"))
	obj.Set("head", h.createRouteHandler(server, "HEAD"))
	obj.Set("options", h.createRouteHandler(server, "OPTIONS"))

	// 通用路由方法
	obj.Set("route", h.createGenericRouteHandler(server))
	obj.Set("use", h.createMiddlewareHandler(server))

	// 静态文件服务
	obj.Set("static", h.createStaticHandler(server))

	// WebSocket 路由
	obj.Set("ws", h.createWebSocketHandler(server))

	// 服务器控制
	obj.Set("listen", h.createListenHandler(server))
	obj.Set("close", h.createCloseHandler(server))

	return obj
}

// createRouteHandler 创建路由处理器
func (h *HTTPServerModule) createRouteHandler(server *HTTPServer, method string) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(h.vm.NewTypeError(fmt.Sprintf("%s requires path and handler", strings.ToLower(method))))
		}

		path := call.Arguments[0].String()
		handler := call.Arguments[1]

		if _, ok := goja.AssertFunction(handler); !ok {
			panic(h.vm.NewTypeError("Handler must be a function"))
		}

		server.mutex.Lock()
		server.routes[method][path] = handler
		server.mutex.Unlock()

		// 注册到 mux
		server.mux.HandleFunc(path, h.createHTTPHandler(server, method, path))

		return goja.Undefined()
	}
}

// createGenericRouteHandler 创建通用路由处理器
func (h *HTTPServerModule) createGenericRouteHandler(server *HTTPServer) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			panic(h.vm.NewTypeError("route requires method, path and handler"))
		}

		method := strings.ToUpper(call.Arguments[0].String())
		path := call.Arguments[1].String()
		handler := call.Arguments[2]

		if _, ok := goja.AssertFunction(handler); !ok {
			panic(h.vm.NewTypeError("Handler must be a function"))
		}

		server.mutex.Lock()
		if server.routes[method] == nil {
			server.routes[method] = make(map[string]goja.Value)
		}
		server.routes[method][path] = handler
		server.mutex.Unlock()

		// 注册到 mux
		server.mux.HandleFunc(path, h.createHTTPHandler(server, method, path))

		return goja.Undefined()
	}
}

// createMiddlewareHandler 创建中间件处理器
func (h *HTTPServerModule) createMiddlewareHandler(server *HTTPServer) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(h.vm.NewTypeError("use requires middleware function"))
		}

		middleware := call.Arguments[0]
		if _, ok := goja.AssertFunction(middleware); !ok {
			panic(h.vm.NewTypeError("Middleware must be a function"))
		}

		server.mutex.Lock()
		server.middleware = append(server.middleware, middleware)
		server.mutex.Unlock()

		return goja.Undefined()
	}
}

// createStaticHandler 创建静态文件处理器
func (h *HTTPServerModule) createStaticHandler(server *HTTPServer) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(h.vm.NewTypeError("static requires directory path"))
		}

		dir := call.Arguments[0].String()
		prefix := "/"
		if len(call.Arguments) > 1 {
			prefix = call.Arguments[1].String()
		}

		fileServer := http.FileServer(http.Dir(dir))
		server.mux.Handle(prefix, http.StripPrefix(prefix, fileServer))

		return goja.Undefined()
	}
}

// createListenHandler 创建监听处理器
func (h *HTTPServerModule) createListenHandler(server *HTTPServer) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(h.vm.NewTypeError("listen requires port"))
		}

		port := call.Arguments[0].String()
		if !strings.Contains(port, ":") {
			port = ":" + port
		}

		var callback goja.Value
		if len(call.Arguments) > 1 {
			if _, ok := goja.AssertFunction(call.Arguments[1]); ok {
				callback = call.Arguments[1]
			}
		}

		promise, resolve, reject := h.vm.NewPromise()

		// 标记有 HTTP 服务器在运行
		atomic.AddInt32(&httpServerRunning, 1)

		go func() {
			server.server = &http.Server{
				Addr:    port,
				Handler: server.mux,
			}

			h.mutex.Lock()
			h.servers[port] = server
			h.mutex.Unlock()

			// 调用回调函数
			if callback != nil {
				if fn, ok := goja.AssertFunction(callback); ok {
					_, err := fn(goja.Undefined())
					if err != nil {
						// 忽略回调函数中的错误，不影响服务器启动
						fmt.Printf("Callback error: %v\n", err)
					}
				}
			}

			resolve(h.vm.ToValue(fmt.Sprintf("Server listening on %s", port)))

			// 启动服务器
			if err := server.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				atomic.AddInt32(&httpServerRunning, -1)
				reject(h.vm.NewGoError(err))
			}
		}()

		return h.vm.ToValue(promise)
	}
}

// createCloseHandler 创建关闭处理器
func (h *HTTPServerModule) createCloseHandler(server *HTTPServer) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		promise, resolve, reject := h.vm.NewPromise()

		go func() {
			if server.server != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := server.server.Shutdown(ctx); err != nil {
					reject(h.vm.NewGoError(err))
				} else {
					resolve(h.vm.ToValue("Server closed"))
				}
			} else {
				resolve(h.vm.ToValue("Server not running"))
			}
		}()

		return h.vm.ToValue(promise)
	}
}

// createHTTPHandler 创建 HTTP 处理器
func (h *HTTPServerModule) createHTTPHandler(server *HTTPServer, method, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 检查方法匹配
		if r.Method != method {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		server.mutex.RLock()
		handler, exists := server.routes[method][path]
		middleware := server.middleware
		server.mutex.RUnlock()

		if !exists {
			http.NotFound(w, r)
			return
		}

		// 使用 ResponseWriter 包装器来捕获响应
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     200,
		}

		// 创建请求对象
		reqObj := h.createRequestObject(r)
		resObj := h.createResponseObjectWithWrapper(rw, r)

		// 执行中间件
		middlewareIndex := 0
		var executeNext func()
		executeNext = func() {
			if middlewareIndex >= len(middleware) {
				// 所有中间件执行完毕，执行路由处理器
				fmt.Printf("DEBUG: Executing route handler\n")
				if fn, ok := goja.AssertFunction(handler); ok {
					_, err := fn(goja.Undefined(), reqObj, resObj)
					if err != nil {
						fmt.Printf("DEBUG: Handler error: %v\n", err)
						if !rw.written {
							http.Error(w, "Handler error: "+err.Error(), http.StatusInternalServerError)
						}
						return
					}
					fmt.Printf("DEBUG: Handler executed successfully\n")
				} else {
					fmt.Printf("DEBUG: Handler is not a function\n")
				}
				return
			}

			mw := middleware[middlewareIndex]
			middlewareIndex++

			fmt.Printf("DEBUG: Executing middleware %d\n", middlewareIndex-1)
			if fn, ok := goja.AssertFunction(mw); ok {
				nextFunc := h.vm.ToValue(executeNext)
				_, err := fn(goja.Undefined(), reqObj, resObj, nextFunc)
				if err != nil {
					fmt.Printf("DEBUG: Middleware error: %v\n", err)
					if !rw.written {
						http.Error(w, "Middleware error", http.StatusInternalServerError)
					}
					return
				}
			}
		}

		executeNext()

		// 确保响应被写入
		if !rw.written && len(rw.body) > 0 {
			w.WriteHeader(rw.statusCode)
			w.Write(rw.body)
		}
	}
}

// responseWriter 包装器
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
	written    bool
}

// createRequestObject 创建请求对象
func (h *HTTPServerModule) createRequestObject(r *http.Request) goja.Value {
	obj := h.vm.NewObject()

	// 基本信息
	obj.Set("method", r.Method)
	obj.Set("url", r.URL.String())
	obj.Set("path", r.URL.Path)
	obj.Set("query", r.URL.RawQuery)

	// 请求头
	headers := h.vm.NewObject()
	for key, values := range r.Header {
		if len(values) == 1 {
			headers.Set(key, values[0])
		} else {
			headers.Set(key, values)
		}
	}
	obj.Set("headers", headers)

	// 查询参数
	params := h.vm.NewObject()
	for key, values := range r.URL.Query() {
		if len(values) == 1 {
			params.Set(key, values[0])
		} else {
			params.Set(key, values)
		}
	}
	obj.Set("params", params)

	// 读取请求体
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err == nil {
			obj.Set("body", string(body))

			// 尝试解析 JSON
			if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
				var jsonData interface{}
				if json.Unmarshal(body, &jsonData) == nil {
					obj.Set("json", h.vm.ToValue(jsonData))
				}
			}
		}
	}

	// 客户端信息
	obj.Set("ip", r.RemoteAddr)
	obj.Set("userAgent", r.UserAgent())

	return obj
}

// createResponseObject 创建响应对象
func (h *HTTPServerModule) createResponseObject(w http.ResponseWriter, r *http.Request) goja.Value {
	return h.createResponseObjectWithWrapper(&responseWriter{ResponseWriter: w, statusCode: 200}, r)
}

// createResponseObjectWithWrapper 使用包装器创建响应对象
func (h *HTTPServerModule) createResponseObjectWithWrapper(rw *responseWriter, r *http.Request) goja.Value {
	obj := h.vm.NewObject()
	w := rw.ResponseWriter

	// 设置状态码
	obj.Set("status", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			if code, ok := call.Arguments[0].Export().(int64); ok {
				rw.statusCode = int(code)
			}
		}
		return obj
	})

	// 设置响应头
	obj.Set("header", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) >= 2 {
			key := call.Arguments[0].String()
			value := call.Arguments[1].String()
			w.Header().Set(key, value)
		}
		return obj
	})

	// 发送文本响应
	obj.Set("send", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			data := call.Arguments[0].String()
			if w.Header().Get("Content-Type") == "" {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			}
			w.WriteHeader(rw.statusCode)
			w.Write([]byte(data))
			rw.written = true
		}
		return obj
	})

	// 发送 JSON 响应
	obj.Set("json", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			data := call.Arguments[0].Export()
			fmt.Printf("DEBUG: json() called with data: %v\n", data)
			jsonData, err := json.Marshal(data)
			if err == nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(rw.statusCode)
				n, writeErr := w.Write(jsonData)
				rw.written = true
				fmt.Printf("DEBUG: json() wrote %d bytes, error: %v\n", n, writeErr)
			} else {
				fmt.Printf("DEBUG: json() marshal error: %v\n", err)
			}
		} else {
			fmt.Printf("DEBUG: json() called with no arguments\n")
		}
		return obj
	})

	// 发送 HTML 响应
	obj.Set("html", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			data := call.Arguments[0].String()
			fmt.Printf("DEBUG: html() called with %d bytes\n", len(data))
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(rw.statusCode)
			n, err := w.Write([]byte(data))
			rw.written = true
			fmt.Printf("DEBUG: html() wrote %d bytes, error: %v\n", n, err)
		}
		return obj
	})

	// 发送文件
	obj.Set("sendFile", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			filePath := call.Arguments[0].String()
			h.sendFileResponse(w, rw, filePath)
		}
		return obj
	})

	// 下载文件
	obj.Set("download", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			filePath := call.Arguments[0].String()
			filename := filepath.Base(filePath)

			// 可选的自定义文件名
			if len(call.Arguments) > 1 {
				filename = call.Arguments[1].String()
			}

			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
			h.sendFileResponse(w, rw, filePath)
		}
		return obj
	})

	// 重定向
	obj.Set("redirect", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			url := call.Arguments[0].String()
			code := 302
			if len(call.Arguments) > 1 {
				if c, ok := call.Arguments[1].Export().(int64); ok {
					code = int(c)
				}
			}
			http.Redirect(w, r, url, code)
			rw.written = true
		}
		return obj
	})

	return obj
}

// sendFileResponse 发送文件响应
func (h *HTTPServerModule) sendFileResponse(w http.ResponseWriter, rw *responseWriter, filePath string) {
	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("File not found"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Error accessing file"))
		}
		rw.written = true
		return
	}

	// 检查是否为目录
	if fileInfo.IsDir() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Cannot send directory"))
		rw.written = true
		return
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error reading file"))
		rw.written = true
		return
	}

	// 检测 MIME 类型
	contentType := h.detectContentType(filePath, content)
	w.Header().Set("Content-Type", contentType)

	// 设置缓存控制头
	w.Header().Set("Last-Modified", fileInfo.ModTime().UTC().Format(http.TimeFormat))
	w.Header().Set("Cache-Control", "public, max-age=3600")

	// 发送文件内容
	w.WriteHeader(rw.statusCode)
	w.Write(content)
	rw.written = true
}

// detectContentType 检测文件的 MIME 类型
func (h *HTTPServerModule) detectContentType(filePath string, content []byte) string {
	// 首先根据文件扩展名判断
	ext := filepath.Ext(filePath)
	if ext != "" {
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			return mimeType
		}
	}

	// 如果无法从扩展名判断,使用内容检测
	if len(content) > 0 {
		return http.DetectContentType(content)
	}

	return "application/octet-stream"
}

// createWebSocketHandler 创建 WebSocket 路由处理器
func (h *HTTPServerModule) createWebSocketHandler(server *HTTPServer) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(h.vm.NewTypeError("ws requires path and handler"))
		}

		path := call.Arguments[0].String()
		handler := call.Arguments[1]

		if _, ok := goja.AssertFunction(handler); !ok {
			panic(h.vm.NewTypeError("Handler must be a function"))
		}

		server.mutex.Lock()
		server.ws[path] = handler
		server.mutex.Unlock()

		// 注册 WebSocket 路由
		server.mux.HandleFunc(path, h.createWebSocketHTTPHandler(server, path))

		return goja.Undefined()
	}
}

// createWebSocketHTTPHandler 创建 WebSocket HTTP 处理器
func (h *HTTPServerModule) createWebSocketHTTPHandler(server *HTTPServer, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 升级到 WebSocket
		conn, err := server.upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("WebSocket upgrade error: %v\n", err)
			return
		}

		server.mutex.RLock()
		handler, exists := server.ws[path]
		server.mutex.RUnlock()

		if !exists {
			conn.Close()
			return
		}

		// 创建 WebSocket 连接对象
		wsObj := h.createWebSocketObject(conn)

		// 调用处理器
		if fn, ok := goja.AssertFunction(handler); ok {
			_, err := fn(goja.Undefined(), wsObj)
			if err != nil {
				fmt.Printf("WebSocket handler error: %v\n", err)
			}
		}
	}
}

// createWebSocketObject 创建 WebSocket 连接对象
func (h *HTTPServerModule) createWebSocketObject(conn *websocket.Conn) goja.Value {
	obj := h.vm.NewObject()

	// 事件监听器
	listeners := make(map[string][]goja.Value)
	var listenersMutex sync.RWMutex

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

	// 发送消息
	obj.Set("send", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}

		data := call.Arguments[0]
		var message []byte
		var err error

		// 判断数据类型
		if data.ExportType().Kind() == reflect.String {
			// 文本消息
			message = []byte(data.String())
			err = conn.WriteMessage(websocket.TextMessage, message)
		} else {
			// JSON 消息
			exported := data.Export()
			message, err = json.Marshal(exported)
			if err == nil {
				err = conn.WriteMessage(websocket.TextMessage, message)
			}
		}

		if err != nil {
			fmt.Printf("WebSocket send error: %v\n", err)
			return h.vm.ToValue(false)
		}

		return h.vm.ToValue(true)
	})

	// 发送 JSON
	obj.Set("sendJSON", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}

		data := call.Arguments[0].Export()
		message, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("WebSocket JSON marshal error: %v\n", err)
			return h.vm.ToValue(false)
		}

		err = conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			fmt.Printf("WebSocket send error: %v\n", err)
			return h.vm.ToValue(false)
		}

		return h.vm.ToValue(true)
	})

	// 关闭连接
	obj.Set("close", func(call goja.FunctionCall) goja.Value {
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

		return goja.Undefined()
	})

	// 启动消息接收循环
	go func() {
		defer conn.Close()

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					fmt.Printf("WebSocket read error: %v\n", err)
				}
				break
			}

			// 触发 message 事件
			listenersMutex.RLock()
			handlers := listeners["message"]
			listenersMutex.RUnlock()

			if len(handlers) > 0 {
				var data interface{}
				if messageType == websocket.TextMessage {
					// 尝试解析为 JSON
					if json.Unmarshal(message, &data) != nil {
						data = string(message)
					}
				} else {
					data = message
				}

				// 注意: 这里直接调用 goja 函数可能不安全
				// 在生产环境中应该使用事件队列
				for _, handler := range handlers {
					if fn, ok := goja.AssertFunction(handler); ok {
						// 使用 defer/recover 保护
						func() {
							defer func() {
								if r := recover(); r != nil {
									fmt.Printf("WebSocket handler panic: %v\n", r)
								}
							}()
							fn(goja.Undefined(), h.vm.ToValue(data))
						}()
					}
				}
			}
		}
	}()

	return obj
}
