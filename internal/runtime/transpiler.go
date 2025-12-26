package runtime

import (
	"fmt"
	"sync"

	"github.com/evanw/esbuild/pkg/api"
)

// TranspilerPool TypeScript 编译器池
type TranspilerPool struct {
	pool sync.Pool
}

// GlobalTranspilerPool 全局编译器池
var GlobalTranspilerPool = &TranspilerPool{
	pool: sync.Pool{
		New: func() interface{} {
			return &api.TransformOptions{
				Loader:            api.LoaderTS,
				Target:            api.ES2020,
				Format:            api.FormatCommonJS,
				MinifyWhitespace:  false,
				MinifyIdentifiers: false,
				MinifySyntax:      false,
			}
		},
	},
}

// GetTransformOptions 获取编译选项
func (tp *TranspilerPool) GetTransformOptions() *api.TransformOptions {
	return tp.pool.Get().(*api.TransformOptions)
}

// PutTransformOptions 归还编译选项
func (tp *TranspilerPool) PutTransformOptions(opts *api.TransformOptions) {
	// 重置选项到默认值
	opts.Loader = api.LoaderTS
	opts.Target = api.ES2020
	opts.Format = api.FormatCommonJS
	opts.MinifyWhitespace = false
	opts.MinifyIdentifiers = false
	opts.MinifySyntax = false
	opts.Sourcefile = ""

	tp.pool.Put(opts)
}

// transpileTS 使用 esbuild 将 TypeScript 转换为 JavaScript
func transpileTS(code string, filename string) (string, error) {
	// 使用对象池获取编译选项
	opts := GlobalTranspilerPool.GetTransformOptions()
	defer GlobalTranspilerPool.PutTransformOptions(opts)

	// 设置文件名
	opts.Sourcefile = filename

	result := api.Transform(code, *opts)

	if len(result.Errors) > 0 {
		return "", fmt.Errorf("transpile error: %s", result.Errors[0].Text)
	}

	return string(result.Code), nil
}
