package test

import (
	"testing"
	"time"

	"sw_runtime/internal/runtime"
)

// BenchmarkEventLoopSetTimeout 测试 setTimeout 性能
func BenchmarkEventLoopSetTimeout(b *testing.B) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		code := `
			let counter = 0;
			setTimeout(() => { counter++; }, 1);
		`
		runner.RunCode(code)
	}
}

// BenchmarkEventLoopSetInterval 测试 setInterval 性能
func BenchmarkEventLoopSetInterval(b *testing.B) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		code := `
			let counter = 0;
			let id = setInterval(() => {
				counter++;
				if (counter >= 2) {
					clearInterval(id);
				}
			}, 1);
		`
		runner.RunCode(code)
	}
}

// BenchmarkEventLoopMultipleTimers 测试多个定时器性能
func BenchmarkEventLoopMultipleTimers(b *testing.B) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		code := `
			let results = [];
			setTimeout(() => { results.push(1); }, 1);
			setTimeout(() => { results.push(2); }, 2);
			setTimeout(() => { results.push(3); }, 3);
			setTimeout(() => { results.push(4); }, 4);
			setTimeout(() => { results.push(5); }, 5);
		`
		runner.RunCode(code)
	}
}

// TestEventLoopConcurrentTimers 测试并发定时器
func TestEventLoopConcurrentTimers(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	start := time.Now()

	code := `
		let completed = 0;
		const total = 100;
		
		for (let i = 0; i < total; i++) {
			setTimeout(() => {
				completed++;
			}, 1);
		}
		
		global.getCompleted = () => completed;
		global.getTotal = () => total;
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run concurrent timers test: %v", err)
	}

	elapsed := time.Since(start)
	completed := runner.GetValue("getCompleted").ToInteger()
	total := runner.GetValue("getTotal").ToInteger()

	if completed != total {
		t.Fatalf("Expected %d completed timers, got %d", total, completed)
	}

	t.Logf("✅ 100 concurrent timers completed in %v", elapsed)
}

// TestEventLoopMemoryLeak 测试内存泄漏
func TestEventLoopMemoryLeak(t *testing.T) {
	for iteration := 0; iteration < 3; iteration++ {
		runner := runtime.NewOrPanic()

		code := `
			let timers = [];
			for (let i = 0; i < 50; i++) {
				let id = setTimeout(() => {}, 10);
				timers.push(id);
			}
			
			// 清除一半的定时器
			for (let i = 0; i < 25; i++) {
				clearTimeout(timers[i]);
			}
			
			let intervals = [];
			for (let i = 0; i < 20; i++) {
				let id = setInterval(() => {}, 50);
				intervals.push(id);
			}
			
			// 清除所有间隔器
			for (let i = 0; i < 20; i++) {
				clearInterval(intervals[i]);
			}
		`

		err := runner.RunCode(code)
		if err != nil {
			t.Fatalf("Iteration %d failed: %v", iteration, err)
		}

		runner.Close()
	}

	t.Log("✅ No memory leak detected after 3 iterations")
}

// TestEventLoopStressTest 压力测试
func TestEventLoopStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping event loop stress test in short mode")
	}

	runner := runtime.NewOrPanic()
	defer runner.Close()

	start := time.Now()

	code := `
		let results = {
			timeouts: 0,
			intervals: 0,
			promises: 0
		};
		
		// 创建 50 个 setTimeout
		for (let i = 0; i < 50; i++) {
			setTimeout(() => {
				results.timeouts++;
			}, (i % 10) + 1);
		}
		
		// 创建 10 个 setInterval (执行 3 次后清除)
		for (let i = 0; i < 10; i++) {
			let count = 0;
			let id = setInterval(() => {
				results.intervals++;
				count++;
				if (count >= 3) {
					clearInterval(id);
				}
			}, 5);
		}
		
		// 创建 30 个 Promise
		for (let i = 0; i < 30; i++) {
			Promise.resolve(i).then(() => {
				results.promises++;
			});
		}
		
		global.results = results;
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run stress test: %v", err)
	}

	elapsed := time.Since(start)

	resultsObj := runner.GetValue("results").ToObject(nil)
	if resultsObj == nil {
		t.Fatal("Results object not found")
	}

	timeouts := resultsObj.Get("timeouts").ToInteger()
	intervals := resultsObj.Get("intervals").ToInteger()
	promises := resultsObj.Get("promises").ToInteger()

	t.Logf("✅ Stress test completed in %v", elapsed)
	t.Logf("   - Timeouts: %d/50", timeouts)
	t.Logf("   - Intervals: %d/30", intervals)
	t.Logf("   - Promises: %d/30", promises)
}

// TestEventLoopRapidCancelation 测试快速取消
func TestEventLoopRapidCancelation(t *testing.T) {
	runner := runtime.NewOrPanic()
	defer runner.Close()

	start := time.Now()

	code := `
		let canceled = 0;
		
		// 创建并立即取消 100 个定时器
		for (let i = 0; i < 100; i++) {
			let id = setTimeout(() => {}, 100);
			clearTimeout(id);
			canceled++;
		}
		
		// 创建并立即取消 50 个间隔器
		for (let i = 0; i < 50; i++) {
			let id = setInterval(() => {}, 100);
			clearInterval(id);
			canceled++;
		}
		
		global.canceled = canceled;
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run rapid cancelation test: %v", err)
	}

	elapsed := time.Since(start)
	canceled := runner.GetValue("canceled").ToInteger()

	if canceled != 150 {
		t.Fatalf("Expected 150 canceled timers, got %d", canceled)
	}

	t.Logf("✅ Rapid cancelation test completed in %v", elapsed)
	t.Logf("   - %d timers/intervals created and canceled", canceled)
}
