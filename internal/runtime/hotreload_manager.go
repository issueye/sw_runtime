package runtime

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// RunnerManager 运行器管理器，支持热重载
type RunnerManager struct {
	scriptPath     string
	workingDir     string
	clearCache     bool
	decryptKey     string
	decryptKeyFile string
	verbose        bool
	quiet          bool

	currentRunner *Runner
	restarting    bool
	mu            sync.RWMutex
	reloader      *HotReloader
	restartChan   chan struct{}
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// NewRunnerManager 创建新的运行器管理器
func NewRunnerManager(scriptPath, workingDir string, clearCache bool,
	decryptKey, decryptKeyFile string, verbose, quiet bool) *RunnerManager {

	return &RunnerManager{
		scriptPath:     scriptPath,
		workingDir:     workingDir,
		clearCache:     clearCache,
		decryptKey:     decryptKey,
		decryptKeyFile: decryptKeyFile,
		verbose:        verbose,
		quiet:          quiet,
		restartChan:    make(chan struct{}, 1),
		stopChan:       make(chan struct{}),
	}
}

// Start 启动运行器管理器
func (rm *RunnerManager) Start() error {
	// 处理中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	rm.wg.Add(1)
	go rm.runLoop(sigChan)

	// 等待运行结束
	rm.wg.Wait()
	return nil
}

// Stop 停止运行器管理器
func (rm *RunnerManager) Stop() {
	close(rm.stopChan)
	rm.wg.Wait()
}

// runLoop 运行主循环
func (rm *RunnerManager) runLoop(sigChan chan os.Signal) {
	defer rm.wg.Done()

	// 首次创建运行器
	if err := rm.createAndRunRunner(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ 运行失败: %v\n", err)
		return
	}

	for {
		select {
		case <-rm.stopChan:
			rm.stopCurrentRunner()
			return

		case sig := <-sigChan:
			if rm.verbose && !rm.quiet {
				fmt.Printf("\n📭 接收到信号 %v，正在停止...\n", sig)
			}
			rm.stopCurrentRunner()
			return

		case <-rm.restartChan:
			if rm.verbose && !rm.quiet {
				fmt.Println("\n🔄 检测到文件变化，重新加载...")
			}
			rm.stopCurrentRunner()

			// 重新创建运行器
			if err := rm.createAndRunRunner(); err != nil {
				fmt.Fprintf(os.Stderr, "❌ 重新加载失败: %v\n", err)
				return
			}
		}
	}
}

// createAndRunRunner 创建并运行运行器
func (rm *RunnerManager) createAndRunRunner() error {
	// 创建运行器
	var runner *Runner
	if rm.workingDir != "" {
		// 确保目录存在
		if _, err := os.Stat(rm.workingDir); os.IsNotExist(err) {
			return fmt.Errorf("工作目录不存在: %s", rm.workingDir)
		}
		runner = NewOrPanicWithWorkingDir(rm.workingDir)
	} else {
		runner = NewOrPanic()
	}

	// 设置当前运行器
	rm.mu.Lock()
	rm.currentRunner = runner
	rm.mu.Unlock()

	// 清理函数
	defer func() {
		rm.mu.Lock()
		if rm.currentRunner == runner {
			rm.currentRunner = nil
		}
		rm.mu.Unlock()
	}()

	// 如果需要清除缓存
	if rm.clearCache {
		runner.ClearModuleCache()
		if rm.verbose && !rm.quiet {
			fmt.Println("🧹 已清除模块缓存")
		}
	}

	// 处理加密文件
	actualScriptPath, cleanup, err := rm.resolveScriptPath()
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	// 如果启用监控模式，设置热加载
	if rm.reloader == nil {
		reloader, err := NewHotReloader(func() {
			// 设置重启标志
			rm.mu.Lock()
			rm.restarting = true
			rm.mu.Unlock()

			// 发送重启信号
			select {
			case rm.restartChan <- struct{}{}:
			default:
				// 通道已满，跳过
			}
		})
		if err != nil {
			return fmt.Errorf("创建热加载管理器失败: %w", err)
		}
		rm.reloader = reloader

		// 添加监控路径
		if err := rm.reloader.AddWatch(rm.scriptPath); err != nil {
			return fmt.Errorf("添加文件监控失败: %w", err)
		}

		// 启动监控
		rm.reloader.Start()
	}

	// 运行脚本
	if rm.verbose && !rm.quiet {
		fmt.Printf("🚀 正在运行: %s\n", rm.scriptPath)
		if rm.reloader != nil {
			fmt.Println("👀 正在监控文件变化... (按 Ctrl+C 退出)")
		}
	}

	err = runner.RunFile(actualScriptPath)
	if err != nil {
		// 检查是否是热重启导致的错误
		rm.mu.RLock()
		restarting := rm.restarting
		rm.mu.RUnlock()

		if restarting {
			// 重置重启标志
			rm.mu.Lock()
			rm.restarting = false
			rm.mu.Unlock()
			return nil // 正常重启，不返回错误
		}
		return fmt.Errorf("运行失败: %w", err)
	}

	if rm.verbose && !rm.quiet {
		fmt.Println("✅ 执行完成")
	}

	return nil
}

// resolveScriptPath 解析脚本路径，处理加密文件
func (rm *RunnerManager) resolveScriptPath() (string, func(), error) {
	// 监控模式不支持加密文件，已在调用处检查
	return rm.scriptPath, nil, nil
}

// stopCurrentRunner 停止当前运行器
func (rm *RunnerManager) stopCurrentRunner() {
	rm.mu.Lock()
	runner := rm.currentRunner
	rm.currentRunner = nil
	rm.restarting = false
	rm.mu.Unlock()

	if runner != nil {
		runner.Close()
	}

	if rm.reloader != nil {
		rm.reloader.Stop()
		rm.reloader = nil
	}
}