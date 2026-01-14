package process

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/dop251/goja"
)

// ExecModule 命令执行模块
type ExecModule struct {
	vm *goja.Runtime
}

// NewExecModule 创建命令执行模块
func NewExecModule(vm *goja.Runtime) *ExecModule {
	return &ExecModule{vm: vm}
}

// GetModule 获取模块对象
func (e *ExecModule) GetModule() *goja.Object {
	obj := e.vm.NewObject()

	// 同步执行命令
	obj.Set("exec", e.execSync)
	obj.Set("execSync", e.execSync)

	// 异步执行命令
	obj.Set("execAsync", e.execAsync)

	// 执行命令并返回输出
	obj.Set("run", e.run)

	// 执行 shell 命令
	obj.Set("shell", e.shell)

	// 执行命令（带超时）
	obj.Set("execWithTimeout", e.execWithTimeout)

	// 获取环境变量
	obj.Set("env", e.getEnv)
	obj.Set("getEnv", e.getEnv)
	obj.Set("setEnv", e.setEnv)

	// 获取当前工作目录
	obj.Set("cwd", e.getCwd)
	obj.Set("chdir", e.chdir)

	// 获取系统信息
	obj.Set("platform", runtime.GOOS)
	obj.Set("arch", runtime.GOARCH)

	// 命令是否存在
	obj.Set("which", e.which)
	obj.Set("commandExists", e.commandExists)

	return obj
}

// execSync 同步执行命令
func (e *ExecModule) execSync(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(e.vm.NewTypeError("exec requires command"))
	}

	command := call.Arguments[0].String()
	args := []string{}

	// 解析参数
	if len(call.Arguments) > 1 {
		if argsVal := call.Arguments[1]; argsVal != nil && !goja.IsUndefined(argsVal) && !goja.IsNull(argsVal) {
			if argsExport, ok := argsVal.Export().([]interface{}); ok {
				for _, arg := range argsExport {
					args = append(args, fmt.Sprintf("%v", arg))
				}
			}
		}
	}

	// 解析选项
	options := e.parseOptions(call, 2)

	// 创建命令
	cmd := exec.Command(command, args...)

	// 设置工作目录
	if options.cwd != "" {
		cmd.Dir = options.cwd
	}

	// 设置环境变量
	if len(options.env) > 0 {
		cmd.Env = append(os.Environ(), options.env...)
	}

	// 执行命令
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// 创建结果对象
	result := e.vm.NewObject()
	result.Set("stdout", stdout.String())
	result.Set("stderr", stderr.String())
	result.Set("command", command)
	result.Set("args", args)

	if err != nil {
		result.Set("error", err.Error())
		result.Set("success", false)

		// 获取退出码
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				result.Set("exitCode", status.ExitStatus())
			}
		} else {
			result.Set("exitCode", -1)
		}
	} else {
		result.Set("error", nil)
		result.Set("success", true)
		result.Set("exitCode", 0)
	}

	return result
}

// execAsync 异步执行命令
func (e *ExecModule) execAsync(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(e.vm.NewTypeError("execAsync requires command"))
	}

	command := call.Arguments[0].String()
	args := []string{}

	// 解析参数
	if len(call.Arguments) > 1 {
		if argsVal := call.Arguments[1]; argsVal != nil && !goja.IsUndefined(argsVal) && !goja.IsNull(argsVal) {
			if argsExport, ok := argsVal.Export().([]interface{}); ok {
				for _, arg := range argsExport {
					args = append(args, fmt.Sprintf("%v", arg))
				}
			}
		}
	}

	// 解析选项
	options := e.parseOptions(call, 2)

	// 创建 Promise
	promise, resolve, _ := e.vm.NewPromise()

	go func() {
		// 创建命令
		cmd := exec.Command(command, args...)

		// 设置工作目录
		if options.cwd != "" {
			cmd.Dir = options.cwd
		}

		// 设置环境变量
		if len(options.env) > 0 {
			cmd.Env = append(os.Environ(), options.env...)
		}

		// 执行命令
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()

		// 创建结果对象
		result := e.vm.NewObject()
		result.Set("stdout", stdout.String())
		result.Set("stderr", stderr.String())
		result.Set("command", command)
		result.Set("args", args)

		if err != nil {
			result.Set("error", err.Error())
			result.Set("success", false)

			if exitError, ok := err.(*exec.ExitError); ok {
				if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
					result.Set("exitCode", status.ExitStatus())
				}
			} else {
				result.Set("exitCode", -1)
			}

			// 如果命令执行失败，仍然 resolve 但带有错误信息
			resolve(result)
		} else {
			result.Set("error", nil)
			result.Set("success", true)
			result.Set("exitCode", 0)
			resolve(result)
		}
	}()

	return e.vm.ToValue(promise)
}

// run 执行命令并直接返回输出
func (e *ExecModule) run(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(e.vm.NewTypeError("run requires command"))
	}

	command := call.Arguments[0].String()

	// 根据操作系统选择 shell
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	// 解析选项
	if len(call.Arguments) > 1 {
		options := e.parseOptions(call, 1)
		if options.cwd != "" {
			cmd.Dir = options.cwd
		}
		if len(options.env) > 0 {
			cmd.Env = append(os.Environ(), options.env...)
		}
	}

	output, err := cmd.CombinedOutput()

	result := e.vm.NewObject()
	result.Set("output", string(output))

	if err != nil {
		result.Set("error", err.Error())
		result.Set("success", false)
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				result.Set("exitCode", status.ExitStatus())
			}
		}
	} else {
		result.Set("error", nil)
		result.Set("success", true)
		result.Set("exitCode", 0)
	}

	return result
}

// shell 执行 shell 命令
func (e *ExecModule) shell(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(e.vm.NewTypeError("shell requires command"))
	}

	command := call.Arguments[0].String()

	// 根据操作系统选择 shell
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	// 解析选项
	if len(call.Arguments) > 1 {
		options := e.parseOptions(call, 1)
		if options.cwd != "" {
			cmd.Dir = options.cwd
		}
		if len(options.env) > 0 {
			cmd.Env = append(os.Environ(), options.env...)
		}
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := e.vm.NewObject()
	result.Set("stdout", stdout.String())
	result.Set("stderr", stderr.String())
	result.Set("command", command)

	if err != nil {
		result.Set("error", err.Error())
		result.Set("success", false)
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				result.Set("exitCode", status.ExitStatus())
			}
		}
	} else {
		result.Set("error", nil)
		result.Set("success", true)
		result.Set("exitCode", 0)
	}

	return result
}

// execWithTimeout 带超时的命令执行
func (e *ExecModule) execWithTimeout(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(e.vm.NewTypeError("execWithTimeout requires command and timeout"))
	}

	command := call.Arguments[0].String()
	timeout := time.Duration(call.Arguments[1].ToInteger()) * time.Millisecond

	args := []string{}
	if len(call.Arguments) > 2 {
		if argsVal := call.Arguments[2]; argsVal != nil && !goja.IsUndefined(argsVal) && !goja.IsNull(argsVal) {
			if argsExport, ok := argsVal.Export().([]interface{}); ok {
				for _, arg := range argsExport {
					args = append(args, fmt.Sprintf("%v", arg))
				}
			}
		}
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := e.vm.NewObject()
	result.Set("stdout", stdout.String())
	result.Set("stderr", stderr.String())
	result.Set("command", command)
	result.Set("args", args)
	result.Set("timeout", timeout.Milliseconds())

	if ctx.Err() == context.DeadlineExceeded {
		result.Set("error", "command timed out")
		result.Set("success", false)
		result.Set("timedOut", true)
		result.Set("exitCode", -1)
	} else if err != nil {
		result.Set("error", err.Error())
		result.Set("success", false)
		result.Set("timedOut", false)
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				result.Set("exitCode", status.ExitStatus())
			}
		}
	} else {
		result.Set("error", nil)
		result.Set("success", true)
		result.Set("timedOut", false)
		result.Set("exitCode", 0)
	}

	return result
}

// getEnv 获取环境变量
func (e *ExecModule) getEnv(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		// 返回所有环境变量
		envMap := e.vm.NewObject()
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				envMap.Set(parts[0], parts[1])
			}
		}
		return envMap
	}

	key := call.Arguments[0].String()
	value := os.Getenv(key)

	if value == "" {
		// 检查是否有默认值
		if len(call.Arguments) > 1 {
			return call.Arguments[1]
		}
		return goja.Undefined()
	}

	return e.vm.ToValue(value)
}

// setEnv 设置环境变量
func (e *ExecModule) setEnv(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(e.vm.NewTypeError("setEnv requires key and value"))
	}

	key := call.Arguments[0].String()
	value := call.Arguments[1].String()

	err := os.Setenv(key, value)
	if err != nil {
		return e.vm.ToValue(false)
	}

	return e.vm.ToValue(true)
}

// getCwd 获取当前工作目录
func (e *ExecModule) getCwd(call goja.FunctionCall) goja.Value {
	cwd, err := os.Getwd()
	if err != nil {
		panic(e.vm.NewGoError(err))
	}
	return e.vm.ToValue(cwd)
}

// chdir 改变当前工作目录
func (e *ExecModule) chdir(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(e.vm.NewTypeError("chdir requires path"))
	}

	path := call.Arguments[0].String()
	err := os.Chdir(path)
	if err != nil {
		return e.vm.ToValue(false)
	}

	return e.vm.ToValue(true)
}

// which 查找命令路径
func (e *ExecModule) which(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(e.vm.NewTypeError("which requires command name"))
	}

	command := call.Arguments[0].String()
	path, err := exec.LookPath(command)
	if err != nil {
		return goja.Null()
	}

	return e.vm.ToValue(path)
}

// commandExists 检查命令是否存在
func (e *ExecModule) commandExists(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(e.vm.NewTypeError("commandExists requires command name"))
	}

	command := call.Arguments[0].String()
	_, err := exec.LookPath(command)

	return e.vm.ToValue(err == nil)
}

// execOptions 命令执行选项
type execOptions struct {
	cwd string
	env []string
}

// parseOptions 解析选项
func (e *ExecModule) parseOptions(call goja.FunctionCall, index int) execOptions {
	options := execOptions{}

	if len(call.Arguments) <= index {
		return options
	}

	optVal := call.Arguments[index]
	if optVal == nil || goja.IsUndefined(optVal) || goja.IsNull(optVal) {
		return options
	}

	if optObj, ok := optVal.Export().(map[string]interface{}); ok {
		if cwd, ok := optObj["cwd"].(string); ok {
			options.cwd = cwd
		}

		if env, ok := optObj["env"].(map[string]interface{}); ok {
			for k, v := range env {
				options.env = append(options.env, fmt.Sprintf("%s=%v", k, v))
			}
		}
	}

	return options
}
