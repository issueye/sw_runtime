package config

import (
	"sw_runtime/internal/builtins/types"

	"github.com/dop251/goja"
)

// Namespace config 命名空间
type Namespace struct {
	vm    *goja.Runtime
	viper *ViperModule
}

// NewNamespace 创建 config 命名空间
func NewNamespace(vm *goja.Runtime) *Namespace {
	return &Namespace{
		vm:    vm,
		viper: NewViperModule(vm),
	}
}

// GetModule 获取命名空间对象
func (c *Namespace) GetModule() *goja.Object {
	obj := c.vm.NewObject()

	// 添加子模块
	viperObj := c.viper.GetModule()
	obj.Set("viper", viperObj)

	return obj
}

// GetSubModule 获取子模块
func (c *Namespace) GetSubModule(name string) (types.BuiltinModule, bool) {
	switch name {
	case "viper":
		return c.viper, true
	}
	return nil, false
}
