package http

import (
	"sw_runtime/internal/builtins/types"

	"github.com/dop251/goja"
)

type Namespace struct {
	vm     *goja.Runtime
	client *HTTPModule
	server *HTTPServerModule
}

func NewNamespace(vm *goja.Runtime) *Namespace {
	return &Namespace{
		vm:     vm,
		client: NewHTTPModule(vm),
		server: NewHTTPServerModule(vm),
	}
}

func (h *Namespace) GetModule() *goja.Object {
	obj := h.vm.NewObject()

	clientObj := h.client.GetModule()
	obj.Set("client", clientObj)

	serverObj := h.server.GetModule()
	obj.Set("server", serverObj)

	return obj
}

func (h *Namespace) GetSubModule(name string) (types.BuiltinModule, bool) {
	switch name {
	case "client":
		return h.client, true
	case "server":
		return h.server, true
	}
	return nil, false
}
