package cmd

import (
	"fmt"
	"sw_runtime/internal/runtime"

	"github.com/spf13/cobra"
)

// infoCmd 代表 info 命令
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "显示运行时信息",
	Long:  `显示 SW Runtime 支持的内置模块和功能信息。`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("╔════════════════════════════════════════════════════════╗")
		fmt.Println("║           SW Runtime 运行时信息                        ║")
		fmt.Println("╚════════════════════════════════════════════════════════╝")
		fmt.Println()

		// 创建临时运行器获取模块信息
		runner := runtime.New()
		defer runner.Close()

		// 内置模块
		fmt.Println("📦 内置模块:")
		builtinModules := runner.GetBuiltinModules()
		for _, module := range builtinModules {
			fmt.Printf("   • %s\n", module)
		}
		fmt.Println()

		// 功能特性
		fmt.Println("✨ 核心功能:")
		features := []string{
			"TypeScript 支持",
			"ES6+ 语法",
			"Promise/async-await",
			"定时器 (setTimeout/setInterval)",
			"HTTP 服务器",
			"WebSocket 支持",
			"文件系统操作",
			"加密和压缩",
			"SQLite 数据库",
			"Redis 客户端",
			"命令执行",
		}
		for _, feature := range features {
			fmt.Printf("   ✓ %s\n", feature)
		}
		fmt.Println()

		// 使用示例
		fmt.Println("💡 快速开始:")
		fmt.Println("   sw_runtime run app.ts        运行脚本")
		fmt.Println("   sw_runtime eval \"code\"       执行代码")
		fmt.Println("   sw_runtime version           查看版本")
		fmt.Println("   sw_runtime --help            查看帮助")
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
