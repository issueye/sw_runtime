package runtime

import (
	"sw_runtime/internal/pool"
	"sync"
	"time"

	"github.com/dop251/goja"
)

// SimpleEventLoop 简化的事件循环实现
type SimpleEventLoop struct {
	vm        *goja.Runtime
	timers    map[int]*time.Timer
	intervals map[int]*time.Ticker
	timerID   int
	mu        sync.Mutex
	running   bool
}

// NewSimpleEventLoop 创建简化的事件循环
func NewSimpleEventLoop(vm *goja.Runtime) *SimpleEventLoop {
	return &SimpleEventLoop{
		vm:        vm,
		timers:    make(map[int]*time.Timer),
		intervals: make(map[int]*time.Ticker),
	}
}

// Start 启动事件循环
func (el *SimpleEventLoop) Start() {
	el.running = true
}

// WaitAndProcess 等待并处理所有任务（简化版）
func (el *SimpleEventLoop) WaitAndProcess() {
	// 简单等待一段时间让异步任务完成
	time.Sleep(10 * time.Millisecond)
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
}
