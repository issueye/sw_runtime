package cmd

import (
	"fmt"
	"os"
	"sw_runtime/internal/runtime"

	"github.com/spf13/cobra"
)

// evalCmd 代表 eval 命令
var evalCmd = &cobra.Command{
	Use:   "eval <code>",
	Short: "执行 JavaScript 代码片段",
	Long: `执行一段 JavaScript 代码并输出结果。

示例:
  sw_runtime eval "console.log('Hello, World!')"
  sw_runtime eval "const x = 10; const y = 20; console.log(x + y)"
  sw_runtime eval "Promise.resolve(42).then(v => console.log(v))"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		code := args[0]

		verbose, _ := cmd.Flags().GetBool("verbose")
		quiet, _ := cmd.Flags().GetBool("quiet")

		if verbose && !quiet {
			fmt.Println("📝 执行代码:")
			fmt.Println(code)
			fmt.Println("---")
		}

		// 创建运行器
		runner := runtime.New()
		defer runner.Close()

		// 执行代码
		err := runner.RunCode(code)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ 执行失败: %v\n", err)
			os.Exit(1)
		}

		if verbose && !quiet {
			fmt.Println("---")
			fmt.Println("✅ 执行完成")
		}
	},
}

func init() {
	rootCmd.AddCommand(evalCmd)
}
