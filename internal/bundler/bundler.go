package bundler

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

// Options 打包选项
type Options struct {
	EntryFile    string   // 入口文件
	OutputFile   string   // 输出文件
	Minify       bool     // 是否压缩
	Sourcemap    bool     // 是否生成 source map
	ExcludeFiles []string // 排除的文件列表
	Encrypt      bool     // 是否加密
	EncryptKey   string   // 加密密钥（如果为空则自动生成）
}

// Result 打包结果
type Result struct {
	Code       string   // 打包后的代码
	Sourcemap  string   // Source map (如果生成)
	Modules    []string // 包含的模块列表
	Encrypted  bool     // 是否加密
	EncryptKey string   // 加密密钥（仅当加密时有效）
}

// Bundler 打包器
type Bundler struct {
	options        Options
	modules        map[string]bool // 已处理的模块
	moduleOrder    []string        // 模块顺序
	builtinModules map[string]bool // 内置模块列表
	excludeSet     map[string]bool // 排除文件集合
	basePath       string          // 基础路径
}

// 内置模块列表
var defaultBuiltinModules = []string{
	"server", "sqlite", "websocket", "ws", "fs", "crypto",
	"zlib", "compression", "http", "redis", "exec",
	"child_process", "path", "httpserver",
}

// New 创建新的打包器
func New(options Options) *Bundler {
	// 获取入口文件的绝对路径作为基础路径
	entryAbs, _ := filepath.Abs(options.EntryFile)
	basePath := filepath.Dir(entryAbs)

	// 创建排除文件集合
	excludeSet := make(map[string]bool)
	for _, file := range options.ExcludeFiles {
		absPath, _ := filepath.Abs(file)
		excludeSet[absPath] = true
	}

	// 创建内置模块映射
	builtinMap := make(map[string]bool)
	for _, mod := range defaultBuiltinModules {
		builtinMap[mod] = true
	}

	return &Bundler{
		options:        options,
		modules:        make(map[string]bool),
		moduleOrder:    make([]string, 0),
		builtinModules: builtinMap,
		excludeSet:     excludeSet,
		basePath:       basePath,
	}
}

// Bundle 执行打包
func (b *Bundler) Bundle() (*Result, error) {
	// 解析入口文件及其依赖
	entryAbs, err := filepath.Abs(b.options.EntryFile)
	if err != nil {
		return nil, fmt.Errorf("无法解析入口文件路径: %w", err)
	}

	err = b.analyzeModule(entryAbs, "")
	if err != nil {
		return nil, err
	}

	// 使用 esbuild 进行打包
	buildOptions := api.BuildOptions{
		EntryPoints:       []string{entryAbs},
		Bundle:            true,
		Write:             false,
		Platform:          api.PlatformNode,
		Format:            api.FormatCommonJS,
		Target:            api.ES2020,
		MinifyWhitespace:  b.options.Minify,
		MinifyIdentifiers: b.options.Minify,
		MinifySyntax:      b.options.Minify,
		Sourcemap:         api.SourceMapNone,
		External:          b.getExternalModules(),
	}

	if b.options.Sourcemap {
		buildOptions.Sourcemap = api.SourceMapInline
	}

	result := api.Build(buildOptions)

	if len(result.Errors) > 0 {
		errorMsg := "打包错误:\n"
		for _, err := range result.Errors {
			errorMsg += fmt.Sprintf("  %s\n", err.Text)
		}
		return nil, fmt.Errorf(errorMsg)
	}

	if len(result.OutputFiles) == 0 {
		return nil, fmt.Errorf("打包未生成输出文件")
	}

	code := string(result.OutputFiles[0].Contents)
	sourcemap := ""

	// 提取 sourcemap（如果存在）
	if b.options.Sourcemap && len(result.OutputFiles) > 1 {
		sourcemap = string(result.OutputFiles[1].Contents)
	}

	// 如果需要加密
	encryptKey := ""
	encrypted := false
	if b.options.Encrypt {
		// 生成或使用提供的密钥
		if b.options.EncryptKey != "" {
			encryptKey = b.options.EncryptKey
		} else {
			var err error
			encryptKey, err = generateEncryptKey()
			if err != nil {
				return nil, fmt.Errorf("生成加密密钥失败: %w", err)
			}
		}

		// 加密代码
		encryptedCode, err := encryptCode(code, encryptKey)
		if err != nil {
			return nil, fmt.Errorf("加密代码失败: %w", err)
		}

		// 生成加密包装器
		code = wrapEncryptedCode(encryptedCode)
		encrypted = true
	}

	return &Result{
		Code:       code,
		Sourcemap:  sourcemap,
		Modules:    b.moduleOrder,
		Encrypted:  encrypted,
		EncryptKey: encryptKey,
	}, nil
}

// analyzeModule 分析模块及其依赖
func (b *Bundler) analyzeModule(modulePath string, parentPath string) error {
	// 规范化路径
	absPath, err := filepath.Abs(modulePath)
	if err != nil {
		return err
	}

	// 检查是否已处理或需要排除
	if b.modules[absPath] || b.excludeSet[absPath] {
		return nil
	}

	// 标记为已处理
	b.modules[absPath] = true
	b.moduleOrder = append(b.moduleOrder, absPath)

	// 读取文件内容
	content, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("读取模块 %s 失败: %w", absPath, err)
	}

	// 如果是 TypeScript，先编译
	ext := filepath.Ext(absPath)
	code := string(content)
	if ext == ".ts" || ext == ".tsx" {
		result := api.Transform(code, api.TransformOptions{
			Loader: api.LoaderTS,
			Target: api.ES2020,
			Format: api.FormatCommonJS,
		})

		if len(result.Errors) > 0 {
			return fmt.Errorf("编译 TypeScript 失败: %s", result.Errors[0].Text)
		}

		code = string(result.Code)
	}

	// 解析依赖
	dependencies := b.extractDependencies(code)

	// 递归处理依赖
	for _, dep := range dependencies {
		// 跳过内置模块
		if b.builtinModules[dep] {
			continue
		}

		// 解析依赖路径
		resolvedPath, err := b.resolveModule(dep, absPath)
		if err != nil {
			// 如果解析失败，可能是内置模块或外部模块，跳过
			continue
		}

		// 递归分析依赖
		err = b.analyzeModule(resolvedPath, absPath)
		if err != nil {
			return err
		}
	}

	return nil
}

// extractDependencies 从代码中提取依赖
func (b *Bundler) extractDependencies(code string) []string {
	dependencies := make([]string, 0)

	// 匹配 require('module') 和 require("module")
	requireRegex := regexp.MustCompile(`require\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	matches := requireRegex.FindAllStringSubmatch(code, -1)

	for _, match := range matches {
		if len(match) > 1 {
			dependencies = append(dependencies, match[1])
		}
	}

	// 匹配 import ... from 'module'
	importRegex := regexp.MustCompile(`import\s+.*?\s+from\s+['"]([^'"]+)['"]`)
	matches = importRegex.FindAllStringSubmatch(code, -1)

	for _, match := range matches {
		if len(match) > 1 {
			dependencies = append(dependencies, match[1])
		}
	}

	// 匹配 import('module') 动态导入
	dynamicImportRegex := regexp.MustCompile(`import\s*\(\s*['"]([^'"]+)['"]\s*\)`)
	matches = dynamicImportRegex.FindAllStringSubmatch(code, -1)

	for _, match := range matches {
		if len(match) > 1 {
			dependencies = append(dependencies, match[1])
		}
	}

	return dependencies
}

// resolveModule 解析模块路径
func (b *Bundler) resolveModule(id string, parentPath string) (string, error) {
	// 相对路径
	if strings.HasPrefix(id, "./") || strings.HasPrefix(id, "../") {
		basePath := filepath.Dir(parentPath)
		resolved := filepath.Join(basePath, id)

		// 尝试不同的扩展名
		extensions := []string{"", ".js", ".ts", ".tsx", ".json"}
		for _, ext := range extensions {
			fullPath := resolved + ext
			if _, err := os.Stat(fullPath); err == nil {
				return filepath.Abs(fullPath)
			}
		}

		// 尝试 index 文件
		indexExtensions := []string{"/index.js", "/index.ts", "/index.tsx", "/index.json"}
		for _, ext := range indexExtensions {
			fullPath := resolved + ext
			if _, err := os.Stat(fullPath); err == nil {
				return filepath.Abs(fullPath)
			}
		}

		return "", fmt.Errorf("模块未找到: %s", id)
	}

	// 绝对路径
	if filepath.IsAbs(id) {
		if _, err := os.Stat(id); err == nil {
			return filepath.Abs(id)
		}
		return "", fmt.Errorf("模块未找到: %s", id)
	}

	// node_modules 查找（简化实现）
	nodeModulesPath := filepath.Join(b.basePath, "node_modules", id)
	if _, err := os.Stat(nodeModulesPath); err == nil {
		return filepath.Abs(nodeModulesPath)
	}

	return "", fmt.Errorf("模块未找到: %s", id)
}

// getExternalModules 获取外部模块列表（内置模块）
func (b *Bundler) getExternalModules() []string {
	external := make([]string, 0, len(b.builtinModules))
	for mod := range b.builtinModules {
		external = append(external, mod)
	}
	return external
}

// generateEncryptKey 生成随机加密密钥 (32字节 = 256位)
func generateEncryptKey() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// encryptCode 使用 AES-256-GCM 加密代码
func encryptCode(code string, keyStr string) (string, error) {
	// 解码 base64 密钥
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return "", fmt.Errorf("无效的密钥格式: %w", err)
	}

	// 创建 AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 创建 GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机 nonce
	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return "", err
	}

	// 加密
	ciphertext := gcm.Seal(nonce, nonce, []byte(code), nil)

	// 返回 base64 编码的加密数据
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// wrapEncryptedCode 包装加密代码，生成带解密器的脚本
func wrapEncryptedCode(encryptedCode string) string {
	return fmt.Sprintf(`// SW Runtime Encrypted Bundle
// This file contains encrypted code and requires a decryption key to run.
// Use: sw_runtime run --decrypt-key=<key> <file>

const ENCRYPTED_CODE = %q;

if (typeof __SW_DECRYPT_KEY__ === 'undefined') {
    console.error('❌ 这是一个加密的 bundle 文件，需要解密密钥运行。');
    console.error('使用方法: sw_runtime run --decrypt-key=<key> <file>');
    throw new Error('缺少解密密钥');
}

// 这里的解密逻辑由 SW Runtime 在运行时处理
// __SW_DECRYPT_KEY__ 会被 SW Runtime 注入
eval(__SW_DECRYPT_KEY__);
`, encryptedCode)
}
