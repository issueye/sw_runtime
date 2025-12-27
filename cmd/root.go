package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
)

// rootCmd 代表基础命令
var rootCmd = &cobra.Command{
	Use:   "sw_runtime",
	Short: "SW Runtime - 高性能 JavaScript/TypeScript 运行时",
	Long: `SW Runtime 是一个基于 Go 的轻量级 JavaScript/TypeScript 运行时环境。

特性:
  • TypeScript 支持 - 自动编译 .ts 文件
  • 内置模块 - HTTP、WebSocket、文件系统、加密等
  • 异步支持 - Promise、setTimeout、setInterval
  • 高性能 - 优化的事件循环和并发处理
  • 零依赖 - 无需安装 Node.js

示例:
  sw_runtime run app.ts                   运行 TypeScript 脚本
  sw_runtime run app.js                   运行 JavaScript 脚本
  sw_runtime eval "console.log('Hello')"  执行 JavaScript 代码
  sw_runtime version                      显示版本信息`,
	Version: version,
}

// Execute 添加所有子命令到 root 命令并适当设置标志
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// 全局标志
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "详细输出模式")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "静默模式")
}
