package runtime

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dop251/goja"
)

func TestEventLoopSetTimeout(t *testing.T) {
	vm := goja.New()
	el := NewEventLoop(vm)
	el.Start()
	defer el.Stop()

	var executed atomic.Bool

	vm.Set("setTimeout", el.SetTimeout)
	vm.Set("callback", func() {
		executed.Store(true)
	})

	_, err := vm.RunString(`setTimeout(callback, 10)`)
	if err != nil {
		t.Fatalf("Failed to run setTimeout: %v", err)
	}

	// 等待执行
	time.Sleep(50 * time.Millisecond)

	if !executed.Load() {
		t.Error("setTimeout callback was not executed")
	}
}

func TestEventLoopClearTimeout(t *testing.T) {
	vm := goja.New()
	el := NewEventLoop(vm)
	el.Start()
	defer el.Stop()

	var executed atomic.Bool

	vm.Set("setTimeout", el.SetTimeout)
	vm.Set("clearTimeout", el.ClearTimeout)
	vm.Set("callback", func() {
		executed.Store(true)
	})

	_, err := vm.RunString(`
		var id = setTimeout(callback, 50);
		clearTimeout(id);
	`)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if executed.Load() {
		t.Error("Cleared timeout should not execute")
	}
}

func TestEventLoopSetInterval(t *testing.T) {
	vm := goja.New()
	el := NewEventLoop(vm)
	el.Start()
	defer el.Stop()

	var count atomic.Int32

	vm.Set("setTimeout", el.SetTimeout)
	vm.Set("setInterval", el.SetInterval)
	vm.Set("clearInterval", el.ClearInterval)
	vm.Set("increment", func() {
		count.Add(1)
	})

	_, err := vm.RunString(`
		var id = setInterval(increment, 20);
		setTimeout(function() { clearInterval(id); }, 100);
	`)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	time.Sleep(150 * time.Millisecond)

	c := count.Load()
	if c < 3 || c > 6 {
		t.Errorf("Expected 3-6 interval executions, got %d", c)
	}
}

func TestEventLoopMultipleTimers(t *testing.T) {
	vm := goja.New()
	el := NewEventLoop(vm)
	el.Start()
	defer el.Stop()

	var order []int
	var mu sync.Mutex

	vm.Set("setTimeout", el.SetTimeout)
	vm.Set("record", func(n int) {
		mu.Lock()
		order = append(order, n)
		mu.Unlock()
	})

	_, err := vm.RunString(`
		setTimeout(function() { record(3); }, 30);
		setTimeout(function() { record(1); }, 10);
		setTimeout(function() { record(2); }, 20);
	`)
	if err != nil {
		t.Fatalf("Failed to run script: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(order) != 3 {
		t.Fatalf("Expected 3 executions, got %d", len(order))
	}

	// 验证执行顺序
	for i, expected := range []int{1, 2, 3} {
		if order[i] != expected {
			t.Errorf("Expected order[%d]=%d, got %d", i, expected, order[i])
		}
	}
}

func TestEventLoopSetImmediate(t *testing.T) {
	vm := goja.New()
	el := NewEventLoop(vm)
	el.Start()
	defer el.Stop()

	var executed atomic.Bool

	vm.Set("setImmediate", el.SetImmediate)
	vm.Set("callback", func() {
		executed.Store(true)
	})

	_, err := vm.RunString(`setImmediate(callback)`)
	if err != nil {
		t.Fatalf("Failed to run setImmediate: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	if !executed.Load() {
		t.Error("setImmediate callback was not executed")
	}
}

func TestEventLoopConcurrentTimers(t *testing.T) {
	vm := goja.New()
	el := NewEventLoop(vm)
	el.Start()
	defer el.Stop()

	var count atomic.Int32
	const numTimers = 100

	vm.Set("setTimeout", el.SetTimeout)
	vm.Set("increment", func() {
		count.Add(1)
	})

	// 创建多个定时器
	for i := 0; i < numTimers; i++ {
		_, err := vm.RunString(`setTimeout(increment, 10)`)
		if err != nil {
			t.Fatalf("Failed to create timer %d: %v", i, err)
		}
	}

	time.Sleep(200 * time.Millisecond)

	if count.Load() != numTimers {
		t.Errorf("Expected %d executions, got %d", numTimers, count.Load())
	}
}

func BenchmarkEventLoopSetTimeout(b *testing.B) {
	vm := goja.New()
	el := NewEventLoop(vm)
	el.Start()

	vm.Set("setTimeout", el.SetTimeout)
	vm.Set("noop", func() {})

	// 预编译脚本 - 使用长延迟避免回调执行
	prog, _ := goja.Compile("", `setTimeout(noop, 100000)`, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		el.vmMu.Lock()
		vm.RunProgram(prog)
		el.vmMu.Unlock()
	}
	b.StopTimer()

	// 清理定时器
	el.Stop()
}

func BenchmarkEventLoopSetInterval(b *testing.B) {
	vm := goja.New()
	el := NewEventLoop(vm)
	el.Start()

	vm.Set("setInterval", el.SetInterval)
	vm.Set("clearInterval", el.ClearInterval)
	vm.Set("noop", func() {})

	// 预编译脚本
	prog, _ := goja.Compile("", `var id = setInterval(noop, 100000); clearInterval(id);`, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		el.vmMu.Lock()
		vm.RunProgram(prog)
		el.vmMu.Unlock()
	}
	b.StopTimer()

	el.Stop()
}

func BenchmarkSimpleEventLoopSetTimeout(b *testing.B) {
	vm := goja.New()
	el := NewSimpleEventLoop(vm)
	el.Start()

	vm.Set("setTimeout", el.SetTimeout)
	vm.Set("noop", func() {})

	// 预编译脚本 - 使用长延迟避免回调执行
	prog, _ := goja.Compile("", `setTimeout(noop, 100000)`, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		el.vmMu.Lock()
		vm.RunProgram(prog)
		el.vmMu.Unlock()
	}
	b.StopTimer()

	// 清理定时器
	el.Stop()
}

func BenchmarkSimpleEventLoopSetInterval(b *testing.B) {
	vm := goja.New()
	el := NewSimpleEventLoop(vm)
	el.Start()

	vm.Set("setInterval", el.SetInterval)
	vm.Set("clearInterval", el.ClearInterval)
	vm.Set("noop", func() {})

	// 预编译脚本
	prog, _ := goja.Compile("", `var id = setInterval(noop, 100000); clearInterval(id);`, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		el.vmMu.Lock()
		vm.RunProgram(prog)
		el.vmMu.Unlock()
	}
	b.StopTimer()

	el.Stop()
}
