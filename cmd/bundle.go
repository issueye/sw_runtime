package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sw_runtime/internal/bundler"

	"github.com/spf13/cobra"
)

var (
	outputFile   string
	minify       bool
	sourcemap    bool
	excludeFiles []string
	encrypt      bool
	encryptKey   string
)

var bundleCmd = &cobra.Command{
	Use:   "bundle <entry-file>",
	Short: "将多个脚本打包成单个文件",
	Long: `将 JavaScript/TypeScript 项目打包成单个可执行文件

bundle 命令会从入口文件开始，递归分析所有依赖的模块，并将它们
打包成一个独立的脚本文件。

特性:
  • 自动解析 require() 依赖
  • 支持 TypeScript 自动编译
  • 排除内置模块
  • 可选代码压缩
  • 支持 Source Map
  • 代码加密保护 (AES-256-GCM)

示例:
  sw_runtime bundle app.ts -o bundle.js
  sw_runtime bundle main.js -o dist/app.js --minify
  sw_runtime bundle server.ts -o server.bundle.js --exclude utils.js,helpers.js
  sw_runtime bundle app.js --encrypt -o app.encrypted.js`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		entryFile := args[0]

		// 检查入口文件是否存在
		if _, err := os.Stat(entryFile); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "❌ 入口文件不存在: %s\n", entryFile)
			os.Exit(1)
		}

		// 如果没有指定输出文件，自动生成
		if outputFile == "" {
			ext := filepath.Ext(entryFile)
			base := strings.TrimSuffix(entryFile, ext)
			outputFile = base + ".bundle.js"
		}

		// 创建打包器
		b := bundler.New(bundler.Options{
			EntryFile:    entryFile,
			OutputFile:   outputFile,
			Minify:       minify,
			Sourcemap:    sourcemap,
			ExcludeFiles: excludeFiles,
			Encrypt:      encrypt,
			EncryptKey:   encryptKey,
		})

		// 执行打包
		quietMode, _ := cmd.Flags().GetBool("quiet")
		if !quietMode {
			fmt.Printf("📦 正在打包: %s\n", entryFile)
		}

		result, err := b.Bundle()
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ 打包失败: %v\n", err)
			os.Exit(1)
		}

		// 写入输出文件
		err = os.WriteFile(outputFile, []byte(result.Code), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ 写入文件失败: %v\n", err)
			os.Exit(1)
		}

		// 如果需要 sourcemap，写入 map 文件
		if sourcemap && result.Sourcemap != "" {
			mapFile := outputFile + ".map"
			err = os.WriteFile(mapFile, []byte(result.Sourcemap), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  写入 sourcemap 失败: %v\n", err)
			}
		}

		// 如果加密了，保存密钥到 .key 文件
		if result.Encrypted && result.EncryptKey != "" {
			keyFile := outputFile + ".key"
			err = os.WriteFile(keyFile, []byte(result.EncryptKey), 0600)
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  写入密钥文件失败: %v\n", err)
			}
		}

		// 显示结果
		if !quietMode {
			verboseMode, _ := cmd.Flags().GetBool("verbose")
			fmt.Printf("\n✅ 打包完成!\n\n")
			fmt.Printf("📄 输出文件: %s\n", outputFile)
			fmt.Printf("📊 文件大小: %.2f KB\n", float64(len(result.Code))/1024)
			fmt.Printf("📦 包含模块: %d 个\n", len(result.Modules))

			if verboseMode {
				fmt.Printf("\n包含的模块:\n")
				for _, mod := range result.Modules {
					fmt.Printf("  • %s\n", mod)
				}
			}

			if sourcemap {
				fmt.Printf("🗺️  Source Map: %s.map\n", outputFile)
			}

			// 显示加密信息
			if result.Encrypted {
				fmt.Printf("\n🔒 加密信息:\n")
				fmt.Printf("✅ 代码已加密 (AES-256-GCM)\n")
				fmt.Printf("🔑 密钥文件: %s.key\n", outputFile)
				fmt.Printf("📝 密钥内容: %s\n", result.EncryptKey)
				fmt.Printf("\n⚠️  请保管好密钥文件，运行时需要：\n")
				fmt.Printf("   sw_runtime run --decrypt-key=%s %s\n", result.EncryptKey, outputFile)
				fmt.Printf("   或\n")
				fmt.Printf("   sw_runtime run --decrypt-key-file=%s.key %s\n", outputFile, outputFile)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(bundleCmd)

	bundleCmd.Flags().StringVarP(&outputFile, "output", "o", "", "输出文件路径 (默认: <entry>.bundle.js)")
	bundleCmd.Flags().BoolVarP(&minify, "minify", "m", false, "压缩输出代码")
	bundleCmd.Flags().BoolVar(&sourcemap, "sourcemap", false, "生成 source map")
	bundleCmd.Flags().StringSliceVar(&excludeFiles, "exclude", []string{}, "排除指定文件（逗号分隔）")
	bundleCmd.Flags().BoolVar(&encrypt, "encrypt", false, "加密打包后的代码 (AES-256-GCM)")
	bundleCmd.Flags().StringVar(&encryptKey, "encrypt-key", "", "指定加密密钥（不指定则自动生成）")
}
