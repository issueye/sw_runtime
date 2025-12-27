package test

import (
	"testing"

	"sw_runtime/internal/builtins"
	"sw_runtime/internal/runtime"

	"github.com/dop251/goja"
)

func TestBuiltinManager(t *testing.T) {
	vm := goja.New()
	manager := builtins.NewManager(vm)

	// 测试模块是否存在
	expectedModules := []string{"path", "fs", "crypto", "zlib", "http", "redis", "sqlite"}

	for _, moduleName := range expectedModules {
		if !manager.HasModule(moduleName) {
			t.Fatalf("Expected builtin module '%s' not found", moduleName)
		}

		module, exists := manager.GetModule(moduleName)
		if !exists {
			t.Fatalf("Failed to get builtin module '%s'", moduleName)
		}

		if module == nil {
			t.Fatalf("Builtin module '%s' is nil", moduleName)
		}

		// 测试模块对象
		moduleObj := module.GetModule()
		if moduleObj == nil {
			t.Fatalf("Module object for '%s' is nil", moduleName)
		}
	}
}

func TestPathModule(t *testing.T) {
	runner := runtime.NewOrPanic()

	code := `
		const path = require('path');
		
		// 测试 path.join
		const joined = path.join('test', 'dir', 'file.txt');
		console.log('Joined path:', joined);
		
		// 测试 path.resolve
		const resolved = path.resolve('test.txt');
		console.log('Resolved path:', resolved);
		
		// 测试 path.basename
		const basename = path.basename('/path/to/file.txt');
		console.log('Basename:', basename);
		
		// 测试 path.dirname
		const dirname = path.dirname('/path/to/file.txt');
		console.log('Dirname:', dirname);
		
		// 测试 path.extname
		const extname = path.extname('file.txt');
		console.log('Extension:', extname);
		
		// 设置测试结果
		global.pathTestResults = {
			joined: joined,
			basename: basename,
			dirname: dirname,
			extname: extname
		};
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run path module test: %v", err)
	}

	// 验证结果
	results := runner.GetValue("pathTestResults")
	if results == nil {
		t.Fatal("Path test results not found")
	}
}

func TestFSModule(t *testing.T) {
	runner := runtime.NewOrPanic()

	code := `
		const fs = require('fs');
		
		// 测试文件操作（这里只测试方法是否存在）
		let fsTestResults = {
			hasReadFile: typeof fs.readFile === 'function',
			hasWriteFile: typeof fs.writeFile === 'function',
			hasExists: typeof fs.exists === 'function',
			hasExistsSync: typeof fs.existsSync === 'function',
			hasMkdir: typeof fs.mkdir === 'function',
			hasReaddir: typeof fs.readdir === 'function'
		};
		
		// 测试同步 exists 方法
		try {
			fsTestResults.existsSyncWorks = typeof fs.existsSync('.') === 'boolean';
		} catch (e) {
			fsTestResults.existsSyncWorks = false;
		}
		
		global.fsTestResults = fsTestResults;
		console.log('FS module methods available:', fsTestResults);
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run fs module test: %v", err)
	}

	results := runner.GetValue("fsTestResults")
	if results == nil {
		t.Fatal("FS test results not found")
	}
}

func TestCryptoModule(t *testing.T) {
	runner := runtime.NewOrPanic()

	code := `
		const crypto = require('crypto');
		
		// 测试哈希功能
		let cryptoTestResults = {
			hasMd5: typeof crypto.md5 === 'function',
			hasSha1: typeof crypto.sha1 === 'function',
			hasSha256: typeof crypto.sha256 === 'function',
			hasBase64Encode: typeof crypto.base64Encode === 'function',
			hasBase64Decode: typeof crypto.base64Decode === 'function'
		};
		
		// 如果方法存在，测试基本功能
		if (cryptoTestResults.hasMd5) {
			try {
				const hash = crypto.md5('test');
				cryptoTestResults.md5Works = typeof hash === 'string' && hash.length > 0;
			} catch (e) {
				cryptoTestResults.md5Works = false;
			}
		}
		
		if (cryptoTestResults.hasBase64Encode) {
			try {
				const encoded = crypto.base64Encode('hello');
				cryptoTestResults.base64Works = typeof encoded === 'string' && encoded.length > 0;
			} catch (e) {
				cryptoTestResults.base64Works = false;
			}
		}
		
		global.cryptoTestResults = cryptoTestResults;
		console.log('Crypto module test results:', cryptoTestResults);
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run crypto module test: %v", err)
	}

	results := runner.GetValue("cryptoTestResults")
	if results == nil {
		t.Fatal("Crypto test results not found")
	}
}

func TestHTTPModule(t *testing.T) {
	runner := runtime.NewOrPanic()

	code := `
		const http = require('http');
		
		// 测试 HTTP 模块方法
		let httpTestResults = {
			hasGet: typeof http.get === 'function',
			hasPost: typeof http.post === 'function',
			hasPut: typeof http.put === 'function',
			hasDelete: typeof http.delete === 'function',
			hasRequest: typeof http.request === 'function'
		};
		
		global.httpTestResults = httpTestResults;
		console.log('HTTP module methods available:', httpTestResults);
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run http module test: %v", err)
	}

	results := runner.GetValue("httpTestResults")
	if results == nil {
		t.Fatal("HTTP test results not found")
	}
}

func TestCompressionModule(t *testing.T) {
	runner := runtime.NewOrPanic()

	code := `
		const zlib = require('zlib');
		
		// 测试压缩模块方法
		let zlibTestResults = {
			hasGzip: typeof zlib.gzip === 'function',
			hasGunzip: typeof zlib.gunzip === 'function',
			hasDeflate: typeof zlib.deflate === 'function',
			hasInflate: typeof zlib.inflate === 'function'
		};
		
		// 测试基本压缩功能
		if (zlibTestResults.hasGzip) {
			try {
				const testData = 'Hello, World! This is a test string for compression.';
				const compressed = zlib.gzip(testData);
				zlibTestResults.gzipWorks = typeof compressed === 'string' && compressed.length > 0;
				
				// 测试解压
				if (zlibTestResults.hasGunzip) {
					const decompressed = zlib.gunzip(compressed);
					zlibTestResults.gzipRoundTrip = decompressed === testData;
				}
			} catch (e) {
				zlibTestResults.gzipWorks = false;
				zlibTestResults.gzipError = e.message;
			}
		}
		
		// 测试 deflate/inflate
		if (zlibTestResults.hasDeflate && zlibTestResults.hasInflate) {
			try {
				const testData = 'Test data for deflate/inflate';
				const compressed = zlib.deflate(testData);
				const decompressed = zlib.inflate(compressed);
				zlibTestResults.deflateWorks = decompressed === testData;
			} catch (e) {
				zlibTestResults.deflateWorks = false;
				zlibTestResults.deflateError = e.message;
			}
		}
		
		global.zlibTestResults = zlibTestResults;
		console.log('Zlib module test results:', zlibTestResults);
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run compression module test: %v", err)
	}

	results := runner.GetValue("zlibTestResults")
	if results == nil {
		t.Fatal("Compression test results not found")
	}
}

func TestModuleRegistration(t *testing.T) {
	vm := goja.New()
	manager := builtins.NewManager(vm)

	// 获取初始模块数量
	initialModules := manager.GetModuleNames()
	initialCount := len(initialModules)

	// 创建一个简单的测试模块
	testModule := &TestModule{vm: vm}
	manager.RegisterModule("testmodule", testModule)

	// 验证模块已注册
	if !manager.HasModule("testmodule") {
		t.Fatal("Test module was not registered")
	}

	// 验证模块数量增加
	newModules := manager.GetModuleNames()
	if len(newModules) != initialCount+1 {
		t.Fatalf("Expected %d modules, got %d", initialCount+1, len(newModules))
	}

	// 验证可以获取模块
	module, exists := manager.GetModule("testmodule")
	if !exists {
		t.Fatal("Failed to get registered test module")
	}

	if module != testModule {
		t.Fatal("Retrieved module is not the same as registered module")
	}
}

// TestModule 用于测试的简单模块实现
type TestModule struct {
	vm *goja.Runtime
}

func (tm *TestModule) GetModule() *goja.Object {
	module := tm.vm.NewObject()
	module.Set("test", func() string {
		return "test module works"
	})
	return module
}
