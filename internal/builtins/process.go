package builtins

import (
	"os"
	"runtime"
	"time"

	"github.com/dop251/goja"
	"github.com/shirou/gopsutil/v3/process"
)

// ProcessModule 进程模块
type ProcessModule struct {
	vm      *goja.Runtime
	manager *Manager
}

// NewProcessModule 创建进程模块
func NewProcessModule(vm *goja.Runtime, manager *Manager) *ProcessModule {
	return &ProcessModule{
		vm:      vm,
		manager: manager,
	}
}

// GetModule 获取模块对象
func (p *ProcessModule) GetModule() *goja.Object {
	obj := p.vm.NewObject()

	// 基础属性
	obj.Set("pid", os.Getpid())
	obj.Set("platform", runtime.GOOS)
	obj.Set("arch", runtime.GOARCH)
	obj.Set("version", "v1.0.0") // Runtime 版本

	// versions 对象
	versions := p.vm.NewObject()
	versions.Set("node", "18.0.0") // 兼容性声明
	versions.Set("sw_runtime", "1.0.0")
	versions.Set("go", runtime.Version())
	obj.Set("versions", versions)

	// 方法
	obj.Set("cwd", p.cwd)
	obj.Set("chdir", p.chdir)
	obj.Set("exit", p.exit)
	obj.Set("uptime", p.uptime)
	obj.Set("memoryUsage", p.memoryUsage)
	obj.Set("nextTick", p.nextTick)
	obj.Set("hrtime", p.hrtime)
	obj.Set("kill", p.kill)

	// 标准流
	obj.Set("stdout", p.getStdout())
	obj.Set("stderr", p.getStderr())

	// 环境变量和参数 - 使用 Getter 确保在 SetArgv 之后也能正确获取
	obj.DefineAccessorProperty("env", p.vm.ToValue(p.envGetter), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)
	obj.DefineAccessorProperty("argv", p.vm.ToValue(p.argvGetter), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

	return obj
}

// envGetter 环境变量 Getter
func (p *ProcessModule) envGetter(call goja.FunctionCall) goja.Value {
	obj := p.vm.NewObject()
	for _, env := range os.Environ() {
		for i := 0; i < len(env); i++ {
			if env[i] == '=' {
				obj.Set(env[:i], env[i+1:])
				break
			}
		}
	}
	return obj
}

// argvGetter 命令行参数 Getter
func (p *ProcessModule) argvGetter(call goja.FunctionCall) goja.Value {
	if p.manager.argv == nil {
		return p.vm.NewArray()
	}
	return p.vm.ToValue(p.manager.argv)
}

// getStdout 获取 stdout 对象
func (p *ProcessModule) getStdout() *goja.Object {
	obj := p.vm.NewObject()
	obj.Set("write", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			os.Stdout.WriteString(call.Arguments[0].String())
		}
		return goja.Undefined()
	})
	return obj
}

// getStderr 获取 stderr 对象
func (p *ProcessModule) getStderr() *goja.Object {
	obj := p.vm.NewObject()
	obj.Set("write", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			os.Stderr.WriteString(call.Arguments[0].String())
		}
		return goja.Undefined()
	})
	return obj
}

// hrtime 高精度时间
func (p *ProcessModule) hrtime(call goja.FunctionCall) goja.Value {
	now := time.Now()

	if len(call.Arguments) > 0 {
		if arg, ok := call.Arguments[0].Export().([]interface{}); ok && len(arg) == 2 {
			sec := arg[0].(int64)
			nsec := arg[1].(int64)

			duration := now.Sub(time.Unix(sec, nsec))
			res := p.vm.NewArray()
			res.Set("0", int64(duration.Seconds()))
			res.Set("1", int64(duration.Nanoseconds())%1e9)
			return res
		}
	}

	res := p.vm.NewArray()
	res.Set("0", now.Unix())
	res.Set("1", int64(now.Nanosecond()))
	return res
}

// kill 发送信号到进程
func (p *ProcessModule) kill(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(p.vm.NewTypeError("kill requires pid"))
	}

	pid := int(call.Arguments[0].ToInteger())
	sig := os.Interrupt

	if len(call.Arguments) > 1 {
		// 这里简化处理，仅支持字符串形式的信号，如 'SIGINT'
		sigPrefix := call.Arguments[1].String()
		switch sigPrefix {
		case "SIGINT":
			sig = os.Interrupt
		case "SIGKILL":
			sig = os.Kill
		}
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return p.vm.ToValue(false)
	}

	err = proc.Signal(sig)
	return p.vm.ToValue(err == nil)
}

// cwd 获取当前工作目录
func (p *ProcessModule) cwd(call goja.FunctionCall) goja.Value {
	dir, err := os.Getwd()
	if err != nil {
		panic(p.vm.NewGoError(err))
	}
	return p.vm.ToValue(dir)
}

// chdir 改变工作目录
func (p *ProcessModule) chdir(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(p.vm.NewTypeError("chdir requires path"))
	}
	err := os.Chdir(call.Arguments[0].String())
	if err != nil {
		panic(p.vm.NewGoError(err))
	}
	return goja.Undefined()
}

// exit 退出进程
func (p *ProcessModule) exit(call goja.FunctionCall) goja.Value {
	code := 0
	if len(call.Arguments) > 0 {
		code = int(call.Arguments[0].ToInteger())
	}
	os.Exit(code)
	return goja.Undefined()
}

// uptime 进程运行时间（秒）
func (p *ProcessModule) uptime(call goja.FunctionCall) goja.Value {
	elapsed := time.Since(p.manager.startTime)
	return p.vm.ToValue(elapsed.Seconds())
}

// memoryUsage 内存使用情况
func (p *ProcessModule) memoryUsage(call goja.FunctionCall) goja.Value {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	result := p.vm.NewObject()
	result.Set("rss", m.Sys)
	result.Set("heapTotal", m.HeapSys)
	result.Set("heapUsed", m.HeapAlloc)
	result.Set("external", 0)
	result.Set("arrayBuffers", 0)

	// 如果能获取到更详细的进程信息
	if proc, err := process.NewProcess(int32(os.Getpid())); err == nil {
		if mem, err := proc.MemoryInfo(); err == nil {
			result.Set("rss", mem.RSS)
		}
	}

	return result
}

// nextTick 调度微任务
func (p *ProcessModule) nextTick(call goja.FunctionCall) goja.Value {
	if p.manager.nextTickFunc != nil {
		return p.manager.nextTickFunc(call)
	}
	return goja.Undefined()
}
