package test

import (
	"strings"
	"testing"
	"time"

	"sw_runtime/internal/runtime"
)

func TestRunnerBasicFunctionality(t *testing.T) {
	runner := runtime.NewOrPanic()

	// 测试基本 JavaScript 代码执行
	code := `
		let result = 1 + 2;
		console.log("Basic test:", result);
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run basic code: %v", err)
	}
}

func TestRunnerTypeScriptSupport(t *testing.T) {
	runner := runtime.NewOrPanic()

	// 测试 TypeScript 代码执行
	tsCode := `
		interface User {
			name: string;
			age: number;
		}

		const user: User = {
			name: "Alice",
			age: 30
		};

		console.log("TypeScript test:", user.name, user.age);
	`

	err := runner.RunCode(tsCode)
	if err != nil {
		t.Fatalf("Failed to run TypeScript code: %v", err)
	}
}

func TestRunnerConsoleOutput(t *testing.T) {
	runner := runtime.NewOrPanic()

	// 测试 console 对象的各种方法
	code := `
		console.log("Log message");
		console.error("Error message");
		console.warn("Warning message");
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run console test: %v", err)
	}
}

func TestRunnerGlobalVariables(t *testing.T) {
	runner := runtime.NewOrPanic()

	// 设置全局变量
	runner.SetValue("testVar", "Hello World")

	code := `
		console.log("Global variable:", testVar);
		globalResult = testVar + " from JS";
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run global variable test: %v", err)
	}

	// 获取 JavaScript 中设置的变量
	result := runner.GetValue("globalResult")
	if result == nil {
		t.Fatal("Failed to get global variable from JS")
	}

	resultStr := result.String()
	expected := "Hello World from JS"
	if resultStr != expected {
		t.Fatalf("Expected %s, got %s", expected, resultStr)
	}
}

func TestRunnerAsyncOperations(t *testing.T) {
	runner := runtime.NewOrPanic()

	// 测试 setTimeout
	code := `
		let completed = false;
		setTimeout(() => {
			completed = true;
			console.log("Timeout completed");
		}, 100);
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run async test: %v", err)
	}

	// 等待一段时间确保异步操作完成
	time.Sleep(200 * time.Millisecond)

	completed := runner.GetValue("completed")
	if !completed.ToBoolean() {
		t.Fatal("Async operation did not complete")
	}
}

func TestRunnerPromiseSupport(t *testing.T) {
	runner := runtime.NewOrPanic()

	// 测试 Promise
	code := `
		let promiseResult = null;
		
		const promise = new Promise((resolve) => {
			setTimeout(() => {
				resolve("Promise resolved");
			}, 50);
		});

		promise.then((result) => {
			promiseResult = result;
			console.log("Promise result:", result);
		});
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run promise test: %v", err)
	}

	// 等待 Promise 完成
	time.Sleep(100 * time.Millisecond)

	result := runner.GetValue("promiseResult")
	if result == nil || result.String() != "Promise resolved" {
		t.Fatal("Promise was not resolved correctly")
	}
}

func TestRunnerModuleSystem(t *testing.T) {
	runner := runtime.NewOrPanic()

	// 测试内置模块
	code := `
		const path = require('path');
		const result = path.join('test', 'file.txt');
		console.log("Path join result:", result);
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run module test: %v", err)
	}
}

func TestRunnerErrorHandling(t *testing.T) {
	runner := runtime.NewOrPanic()

	// 测试语法错误
	invalidCode := `
		let x = ;
	`

	err := runner.RunCode(invalidCode)
	if err == nil {
		t.Fatal("Expected error for invalid syntax, but got none")
	}

	// 测试运行时错误
	runtimeErrorCode := `
		throw new Error("Test runtime error");
	`

	err = runner.RunCode(runtimeErrorCode)
	if err == nil {
		t.Fatal("Expected runtime error, but got none")
	}

	if !strings.Contains(err.Error(), "Test runtime error") {
		t.Fatalf("Expected error message to contain 'Test runtime error', got: %v", err)
	}
}

func TestRunnerModuleCache(t *testing.T) {
	runner := runtime.NewOrPanic()

	// 测试模块缓存功能
	loadedModules := runner.GetLoadedModules()
	initialCount := len(loadedModules)

	// 加载一个模块
	code := `
		const path = require('path');
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// 检查模块是否被缓存
	loadedModules = runner.GetLoadedModules()
	if len(loadedModules) <= initialCount {
		t.Fatal("Module was not cached")
	}

	// 清除缓存
	runner.ClearModuleCache()
	loadedModules = runner.GetLoadedModules()
	if len(loadedModules) != 0 {
		t.Fatal("Module cache was not cleared")
	}
}

func TestRunnerBuiltinModules(t *testing.T) {
	runner := runtime.NewOrPanic()

	// 获取内置模块列表
	builtinModules := runner.GetBuiltinModules()
	if len(builtinModules) == 0 {
		t.Fatal("No builtin modules found")
	}

	// 检查是否包含预期的内置模块
	expectedModules := []string{"path", "fs", "crypto", "http"}
	for _, expected := range expectedModules {
		found := false
		for _, builtin := range builtinModules {
			if builtin == expected {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Expected builtin module '%s' not found", expected)
		}
	}
}
