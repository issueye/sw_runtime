package utils

import (
	"sw_runtime/internal/builtins/types"

	"github.com/dop251/goja"
)

// Namespace utils 命名空间
type Namespace struct {
	vm          *goja.Runtime
	path        *PathModule
	time        *TimeModule
	util        *UtilModule
	compression *CompressionModule
	crypto      *CryptoModule
}

// NewNamespace 创建 utils 命名空间
func NewNamespace(vm *goja.Runtime) *Namespace {
	return &Namespace{
		vm:          vm,
		path:        NewPathModule(vm),
		time:        NewTimeModule(vm),
		util:        NewUtilModule(vm),
		compression: NewCompressionModule(vm),
		crypto:      NewCryptoModule(vm),
	}
}

// GetModule 获取命名空间对象
func (u *Namespace) GetModule() *goja.Object {
	obj := u.vm.NewObject()

	// 添加子模块
	pathObj := u.path.GetModule()
	obj.Set("path", pathObj)

	timeObj := u.time.GetModule()
	obj.Set("time", timeObj)

	utilObj := u.util.GetModule()
	obj.Set("util", utilObj)

	compressionObj := u.compression.GetModule()
	obj.Set("compression", compressionObj)

	cryptoObj := u.crypto.GetModule()
	obj.Set("crypto", cryptoObj)

	return obj
}

// GetSubModule 获取子模块
func (u *Namespace) GetSubModule(name string) (types.BuiltinModule, bool) {
	switch name {
	case "path":
		return u.path, true
	case "time":
		return u.time, true
	case "util":
		return u.util, true
	case "compression":
		return u.compression, true
	case "crypto":
		return u.crypto, true
	}
	return nil, false
}
