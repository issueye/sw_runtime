package builtins

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"io"
	"sw_runtime/internal/pool"

	"github.com/dop251/goja"
)

// CompressionModule 压缩模块
type CompressionModule struct {
	vm *goja.Runtime
}

// NewCompressionModule 创建压缩模块
func NewCompressionModule(vm *goja.Runtime) *CompressionModule {
	return &CompressionModule{vm: vm}
}

// GetModule 获取压缩模块对象
func (c *CompressionModule) GetModule() *goja.Object {
	obj := c.vm.NewObject()

	// 标准的 zlib API
	obj.Set("gzip", c.gzipCompress)
	obj.Set("gunzip", c.gzipDecompress)
	obj.Set("deflate", c.zlibCompress)
	obj.Set("inflate", c.zlibDecompress)

	// 兼容的别名
	obj.Set("gzipCompress", c.gzipCompress)
	obj.Set("gzipDecompress", c.gzipDecompress)
	obj.Set("zlibCompress", c.zlibCompress)
	obj.Set("zlibDecompress", c.zlibDecompress)

	return obj
}

// gzipCompress Gzip 压缩
func (c *CompressionModule) gzipCompress(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(c.vm.NewTypeError("gzipCompress() missing data"))
	}

	data := call.Arguments[0].String()

	// 使用对象池获取缓冲区
	buf := pool.GlobalManager.GetByteBuffer()
	defer pool.GlobalManager.PutByteBuffer(buf)

	writer := gzip.NewWriter(buf)

	_, err := writer.Write([]byte(data))
	if err != nil {
		panic(c.vm.NewGoError(err))
	}

	err = writer.Close()
	if err != nil {
		panic(c.vm.NewGoError(err))
	}

	// 返回 base64 编码的压缩数据
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return c.vm.ToValue(encoded)
}

// gzipDecompress Gzip 解压
func (c *CompressionModule) gzipDecompress(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(c.vm.NewTypeError("gzipDecompress() missing data"))
	}

	data := call.Arguments[0].String()

	// 使用对象池获取字节切片
	compressed := pool.GlobalManager.GetByteSlice()
	defer pool.GlobalManager.PutByteSlice(compressed)

	// 解码 base64
	var err error
	compressed, err = base64.StdEncoding.AppendDecode(compressed, []byte(data))
	if err != nil {
		panic(c.vm.NewGoError(err))
	}

	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		panic(c.vm.NewGoError(err))
	}
	defer reader.Close()

	// 使用对象池获取解压缓冲区
	decompressed := pool.GlobalManager.GetByteSlice()
	defer pool.GlobalManager.PutByteSlice(decompressed)

	// 读取解压数据
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			decompressed = append(decompressed, buf[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(c.vm.NewGoError(err))
		}
	}

	return c.vm.ToValue(string(decompressed))
}

// zlibCompress Zlib 压缩
func (c *CompressionModule) zlibCompress(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(c.vm.NewTypeError("zlibCompress() missing data"))
	}

	data := call.Arguments[0].String()

	// 使用对象池获取缓冲区
	buf := pool.GlobalManager.GetByteBuffer()
	defer pool.GlobalManager.PutByteBuffer(buf)

	writer := zlib.NewWriter(buf)

	_, err := writer.Write([]byte(data))
	if err != nil {
		panic(c.vm.NewGoError(err))
	}

	err = writer.Close()
	if err != nil {
		panic(c.vm.NewGoError(err))
	}

	// 返回 base64 编码的压缩数据
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return c.vm.ToValue(encoded)
}

// zlibDecompress Zlib 解压
func (c *CompressionModule) zlibDecompress(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(c.vm.NewTypeError("zlibDecompress() missing data"))
	}

	data := call.Arguments[0].String()

	// 使用对象池获取字节切片
	compressed := pool.GlobalManager.GetByteSlice()
	defer pool.GlobalManager.PutByteSlice(compressed)

	// 解码 base64
	var err error
	compressed, err = base64.StdEncoding.AppendDecode(compressed, []byte(data))
	if err != nil {
		panic(c.vm.NewGoError(err))
	}

	reader, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		panic(c.vm.NewGoError(err))
	}
	defer reader.Close()

	// 使用对象池获取解压缓冲区
	decompressed := pool.GlobalManager.GetByteSlice()
	defer pool.GlobalManager.PutByteSlice(decompressed)

	// 读取解压数据
	buf := make([]byte, 4096)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			decompressed = append(decompressed, buf[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(c.vm.NewGoError(err))
		}
	}

	return c.vm.ToValue(string(decompressed))
}
