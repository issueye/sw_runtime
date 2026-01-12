package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"sw_runtime/internal/modules"
	"sw_runtime/internal/pool"

	"time"

	"github.com/dop251/goja"
)

// EventLoopType 事件循环类型
type EventLoopType int

const (
	// EventLoopSimple 简单事件循环（轮询模式）
	EventLoopSimple EventLoopType = iota
	// EventLoopOptimized 优化事件循环（事件驱动模式）
	EventLoopOptimized
)

// DefaultEventLoopType 默认事件循环类型
var DefaultEventLoopType = EventLoopOptimized

// eventLoopInterface 事件循环接口
type eventLoopInterface interface {
	Start()
	Stop()
	AddJob()
	DoneJob()
	SetLongLived()
	WaitAndProcess()
	SetTimeout(call goja.FunctionCall) goja.Value
	ClearTimeout(call goja.FunctionCall) goja.Value
	SetInterval(call goja.FunctionCall) goja.Value
	ClearInterval(call goja.FunctionCall) goja.Value
	NextTick(call goja.FunctionCall) goja.Value
	RunOnLoopSync(func(*goja.Runtime) interface{}) interface{}
}

// Runner JavaScript/TypeScript 运行器
type Runner struct {
	vm      *goja.Runtime
	loop    eventLoopInterface
	modules *modules.System
	argv    []string
	start   time.Time
}

// RunnerPool Runner 对象池，用于复用 Runner 实例以减少频繁创建开销。
// 注意：池中的 Runner 会复用同一个 JS VM，因此全局状态（global 上挂的变量等）不会自动重置。
// 仅在你能接受跨调用共享全局状态，或者每次手动清理全局变量的场景下使用。
type RunnerPool struct {
	pool sync.Pool
}

// defaultRunnerPool 默认全局 Runner 池。
var defaultRunnerPool = NewRunnerPool()

// NewRunnerPool 创建新的 Runner 池。
func NewRunnerPool() *RunnerPool {
	rp := &RunnerPool{}
	rp.pool.New = func() interface{} {
		// 对于池中暂无可用实例时，退回到正常的 New 创建逻辑。
		r, err := New()
		if err != nil {
			panic(err)
		}
		return r
	}
	return rp
}

// Acquire 从默认 Runner 池获取一个 Runner。
func (rp *RunnerPool) Acquire() *Runner {
	return rp.pool.Get().(*Runner)
}

// Release 将 Runner 放回默认 Runner 池以便复用。
// 当前实现仅清空模块缓存，不会重置 JS 全局状态，也不会主动关闭 HTTP 等长连接服务。
// 如需完全隔离环境，请继续使用 New/NewOrPanic + Close，而不要复用池。
func (rp *RunnerPool) Release(r *Runner) {
	if r == nil {
		return
	}

	// 清理模块缓存，避免上一次加载的文件模块残留。
	r.ClearModuleCache()

	rp.pool.Put(r)
}

// AcquireRunner 从默认池获取 Runner 的便捷函数。
func AcquireRunner() *Runner {
	return defaultRunnerPool.Acquire()
}

// ReleaseRunner 将 Runner 归还给默认池的便捷函数。
func ReleaseRunner(r *Runner) {
	defaultRunnerPool.Release(r)
}

// New 创建新的运行器（使用默认事件循环类型）
func New() (*Runner, error) {
	return NewWithEventLoop(DefaultEventLoopType)
}

// NewWithEventLoop 创建新的运行器，指定事件循环类型
func NewWithEventLoop(loopType EventLoopType) (*Runner, error) {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	// 根据类型创建事件循环
	var loop eventLoopInterface
	switch loopType {
	case EventLoopOptimized:
		loop = NewEventLoop(vm)
	default:
		loop = NewSimpleEventLoop(vm)
	}

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
	r.setupBuiltinsWithDir(basePath)

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

// NewWithWorkingDir 创建新的运行器，使用指定的工作目录
func NewWithWorkingDir(workingDir string) (*Runner, error) {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	// 根据类型创建事件循环
	var loop eventLoopInterface
	switch DefaultEventLoopType {
	case EventLoopOptimized:
		loop = NewEventLoop(vm)
	default:
		loop = NewSimpleEventLoop(vm)
	}

	// 使用指定的工作目录
	basePath := filepath.Clean(workingDir)

	moduleSystem := modules.NewSystem(vm, basePath)

	r := &Runner{
		vm:      vm,
		loop:    loop,
		modules: moduleSystem,
	}
	r.setupBuiltinsWithDir(basePath)

	// 增加 Runner 计数
	pool.GlobalMemoryMonitor.IncrementRunnerCount()

	return r, nil
}

// NewOrPanicWithWorkingDir 创建新的运行器，使用指定工作目录
func NewOrPanicWithWorkingDir(workingDir string) *Runner {
	r, err := NewWithWorkingDir(workingDir)
	if err != nil {
		panic(err)
	}
	return r
}

// setupBuiltinsWithDir 注册内置函数，使用指定的工作目录
func (r *Runner) setupBuiltinsWithDir(workingDir string) {
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
			module, err := r.modules.LoadModule(id, workingDir)
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
	r.vm.Set("__dirname", workingDir)
	r.vm.Set("__filename", "")

	// 设置 process 对象
	process := r.modules.GetBuiltinModule("process")
	if process != nil {
		r.vm.Set("process", process)
		// 设置 nextTick 函数到管理器，由 process 模块调用
		r.modules.SetNextTick(r.loop.NextTick)
		// 设置 RunOnLoopSync 函数，由 Raft 等模块使用
		r.modules.SetRunOnLoopSync(r.loop.RunOnLoopSync)
	}

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

// SafeRunCode 执行代码并捕获底层运行时 panic，
// 将其包装为 error 返回，避免测试或调用方进程直接崩溃。
func (r *Runner) SafeRunCode(code string) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("runtime panic: %v", v)
		}
	}()
	return r.RunCode(code)
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

// SafeRunFile 执行文件并捕获底层运行时 panic。
func (r *Runner) SafeRunFile(filename string) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("runtime panic: %v", v)
		}
	}()
	return r.RunFile(filename)
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

// SetArgv 设置命令行参数
func (r *Runner) SetArgv(argv []string) {
	r.argv = argv
	r.modules.SetArgv(argv)
}

// GetArgv 获取命令行参数
func (r *Runner) GetArgv() []string {
	return r.argv
}

// SetStartTime 设置起始时间
func (r *Runner) SetStartTime(t time.Time) {
	r.start = t
	r.modules.SetStartTime(t)
}

// GetStartTime 获取起始时间
func (r *Runner) GetStartTime() time.Time {
	return r.start
}
