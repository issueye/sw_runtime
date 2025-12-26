package pool

import (
	"bytes"
	"sync"
)

// Manager 对象池管理器
type Manager struct {
	// 字节缓冲池
	byteBufferPool sync.Pool

	// 字符串切片池
	stringSlicePool sync.Pool

	// 接口切片池
	interfaceSlicePool sync.Pool

	// 字节切片池
	byteSlicePool sync.Pool

	// 小对象池 (用于频繁创建的小对象)
	smallObjectPool sync.Pool
}

// GlobalManager 全局对象池管理器
var GlobalManager = NewManager()

// NewManager 创建新的对象池管理器
func NewManager() *Manager {
	pm := &Manager{}

	// 初始化字节缓冲池
	pm.byteBufferPool = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}

	// 初始化字符串切片池
	pm.stringSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 16) // 预分配16个元素
		},
	}

	// 初始化接口切片池
	pm.interfaceSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]interface{}, 0, 16) // 预分配16个元素
		},
	}

	// 初始化字节切片池
	pm.byteSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 1024) // 预分配1KB
		},
	}

	// 初始化小对象池
	pm.smallObjectPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]interface{})
		},
	}

	return pm
}

// GetByteBuffer 获取字节缓冲区
func (pm *Manager) GetByteBuffer() *bytes.Buffer {
	buf := pm.byteBufferPool.Get().(*bytes.Buffer)
	buf.Reset() // 重置缓冲区
	return buf
}

// PutByteBuffer 归还字节缓冲区
func (pm *Manager) PutByteBuffer(buf *bytes.Buffer) {
	if buf.Cap() > 64*1024 { // 如果缓冲区太大，不放回池中
		return
	}
	pm.byteBufferPool.Put(buf)
}

// GetStringSlice 获取字符串切片
func (pm *Manager) GetStringSlice() []string {
	slice := pm.stringSlicePool.Get().([]string)
	return slice[:0] // 重置长度但保留容量
}

// PutStringSlice 归还字符串切片
func (pm *Manager) PutStringSlice(slice []string) {
	if cap(slice) > 256 { // 如果切片太大，不放回池中
		return
	}
	pm.stringSlicePool.Put(slice)
}

// GetInterfaceSlice 获取接口切片
func (pm *Manager) GetInterfaceSlice() []interface{} {
	slice := pm.interfaceSlicePool.Get().([]interface{})
	return slice[:0] // 重置长度但保留容量
}

// PutInterfaceSlice 归还接口切片
func (pm *Manager) PutInterfaceSlice(slice []interface{}) {
	if cap(slice) > 256 { // 如果切片太大，不放回池中
		return
	}
	// 清空引用，避免内存泄漏
	for i := range slice {
		slice[i] = nil
	}
	pm.interfaceSlicePool.Put(slice)
}

// GetByteSlice 获取字节切片
func (pm *Manager) GetByteSlice() []byte {
	slice := pm.byteSlicePool.Get().([]byte)
	return slice[:0] // 重置长度但保留容量
}

// PutByteSlice 归还字节切片
func (pm *Manager) PutByteSlice(slice []byte) {
	if cap(slice) > 64*1024 { // 如果切片太大，不放回池中
		return
	}
	pm.byteSlicePool.Put(slice)
}

// GetSmallObject 获取小对象 (map)
func (pm *Manager) GetSmallObject() map[string]interface{} {
	obj := pm.smallObjectPool.Get().(map[string]interface{})
	// 清空 map
	for k := range obj {
		delete(obj, k)
	}
	return obj
}

// PutSmallObject 归还小对象
func (pm *Manager) PutSmallObject(obj map[string]interface{}) {
	if len(obj) > 64 { // 如果对象太大，不放回池中
		return
	}
	pm.smallObjectPool.Put(obj)
}

// Stats 获取池统计信息
type Stats struct {
	ByteBufferPoolSize     int
	StringSlicePoolSize    int
	InterfaceSlicePoolSize int
	ByteSlicePoolSize      int
	SmallObjectPoolSize    int
}

// GetStats 获取池统计信息 (用于监控)
func (pm *Manager) GetStats() Stats {
	// 注意: sync.Pool 没有提供获取当前大小的方法
	// 这里返回的是估算值，主要用于监控趋势
	return Stats{
		ByteBufferPoolSize:     0, // sync.Pool 不提供大小信息
		StringSlicePoolSize:    0,
		InterfaceSlicePoolSize: 0,
		ByteSlicePoolSize:      0,
		SmallObjectPoolSize:    0,
	}
}

// Clear 清空所有池 (用于测试或重置)
func (pm *Manager) Clear() {
	pm.byteBufferPool = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
	pm.stringSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 16)
		},
	}
	pm.interfaceSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]interface{}, 0, 16)
		},
	}
	pm.byteSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 1024)
		},
	}
	pm.smallObjectPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]interface{})
		},
	}
}
