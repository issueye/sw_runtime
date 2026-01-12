package builtins

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
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
	routes     map[string][]routeEntry // method -> route entries
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

	// 路径注册追踪（防止重复注册到 http.ServeMux）
	registeredMuxPaths map[string]bool
}

// routeEntry 路由条目
type routeEntry struct {
	path    string
	pattern *routePattern
	handler goja.Value
}

// routePattern 预编译的路由模式
type routePattern struct {
	parts    []string
	isStatic bool
	params   []string
}

// parseRoutePattern 解析路由模式
func parseRoutePattern(path string) *routePattern {
	if !strings.Contains(path, ":") {
		return &routePattern{isStatic: true}
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	p := &routePattern{parts: parts, isStatic: false}
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			p.params = append(p.params, part[1:])
			p.parts[i] = ":"
		}
	}
	return p
}

// match 匹配路径并提取参数
func (p *routePattern) match(path string) (map[string]string, bool) {
	if p.isStatic {
		return nil, false
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != len(p.parts) {
		return nil, false
	}

	params := make(map[string]string)
	paramIdx := 0
	for i, part := range p.parts {
		if part == ":" {
			params[p.params[paramIdx]] = parts[i]
			paramIdx++
		} else if part != parts[i] {
			return nil, false
		}
	}
	return params, true
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
		mux:                http.NewServeMux(),
		vm:                 h.vm,
		routes:             make(map[string][]routeEntry),
		middleware:         make([]goja.Value, 0),
		wsAllowedOrigins:   []string{},
		wsAllowAll:         false, // 默认不允许所有来源
		requestChan:        make(chan func(*goja.Runtime), 100),
		stopChan:           make(chan struct{}),
		registeredMuxPaths: make(map[string]bool), // 初始化路径追踪
		// 默认超时配置
		readTimeout:       consts.DefaultReadTimeout,
		writeTimeout:      consts.DefaultWriteTimeout,
		idleTimeout:       consts.DefaultIdleTimeout,
		readHeaderTimeout: 10 * time.Second,
		maxHeaderBytes:    consts.MaxHeaderSize,
	}

	// 初始化路由列表
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, m := range methods {
		server.routes[m] = make([]routeEntry, 0)
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
		// 先发送停止信号
		close(s.stopChan)
		// 等待一小段时间让 VMProcessor 处理完当前任务
		time.Sleep(100 * time.Millisecond)
		// 关闭请求通道
		close(s.requestChan)
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
		server.routes[method] = append(server.routes[method], routeEntry{
			path:    path,
			pattern: parseRoutePattern(path),
			handler: handler,
		})

		// 路由分发逻辑优化：注册到 http.ServeMux
		muxPath := path
		if strings.Contains(path, ":") {
			idx := strings.Index(path, ":")
			if idx > 0 {
				muxPath = path[:idx]
				if !strings.HasSuffix(muxPath, "/") {
					lastSlash := strings.LastIndex(muxPath, "/")
					if lastSlash != -1 {
						muxPath = muxPath[:lastSlash+1]
					} else {
						muxPath = "/"
					}
				}
			} else {
				muxPath = "/"
			}
		}

		if !server.registeredMuxPaths[muxPath] {
			server.registeredMuxPaths[muxPath] = true
			server.mutex.Unlock()
			server.mux.HandleFunc(muxPath, h.createHTTPHandler(server, "", muxPath))
		} else {
			server.mutex.Unlock()
		}

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
			server.routes[method] = make([]routeEntry, 0)
		}
		server.routes[method] = append(server.routes[method], routeEntry{
			path:    path,
			pattern: parseRoutePattern(path),
			handler: handler,
		})

		muxPath := path
		if strings.Contains(path, ":") {
			idx := strings.Index(path, ":")
			if idx > 0 {
				muxPath = path[:idx]
				if !strings.HasSuffix(muxPath, "/") {
					lastSlash := strings.LastIndex(muxPath, "/")
					if lastSlash != -1 {
						muxPath = muxPath[:lastSlash+1]
					} else {
						muxPath = "/"
					}
				}
			} else {
				muxPath = "/"
			}
		}

		if !server.registeredMuxPaths[muxPath] {
			server.registeredMuxPaths[muxPath] = true
			server.mutex.Unlock()
			server.mux.HandleFunc(muxPath, h.createHTTPHandler(server, "", muxPath))
		} else {
			server.mutex.Unlock()
		}

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

			// 通过 requestChan 调用回调函数和 resolve（确保线程安全）
			if callback != nil {
				if fn, ok := goja.AssertFunction(callback); ok {
					done := make(chan struct{})
					select {
					case server.requestChan <- func(vm *goja.Runtime) {
						defer close(done)
						defer func() {
							if r := recover(); r != nil {
								fmt.Printf("Callback panic: %v\n", r)
							}
						}()
						_, err := fn(goja.Undefined())
						if err != nil {
							fmt.Printf("Callback error: %v\n", err)
						}
					}:
						<-done
					case <-time.After(5 * time.Second):
						fmt.Printf("Callback timeout\n")
					}
				}
			}

			// resolve promise（通过 requestChan 确保线程安全）
			resolveDone := make(chan struct{})
			select {
			case server.requestChan <- func(vm *goja.Runtime) {
				defer close(resolveDone)
				resolve(vm.ToValue(fmt.Sprintf("Server listening on %s", port)))
			}:
				<-resolveDone
			case <-time.After(5 * time.Second):
				fmt.Printf("Resolve timeout\n")
			}

			// 启动服务器
			if err := server.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				unregisterServer(server)
				// reject promise（通过 requestChan 确保线程安全）
				rejectDone := make(chan struct{})
				select {
				case server.requestChan <- func(vm *goja.Runtime) {
					defer close(rejectDone)
					reject(vm.NewGoError(err))
				}:
					<-rejectDone
				case <-time.After(5 * time.Second):
					fmt.Printf("Reject timeout\n")
				}
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

			// 通过 requestChan 调用回调函数和 resolve（确保线程安全）
			if callback != nil {
				if fn, ok := goja.AssertFunction(callback); ok {
					done := make(chan struct{})
					select {
					case server.requestChan <- func(vm *goja.Runtime) {
						defer close(done)
						defer func() {
							if r := recover(); r != nil {
								fmt.Printf("Callback panic: %v\n", r)
							}
						}()
						_, err := fn(goja.Undefined())
						if err != nil {
							fmt.Printf("Callback error: %v\n", err)
						}
					}:
						<-done
					case <-time.After(5 * time.Second):
						fmt.Printf("Callback timeout\n")
					}
				}
			}

			// resolve promise（通过 requestChan 确保线程安全）
			resolveDone := make(chan struct{})
			select {
			case server.requestChan <- func(vm *goja.Runtime) {
				defer close(resolveDone)
				resolve(vm.ToValue(fmt.Sprintf("HTTPS Server listening on %s", port)))
			}:
				<-resolveDone
			case <-time.After(5 * time.Second):
				fmt.Printf("Resolve timeout\n")
			}

			// 启动 HTTPS 服务器
			if err := server.server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
				unregisterServer(server)
				// reject promise（通过 requestChan 确保线程安全）
				rejectDone := make(chan struct{})
				select {
				case server.requestChan <- func(vm *goja.Runtime) {
					defer close(rejectDone)
					reject(vm.NewGoError(err))
				}:
					<-rejectDone
				case <-time.After(5 * time.Second):
					fmt.Printf("Reject timeout\n")
				}
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
		go func() {
			// 1. 关闭 HTTP 服务器
			if server.server != nil {
				ctx, cancel := context.WithTimeout(context.Background(), consts.DefaultHTTPTimeout)
				defer cancel()

				if err := server.server.Shutdown(ctx); err != nil {
					fmt.Printf("Server shutdown error: %v\n", err)
					return
				}
			}

			// 2. 停止 VM 处理器
			server.stopVMProcessor()

			// 3. 等待所有 goroutine 完成
			done := make(chan struct{})
			go func() {
				server.wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				fmt.Println("Server closed successfully")
			case <-time.After(10 * time.Second):
				fmt.Printf("Timeout waiting for goroutines to finish\n")
			}
		}()

		return goja.Undefined()
	}
}

// createHTTPHandler 创建 HTTP 处理器
// method 参数为空字符串时，会根据实际请求的方法动态查找对应的 handler
func (h *HTTPServerModule) createHTTPHandler(server *HTTPServer, method, path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 如果 method 为空，使用请求的实际方法
		actualMethod := method
		if actualMethod == "" {
			actualMethod = r.Method
		}

		server.mutex.RLock()
		routes := server.routes[actualMethod]
		middleware := server.middleware
		server.mutex.RUnlock()

		var handler goja.Value
		var params map[string]string
		found := false

		// 匹配路由
		for _, entry := range routes {
			if entry.pattern.isStatic {
				if entry.path == r.URL.Path {
					handler = entry.handler
					found = true
					break
				}
			} else {
				if p, match := entry.pattern.match(r.URL.Path); match {
					handler = entry.handler
					params = p
					found = true
					break
				}
			}
		}

		if !found {
			// 如果没有匹配到路径参数路由，尝试精确匹配
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
					fmt.Printf("Handler panic at %s: %v\n", path, r)
					if !rw.written {
						// 可以根据环境变量判断是否显示详细信息，这里先默认显示简略信息
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					}
				}
			}()

			// 在 VM goroutine 中创建请求和响应对象
			reqObj := h.createRequestObject(r)
			if params != nil {
				// 将路径参数注入到 req.params
				pObj := reqObj.ToObject(vm).Get("params").ToObject(vm)
				for k, v := range params {
					pObj.Set(k, v)
				}
			}
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

// createRequestObject 创建请求对象 (增强版)
func (h *HTTPServerModule) createRequestObject(r *http.Request) goja.Value {
	obj := h.vm.NewObject()

	// 1. 基本信息
	obj.Set("method", r.Method)
	obj.Set("url", r.URL.String())
	obj.Set("path", r.URL.Path)
	obj.Set("query", r.URL.RawQuery)
	obj.Set("originalUrl", r.URL.RequestURI())

	// 2. 协议与安全
	protocol := "http"
	if r.TLS != nil {
		protocol = "https"
	} else if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		protocol = proto
	}
	obj.Set("protocol", protocol)
	obj.Set("secure", protocol == "https")

	// 3. 主机名 (不含端口)
	host := r.Host
	if strings.Contains(host, ":") {
		h, _, _ := net.SplitHostPort(host)
		host = h
	}
	obj.Set("hostname", host)
	obj.Set("host", r.Host)

	// 4. XHR 判断
	obj.Set("xhr", strings.ToLower(r.Header.Get("X-Requested-With")) == "xmlhttprequest")

	// 5. 请求头
	headers := h.vm.NewObject()
	for key, values := range r.Header {
		if len(values) == 1 {
			headers.Set(key, values[0])
		} else {
			headers.Set(key, values)
		}
	}
	obj.Set("headers", headers)

	// 获取头部方法 (类似 Express 的 req.get)
	obj.Set("get", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			name := call.Arguments[0].String()
			return h.vm.ToValue(r.Header.Get(name))
		}
		return goja.Undefined()
	})

	// 6. Cookies 解析
	cookiesObj := h.vm.NewObject()
	cookies := r.Cookies()
	for _, cookie := range cookies {
		cookiesObj.Set(cookie.Name, cookie.Value)
	}
	obj.Set("cookies", cookiesObj)

	// 7. 查询参数
	params := h.vm.NewObject()
	for key, values := range r.URL.Query() {
		if len(values) == 1 {
			params.Set(key, values[0])
		} else {
			params.Set(key, values)
		}
	}
	obj.Set("params", params)

	// 8. Body 解析增强
	if r.Body != nil {
		// 注意：io.ReadAll 会消耗 Body，如果后续需要重新读取，这里需要处理
		// 目前方案是解析后存入对象。对于大 Body，可能需要流式支持
		body, err := io.ReadAll(r.Body)
		if err == nil {
			bodyStr := string(body)
			obj.Set("body", bodyStr)

			contentType := r.Header.Get("Content-Type")

			// JSON 解析
			if strings.Contains(contentType, "application/json") {
				var jsonData interface{}
				if json.Unmarshal(body, &jsonData) == nil {
					obj.Set("json", h.vm.ToValue(jsonData))
				}
			} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
				// Form 表单解析
				if values, err := url.ParseQuery(bodyStr); err == nil {
					formObj := h.vm.NewObject()
					for k, v := range values {
						if len(v) == 1 {
							formObj.Set(k, v[0])
						} else {
							formObj.Set(k, v)
						}
					}
					obj.Set("form", formObj)
				}
			}
		}
	}

	// 9. 类型检查方法 (类似 Express 的 req.is)
	obj.Set("is", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			target := call.Arguments[0].String()
			contentType := r.Header.Get("Content-Type")
			return h.vm.ToValue(strings.Contains(contentType, target))
		}
		return h.vm.ToValue(false)
	})

	// 10. 客户端信息
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
			h.sendFileResponse(w, rw, r, filePath)
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
			h.sendFileResponse(w, rw, r, filePath)
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

// sendFileResponse 发送文件响应 (优化为流式传输)
func (h *HTTPServerModule) sendFileResponse(w http.ResponseWriter, rw *responseWriter, r *http.Request, filePath string) {
	// 路径验证 - 防止路径遍历攻击
	cleanPath := filepath.Clean(filePath)

	// 检查是否包含路径遍历模式
	if strings.Contains(cleanPath, "..") {
		http.Error(w, "Access denied: path traversal not allowed", http.StatusForbidden)
		rw.written = true
		return
	}

	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		rw.written = true
		return
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error accessing file", http.StatusInternalServerError)
		}
		rw.written = true
		return
	}

	// 检查是否为目录
	if fileInfo.IsDir() {
		http.Error(w, "Cannot send directory", http.StatusBadRequest)
		rw.written = true
		return
	}

	// 设置响应状态码并使用 Go 标准库的 ServeFile
	// ServeFile 会自动处理 Content-Type, Range 请求, Last-Modified 等
	if rw.statusCode != 0 && rw.statusCode != 200 {
		w.WriteHeader(rw.statusCode)
	}
	http.ServeFile(w, r, absPath)
	rw.written = true
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
