package cmd

import (
	"fmt"
	"os"
	"sw_runtime/internal/runtime"

	"github.com/spf13/cobra"
)

var (
	clearCache bool
)

// runCmd 代表 run 命令
var runCmd = &cobra.Command{
	Use:   "run <script>",
	Short: "运行 JavaScript 或 TypeScript 脚本",
	Long: `运行指定的 JavaScript 或 TypeScript 脚本文件。

支持的文件类型:
  • .js  - JavaScript 文件
  • .ts  - TypeScript 文件 (自动编译)
  • .tsx - TypeScript JSX 文件

示例:
  sw_runtime run app.ts
  sw_runtime run server.js
  sw_runtime run --clear-cache app.ts`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		scriptPath := args[0]

		// 检查文件是否存在
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "❌ 文件不存在: %s\n", scriptPath)
			os.Exit(1)
		}

		verbose, _ := cmd.Flags().GetBool("verbose")
		quiet, _ := cmd.Flags().GetBool("quiet")

		if verbose && !quiet {
			fmt.Printf("🚀 正在运行: %s\n", scriptPath)
		}

		// 创建运行器
		runner := runtime.New()
		defer runner.Close()

		// 如果需要清除缓存
		if clearCache {
			runner.ClearModuleCache()
			if verbose && !quiet {
				fmt.Println("🧹 已清除模块缓存")
			}
		}

		// 运行脚本
		err := runner.RunFile(scriptPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ 运行失败: %v\n", err)
			os.Exit(1)
		}

		if verbose && !quiet {
			fmt.Println("✅ 执行完成")
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// 本地标志
	runCmd.Flags().BoolVarP(&clearCache, "clear-cache", "c", false, "运行前清除模块缓存")
}
