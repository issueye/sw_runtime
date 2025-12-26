package pool

import (
	"runtime"
	"sync"
	"time"
)

// MemoryStats 内存统计信息
type MemoryStats struct {
	// Go 运行时内存统计
	Alloc        uint64 // 当前分配的内存 (字节)
	TotalAlloc   uint64 // 累计分配的内存 (字节)
	Sys          uint64 // 系统内存 (字节)
	NumGC        uint32 // GC 次数
	PauseTotalNs uint64 // GC 暂停总时间 (纳秒)

	// 对象池统计
	PoolStats Stats

	// 自定义统计
	RunnerCount int64 // 活跃的 Runner 数量
	ModuleCount int64 // 加载的模块数量
	TimerCount  int64 // 活跃的定时器数量

	// 时间戳
	Timestamp time.Time
}

// MemoryMonitor 内存监控器
type MemoryMonitor struct {
	mu          sync.RWMutex
	stats       MemoryStats
	runnerCount int64
	moduleCount int64
	timerCount  int64

	// 监控配置
	enabled  bool
	interval time.Duration
	stopCh   chan struct{}

	// 历史统计 (保留最近100个数据点)
	history    []MemoryStats
	maxHistory int
}

// GlobalMemoryMonitor 全局内存监控器
var GlobalMemoryMonitor = NewMemoryMonitor()

// NewMemoryMonitor 创建内存监控器
func NewMemoryMonitor() *MemoryMonitor {
	return &MemoryMonitor{
		enabled:    false,
		interval:   time.Second * 5, // 默认5秒采集一次
		maxHistory: 100,
		history:    make([]MemoryStats, 0, 100),
		stopCh:     make(chan struct{}),
	}
}

// Start 启动内存监控
func (mm *MemoryMonitor) Start() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if mm.enabled {
		return // 已经启动
	}

	mm.enabled = true
	go mm.monitorLoop()
}

// Stop 停止内存监控
func (mm *MemoryMonitor) Stop() {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if !mm.enabled {
		return // 已经停止
	}

	mm.enabled = false
	close(mm.stopCh)
	mm.stopCh = make(chan struct{}) // 重新创建 channel 以便下次启动
}

// SetInterval 设置监控间隔
func (mm *MemoryMonitor) SetInterval(interval time.Duration) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.interval = interval
}

// monitorLoop 监控循环
func (mm *MemoryMonitor) monitorLoop() {
	ticker := time.NewTicker(mm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mm.collectStats()
		case <-mm.stopCh:
			return
		}
	}
}

// collectStats 收集统计信息
func (mm *MemoryMonitor) collectStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	mm.mu.Lock()
	defer mm.mu.Unlock()

	// 更新统计信息
	mm.stats = MemoryStats{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGC:        m.NumGC,
		PauseTotalNs: m.PauseTotalNs,
		PoolStats:    GlobalManager.GetStats(),
		RunnerCount:  mm.runnerCount,
		ModuleCount:  mm.moduleCount,
		TimerCount:   mm.timerCount,
		Timestamp:    time.Now(),
	}

	// 添加到历史记录
	mm.history = append(mm.history, mm.stats)
	if len(mm.history) > mm.maxHistory {
		// 移除最旧的记录
		copy(mm.history, mm.history[1:])
		mm.history = mm.history[:mm.maxHistory]
	}
}

// GetStats 获取当前统计信息
func (mm *MemoryMonitor) GetStats() MemoryStats {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return mm.stats
}

// GetHistory 获取历史统计信息
func (mm *MemoryMonitor) GetHistory() []MemoryStats {
	mm.mu.RLock()
	defer mm.mu.RUnlock()

	// 返回副本以避免并发问题
	history := make([]MemoryStats, len(mm.history))
	copy(history, mm.history)
	return history
}

// IncrementRunnerCount 增加 Runner 计数
func (mm *MemoryMonitor) IncrementRunnerCount() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.runnerCount++
}

// DecrementRunnerCount 减少 Runner 计数
func (mm *MemoryMonitor) DecrementRunnerCount() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	if mm.runnerCount > 0 {
		mm.runnerCount--
	}
}

// IncrementModuleCount 增加模块计数
func (mm *MemoryMonitor) IncrementModuleCount() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.moduleCount++
}

// DecrementModuleCount 减少模块计数
func (mm *MemoryMonitor) DecrementModuleCount() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	if mm.moduleCount > 0 {
		mm.moduleCount--
	}
}

// IncrementTimerCount 增加定时器计数
func (mm *MemoryMonitor) IncrementTimerCount() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.timerCount++
}

// DecrementTimerCount 减少定时器计数
func (mm *MemoryMonitor) DecrementTimerCount() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	if mm.timerCount > 0 {
		mm.timerCount--
	}
}

// ForceGC 强制执行垃圾回收
func (mm *MemoryMonitor) ForceGC() {
	runtime.GC()
	mm.collectStats() // 立即更新统计信息
}

// GetMemoryUsagePercent 获取内存使用百分比 (相对于系统内存)
func (mm *MemoryMonitor) GetMemoryUsagePercent() float64 {
	stats := mm.GetStats()
	if stats.Sys == 0 {
		return 0
	}
	return float64(stats.Alloc) / float64(stats.Sys) * 100
}

// IsMemoryPressureHigh 检查内存压力是否过高
func (mm *MemoryMonitor) IsMemoryPressureHigh() bool {
	return mm.GetMemoryUsagePercent() > 80 // 超过80%认为压力过高
}

// GetGCFrequency 获取 GC 频率 (每秒 GC 次数)
func (mm *MemoryMonitor) GetGCFrequency() float64 {
	history := mm.GetHistory()
	if len(history) < 2 {
		return 0
	}

	latest := history[len(history)-1]
	earliest := history[0]

	duration := latest.Timestamp.Sub(earliest.Timestamp).Seconds()
	if duration == 0 {
		return 0
	}

	gcCount := latest.NumGC - earliest.NumGC
	return float64(gcCount) / duration
}
