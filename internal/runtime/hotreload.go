package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// HotReloader 热加载管理器
type HotReloader struct {
	watcher    *fsnotify.Watcher
	watchPaths map[string]bool // 正在监控的路径
	callback   func()         // 文件变化时的回调函数
	done       chan struct{}  // 停止信号
}

// NewHotReloader 创建新的热加载管理器
func NewHotReloader(callback func()) (*HotReloader, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("创建文件监控器失败: %w", err)
	}

	return &HotReloader{
		watcher:    watcher,
		watchPaths: make(map[string]bool),
		callback:   callback,
		done:       make(chan struct{}),
	}, nil
}

// AddWatch 添加监控路径
func (hr *HotReloader) AddWatch(path string) error {
	// 获取绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %w", err)
	}

	// 检查是否是目录
	info, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("检查路径失败: %w", err)
	}

	if info.IsDir() {
		// 监控目录
		if err := hr.watcher.Add(absPath); err != nil {
			return fmt.Errorf("添加目录监控失败: %w", err)
		}
		hr.watchPaths[absPath] = true
	} else {
		// 监控文件所在目录
		dir := filepath.Dir(absPath)
		if err := hr.watcher.Add(dir); err != nil {
			return fmt.Errorf("添加文件目录监控失败: %w", err)
		}
		hr.watchPaths[dir] = true
	}

	return nil
}

// Start 启动热加载监控
func (hr *HotReloader) Start() {
	go hr.run()
}

// Stop 停止热加载监控
func (hr *HotReloader) Stop() {
	close(hr.done)
	hr.watcher.Close()
}

// run 运行文件监控循环
func (hr *HotReloader) run() {
	// 防抖计时器
	var debounceTimer *time.Timer
	var debounceDuration = 500 * time.Millisecond

	for {
		select {
		case <-hr.done:
			if debounceTimer != nil {
				debounceTimer.Stop()
			}
			return

		case event, ok := <-hr.watcher.Events:
			if !ok {
				return
			}

			// 只处理写入、创建、重命名事件
			if event.Op&fsnotify.Write == fsnotify.Write ||
				event.Op&fsnotify.Create == fsnotify.Create ||
				event.Op&fsnotify.Rename == fsnotify.Rename {

				// 检查是否是监控的文件类型
				ext := filepath.Ext(event.Name)
				if ext == ".js" || ext == ".ts" || ext == ".tsx" || ext == ".json" {
					// 防抖处理：延迟执行回调
					if debounceTimer != nil {
						debounceTimer.Stop()
					}

					debounceTimer = time.AfterFunc(debounceDuration, func() {
						fmt.Printf("🔄 检测到文件变化: %s\n", event.Name)
						hr.callback()
					})
				}
			}

		case err, ok := <-hr.watcher.Errors:
			if !ok {
				return
			}
			fmt.Fprintf(os.Stderr, "文件监控错误: %v\n", err)
		}
	}
}