package consts

import "time"

// 文件系统权限
const (
	FilePermReadWrite = 0644
	DirPermReadWrite  = 0755
	FilePermExclusive = 0600
	DirPermPrivate    = 0700
)

// 网络相关
const (
	DefaultHTTPTimeout  = 30 * time.Second
	DefaultReadTimeout  = 10 * time.Second
	DefaultWriteTimeout = 10 * time.Second
	DefaultIdleTimeout  = 120 * time.Second

	// HTTP 状态码
	StatusOK                  = 200
	StatusCreated             = 201
	StatusNoContent           = 204
	StatusBadRequest          = 400
	StatusUnauthorized        = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusMethodNotAllowed    = 405
	StatusRequestTimeout      = 408
	StatusInternalServerError = 500
	StatusBadGateway          = 502
	StatusServiceUnavailable  = 503
)

// 缓存大小
const (
	DefaultTimerCacheSize    = 64
	DefaultIntervalCacheSize = 32
	DefaultModuleCacheSize   = 1000
)

// WebSocket
const (
	WSReadBufferSize  = 1024
	WSWriteBufferSize = 1024
	WSMaxMessageSize  = 10 * 1024 * 1024 // 10MB
	WSReadTimeout     = 60 * time.Second
	WSWriteTimeout    = 10 * time.Second
)

// 缓冲区大小
const (
	SmallBufferSize  = 4 * 1024        // 4KB
	MediumBufferSize = 64 * 1024       // 64KB
	LargeBufferSize  = 1024 * 1024     // 1MB
	MaxBufferSize    = 10 * 1024 * 1024 // 10MB
)

// 安全相关
const (
	MaxPathLength      = 4096
	MaxURLLength       = 2083
	MaxHeaderSize      = 8192
	MaxRequestBodySize = 10 * 1024 * 1024 // 10MB
)

// 数据库相关
const (
	DefaultMaxOpenConns = 25
	DefaultMaxIdleConns = 5
	DefaultConnMaxLifetime = 5 * time.Minute
	DefaultConnMaxIdleTime = 1 * time.Minute
)

// 日志级别
const (
	LogLevelDebug = "DEBUG"
	LogLevelInfo  = "INFO"
	LogLevelWarn  = "WARN"
	LogLevelError = "ERROR"
)

// 默认值
const (
	DefaultHost = "0.0.0.0"
	DefaultPort = "8080"
)
