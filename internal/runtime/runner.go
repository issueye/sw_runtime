package runtime

import (
	"fmt"
	"os"
	"path/filepath"

	"sw_runtime/internal/modules"
	"sw_runtime/internal/pool"

	"github.com/dop251/goja"
)

// Runner JavaScript/TypeScript 运行器
type Runner struct {
	vm      *goja.Runtime
	loop    *SimpleEventLoop
	modules *modules.System
}

// New 创建新的运行器
func New() (*Runner, error) {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	loop := NewSimpleEventLoop(vm)

	// 获取当前工作目录作为基础路径
	basePath, err := os.Getwd()
	if err != nil {
		// 如果无法获取当前目录，使用临时目录作为后备
		basePath = os.TempDir()
	}

	moduleSystem := modules.NewSystem(vm, basePath)

	r := &Runner{
		vm:      vm,
		loop:    loop,
		modules: moduleSystem,
	}
	r.setupBuiltins()

	// 增加 Runner 计数
	pool.GlobalMemoryMonitor.IncrementRunnerCount()

	return r, nil
}

// NewOrPanic 创建新的运行器，出错时 panic（为了向后兼容）
func NewOrPanic() *Runner {
	r, err := New()
	if err != nil {
		panic(err)
	}
	return r
}

// setupBuiltins 注册内置函数
func (r *Runner) setupBuiltins() {
	// console 对象
	console := r.vm.NewObject()
	console.Set("log", func(call goja.FunctionCall) goja.Value {
		args := make([]interface{}, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.Export()
		}
		fmt.Println(args...)
		return goja.Undefined()
	})
	console.Set("error", func(call goja.FunctionCall) goja.Value {
		args := make([]interface{}, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.Export()
		}
		fmt.Fprintln(os.Stderr, args...)
		return goja.Undefined()
	})
	console.Set("warn", func(call goja.FunctionCall) goja.Value {
		args := make([]interface{}, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.Export()
		}
		fmt.Println("[WARN]", args)
		return goja.Undefined()
	})
	r.vm.Set("console", console)

	// 定时器函数 - 使用事件循环
	r.vm.Set("setTimeout", r.loop.SetTimeout)
	r.vm.Set("clearTimeout", r.loop.ClearTimeout)
	r.vm.Set("setInterval", r.loop.SetInterval)
	r.vm.Set("clearInterval", r.loop.ClearInterval)

	// 模块系统
	r.vm.Set("require", r.modules.Require)

	// 动态 import 支持
	r.vm.Set("import", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			panic(r.vm.NewTypeError("import() missing path"))
		}

		id := call.Arguments[0].String()

		// 创建 Promise
		promise, resolve, reject := r.vm.NewPromise()

		// 异步加载模块
		go func() {
			currentDir, err := os.Getwd()
			if err != nil {
				currentDir = "."
			}
			module, err := r.modules.LoadModule(id, currentDir)
			if err != nil {
				reject(r.vm.NewGoError(err))
			} else {
				resolve(module.Exports)
			}
		}()

		return r.vm.ToValue(promise)
	})

	// 全局变量
	r.vm.Set("global", r.vm.GlobalObject())
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	r.vm.Set("__dirname", dir)
	r.vm.Set("__filename", "")

	// 启用 Promise
	r.vm.SetPromiseRejectionTracker(func(p *goja.Promise, op goja.PromiseRejectionOperation) {
		if op == goja.PromiseRejectionReject {
			fmt.Fprintf(os.Stderr, "Unhandled promise rejection: %v\n", p.Result())
		}
	})
}

// RunCode 执行 TypeScript/JavaScript 代码
func (r *Runner) RunCode(code string) error {
	// 尝试作为 TypeScript 编译
	jsCode, err := transpileTS(code, "inline.ts")
	if err != nil {
		// 如果编译失败，尝试直接作为 JS 执行
		jsCode = code
	}

	r.loop.Start()
	_, err = r.vm.RunString(jsCode)
	if err != nil {
		return err
	}

	// 处理异步任务
	r.loop.WaitAndProcess()
	return nil
}

// RunFile 执行 TypeScript/JavaScript 文件
func (r *Runner) RunFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	code := string(content)
	ext := filepath.Ext(filename)

	// 如果是 .ts 或 .tsx 文件，先编译
	if ext == ".ts" || ext == ".tsx" {
		code, err = transpileTS(code, filename)
		if err != nil {
			return err
		}
	}

	r.loop.Start()
	_, err = r.vm.RunString(code)
	if err != nil {
		return err
	}

	// 处理异步任务
	r.loop.WaitAndProcess()
	return nil
}

// SetValue 设置全局变量
func (r *Runner) SetValue(name string, value interface{}) {
	r.vm.Set(name, value)
}

// GetValue 获取全局变量
func (r *Runner) GetValue(name string) goja.Value {
	return r.vm.Get(name)
}

// ClearModuleCache 清除模块缓存
func (r *Runner) ClearModuleCache() {
	r.modules.ClearCache()
}

// GetLoadedModules 获取已加载的模块列表
func (r *Runner) GetLoadedModules() []string {
	return r.modules.GetLoadedModules()
}

// GetBuiltinModules 获取内置模块列表
func (r *Runner) GetBuiltinModules() []string {
	return r.modules.GetBuiltinModules()
}

// Close 关闭运行器并清理资源
func (r *Runner) Close() {
	// 停止事件循环
	r.loop.Stop()

	// 关闭模块系统（包括所有 HTTP 服务器）
	r.modules.Close()

	// 减少 Runner 计数
	pool.GlobalMemoryMonitor.DecrementRunnerCount()
}

// GetMemoryStats 获取内存统计信息
func (r *Runner) GetMemoryStats() pool.MemoryStats {
	return pool.GlobalMemoryMonitor.GetStats()
}
