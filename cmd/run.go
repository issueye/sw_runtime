package cmd

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"os"
	"regexp"
	"sw_runtime/internal/runtime"

	"github.com/spf13/cobra"
)

var (
	clearCache     bool
	decryptKey     string
	decryptKeyFile string
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
  sw_runtime run --clear-cache app.ts
  sw_runtime run --decrypt-key=<key> encrypted.bundle.js
  sw_runtime run --decrypt-key-file=bundle.key encrypted.bundle.js`,
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
		runner := runtime.NewOrPanic()
		defer runner.Close()

		// 如果需要清除缓存
		if clearCache {
			runner.ClearModuleCache()
			if verbose && !quiet {
				fmt.Println("🧹 已清除模块缓存")
			}
		}

		// 处理加密文件
		var actualScriptPath = scriptPath
		if decryptKey != "" || decryptKeyFile != "" {
			// 读取密钥
			key := decryptKey
			if decryptKeyFile != "" {
				keyData, err := os.ReadFile(decryptKeyFile)
				if err != nil {
					fmt.Fprintf(os.Stderr, "❌ 读取密钥文件失败: %v\n", err)
					os.Exit(1)
				}
				key = string(keyData)
			}

			if verbose && !quiet {
				fmt.Println("🔓 正在解密文件...")
			}

			// 解密文件
			decryptedPath, err := decryptBundleFile(scriptPath, key)
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ 解密失败: %v\n", err)
				os.Exit(1)
			}
			actualScriptPath = decryptedPath
			defer os.Remove(decryptedPath) // 运行后删除临时文件

			if verbose && !quiet {
				fmt.Println("✅ 解密成功")
			}
		}

		// 运行脚本
		err := runner.RunFile(actualScriptPath)
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
	runCmd.Flags().StringVar(&decryptKey, "decrypt-key", "", "解密密钥（用于加密的 bundle 文件）")
	runCmd.Flags().StringVar(&decryptKeyFile, "decrypt-key-file", "", "解密密钥文件路径")
}

// decryptBundleFile 解密 bundle 文件
func decryptBundleFile(encryptedFile string, keyStr string) (string, error) {
	// 读取加密文件
	content, err := os.ReadFile(encryptedFile)
	if err != nil {
		return "", err
	}

	// 提取 ENCRYPTED_CODE 变量
	re := regexp.MustCompile(`const ENCRYPTED_CODE = "([^"]+)";`)
	matches := re.FindStringSubmatch(string(content))
	if len(matches) < 2 {
		return "", fmt.Errorf("文件不是加密的 bundle 文件")
	}

	encryptedCode := matches[1]

	// 解密
	decryptedCode, err := decryptCode(encryptedCode, keyStr)
	if err != nil {
		return "", err
	}

	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "sw_decrypted_*.js")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	_, err = tmpFile.WriteString(decryptedCode)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

// decryptCode 使用 AES-256-GCM 解密代码
func decryptCode(encryptedCode string, keyStr string) (string, error) {
	// 解码 base64 加密数据
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedCode)
	if err != nil {
		return "", fmt.Errorf("无效的加密数据: %w", err)
	}

	// 解码 base64 密钥
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return "", fmt.Errorf("无效的密钥格式: %w", err)
	}

	// 创建 AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// 创建 GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 提取 nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("加密数据太短")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败: %w", err)
	}

	return string(plaintext), nil
}
