package test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"sw_runtime/internal/runtime"
)

func TestModuleSystemBasicRequire(t *testing.T) {
	runner := runtime.NewOrPanic()

	// 测试内置模块加载
	code := `
		const path = require('utils/path');
		const fs = require('fs/fs');
		const crypto = require('utils/crypto');

		let moduleTestResults = {
			pathLoaded: typeof path === 'object' && path !== null,
			fsLoaded: typeof fs === 'object' && fs !== null,
			cryptoLoaded: typeof crypto === 'object' && crypto !== null
		};

		global.moduleTestResults = moduleTestResults;
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run module require test: %v", err)
	}

	results := runner.GetValue("moduleTestResults")
	if results == nil {
		t.Fatal("Module test results not found")
	}
}

func TestModuleSystemFileModule(t *testing.T) {
	// 创建临时测试模块文件
	tempDir := t.TempDir()
	moduleFile := filepath.Join(tempDir, "testmodule.js")

	moduleContent := `
		exports.hello = function(name) {
			return "Hello, " + name + "!";
		};

		exports.version = "1.0.0";

		exports.data = {
			items: [1, 2, 3],
			config: { debug: true }
		};
	`

	err := os.WriteFile(moduleFile, []byte(moduleContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test module file: %v", err)
	}

	// 切换到临时目录
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	runner := runtime.NewOrPanic()

	code := `
		const testModule = require('./testmodule.js');

		let fileModuleResults = {
			hasHello: typeof testModule.hello === 'function',
			hasVersion: typeof testModule.version === 'string',
			hasData: typeof testModule.data === 'object',
			helloResult: null,
			versionValue: testModule.version
		};

		if (fileModuleResults.hasHello) {
			fileModuleResults.helloResult = testModule.hello('World');
		}

		global.fileModuleResults = fileModuleResults;
	`

	err = runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run file module test: %v", err)
	}

	results := runner.GetValue("fileModuleResults")
	if results == nil {
		t.Fatal("File module test results not found")
	}
}

func TestModuleSystemTypeScriptModule(t *testing.T) {
	// 创建临时 TypeScript 模块文件
	tempDir := t.TempDir()
	moduleFile := filepath.Join(tempDir, "tsmodule.ts")

	moduleContent := `
		interface User {
			name: string;
			age: number;
		}

		export function createUser(name: string, age: number): User {
			return { name, age };
		}

		export const defaultUser: User = {
			name: "Anonymous",
			age: 0
		};

		export class UserManager {
			private users: User[] = [];

			addUser(user: User): void {
				this.users.push(user);
			}

			getUserCount(): number {
				return this.users.length;
			}
		}
	`

	err := os.WriteFile(moduleFile, []byte(moduleContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create TypeScript test module file: %v", err)
	}

	// 切换到临时目录
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	runner := runtime.NewOrPanic()

	code := `
		const tsModule = require('./tsmodule.ts');

		let tsModuleResults = {
			hasCreateUser: typeof tsModule.createUser === 'function',
			hasDefaultUser: typeof tsModule.defaultUser === 'object',
			hasUserManager: typeof tsModule.UserManager === 'function',
			userCreated: null,
			managerWorks: false
		};

		if (tsModuleResults.hasCreateUser) {
			tsModuleResults.userCreated = tsModule.createUser('Alice', 25);
		}

		if (tsModuleResults.hasUserManager) {
			const manager = new tsModule.UserManager();
			manager.addUser({ name: 'Bob', age: 30 });
			tsModuleResults.managerWorks = manager.getUserCount() === 1;
		}

		global.tsModuleResults = tsModuleResults;
	`

	err = runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run TypeScript module test: %v", err)
	}

	results := runner.GetValue("tsModuleResults")
	if results == nil {
		t.Fatal("TypeScript module test results not found")
	}
}

func TestModuleSystemCircularDependency(t *testing.T) {
	// 创建循环依赖的模块
	tempDir := t.TempDir()

	moduleA := filepath.Join(tempDir, "moduleA.js")
	moduleB := filepath.Join(tempDir, "moduleB.js")

	moduleAContent := `
		const moduleB = require('./moduleB.js');

		exports.name = 'Module A';
		exports.getB = function() {
			return moduleB.name;
		};
	`

	moduleBContent := `
		const moduleA = require('./moduleA.js');

		exports.name = 'Module B';
		exports.getA = function() {
			return moduleA.name;
		};
	`

	err := os.WriteFile(moduleA, []byte(moduleAContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create module A: %v", err)
	}

	err = os.WriteFile(moduleB, []byte(moduleBContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create module B: %v", err)
	}

	// 切换到临时目录
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	runner := runtime.NewOrPanic()

	code := `
		let circularResults = {
			error: null,
			moduleALoaded: false,
			moduleBLoaded: false
		};

		try {
			const moduleA = require('./moduleA.js');
			circularResults.moduleALoaded = typeof moduleA === 'object';

			const moduleB = require('./moduleB.js');
			circularResults.moduleBLoaded = typeof moduleB === 'object';
		} catch (e) {
			circularResults.error = e.message;
		}

		global.circularResults = circularResults;
	`

	err = runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run circular dependency test: %v", err)
	}

	results := runner.GetValue("circularResults")
	if results == nil {
		t.Fatal("Circular dependency test results not found")
	}
}

func TestModuleSystemCaching(t *testing.T) {
	// 创建测试模块
	tempDir := t.TempDir()
	moduleFile := filepath.Join(tempDir, "cached.js")

	moduleContent := `
		let loadCount = 0;
		loadCount++;

		exports.getLoadCount = function() {
			return loadCount;
		};

		exports.increment = function() {
			loadCount++;
		};
	`

	err := os.WriteFile(moduleFile, []byte(moduleContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create cached test module: %v", err)
	}

	// 切换到临时目录
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	runner := runtime.NewOrPanic()

	code := `
		// 第一次加载
		const cached1 = require('./cached.js');
		const firstLoadCount = cached1.getLoadCount();

		// 增加计数
		cached1.increment();

		// 第二次加载（应该使用缓存）
		const cached2 = require('./cached.js');
		const secondLoadCount = cached2.getLoadCount();

		let cachingResults = {
			firstLoadCount: firstLoadCount,
			secondLoadCount: secondLoadCount,
			sameInstance: cached1 === cached2
		};

		global.cachingResults = cachingResults;
	`

	err = runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run module caching test: %v", err)
	}

	results := runner.GetValue("cachingResults")
	if results == nil {
		t.Fatal("Module caching test results not found")
	}
}

func TestModuleSystemDynamicImport(t *testing.T) {
	// 创建测试模块
	tempDir := t.TempDir()
	moduleFile := filepath.Join(tempDir, "dynamic.js")

	moduleContent := `
		exports.message = "Dynamic import works!";
		exports.timestamp = Date.now();
	`

	err := os.WriteFile(moduleFile, []byte(moduleContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create dynamic import test module: %v", err)
	}

	// 切换到临时目录
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	runner := runtime.NewOrPanic()

	code := `
		let dynamicResults = {
			importSucceeded: false,
			message: null,
			error: null
		};

		// 使用 Promise 包装动态导入
		new Promise((resolve, reject) => {
			try {
				const module = require('./dynamic.js');
				resolve(module);
			} catch (error) {
				reject(error);
			}
		})
		.then((module) => {
			dynamicResults.importSucceeded = true;
			dynamicResults.message = module.message;
		})
		.catch((error) => {
			dynamicResults.error = error.message;
		});
	`

	err = runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run dynamic import test: %v", err)
	}

	// 等待 Promise 完成
	time.Sleep(100 * time.Millisecond)

	results := runner.GetValue("dynamicResults")
	if results == nil {
		t.Fatal("Dynamic import test results not found")
	}
}
