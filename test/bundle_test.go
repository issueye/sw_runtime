package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sw_runtime/internal/bundler"
)

func TestBundlerBasic(t *testing.T) {
	// 创建临时测试目录
	tempDir := t.TempDir()

	// 创建测试文件
	utilsContent := `
exports.add = function(a, b) {
	return a + b;
};

exports.multiply = function(a, b) {
	return a * b;
};
`
	utilsFile := filepath.Join(tempDir, "utils.js")
	err := os.WriteFile(utilsFile, []byte(utilsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create utils.js: %v", err)
	}

	// 创建主文件
	mainContent := `
const utils = require('./utils.js');

const result1 = utils.add(5, 3);
const result2 = utils.multiply(4, 7);

console.log('Add:', result1);
console.log('Multiply:', result2);
`
	mainFile := filepath.Join(tempDir, "main.js")
	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create main.js: %v", err)
	}

	// 执行打包
	outputFile := filepath.Join(tempDir, "bundle.js")
	b := bundler.New(bundler.Options{
		EntryFile:  mainFile,
		OutputFile: outputFile,
		Minify:     false,
	})

	result, err := b.Bundle()
	if err != nil {
		t.Fatalf("Bundle failed: %v", err)
	}

	// 验证结果
	if len(result.Code) == 0 {
		t.Fatal("Bundle code is empty")
	}

	if len(result.Modules) != 2 {
		t.Fatalf("Expected 2 modules, got %d", len(result.Modules))
	}

	// 验证包含的模块
	if !strings.Contains(strings.Join(result.Modules, ","), "main.js") {
		t.Error("Bundle should contain main.js")
	}

	if !strings.Contains(strings.Join(result.Modules, ","), "utils.js") {
		t.Error("Bundle should contain utils.js")
	}

	t.Logf("Bundle successful, size: %d bytes, modules: %d", len(result.Code), len(result.Modules))
}

func TestBundlerTypeScript(t *testing.T) {
	tempDir := t.TempDir()

	// 创建 TypeScript 模块
	tsContent := `
function greet(name: string): string {
	return "Hello, " + name + "!";
}

exports.greet = greet;
`
	tsFile := filepath.Join(tempDir, "greeter.ts")
	err := os.WriteFile(tsFile, []byte(tsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create greeter.ts: %v", err)
	}

	// 创建主文件
	mainContent := `
const greeter = require('./greeter.ts');
console.log(greeter.greet('TypeScript'));
`
	mainFile := filepath.Join(tempDir, "main.js")
	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create main.js: %v", err)
	}

	// 执行打包
	outputFile := filepath.Join(tempDir, "bundle.js")
	b := bundler.New(bundler.Options{
		EntryFile:  mainFile,
		OutputFile: outputFile,
		Minify:     false,
	})

	result, err := b.Bundle()
	if err != nil {
		t.Fatalf("Bundle failed: %v", err)
	}

	// 验证 TypeScript 被编译
	if !strings.Contains(result.Code, "Hello") {
		t.Error("Bundle should contain compiled TypeScript code")
	}

	if len(result.Modules) != 2 {
		t.Fatalf("Expected 2 modules, got %d", len(result.Modules))
	}

	t.Logf("TypeScript bundle successful, size: %d bytes", len(result.Code))
}

func TestBundlerMinify(t *testing.T) {
	tempDir := t.TempDir()

	// 创建简单模块
	content := `
exports.longFunctionName = function(veryLongParameterName) {
	const anotherLongVariableName = veryLongParameterName * 2;
	return anotherLongVariableName + 10;
};
`
	moduleFile := filepath.Join(tempDir, "module.js")
	err := os.WriteFile(moduleFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create module.js: %v", err)
	}

	mainContent := `
const mod = require('./module.js');
console.log(mod.longFunctionName(5));
`
	mainFile := filepath.Join(tempDir, "main.js")
	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create main.js: %v", err)
	}

	// 不压缩打包
	b1 := bundler.New(bundler.Options{
		EntryFile:  mainFile,
		OutputFile: filepath.Join(tempDir, "bundle.js"),
		Minify:     false,
	})
	result1, err := b1.Bundle()
	if err != nil {
		t.Fatalf("Non-minified bundle failed: %v", err)
	}

	// 压缩打包
	b2 := bundler.New(bundler.Options{
		EntryFile:  mainFile,
		OutputFile: filepath.Join(tempDir, "bundle.min.js"),
		Minify:     true,
	})
	result2, err := b2.Bundle()
	if err != nil {
		t.Fatalf("Minified bundle failed: %v", err)
	}

	// 压缩后应该更小
	if len(result2.Code) >= len(result1.Code) {
		t.Errorf("Minified bundle (%d bytes) should be smaller than non-minified (%d bytes)",
			len(result2.Code), len(result1.Code))
	}

	reduction := float64(len(result1.Code)-len(result2.Code)) / float64(len(result1.Code)) * 100
	t.Logf("Minification reduced size by %.1f%% (%d -> %d bytes)",
		reduction, len(result1.Code), len(result2.Code))
}

func TestBundlerBuiltinExclusion(t *testing.T) {
	tempDir := t.TempDir()

	// 创建使用内置模块的文件
	mainContent := `
const fs = require('fs');
const http = require('http');
const crypto = require('crypto');

console.log('Using builtin modules');
`
	mainFile := filepath.Join(tempDir, "main.js")
	err := os.WriteFile(mainFile, []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create main.js: %v", err)
	}

	// 执行打包
	b := bundler.New(bundler.Options{
		EntryFile:  mainFile,
		OutputFile: filepath.Join(tempDir, "bundle.js"),
		Minify:     false,
	})

	result, err := b.Bundle()
	if err != nil {
		t.Fatalf("Bundle failed: %v", err)
	}

	// 应该只有一个模块（main.js），内置模块被排除
	if len(result.Modules) != 1 {
		t.Fatalf("Expected 1 module (main.js only), got %d", len(result.Modules))
	}

	// 验证代码中仍然包含 require 调用（内置模块在运行时加载）
	if !strings.Contains(result.Code, "require") {
		t.Error("Bundle should still contain require calls for builtin modules")
	}

	t.Logf("Builtin modules correctly excluded, bundle size: %d bytes", len(result.Code))
}

func TestBundlerExcludeFiles(t *testing.T) {
	tempDir := t.TempDir()

	// 创建多个模块
	module1Content := `exports.value = 'module1';`
	module1File := filepath.Join(tempDir, "module1.js")
	os.WriteFile(module1File, []byte(module1Content), 0644)

	module2Content := `exports.value = 'module2';`
	module2File := filepath.Join(tempDir, "module2.js")
	os.WriteFile(module2File, []byte(module2Content), 0644)

	mainContent := `
const m1 = require('./module1.js');
const m2 = require('./module2.js');
console.log(m1.value, m2.value);
`
	mainFile := filepath.Join(tempDir, "main.js")
	os.WriteFile(mainFile, []byte(mainContent), 0644)

	// 执行打包，排除 module2.js
	b := bundler.New(bundler.Options{
		EntryFile:    mainFile,
		OutputFile:   filepath.Join(tempDir, "bundle.js"),
		Minify:       false,
		ExcludeFiles: []string{module2File},
	})

	result, err := b.Bundle()
	if err != nil {
		t.Fatalf("Bundle failed: %v", err)
	}

	// 应该只包含 main.js 和 module1.js
	if len(result.Modules) != 2 {
		t.Fatalf("Expected 2 modules (excluding module2), got %d", len(result.Modules))
	}

	// 验证不包含 module2
	moduleList := strings.Join(result.Modules, ",")
	if strings.Contains(moduleList, "module2.js") {
		t.Error("Bundle should not contain excluded module2.js")
	}

	if !strings.Contains(moduleList, "module1.js") {
		t.Error("Bundle should contain module1.js")
	}

	t.Logf("Exclude files working correctly, modules: %d", len(result.Modules))
}
