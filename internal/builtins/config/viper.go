package config

import (
	"fmt"
	"strings"
	"sync"

	"github.com/dop251/goja"
	"github.com/spf13/viper"
)

// ViperModule viper 配置模块
type ViperModule struct {
	vm   *goja.Runtime
	mu   sync.RWMutex
	vips map[string]*viper.Viper // 支持多个 viper 实例
}

// NewViperModule 创建 viper 配置模块
func NewViperModule(vm *goja.Runtime) *ViperModule {
	return &ViperModule{
		vm:   vm,
		vips: make(map[string]*viper.Viper),
	}
}

// GetModule 获取模块对象
func (v *ViperModule) GetModule() *goja.Object {
	obj := v.vm.NewObject()

	// 创建新的 viper 实例
	obj.Set("new", v.newViper)

	// 获取默认值
	obj.Set("getDefault", v.getDefault)

	return obj
}

// newViper 创建新的 viper 实例
func (v *ViperModule) newViper(call goja.FunctionCall) goja.Value {
	name := "default"
	if len(call.Arguments) > 0 && call.Arguments[0] != goja.Undefined() {
		name = call.Arguments[0].String()
	}

	vp := viper.New()

	v.mu.Lock()
	v.vips[name] = vp
	v.mu.Unlock()

	return v.createViperObject(vp, name)
}

// createViperObject 创建 viper 实例的 JS 对象
func (v *ViperModule) createViperObject(vp *viper.Viper, name string) goja.Value {
	obj := v.vm.NewObject()

	// 设置实例名
	obj.Set("name", name)

	// === 文件配置 ===
	// SetConfigFile 设置配置文件路径
	obj.Set("setConfigFile", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(v.vm.NewTypeError("setConfigFile requires configFile argument"))
		}
		configFile := call.Arguments[0].String()
		vp.SetConfigFile(configFile)
		return goja.Undefined()
	})

	// SetConfigName 设置配置文件名（不含扩展名）
	obj.Set("setConfigName", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(v.vm.NewTypeError("setConfigName requires name argument"))
		}
		name := call.Arguments[0].String()
		vp.SetConfigName(name)
		return goja.Undefined()
	})

	// AddConfigPath 添加搜索配置文件的路径
	obj.Set("addConfigPath", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(v.vm.NewTypeError("addConfigPath requires path argument"))
		}
		path := call.Arguments[0].String()
		vp.AddConfigPath(path)
		return goja.Undefined()
	})

	// SetConfigType 设置配置类型
	obj.Set("setConfigType", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(v.vm.NewTypeError("setConfigType requires type argument"))
		}
		configType := call.Arguments[0].String()
		vp.SetConfigType(configType)
		return goja.Undefined()
	})

	// === 读取配置 ===
	// ReadInConfig 读取配置文件
	obj.Set("readInConfig", func(call goja.FunctionCall) goja.Value {
		err := vp.ReadInConfig()
		if err != nil {
			panic(v.vm.NewGoError(err))
		}
		return goja.Undefined()
	})

	// SafeWriteConfig 安全写入配置
	obj.Set("safeWriteConfig", func(call goja.FunctionCall) goja.Value {
		err := vp.SafeWriteConfig()
		if err != nil {
			panic(v.vm.NewGoError(err))
		}
		return goja.Undefined()
	})

	// === 获取值 ===
	// Get 获取值
	obj.Set("get", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		key := call.Arguments[0].String()
		return v.toJSValue(vp.Get(key))
	})

	// GetString 获取字符串
	obj.Set("getString", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		key := call.Arguments[0].String()
		return v.vm.ToValue(vp.GetString(key))
	})

	// GetInt 获取整数
	obj.Set("getInt", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		key := call.Arguments[0].String()
		return v.vm.ToValue(vp.GetInt(key))
	})

	// GetInt64 获取 64 位整数
	obj.Set("getInt64", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		key := call.Arguments[0].String()
		return v.vm.ToValue(vp.GetInt64(key))
	})

	// GetFloat64 获取浮点数
	obj.Set("getFloat64", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		key := call.Arguments[0].String()
		return v.vm.ToValue(vp.GetFloat64(key))
	})

	// GetBool 获取布尔值
	obj.Set("getBool", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		key := call.Arguments[0].String()
		return v.vm.ToValue(vp.GetBool(key))
	})

	// GetStringSlice 获取字符串数组
	obj.Set("getStringSlice", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		key := call.Arguments[0].String()
		return v.vm.ToValue(vp.GetStringSlice(key))
	})

	// GetIntSlice 获取整数数组
	obj.Set("getIntSlice", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		key := call.Arguments[0].String()
		return v.vm.ToValue(vp.GetIntSlice(key))
	})

	// GetStringMap 获取字符串映射
	obj.Set("getStringMap", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		key := call.Arguments[0].String()
		return v.vm.ToValue(vp.GetStringMap(key))
	})

	// GetStringMapString 获取字符串->字符串映射
	obj.Set("getStringMapString", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		key := call.Arguments[0].String()
		return v.vm.ToValue(vp.GetStringMapString(key))
	})

	// === 设置值 ===
	// Set 设置值
	obj.Set("set", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(v.vm.NewTypeError("set requires key and value arguments"))
		}
		key := call.Arguments[0].String()
		value := call.Arguments[1].Export()
		vp.Set(key, value)
		return goja.Undefined()
	})

	// SetDefault 设置默认值
	obj.Set("setDefault", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(v.vm.NewTypeError("setDefault requires key and value arguments"))
		}
		key := call.Arguments[0].String()
		value := call.Arguments[1].Export()
		vp.SetDefault(key, value)
		return goja.Undefined()
	})

	// === 检查键是否存在 ===
	// IsSet 检查键是否已设置
	obj.Set("isSet", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		key := call.Arguments[0].String()
		return v.vm.ToValue(vp.IsSet(key))
	})

	// IsSet 检查键是否有默认值（通过 Get 比较）
	obj.Set("isSetDefault", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return goja.Undefined()
		}
		key := call.Arguments[0].String()
		// 检查是否有默认值
		defaultVal := vp.Get(key)
		return v.vm.ToValue(defaultVal != nil)
	})

	// === 环境变量 ===
	// BindEnv 绑定环境变量
	obj.Set("bindEnv", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(v.vm.NewTypeError("bindEnv requires at least one argument"))
		}
		args := make([]string, len(call.Arguments))
		for i, arg := range call.Arguments {
			args[i] = arg.String()
		}
		vp.BindEnv(args...)
		return goja.Undefined()
	})

	// SetEnvPrefix 设置环境变量前缀
	obj.Set("setEnvPrefix", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(v.vm.NewTypeError("setEnvPrefix requires prefix argument"))
		}
		prefix := call.Arguments[0].String()
		vp.SetEnvPrefix(prefix)
		return goja.Undefined()
	})

	// AllowEmptyEnv 允许空环境变量
	obj.Set("allowEmptyEnv", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			vp.AllowEmptyEnv(true)
		} else {
			vp.AllowEmptyEnv(call.Arguments[0].ToBoolean())
		}
		return goja.Undefined()
	})

	// SetEnvKeyReplacer 设置环境变量键替换规则
	obj.Set("setEnvKeyReplacer", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			panic(v.vm.NewTypeError("setEnvKeyReplacer requires old and new arguments"))
		}
		oldStr := call.Arguments[0].String()
		newStr := call.Arguments[1].String()
		vp.SetEnvKeyReplacer(strings.NewReplacer(oldStr, newStr))
		return goja.Undefined()
	})

	// === 配置信息 ===
	// AllSettings 获取所有设置
	obj.Set("allSettings", func(call goja.FunctionCall) goja.Value {
		return v.toJSValue(vp.AllSettings())
	})

	// Keys 获取所有键
	obj.Set("keys", func(call goja.FunctionCall) goja.Value {
		keys := vp.AllKeys()
		arr := v.vm.NewArray()
		for i, key := range keys {
			arr.Set(fmt.Sprintf("%d", i), key)
		}
		return arr
	})

	// ConfigFileUsed 获取使用的配置文件路径
	obj.Set("configFileUsed", func(call goja.FunctionCall) goja.Value {
		return v.vm.ToValue(vp.ConfigFileUsed())
	})

	// Unmarshal 解码配置到对象
	obj.Set("unmarshal", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(v.vm.NewTypeError("unmarshal requires config object argument"))
		}
		cfg := call.Arguments[0].ToObject(v.vm)
		err := vp.Unmarshal(cfg)
		if err != nil {
			panic(v.vm.NewGoError(err))
		}
		return goja.Undefined()
	})

	// UnmarshalExact 解码配置到对象（精确匹配）
	obj.Set("unmarshalExact", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			panic(v.vm.NewTypeError("unmarshalExact requires config object argument"))
		}
		cfg := call.Arguments[0].ToObject(v.vm)
		err := vp.UnmarshalExact(cfg)
		if err != nil {
			panic(v.vm.NewGoError(err))
		}
		return goja.Undefined()
	})

	// === 搜索路径 ===
	// GetSearchPath 获取搜索路径
	obj.Set("getSearchPath", func(call goja.FunctionCall) goja.Value {
		// viper v1.x 没有 ConfigPathsUsed，使用 ConfigFileUsed 代替
		configFile := vp.ConfigFileUsed()
		if configFile == "" {
			return goja.Undefined()
		}
		// 提取目录路径
		idx := strings.LastIndex(configFile, "/")
		if idx < 0 {
			idx = strings.LastIndex(configFile, "\\")
		}
		if idx >= 0 {
			return v.vm.ToValue(configFile[:idx])
		}
		return goja.Undefined()
	})

	return obj
}

// getDefault 获取默认值（全局方法）
func (v *ViperModule) getDefault(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(v.vm.NewTypeError("getDefault requires key and defaultValue arguments"))
	}
	key := call.Arguments[0].String()
	defaultValue := call.Arguments[1].Export()
	// 使用 viper 全局实例获取默认值
	if viper.IsSet(key) {
		return v.toJSValue(viper.Get(key))
	}
	return v.toJSValue(defaultValue)
}

// toJSValue 将 Go 值转换为 JS 值
func (v *ViperModule) toJSValue(val interface{}) goja.Value {
	if val == nil {
		return goja.Undefined()
	}
	return v.vm.ToValue(val)
}
