package test

import (
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dop251/goja"

	"sw_runtime/internal/builtins"
)

// TestVMProcessorPerformance VMProcessor 性能测试
func TestVMProcessorPerformance(t *testing.T) {
	t.Log("==========================================")
	t.Log("VMProcessor 性能测试报告")
	t.Log("==========================================\n")

	// 测试 1: 串行请求性能
	t.Run("SerialRequests", func(t *testing.T) {
		testSerialRequestPerformance(t)
	})

	// 测试 2: 并发请求性能
	t.Run("ConcurrentRequests", func(t *testing.T) {
		testConcurrentRequestPerformance(t)
	})

	// 测试 3: 高并发压力测试
	t.Run("HighConcurrencyStress", func(t *testing.T) {
		testHighConcurrencyStress(t)
	})

	// 测试 4: 混合路由性能
	t.Run("MixedRoutes", func(t *testing.T) {
		testMixedRoutesPerformance(t)
	})

	// 测试 5: 长时间稳定性测试
	t.Run("LongRunningStability", func(t *testing.T) {
		testLongRunningStability(t)
	})

	t.Log("\n==========================================")
	t.Log("性能测试完成")
	t.Log("==========================================")
}

// testSerialRequestPerformance 串行请求性能测试
func testSerialRequestPerformance(t *testing.T) {
	vm := goja.New()
	httpModule := builtins.NewHTTPServerModule(vm)
	vm.Set("httpserver", httpModule.GetModule())

	script := `
		const server = httpserver.createServer({
			readTimeout: 30,
			writeTimeout: 30
		});

		server.get('/test', (req, res) => {
			res.json({ message: 'Serial test', timestamp: Date.now() });
		});

		server.listen('38901');
	`

	_, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("创建服务器失败: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// 串行请求测试
	requestCount := 100
	startTime := time.Now()

	for i := 0; i < requestCount; i++ {
		resp, err := http.Get("http://localhost:38901/test")
		if err != nil {
			t.Errorf("请求失败: %v", err)
			continue
		}
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}

	duration := time.Since(startTime)
	avgLatency := duration / time.Duration(requestCount)
	throughput := float64(requestCount) / duration.Seconds()

	t.Logf("\n📊 串行请求性能:")
	t.Logf("   - 请求数量: %d", requestCount)
	t.Logf("   - 总耗时: %v", duration)
	t.Logf("   - 平均延迟: %v", avgLatency)
	t.Logf("   - 吞吐量: %.2f req/s", throughput)

	if avgLatency > 50*time.Millisecond {
		t.Logf("   ⚠️  平均延迟偏高")
	} else {
		t.Logf("   ✅ 延迟表现良好")
	}
}

// testConcurrentRequestPerformance 并发请求性能测试
func testConcurrentRequestPerformance(t *testing.T) {
	vm := goja.New()
	httpModule := builtins.NewHTTPServerModule(vm)
	vm.Set("httpserver", httpModule.GetModule())

	script := `
		const server = httpserver.createServer({
			readTimeout: 30,
			writeTimeout: 30
		});

		server.get('/concurrent', (req, res) => {
			res.json({ message: 'Concurrent test', id: Math.random() });
		});

		server.listen('38902');
	`

	_, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("创建服务器失败: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// 并发请求测试
	concurrency := 10
	requestsPerWorker := 50
	totalRequests := concurrency * requestsPerWorker

	var wg sync.WaitGroup
	var successCount, errorCount int32
	startTime := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				resp, err := http.Get("http://localhost:38902/concurrent")
				if err != nil {
					atomic.AddInt32(&errorCount, 1)
					continue
				}
				io.ReadAll(resp.Body)
				resp.Body.Close()
				atomic.AddInt32(&successCount, 1)
			}
		}()
	}

	wg.Wait()
	duration := time.Since(startTime)
	avgLatency := duration / time.Duration(totalRequests)
	throughput := float64(totalRequests) / duration.Seconds()
	successRate := float64(successCount) / float64(totalRequests) * 100

	t.Logf("\n📊 并发请求性能:")
	t.Logf("   - 并发数: %d", concurrency)
	t.Logf("   - 总请求数: %d", totalRequests)
	t.Logf("   - 成功请求: %d", successCount)
	t.Logf("   - 失败请求: %d", errorCount)
	t.Logf("   - 成功率: %.2f%%", successRate)
	t.Logf("   - 总耗时: %v", duration)
	t.Logf("   - 平均延迟: %v", avgLatency)
	t.Logf("   - 吞吐量: %.2f req/s", throughput)

	if successRate < 95 {
		t.Errorf("   ❌ 成功率过低: %.2f%%", successRate)
	} else if successRate < 99 {
		t.Logf("   ⚠️  成功率可以提升: %.2f%%", successRate)
	} else {
		t.Logf("   ✅ 成功率优秀")
	}
}

// testHighConcurrencyStress 高并发压力测试
func testHighConcurrencyStress(t *testing.T) {
	vm := goja.New()
	httpModule := builtins.NewHTTPServerModule(vm)
	vm.Set("httpserver", httpModule.GetModule())

	script := `
		const server = httpserver.createServer({
			readTimeout: 30,
			writeTimeout: 30
		});

		server.get('/stress', (req, res) => {
			res.json({ status: 'ok', load: 'high' });
		});

		server.listen('38903');
	`

	_, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("创建服务器失败: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// 高并发压力测试
	concurrency := 50
	requestsPerWorker := 20
	totalRequests := concurrency * requestsPerWorker

	var wg sync.WaitGroup
	var successCount, errorCount int32
	var totalLatency int64 // 纳秒

	startTime := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				reqStart := time.Now()
				resp, err := http.Get("http://localhost:38903/stress")
				reqDuration := time.Since(reqStart)

				if err != nil {
					atomic.AddInt32(&errorCount, 1)
					continue
				}
				io.ReadAll(resp.Body)
				resp.Body.Close()
				atomic.AddInt32(&successCount, 1)
				atomic.AddInt64(&totalLatency, reqDuration.Nanoseconds())
			}
		}()
	}

	wg.Wait()
	duration := time.Since(startTime)
	avgLatency := time.Duration(totalLatency / int64(successCount))
	throughput := float64(totalRequests) / duration.Seconds()
	successRate := float64(successCount) / float64(totalRequests) * 100

	t.Logf("\n📊 高并发压力测试:")
	t.Logf("   - 并发数: %d", concurrency)
	t.Logf("   - 总请求数: %d", totalRequests)
	t.Logf("   - 成功请求: %d", successCount)
	t.Logf("   - 失败请求: %d", errorCount)
	t.Logf("   - 成功率: %.2f%%", successRate)
	t.Logf("   - 总耗时: %v", duration)
	t.Logf("   - 平均延迟: %v", avgLatency)
	t.Logf("   - 吞吐量: %.2f req/s", throughput)

	// 性能评估
	if throughput > 10000 {
		t.Logf("   🚀 性能优秀 (>10k req/s)")
	} else if throughput > 5000 {
		t.Logf("   ✅ 性能良好 (>5k req/s)")
	} else if throughput > 2000 {
		t.Logf("   ⚠️  性能一般 (>2k req/s)")
	} else {
		t.Logf("   ❌ 性能偏低 (<2k req/s)")
	}

	if successRate < 95 {
		t.Errorf("   ❌ 高并发下成功率过低: %.2f%%", successRate)
	}
}

// testMixedRoutesPerformance 混合路由性能测试
func testMixedRoutesPerformance(t *testing.T) {
	vm := goja.New()
	httpModule := builtins.NewHTTPServerModule(vm)
	vm.Set("httpserver", httpModule.GetModule())

	script := `
		const server = httpserver.createServer({
			readTimeout: 30,
			writeTimeout: 30
		});

		server.get('/fast', (req, res) => {
			res.json({ type: 'fast' });
		});

		server.get('/medium', (req, res) => {
			res.json({ type: 'medium', data: 'some data here' });
		});

		server.post('/echo', (req, res) => {
			res.json({ type: 'echo', body: req.body });
		});

		server.listen('38904');
	`

	_, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("创建服务器失败: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// 混合路由测试
	concurrency := 20
	requestsPerWorker := 15
	totalRequests := concurrency * requestsPerWorker * 3 // 3 种路由

	var wg sync.WaitGroup
	var successCount, errorCount int32
	startTime := time.Now()

	routes := []string{"/fast", "/medium", "/echo"}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestsPerWorker; j++ {
				for _, route := range routes {
					var resp *http.Response
					var err error

					if route == "/echo" {
						// POST 请求
						resp, err = http.Post("http://localhost:38904"+route, "application/json", nil)
					} else {
						// GET 请求
						resp, err = http.Get("http://localhost:38904" + route)
					}

					if err != nil {
						atomic.AddInt32(&errorCount, 1)
						continue
					}
					io.ReadAll(resp.Body)
					resp.Body.Close()
					atomic.AddInt32(&successCount, 1)
				}
			}
		}()
	}

	wg.Wait()
	duration := time.Since(startTime)
	avgLatency := duration / time.Duration(totalRequests)
	throughput := float64(totalRequests) / duration.Seconds()
	successRate := float64(successCount) / float64(totalRequests) * 100

	t.Logf("\n📊 混合路由性能:")
	t.Logf("   - 并发数: %d", concurrency)
	t.Logf("   - 路由数: %d", len(routes))
	t.Logf("   - 总请求数: %d", totalRequests)
	t.Logf("   - 成功请求: %d", successCount)
	t.Logf("   - 失败请求: %d", errorCount)
	t.Logf("   - 成功率: %.2f%%", successRate)
	t.Logf("   - 总耗时: %v", duration)
	t.Logf("   - 平均延迟: %v", avgLatency)
	t.Logf("   - 吞吐量: %.2f req/s", throughput)

	if successRate >= 99 {
		t.Logf("   ✅ 混合路由处理稳定")
	} else if successRate >= 95 {
		t.Logf("   ⚠️  混合路由处理基本稳定")
	} else {
		t.Errorf("   ❌ 混合路由处理不稳定")
	}
}

// testLongRunningStability 长时间稳定性测试
func testLongRunningStability(t *testing.T) {
	vm := goja.New()
	httpModule := builtins.NewHTTPServerModule(vm)
	vm.Set("httpserver", httpModule.GetModule())

	script := `
		const server = httpserver.createServer({
			readTimeout: 30,
			writeTimeout: 30
		});

		server.get('/stable', (req, res) => {
			res.json({ status: 'stable', timestamp: Date.now() });
		});

		server.listen('38905');
	`

	_, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("创建服务器失败: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// 长时间稳定性测试 (持续 10 秒)
	duration := 10 * time.Second
	concurrency := 10
	var successCount, errorCount int32
	var wg sync.WaitGroup
	stopChan := make(chan struct{})

	startTime := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stopChan:
					return
				default:
					resp, err := http.Get("http://localhost:38905/stable")
					if err != nil {
						atomic.AddInt32(&errorCount, 1)
						continue
					}
					io.ReadAll(resp.Body)
					resp.Body.Close()
					atomic.AddInt32(&successCount, 1)
					time.Sleep(10 * time.Millisecond) // 控制请求频率
				}
			}
		}()
	}

	// 等待测试时间
	time.Sleep(duration)
	close(stopChan)
	wg.Wait()

	actualDuration := time.Since(startTime)
	totalRequests := successCount + errorCount
	throughput := float64(totalRequests) / actualDuration.Seconds()
	successRate := float64(successCount) / float64(totalRequests) * 100

	t.Logf("\n📊 长时间稳定性测试:")
	t.Logf("   - 测试时长: %v", actualDuration)
	t.Logf("   - 并发数: %d", concurrency)
	t.Logf("   - 总请求数: %d", totalRequests)
	t.Logf("   - 成功请求: %d", successCount)
	t.Logf("   - 失败请求: %d", errorCount)
	t.Logf("   - 成功率: %.2f%%", successRate)
	t.Logf("   - 平均吞吐量: %.2f req/s", throughput)

	if successRate >= 99.9 {
		t.Logf("   🌟 长时间运行非常稳定")
	} else if successRate >= 99 {
		t.Logf("   ✅ 长时间运行稳定")
	} else if successRate >= 95 {
		t.Logf("   ⚠️  长时间运行基本稳定")
	} else {
		t.Errorf("   ❌ 长时间运行不稳定")
	}
}
