package runtime

import (
	"container/heap"
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

// EventLoop 事件驱动的事件循环实现
// 使用最小堆管理定时器，通过 channel 驱动而非轮询
type EventLoop struct {
	vm *goja.Runtime

	// 定时器管理 - 使用最小堆
	timerHeap   timerHeap
	timerMu     sync.Mutex
	timerID     atomic.Int64
	timerNotify chan struct{} // 通知有新定时器加入

	// 间隔定时器
	intervals  map[int]*intervalTask
	intervalMu sync.RWMutex

	// VM 访问队列 - 串行化所有 JS 执行
	vmQueue chan vmTask
	vmMu    sync.Mutex // 备用锁，用于非队列场景

	// 状态管理
	running      atomic.Bool
	activeJobs   atomic.Int32
	hasLongLived atomic.Bool

	// 生命周期控制
	ctx       context.Context
	cancel    context.CancelFunc
	stopChan  chan struct{}
	stoppedCh chan struct{} // 用于等待完全停止

	// 配置
	idleTimeout time.Duration // 空闲超时时间
}

// vmTask VM 任务
type vmTask struct {
	fn   func()
	done chan struct{}
}

// timerTask 定时器任务
type timerTask struct {
	id       int
	deadline time.Time
	callback func()
	canceled atomic.Bool
	index    int // 堆索引
}

// intervalTask 间隔定时器任务
type intervalTask struct {
	id       int
	interval time.Duration
	callback func()
	canceled atomic.Bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// timerHeap 定时器最小堆
type timerHeap []*timerTask

func (h timerHeap) Len() int { return len(h) }

func (h timerHeap) Less(i, j int) bool {
	return h[i].deadline.Before(h[j].deadline)
}

func (h timerHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *timerHeap) Push(x interface{}) {
	n := len(*h)
	task := x.(*timerTask)
	task.index = n
	*h = append(*h, task)
}

func (h *timerHeap) Pop() interface{} {
	old := *h
	n := len(old)
	task := old[n-1]
	old[n-1] = nil  // 避免内存泄漏
	task.index = -1 // 标记已移除
	*h = old[0 : n-1]
	return task
}

// NewEventLoop 创建事件驱动的事件循环
func NewEventLoop(vm *goja.Runtime) *EventLoop {
	ctx, cancel := context.WithCancel(context.Background())
	el := &EventLoop{
		vm:          vm,
		timerHeap:   make(timerHeap, 0, 64),
		timerNotify: make(chan struct{}, 1),
		intervals:   make(map[int]*intervalTask, 32),
		vmQueue:     make(chan vmTask, 256),
		ctx:         ctx,
		cancel:      cancel,
		stopChan:    make(chan struct{}),
		stoppedCh:   make(chan struct{}),
		idleTimeout: 50 * time.Millisecond, // 默认 50ms 空闲超时
	}
	heap.Init(&el.timerHeap)
	return el
}

// Start 启动事件循环
func (el *EventLoop) Start() {
	if !el.running.CompareAndSwap(false, true) {
		return // 已经在运行
	}

	// 启动 VM 任务处理器
	go el.vmProcessor()

	// 启动定时器处理器
	go el.timerProcessor()
}

// vmProcessor VM 任务处理器 - 串行执行所有 JS 代码
func (el *EventLoop) vmProcessor() {
	for {
		select {
		case <-el.ctx.Done():
			return
		case task := <-el.vmQueue:
			el.safeExecute(task.fn)
			if task.done != nil {
				close(task.done)
			}
		}
	}
}

// timerProcessor 定时器处理器 - 事件驱动
func (el *EventLoop) timerProcessor() {
	for {
		el.timerMu.Lock()
		var nextDeadline time.Time
		var timer *time.Timer

		if el.timerHeap.Len() > 0 {
			nextDeadline = el.timerHeap[0].deadline
			delay := time.Until(nextDeadline)
			if delay <= 0 {
				// 已到期，立即处理
				el.processExpiredTimers()
				el.timerMu.Unlock()
				continue
			}
			timer = time.NewTimer(delay)
		}
		el.timerMu.Unlock()

		// 等待事件
		if timer != nil {
			select {
			case <-el.ctx.Done():
				timer.Stop()
				return
			case <-el.timerNotify:
				// 有新定时器加入，重新计算
				timer.Stop()
				continue
			case <-timer.C:
				// 定时器到期
				el.timerMu.Lock()
				el.processExpiredTimers()
				el.timerMu.Unlock()
			}
		} else {
			// 没有定时器，等待新定时器或停止信号
			select {
			case <-el.ctx.Done():
				return
			case <-el.timerNotify:
				continue
			}
		}
	}
}

// processExpiredTimers 处理到期的定时器（调用时必须持有 timerMu）
func (el *EventLoop) processExpiredTimers() {
	now := time.Now()
	for el.timerHeap.Len() > 0 {
		task := el.timerHeap[0]
		if task.deadline.After(now) {
			break
		}

		heap.Pop(&el.timerHeap)

		if task.canceled.Load() {
			continue
		}

		// 提交到 VM 队列执行
		callback := task.callback
		el.submitTask(func() {
			callback()
		})

		pool.GlobalMemoryMonitor.DecrementTimerCount()
	}
}

// submitTask 提交任务到 VM 队列
func (el *EventLoop) submitTask(fn func()) {
	task := vmTask{fn: fn}
	select {
	case el.vmQueue <- task:
	case <-el.ctx.Done():
	}
}

// submitTaskSync 同步提交任务并等待完成
func (el *EventLoop) submitTaskSync(fn func()) {
	task := vmTask{
		fn:   fn,
		done: make(chan struct{}),
	}
	select {
	case el.vmQueue <- task:
		<-task.done
	case <-el.ctx.Done():
	}
}

// safeExecute 安全执行函数，捕获 panic
func (el *EventLoop) safeExecute(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			// 记录错误但不影响事件循环
		}
	}()
	fn()
}

// notifyTimerChange 通知定时器变化
func (el *EventLoop) notifyTimerChange() {
	select {
	case el.timerNotify <- struct{}{}:
	default:
		// 已有通知待处理
	}
}

// Stop 停止事件循环
func (el *EventLoop) Stop() {
	if !el.running.CompareAndSwap(true, false) {
		return
	}

	el.cancel()

	// 清理所有定时器
	el.timerMu.Lock()
	for el.timerHeap.Len() > 0 {
		task := heap.Pop(&el.timerHeap).(*timerTask)
		task.canceled.Store(true)
		pool.GlobalMemoryMonitor.DecrementTimerCount()
	}
	el.timerMu.Unlock()

	// 清理所有间隔定时器
	el.intervalMu.Lock()
	for id, task := range el.intervals {
		task.canceled.Store(true)
		task.cancel()
		delete(el.intervals, id)
	}
	el.intervalMu.Unlock()

	// 发送停止信号
	select {
	case el.stopChan <- struct{}{}:
	default:
	}

	close(el.stoppedCh)
}

// AddJob 增加活跃任务计数
func (el *EventLoop) AddJob() {
	el.activeJobs.Add(1)
}

// DoneJob 减少活跃任务计数
func (el *EventLoop) DoneJob() {
	el.activeJobs.Add(-1)
}

// SetLongLived 标记有长期运行的任务
func (el *EventLoop) SetLongLived() {
	el.hasLongLived.Store(true)
}

// WaitAndProcess 等待并处理所有任务
func (el *EventLoop) WaitAndProcess() {
	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// 给异步任务一些启动时间
	time.Sleep(10 * time.Millisecond)

	// 空闲检测
	idleStart := time.Time{}
	checkTicker := time.NewTicker(10 * time.Millisecond)
	defer checkTicker.Stop()

	for el.running.Load() {
		select {
		case <-sigChan:
			el.Stop()
			return
		case <-el.stopChan:
			return
		case <-el.ctx.Done():
			return
		case <-checkTicker.C:
			if el.hasWork() {
				idleStart = time.Time{} // 重置空闲计时
				continue
			}

			// 开始空闲计时
			if idleStart.IsZero() {
				idleStart = time.Now()
				continue
			}

			// 检查是否超过空闲超时
			if time.Since(idleStart) >= el.idleTimeout {
				return
			}
		}
	}
}

// hasWork 检查是否有待处理的工作
func (el *EventLoop) hasWork() bool {
	// 检查长期运行任务
	if el.hasLongLived.Load() {
		return true
	}

	// 检查 HTTP 服务器
	if builtins.IsHTTPServerRunning() {
		return true
	}

	// 检查 TCP 服务器
	if builtins.IsTCPServerRunning() {
		return true
	}

	// 检查活跃任务
	if el.activeJobs.Load() > 0 {
		return true
	}

	// 检查定时器
	el.timerMu.Lock()
	hasTimers := el.timerHeap.Len() > 0
	el.timerMu.Unlock()

	if hasTimers {
		return true
	}

	// 检查间隔定时器
	el.intervalMu.RLock()
	hasIntervals := len(el.intervals) > 0
	el.intervalMu.RUnlock()

	return hasIntervals
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
	if delay < 0 {
		delay = 0
	}

	id := int(el.timerID.Add(1))

	// 捕获 VM 引用
	vm := el.vm

	task := &timerTask{
		id:       id,
		deadline: time.Now().Add(time.Duration(delay) * time.Millisecond),
		callback: func() {
			// 使用 vmMu 保护 VM 访问
			el.vmMu.Lock()
			defer el.vmMu.Unlock()
			defer func() {
				if r := recover(); r != nil {
					// 忽略回调中的 panic
				}
			}()
			fn(goja.Undefined())
		},
	}

	el.timerMu.Lock()
	heap.Push(&el.timerHeap, task)
	el.timerMu.Unlock()

	pool.GlobalMemoryMonitor.IncrementTimerCount()

	// 通知定时器处理器
	el.notifyTimerChange()

	return vm.ToValue(id)
}

// ClearTimeout 实现 clearTimeout
func (el *EventLoop) ClearTimeout(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.Undefined()
	}

	id := int(call.Arguments[0].ToInteger())

	el.timerMu.Lock()
	for i, task := range el.timerHeap {
		if task.id == id && !task.canceled.Load() {
			task.canceled.Store(true)
			heap.Remove(&el.timerHeap, i)
			pool.GlobalMemoryMonitor.DecrementTimerCount()
			break
		}
	}
	el.timerMu.Unlock()

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

	id := int(el.timerID.Add(1))

	ctx, cancel := context.WithCancel(el.ctx)
	task := &intervalTask{
		id:       id,
		interval: time.Duration(interval) * time.Millisecond,
		callback: func() {
			el.vmMu.Lock()
			defer el.vmMu.Unlock()
			defer func() {
				if r := recover(); r != nil {
					// 忽略回调中的 panic
				}
			}()
			fn(goja.Undefined())
		},
		ctx:    ctx,
		cancel: cancel,
	}

	el.intervalMu.Lock()
	el.intervals[id] = task
	el.intervalMu.Unlock()

	// 启动间隔执行 goroutine
	go el.runInterval(task)

	return el.vm.ToValue(id)
}

// runInterval 运行间隔定时器
func (el *EventLoop) runInterval(task *intervalTask) {
	ticker := time.NewTicker(task.interval)
	defer ticker.Stop()

	for {
		select {
		case <-task.ctx.Done():
			return
		case <-ticker.C:
			if task.canceled.Load() {
				return
			}
			el.submitTask(task.callback)
		}
	}
}

// ClearInterval 实现 clearInterval
func (el *EventLoop) ClearInterval(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.Undefined()
	}

	id := int(call.Arguments[0].ToInteger())

	el.intervalMu.Lock()
	if task, ok := el.intervals[id]; ok {
		task.canceled.Store(true)
		task.cancel()
		delete(el.intervals, id)
	}
	el.intervalMu.Unlock()

	return goja.Undefined()
}

// SetImmediate 实现 setImmediate（高优先级任务）
func (el *EventLoop) SetImmediate(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.Undefined()
	}

	fn, ok := goja.AssertFunction(call.Arguments[0])
	if !ok {
		return goja.Undefined()
	}

	id := int(el.timerID.Add(1))

	// 立即提交到队列
	el.submitTask(func() {
		el.vmMu.Lock()
		defer el.vmMu.Unlock()
		defer func() {
			if r := recover(); r != nil {
				// 忽略回调中的 panic
			}
		}()
		fn(goja.Undefined())
	})

	return el.vm.ToValue(id)
}

// NextTick 实现 process.nextTick
func (el *EventLoop) NextTick(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		return goja.Undefined()
	}

	fn, ok := goja.AssertFunction(call.Arguments[0])
	if !ok {
		return goja.Undefined()
	}

	// 提交到队列
	el.submitTask(func() {
		el.vmMu.Lock()
		defer el.vmMu.Unlock()
		defer func() {
			if r := recover(); r != nil {
				// 忽略回调中的 panic
			}
		}()
		fn(goja.Undefined())
	})

	return goja.Undefined()
}

// GetPendingTimerCount 获取待处理的定时器数量（用于调试）
func (el *EventLoop) GetPendingTimerCount() int {
	el.timerMu.Lock()
	defer el.timerMu.Unlock()
	return el.timerHeap.Len()
}

// GetActiveIntervalCount 获取活跃的间隔定时器数量（用于调试）
func (el *EventLoop) GetActiveIntervalCount() int {
	el.intervalMu.RLock()
	defer el.intervalMu.RUnlock()
	return len(el.intervals)
}

// RunOnLoopSync 在事件循环中同步执行函数并返回结果
// 用于从其他 goroutine (如 Raft Controller) 同步调用 JS 逻辑
func (el *EventLoop) RunOnLoopSync(fn func(*goja.Runtime) interface{}) interface{} {
	var result interface{}
	done := make(chan struct{})

	task := func() {
		defer close(done)
		el.vmMu.Lock()
		defer el.vmMu.Unlock()

		defer func() {
			if r := recover(); r != nil {
				// 记录 panic 但不崩溃
			}
		}()

		result = fn(el.vm)
	}

	el.submitTask(task)
	<-done
	return result
}
