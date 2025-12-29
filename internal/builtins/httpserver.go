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
	"time"

	"github.com/dop251/goja"
	"github.com/gorilla/websocket"

	"sw_runtime/internal/consts"
)

// 全局变量，标记是否有 HTTP 服务器在运行
var httpServerRunning int32

// 服务器注册表，用于跟踪所有活动服务器
var serverRegistry = struct {
	sync.RWMutex
	servers map[*HTTPServer]struct{}
}{
	servers: make(map[*HTTPServer]struct{}),
}

// IsHTTPServerRunning 检查是否有 HTTP 服务器在运行
func IsHTTPServerRunning() bool {
	serverRegistry.RLock()
	defer serverRegistry.RUnlock()
	return len(serverRegistry.servers) > 0
}

// registerServer 注册服务器
func registerServer(s *HTTPServer) {
	serverRegistry.Lock()
	defer serverRegistry.Unlock()
	serverRegistry.servers[s] = struct{}{}
}

// unregisterServer 注销服务器
func unregisterServer(s *HTTPServer) {
	serverRegistry.Lock()
	defer serverRegistry.Unlock()
	delete(serverRegistry.servers, s)
}

// closeAllHTTPServers 关闭所有注册的 HTTP 服务器
func closeAllHTTPServers() {
	serverRegistry.Lock()
	servers := make([]*HTTPServer, 0, len(serverRegistry.servers))
	for s := range serverRegistry.servers {
		servers = append(servers, s)
	}
	serverRegistry.Unlock()

	// 关闭所有服务器
	for _, s := range servers {
		if s.server != nil {
			ctx, cancel := context.WithTimeout(context.Background(), consts.DefaultHTTPTimeout)
			s.server.Shutdown(ctx)
			cancel()
		}
		s.stopVMProcessor()
	}
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

	// WebSocket 安全配置
	wsAllowedOrigins []string
	wsAllowAll       bool // 默认 false，生产环境应该设为 false

	// 超时配置
	readTimeout       time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	readHeaderTimeout time.Duration
	maxHeaderBytes    int

	// 请求处理队列（用于保护 goja.Runtime 并发访问，通过事件队列处理）
	requestChan chan func(*goja.Runtime)
	wg          sync.WaitGroup
	stopChan    chan struct{}
	stopOnce    sync.Once
	initialized bool
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
		mux:              http.NewServeMux(),
		vm:               h.vm,
		routes:           make(map[string]map[string]goja.Value),
		middleware:       make([]goja.Value, 0),
		wsAllowedOrigins: []string{},
		wsAllowAll:       false, // 默认不允许所有来源
		requestChan:      make(chan func(*goja.Runtime), 100),
		stopChan:         make(chan struct{}),
		// 默认超时配置
		readTimeout:       consts.DefaultReadTimeout,
		writeTimeout:      consts.DefaultWriteTimeout,
		idleTimeout:       consts.DefaultIdleTimeout,
		readHeaderTimeout: 10 * time.Second,
		maxHeaderBytes:    consts.MaxHeaderSize,
	}

	// 解析配置参数
	if len(call.Arguments) > 0 && call.Arguments[0] != goja.Undefined() && call.Arguments[0] != goja.Null() {
		configObj := call.Arguments[0].ToObject(h.vm)
		if configObj != nil {
			// 读取超时配置（秒）
			if readTimeout := configObj.Get("readTimeout"); readTimeout != nil && readTimeout != goja.Undefined() {
				if seconds, ok := readTimeout.Export().(int64); ok {
					server.readTimeout = time.Duration(seconds) * time.Second
				}
			}
			if writeTimeout := configObj.Get("writeTimeout"); writeTimeout != nil && writeTimeout != goja.Undefined() {
				if seconds, ok := writeTimeout.Export().(int64); ok {
					server.writeTimeout = time.Duration(seconds) * time.Second
				}
			}
			if idleTimeout := configObj.Get("idleTimeout"); idleTimeout != nil && idleTimeout != goja.Undefined() {
				if seconds, ok := idleTimeout.Export().(int64); ok {
					server.idleTimeout = time.Duration(seconds) * time.Second
				}
			}
			if readHeaderTimeout := configObj.Get("readHeaderTimeout"); readHeaderTimeout != nil && readHeaderTimeout != goja.Undefined() {
				if seconds, ok := readHeaderTimeout.Export().(int64); ok {
					server.readHeaderTimeout = time.Duration(seconds) * time.Second
				}
			}
			if maxHeaderBytes := configObj.Get("maxHeaderBytes"); maxHeaderBytes != nil && maxHeaderBytes != goja.Undefined() {
				if bytes, ok := maxHeaderBytes.Export().(int64); ok {
					server.maxHeaderBytes = int(bytes)
				}
			}
		}
	}

	// 初始化路由映射
	server.routes["GET"] = make(map[string]goja.Value)
	server.routes["POST"] = make(map[string]goja.Value)
	server.routes["PUT"] = make(map[string]goja.Value)
	server.routes["DELETE"] = make(map[string]goja.Value)
	server.routes["PATCH"] = make(map[string]goja.Value)
	server.routes["HEAD"] = make(map[string]goja.Value)
	server.routes["OPTIONS"] = make(map[string]goja.Value)

	// 初始化 WebSocket（安全的 CORS 配置）
	server.ws = make(map[string]goja.Value)
	server.upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return server.checkWebSocketOrigin(r)
		},
		ReadBufferSize:  consts.WSReadBufferSize,
		WriteBufferSize: consts.WSWriteBufferSize,
	}

	// 启动 VM 处理器
	server.startVMProcessor()

	return h.createServerObject(server)
}

// checkWebSocketOrigin 检查 WebSocket 请求的来源是否允许
func (s *HTTPServer) checkWebSocketOrigin(r *http.Request) bool {
	// 如果明确允许所有来源（仅用于开发环境）
	if s.wsAllowAll {
		return true
	}

	// 获取请求的 Origin
	origin := r.Header.Get("Origin")

	// 如果没有 Origin 头，拒绝连接
	if origin == "" {
		return false
	}

	// 检查是否在允许列表中
	for _, allowed := range s.wsAllowedOrigins {
		if allowed == "*" {
			return true // 允许所有来源
		}
		if origin == allowed {
			return true
		}
		// 支持通配符前缀匹配，如 https://*.example.com
		if strings.HasSuffix(allowed, "*") {
			prefix := strings.TrimSuffix(allowed, "*")
			if strings.HasPrefix(origin, prefix) {
				return true
			}
		}
	}

	return false // 默认拒绝
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

	// WebSocket 安全配置
	obj.Set("setWSAllowedOrigins", h.createSetWSAllowedOrigins(server))
	obj.Set("setWSAllowAll", h.createSetWSAllowAll(server))

	// 服务器控制
	obj.Set("listen", h.createListenHandler(server))
	obj.Set("listenTLS", h.createListenTLSHandler(server))
	obj.Set("close", h.createCloseHandler(server))

	return obj
}

// startVMProcessor 启动 VM 处理器（串行化对 goja.Runtime 的访问）
func (s *HTTPServer) startVMProcessor() {
	if s.initialized {
		return
	}
	s.initialized = true
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case fn, ok := <-s.requestChan:
				if !ok {
					return
				}
				if fn != nil {
					fn(s.vm)
				}
			case <-s.stopChan:
				return
			}
		}
	}()
}

// stopVMProcessor 停止 VM 处理器
func (s *HTTPServer) stopVMProcessor() {
	s.stopOnce.Do(func() {
		// 关闭请求通道
		select {
		case <-s.requestChan:
			// 已关闭
		default:
			close(s.requestChan)
		}
		// 发送停止信号
		select {
		case <-s.stopChan:
			// 已经关闭
		default:
			close(s.stopChan)
		}
	})
}

// createSetWSAllowedOrigins 创建设置允许来源的方法
func (h *HTTPServerModule) createSetWSAllowedOrigins(server *HTTPServer) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		server.wsAllowedOrigins = []string{}
		server.wsAllowAll = false

		if len(call.Arguments) > 0 {
			if arg := call.Arguments[0]; arg.ExportType() != nil {
				// 支持字符串数组或单个字符串
				exported := arg.Export()
				switch v := exported.(type) {
				case []string:
					server.wsAllowedOrigins = v
				case []interface{}:
					for _, item := range v {
						if str, ok := item.(string); ok {
							server.wsAllowedOrigins = append(server.wsAllowedOrigins, str)
						}
					}
				case string:
					server.wsAllowedOrigins = []string{v}
				}
			}
		}

		return goja.Undefined()
	}
}

// createSetWSAllowAll 创建设置允许所有来源的方法（仅用于开发）
func (h *HTTPServerModule) createSetWSAllowAll(server *HTTPServer) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		allow := false
		if len(call.Arguments) > 0 {
			allow = call.Arguments[0].ToBoolean()
		}
		server.wsAllowAll = allow
		return goja.Undefined()
	}
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

		// 注册服务器
		registerServer(server)

		go func() {
			server.server = &http.Server{
				Addr:              port,
				Handler:           server.mux,
				ReadTimeout:       server.readTimeout,
				WriteTimeout:      server.writeTimeout,
				IdleTimeout:       server.idleTimeout,
				ReadHeaderTimeout: server.readHeaderTimeout,
				MaxHeaderBytes:    server.maxHeaderBytes,
			}

			h.mutex.Lock()
			h.servers[port] = server
			h.mutex.Unlock()

			// 调用回调函数
			// 注意：回调在服务器启动前调用，此时不需要通过 requestChan
			// 因为没有并发访问 VM 的风险
			if callback != nil {
				if fn, ok := goja.AssertFunction(callback); ok {
					func() {
						defer func() {
							if r := recover(); r != nil {
								fmt.Printf("Callback panic: %v\n", r)
							}
						}()
						_, err := fn(goja.Undefined())
						if err != nil {
							fmt.Printf("Callback error: %v\n", err)
						}
					}()
				}
			}

			// resolve promise
			resolve(h.vm.ToValue(fmt.Sprintf("Server listening on %s", port)))

			// 启动服务器
			if err := server.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				unregisterServer(server)
				// reject promise
				reject(h.vm.NewGoError(err))
			} else {
				// 服务器正常关闭
				unregisterServer(server)
			}
		}()

		return h.vm.ToValue(promise)
	}
}

// createListenTLSHandler 创建 HTTPS 监听处理器
func (h *HTTPServerModule) createListenTLSHandler(server *HTTPServer) func(goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 3 {
			panic(h.vm.NewTypeError("listenTLS requires port, certFile, and keyFile"))
		}

		port := call.Arguments[0].String()
		if !strings.Contains(port, ":") {
			port = ":" + port
		}

		certFile := call.Arguments[1].String()
		keyFile := call.Arguments[2].String()

		var callback goja.Value
		if len(call.Arguments) > 3 {
			if _, ok := goja.AssertFunction(call.Arguments[3]); ok {
				callback = call.Arguments[3]
			}
		}

		promise, resolve, reject := h.vm.NewPromise()

		// 注册服务器
		registerServer(server)

		go func() {
			server.server = &http.Server{
				Addr:              port,
				Handler:           server.mux,
				ReadTimeout:       server.readTimeout,
				WriteTimeout:      server.writeTimeout,
				IdleTimeout:       server.idleTimeout,
				ReadHeaderTimeout: server.readHeaderTimeout,
				MaxHeaderBytes:    server.maxHeaderBytes,
			}

			h.mutex.Lock()
			h.servers[port] = server
			h.mutex.Unlock()

			// 调用回调函数
			if callback != nil {
				if fn, ok := goja.AssertFunction(callback); ok {
					func() {
						defer func() {
							if r := recover(); r != nil {
								fmt.Printf("Callback panic: %v\n", r)
							}
						}()
						_, err := fn(goja.Undefined())
						if err != nil {
							fmt.Printf("Callback error: %v\n", err)
						}
					}()
				}
			}

			// resolve promise
			resolve(h.vm.ToValue(fmt.Sprintf("HTTPS Server listening on %s", port)))

			// 启动 HTTPS 服务器
			if err := server.server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				unregisterServer(server)
				// reject promise
				reject(h.vm.NewGoError(err))
			} else {
				// 服务器正常关闭
				unregisterServer(server)
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
			// 1. 停止 VM 处理器
			server.stopVMProcessor()

			// 2. 关闭 HTTP 服务器
			if server.server != nil {
				ctx, cancel := context.WithTimeout(context.Background(), consts.DefaultHTTPTimeout)
				defer cancel()

				if err := server.server.Shutdown(ctx); err != nil {
					reject(h.vm.NewGoError(err))
					return
				}
			}

			// 3. 等待所有 goroutine 完成
			done := make(chan struct{})
			go func() {
				server.wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				resolve(h.vm.ToValue("Server closed"))
			case <-time.After(10 * time.Second):
				reject(h.vm.NewGoError(fmt.Errorf("timeout waiting for goroutines to finish")))
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

		// 使用 channel 等待处理完成
		done := make(chan struct{})

		// 提交到 VM 处理队列异步执行
		select {
		case server.requestChan <- func(vm *goja.Runtime) {
			defer close(done)
			defer func() {
				if r := recover(); r != nil {
					if !rw.written {
						http.Error(w, fmt.Sprintf("Handler panic: %v", r), http.StatusInternalServerError)
					}
				}
			}()

			// 在 VM goroutine 中创建请求和响应对象
			reqObj := h.createRequestObject(r)
			resObj := h.createResponseObjectWithWrapper(rw, r)

			// 执行中间件链
			middlewareIndex := 0
			var executeNext func()
			executeNext = func() {
				if middlewareIndex >= len(middleware) {
					// 所有中间件执行完毕，执行路由处理器
					if fn, ok := goja.AssertFunction(handler); ok {
						_, err := fn(goja.Undefined(), reqObj, resObj)
						if err != nil && !rw.written {
							http.Error(w, "Handler error: "+err.Error(), http.StatusInternalServerError)
						}
					}
					return
				}

				mw := middleware[middlewareIndex]
				middlewareIndex++

				if fn, ok := goja.AssertFunction(mw); ok {
					nextFunc := vm.ToValue(executeNext)
					_, err := fn(goja.Undefined(), reqObj, resObj, nextFunc)
					if err != nil && !rw.written {
						http.Error(w, "Middleware error", http.StatusInternalServerError)
					}
				}
			}

			executeNext()

			// 确保响应被写入
			if !rw.written && len(rw.body) > 0 {
				w.WriteHeader(rw.statusCode)
				w.Write(rw.body)
			}
		}:
			// 等待处理完成
			<-done
		case <-time.After(30 * time.Second):
			// 超时处理
			http.Error(w, "Request processing timeout", http.StatusRequestTimeout)
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
			jsonData, err := json.Marshal(data)
			if err == nil {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(rw.statusCode)
				w.Write(jsonData)
				rw.written = true
			}
		}
		return obj
	})

	// 发送 HTML 响应
	obj.Set("html", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			data := call.Arguments[0].String()
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(rw.statusCode)
			w.Write([]byte(data))
			rw.written = true
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
	// 路径验证 - 防止路径遍历攻击
	cleanPath := filepath.Clean(filePath)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid file path"))
		rw.written = true
		return
	}

	// 检查是否包含路径遍历模式
	if strings.Contains(cleanPath, "..") {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Access denied: path traversal not allowed"))
		rw.written = true
		return
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(absPath)
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
	content, err := os.ReadFile(absPath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error reading file"))
		rw.written = true
		return
	}

	// 检测 MIME 类型
	contentType := h.detectContentType(absPath, content)
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

		// 使用 channel 等待处理完成
		done := make(chan struct{})

		// 提交到 VM 处理队列异步执行
		select {
		case server.requestChan <- func(vm *goja.Runtime) {
			defer close(done)
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("WebSocket handler panic: %v\n", r)
				}
			}()

			// 在 VM goroutine 中创建 WebSocket 连接对象
			wsObj := h.createWebSocketObject(server, conn)

			// 调用处理器
			if fn, ok := goja.AssertFunction(handler); ok {
				_, err := fn(goja.Undefined(), wsObj)
				if err != nil {
					fmt.Printf("WebSocket handler error: %v\n", err)
				}
			}
		}:
			// 等待处理完成
			<-done
		case <-time.After(30 * time.Second):
			// 超时处理
			fmt.Printf("WebSocket handler timeout\n")
			conn.Close()
		}
	}
}

// createWebSocketObject 创建 WebSocket 连接对象
func (h *HTTPServerModule) createWebSocketObject(server *HTTPServer, conn *websocket.Conn) goja.Value {
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

				// 调用事件处理器（提交到 VM 处理队列）
				for _, handler := range handlers {
					if fn, ok := goja.AssertFunction(handler); ok {
						// 提交到 VM 处理队列异步执行
						select {
						case server.requestChan <- func(vm *goja.Runtime) {
							defer func() {
								if r := recover(); r != nil {
									fmt.Printf("WebSocket message handler panic: %v\n", r)
								}
							}()
							fn(goja.Undefined(), h.vm.ToValue(data))
						}:
							// 提交成功，继续
						default:
							// 队列已满，跳过
							fmt.Printf("WebSocket message queue full, dropping message\n")
						}
					}
				}
			}
		}
	}()

	return obj
}
