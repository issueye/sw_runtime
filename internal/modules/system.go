package modules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sw_runtime/internal/builtins"
	"sync"

	"github.com/dop251/goja"
)

// System 模块系统
type System struct {
	vm             *goja.Runtime
	cache          map[string]*Module
	builtinManager *builtins.Manager
	mu             sync.RWMutex
	basePath       string
	nodeModules    []string
}

// Module 表示一个模块
type Module struct {
	ID       string
	Filename string
	Exports  *goja.Object
	Loaded   bool
	Children []string
	Parent   string
}

// NewSystem 创建新的模块系统
func NewSystem(vm *goja.Runtime, basePath string) *System {
	ms := &System{
		vm:             vm,
		cache:          make(map[string]*Module),
		builtinManager: builtins.NewManager(vm, basePath),
		basePath:       basePath,
		nodeModules: []string{
			filepath.Join(basePath, "node_modules"),
		},
	}

	return ms
}

// resolveModule 解析模块路径
func (ms *System) resolveModule(id string, parentPath string) (string, error) {
	// 内置模块
	if ms.builtinManager.HasModule(id) {
		return id, nil
	}

	// 相对路径
	if strings.HasPrefix(id, "./") || strings.HasPrefix(id, "../") {
		var basePath string
		if parentPath == "" {
			basePath = ms.basePath
		} else {
			// 如果 parentPath 是文件，取其目录；如果是目录，直接使用
			if filepath.Ext(parentPath) != "" {
				basePath = filepath.Dir(parentPath)
			} else {
				basePath = parentPath
			}
		}
		resolved := filepath.Join(basePath, id)

		// 尝试不同的扩展名
		extensions := []string{"", ".js", ".ts", ".json"}
		for _, ext := range extensions {
			fullPath := resolved + ext
			if _, err := os.Stat(fullPath); err == nil {
				abs, _ := filepath.Abs(fullPath)
				return abs, nil
			}
		}

		// 尝试 index 文件
		indexExtensions := []string{"/index.js", "/index.ts", "/index.json"}
		for _, ext := range indexExtensions {
			fullPath := resolved + ext
			if _, err := os.Stat(fullPath); err == nil {
				abs, _ := filepath.Abs(fullPath)
				return abs, nil
			}
		}

		return "", fmt.Errorf("module not found: %s", id)
	}

	// 绝对路径
	if filepath.IsAbs(id) {
		if _, err := os.Stat(id); err == nil {
			return id, nil
		}
		return "", fmt.Errorf("module not found: %s", id)
	}

	// node_modules 查找
	for _, nodeModulesPath := range ms.nodeModules {
		modulePath := filepath.Join(nodeModulesPath, id)

		// 尝试不同的扩展名
		extensions := []string{"", ".js", ".ts", ".json"}
		for _, ext := range extensions {
			fullPath := modulePath + ext
			if _, err := os.Stat(fullPath); err == nil {
				abs, _ := filepath.Abs(fullPath)
				return abs, nil
			}
		}

		// 尝试 package.json 中的 main 字段
		packageJsonPath := filepath.Join(modulePath, "package.json")
		if _, err := os.Stat(packageJsonPath); err == nil {
			// 这里简化处理，直接尝试 index.js
			indexPath := filepath.Join(modulePath, "index.js")
			if _, err := os.Stat(indexPath); err == nil {
				abs, _ := filepath.Abs(indexPath)
				return abs, nil
			}
		}
	}

	return "", fmt.Errorf("module not found: %s", id)
}

// LoadModule 加载模块
func (ms *System) LoadModule(id string, parentPath string) (*Module, error) {
	resolvedPath, err := ms.resolveModule(id, parentPath)
	if err != nil {
		return nil, err
	}

	ms.mu.RLock()
	if cached, exists := ms.cache[resolvedPath]; exists {
		ms.mu.RUnlock()
		return cached, nil
	}
	ms.mu.RUnlock()

	// 内置模块
	if builtinModule, exists := ms.builtinManager.GetModule(resolvedPath); exists {
		module := &Module{
			ID:       resolvedPath,
			Filename: resolvedPath,
			Exports:  builtinModule.GetModule(),
			Loaded:   true,
		}

		ms.mu.Lock()
		ms.cache[resolvedPath] = module
		ms.mu.Unlock()

		return module, nil
	}

	// 文件模块
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read module %s: %w", resolvedPath, err)
	}

	module := &Module{
		ID:       resolvedPath,
		Filename: resolvedPath,
		Exports:  ms.vm.NewObject(),
		Loaded:   false,
		Parent:   parentPath,
	}

	ms.mu.Lock()
	ms.cache[resolvedPath] = module
	ms.mu.Unlock()

	// 执行模块代码
	err = ms.executeModule(string(content), module)
	if err != nil {
		ms.mu.Lock()
		delete(ms.cache, resolvedPath)
		ms.mu.Unlock()
		return nil, err
	}

	module.Loaded = true
	return module, nil
}

// executeModule 执行模块代码
func (ms *System) executeModule(code string, module *Module) error {
	ext := filepath.Ext(module.Filename)

	// 如果是 TypeScript 文件，先编译
	if ext == ".ts" || ext == ".tsx" {
		jsCode, err := transpileTS(code, module.Filename)
		if err != nil {
			return fmt.Errorf("failed to transpile %s: %w", module.Filename, err)
		}
		code = jsCode
	}

	// JSON 文件直接解析
	if ext == ".json" {
		result, err := ms.vm.RunString("(" + code + ")")
		if err != nil {
			return fmt.Errorf("failed to parse JSON %s: %w", module.Filename, err)
		}
		if result != nil {
			module.Exports = result.ToObject(ms.vm)
		}
		return nil
	}

	// 创建模块作用域
	moduleObj := ms.vm.NewObject()
	moduleObj.Set("exports", module.Exports)
	moduleObj.Set("id", module.ID)
	moduleObj.Set("filename", module.Filename)
	moduleObj.Set("loaded", false)
	moduleObj.Set("children", ms.vm.NewArray())
	moduleObj.Set("parent", nil)

	// 创建 require 函数
	requireFunc := func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			panic(ms.vm.NewTypeError("require() missing path"))
		}

		id := call.Arguments[0].String()
		requiredModule, err := ms.LoadModule(id, module.Filename)
		if err != nil {
			panic(ms.vm.NewGoError(err))
		}

		return requiredModule.Exports
	}

	// 创建动态 import 函数 (返回 Promise)
	importFunc := func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			panic(ms.vm.NewTypeError("import() missing path"))
		}

		id := call.Arguments[0].String()

		// 创建 Promise
		promise, resolve, reject := ms.vm.NewPromise()

		// 异步加载模块
		go func() {
			requiredModule, err := ms.LoadModule(id, module.Filename)
			if err != nil {
				reject(ms.vm.NewGoError(err))
			} else {
				resolve(requiredModule.Exports)
			}
		}()

		return ms.vm.ToValue(promise)
	}

	// 包装代码为 CommonJS 模块格式，同时支持动态 import
	wrappedCode := fmt.Sprintf(`
(function(exports, require, module, __filename, __dirname, importFunc) {
	// 添加动态 import 支持
	if (typeof globalThis !== 'undefined') {
		globalThis.import = importFunc;
	} else if (typeof global !== 'undefined') {
		global.import = importFunc;
	}
	
%s
});`, code)

	// 编译并执行
	program, err := goja.Compile(module.Filename, wrappedCode, false)
	if err != nil {
		return fmt.Errorf("failed to compile module %s: %w", module.Filename, err)
	}

	moduleFunc, err := ms.vm.RunProgram(program)
	if err != nil {
		return fmt.Errorf("failed to run module %s: %w", module.Filename, err)
	}

	// 调用模块函数
	dirname := filepath.Dir(module.Filename)
	callable, ok := goja.AssertFunction(moduleFunc)
	if !ok {
		return fmt.Errorf("module %s did not return a function", module.Filename)
	}

	_, err = callable(goja.Undefined(),
		module.Exports,
		ms.vm.ToValue(requireFunc),
		moduleObj,
		ms.vm.ToValue(module.Filename),
		ms.vm.ToValue(dirname),
		ms.vm.ToValue(importFunc),
	)

	if err != nil {
		return fmt.Errorf("failed to execute module %s: %w", module.Filename, err)
	}

	// 更新 exports (可能被重新赋值)
	if newExports := moduleObj.Get("exports"); newExports != nil {
		module.Exports = newExports.ToObject(ms.vm)
	}

	return nil
}

// Require 实现全局 require 函数
func (ms *System) Require(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(ms.vm.NewTypeError("require() missing path"))
	}

	id := call.Arguments[0].String()
	// 使用当前工作目录作为父路径
	currentDir, _ := os.Getwd()
	module, err := ms.LoadModule(id, currentDir)
	if err != nil {
		panic(ms.vm.NewGoError(err))
	}

	return module.Exports
}

// ClearCache 清除模块缓存
func (ms *System) ClearCache() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.cache = make(map[string]*Module)
}

// GetLoadedModules 获取已加载的模块列表
func (ms *System) GetLoadedModules() []string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	modules := make([]string, 0, len(ms.cache))
	for id := range ms.cache {
		modules = append(modules, id)
	}
	return modules
}

// GetBuiltinModules 获取内置模块列表
func (ms *System) GetBuiltinModules() []string {
	return ms.builtinManager.GetModuleNames()
}

// Close 关闭模块系统并清理资源
func (ms *System) Close() {
	ms.builtinManager.Close()
	ms.ClearCache()
}
