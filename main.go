package main

import (
	"fmt"
	"os"
	"sw_runtime/internal/runtime"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: sw_runtime <script.js|script.ts>")
		fmt.Println("示例: sw_runtime examples/calculator-app.ts")
		os.Exit(1)
	}

	scriptPath := os.Args[1]

	// 创建运行器
	runner := runtime.New()

	// 运行脚本
	err := runner.RunFile(scriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "运行脚本失败: %v\n", err)
		os.Exit(1)
	}
}
