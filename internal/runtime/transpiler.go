package runtime

import (
	"fmt"

	"github.com/evanw/esbuild/pkg/api"
)

// transpileTS 使用 esbuild 将 TypeScript 转换为 JavaScript
func transpileTS(code string, filename string) (string, error) {
	result := api.Transform(code, api.TransformOptions{
		Loader:            api.LoaderTS,
		Target:            api.ES2020,
		Format:            api.FormatCommonJS,
		Sourcefile:        filename,
		MinifyWhitespace:  false,
		MinifyIdentifiers: false,
		MinifySyntax:      false,
	})

	if len(result.Errors) > 0 {
		return "", fmt.Errorf("transpile error: %s", result.Errors[0].Text)
	}

	return string(result.Code), nil
}
