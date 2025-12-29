package builtins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dop251/goja"

	"sw_runtime/internal/consts"
	"sw_runtime/internal/security"
)

// FSModule 文件系统模块
type FSModule struct {
	vm        *goja.Runtime
	basePath  string
	validator *security.PathValidator
}

// NewFSModule 创建文件系统模块
func NewFSModule(vm *goja.Runtime, basePath string) *FSModule {
	// 使用传入的基础路径，如果为空则使用当前工作目录
	if basePath == "" {
		var err error
		basePath, err = os.Getwd()
		if err != nil {
			basePath = os.TempDir()
		}
	}

	return &FSModule{
		vm:        vm,
		basePath:  basePath,
		validator: security.NewPathValidator(basePath),
	}
}

// sanitizePath 验证并清理路径，防止路径遍历攻击
func (f *FSModule) sanitizePath(path string) (string, error) {
	// 检查路径长度
	if len(path) > consts.MaxPathLength {
		return "", fmt.Errorf("path too long: max %d characters", consts.MaxPathLength)
	}

	// 检查是否包含空字节
	if strings.ContainsAny(path, "\x00") {
		return "", fmt.Errorf("path contains null bytes")
	}

	// 清理路径
	cleanPath := filepath.Clean(path)

	// 检查是否包含路径遍历模式
	if strings.Contains(cleanPath, "..") {
		// 使用验证器检查
		return f.validator.Validate(cleanPath)
	}

	// 对于相对路径，使用验证器检查
	if !filepath.IsAbs(cleanPath) {
		return f.validator.Validate(cleanPath)
	}

	// 对于绝对路径，也需要验证
	return f.validator.Validate(cleanPath)
}

// validatePath 验证路径并返回绝对路径
func (f *FSModule) validatePath(path string) (string, error) {
	return f.sanitizePath(path)
}

// GetModule 获取文件系统模块对象
func (f *FSModule) GetModule() *goja.Object {
	obj := f.vm.NewObject()

	// 同步方法
	obj.Set("readFileSync", f.readFileSync)
	obj.Set("writeFileSync", f.writeFileSync)
	obj.Set("existsSync", f.existsSync)
	obj.Set("statSync", f.statSync)
	obj.Set("mkdirSync", f.mkdirSync)
	obj.Set("readdirSync", f.readdirSync)
	obj.Set("unlinkSync", f.unlinkSync)
	obj.Set("rmdirSync", f.rmdirSync)
	obj.Set("copyFileSync", f.copyFileSync)
	obj.Set("renameSync", f.renameSync)

	// 异步方法 (简化版，返回 Promise)
	obj.Set("readFile", f.readFile)
	obj.Set("writeFile", f.writeFile)
	obj.Set("exists", f.exists)
	obj.Set("stat", f.stat)
	obj.Set("mkdir", f.mkdir)
	obj.Set("readdir", f.readdir)
	obj.Set("unlink", f.unlink)
	obj.Set("rmdir", f.rmdir)
	obj.Set("copyFile", f.copyFile)
	obj.Set("rename", f.rename)

	return obj
}

// readFileSync 同步读取文件
func (f *FSModule) readFileSync(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(f.vm.NewTypeError("readFileSync requires a filename"))
	}

	filename := call.Arguments[0].String()
	safePath, err := f.validatePath(filename)
	if err != nil {
		panic(f.vm.NewGoError(fmt.Errorf("access denied: %w", err)))
	}

	content, err := os.ReadFile(safePath)
	if err != nil {
		panic(f.vm.NewGoError(err))
	}

	// 检查编码选项
	if len(call.Arguments) > 1 {
		options := call.Arguments[1]
		if options.ExportType().Kind().String() == "string" {
			encoding := options.String()
			if encoding == "utf8" || encoding == "utf-8" {
				return f.vm.ToValue(string(content))
			}
		} else if obj := options.ToObject(f.vm); obj != nil {
			if enc := obj.Get("encoding"); enc != nil && enc != goja.Undefined() {
				encoding := enc.String()
				if encoding == "utf8" || encoding == "utf-8" {
					return f.vm.ToValue(string(content))
				}
			}
		}
	}

	return f.vm.ToValue(string(content))
}

// writeFileSync 同步写入文件
func (f *FSModule) writeFileSync(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(f.vm.NewTypeError("writeFileSync requires filename and data"))
	}

	filename := call.Arguments[0].String()
	safePath, err := f.validatePath(filename)
	if err != nil {
		panic(f.vm.NewGoError(fmt.Errorf("access denied: %w", err)))
	}

	data := call.Arguments[1].String()

	err = os.WriteFile(safePath, []byte(data), consts.FilePermReadWrite)
	if err != nil {
		panic(f.vm.NewGoError(err))
	}

	return goja.Undefined()
}

// existsSync 检查文件是否存在
func (f *FSModule) existsSync(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		return f.vm.ToValue(false)
	}
	filename := call.Arguments[0].String()
	safePath, err := f.validatePath(filename)
	if err != nil {
		fmt.Println("validatePath Error:", err)
		return f.vm.ToValue(false)
	}
	_, err = os.Stat(safePath)
	if err != nil {
		fmt.Println("Error:", err)
		return f.vm.ToValue(false)
	}
	return f.vm.ToValue(err == nil)
}

// statSync 获取文件信息
func (f *FSModule) statSync(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(f.vm.NewTypeError("statSync requires a filename"))
	}

	filename := call.Arguments[0].String()
	safePath, err := f.validatePath(filename)
	if err != nil {
		panic(f.vm.NewGoError(fmt.Errorf("access denied: %w", err)))
	}

	info, err := os.Stat(safePath)
	if err != nil {
		panic(f.vm.NewGoError(err))
	}

	stat := f.vm.NewObject()
	stat.Set("isFile", func() bool { return info.Mode().IsRegular() })
	stat.Set("isDirectory", func() bool { return info.IsDir() })
	stat.Set("size", info.Size())
	stat.Set("mode", int(info.Mode()))
	stat.Set("mtime", info.ModTime())
	stat.Set("name", info.Name())

	return stat
}

// mkdirSync 创建目录
func (f *FSModule) mkdirSync(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(f.vm.NewTypeError("mkdirSync requires a path"))
	}

	path := call.Arguments[0].String()
	safePath, err := f.validatePath(path)
	if err != nil {
		panic(f.vm.NewGoError(fmt.Errorf("access denied: %w", err)))
	}

	// 检查选项
	recursive := false
	if len(call.Arguments) > 1 {
		if obj := call.Arguments[1].ToObject(f.vm); obj != nil {
			if rec := obj.Get("recursive"); rec != nil && rec != goja.Undefined() {
				recursive = rec.ToBoolean()
			}
		}
	}

	var errMkdir error
	if recursive {
		errMkdir = os.MkdirAll(safePath, consts.DirPermReadWrite)
	} else {
		errMkdir = os.Mkdir(safePath, consts.DirPermReadWrite)
	}

	if errMkdir != nil {
		panic(f.vm.NewGoError(errMkdir))
	}

	return goja.Undefined()
}

// readdirSync 读取目录
func (f *FSModule) readdirSync(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(f.vm.NewTypeError("readdirSync requires a path"))
	}

	path := call.Arguments[0].String()
	safePath, err := f.validatePath(path)
	if err != nil {
		panic(f.vm.NewGoError(fmt.Errorf("access denied: %w", err)))
	}

	entries, err := os.ReadDir(safePath)
	if err != nil {
		panic(f.vm.NewGoError(err))
	}

	result := f.vm.NewArray()
	for i, entry := range entries {
		result.Set(fmt.Sprintf("%d", i), f.vm.ToValue(entry.Name()))
	}

	return result
}

// unlinkSync 删除文件
func (f *FSModule) unlinkSync(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(f.vm.NewTypeError("unlinkSync requires a filename"))
	}

	filename := call.Arguments[0].String()
	safePath, err := f.validatePath(filename)
	if err != nil {
		panic(f.vm.NewGoError(fmt.Errorf("access denied: %w", err)))
	}

	err = os.Remove(safePath)
	if err != nil {
		panic(f.vm.NewGoError(err))
	}

	return goja.Undefined()
}

// rmdirSync 删除目录
func (f *FSModule) rmdirSync(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(f.vm.NewTypeError("rmdirSync requires a path"))
	}

	path := call.Arguments[0].String()
	safePath, err := f.validatePath(path)
	if err != nil {
		panic(f.vm.NewGoError(fmt.Errorf("access denied: %w", err)))
	}

	// 检查选项
	recursive := false
	if len(call.Arguments) > 1 {
		if obj := call.Arguments[1].ToObject(f.vm); obj != nil {
			if rec := obj.Get("recursive"); rec != nil && rec != goja.Undefined() {
				recursive = rec.ToBoolean()
			}
		}
	}

	var errRmdir error
	if recursive {
		errRmdir = os.RemoveAll(safePath)
	} else {
		errRmdir = os.Remove(safePath)
	}

	if errRmdir != nil {
		panic(f.vm.NewGoError(errRmdir))
	}

	return goja.Undefined()
}

// copyFileSync 复制文件
func (f *FSModule) copyFileSync(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(f.vm.NewTypeError("copyFileSync requires source and destination"))
	}

	src := call.Arguments[0].String()
	dst := call.Arguments[1].String()

	safeSrc, err := f.validatePath(src)
	if err != nil {
		panic(f.vm.NewGoError(fmt.Errorf("access denied (source): %w", err)))
	}

	safeDst, err := f.validatePath(dst)
	if err != nil {
		panic(f.vm.NewGoError(fmt.Errorf("access denied (destination): %w", err)))
	}

	data, err := os.ReadFile(safeSrc)
	if err != nil {
		panic(f.vm.NewGoError(err))
	}

	err = os.WriteFile(safeDst, data, consts.FilePermReadWrite)
	if err != nil {
		panic(f.vm.NewGoError(err))
	}

	return goja.Undefined()
}

// renameSync 重命名文件
func (f *FSModule) renameSync(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(f.vm.NewTypeError("renameSync requires old and new paths"))
	}

	oldPath := call.Arguments[0].String()
	newPath := call.Arguments[1].String()

	safeOld, err := f.validatePath(oldPath)
	if err != nil {
		panic(f.vm.NewGoError(fmt.Errorf("access denied (old path): %w", err)))
	}

	safeNew, err := f.validatePath(newPath)
	if err != nil {
		panic(f.vm.NewGoError(fmt.Errorf("access denied (new path): %w", err)))
	}

	err = os.Rename(safeOld, safeNew)
	if err != nil {
		panic(f.vm.NewGoError(err))
	}

	return goja.Undefined()
}

// 异步方法实现 (返回 Promise)

// readFile 异步读取文件
func (f *FSModule) readFile(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(f.vm.NewTypeError("readFile requires a filename"))
	}

	filename := call.Arguments[0].String()
	promise, resolve, reject := f.vm.NewPromise()

	go func() {
		safePath, err := f.validatePath(filename)
		if err != nil {
			reject(f.vm.NewGoError(fmt.Errorf("access denied: %w", err)))
			return
		}

		content, err := os.ReadFile(safePath)
		if err != nil {
			reject(f.vm.NewGoError(err))
		} else {
			resolve(f.vm.ToValue(string(content)))
		}
	}()

	return f.vm.ToValue(promise)
}

// writeFile 异步写入文件
func (f *FSModule) writeFile(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(f.vm.NewTypeError("writeFile requires filename and data"))
	}

	filename := call.Arguments[0].String()
	data := call.Arguments[1].String()
	promise, resolve, reject := f.vm.NewPromise()

	go func() {
		safePath, err := f.validatePath(filename)
		if err != nil {
			reject(f.vm.NewGoError(fmt.Errorf("access denied: %w", err)))
			return
		}

		err = os.WriteFile(safePath, []byte(data), consts.FilePermReadWrite)
		if err != nil {
			reject(f.vm.NewGoError(err))
		} else {
			resolve(goja.Undefined())
		}
	}()

	return f.vm.ToValue(promise)
}

// exists 异步检查文件是否存在
func (f *FSModule) exists(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(f.vm.NewTypeError("exists requires a filename"))
	}

	filename := call.Arguments[0].String()
	promise, resolve, _ := f.vm.NewPromise()

	go func() {
		safePath, err := f.validatePath(filename)
		if err != nil {
			resolve(f.vm.ToValue(false))
			return
		}

		_, err = os.Stat(safePath)
		resolve(f.vm.ToValue(err == nil))
	}()

	return f.vm.ToValue(promise)
}

// stat 异步获取文件信息
func (f *FSModule) stat(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(f.vm.NewTypeError("stat requires a filename"))
	}

	filename := call.Arguments[0].String()
	promise, resolve, reject := f.vm.NewPromise()

	go func() {
		safePath, err := f.validatePath(filename)
		if err != nil {
			reject(f.vm.NewGoError(fmt.Errorf("access denied: %w", err)))
			return
		}

		info, err := os.Stat(safePath)
		if err != nil {
			reject(f.vm.NewGoError(err))
		} else {
			stat := f.vm.NewObject()
			stat.Set("isFile", func() bool { return info.Mode().IsRegular() })
			stat.Set("isDirectory", func() bool { return info.IsDir() })
			stat.Set("size", info.Size())
			stat.Set("mode", int(info.Mode()))
			stat.Set("mtime", info.ModTime())
			stat.Set("name", info.Name())
			resolve(stat)
		}
	}()

	return f.vm.ToValue(promise)
}

// mkdir 异步创建目录
func (f *FSModule) mkdir(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(f.vm.NewTypeError("mkdir requires a path"))
	}

	path := call.Arguments[0].String()
	promise, resolve, reject := f.vm.NewPromise()

	go func() {
		safePath, err := f.validatePath(path)
		if err != nil {
			reject(f.vm.NewGoError(fmt.Errorf("access denied: %w", err)))
			return
		}

		err = os.MkdirAll(safePath, consts.DirPermReadWrite)
		if err != nil {
			reject(f.vm.NewGoError(err))
		} else {
			resolve(goja.Undefined())
		}
	}()

	return f.vm.ToValue(promise)
}

// readdir 异步读取目录
func (f *FSModule) readdir(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(f.vm.NewTypeError("readdir requires a path"))
	}

	path := call.Arguments[0].String()
	promise, resolve, reject := f.vm.NewPromise()

	go func() {
		safePath, err := f.validatePath(path)
		if err != nil {
			reject(f.vm.NewGoError(fmt.Errorf("access denied: %w", err)))
			return
		}

		entries, err := os.ReadDir(safePath)
		if err != nil {
			reject(f.vm.NewGoError(err))
		} else {
			result := f.vm.NewArray()
			for i, entry := range entries {
				result.Set(fmt.Sprintf("%d", i), f.vm.ToValue(entry.Name()))
			}
			resolve(result)
		}
	}()

	return f.vm.ToValue(promise)
}

// unlink 异步删除文件
func (f *FSModule) unlink(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(f.vm.NewTypeError("unlink requires a filename"))
	}

	filename := call.Arguments[0].String()
	promise, resolve, reject := f.vm.NewPromise()

	go func() {
		safePath, err := f.validatePath(filename)
		if err != nil {
			reject(f.vm.NewGoError(fmt.Errorf("access denied: %w", err)))
			return
		}

		err = os.Remove(safePath)
		if err != nil {
			reject(f.vm.NewGoError(err))
		} else {
			resolve(goja.Undefined())
		}
	}()

	return f.vm.ToValue(promise)
}

// rmdir 异步删除目录
func (f *FSModule) rmdir(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(f.vm.NewTypeError("rmdir requires a path"))
	}

	path := call.Arguments[0].String()
	promise, resolve, reject := f.vm.NewPromise()

	go func() {
		safePath, err := f.validatePath(path)
		if err != nil {
			reject(f.vm.NewGoError(fmt.Errorf("access denied: %w", err)))
			return
		}

		err = os.RemoveAll(safePath)
		if err != nil {
			reject(f.vm.NewGoError(err))
		} else {
			resolve(goja.Undefined())
		}
	}()

	return f.vm.ToValue(promise)
}

// copyFile 异步复制文件
func (f *FSModule) copyFile(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(f.vm.NewTypeError("copyFile requires source and destination"))
	}

	src := call.Arguments[0].String()
	dst := call.Arguments[1].String()
	promise, resolve, reject := f.vm.NewPromise()

	go func() {
		safeSrc, err := f.validatePath(src)
		if err != nil {
			reject(f.vm.NewGoError(fmt.Errorf("access denied (source): %w", err)))
			return
		}

		safeDst, err := f.validatePath(dst)
		if err != nil {
			reject(f.vm.NewGoError(fmt.Errorf("access denied (destination): %w", err)))
			return
		}

		data, err := os.ReadFile(safeSrc)
		if err != nil {
			reject(f.vm.NewGoError(err))
			return
		}

		err = os.WriteFile(safeDst, data, consts.FilePermReadWrite)
		if err != nil {
			reject(f.vm.NewGoError(err))
		} else {
			resolve(goja.Undefined())
		}
	}()

	return f.vm.ToValue(promise)
}

// rename 异步重命名文件
func (f *FSModule) rename(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(f.vm.NewTypeError("rename requires old and new paths"))
	}

	oldPath := call.Arguments[0].String()
	newPath := call.Arguments[1].String()
	promise, resolve, reject := f.vm.NewPromise()

	go func() {
		safeOld, err := f.validatePath(oldPath)
		if err != nil {
			reject(f.vm.NewGoError(fmt.Errorf("access denied (old path): %w", err)))
			return
		}

		safeNew, err := f.validatePath(newPath)
		if err != nil {
			reject(f.vm.NewGoError(fmt.Errorf("access denied (new path): %w", err)))
			return
		}

		err = os.Rename(safeOld, safeNew)
		if err != nil {
			reject(f.vm.NewGoError(err))
		} else {
			resolve(goja.Undefined())
		}
	}()

	return f.vm.ToValue(promise)
}
