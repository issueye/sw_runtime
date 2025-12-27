package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sw_runtime/internal/bundler"
)

func TestBundlerEncryption(t *testing.T) {
	tempDir := t.TempDir()

	// 创建测试模块
	utilsContent := `
exports.secret = function() {
	return "This is secret code";
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
console.log(utils.secret());
`
	mainFile := filepath.Join(tempDir, "main.js")
	err = os.WriteFile(mainFile, []byte(mainContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create main.js: %v", err)
	}

	// 测试加密打包
	outputFile := filepath.Join(tempDir, "bundle.js")
	b := bundler.New(bundler.Options{
		EntryFile:  mainFile,
		OutputFile: outputFile,
		Encrypt:    true,
	})

	result, err := b.Bundle()
	if err != nil {
		t.Fatalf("Encryption bundle failed: %v", err)
	}

	// 验证加密结果
	if !result.Encrypted {
		t.Error("Result should be encrypted")
	}

	if result.EncryptKey == "" {
		t.Error("Encryption key should be generated")
	}

	if len(result.EncryptKey) == 0 {
		t.Error("Encryption key should not be empty")
	}

	// 验证生成的代码包含加密标记
	if !strings.Contains(result.Code, "ENCRYPTED_CODE") {
		t.Error("Encrypted bundle should contain ENCRYPTED_CODE")
	}

	if !strings.Contains(result.Code, "SW Runtime Encrypted Bundle") {
		t.Error("Encrypted bundle should contain header comment")
	}

	// 验证原始代码不在加密文件中
	if strings.Contains(result.Code, "This is secret code") {
		t.Error("Original code should not be visible in encrypted bundle")
	}

	t.Logf("Encryption successful, key: %s", result.EncryptKey)
	t.Logf("Bundle size: %d bytes", len(result.Code))
}

func TestBundlerEncryptionWithCustomKey(t *testing.T) {
	tempDir := t.TempDir()

	// 创建简单模块
	content := `exports.value = 42;`
	moduleFile := filepath.Join(tempDir, "module.js")
	os.WriteFile(moduleFile, []byte(content), 0644)

	mainContent := `
const mod = require('./module.js');
console.log(mod.value);
`
	mainFile := filepath.Join(tempDir, "main.js")
	os.WriteFile(mainFile, []byte(mainContent), 0644)

	// 使用自定义密钥（32字节）
	// "12345678901234567890123456789012" = 32 bytes
	customKey := "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI="

	b := bundler.New(bundler.Options{
		EntryFile:  mainFile,
		OutputFile: filepath.Join(tempDir, "bundle.js"),
		Encrypt:    true,
		EncryptKey: customKey,
	})

	result, err := b.Bundle()
	if err != nil {
		t.Fatalf("Bundle with custom key failed: %v", err)
	}

	// 验证使用的是自定义密钥
	if result.EncryptKey != customKey {
		t.Errorf("Expected custom key %s, got %s", customKey, result.EncryptKey)
	}

	t.Logf("Custom key encryption successful")
}

func TestBundlerEncryptionAndMinify(t *testing.T) {
	tempDir := t.TempDir()

	// 创建测试文件
	content := `
function longFunctionName(veryLongParameterName) {
	const anotherLongVariableName = veryLongParameterName * 2;
	return anotherLongVariableName + 10;
}
console.log(longFunctionName(5));
`
	mainFile := filepath.Join(tempDir, "main.js")
	os.WriteFile(mainFile, []byte(content), 0644)

	// 加密 + 压缩
	b := bundler.New(bundler.Options{
		EntryFile:  mainFile,
		OutputFile: filepath.Join(tempDir, "bundle.js"),
		Encrypt:    true,
		Minify:     true,
	})

	result, err := b.Bundle()
	if err != nil {
		t.Fatalf("Encryption + minify failed: %v", err)
	}

	// 验证同时启用了加密和压缩
	if !result.Encrypted {
		t.Error("Should be encrypted")
	}

	// 加密文件应该更小（因为加密前的代码被压缩了）
	if len(result.Code) == 0 {
		t.Error("Bundle should not be empty")
	}

	t.Logf("Encryption + minify successful, size: %d bytes", len(result.Code))
}

func TestBundlerNonEncrypted(t *testing.T) {
	tempDir := t.TempDir()

	content := `console.log("Hello");`
	mainFile := filepath.Join(tempDir, "main.js")
	os.WriteFile(mainFile, []byte(content), 0644)

	// 不加密
	b := bundler.New(bundler.Options{
		EntryFile:  mainFile,
		OutputFile: filepath.Join(tempDir, "bundle.js"),
		Encrypt:    false,
	})

	result, err := b.Bundle()
	if err != nil {
		t.Fatalf("Non-encrypted bundle failed: %v", err)
	}

	// 验证未加密
	if result.Encrypted {
		t.Error("Should not be encrypted")
	}

	if result.EncryptKey != "" {
		t.Error("Non-encrypted bundle should not have encryption key")
	}

	// 原始代码应该可见
	if !strings.Contains(result.Code, "Hello") {
		t.Error("Original code should be visible in non-encrypted bundle")
	}

	t.Logf("Non-encrypted bundle successful")
}
