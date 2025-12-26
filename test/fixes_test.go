package test

import (
	"testing"
	"time"

	"sw_runtime/internal/runtime"

	"github.com/dop251/goja"
)

// TestFixedEventLoopInterval 测试修复后的事件循环间隔功能
func TestFixedEventLoopInterval(t *testing.T) {
	runner := runtime.New()

	code := `
		let intervalExecuted = false;
		let intervalId = null;
		
		intervalId = setInterval(() => {
			intervalExecuted = true;
			console.log('Interval executed successfully');
			clearInterval(intervalId);
		}, 50);
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run fixed interval test: %v", err)
	}

	// 等待间隔执行
	time.Sleep(200 * time.Millisecond)

	executed := runner.GetValue("intervalExecuted")
	if !executed.ToBoolean() {
		t.Fatal("Interval was not executed")
	}

	t.Log("✅ Event loop interval fix verified")
}

// TestFixedCompressionModule 测试修复后的压缩模块
func TestFixedCompressionModule(t *testing.T) {
	runner := runtime.New()

	code := `
		const zlib = require('zlib');
		
		let compressionResults = {
			testsPassed: 0,
			totalTests: 0,
			errors: []
		};
		
		// 测试 gzip/gunzip
		try {
			compressionResults.totalTests++;
			const originalData = 'Hello, World! This is a test for gzip compression.';
			const compressed = zlib.gzip(originalData);
			const decompressed = zlib.gunzip(compressed);
			
			if (decompressed === originalData) {
				compressionResults.testsPassed++;
				console.log('✅ Gzip/Gunzip test passed');
			} else {
				compressionResults.errors.push('Gzip round-trip failed');
			}
		} catch (e) {
			compressionResults.errors.push('Gzip test error: ' + e.message);
		}
		
		// 测试 deflate/inflate
		try {
			compressionResults.totalTests++;
			const originalData = 'This is test data for deflate/inflate compression.';
			const compressed = zlib.deflate(originalData);
			const decompressed = zlib.inflate(compressed);
			
			if (decompressed === originalData) {
				compressionResults.testsPassed++;
				console.log('✅ Deflate/Inflate test passed');
			} else {
				compressionResults.errors.push('Deflate round-trip failed');
			}
		} catch (e) {
			compressionResults.errors.push('Deflate test error: ' + e.message);
		}
		
		// 测试空字符串
		try {
			compressionResults.totalTests++;
			const emptyData = '';
			const compressed = zlib.gzip(emptyData);
			const decompressed = zlib.gunzip(compressed);
			
			if (decompressed === emptyData) {
				compressionResults.testsPassed++;
				console.log('✅ Empty string compression test passed');
			} else {
				compressionResults.errors.push('Empty string compression failed');
			}
		} catch (e) {
			compressionResults.errors.push('Empty string test error: ' + e.message);
		}
		
		// 测试大数据
		try {
			compressionResults.totalTests++;
			const largeData = 'A'.repeat(10000);
			const compressed = zlib.gzip(largeData);
			const decompressed = zlib.gunzip(compressed);
			
			if (decompressed === largeData && compressed.length < largeData.length) {
				compressionResults.testsPassed++;
				console.log('✅ Large data compression test passed');
				console.log('Original size:', largeData.length, 'Compressed size:', compressed.length);
			} else {
				compressionResults.errors.push('Large data compression failed');
			}
		} catch (e) {
			compressionResults.errors.push('Large data test error: ' + e.message);
		}
		
		global.compressionResults = compressionResults;
		console.log('Compression test results:', compressionResults);
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run compression test: %v", err)
	}

	results := runner.GetValue("compressionResults")
	if results == nil {
		t.Fatal("Compression results not found")
	}

	resultsObj := results.(*goja.Object)
	passed := resultsObj.Get("testsPassed").ToInteger()
	total := resultsObj.Get("totalTests").ToInteger()

	if passed != total {
		errors := resultsObj.Get("errors")
		t.Fatalf("Compression tests failed: %d/%d passed. Errors: %v", passed, total, errors)
	}

	t.Logf("All compression tests passed: %d/%d", passed, total)
}

// TestFixedFileSystemModule 测试修复后的文件系统模块
func TestFixedFileSystemModule(t *testing.T) {
	runner := runtime.New()

	code := `
		const fs = require('fs');
		
		let fsResults = {
			testsPassed: 0,
			totalTests: 0,
			errors: []
		};
		
		// 测试 existsSync
		try {
			fsResults.totalTests++;
			const currentDirExists = fs.existsSync('.');
			const nonExistentExists = fs.existsSync('non-existent-file-12345');
			
			if (currentDirExists === true && nonExistentExists === false) {
				fsResults.testsPassed++;
				console.log('✅ existsSync test passed');
			} else {
				fsResults.errors.push('existsSync test failed');
			}
		} catch (e) {
			fsResults.errors.push('existsSync error: ' + e.message);
		}
		
		// 测试异步 exists
		fsResults.totalTests++;
		fs.exists('.')
			.then((exists) => {
				if (exists === true) {
					fsResults.testsPassed++;
					console.log('✅ async exists test passed');
				} else {
					fsResults.errors.push('async exists test failed');
				}
				
				// 测试不存在的文件
				return fs.exists('non-existent-file-12345');
			})
			.then((exists) => {
				fsResults.totalTests++;
				if (exists === false) {
					fsResults.testsPassed++;
					console.log('✅ async exists (non-existent) test passed');
				} else {
					fsResults.errors.push('async exists (non-existent) test failed');
				}
			})
			.catch((e) => {
				fsResults.errors.push('async exists error: ' + e.message);
			});
		
		global.fsResults = fsResults;
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run fs test: %v", err)
	}

	// 等待异步操作完成
	time.Sleep(100 * time.Millisecond)

	results := runner.GetValue("fsResults")
	if results == nil {
		t.Fatal("FS results not found")
	}

	resultsObj := results.(*goja.Object)
	passed := resultsObj.Get("testsPassed").ToInteger()
	total := resultsObj.Get("totalTests").ToInteger()

	if passed != total {
		errors := resultsObj.Get("errors")
		t.Fatalf("FS tests failed: %d/%d passed. Errors: %v", passed, total, errors)
	}

	t.Logf("All FS tests passed: %d/%d", passed, total)
}

// TestAllFixesIntegration 综合测试所有修复
func TestAllFixesIntegration(t *testing.T) {
	runner := runtime.New()

	code := `
		console.log('🚀 Starting comprehensive fixes integration test...');
		
		let integrationResults = {
			eventLoopFixed: false,
			compressionFixed: false,
			fileSystemFixed: false,
			allTestsPassed: false
		};
		
		// 测试事件循环修复
		let intervalExecuted = false;
		const intervalId = setInterval(() => {
			intervalExecuted = true;
			clearInterval(intervalId);
			integrationResults.eventLoopFixed = true;
			console.log('✅ Event loop interval fix verified');
			checkAllComplete();
		}, 30);
		
		// 测试压缩模块修复
		try {
			const zlib = require('zlib');
			const testData = 'Integration test data for compression';
			const compressed = zlib.gzip(testData);
			const decompressed = zlib.gunzip(compressed);
			
			if (decompressed === testData) {
				integrationResults.compressionFixed = true;
				console.log('✅ Compression module fix verified');
				checkAllComplete();
			}
		} catch (e) {
			console.error('❌ Compression test failed:', e.message);
		}
		
		// 测试文件系统修复
		try {
			const fs = require('fs');
			if (typeof fs.exists === 'function' && typeof fs.existsSync === 'function') {
				const exists = fs.existsSync('.');
				if (exists === true) {
					integrationResults.fileSystemFixed = true;
					console.log('✅ File system module fix verified');
					checkAllComplete();
				}
			}
		} catch (e) {
			console.error('❌ File system test failed:', e.message);
		}
		
		function checkAllComplete() {
			if (integrationResults.eventLoopFixed && 
				integrationResults.compressionFixed && 
				integrationResults.fileSystemFixed) {
				integrationResults.allTestsPassed = true;
				console.log('🎉 All fixes verified successfully!');
			}
		}
		
		global.integrationResults = integrationResults;
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run integration test: %v", err)
	}

	// 等待所有异步操作完成
	time.Sleep(150 * time.Millisecond)

	results := runner.GetValue("integrationResults")
	if results == nil {
		t.Fatal("Integration results not found")
	}

	resultsObj := results.(*goja.Object)
	allPassed := resultsObj.Get("allTestsPassed").ToBoolean()

	if !allPassed {
		eventLoop := resultsObj.Get("eventLoopFixed").ToBoolean()
		compression := resultsObj.Get("compressionFixed").ToBoolean()
		fileSystem := resultsObj.Get("fileSystemFixed").ToBoolean()

		t.Fatalf("Integration test failed. Event Loop: %v, Compression: %v, File System: %v",
			eventLoop, compression, fileSystem)
	}

	t.Log("🎉 All fixes integration test passed successfully!")
}
