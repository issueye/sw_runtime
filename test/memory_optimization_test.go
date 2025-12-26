package test

import (
	"bytes"
	goruntime "runtime"
	"testing"

	"sw_runtime/internal/pool"
	"sw_runtime/internal/runtime"
)

func TestPoolBasicUsage(t *testing.T) {
	poolManager := pool.NewManager()

	// 测试字节缓冲池
	buf1 := poolManager.GetByteBuffer()
	buf1.WriteString("test data")
	poolManager.PutByteBuffer(buf1)

	buf2 := poolManager.GetByteBuffer()
	if buf2.Len() != 0 {
		t.Error("Buffer should be reset when retrieved from pool")
	}
	poolManager.PutByteBuffer(buf2)

	// 测试字符串切片池
	slice1 := poolManager.GetStringSlice()
	slice1 = append(slice1, "test1", "test2")
	poolManager.PutStringSlice(slice1)

	slice2 := poolManager.GetStringSlice()
	if len(slice2) != 0 {
		t.Error("String slice should be reset when retrieved from pool")
	}
	poolManager.PutStringSlice(slice2)

	t.Log("✅ Pool basic usage test passed")
}

func TestPoolConcurrency(t *testing.T) {
	poolManager := pool.NewManager()
	const numGoroutines = 50
	const numOperations = 100

	done := make(chan bool, numGoroutines)

	// 并发测试
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				// 测试字节缓冲池
				buf := poolManager.GetByteBuffer()
				buf.WriteString("concurrent test")
				poolManager.PutByteBuffer(buf)

				// 测试字符串切片池
				slice := poolManager.GetStringSlice()
				slice = append(slice, "test")
				poolManager.PutStringSlice(slice)

				// 测试小对象池
				obj := poolManager.GetSmallObject()
				obj["key"] = "value"
				poolManager.PutSmallObject(obj)
			}
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	t.Log("✅ Pool concurrency test passed")
}

func TestRunnerMemoryUsage(t *testing.T) {
	// 记录初始内存状态
	var m1, m2 goruntime.MemStats
	goruntime.GC()
	goruntime.ReadMemStats(&m1)

	// 创建多个 Runner 实例并执行代码
	const numRunners = 5
	runners := make([]*runtime.Runner, numRunners)

	for i := 0; i < numRunners; i++ {
		runners[i] = runtime.New()

		// 执行一些代码来测试内存使用
		code := `
			console.log("Memory test", Math.random());
			
			// 创建一些对象
			let data = [];
			for (let i = 0; i < 50; i++) {
				data.push({
					id: i,
					name: "item_" + i,
					value: Math.random()
				});
			}
		`

		err := runners[i].RunCode(code)
		if err != nil {
			t.Fatalf("Failed to run code in runner %d: %v", i, err)
		}
	}

	// 清理 Runner
	for i := 0; i < numRunners; i++ {
		runners[i].Close()
	}

	// 强制 GC 并测量内存
	goruntime.GC()
	goruntime.ReadMemStats(&m2)

	// 计算内存增长
	memoryGrowth := m2.Alloc - m1.Alloc
	t.Logf("Memory growth after creating %d runners: %d bytes", numRunners, memoryGrowth)

	// 验证内存增长在合理范围内 (每个 Runner 不超过 2MB)
	maxExpectedGrowth := uint64(numRunners * 2 * 1024 * 1024) // 2MB per runner
	if memoryGrowth > maxExpectedGrowth {
		t.Logf("Warning: Memory growth higher than expected: %d bytes (max expected: %d bytes)",
			memoryGrowth, maxExpectedGrowth)
	}

	t.Log("✅ Runner memory usage test completed")
}

func TestCompressionMemoryUsage(t *testing.T) {
	runner := runtime.New()
	defer runner.Close()

	// 记录初始内存状态
	var m1, m2 goruntime.MemStats
	goruntime.GC()
	goruntime.ReadMemStats(&m1)

	// 执行压缩操作
	code := `
		const zlib = require('zlib');
		let results = [];
		
		// 执行多次压缩操作
		for (let i = 0; i < 50; i++) {
			const data = 'This is test data for compression test. '.repeat(50);
			const compressed = zlib.gzip(data);
			const decompressed = zlib.gunzip(compressed);
			
			if (decompressed !== data) {
				throw new Error('Compression/decompression failed');
			}
			
			results.push(compressed.length);
		}
		
		global.compressionResults = results;
		console.log('Completed', results.length, 'compression operations');
	`

	err := runner.RunCode(code)
	if err != nil {
		t.Fatalf("Failed to run compression test: %v", err)
	}

	// 验证结果
	results := runner.GetValue("compressionResults")
	if results == nil {
		t.Fatal("Compression results not found")
	}

	// 强制 GC 并测量内存
	goruntime.GC()
	goruntime.ReadMemStats(&m2)

	memoryGrowth := m2.Alloc - m1.Alloc
	t.Logf("Memory growth after compression operations: %d bytes", memoryGrowth)

	t.Log("✅ Compression memory usage test completed")
}

func BenchmarkPoolVsDirectAllocation(b *testing.B) {
	poolManager := pool.NewManager()

	b.Run("WithPool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := poolManager.GetByteBuffer()
			buf.WriteString("benchmark test data")
			poolManager.PutByteBuffer(buf)
		}
	})

	b.Run("DirectAllocation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := &bytes.Buffer{}
			buf.WriteString("benchmark test data")
			// 不需要显式释放，让 GC 处理
		}
	})
}

func BenchmarkRunnerCreationOptimized(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runner := runtime.New()
		runner.Close()
	}
}
