package net

import (
	"sw_runtime/internal/builtins/types"

	"github.com/dop251/goja"
)

// Namespace net 命名空间
type Namespace struct {
	vm       *goja.Runtime
	net      *NetModule
	proxy    *ProxyModule
	websocket *WebSocketModule
}

// NewNamespace 创建 net 命名空间
func NewNamespace(vm *goja.Runtime) *Namespace {
	return &Namespace{
		vm:       vm,
		net:      NewNetModule(vm),
		proxy:    NewProxyModule(vm),
		websocket: NewWebSocketModule(vm),
	}
}

// GetModule 获取命名空间对象
func (n *Namespace) GetModule() *goja.Object {
	obj := n.vm.NewObject()

	// 添加子模块
	netObj := n.net.GetModule()
	obj.Set("net", netObj)

	proxyObj := n.proxy.GetModule()
	obj.Set("proxy", proxyObj)

	websocketObj := n.websocket.GetModule()
	obj.Set("websocket", websocketObj)

	return obj
}

// GetSubModule 获取子模块
func (n *Namespace) GetSubModule(name string) (types.BuiltinModule, bool) {
	switch name {
	case "net":
		return n.net, true
	case "proxy":
		return n.proxy, true
	case "websocket":
		return n.websocket, true
	}
	return nil, false
}
