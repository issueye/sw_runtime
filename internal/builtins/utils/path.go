package utils

import (
	"path/filepath"

	"github.com/dop251/goja"
)

// PathModule 路径模块
type PathModule struct {
	vm *goja.Runtime
}

// NewPathModule 创建路径模块
func NewPathModule(vm *goja.Runtime) *PathModule {
	return &PathModule{vm: vm}
}

// GetModule 获取路径模块对象
func (p *PathModule) GetModule() *goja.Object {
	obj := p.vm.NewObject()

	obj.Set("join", p.join)
	obj.Set("resolve", p.resolve)
	obj.Set("dirname", p.dirname)
	obj.Set("basename", p.basename)
	obj.Set("extname", p.extname)
	obj.Set("sep", filepath.Separator)
	obj.Set("delimiter", string(filepath.ListSeparator))
	obj.Set("isAbsolute", p.isAbsolute)
	obj.Set("relative", p.relative)
	obj.Set("normalize", p.normalize)

	return obj
}

// join 路径连接
func (p *PathModule) join(call goja.FunctionCall) goja.Value {
	parts := make([]string, len(call.Arguments))
	for i, arg := range call.Arguments {
		parts[i] = arg.String()
	}
	return p.vm.ToValue(filepath.Join(parts...))
}

// resolve 路径解析
func (p *PathModule) resolve(call goja.FunctionCall) goja.Value {
	parts := make([]string, len(call.Arguments))
	for i, arg := range call.Arguments {
		parts[i] = arg.String()
	}
	abs, _ := filepath.Abs(filepath.Join(parts...))
	return p.vm.ToValue(abs)
}

// dirname 获取目录名
func (p *PathModule) dirname(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		return goja.Undefined()
	}
	return p.vm.ToValue(filepath.Dir(call.Arguments[0].String()))
}

// basename 获取基础名
func (p *PathModule) basename(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		return goja.Undefined()
	}
	path := call.Arguments[0].String()
	if len(call.Arguments) > 1 {
		ext := call.Arguments[1].String()
		base := filepath.Base(path)
		if filepath.Ext(base) == ext {
			return p.vm.ToValue(base[:len(base)-len(ext)])
		}
		return p.vm.ToValue(base)
	}
	return p.vm.ToValue(filepath.Base(path))
}

// extname 获取扩展名
func (p *PathModule) extname(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		return goja.Undefined()
	}
	return p.vm.ToValue(filepath.Ext(call.Arguments[0].String()))
}

// isAbsolute 判断是否为绝对路径
func (p *PathModule) isAbsolute(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		return p.vm.ToValue(false)
	}
	return p.vm.ToValue(filepath.IsAbs(call.Arguments[0].String()))
}

// relative 获取相对路径
func (p *PathModule) relative(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		return goja.Undefined()
	}
	from := call.Arguments[0].String()
	to := call.Arguments[1].String()
	rel, err := filepath.Rel(from, to)
	if err != nil {
		panic(p.vm.NewGoError(err))
	}
	return p.vm.ToValue(rel)
}

// normalize 规范化路径
func (p *PathModule) normalize(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		return goja.Undefined()
	}
	return p.vm.ToValue(filepath.Clean(call.Arguments[0].String()))
}
