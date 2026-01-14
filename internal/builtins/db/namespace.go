package db

import (
	"sw_runtime/internal/builtins/types"

	"github.com/dop251/goja"
)

type Namespace struct {
	vm     *goja.Runtime
	redis  *RedisModule
	sqlite *SQLiteModule
}

func NewNamespace(vm *goja.Runtime) *Namespace {
	return &Namespace{
		vm:     vm,
		redis:  NewRedisModule(vm),
		sqlite: NewSQLiteModule(vm),
	}
}

func (h *Namespace) GetModule() *goja.Object {
	obj := h.vm.NewObject()

	clientObj := h.redis.GetModule()
	obj.Set("redis", clientObj)

	serverObj := h.sqlite.GetModule()
	obj.Set("sqlite", serverObj)

	return obj
}

func (h *Namespace) GetSubModule(name string) (types.BuiltinModule, bool) {
	switch name {
	case "redis":
		return h.redis, true
	case "sqlite":
		return h.sqlite, true
	}
	return nil, false
}
