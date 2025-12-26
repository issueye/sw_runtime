package runtime

import (
	"sync"
	"time"

	"github.com/dop251/goja"
)

// EventLoop 实现简单的事件循环，支持异步操作
type EventLoop struct {
	vm        *goja.Runtime
	jobQueue  chan func()
	timers    map[int]*time.Timer
	intervals map[int]*time.Ticker
	timerID   int
	mu        sync.Mutex
	running   bool
	wg        sync.WaitGroup
}

// NewEventLoop 创建新的事件循环
func NewEventLoop(vm *goja.Runtime) *EventLoop {
	return &EventLoop{
		vm:        vm,
		jobQueue:  make(chan func(), 100),
		timers:    make(map[int]*time.Timer),
		intervals: make(map[int]*time.Ticker),
	}
}

// Start 启动事件循环
func (el *EventLoop) Start() {
	el.running = true
}

// RunOnLoop 在事件循环中执行任务
func (el *EventLoop) RunOnLoop(fn func()) {
	el.wg.Add(1)
	el.jobQueue <- func() {
		defer el.wg.Done()
		fn()
	}
}

// WaitAndProcess 等待并处理所有任务
func (el *EventLoop) WaitAndProcess() {
	done := make(chan struct{})
	go func() {
		el.wg.Wait()
		close(done)
	}()

	for {
		select {
		case job := <-el.jobQueue:
			job()
		case <-done:
			// 再处理一下剩余的任务
			for {
				select {
				case job := <-el.jobQueue:
					job()
				default:
					return
				}
			}
		}
	}
}

// SetTimeout 实现 setTimeout
func (el *EventLoop) SetTimeout(call goja.FunctionCall) goja.Value {
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

	el.wg.Add(1)
	timer := time.AfterFunc(time.Duration(delay)*time.Millisecond, func() {
		el.jobQueue <- func() {
			defer el.wg.Done()
			fn(goja.Undefined())
		}
		el.mu.Lock()
		delete(el.timers, id)
		el.mu.Unlock()
	})

	el.mu.Lock()
	el.timers[id] = timer
	el.mu.Unlock()

	return el.vm.ToValue(id)
}

// ClearTimeout 实现 clearTimeout
func (el *EventLoop) ClearTimeout(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.Undefined()
	}

	id := int(call.Arguments[0].ToInteger())

	el.mu.Lock()
	defer el.mu.Unlock()

	if timer, ok := el.timers[id]; ok {
		timer.Stop()
		delete(el.timers, id)
		// 安全地减少 WaitGroup 计数
		go func() {
			defer func() {
				if r := recover(); r != nil {
					// 忽略 WaitGroup 计数器为负的 panic
				}
			}()
			el.wg.Done()
		}()
	}

	return goja.Undefined()
}

// SetInterval 实现 setInterval
func (el *EventLoop) SetInterval(call goja.FunctionCall) goja.Value {
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

	el.wg.Add(1)
	go func() {
		defer func() {
			el.wg.Done()
			ticker.Stop()
			el.mu.Lock()
			delete(el.intervals, id)
			el.mu.Unlock()
		}()

		for {
			select {
			case <-ticker.C:
				el.mu.Lock()
				_, exists := el.intervals[id]
				el.mu.Unlock()

				if !exists {
					return
				}

				el.jobQueue <- func() {
					fn(goja.Undefined())
				}
			}
		}
	}()

	return el.vm.ToValue(id)
}

// ClearInterval 实现 clearInterval
func (el *EventLoop) ClearInterval(call goja.FunctionCall) goja.Value {
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
func (el *EventLoop) Stop() {
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
