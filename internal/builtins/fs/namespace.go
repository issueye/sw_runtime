package fs

import (
	"sw_runtime/internal/builtins/types"

	"github.com/dop251/goja"
)

// Namespace fs 命名空间
type Namespace struct {
	vm  *goja.Runtime
	fs  *FSModule
	os  *OSModule
}

// NewNamespace 创建 fs 命名空间
func NewNamespace(vm *goja.Runtime, basePath string) *Namespace {
	return &Namespace{
		vm:  vm,
		fs:  NewFSModule(vm, basePath),
		os:  NewOSModule(vm),
	}
}

// GetModule 获取命名空间对象
func (f *Namespace) GetModule() *goja.Object {
	obj := f.vm.NewObject()

	// 添加子模块
	fsObj := f.fs.GetModule()
	obj.Set("fs", fsObj)

	osObj := f.os.GetModule()
	obj.Set("os", osObj)

	return obj
}

// GetSubModule 获取子模块
func (f *Namespace) GetSubModule(name string) (types.BuiltinModule, bool) {
	switch name {
	case "fs":
		return f.fs, true
	case "os":
		return f.os, true
	}
	return nil, false
}
