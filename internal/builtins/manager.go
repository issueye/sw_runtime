package builtins

import (
	"github.com/dop251/goja"
)

// BuiltinModule 内置模块接口
type BuiltinModule interface {
	GetModule() *goja.Object
}

// Manager 内置模块管理器
type Manager struct {
	vm      *goja.Runtime
	modules map[string]BuiltinModule
}

// NewManager 创建内置模块管理器
func NewManager(vm *goja.Runtime) *Manager {
	m := &Manager{
		vm:      vm,
		modules: make(map[string]BuiltinModule),
	}

	// 注册内置模块
	m.registerBuiltinModules()

	return m
}

// registerBuiltinModules 注册所有内置模块
func (m *Manager) registerBuiltinModules() {
	m.modules["path"] = NewPathModule(m.vm)
	m.modules["fs"] = NewFSModule(m.vm)
	m.modules["crypto"] = NewCryptoModule(m.vm)
	m.modules["zlib"] = NewCompressionModule(m.vm)
	m.modules["compression"] = NewCompressionModule(m.vm)
	m.modules["http"] = NewHTTPModule(m.vm)
	m.modules["httpserver"] = NewHTTPServerModule(m.vm)
	m.modules["server"] = NewHTTPServerModule(m.vm) // 别名
	m.modules["redis"] = NewRedisModule(m.vm)
	m.modules["sqlite"] = NewSQLiteModule(m.vm)
	m.modules["exec"] = NewExecModule(m.vm)
	m.modules["child_process"] = NewExecModule(m.vm) // Node.js 风格别名
	m.modules["websocket"] = NewWebSocketModule(m.vm)
	m.modules["ws"] = NewWebSocketModule(m.vm) // 简短别名
	m.modules["net"] = NewNetModule(m.vm)
	m.modules["proxy"] = NewProxyModule(m.vm)
	m.modules["time"] = NewTimeModule(m.vm)
}

// GetModule 获取内置模块
func (m *Manager) GetModule(name string) (BuiltinModule, bool) {
	module, exists := m.modules[name]
	return module, exists
}

// HasModule 检查模块是否存在
func (m *Manager) HasModule(name string) bool {
	_, exists := m.modules[name]
	return exists
}

// GetModuleNames 获取所有模块名称
func (m *Manager) GetModuleNames() []string {
	names := make([]string, 0, len(m.modules))
	for name := range m.modules {
		names = append(names, name)
	}
	return names
}

// RegisterModule 注册自定义模块
func (m *Manager) RegisterModule(name string, module BuiltinModule) {
	m.modules[name] = module
}

// Close 关闭所有内置模块并清理资源
func (m *Manager) Close() {
	closeAllHTTPServers()
}
