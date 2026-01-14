package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/dop251/goja"
)

// CryptoModule 加密模块
type CryptoModule struct {
	vm *goja.Runtime
}

// NewCryptoModule 创建加密模块
func NewCryptoModule(vm *goja.Runtime) *CryptoModule {
	return &CryptoModule{vm: vm}
}

// GetModule 获取加密模块对象
func (c *CryptoModule) GetModule() *goja.Object {
	obj := c.vm.NewObject()

	// Hash 函数
	obj.Set("md5", c.md5Hash)
	obj.Set("sha1", c.sha1Hash)
	obj.Set("sha256", c.sha256Hash)
	obj.Set("sha512", c.sha512Hash)

	// Base64 编解码
	obj.Set("base64Encode", c.base64Encode)
	obj.Set("base64Decode", c.base64Decode)

	// Hex 编解码
	obj.Set("hexEncode", c.hexEncode)
	obj.Set("hexDecode", c.hexDecode)

	// AES 加解密
	obj.Set("aesEncrypt", c.aesEncrypt)
	obj.Set("aesDecrypt", c.aesDecrypt)

	// 随机数生成
	obj.Set("randomBytes", c.randomBytes)

	return obj
}

// md5Hash MD5 哈希
func (c *CryptoModule) md5Hash(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(c.vm.NewTypeError("md5() missing data"))
	}

	data := call.Arguments[0].String()
	hash := md5.Sum([]byte(data))
	return c.vm.ToValue(hex.EncodeToString(hash[:]))
}

// sha1Hash SHA1 哈希
func (c *CryptoModule) sha1Hash(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(c.vm.NewTypeError("sha1() missing data"))
	}

	data := call.Arguments[0].String()
	hash := sha1.Sum([]byte(data))
	return c.vm.ToValue(hex.EncodeToString(hash[:]))
}

// sha256Hash SHA256 哈希
func (c *CryptoModule) sha256Hash(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(c.vm.NewTypeError("sha256() missing data"))
	}

	data := call.Arguments[0].String()
	hash := sha256.Sum256([]byte(data))
	return c.vm.ToValue(hex.EncodeToString(hash[:]))
}

// sha512Hash SHA512 哈希
func (c *CryptoModule) sha512Hash(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(c.vm.NewTypeError("sha512() missing data"))
	}

	data := call.Arguments[0].String()
	hash := sha512.Sum512([]byte(data))
	return c.vm.ToValue(hex.EncodeToString(hash[:]))
}

// base64Encode Base64 编码
func (c *CryptoModule) base64Encode(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(c.vm.NewTypeError("base64Encode() missing data"))
	}

	data := call.Arguments[0].String()
	encoded := base64.StdEncoding.EncodeToString([]byte(data))
	return c.vm.ToValue(encoded)
}

// base64Decode Base64 解码
func (c *CryptoModule) base64Decode(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(c.vm.NewTypeError("base64Decode() missing data"))
	}

	data := call.Arguments[0].String()
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		panic(c.vm.NewGoError(err))
	}
	return c.vm.ToValue(string(decoded))
}

// hexEncode Hex 编码
func (c *CryptoModule) hexEncode(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(c.vm.NewTypeError("hexEncode() missing data"))
	}

	data := call.Arguments[0].String()
	encoded := hex.EncodeToString([]byte(data))
	return c.vm.ToValue(encoded)
}

// hexDecode Hex 解码
func (c *CryptoModule) hexDecode(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		panic(c.vm.NewTypeError("hexDecode() missing data"))
	}

	data := call.Arguments[0].String()
	decoded, err := hex.DecodeString(data)
	if err != nil {
		panic(c.vm.NewGoError(err))
	}
	return c.vm.ToValue(string(decoded))
}

// aesEncrypt AES 加密
func (c *CryptoModule) aesEncrypt(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(c.vm.NewTypeError("aesEncrypt() missing data or key"))
	}

	data := call.Arguments[0].String()
	key := call.Arguments[1].String()

	// 确保密钥长度为 32 字节 (AES-256)
	keyBytes := make([]byte, 32)
	copy(keyBytes, []byte(key))

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		panic(c.vm.NewGoError(err))
	}

	// 使用 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(c.vm.NewGoError(err))
	}

	// 生成随机 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic(c.vm.NewGoError(err))
	}

	// 加密
	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return c.vm.ToValue(encoded)
}

// aesDecrypt AES 解密
func (c *CryptoModule) aesDecrypt(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(c.vm.NewTypeError("aesDecrypt() missing data or key"))
	}

	data := call.Arguments[0].String()
	key := call.Arguments[1].String()

	// 解码 base64
	ciphertext, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		panic(c.vm.NewGoError(err))
	}

	// 确保密钥长度为 32 字节 (AES-256)
	keyBytes := make([]byte, 32)
	copy(keyBytes, []byte(key))

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		panic(c.vm.NewGoError(err))
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(c.vm.NewGoError(err))
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		panic(c.vm.NewGoError(fmt.Errorf("ciphertext too short")))
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(c.vm.NewGoError(err))
	}

	return c.vm.ToValue(string(plaintext))
}

// randomBytes 生成随机字节
func (c *CryptoModule) randomBytes(call goja.FunctionCall) goja.Value {
	size := 16 // 默认 16 字节
	if len(call.Arguments) > 0 {
		size = int(call.Arguments[0].ToInteger())
	}

	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		panic(c.vm.NewGoError(err))
	}

	return c.vm.ToValue(hex.EncodeToString(bytes))
}
