package runtime

import (
	"context"
	"os"
	"os/signal"
	"sw_runtime/internal/builtins"
	"sw_runtime/internal/pool"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dop251/goja"
)

// SimpleEventLoop 简化的事件循环实现
type SimpleEventLoop struct {
	vm           *goja.Runtime
	timers       map[int]*timerEntry    // 定时器条目
	intervals    map[int]*intervalEntry // 间隔定时器条目
	timerID      atomic.Int64           // 原子计数器,避免锁竞争
	mu           sync.RWMutex           // 读写锁,提升并发性能
	vmMu         sync.Mutex             // 保护 goja.Runtime 并发访问(非线程安全)
	running      atomic.Bool            // 原子布尔值,避免锁
	activeJobs   atomic.Int32           // 活跃的异步任务计数
	stopChan     chan struct{}          // 停止信号通道
	hasLongLived atomic.Bool            // 是否有长期运行的任务(如 HTTP 服务器)
	ctx          context.Context        // 上下文控制
	cancel       context.CancelFunc     // 取消函数
}

// timerEntry 定时器条目
type timerEntry struct {
	timer    *time.Timer
	canceled atomic.Bool // 是否已取消
}

// intervalEntry 间隔定时器条目
type intervalEntry struct {
	ticker   *time.Ticker
	ctx      context.Context
	cancel   context.CancelFunc
	canceled atomic.Bool // 是否已取消
}

// NewSimpleEventLoop 创建简化的事件循环
func NewSimpleEventLoop(vm *goja.Runtime) *SimpleEventLoop {
	ctx, cancel := context.WithCancel(context.Background())
	el := &SimpleEventLoop{
		vm:        vm,
		timers:    make(map[int]*timerEntry, 64),    // 增大初始容量,减少扩容开销
		intervals: make(map[int]*intervalEntry, 32), // 增大初始容量
		stopChan:  make(chan struct{}, 1),           // 带缓冲,避免阻塞
		ctx:       ctx,
		cancel:    cancel,
	}
	return el
}

// Start 启动事件循环
func (el *SimpleEventLoop) Start() {
	el.running.Store(true)
}

// Stop 停止事件循环
func (el *SimpleEventLoop) Stop() {
	if !el.running.CompareAndSwap(true, false) {
		return // 已经停止
	}

	// 取消上下文,通知所有goroutine退出
	el.cancel()

	el.mu.Lock()
	// 停止所有定时器
	for id, entry := range el.timers {
		entry.canceled.Store(true)
		entry.timer.Stop()
		delete(el.timers, id)
		pool.GlobalMemoryMonitor.DecrementTimerCount()
	}

	// 停止所有间隔定时器
	for id, entry := range el.intervals {
		entry.canceled.Store(true)
		entry.cancel()
		entry.ticker.Stop()
		delete(el.intervals, id)
	}
	el.mu.Unlock()

	// 发送停止信号(非阻塞)
	select {
	case el.stopChan <- struct{}{}:
	default:
	}
}

// AddJob 增加活跃任务计数
func (el *SimpleEventLoop) AddJob() {
	el.activeJobs.Add(1)
}

// DoneJob 减少活跃任务计数
func (el *SimpleEventLoop) DoneJob() {
	el.activeJobs.Add(-1)
}

// SetLongLived 标记有长期运行的任务
func (el *SimpleEventLoop) SetLongLived() {
	el.hasLongLived.Store(true)
}

// WaitAndProcess 等待并处理所有任务
func (el *SimpleEventLoop) WaitAndProcess() {
	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// 给异步任务一些启动时间
	time.Sleep(10 * time.Millisecond)

	// 使用更高频率的 ticker 提升响应速度
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()

	// 空闲检测计数器 - 优化退出条件
	idleCount := 0
	const maxIdleChecks = 3 // 5ms × 3 = 15ms 无任务后快速退出

	for el.running.Load() {
		select {
		case <-sigChan:
			// 收到终止信号,优雅退出
			el.Stop()
			return
		case <-el.stopChan:
			// 收到停止信号
			return
		case <-el.ctx.Done():
			// 上下文取消
			return
		case <-ticker.C:
			// 先检查长期运行任务(最快路径)
			isLongLived := el.hasLongLived.Load()
			if isLongLived {
				idleCount = 0
				continue
			}

			// 再检查 HTTP 服务器
			hasHTTPServer := builtins.IsHTTPServerRunning()
			if hasHTTPServer {
				idleCount = 0
				continue
			}

			// 检查活跃任务(无锁操作)
			activeJobs := el.activeJobs.Load()
			if activeJobs > 0 {
				idleCount = 0
				continue
			}

			// 最后才检查定时器(需要锁)
			el.mu.RLock()
			hasTimers := len(el.timers) > 0
			hasIntervals := len(el.intervals) > 0
			el.mu.RUnlock()

			if hasTimers || hasIntervals {
				idleCount = 0
				continue
			}

			// 没有任何活跃任务,增加空闲计数
			idleCount++
			if idleCount >= maxIdleChecks {
				// 确实没有任务了,快速退出
				return
			}
		}
	}
}

// SetTimeout 实现 setTimeout
func (el *SimpleEventLoop) SetTimeout(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.Undefined()
	}

	fn, ok := goja.AssertFunction(call.Arguments[0])
	if !ok {
		return goja.Undefined()
	}

	delay := int64(0)
	if len(call.Arguments) >= 2 {
		delay = call.Arguments[1].ToInteger()
	}
	if delay < 0 {
		delay = 0
	}

	// 使用原子操作生成ID,避免锁
	id := int(el.timerID.Add(1))

	entry := &timerEntry{}
	entry.timer = time.AfterFunc(time.Duration(delay)*time.Millisecond, func() {
		// 快速路径：检查是否已取消
		if entry.canceled.Load() {
			return
		}

		// 优化：减少 defer 开销，直接执行清理
		// 先执行回调（需要加锁保护 goja.Runtime）
		el.vmMu.Lock()
		func() {
			defer func() {
				if r := recover(); r != nil {
					// 记录错误但不影响清理
				}
			}()
			fn(goja.Undefined())
		}()
		el.vmMu.Unlock()

		// 然后清理资源
		el.mu.Lock()
		delete(el.timers, id)
		el.mu.Unlock()
		pool.GlobalMemoryMonitor.DecrementTimerCount()
	})

	el.mu.Lock()
	el.timers[id] = entry
	el.mu.Unlock()

	// 增加定时器计数
	pool.GlobalMemoryMonitor.IncrementTimerCount()

	return el.vm.ToValue(id)
}

// ClearTimeout 实现 clearTimeout
func (el *SimpleEventLoop) ClearTimeout(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.Undefined()
	}

	id := int(call.Arguments[0].ToInteger())

	el.mu.Lock()
	entry, ok := el.timers[id]
	if ok {
		entry.canceled.Store(true)
		entry.timer.Stop()
		delete(el.timers, id)
		el.mu.Unlock()
		// 减少定时器计数
		pool.GlobalMemoryMonitor.DecrementTimerCount()
	} else {
		el.mu.Unlock()
	}

	return goja.Undefined()
}

// SetInterval 实现 setInterval
func (el *SimpleEventLoop) SetInterval(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.Undefined()
	}

	fn, ok := goja.AssertFunction(call.Arguments[0])
	if !ok {
		return goja.Undefined()
	}

	interval := int64(100) // 默认 100ms
	if len(call.Arguments) >= 2 {
		interval = call.Arguments[1].ToInteger()
	}
	if interval < 1 {
		interval = 1
	}

	// 使用原子操作生成ID
	id := int(el.timerID.Add(1))

	// 创建间隔定时器条目
	ctx, cancel := context.WithCancel(el.ctx)
	entry := &intervalEntry{
		ticker: time.NewTicker(time.Duration(interval) * time.Millisecond),
		ctx:    ctx,
		cancel: cancel,
	}

	el.mu.Lock()
	el.intervals[id] = entry
	el.mu.Unlock()

	// 启动间隔执行的 goroutine
	go func() {
		defer func() {
			entry.ticker.Stop()
			el.mu.Lock()
			delete(el.intervals, id)
			el.mu.Unlock()
		}()

		for {
			select {
			case <-entry.ctx.Done():
				// 上下文取消,退出
				return
			case <-entry.ticker.C:
				// 快速路径：检查是否已取消
				if entry.canceled.Load() {
					return
				}

				// 优化：内联 panic 处理，减少函数调用开销（需要加锁保护 goja.Runtime）
				el.vmMu.Lock()
				func() {
					defer func() {
						if r := recover(); r != nil {
							// 记录错误但继续运行
						}
					}()
					fn(goja.Undefined())
				}()
				el.vmMu.Unlock()
			}
		}
	}()

	return el.vm.ToValue(id)
}

// ClearInterval 实现 clearInterval
func (el *SimpleEventLoop) ClearInterval(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.Undefined()
	}

	id := int(call.Arguments[0].ToInteger())

	el.mu.Lock()
	entry, ok := el.intervals[id]
	if ok {
		entry.canceled.Store(true)
		entry.cancel() // 取消上下文,通知 goroutine 退出
		entry.ticker.Stop()
		delete(el.intervals, id)
	}
	el.mu.Unlock()

	return goja.Undefined()
}
