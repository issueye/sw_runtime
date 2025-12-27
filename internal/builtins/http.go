package builtins

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dop251/goja"
)

// HTTPModule HTTP 客户端模块
type HTTPModule struct {
	vm                  *goja.Runtime
	client              *http.Client
	requestInterceptor  goja.Callable
	responseInterceptor goja.Callable
}

// NewHTTPModule 创建 HTTP 模块
func NewHTTPModule(vm *goja.Runtime) *HTTPModule {
	return &HTTPModule{
		vm: vm,
		client: &http.Client{
			Timeout: 30 * time.Second,
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

	// 状态码常量
	statusCodes := h.vm.NewObject()
	statusCodes.Set("OK", 200)
	statusCodes.Set("CREATED", 201)
	statusCodes.Set("NO_CONTENT", 204)
	statusCodes.Set("BAD_REQUEST", 400)
	statusCodes.Set("UNAUTHORIZED", 401)
	statusCodes.Set("FORBIDDEN", 403)
	statusCodes.Set("NOT_FOUND", 404)
	statusCodes.Set("INTERNAL_SERVER_ERROR", 500)
	obj.Set("STATUS_CODES", statusCodes)

	return obj
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
		}
	}

	return config
}

// makeRequest 执行 HTTP 请求
func (h *HTTPModule) makeRequest(config *HTTPConfig) (*HTTPResponse, error) {
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
	if config.Data != nil {
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

	// 创建请求
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout)*time.Second)
	defer cancel()

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
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 构建响应对象
	response := &HTTPResponse{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Headers:    make(map[string]string),
		Text:       string(respBody),
		URL:        reqURL,
		Config:     make(map[string]interface{}),
	}

	// 复制响应头
	for key, values := range resp.Header {
		if len(values) > 0 {
			response.Headers[key] = values[0]
		}
	}

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
