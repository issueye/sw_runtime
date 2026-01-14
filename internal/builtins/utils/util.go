package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/dop251/goja"
)

// UtilModule 工具模块
type UtilModule struct {
	vm *goja.Runtime
}

// NewUtilModule 创建工具模块
func NewUtilModule(vm *goja.Runtime) *UtilModule {
	return &UtilModule{vm: vm}
}

// GetModule 获取工具模块对象
func (u *UtilModule) GetModule() *goja.Object {
	obj := u.vm.NewObject()

	obj.Set("format", u.format)
	obj.Set("inspect", u.inspect)
	obj.Set("isDeepStrictEqual", u.isDeepStrictEqual)

	// Types 子模块
	types := u.vm.NewObject()
	types.Set("isDate", u.isDate)
	types.Set("isRegExp", u.isRegExp)
	types.Set("isPromise", u.isPromise)
	obj.Set("types", types)

	return obj
}

// format 格式化字符串
func (u *UtilModule) format(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		return u.vm.ToValue("")
	}

	f := call.Arguments[0].String()
	if len(call.Arguments) == 1 {
		return u.vm.ToValue(f)
	}

	args := make([]interface{}, len(call.Arguments)-1)
	for i := 0; i < len(call.Arguments)-1; i++ {
		args[i] = call.Arguments[i+1].Export()
	}

	// 简单的 printf 风格实现
	result := fmt.Sprintf(strings.ReplaceAll(f, "%j", "%v"), args...)
	return u.vm.ToValue(result)
}

// inspect 对象检查
func (u *UtilModule) inspect(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		return goja.Undefined()
	}

	arg := call.Arguments[0]
	// 简单实现：使用 Export 并格式化
	return u.vm.ToValue(fmt.Sprintf("%#v", arg.Export()))
}

// isDeepStrictEqual 深度比较
func (u *UtilModule) isDeepStrictEqual(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		return u.vm.ToValue(false)
	}

	val1 := call.Arguments[0].Export()
	val2 := call.Arguments[1].Export()

	return u.vm.ToValue(reflect.DeepEqual(val1, val2))
}

// isDate 是否为日期对象
func (u *UtilModule) isDate(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		return u.vm.ToValue(false)
	}
	// 在 goja 中，Date 对象导出为 time.Time
	_, ok := call.Arguments[0].Export().(reflect.Value)
	if !ok {
		// 尝试通过内置标识检查 (简化版)
		return u.vm.ToValue(strings.Contains(call.Arguments[0].String(), "GMT"))
	}
	return u.vm.ToValue(false)
}

// isRegExp 是否为正则对象
func (u *UtilModule) isRegExp(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		return u.vm.ToValue(false)
	}
	// goja 中正则对象通常导出为 *regexp.Regexp 或特定内部结构
	return u.vm.ToValue(strings.HasPrefix(call.Arguments[0].String(), "/"))
}

// isPromise 是否为 Promise 对象
func (u *UtilModule) isPromise(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		return u.vm.ToValue(false)
	}
	_, ok := call.Arguments[0].Export().(*goja.Promise)
	return u.vm.ToValue(ok)
}
