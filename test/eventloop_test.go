package test

import (
	"testing"
	"time"

	"sw_runtime/internal/runtime"

	"github.com/dop251/goja"
)

func TestEventLoopBasicTimer(t *testing.T) {
	runner := runtime.NewOrPanic()

	code := `
		let timerExecuted = false;
		let timerValue = null;
		
		setTimeout(() => {
			timerExecuted = true;
			timerValue = "timer executed";
		}, 50);
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run timer test: %v", err)
	}

	// 等待定时器执行
	time.Sleep(100 * time.Millisecond)

	executed := runner.GetValue("timerExecuted")
	if !executed.ToBoolean() {
		t.Fatal("Timer was not executed")
	}

	value := runner.GetValue("timerValue")
	if value.String() != "timer executed" {
		t.Fatal("Timer did not set correct value")
	}
}

func TestEventLoopInterval(t *testing.T) {
	runner := runtime.NewOrPanic()

	code := `
		let intervalCount = 0;
		let intervalId = null;
		
		intervalId = setInterval(() => {
			intervalCount++;
			console.log('Interval executed:', intervalCount);
			if (intervalCount >= 2) {
				clearInterval(intervalId);
			}
		}, 50);
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run interval test: %v", err)
	}

	// 等待间隔执行
	time.Sleep(300 * time.Millisecond)

	count := runner.GetValue("intervalCount")
	countNum := count.ToInteger()
	if countNum < 2 {
		t.Fatalf("Expected interval to execute at least 2 times, got %d", countNum)
	}
}

func TestEventLoopClearTimeout(t *testing.T) {
	runner := runtime.NewOrPanic()

	code := `
		let shouldNotExecute = false;
		
		const timerId = setTimeout(() => {
			shouldNotExecute = true;
		}, 50);
		
		// 立即清除定时器
		clearTimeout(timerId);
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run clear timeout test: %v", err)
	}

	// 等待足够长的时间
	time.Sleep(100 * time.Millisecond)

	executed := runner.GetValue("shouldNotExecute")
	if executed.ToBoolean() {
		t.Fatal("Cleared timeout should not have executed")
	}
}

func TestEventLoopMultipleTimers(t *testing.T) {
	runner := runtime.NewOrPanic()

	code := `
		let results = [];
		
		setTimeout(() => {
			results.push("timer1");
		}, 30);
		
		setTimeout(() => {
			results.push("timer2");
		}, 20);
		
		setTimeout(() => {
			results.push("timer3");
		}, 40);
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run multiple timers test: %v", err)
	}

	// 等待所有定时器执行
	time.Sleep(100 * time.Millisecond)

	results := runner.GetValue("results")
	if results == nil {
		t.Fatal("Results array not found")
	}

	// 验证所有定时器都执行了
	resultsObj := results.(*goja.Object)
	length := resultsObj.Get("length").ToInteger()
	if length != 3 {
		t.Fatalf("Expected 3 timer results, got %d", length)
	}
}

func TestEventLoopNestedTimers(t *testing.T) {
	runner := runtime.NewOrPanic()

	code := `
		let nestedResult = null;
		
		setTimeout(() => {
			setTimeout(() => {
				nestedResult = "nested timer executed";
			}, 20);
		}, 30);
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run nested timers test: %v", err)
	}

	// 等待嵌套定时器执行
	time.Sleep(100 * time.Millisecond)

	result := runner.GetValue("nestedResult")
	if result.String() != "nested timer executed" {
		t.Fatal("Nested timer was not executed correctly")
	}
}

func TestEventLoopPromiseIntegration(t *testing.T) {
	runner := runtime.NewOrPanic()

	code := `
		let promiseResult = null;
		let timerResult = null;
		
		// Promise 应该在下一个 tick 执行
		Promise.resolve("promise value").then((value) => {
			promiseResult = value;
		});
		
		// Timer 应该在指定延迟后执行
		setTimeout(() => {
			timerResult = "timer value";
		}, 20);
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run promise integration test: %v", err)
	}

	// 等待异步操作完成
	time.Sleep(50 * time.Millisecond)

	promiseResult := runner.GetValue("promiseResult")
	if promiseResult.String() != "promise value" {
		t.Fatal("Promise was not resolved correctly")
	}

	timerResult := runner.GetValue("timerResult")
	if timerResult.String() != "timer value" {
		t.Fatal("Timer was not executed correctly")
	}
}

func TestEventLoopErrorHandling(t *testing.T) {
	runner := runtime.NewOrPanic()

	code := `
		let errorCaught = false;
		let normalTimerExecuted = false;
		
		// 这个定时器会抛出错误
		setTimeout(() => {
			try {
				throw new Error("Timer error");
			} catch (e) {
				errorCaught = true;
			}
		}, 20);
		
		// 这个定时器应该正常执行
		setTimeout(() => {
			normalTimerExecuted = true;
		}, 40);
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run error handling test: %v", err)
	}

	// 等待定时器执行
	time.Sleep(80 * time.Millisecond)

	normalExecuted := runner.GetValue("normalTimerExecuted")
	if !normalExecuted.ToBoolean() {
		t.Fatal("Normal timer should have executed despite error in another timer")
	}

	errorCaught := runner.GetValue("errorCaught")
	if !errorCaught.ToBoolean() {
		t.Fatal("Error should have been caught")
	}
}
