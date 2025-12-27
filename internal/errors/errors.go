package errors

import (
	"fmt"
	"runtime"

	"github.com/dop251/goja"
)

// RuntimeError 运行时错误类型
type RuntimeError struct {
	Code    string
	Message string
	File    string
	Line    int
}

// Error 实现 error 接口
func (e *RuntimeError) Error() string {
	if e.File != "" {
		return fmt.Sprintf("[%s] %s (%s:%d)", e.Code, e.Message, e.File, e.Line)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap 支持 errors.Unwrap
func (e *RuntimeError) Unwrap() error {
	return nil
}

// Is 支持错误比较
func (e *RuntimeError) Is(target error) bool {
	if t, ok := target.(*RuntimeError); ok {
		return e.Code == t.Code
	}
	return false
}

// New 创建新的运行时错误
func New(code, message string) error {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return &RuntimeError{
			Code:    code,
			Message: message,
		}
	}
	return &RuntimeError{
		Code:    code,
		Message: message,
		File:    file,
		Line:    line,
	}
}

// Wrap 包装错误并添加代码
func Wrap(code string, err error) error {
	if err == nil {
		return nil
	}
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return &RuntimeError{
			Code:    code,
			Message: err.Error(),
		}
	}
	return &RuntimeError{
		Code:    code,
		Message: err.Error(),
		File:    file,
		Line:    line,
	}
}

// Wrapf 包装错误并添加格式化消息
func Wrapf(code string, err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	message := fmt.Sprintf("%s: %s", fmt.Sprintf(format, args...), err.Error())
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return &RuntimeError{
			Code:    code,
			Message: message,
		}
	}
	return &RuntimeError{
		Code:    code,
		Message: message,
		File:    file,
		Line:    line,
	}
}

// JSError 创建标准化的 JavaScript 错误对象
func JSError(vm *goja.Runtime, code, message string) goja.Value {
	errObj := vm.NewObject()
	errObj.Set("code", code)
	errObj.Set("message", message)
	errObj.Set("name", "RuntimeError")
	return vm.ToValue(errObj)
}

// JSErrorWithStack 创建带堆栈跟踪的 JavaScript 错误对象
func JSErrorWithStack(vm *goja.Runtime, code, message string) goja.Value {
	errObj := vm.NewObject()
	errObj.Set("code", code)
	errObj.Set("message", message)
	errObj.Set("name", "RuntimeError")

	// 获取堆栈跟踪
	errObj.Set("stack", getStackTrace(2))

	return vm.ToValue(errObj)
}

// getStackTrace 获取堆栈跟踪
func getStackTrace(skip int) string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// ThrowError 抛出标准化错误到 JavaScript
func ThrowError(vm *goja.Runtime, code, message string) {
	panic(JSError(vm, code, message))
}

// ThrowErrorWithStack 抛出带堆栈的标准化错误到 JavaScript
func ThrowErrorWithStack(vm *goja.Runtime, code, message string) {
	panic(JSErrorWithStack(vm, code, message))
}

// ThrowGoError 抛出 Go 错误到 JavaScript
func ThrowGoError(vm *goja.Runtime, err error) {
	if err == nil {
		return
	}

	// 如果是 RuntimeError，提取代码和消息
	if re, ok := err.(*RuntimeError); ok {
		ThrowError(vm, re.Code, re.Message)
	}

	// 否则作为普通错误处理
	panic(vm.NewGoError(err))
}

// 错误代码常量
const (
	// 文件系统错误代码
	ErrCodeFSAccessDenied   = "FS_ACCESS_DENIED"
	ErrCodeFSNotFound       = "FS_NOT_FOUND"
	ErrCodeFSPermission     = "FS_PERMISSION_DENIED"
	ErrCodeFSAlreadyExists  = "FS_ALREADY_EXISTS"
	ErrCodeFSInvalidPath    = "FS_INVALID_PATH"
	ErrCodeFSTraversal      = "FS_PATH_TRAVERSAL"
	ErrCodeFSSandbox        = "FS_SANDBOX_VIOLATION"

	// 数据库错误代码
	ErrCodeDBQueryFailed  = "DB_QUERY_FAILED"
	ErrCodeDBConnection   = "DB_CONNECTION_FAILED"
	ErrCodeDBTxFailed     = "DB_TRANSACTION_FAILED"
	ErrCodeDBInvalidSQL   = "DB_INVALID_SQL"
	ErrCodeDBStmtFailed   = "DB_STATEMENT_FAILED"

	// HTTP 错误代码
	ErrCodeHTTPInvalidURL     = "HTTP_INVALID_URL"
	ErrCodeHTTPSSRF           = "HTTP_SSRF_BLOCKED"
	ErrCodeHTTPRequestFailed  = "HTTP_REQUEST_FAILED"
	ErrCodeHTTPTimeout        = "HTTP_TIMEOUT"
	ErrCodeHTTPInvalidHeader  = "HTTP_INVALID_HEADER"

	// 模块错误代码
	ErrCodeModuleNotFound  = "MODULE_NOT_FOUND"
	ErrCodeModuleLoadError = "MODULE_LOAD_ERROR"
	ErrCodeModuleSyntax     = "MODULE_SYNTAX_ERROR"

	// 验证错误代码
	ErrCodeValidationFailed = "VALIDATION_FAILED"
	ErrCodeInvalidInput     = "INVALID_INPUT"
	ErrCodeInvalidType      = "INVALID_TYPE"

	// 运行时错误代码
	ErrCodeRuntime      = "RUNTIME_ERROR"
	ErrCodePanic        = "PANIC"
	ErrCodeTimeout      = "TIMEOUT"
	ErrCodeCancelled    = "CANCELLED"
	ErrCodeNotAllowed   = "NOT_ALLOWED"
)

// IsFSAccessError 检查是否是文件系统访问错误
func IsFSAccessError(err error) bool {
	if re, ok := err.(*RuntimeError); ok {
		return re.Code == ErrCodeFSAccessDenied ||
			re.Code == ErrCodeFSPermission ||
			re.Code == ErrCodeFSTraversal ||
			re.Code == ErrCodeFSSandbox
	}
	return false
}

// IsNotFoundError 检查是否是未找到错误
func IsNotFoundError(err error) bool {
	if re, ok := err.(*RuntimeError); ok {
		return re.Code == ErrCodeFSNotFound ||
			re.Code == ErrCodeModuleNotFound
	}
	return false
}

// IsDBError 检查是否是数据库错误
func IsDBError(err error) bool {
	if re, ok := err.(*RuntimeError); ok {
		return re.Code == ErrCodeDBQueryFailed ||
			re.Code == ErrCodeDBConnection ||
			re.Code == ErrCodeDBTxFailed ||
			re.Code == ErrCodeDBInvalidSQL ||
			re.Code == ErrCodeDBStmtFailed
	}
	return false
}

// IsValidationError 检查是否是验证错误
func IsValidationError(err error) bool {
	if re, ok := err.(*RuntimeError); ok {
		return re.Code == ErrCodeValidationFailed ||
			re.Code == ErrCodeInvalidInput ||
			re.Code == ErrCodeInvalidType
	}
	return false
}
