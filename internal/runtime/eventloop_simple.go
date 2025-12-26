package runtime

import (
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
	timers       map[int]*time.Timer
	intervals    map[int]*time.Ticker
	timerID      int
	mu           sync.Mutex
	running      bool
	activeJobs   int32         // 活跃的异步任务计数
	stopChan     chan struct{} // 停止信号通道
	hasLongLived bool          // 是否有长期运行的任务（如 HTTP 服务器）
}

// NewSimpleEventLoop 创建简化的事件循环
func NewSimpleEventLoop(vm *goja.Runtime) *SimpleEventLoop {
	return &SimpleEventLoop{
		vm:        vm,
		timers:    make(map[int]*time.Timer),
		intervals: make(map[int]*time.Ticker),
		stopChan:  make(chan struct{}),
	}
}

// Start 启动事件循环
func (el *SimpleEventLoop) Start() {
	el.running = true
}

// Stop 停止事件循环
func (el *SimpleEventLoop) Stop() {
	el.mu.Lock()
	defer el.mu.Unlock()

	for id, timer := range el.timers {
		timer.Stop()
		delete(el.timers, id)
	}

	for id, ticker := range el.intervals {
		ticker.Stop()
		delete(el.intervals, id)
	}

	el.running = false

	// 发送停止信号
	select {
	case el.stopChan <- struct{}{}:
	default:
	}
}

// AddJob 增加活跃任务计数
func (el *SimpleEventLoop) AddJob() {
	atomic.AddInt32(&el.activeJobs, 1)
}

// DoneJob 减少活跃任务计数
func (el *SimpleEventLoop) DoneJob() {
	atomic.AddInt32(&el.activeJobs, -1)
}

// SetLongLived 标记有长期运行的任务
func (el *SimpleEventLoop) SetLongLived() {
	el.hasLongLived = true
}

// WaitAndProcess 等待并处理所有任务
func (el *SimpleEventLoop) WaitAndProcess() {
	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 给异步任务一些启动时间
	time.Sleep(50 * time.Millisecond)

	for el.running {
		select {
		case <-sigChan:
			// 收到终止信号，优雅退出
			el.Stop()
			return
		case <-el.stopChan:
			// 收到停止信号
			return
		default:
			el.mu.Lock()
			hasTimers := len(el.timers) > 0
			hasIntervals := len(el.intervals) > 0
			el.mu.Unlock()

			activeJobs := atomic.LoadInt32(&el.activeJobs)

			// 检查是否有 HTTP 服务器在运行
			hasHTTPServer := builtins.IsHTTPServerRunning()

			// 如果有长期运行的任务（如 HTTP 服务器），持续运行
			if el.hasLongLived || hasHTTPServer {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// 如果有活跃的定时器、间隔器或异步任务，继续等待
			if hasTimers || hasIntervals || activeJobs > 0 {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			// 没有任何活跃任务，再等待一小段时间确认
			time.Sleep(50 * time.Millisecond)

			// 再次检查
			el.mu.Lock()
			hasTimers = len(el.timers) > 0
			hasIntervals = len(el.intervals) > 0
			el.mu.Unlock()
			activeJobs = atomic.LoadInt32(&el.activeJobs)
			hasHTTPServer = builtins.IsHTTPServerRunning()

			if !hasTimers && !hasIntervals && activeJobs == 0 && !el.hasLongLived && !hasHTTPServer {
				// 确实没有任务了，退出
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

	el.mu.Lock()
	el.timerID++
	id := el.timerID
	el.mu.Unlock()

	timer := time.AfterFunc(time.Duration(delay)*time.Millisecond, func() {
		// 使用 defer 和 recover 来处理可能的 panic
		defer func() {
			if r := recover(); r != nil {
				// 忽略 panic，避免影响其他操作
			}
			// 减少定时器计数
			pool.GlobalMemoryMonitor.DecrementTimerCount()
		}()

		fn(goja.Undefined())
		el.mu.Lock()
		delete(el.timers, id)
		el.mu.Unlock()
	})

	el.mu.Lock()
	el.timers[id] = timer
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
	defer el.mu.Unlock()

	if timer, ok := el.timers[id]; ok {
		timer.Stop()
		delete(el.timers, id)
		// 减少定时器计数
		pool.GlobalMemoryMonitor.DecrementTimerCount()
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

	el.mu.Lock()
	el.timerID++
	id := el.timerID
	el.mu.Unlock()

	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)

	el.mu.Lock()
	el.intervals[id] = ticker
	el.mu.Unlock()

	go func() {
		defer func() {
			ticker.Stop()
			el.mu.Lock()
			delete(el.intervals, id)
			el.mu.Unlock()
		}()

		for range ticker.C {
			el.mu.Lock()
			_, exists := el.intervals[id]
			el.mu.Unlock()

			if !exists {
				return
			}

			// 直接调用函数，避免复杂的队列机制
			// 使用 defer 和 recover 来处理可能的 panic
			func() {
				defer func() {
					if r := recover(); r != nil {
						// 忽略 panic，避免影响其他操作
					}
				}()
				fn(goja.Undefined())
			}()
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
	defer el.mu.Unlock()

	if ticker, ok := el.intervals[id]; ok {
		ticker.Stop()
		delete(el.intervals, id)
	}

	return goja.Undefined()
}
