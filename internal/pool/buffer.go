package pool

import (
	"bytes"
	"sync"
)

// BufferPool 缓冲区池
type BufferPool struct {
	small  sync.Pool
	medium sync.Pool
	large  sync.Pool
}

// GlobalBufferPool 全局缓冲区池
var GlobalBufferPool = &BufferPool{
	small: sync.Pool{
		New: func() interface{} {
			return make([]byte, 4*1024) // 4KB
		},
	},
	medium: sync.Pool{
		New: func() interface{} {
			return make([]byte, 64*1024) // 64KB
		},
	},
	large: sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024*1024) // 1MB
		},
	},
}

// GetSmall 获取小缓冲区 (4KB)
func (bp *BufferPool) GetSmall() []byte {
	return bp.small.Get().([]byte)
}

// PutSmall 归还小缓冲区
func (bp *BufferPool) PutSmall(buf []byte) {
	if cap(buf) <= 4*1024 {
		bp.small.Put(buf)
	}
}

// GetMedium 获取中等缓冲区 (64KB)
func (bp *BufferPool) GetMedium() []byte {
	return bp.medium.Get().([]byte)
}

// PutMedium 归还中等缓冲区
func (bp *BufferPool) PutMedium(buf []byte) {
	if cap(buf) <= 64*1024 {
		bp.medium.Put(buf)
	}
}

// GetLarge 获取大缓冲区 (1MB)
func (bp *BufferPool) GetLarge() []byte {
	return bp.large.Get().([]byte)
}

// PutLarge 归还大缓冲区
func (bp *BufferPool) PutLarge(buf []byte) {
	if cap(buf) <= 1024*1024 {
		bp.large.Put(buf)
	}
}

// GetSize 获取指定大小的缓冲区
func (bp *BufferPool) GetSize(size int) []byte {
	switch {
	case size <= 4*1024:
		return bp.GetSmall()
	case size <= 64*1024:
		return bp.GetMedium()
	case size <= 1024*1024:
		return bp.GetLarge()
	default:
		return make([]byte, size)
	}
}

// PutSize 归还缓冲区
func (bp *BufferPool) PutSize(buf []byte, size int) {
	switch {
	case size <= 4*1024:
		bp.PutSmall(buf)
	case size <= 64*1024:
		bp.PutMedium(buf)
	case size <= 1024*1024:
		bp.PutLarge(buf)
	// 超过 1MB 的缓冲区不回收
	}
}

// ByteBufferPool bytes.Buffer 池
var ByteBufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// GetByteBuffer 获取 bytes.Buffer
func GetByteBuffer() *bytes.Buffer {
	buf := ByteBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

// PutByteBuffer 归还 bytes.Buffer
func PutByteBuffer(buf *bytes.Buffer) {
	if buf.Cap() <= 1024*1024 { // 限制回收的缓冲区大小
		ByteBufferPool.Put(buf)
	}
}

// ByteSlicePool 字节切片池（通用）
type ByteSlicePool struct {
	pool sync.Pool
}

// NewByteSlicePool 创建新的字节切片池
func NewByteSlicePool(size int) *ByteSlicePool {
	return &ByteSlicePool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, size)
			},
		},
	}
}

// Get 获取字节切片
func (bsp *ByteSlicePool) Get() []byte {
	return bsp.pool.Get().([]byte)
}

// Put 归还字节切片
func (bsp *ByteSlicePool) Put(buf []byte) {
	bsp.pool.Put(buf)
}

// ReaderPool io.Reader 限制池（用于限制读取大小）
type LimitedReaderPool struct {
	pool sync.Pool
}

// NewLimitedReaderPool 创建限制读取池
func NewLimitedReaderPool(limit int) *LimitedReaderPool {
	return &LimitedReaderPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, limit)
			},
		},
	}
}

// Get 获取缓冲区
func (lrp *LimitedReaderPool) Get() []byte {
	return lrp.pool.Get().([]byte)
}

// Put 归还缓冲区
func (lrp *LimitedReaderPool) Put(buf []byte) {
	lrp.pool.Put(buf)
}
