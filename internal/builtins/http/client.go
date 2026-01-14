package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dop251/goja"

	"sw_runtime/internal/consts"
	"sw_runtime/internal/security"
)

// HTTPModule HTTP 客户端模块
type HTTPModule struct {
	vm                  *goja.Runtime
	client              *http.Client
	urlValidator        *security.URLValidator
	requestInterceptor  goja.Callable
	responseInterceptor goja.Callable
}

// NewHTTPModule 创建 HTTP 模块
func NewHTTPModule(vm *goja.Runtime) *HTTPModule {
	return &HTTPModule{
		vm:           vm,
		urlValidator: security.NewURLValidator(), // 默认阻止内网访问
		client: &http.Client{
			Timeout: consts.DefaultHTTPTimeout,
		},
	}
}

// GetModule 获取 HTTP 模块对象
func (h *HTTPModule) GetModule() *goja.Object {
	obj := h.vm.NewObject()

	// HTTP 方法
	obj.Set("get", h.get)
	obj.Set("post", h.post)
	obj.Set("put", h.put)
	obj.Set("delete", h.delete)
	obj.Set("patch", h.patch)
	obj.Set("head", h.head)
	obj.Set("options", h.options)

	// 通用请求方法
	obj.Set("request", h.request)

	// 创建客户端实例
	obj.Set("createClient", h.createClient)

	// 拦截器
	obj.Set("setRequestInterceptor", h.setRequestInterceptor)
	obj.Set("setResponseInterceptor", h.setResponseInterceptor)

	// SSRF 保护配置
	obj.Set("allowPrivateNetwork", h.allowPrivateNetwork)
	obj.Set("addBlockedHost", h.addBlockedHost)
	obj.Set("addBlockedCIDR", h.addBlockedCIDR)

	// 状态码常量
	statusCodes := h.vm.NewObject()
	statusCodes.Set("OK", consts.StatusOK)
	statusCodes.Set("CREATED", consts.StatusCreated)
	statusCodes.Set("NO_CONTENT", consts.StatusNoContent)
	statusCodes.Set("BAD_REQUEST", consts.StatusBadRequest)
	statusCodes.Set("UNAUTHORIZED", consts.StatusUnauthorized)
	statusCodes.Set("FORBIDDEN", consts.StatusForbidden)
	statusCodes.Set("NOT_FOUND", consts.StatusNotFound)
	statusCodes.Set("INTERNAL_SERVER_ERROR", consts.StatusInternalServerError)
	obj.Set("STATUS_CODES", statusCodes)

	return obj
}

// allowPrivateNetwork 允许访问私有网络（仅用于开发环境）
func (h *HTTPModule) allowPrivateNetwork(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) > 0 {
		allow := call.Arguments[0].ToBoolean()
		if allow {
			h.urlValidator = security.NewURLValidatorWithPrivate()
		} else {
			h.urlValidator = security.NewURLValidator()
		}
	}
	return goja.Undefined()
}

// addBlockedHost 添加阻止的主机
func (h *HTTPModule) addBlockedHost(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) > 0 {
		host := call.Arguments[0].String()
		h.urlValidator.AddBlockedHost(host)
	}
	return goja.Undefined()
}

// addBlockedCIDR 添加阻止的 IP 网段
func (h *HTTPModule) addBlockedCIDR(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) > 0 {
		cidr := call.Arguments[0].String()
		if err := h.urlValidator.AddBlockedCIDR(cidr); err != nil {
			panic(h.vm.NewGoError(err))
		}
	}
	return goja.Undefined()
}

// HTTPResponse HTTP 响应结构
type HTTPResponse struct {
	Status     int                    `json:"status"`
	StatusText string                 `json:"statusText"`
	Headers    map[string]string      `json:"headers"`
	Data       interface{}            `json:"data"`
	Text       string                 `json:"text"`
	URL        string                 `json:"url"`
	Config     map[string]interface{} `json:"config"`
	Stream     *StreamResponse        `json:"-"`
}

// StreamResponse 流式响应结构（用于大文件下载）
type StreamResponse struct {
	vm        *goja.Runtime
	Body      io.ReadCloser
	Headers   map[string]string
	URL       string
	Status    int
	StatusText string
}

// Read 读取流式数据
func (s *StreamResponse) Read(call goja.FunctionCall) goja.Value {
	var n int
	var err error

	if len(call.Arguments) > 0 {
		size := call.Arguments[0].ToInteger()
		buf := make([]byte, size)
		n, err = s.Body.Read(buf)
		if err != nil && err != io.EOF {
			panic(s.vm.NewGoError(err))
		}
		if n == 0 {
			return s.vm.ToValue("")
		}
		return s.vm.ToValue(string(buf[:n]))
	}
	// 默认读取一块数据
	buf := make([]byte, 4096)
	n, err = s.Body.Read(buf)
	if err != nil && err != io.EOF {
		panic(s.vm.NewGoError(err))
	}
	if n == 0 {
		return s.vm.ToValue("")
	}
	return s.vm.ToValue(string(buf[:n]))
}

// Close 关闭流
func (s *StreamResponse) Close(call goja.FunctionCall) goja.Value {
	s.Body.Close()
	return goja.Undefined()
}

// PipeToFile 将流写入文件
func (s *StreamResponse) PipeToFile(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(s.vm.NewGoError(fmt.Errorf("file path required")))
	}
	filePath := call.Arguments[0].String()

	file, err := os.Create(filePath)
	if err != nil {
		panic(s.vm.NewGoError(err))
	}
	defer file.Close()

	_, err = io.Copy(file, s.Body)
	if err != nil {
		panic(s.vm.NewGoError(err))
	}
	return goja.Undefined()
}

// Copy 复制流到目标（支持自定义 writer）
func (s *StreamResponse) Copy(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(s.vm.NewGoError(fmt.Errorf("destination required")))
	}

	dest := call.Arguments[0].ToObject(s.vm)
	if dest == nil {
		panic(s.vm.NewGoError(fmt.Errorf("destination must be an object with write method")))
	}

	// 检查是否有 write 方法
	writeFn := dest.Get("write")
	if writeFn == nil || writeFn == goja.Undefined() {
		panic(s.vm.NewGoError(fmt.Errorf("destination must have a write method")))
	}

	writeCallable, ok := goja.AssertFunction(writeFn)
	if !ok {
		panic(s.vm.NewGoError(fmt.Errorf("write must be a function")))
	}

	// 使用 io.Copy 复制数据，通过 writer 回调处理
	writer := &callbackWriter{
		vm:   s.vm,
		fn:   writeCallable,
		dest: dest,
	}

	written, err := io.Copy(writer, s.Body)
	if err != nil {
		panic(s.vm.NewGoError(err))
	}

	return s.vm.ToValue(written)
}

// callbackWriter 包装 goja 函数为 io.Writer
type callbackWriter struct {
	vm   *goja.Runtime
	fn   goja.Callable
	dest *goja.Object
}

func (w *callbackWriter) Write(p []byte) (n int, err error) {
	// 跳过空数据
	if len(p) == 0 {
		return 0, nil
	}
	// 将字节切片转换为 JS 字符串
	data := string(p)
	_, err = w.fn(goja.Undefined(), w.vm.ToValue(data))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// HTTPConfig HTTP 请求配置
type HTTPConfig struct {
	Method            string                 `json:"method"`
	URL               string                 `json:"url"`
	Headers           map[string]string      `json:"headers"`
	Data              interface{}            `json:"data"`
	Params            map[string]string      `json:"params"`
	Timeout           int                    `json:"timeout"`
	Auth              map[string]string      `json:"auth"`
	Proxy             string                 `json:"proxy"`
	Cookies           map[string]string      `json:"cookies"`
	Config            map[string]interface{} `json:"config"`
	ResponseType      string                 `json:"responseType"` // "json" | "text" | "stream"
	FilePath          string                 `json:"filePath"`     // 上传文件路径
	BeforeRequest     goja.Callable          `json:"-"`
	AfterResponse     goja.Callable          `json:"-"`
	TransformRequest  goja.Callable          `json:"-"`
	TransformResponse goja.Callable          `json:"-"`
}

// parseConfig 解析请求配置
func (h *HTTPModule) parseConfig(args []goja.Value) *HTTPConfig {
	config := &HTTPConfig{
		Headers: make(map[string]string),
		Params:  make(map[string]string),
		Auth:    make(map[string]string),
		Cookies: make(map[string]string),
		Config:  make(map[string]interface{}),
		Timeout: 30,
	}

	if len(args) > 0 {
		config.URL = args[0].String()
	}

	if len(args) > 1 && args[1] != goja.Undefined() {
		configObj := args[1].ToObject(h.vm)
		if configObj != nil {
			// 解析各种配置选项
			if method := configObj.Get("method"); method != nil && method != goja.Undefined() {
				config.Method = method.String()
			}
			if headers := configObj.Get("headers"); headers != nil && headers != goja.Undefined() {
				headersObj := headers.ToObject(h.vm)
				if headersObj != nil {
					for _, key := range headersObj.Keys() {
						config.Headers[key] = headersObj.Get(key).String()
					}
				}
			}
			if data := configObj.Get("data"); data != nil && data != goja.Undefined() {
				config.Data = data.Export()
			}
			if params := configObj.Get("params"); params != nil && params != goja.Undefined() {
				paramsObj := params.ToObject(h.vm)
				if paramsObj != nil {
					for _, key := range paramsObj.Keys() {
						config.Params[key] = paramsObj.Get(key).String()
					}
				}
			}
			if timeout := configObj.Get("timeout"); timeout != nil && timeout != goja.Undefined() {
				config.Timeout = int(timeout.ToInteger())
			}
			if auth := configObj.Get("auth"); auth != nil && auth != goja.Undefined() {
				authObj := auth.ToObject(h.vm)
				if authObj != nil {
					for _, key := range authObj.Keys() {
						config.Auth[key] = authObj.Get(key).String()
					}
				}
			}
			// 解析拦截器
			if beforeRequest := configObj.Get("beforeRequest"); beforeRequest != nil && beforeRequest != goja.Undefined() {
				if fn, ok := goja.AssertFunction(beforeRequest); ok {
					config.BeforeRequest = fn
				}
			}
			if afterResponse := configObj.Get("afterResponse"); afterResponse != nil && afterResponse != goja.Undefined() {
				if fn, ok := goja.AssertFunction(afterResponse); ok {
					config.AfterResponse = fn
				}
			}
			if transformRequest := configObj.Get("transformRequest"); transformRequest != nil && transformRequest != goja.Undefined() {
				if fn, ok := goja.AssertFunction(transformRequest); ok {
					config.TransformRequest = fn
				}
			}
			if transformResponse := configObj.Get("transformResponse"); transformResponse != nil && transformResponse != goja.Undefined() {
				if fn, ok := goja.AssertFunction(transformResponse); ok {
					config.TransformResponse = fn
				}
			}
			// 解析响应类型（stream 用于大文件下载）
			if responseType := configObj.Get("responseType"); responseType != nil && responseType != goja.Undefined() {
				config.ResponseType = responseType.String()
			}
			// 解析文件路径（用于上传文件）
			if filePath := configObj.Get("filePath"); filePath != nil && filePath != goja.Undefined() {
				config.FilePath = filePath.String()
			}
		}
	}

	return config
}

// makeRequest 执行 HTTP 请求
func (h *HTTPModule) makeRequest(config *HTTPConfig) (*HTTPResponse, error) {
	// 验证 URL 安全性（防止 SSRF 攻击）
	if err := h.urlValidator.Validate(config.URL); err != nil {
		return nil, fmt.Errorf("URL validation failed: %w", err)
	}

	// 应用全局请求拦截器
	if h.requestInterceptor != nil {
		configObj := h.vm.ToValue(config).ToObject(h.vm)
		result, err := h.requestInterceptor(goja.Undefined(), configObj)
		if err != nil {
			return nil, err
		}
		// 更新配置
		if resultObj := result.ToObject(h.vm); resultObj != nil {
			if url := resultObj.Get("url"); url != nil && url != goja.Undefined() {
				config.URL = url.String()
				// 拦截器修改后也要验证 URL
				if err := h.urlValidator.Validate(config.URL); err != nil {
					return nil, fmt.Errorf("URL validation failed (after interceptor): %w", err)
				}
			}
			if headers := resultObj.Get("headers"); headers != nil && headers != goja.Undefined() {
				headersObj := headers.ToObject(h.vm)
				if headersObj != nil {
					config.Headers = make(map[string]string)
					for _, key := range headersObj.Keys() {
						config.Headers[key] = headersObj.Get(key).String()
					}
				}
			}
			if data := resultObj.Get("data"); data != nil && data != goja.Undefined() {
				config.Data = data.Export()
			}
			if params := resultObj.Get("params"); params != nil && params != goja.Undefined() {
				paramsObj := params.ToObject(h.vm)
				if paramsObj != nil {
					config.Params = make(map[string]string)
					for _, key := range paramsObj.Keys() {
						config.Params[key] = paramsObj.Get(key).String()
					}
				}
			}
		}
	}

	// 应用 beforeRequest 拦截器
	if config.BeforeRequest != nil {
		configObj := h.vm.ToValue(config).ToObject(h.vm)
		result, err := config.BeforeRequest(goja.Undefined(), configObj)
		if err != nil {
			return nil, err
		}
		// 更新配置
		if resultObj := result.ToObject(h.vm); resultObj != nil {
			if url := resultObj.Get("url"); url != nil && url != goja.Undefined() {
				config.URL = url.String()
				// beforeRequest 修改后也要验证 URL
				if err := h.urlValidator.Validate(config.URL); err != nil {
					return nil, fmt.Errorf("URL validation failed (after beforeRequest): %w", err)
				}
			}
			if headers := resultObj.Get("headers"); headers != nil && headers != goja.Undefined() {
				headersObj := headers.ToObject(h.vm)
				if headersObj != nil {
					config.Headers = make(map[string]string)
					for _, key := range headersObj.Keys() {
						config.Headers[key] = headersObj.Get(key).String()
					}
				}
			}
			if data := resultObj.Get("data"); data != nil && data != goja.Undefined() {
				config.Data = data.Export()
			}
			if params := resultObj.Get("params"); params != nil && params != goja.Undefined() {
				paramsObj := params.ToObject(h.vm)
				if paramsObj != nil {
					config.Params = make(map[string]string)
					for _, key := range paramsObj.Keys() {
						config.Params[key] = paramsObj.Get(key).String()
					}
				}
			}
		}
	}

	// 应用 transformRequest 拦截器
	if config.TransformRequest != nil && config.Data != nil {
		result, err := config.TransformRequest(goja.Undefined(), h.vm.ToValue(config.Data))
		if err != nil {
			return nil, err
		}
		config.Data = result.Export()
	}

	// 构建 URL
	reqURL := config.URL
	if len(config.Params) > 0 {
		u, err := url.Parse(reqURL)
		if err != nil {
			return nil, err
		}
		q := u.Query()
		for key, value := range config.Params {
			q.Add(key, value)
		}
		u.RawQuery = q.Encode()
		reqURL = u.String()
	}

	// 准备请求体
	var body io.Reader
	if config.FilePath != "" {
		// 文件上传模式
		file, err := os.Open(config.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		body = file
		if config.Headers["Content-Type"] == "" {
			// 根据文件扩展名推断 Content-Type
			contentType := "application/octet-stream"
			switch {
			case strings.HasSuffix(config.FilePath, ".json"):
				contentType = "application/json"
			case strings.HasSuffix(config.FilePath, ".xml"):
				contentType = "application/xml"
			case strings.HasSuffix(config.FilePath, ".txt"):
				contentType = "text/plain"
			case strings.HasSuffix(config.FilePath, ".html") || strings.HasSuffix(config.FilePath, ".htm"):
				contentType = "text/html"
			case strings.HasSuffix(config.FilePath, ".css"):
				contentType = "text/css"
			case strings.HasSuffix(config.FilePath, ".js"):
				contentType = "application/javascript"
			case strings.HasSuffix(config.FilePath, ".pdf"):
				contentType = "application/pdf"
			case strings.HasSuffix(config.FilePath, ".zip"):
				contentType = "application/zip"
			case strings.HasSuffix(config.FilePath, ".png"):
				contentType = "image/png"
			case strings.HasSuffix(config.FilePath, ".jpg") || strings.HasSuffix(config.FilePath, ".jpeg"):
				contentType = "image/jpeg"
			case strings.HasSuffix(config.FilePath, ".gif"):
				contentType = "image/gif"
			case strings.HasSuffix(config.FilePath, ".ts"):
				contentType = "video/mp2t"
			case strings.HasSuffix(config.FilePath, ".m4s"):
				contentType = "video/mp4"
			}
			config.Headers["Content-Type"] = contentType
		}
	} else if config.Data != nil {
		switch data := config.Data.(type) {
		case string:
			body = strings.NewReader(data)
		case map[string]interface{}:
			jsonData, err := json.Marshal(data)
			if err != nil {
				return nil, err
			}
			body = bytes.NewReader(jsonData)
			if config.Headers["Content-Type"] == "" {
				config.Headers["Content-Type"] = "application/json"
			}
		default:
			jsonData, err := json.Marshal(data)
			if err != nil {
				return nil, err
			}
			body = bytes.NewReader(jsonData)
			if config.Headers["Content-Type"] == "" {
				config.Headers["Content-Type"] = "application/json"
			}
		}
	}

	// 流式响应模式自动禁用超时
	if config.ResponseType == "stream" {
		config.Timeout = 0
	}

	// 创建请求（timeout <= 0 表示不超时）
	var ctx context.Context
	var cancel context.CancelFunc
	if config.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
		defer cancel()
	} else {
		ctx = context.Background()
	}

	req, err := http.NewRequestWithContext(ctx, config.Method, reqURL, body)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// 设置认证
	if username, ok := config.Auth["username"]; ok {
		if password, ok := config.Auth["password"]; ok {
			req.SetBasicAuth(username, password)
		}
	}
	if token, ok := config.Auth["token"]; ok {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// 执行请求
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, err
	}

	// 构建响应对象
	response := &HTTPResponse{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Headers:    make(map[string]string),
		URL:        reqURL,
		Config:     make(map[string]interface{}),
	}

	// 复制响应头
	for key, values := range resp.Header {
		if len(values) > 0 {
			response.Headers[key] = values[0]
		}
	}

	// 流式响应模式
	if config.ResponseType == "stream" {
		// 创建 StreamResponse，保留 Body 不关闭（由用户调用 close）
		streamResponse := &StreamResponse{
			vm:         h.vm,
			Body:       resp.Body,
			Headers:    response.Headers,
			URL:        reqURL,
			Status:     resp.StatusCode,
			StatusText: resp.Status,
		}

		// 将 StreamResponse 暴露给 JS
		streamObj := h.vm.NewObject()
		streamObj.Set("read", streamResponse.Read)
		streamObj.Set("close", streamResponse.Close)
		streamObj.Set("pipeToFile", streamResponse.PipeToFile)
		streamObj.Set("copy", streamResponse.Copy)
		streamObj.Set("headers", h.vm.ToValue(response.Headers))
		streamObj.Set("status", response.Status)
		streamObj.Set("statusText", response.StatusText)
		streamObj.Set("url", response.URL)

		// 在 Stream 中存储引用以便调用方法
		response.Stream = streamResponse
		response.Data = streamObj
		return response, nil
	}

	// 非流式响应：关闭 Body 并读取内容
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	response.Text = string(respBody)

	// 尝试解析 JSON
	var jsonData interface{}
	if err := json.Unmarshal(respBody, &jsonData); err == nil {
		response.Data = jsonData
	} else {
		response.Data = string(respBody)
	}

	// 应用 transformResponse 拦截器
	if config.TransformResponse != nil {
		result, err := config.TransformResponse(goja.Undefined(), h.vm.ToValue(response.Data))
		if err == nil {
			response.Data = result.Export()
		}
	}

	// 应用 afterResponse 拦截器
	if config.AfterResponse != nil {
		responseObj := h.vm.ToValue(response).ToObject(h.vm)
		result, err := config.AfterResponse(goja.Undefined(), responseObj)
		if err == nil {
			if resultObj := result.ToObject(h.vm); resultObj != nil {
				if data := resultObj.Get("data"); data != nil && data != goja.Undefined() {
					response.Data = data.Export()
				}
				if headers := resultObj.Get("headers"); headers != nil && headers != goja.Undefined() {
					headersObj := headers.ToObject(h.vm)
					if headersObj != nil {
						response.Headers = make(map[string]string)
						for _, key := range headersObj.Keys() {
							response.Headers[key] = headersObj.Get(key).String()
						}
					}
				}
			}
		}
	}

	// 应用全局响应拦截器
	if h.responseInterceptor != nil {
		responseObj := h.vm.ToValue(response).ToObject(h.vm)
		result, err := h.responseInterceptor(goja.Undefined(), responseObj)
		if err == nil {
			if resultObj := result.ToObject(h.vm); resultObj != nil {
				if data := resultObj.Get("data"); data != nil && data != goja.Undefined() {
					response.Data = data.Export()
				}
				if headers := resultObj.Get("headers"); headers != nil && headers != goja.Undefined() {
					headersObj := headers.ToObject(h.vm)
					if headersObj != nil {
						response.Headers = make(map[string]string)
						for _, key := range headersObj.Keys() {
							response.Headers[key] = headersObj.Get(key).String()
						}
					}
				}
			}
		}
	}

	return response, nil
}

// get GET 请求
func (h *HTTPModule) get(call goja.FunctionCall) goja.Value {
	config := h.parseConfig(call.Arguments)
	config.Method = "GET"

	promise, resolve, reject := h.vm.NewPromise()

	go func() {
		response, err := h.makeRequest(config)
		if err != nil {
			reject(h.vm.NewGoError(err))
		} else {
			resolve(h.vm.ToValue(response))
		}
	}()

	return h.vm.ToValue(promise)
}

// post POST 请求
func (h *HTTPModule) post(call goja.FunctionCall) goja.Value {
	config := h.parseConfig(call.Arguments)
	config.Method = "POST"

	promise, resolve, reject := h.vm.NewPromise()

	go func() {
		response, err := h.makeRequest(config)
		if err != nil {
			reject(h.vm.NewGoError(err))
		} else {
			resolve(h.vm.ToValue(response))
		}
	}()

	return h.vm.ToValue(promise)
}

// put PUT 请求
func (h *HTTPModule) put(call goja.FunctionCall) goja.Value {
	config := h.parseConfig(call.Arguments)
	config.Method = "PUT"

	promise, resolve, reject := h.vm.NewPromise()

	go func() {
		response, err := h.makeRequest(config)
		if err != nil {
			reject(h.vm.NewGoError(err))
		} else {
			resolve(h.vm.ToValue(response))
		}
	}()

	return h.vm.ToValue(promise)
}

// delete DELETE 请求
func (h *HTTPModule) delete(call goja.FunctionCall) goja.Value {
	config := h.parseConfig(call.Arguments)
	config.Method = "DELETE"

	promise, resolve, reject := h.vm.NewPromise()

	go func() {
		response, err := h.makeRequest(config)
		if err != nil {
			reject(h.vm.NewGoError(err))
		} else {
			resolve(h.vm.ToValue(response))
		}
	}()

	return h.vm.ToValue(promise)
}

// patch PATCH 请求
func (h *HTTPModule) patch(call goja.FunctionCall) goja.Value {
	config := h.parseConfig(call.Arguments)
	config.Method = "PATCH"

	promise, resolve, reject := h.vm.NewPromise()

	go func() {
		response, err := h.makeRequest(config)
		if err != nil {
			reject(h.vm.NewGoError(err))
		} else {
			resolve(h.vm.ToValue(response))
		}
	}()

	return h.vm.ToValue(promise)
}

// head HEAD 请求
func (h *HTTPModule) head(call goja.FunctionCall) goja.Value {
	config := h.parseConfig(call.Arguments)
	config.Method = "HEAD"

	promise, resolve, reject := h.vm.NewPromise()

	go func() {
		response, err := h.makeRequest(config)
		if err != nil {
			reject(h.vm.NewGoError(err))
		} else {
			resolve(h.vm.ToValue(response))
		}
	}()

	return h.vm.ToValue(promise)
}

// options OPTIONS 请求
func (h *HTTPModule) options(call goja.FunctionCall) goja.Value {
	config := h.parseConfig(call.Arguments)
	config.Method = "OPTIONS"

	promise, resolve, reject := h.vm.NewPromise()

	go func() {
		response, err := h.makeRequest(config)
		if err != nil {
			reject(h.vm.NewGoError(err))
		} else {
			resolve(h.vm.ToValue(response))
		}
	}()

	return h.vm.ToValue(promise)
}

// request 通用请求方法
func (h *HTTPModule) request(call goja.FunctionCall) goja.Value {
	config := h.parseConfig(call.Arguments)
	if config.Method == "" {
		config.Method = "GET"
	}

	promise, resolve, reject := h.vm.NewPromise()

	go func() {
		response, err := h.makeRequest(config)
		if err != nil {
			reject(h.vm.NewGoError(err))
		} else {
			resolve(h.vm.ToValue(response))
		}
	}()

	return h.vm.ToValue(promise)
}

// createClient 创建自定义 HTTP 客户端
func (h *HTTPModule) createClient(call goja.FunctionCall) goja.Value {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if len(call.Arguments) > 0 && call.Arguments[0] != goja.Undefined() {
		configObj := call.Arguments[0].ToObject(h.vm)
		if configObj != nil {
			if timeout := configObj.Get("timeout"); timeout != nil && timeout != goja.Undefined() {
				client.Timeout = time.Duration(timeout.ToInteger()) * time.Second
			}
		}
	}

	// 创建客户端实例对象
	clientObj := h.vm.NewObject()
	httpModule := &HTTPModule{vm: h.vm, client: client}

	clientObj.Set("get", httpModule.get)
	clientObj.Set("post", httpModule.post)
	clientObj.Set("put", httpModule.put)
	clientObj.Set("delete", httpModule.delete)
	clientObj.Set("patch", httpModule.patch)
	clientObj.Set("head", httpModule.head)
	clientObj.Set("options", httpModule.options)
	clientObj.Set("request", httpModule.request)

	return clientObj
}

// setRequestInterceptor 设置请求拦截器
func (h *HTTPModule) setRequestInterceptor(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) > 0 {
		if fn, ok := goja.AssertFunction(call.Arguments[0]); ok {
			h.requestInterceptor = fn
		}
	}
	return goja.Undefined()
}

// setResponseInterceptor 设置响应拦截器
func (h *HTTPModule) setResponseInterceptor(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) > 0 {
		if fn, ok := goja.AssertFunction(call.Arguments[0]); ok {
			h.responseInterceptor = fn
		}
	}
	return goja.Undefined()
}
