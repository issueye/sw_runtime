package builtins

import (
	"github.com/dop251/goja"
)

type HTTPNamespace struct {
	vm     *goja.Runtime
	client *HTTPModule
	server *HTTPServerModule
}

func NewHTTPNamespace(vm *goja.Runtime) *HTTPNamespace {
	return &HTTPNamespace{
		vm:     vm,
		client: NewHTTPModule(vm),
		server: NewHTTPServerModule(vm),
	}
}

func (h *HTTPNamespace) GetModule() *goja.Object {
	obj := h.vm.NewObject()

	clientObj := h.client.GetModule()
	obj.Set("client", clientObj)

	serverObj := h.server.GetModule()
	obj.Set("server", serverObj)

	return obj
}

func (h *HTTPNamespace) GetSubModule(name string) (BuiltinModule, bool) {
	switch name {
	case "client":
		return h.client, true
	case "server":
		return h.server, true
	}
	return nil, false
}
