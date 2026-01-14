package builtins

import (
	"strings"
	"sw_runtime/internal/builtins/db"
	"sw_runtime/internal/builtins/http"
	"sw_runtime/internal/builtins/types"
	"time"

	"github.com/dop251/goja"
)

type Manager struct {
	vm         *goja.Runtime
	modules    map[string]types.BuiltinModule
	namespaces map[string]types.NamespaceModule
	basePath   string
	startTime  time.Time
	argv       []string
}

func NewManager(vm *goja.Runtime, basePath string) *Manager {
	m := &Manager{
		vm:         vm,
		modules:    make(map[string]types.BuiltinModule),
		namespaces: make(map[string]types.NamespaceModule),
		basePath:   basePath,
		startTime:  time.Now(),
		argv:       make([]string, 0),
	}
	m.registerBuiltinModules()
	return m
}

func (m *Manager) registerBuiltinModules() {
	httpNS := http.NewNamespace(m.vm)
	m.namespaces["http"] = httpNS
	m.modules["http"] = httpNS

	db := db.NewNamespace(m.vm)
	m.namespaces["db"] = db
	m.modules["db"] = db

	m.modules["path"] = NewPathModule(m.vm)
	m.modules["fs"] = NewFSModule(m.vm, m.basePath)
	m.modules["crypto"] = NewCryptoModule(m.vm)
	m.modules["zlib"] = NewCompressionModule(m.vm)
	m.modules["compression"] = NewCompressionModule(m.vm)
	m.modules["exec"] = NewExecModule(m.vm)
	m.modules["child_process"] = NewExecModule(m.vm)
	m.modules["websocket"] = NewWebSocketModule(m.vm)
	m.modules["ws"] = NewWebSocketModule(m.vm)
	m.modules["net"] = NewNetModule(m.vm)
	m.modules["proxy"] = NewProxyModule(m.vm)
	m.modules["time"] = NewTimeModule(m.vm)
	m.modules["viper"] = NewViperModule(m.vm)
}

func (m *Manager) GetModule(name string) (types.BuiltinModule, bool) {
	module, exists := m.modules[name]
	return module, exists
}

func (m *Manager) GetNamespaceModule(name string) (types.NamespaceModule, bool) {
	module, exists := m.namespaces[name]
	return module, exists
}

func (m *Manager) GetNamespacedModule(fullName string) (types.BuiltinModule, bool) {
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

func (m *Manager) RegisterModule(name string, module types.BuiltinModule) {
	m.modules[name] = module
}

func (m *Manager) Close() {
	http.CloseAllHTTPServers()
	closeAllTCPServers()
}

// SetArgv 设置命令行参数
func (m *Manager) SetArgv(argv []string) {
	m.argv = argv
}

// SetStartTime 设置起始时间
func (m *Manager) SetStartTime(t time.Time) {
	m.startTime = t
}
