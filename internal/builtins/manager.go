package builtins

import (
	"strings"
	"time"

	"github.com/dop251/goja"
)

type BuiltinModule interface {
	GetModule() *goja.Object
}

type NamespaceModule interface {
	GetSubModule(name string) (BuiltinModule, bool)
}

type Manager struct {
	vm         *goja.Runtime
	modules    map[string]BuiltinModule
	namespaces map[string]NamespaceModule
	basePath   string
	startTime  time.Time
	argv       []string
}

func NewManager(vm *goja.Runtime, basePath string) *Manager {
	m := &Manager{
		vm:         vm,
		modules:    make(map[string]BuiltinModule),
		namespaces: make(map[string]NamespaceModule),
		basePath:   basePath,
		startTime:  time.Now(),
		argv:       make([]string, 0),
	}
	m.registerBuiltinModules()
	return m
}

func (m *Manager) registerBuiltinModules() {
	httpNS := NewHTTPNamespace(m.vm)
	m.namespaces["http"] = httpNS
	m.modules["http"] = httpNS

	m.modules["path"] = NewPathModule(m.vm)
	m.modules["fs"] = NewFSModule(m.vm, m.basePath)
	m.modules["crypto"] = NewCryptoModule(m.vm)
	m.modules["zlib"] = NewCompressionModule(m.vm)
	m.modules["compression"] = NewCompressionModule(m.vm)
	m.modules["redis"] = NewRedisModule(m.vm)
	m.modules["sqlite"] = NewSQLiteModule(m.vm)
	m.modules["exec"] = NewExecModule(m.vm)
	m.modules["child_process"] = NewExecModule(m.vm)
	m.modules["websocket"] = NewWebSocketModule(m.vm)
	m.modules["ws"] = NewWebSocketModule(m.vm)
	m.modules["net"] = NewNetModule(m.vm)
	m.modules["proxy"] = NewProxyModule(m.vm)
	m.modules["time"] = NewTimeModule(m.vm)

	m.modules["httpserver"] = NewHTTPServerModule(m.vm)
	m.modules["server"] = NewHTTPServerModule(m.vm)
}

func (m *Manager) GetModule(name string) (BuiltinModule, bool) {
	module, exists := m.modules[name]
	return module, exists
}

func (m *Manager) GetNamespaceModule(name string) (NamespaceModule, bool) {
	module, exists := m.namespaces[name]
	return module, exists
}

func (m *Manager) GetNamespacedModule(fullName string) (BuiltinModule, bool) {
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) != 2 {
		return nil, false
	}

	nsName, subName := parts[0], parts[1]
	ns, ok := m.namespaces[nsName]
	if !ok {
		return nil, false
	}

	return ns.GetSubModule(subName)
}

func (m *Manager) HasModule(name string) bool {
	if _, ok := m.modules[name]; ok {
		return true
	}
	if _, ok := m.namespaces[name]; ok {
		return true
	}
	if strings.Contains(name, "/") {
		_, ok := m.GetNamespacedModule(name)
		return ok
	}
	return false
}

func (m *Manager) GetModuleNames() []string {
	names := make([]string, 0, len(m.modules))
	for name := range m.modules {
		names = append(names, name)
	}
	return names
}

func (m *Manager) RegisterModule(name string, module BuiltinModule) {
	m.modules[name] = module
}

func (m *Manager) Close() {
	closeAllHTTPServers()
}

// SetArgv 设置命令行参数
func (m *Manager) SetArgv(argv []string) {
	m.argv = argv
}

// SetStartTime 设置起始时间
func (m *Manager) SetStartTime(t time.Time) {
	m.startTime = t
}
