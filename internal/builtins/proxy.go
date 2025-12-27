package builtins

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dop251/goja"
)

// ProxyModule 代理模块
type ProxyModule struct {
	vm      *goja.Runtime
	proxies map[string]*ProxyServer
	mutex   sync.RWMutex
}

// ProxyServer 代理服务器
type ProxyServer struct {
	server    *http.Server
	vm        *goja.Runtime
	handlers  map[string]goja.Value // event -> handler
	mutex     sync.RWMutex
	proxyType string // "http" or "tcp"
}

// TCPProxy TCP 代理
type TCPProxy struct {
	listener net.Listener
	vm       *goja.Runtime
	target   string
	handlers map[string]goja.Value
	mutex    sync.RWMutex
	closed   bool
}

// NewProxyModule 创建代理模块
func NewProxyModule(vm *goja.Runtime) *ProxyModule {
	return &ProxyModule{
		vm:      vm,
		proxies: make(map[string]*ProxyServer),
	}
}

// GetModule 获取代理模块对象
func (p *ProxyModule) GetModule() *goja.Object {
	obj := p.vm.NewObject()

	// HTTP 代理
	obj.Set("createHTTPProxy", p.createHTTPProxy)

	// TCP 代理
	obj.Set("createTCPProxy", p.createTCPProxy)

	return obj
}

// createHTTPProxy 创建 HTTP 代理服务器
func (p *ProxyModule) createHTTPProxy(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(p.vm.NewTypeError("createHTTPProxy requires target URL"))
	}

	targetURL := call.Arguments[0].String()

	// 解析目标 URL
	target, err := url.Parse(targetURL)
	if err != nil {
		panic(p.vm.NewGoError(fmt.Errorf("invalid target URL: %v", err)))
	}

	// 创建代理服务器
	proxy := &ProxyServer{
		vm:        p.vm,
		handlers:  make(map[string]goja.Value),
		proxyType: "http",
	}

	// 创建反向代理
	reverseProxy := httputil.NewSingleHostReverseProxy(target)

	// 自定义 Transport 以支持 HTTPS
	reverseProxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// 请求修改器
	originalDirector := reverseProxy.Director
	reverseProxy.Director = func(req *http.Request) {
		originalDirector(req)

		// 触发 request 事件
		proxy.mutex.RLock()
		handler, exists := proxy.handlers["request"]
		proxy.mutex.RUnlock()

		if exists {
			if fn, ok := goja.AssertFunction(handler); ok {
				reqObj := p.vm.NewObject()
				reqObj.Set("method", req.Method)
				reqObj.Set("url", req.URL.String())
				reqObj.Set("path", req.URL.Path)
				reqObj.Set("host", req.Host)
				reqObj.Set("remoteAddr", req.RemoteAddr)

				// 请求头
				headers := p.vm.NewObject()
				for k, v := range req.Header {
					if len(v) > 0 {
						headers.Set(k, v[0])
					}
				}
				reqObj.Set("headers", headers)

				_, err := fn(goja.Undefined(), p.vm.ToValue(reqObj))
				if err != nil {
					fmt.Printf("Request handler error: %v\n", err)
				}
			}
		}
	}

	// 响应修改器
	reverseProxy.ModifyResponse = func(resp *http.Response) error {
		// 触发 response 事件
		proxy.mutex.RLock()
		handler, exists := proxy.handlers["response"]
		proxy.mutex.RUnlock()

		if exists {
			if fn, ok := goja.AssertFunction(handler); ok {
				respObj := p.vm.NewObject()
				respObj.Set("status", resp.StatusCode)
				respObj.Set("statusText", resp.Status)

				// 响应头
				headers := p.vm.NewObject()
				for k, v := range resp.Header {
					if len(v) > 0 {
						headers.Set(k, v[0])
					}
				}
				respObj.Set("headers", headers)

				_, err := fn(goja.Undefined(), p.vm.ToValue(respObj))
				if err != nil {
					fmt.Printf("Response handler error: %v\n", err)
				}
			}
		}
		return nil
	}

	// 错误处理器
	reverseProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		// 触发 error 事件
		proxy.mutex.RLock()
		handler, exists := proxy.handlers["error"]
		proxy.mutex.RUnlock()

		if exists {
			if fn, ok := goja.AssertFunction(handler); ok {
				errObj := p.vm.NewObject()
				errObj.Set("message", err.Error())
				errObj.Set("url", r.URL.String())

				_, handlerErr := fn(goja.Undefined(), p.vm.ToValue(errObj))
				if handlerErr != nil {
					fmt.Printf("Error handler error: %v\n", handlerErr)
				}
			}
		}

		// 默认错误响应
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "Proxy Error: %v", err)
	}

	// 创建 HTTP 服务器
	mux := http.NewServeMux()
	mux.HandleFunc("/", reverseProxy.ServeHTTP)

	proxy.server = &http.Server{
		Handler: mux,
	}

	return p.vm.ToValue(p.createProxyObject(proxy))
}

// createTCPProxy 创建 TCP 代理服务器
func (p *ProxyModule) createTCPProxy(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(p.vm.NewTypeError("createTCPProxy requires target address"))
	}

	target := call.Arguments[0].String()

	tcpProxy := &TCPProxy{
		vm:       p.vm,
		target:   target,
		handlers: make(map[string]goja.Value),
	}

	return p.vm.ToValue(p.createTCPProxyObject(tcpProxy))
}

// createProxyObject 创建 HTTP 代理对象
func (p *ProxyModule) createProxyObject(proxy *ProxyServer) *goja.Object {
	obj := p.vm.NewObject()

	// 注册事件监听器
	obj.Set("on", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(p.vm.NewTypeError("on requires event name and handler"))
		}

		event := call.Arguments[0].String()
		handler := call.Arguments[1]

		proxy.mutex.Lock()
		proxy.handlers[event] = handler
		proxy.mutex.Unlock()

		return goja.Undefined()
	})

	// 启动代理服务器
	obj.Set("listen", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(p.vm.NewTypeError("listen requires port"))
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

		promise, resolve, reject := p.vm.NewPromise()

		go func() {
			proxy.server.Addr = port

			p.mutex.Lock()
			p.proxies[port] = proxy
			p.mutex.Unlock()

			// 调用回调函数
			if callback != nil {
				if fn, ok := goja.AssertFunction(callback); ok {
					_, err := fn(goja.Undefined())
					if err != nil {
						fmt.Printf("Callback error: %v\n", err)
					}
				}
			}

			resolve(p.vm.ToValue(fmt.Sprintf("HTTP Proxy listening on %s", port)))

			// 启动服务器
			if err := proxy.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				reject(p.vm.NewGoError(err))
			}
		}()

		return p.vm.ToValue(promise)
	})

	// 关闭代理服务器
	obj.Set("close", func(call goja.FunctionCall) goja.Value {
		promise, resolve, reject := p.vm.NewPromise()

		go func() {
			if proxy.server != nil {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := proxy.server.Shutdown(ctx); err != nil {
					reject(p.vm.NewGoError(err))
					return
				}
			}
			resolve(p.vm.ToValue("Proxy server closed"))
		}()

		return p.vm.ToValue(promise)
	})

	return obj
}

// createTCPProxyObject 创建 TCP 代理对象
func (p *ProxyModule) createTCPProxyObject(tcpProxy *TCPProxy) *goja.Object {
	obj := p.vm.NewObject()

	// 注册事件监听器
	obj.Set("on", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(p.vm.NewTypeError("on requires event name and handler"))
		}

		event := call.Arguments[0].String()
		handler := call.Arguments[1]

		tcpProxy.mutex.Lock()
		tcpProxy.handlers[event] = handler
		tcpProxy.mutex.Unlock()

		return goja.Undefined()
	})

	// 启动 TCP 代理服务器
	obj.Set("listen", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(p.vm.NewTypeError("listen requires port"))
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

		promise, resolve, reject := p.vm.NewPromise()

		go func() {
			listener, err := net.Listen("tcp", port)
			if err != nil {
				reject(p.vm.NewGoError(err))
				return
			}

			tcpProxy.listener = listener

			// 调用回调函数
			if callback != nil {
				if fn, ok := goja.AssertFunction(callback); ok {
					_, err := fn(goja.Undefined())
					if err != nil {
						fmt.Printf("Callback error: %v\n", err)
					}
				}
			}

			resolve(p.vm.ToValue(fmt.Sprintf("TCP Proxy listening on %s", port)))

			// 接受连接
			for {
				if tcpProxy.closed {
					break
				}

				conn, err := listener.Accept()
				if err != nil {
					if !tcpProxy.closed {
						// 触发 error 事件
						tcpProxy.mutex.RLock()
						handler, exists := tcpProxy.handlers["error"]
						tcpProxy.mutex.RUnlock()

						if exists {
							if fn, ok := goja.AssertFunction(handler); ok {
								errObj := p.vm.NewObject()
								errObj.Set("message", err.Error())
								_, _ = fn(goja.Undefined(), p.vm.ToValue(errObj))
							}
						}
					}
					continue
				}

				// 处理连接
				go p.handleTCPProxyConnection(tcpProxy, conn)
			}
		}()

		return p.vm.ToValue(promise)
	})

	// 关闭 TCP 代理服务器
	obj.Set("close", func(call goja.FunctionCall) goja.Value {
		promise, resolve, reject := p.vm.NewPromise()

		go func() {
			tcpProxy.closed = true
			if tcpProxy.listener != nil {
				if err := tcpProxy.listener.Close(); err != nil {
					reject(p.vm.NewGoError(err))
					return
				}
			}
			resolve(p.vm.ToValue("TCP Proxy server closed"))
		}()

		return p.vm.ToValue(promise)
	})

	return obj
}

// handleTCPProxyConnection 处理 TCP 代理连接
func (p *ProxyModule) handleTCPProxyConnection(tcpProxy *TCPProxy, clientConn net.Conn) {
	defer clientConn.Close()

	// 连接到目标服务器
	targetConn, err := net.DialTimeout("tcp", tcpProxy.target, 10*time.Second)
	if err != nil {
		// 触发 error 事件
		tcpProxy.mutex.RLock()
		handler, exists := tcpProxy.handlers["error"]
		tcpProxy.mutex.RUnlock()

		if exists {
			if fn, ok := goja.AssertFunction(handler); ok {
				errObj := p.vm.NewObject()
				errObj.Set("message", fmt.Sprintf("Failed to connect to target: %v", err))
				_, _ = fn(goja.Undefined(), p.vm.ToValue(errObj))
			}
		}
		return
	}
	defer targetConn.Close()

	// 触发 connection 事件
	tcpProxy.mutex.RLock()
	connHandler, exists := tcpProxy.handlers["connection"]
	tcpProxy.mutex.RUnlock()

	if exists {
		if fn, ok := goja.AssertFunction(connHandler); ok {
			connObj := p.vm.NewObject()
			connObj.Set("remoteAddr", clientConn.RemoteAddr().String())
			connObj.Set("target", tcpProxy.target)
			_, _ = fn(goja.Undefined(), p.vm.ToValue(connObj))
		}
	}

	// 双向转发数据
	var wg sync.WaitGroup
	wg.Add(2)

	// 客户端 -> 目标服务器
	go func() {
		defer wg.Done()
		bytesTransferred, err := io.Copy(targetConn, clientConn)

		// 触发 data 事件
		tcpProxy.mutex.RLock()
		dataHandler, exists := tcpProxy.handlers["data"]
		tcpProxy.mutex.RUnlock()

		if exists && bytesTransferred > 0 {
			if fn, ok := goja.AssertFunction(dataHandler); ok {
				dataObj := p.vm.NewObject()
				dataObj.Set("direction", "client->target")
				dataObj.Set("bytes", bytesTransferred)
				_, _ = fn(goja.Undefined(), p.vm.ToValue(dataObj))
			}
		}

		if err != nil && err != io.EOF {
			// 触发 error 事件
			tcpProxy.mutex.RLock()
			errHandler, exists := tcpProxy.handlers["error"]
			tcpProxy.mutex.RUnlock()

			if exists {
				if fn, ok := goja.AssertFunction(errHandler); ok {
					errObj := p.vm.NewObject()
					errObj.Set("message", err.Error())
					errObj.Set("direction", "client->target")
					_, _ = fn(goja.Undefined(), p.vm.ToValue(errObj))
				}
			}
		}
	}()

	// 目标服务器 -> 客户端
	go func() {
		defer wg.Done()
		bytesTransferred, err := io.Copy(clientConn, targetConn)

		// 触发 data 事件
		tcpProxy.mutex.RLock()
		dataHandler, exists := tcpProxy.handlers["data"]
		tcpProxy.mutex.RUnlock()

		if exists && bytesTransferred > 0 {
			if fn, ok := goja.AssertFunction(dataHandler); ok {
				dataObj := p.vm.NewObject()
				dataObj.Set("direction", "target->client")
				dataObj.Set("bytes", bytesTransferred)
				_, _ = fn(goja.Undefined(), p.vm.ToValue(dataObj))
			}
		}

		if err != nil && err != io.EOF {
			// 触发 error 事件
			tcpProxy.mutex.RLock()
			errHandler, exists := tcpProxy.handlers["error"]
			tcpProxy.mutex.RUnlock()

			if exists {
				if fn, ok := goja.AssertFunction(errHandler); ok {
					errObj := p.vm.NewObject()
					errObj.Set("message", err.Error())
					errObj.Set("direction", "target->client")
					_, _ = fn(goja.Undefined(), p.vm.ToValue(errObj))
				}
			}
		}
	}()

	wg.Wait()

	// 触发 close 事件
	tcpProxy.mutex.RLock()
	closeHandler, exists := tcpProxy.handlers["close"]
	tcpProxy.mutex.RUnlock()

	if exists {
		if fn, ok := goja.AssertFunction(closeHandler); ok {
			_, _ = fn(goja.Undefined())
		}
	}
}

// 全局变量跟踪代理服务器运行状态
var proxyServerRunning int32

// IsProxyServerRunning 检查是否有代理服务器在运行
func IsProxyServerRunning() bool {
	return atomic.LoadInt32(&proxyServerRunning) > 0
}
