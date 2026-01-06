package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	stdruntime "runtime"

	rt "sw_runtime/internal/runtime"
)

// scriptsDir 保存边缘脚本所在的目录路径（与本示例 main.go 同级的 scripts 目录）。
var scriptsDir string

func init() {
	// 尽量用当前文件位置推导脚本目录，避免受工作目录影响。
	if _, file, _, ok := stdruntime.Caller(0); ok {
		base := filepath.Dir(file)
		scriptsDir = filepath.Join(base, "scripts")
	} else {
		scriptsDir = "scripts"
	}
}

// 简单的请求上下文，传给 JS 侧使用。
type EdgeRequest struct {
	Method string            `json:"method"`
	Path   string            `json:"path"`
	Query  map[string]string `json:"query"`
	Body   string            `json:"body,omitempty"`
}

// JS 约定的响应结构。
type EdgeResponse struct {
	Status  int               `json:"status,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
	JSON    any               `json:"json,omitempty"`
}

var runnerPool = rt.NewRunnerPool()

func main() {
	http.HandleFunc("/edge/", edgeHandler)

	addr := ":8080"
	log.Printf("[edge-service] listening on %s", addr)
	log.Printf("try: curl " + "http://localhost:8080/edge/hello?name=SW-Runtime")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("http server error: %v", err)
	}
}

func edgeHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析脚本名称：/edge/{script}
	name := strings.TrimPrefix(r.URL.Path, "/edge/")
	if name == "" {
		name = "hello" // 默认脚本名 hello-edge.ts
	}

	scriptPath := filepath.Join(scriptsDir, fmt.Sprintf("%s-edge.ts", name))
	if _, err := os.Stat(scriptPath); err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "script not found: %s\n", scriptPath)
		return
	}

	// 2. 从池中获取 Runner（每个请求独占一个 Runner，避免并发冲突）。
	runner := rt.AcquireRunner()
	defer rt.ReleaseRunner(runner)

	// 3. 构造给 JS 的请求上下文
	bodyBytes, _ := io.ReadAll(r.Body)
	_ = r.Body.Close()

	query := make(map[string]string, len(r.URL.Query()))
	for k, vs := range r.URL.Query() {
		if len(vs) > 0 {
			query[k] = vs[0]
		}
	}

	edgeReq := EdgeRequest{
		Method: r.Method,
		Path:   r.URL.Path,
		Query:  query,
		Body:   string(bodyBytes),
	}

	// 注入到 JS 全局变量：global.request
	runner.SetValue("request", edgeReq)
	// 清理上一次可能遗留的 response
	runner.SetValue("response", nil)

	// 4. 安全执行脚本文件（捕获 goja panic）
	if err := runner.SafeRunFile(scriptPath); err != nil {
		log.Printf("edge script panic (%s): %v", scriptPath, err)
		http.Error(w, "edge script failed", http.StatusInternalServerError)
		return
	}

	// 5. 从 JS 中读取响应
	respVal := runner.GetValue("response")
	if respVal == nil || respVal.Export() == nil {
		// 如果脚本没有设置 response，就返回一个简单 JSON
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "no response set by script",
			"path":    r.URL.Path,
		})
		return
	}

	exported := respVal.Export()
	respMap, ok := exported.(map[string]any)
	if !ok {
		log.Printf("response has unexpected type %T, falling back to JSON", exported)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(exported)
		return
	}

	// 解析 EdgeResponse 结构
	var edgeResp EdgeResponse
	if b, err := json.Marshal(respMap); err == nil {
		_ = json.Unmarshal(b, &edgeResp)
	}

	status := edgeResp.Status
	if status == 0 {
		status = http.StatusOK
	}

	for k, v := range edgeResp.Headers {
		w.Header().Set(k, v)
	}

	// 自动根据是否有 JSON 字段选择返回 JSON 或纯文本
	if edgeResp.JSON != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(edgeResp.JSON)
		return
	}

	w.WriteHeader(status)
	if edgeResp.Body != "" {
		_, _ = io.WriteString(w, edgeResp.Body)
	} else {
		_, _ = io.WriteString(w, "ok")
	}
}
