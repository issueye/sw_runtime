package process

import (
	"sw_runtime/internal/builtins/types"
	"time"

	"github.com/dop251/goja"
)

// Namespace net 命名空间
type Namespace struct {
	vm      *goja.Runtime
	exec    *ExecModule
	process *ProcessModule
}

// NewNamespace 创建 net 命名空间
func NewNamespace(vm *goja.Runtime, args []string, startTime time.Time) *Namespace {
	return &Namespace{
		vm:      vm,
		exec:    NewExecModule(vm),
		process: NewProcessModule(vm, args, startTime),
	}
}

// GetModule 获取命名空间对象
func (n *Namespace) GetModule() *goja.Object {
	obj := n.vm.NewObject()

	execObj := n.exec.GetModule()
	obj.Set("exec", execObj)

	processObj := n.process.GetModule()
	obj.Set("process", processObj)

	return obj
}

// GetSubModule 获取子模块
func (n *Namespace) GetSubModule(name string) (types.BuiltinModule, bool) {
	switch name {
	case "exec":
		return n.exec, true
	case "process":
		return n.process, true
	}
	return nil, false
}
