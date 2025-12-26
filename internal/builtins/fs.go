package builtins

import (
	"fmt"
	"os"

	"github.com/dop251/goja"
)

// FSModule 文件系统模块
type FSModule struct {
	vm *goja.Runtime
}

// NewFSModule 创建文件系统模块
func NewFSModule(vm *goja.Runtime) *FSModule {
	return &FSModule{vm: vm}
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
	content, err := os.ReadFile(filename)
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
	data := call.Arguments[1].String()

	err := os.WriteFile(filename, []byte(data), 0644)
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
	_, err := os.Stat(filename)
	return f.vm.ToValue(err == nil)
}

// statSync 获取文件信息
func (f *FSModule) statSync(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(f.vm.NewTypeError("statSync requires a filename"))
	}

	filename := call.Arguments[0].String()
	info, err := os.Stat(filename)
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

	// 检查选项
	recursive := false
	if len(call.Arguments) > 1 {
		if obj := call.Arguments[1].ToObject(f.vm); obj != nil {
			if rec := obj.Get("recursive"); rec != nil && rec != goja.Undefined() {
				recursive = rec.ToBoolean()
			}
		}
	}

	var err error
	if recursive {
		err = os.MkdirAll(path, 0755)
	} else {
		err = os.Mkdir(path, 0755)
	}

	if err != nil {
		panic(f.vm.NewGoError(err))
	}

	return goja.Undefined()
}

// readdirSync 读取目录
func (f *FSModule) readdirSync(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(f.vm.NewTypeError("readdirSync requires a path"))
	}

	path := call.Arguments[0].String()
	entries, err := os.ReadDir(path)
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
	err := os.Remove(filename)
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

	// 检查选项
	recursive := false
	if len(call.Arguments) > 1 {
		if obj := call.Arguments[1].ToObject(f.vm); obj != nil {
			if rec := obj.Get("recursive"); rec != nil && rec != goja.Undefined() {
				recursive = rec.ToBoolean()
			}
		}
	}

	var err error
	if recursive {
		err = os.RemoveAll(path)
	} else {
		err = os.Remove(path)
	}

	if err != nil {
		panic(f.vm.NewGoError(err))
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

	data, err := os.ReadFile(src)
	if err != nil {
		panic(f.vm.NewGoError(err))
	}

	err = os.WriteFile(dst, data, 0644)
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

	err := os.Rename(oldPath, newPath)
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
		content, err := os.ReadFile(filename)
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
		err := os.WriteFile(filename, []byte(data), 0644)
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
		_, err := os.Stat(filename)
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
		info, err := os.Stat(filename)
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
		err := os.MkdirAll(path, 0755)
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
		entries, err := os.ReadDir(path)
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
		err := os.Remove(filename)
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
		err := os.RemoveAll(path)
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
		data, err := os.ReadFile(src)
		if err != nil {
			reject(f.vm.NewGoError(err))
			return
		}

		err = os.WriteFile(dst, data, 0644)
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
		err := os.Rename(oldPath, newPath)
		if err != nil {
			reject(f.vm.NewGoError(err))
		} else {
			resolve(goja.Undefined())
		}
	}()

	return f.vm.ToValue(promise)
}
