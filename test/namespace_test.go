package test

import (
	"fmt"
	"sw_runtime/internal/runtime"
	"testing"
)

func TestNamespaceModules(t *testing.T) {
	fmt.Println("=== 测试命名空间模块加载 ===")

	r, err := runtime.New()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	// 测试 1: 加载 http 命名空间
	fmt.Println("测试 1: 加载 http 命名空间")
	code1 := `
		const http = require('http');
		console.log('http.client 存在:', typeof http.client !== 'undefined');
		console.log('http.server 存在:', typeof http.server !== 'undefined');
	`
	err = r.RunCode(code1)
	if err != nil {
		fmt.Println("  失败:", err)
	} else {
		fmt.Println("  通过")
	}

	// 测试 2: 加载 http/client 子模块
	fmt.Println("测试 2: 加载 http/client 子模块")
	code2 := `
		const client = require('http/client');
		console.log('http/client 模块方法:', Object.keys(client).filter(k => typeof client[k] === 'function').slice(0, 5));
	`
	err = r.RunCode(code2)
	if err != nil {
		fmt.Println("  失败:", err)
	} else {
		fmt.Println("  通过")
	}

	// 测试 3: 加载 http/server 子模块
	fmt.Println("测试 3: 加载 http/server 子模块")
	code3 := `
		const server = require('http/server');
		console.log('http/server 模块方法:', Object.keys(server).filter(k => typeof server[k] === 'function').slice(0, 5));
	`
	err = r.RunCode(code3)
	if err != nil {
		fmt.Println("  失败:", err)
	} else {
		fmt.Println("  通过")
	}

	// 测试 4: 向后兼容 - 旧版 httpserver 仍可用
	fmt.Println("测试 4: 向后兼容 - 旧版 httpserver")
	code4 := `
		const oldServer = require('httpserver');
		console.log('httpserver 模块方法:', Object.keys(oldServer).filter(k => typeof oldServer[k] === 'function').slice(0, 3));
	`
	err = r.RunCode(code4)
	if err != nil {
		fmt.Println("  失败:", err)
	} else {
		fmt.Println("  通过")
	}

	fmt.Println("=== 所有测试完成 ===")
}
