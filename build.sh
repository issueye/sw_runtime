#!/bin/bash
# SW Runtime 构建脚本 (Bash)
# 用法: ./build.sh [选项]

set -e  # 遇到错误立即退出

# 项目配置
PROJECT_NAME="sw_runtime"
VERSION="1.0.0"
BUILD_DIR="build"
BIN_DIR="$BUILD_DIR/bin"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 输出函数
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# 显示帮助信息
show_help() {
    cat << EOF
SW Runtime 构建脚本

用法: ./build.sh [选项]

选项:
  clean       清理构建产物和缓存
  test        运行所有测试
  bench       运行基准测试
  release     构建发布版本(优化编译)
  all         构建所有平台版本
  help        显示此帮助信息

示例:
  ./build.sh                # 构建当前平台版本
  ./build.sh test           # 运行测试
  ./build.sh release        # 构建优化版本
  ./build.sh clean all      # 清理后构建所有平台

EOF
}

# 清理函数
clean_build() {
    log_info "清理构建产物..."
    
    if [ -d "$BUILD_DIR" ]; then
        rm -rf "$BUILD_DIR"
        log_success "已删除 $BUILD_DIR 目录"
    fi
    
    # 清理测试缓存
    go clean -testcache
    log_success "已清理测试缓存"
}

# 检查 Go 环境
check_environment() {
    log_info "检查构建环境..."
    
    # 检查 Go 版本
    if ! command -v go &> /dev/null; then
        log_error "未找到 Go 环境，请先安装 Go 1.24+"
        exit 1
    fi
    
    go_version=$(go version)
    log_success "Go 环境: $go_version"
    
    # 检查依赖
    log_info "检查项目依赖..."
    go mod download
    go mod verify
    log_success "依赖检查完成"
}

# 运行测试
run_tests() {
    log_info "运行测试套件..."
    
    export GO_TEST_TIMEOUT="30s"
    
    # 运行所有测试
    if go test ./test -v -timeout 30s; then
        log_success "所有测试通过"
    else
        log_error "测试失败"
        exit 1
    fi
}

# 运行基准测试
run_benchmarks() {
    log_info "运行基准测试..."
    
    # 运行事件循环基准测试
    log_info "事件循环性能测试..."
    go test ./test -bench=BenchmarkEventLoop -benchmem -run=^$ -benchtime=3s
    
    # 运行异步操作基准测试
    log_info "异步操作性能测试..."
    go test ./test -bench=BenchmarkRunnerAsync -benchmem -run=^$ -benchtime=3s
    
    # 运行内存优化基准测试
    log_info "内存优化性能测试..."
    go test ./test -bench=BenchmarkPool -benchmem -run=^$ -benchtime=3s
    
    log_success "基准测试完成"
}

# 构建单个平台
build_platform() {
    local os=$1
    local arch=$2
    local is_release=$3
    
    local output_name="$PROJECT_NAME"
    if [ "$os" = "windows" ]; then
        output_name="${output_name}.exe"
    fi
    
    local output_path="$BIN_DIR/$os-$arch/$output_name"
    
    log_info "构建 $os/$arch..."
    
    # 创建输出目录
    mkdir -p "$(dirname "$output_path")"
    
    # 设置环境变量
    export GOOS=$os
    export GOARCH=$arch
    export CGO_ENABLED=1
    
    # 构建标志
    local ldflags="-s -w -X main.Version=$VERSION -X main.BuildTime=$(date '+%Y-%m-%d_%H:%M:%S')"
    
    if [ "$is_release" = "true" ]; then
        # 发布版本：优化编译
        go build -ldflags "$ldflags" -trimpath -o "$output_path" .
    else
        # 开发版本：保留调试信息
        go build -o "$output_path" .
    fi
    
    if [ $? -eq 0 ]; then
        local file_size=$(du -h "$output_path" | cut -f1)
        log_success "构建成功: $output_path ($file_size)"
    else
        log_error "构建失败: $os/$arch"
        exit 1
    fi
}

# 构建所有平台
build_all_platforms() {
    local is_release=$1
    
    log_info "构建所有平台版本..."
    
    # 定义平台列表
    declare -a platforms=(
        "windows:amd64"
        "windows:arm64"
        "linux:amd64"
        "linux:arm64"
        "darwin:amd64"
        "darwin:arm64"
    )
    
    for platform in "${platforms[@]}"; do
        IFS=':' read -r os arch <<< "$platform"
        build_platform "$os" "$arch" "$is_release"
    done
    
    log_success "所有平台构建完成"
}

# 构建当前平台
build_current() {
    local is_release=$1
    
    local current_os=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$current_os" in
        darwin) current_os="darwin" ;;
        linux) current_os="linux" ;;
        mingw*|cygwin*|msys*) current_os="windows" ;;
    esac
    
    local current_arch=$(uname -m)
    case "$current_arch" in
        x86_64) current_arch="amd64" ;;
        aarch64|arm64) current_arch="arm64" ;;
        armv7l) current_arch="arm" ;;
        i386|i686) current_arch="386" ;;
    esac
    
    build_platform "$current_os" "$current_arch" "$is_release"
}

# 创建发布包
create_release_package() {
    log_info "创建发布包..."
    
    local release_dir="$BUILD_DIR/release"
    mkdir -p "$release_dir"
    
    # 压缩每个平台的构建产物
    for platform_dir in "$BIN_DIR"/*; do
        if [ -d "$platform_dir" ]; then
            local platform_name=$(basename "$platform_dir")
            local archive_name="$release_dir/${PROJECT_NAME}_${VERSION}_${platform_name}.tar.gz"
            
            log_info "打包 $platform_name..."
            tar -czf "$archive_name" -C "$platform_dir" .
            
            local archive_size=$(du -h "$archive_name" | cut -f1)
            log_success "已创建: $archive_name ($archive_size)"
        fi
    done
    
    log_success "发布包创建完成: $release_dir"
}

# 显示构建信息
show_build_info() {
    cat << EOF

========================================
  SW Runtime 构建系统
========================================
  项目名称: $PROJECT_NAME
  版本号:   $VERSION
  构建时间: $(date '+%Y-%m-%d %H:%M:%S')
========================================

EOF
}

# ============ 主流程 ============

# 解析参数
DO_CLEAN=false
DO_TEST=false
DO_BENCH=false
DO_RELEASE=false
DO_ALL=false

for arg in "$@"; do
    case "$arg" in
        clean)
            DO_CLEAN=true
            ;;
        test)
            DO_TEST=true
            ;;
        bench)
            DO_BENCH=true
            ;;
        release)
            DO_RELEASE=true
            ;;
        all)
            DO_ALL=true
            ;;
        help)
            show_help
            exit 0
            ;;
        *)
            log_error "未知选项: $arg"
            show_help
            exit 1
            ;;
    esac
done

# 显示构建信息
show_build_info

# 清理
if [ "$DO_CLEAN" = true ]; then
    clean_build
fi

# 检查环境
check_environment

# 运行测试
if [ "$DO_TEST" = true ]; then
    run_tests
fi

# 运行基准测试
if [ "$DO_BENCH" = true ]; then
    run_benchmarks
fi

# 构建
if [ "$DO_ALL" = true ]; then
    # 构建所有平台
    build_all_platforms "$DO_RELEASE"
    if [ "$DO_RELEASE" = true ]; then
        create_release_package
    fi
else
    # 只构建当前平台
    build_current "$DO_RELEASE"
fi

echo ""
log_success "构建完成！"
echo ""

# 显示构建产物
if [ -d "$BIN_DIR" ]; then
    log_info "构建产物:"
    find "$BIN_DIR" -type f -executable -o -name "*.exe" | while read -r file; do
        echo "  📦 $file"
    done
fi
