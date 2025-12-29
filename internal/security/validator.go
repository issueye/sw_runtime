package security

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"sw_runtime/internal/consts"
)

// PathValidator 路径验证器
type PathValidator struct {
	basePath string
	allowAll bool
}

// NewPathValidator 创建路径验证器
func NewPathValidator(basePath string) *PathValidator {
	return &PathValidator{
		basePath: filepath.Clean(basePath),
		allowAll: false,
	}
}

// NewPathValidatorAllowAll 创建允许所有路径的验证器（仅用于测试）
func NewPathValidatorAllowAll(basePath string) *PathValidator {
	return &PathValidator{
		basePath: filepath.Clean(basePath),
		allowAll: true,
	}
}

// Validate 验证路径是否在允许范围内
func (pv *PathValidator) Validate(path string) (string, error) {
	if pv.allowAll {
		return path, nil
	}

	// 检查路径长度
	if len(path) > consts.MaxPathLength {
		return "", fmt.Errorf("path too long: max %d characters", consts.MaxPathLength)
	}

	// 检查是否包含空字节
	if strings.ContainsAny(path, "\x00") {
		return "", fmt.Errorf("path contains null bytes")
	}

	// 转换为绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// 清理路径
	absPath = filepath.Clean(absPath)
	basePath := filepath.Clean(pv.basePath)

	// 检查是否在基础路径内
	relPath, err := filepath.Rel(basePath, absPath)
	if err != nil {
		// 在 Windows 上，跨驱动器的路径无法计算相对路径
		// 例如：D:\path 和 E:\path 之间没有相对路径
		return "", fmt.Errorf("access denied: cannot resolve path relative to base (different drive or path conflict)")
	}

	// 检查是否尝试逃逸
	if strings.HasPrefix(relPath, "..") || strings.HasPrefix(filepath.ToSlash(relPath), "../") {
		return "", fmt.Errorf("access denied: path outside sandbox (potential traversal attack)")
	}

	// 再次验证最终路径在基础路径内
	if !strings.HasPrefix(absPath+string(filepath.Separator), basePath+string(filepath.Separator)) &&
		absPath != basePath {
		// Windows 路径大小写不敏感检查
		if !strings.EqualFold(absPath+string(filepath.Separator), basePath+string(filepath.Separator)) &&
			!strings.EqualFold(absPath, basePath) {
			return "", fmt.Errorf("access denied: path outside allowed directory")
		}
	}

	return absPath, nil
}

// IsValidPath 检查路径是否有效（不返回清理后的路径）
func (pv *PathValidator) IsValidPath(path string) bool {
	_, err := pv.Validate(path)
	return err == nil
}

// URLValidator URL 验证器
type URLValidator struct {
	blockedNets     []*net.IPNet
	blockedHosts    []string
	allowedSchemes  []string
	allowPrivate    bool
	allowLoopback   bool
	allowLinkLocal  bool
}

// NewURLValidator 创建 URL 验证器
func NewURLValidator() *URLValidator {
	blockedNets := []*net.IPNet{
		parseCIDR("127.0.0.0/8"),    // Loopback
		parseCIDR("10.0.0.0/8"),     // Private Class A
		parseCIDR("172.16.0.0/12"),  // Private Class B
		parseCIDR("192.168.0.0/16"), // Private Class C
		parseCIDR("169.254.0.0/16"), // Link-local
		parseCIDR("::1/128"),        // IPv6 loopback
		parseCIDR("fc00::/7"),       // IPv6 private
		parseCIDR("fe80::/10"),      // IPv6 link-local
	}

	return &URLValidator{
		blockedNets:    blockedNets,
		blockedHosts:   []string{"localhost", "localhost.localdomain"},
		allowedSchemes: []string{"http", "https"},
		allowPrivate:   false,
		allowLoopback:  false,
		allowLinkLocal: false,
	}
}

// NewURLValidatorWithPrivate 创建允许私有网络的 URL 验证器
func NewURLValidatorWithPrivate() *URLValidator {
	uv := NewURLValidator()
	uv.allowPrivate = true
	uv.allowLoopback = true
	uv.allowLinkLocal = true
	uv.blockedNets = nil
	uv.blockedHosts = nil
	return uv
}

// parseCIDR 解析 CIDR 表示法
func parseCIDR(cidr string) *net.IPNet {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil
	}
	return network
}

// Validate 验证 URL 是否安全
func (uv *URLValidator) Validate(urlStr string) error {
	// 检查 URL 长度
	if len(urlStr) > consts.MaxURLLength {
		return fmt.Errorf("URL too long: max %d characters", consts.MaxURLLength)
	}

	// 检查是否包含空字节
	if strings.ContainsAny(urlStr, "\x00\r\n") {
		return fmt.Errorf("URL contains invalid characters")
	}

	// 检查 scheme
	schemeEnd := strings.Index(urlStr, "://")
	if schemeEnd == -1 {
		return fmt.Errorf("invalid URL: missing scheme")
	}

	scheme := urlStr[:schemeEnd]
	allowed := false
	for _, s := range uv.allowedSchemes {
		if strings.EqualFold(scheme, s) {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("unsupported scheme: %s", scheme)
	}

	// 提取主机名
	hostStart := schemeEnd + 3
	hostEnd := strings.Index(urlStr[hostStart:], "/")
	if hostEnd == -1 {
		hostEnd = len(urlStr)
	} else {
		hostEnd += hostStart
	}

	hostWithPort := urlStr[hostStart:hostEnd]
	// 移除端口
	host := hostWithPort
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		// 确保不是 IPv6 地址
		if !strings.Contains(host[:colonIndex], "]") {
			host = host[:colonIndex]
		}
	}

	// 移除 IPv6 的括号
	host = strings.Trim(host, "[]")

	// 检查是否是阻止的主机名
	for _, blocked := range uv.blockedHosts {
		if strings.EqualFold(host, blocked) {
			return fmt.Errorf("access to host '%s' is not allowed", host)
		}
	}

	// 解析 IP 地址
	ip := net.ParseIP(host)
	if ip != nil {
		return uv.validateIP(ip)
	}

	return nil
}

// validateIP 验证 IP 地址是否被允许
func (uv *URLValidator) validateIP(ip net.IP) error {
	// 检查是否在阻止的网段中
	for _, blocked := range uv.blockedNets {
		if blocked.Contains(ip) {
			return fmt.Errorf("access to IP %s is not allowed (private network)", ip.String())
		}
	}

	// 检查是否是私有地址
	if !uv.allowPrivate {
		if ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
			return fmt.Errorf("access to IP %s is not allowed (private network)", ip.String())
		}
	}

	if !uv.allowLoopback && ip.IsLoopback() {
		return fmt.Errorf("access to loopback address %s is not allowed", ip.String())
	}

	if !uv.allowLinkLocal && ip.IsLinkLocalUnicast() {
		return fmt.Errorf("access to link-local address %s is not allowed", ip.String())
	}

	return nil
}

// IsSafeURL 检查 URL 是否安全（不返回错误）
func (uv *URLValidator) IsSafeURL(urlStr string) bool {
	return uv.Validate(urlStr) == nil
}

// SetAllowedSchemes 设置允许的协议
func (uv *URLValidator) SetAllowedSchemes(schemes ...string) {
	uv.allowedSchemes = schemes
}

// AddBlockedHost 添加阻止的主机
func (uv *URLValidator) AddBlockedHost(host string) {
	uv.blockedHosts = append(uv.blockedHosts, host)
}

// AddBlockedCIDR 添加阻止的 IP 网段
func (uv *URLValidator) AddBlockedCIDR(cidr string) error {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR: %w", err)
	}
	uv.blockedNets = append(uv.blockedNets, network)
	return nil
}

// InputValidator 输入验证器
type InputValidator struct {
	maxStringLength int
	maxArrayLength  int
}

// NewInputValidator 创建输入验证器
func NewInputValidator() *InputValidator {
	return &InputValidator{
		maxStringLength: consts.MaxRequestBodySize,
		maxArrayLength:  10000,
	}
}

// ValidateString 验证字符串输入
func (iv *InputValidator) ValidateString(s string) error {
	if len(s) > iv.maxStringLength {
		return fmt.Errorf("string too long: max %d bytes", iv.maxStringLength)
	}

	// 检查是否包含空字节
	if strings.ContainsAny(s, "\x00") {
		return fmt.Errorf("string contains null bytes")
	}

	return nil
}

// ValidateStringSlice 验证字符串数组输入
func (iv *InputValidator) ValidateStringSlice(slice []string) error {
	if len(slice) > iv.maxArrayLength {
		return fmt.Errorf("array too long: max %d elements", iv.maxArrayLength)
	}

	for i, s := range slice {
		if err := iv.ValidateString(s); err != nil {
			return fmt.Errorf("element %d: %w", i, err)
		}
	}

	return nil
}
