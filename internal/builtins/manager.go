package builtins

import (
	"strings"
	"sw_runtime/internal/builtins/config"
	"sw_runtime/internal/builtins/db"
	"sw_runtime/internal/builtins/fs"
	"sw_runtime/internal/builtins/http"
	"sw_runtime/internal/builtins/net"
	"sw_runtime/internal/builtins/process"
	"sw_runtime/internal/builtins/types"
	"sw_runtime/internal/builtins/utils"
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
	// HTTP 命名空间
	httpNS := http.NewNamespace(m.vm)
	m.namespaces["http"] = httpNS
	m.modules["http"] = httpNS

	// DB 命名空间
	dbNS := db.NewNamespace(m.vm)
	m.namespaces["db"] = dbNS
	m.modules["db"] = dbNS

	// Utils 命名空间 (path, time, util, compression, crypto)
	utilsNS := utils.NewNamespace(m.vm)
	m.namespaces["utils"] = utilsNS

	// Net 命名空间 (net, proxy, websocket)
	netNS := net.NewNamespace(m.vm)
	m.namespaces["net"] = netNS

	// FS 命名空间 (fs, os)
	fsNS := fs.NewNamespace(m.vm, m.basePath)
	m.namespaces["fs"] = fsNS

	// Config 命名空间 (viper)
	configNS := config.NewNamespace(m.vm)
	m.namespaces["config"] = configNS

	// Process 命名空间 (process, exec)
	processNS := process.NewNamespace(m.vm, m.argv, m.startTime)
	m.namespaces["process"] = processNS
	m.modules["process"] = processNS
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
	net.CloseAllTCPServers()
}

// SetArgv 设置命令行参数
func (m *Manager) SetArgv(argv []string) {
	m.argv = argv
}

// SetStartTime 设置起始时间
func (m *Manager) SetStartTime(t time.Time) {
	m.startTime = t
}

// NewHTTPServerModule 创建 HTTP 服务器模块（向后兼容导出）
func NewHTTPServerModule(vm *goja.Runtime) *http.HTTPServerModule {
	return http.NewHTTPServerModule(vm)
}
