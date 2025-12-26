# SW Runtime Makefile
# 提供跨平台的构建命令

.PHONY: all build clean test bench run help install dev release

# 项目配置
PROJECT_NAME := sw_runtime
VERSION := 1.0.0
BUILD_DIR := build
BIN_DIR := $(BUILD_DIR)/bin

# Go 编译器配置
GO := go
GOFLAGS := 
LDFLAGS := -s -w -X main.Version=$(VERSION) -X main.BuildTime=$(shell date '+%Y-%m-%d_%H:%M:%S')

# 默认目标
all: build

# 显示帮助信息
help:
	@echo "SW Runtime 构建系统 - Makefile"
	@echo ""
	@echo "可用命令:"
	@echo "  make build        - 构建当前平台版本"
	@echo "  make dev          - 构建开发版本(保留调试信息)"
	@echo "  make release      - 构建发布版本(优化编译)"
	@echo "  make all-platforms - 构建所有平台版本"
	@echo "  make test         - 运行所有测试"
	@echo "  make bench        - 运行基准测试"
	@echo "  make clean        - 清理构建产物"
	@echo "  make run          - 运行示例程序"
	@echo "  make install      - 安装到 GOPATH/bin"
	@echo "  make help         - 显示此帮助信息"
	@echo ""

# 构建开发版本
dev:
	@echo "🔨 构建开发版本..."
	@mkdir -p $(BIN_DIR)
	@$(GO) build -o $(BIN_DIR)/$(PROJECT_NAME) .
	@echo "✅ 构建成功: $(BIN_DIR)/$(PROJECT_NAME)"

# 构建发布版本
build: release

release:
	@echo "🔨 构建发布版本..."
	@mkdir -p $(BIN_DIR)
	@$(GO) build -ldflags "$(LDFLAGS)" -trimpath -o $(BIN_DIR)/$(PROJECT_NAME) .
	@echo "✅ 构建成功: $(BIN_DIR)/$(PROJECT_NAME)"

# 构建所有平台
all-platforms:
	@echo "🔨 构建所有平台版本..."
	@$(MAKE) build-windows-amd64
	@$(MAKE) build-windows-arm64
	@$(MAKE) build-linux-amd64
	@$(MAKE) build-linux-arm64
	@$(MAKE) build-darwin-amd64
	@$(MAKE) build-darwin-arm64
	@echo "✅ 所有平台构建完成"

# Windows AMD64
build-windows-amd64:
	@echo "📦 构建 Windows/AMD64..."
	@mkdir -p $(BIN_DIR)/windows-amd64
	@GOOS=windows GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -trimpath -o $(BIN_DIR)/windows-amd64/$(PROJECT_NAME).exe .

# Windows ARM64
build-windows-arm64:
	@echo "📦 构建 Windows/ARM64..."
	@mkdir -p $(BIN_DIR)/windows-arm64
	@GOOS=windows GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -trimpath -o $(BIN_DIR)/windows-arm64/$(PROJECT_NAME).exe .

# Linux AMD64
build-linux-amd64:
	@echo "📦 构建 Linux/AMD64..."
	@mkdir -p $(BIN_DIR)/linux-amd64
	@GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -trimpath -o $(BIN_DIR)/linux-amd64/$(PROJECT_NAME) .

# Linux ARM64
build-linux-arm64:
	@echo "📦 构建 Linux/ARM64..."
	@mkdir -p $(BIN_DIR)/linux-arm64
	@GOOS=linux GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -trimpath -o $(BIN_DIR)/linux-arm64/$(PROJECT_NAME) .

# macOS AMD64
build-darwin-amd64:
	@echo "📦 构建 macOS/AMD64..."
	@mkdir -p $(BIN_DIR)/darwin-amd64
	@GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -trimpath -o $(BIN_DIR)/darwin-amd64/$(PROJECT_NAME) .

# macOS ARM64
build-darwin-arm64:
	@echo "📦 构建 macOS/ARM64 (Apple Silicon)..."
	@mkdir -p $(BIN_DIR)/darwin-arm64
	@GOOS=darwin GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -trimpath -o $(BIN_DIR)/darwin-arm64/$(PROJECT_NAME) .

# 运行测试
test:
	@echo "🧪 运行测试套件..."
	@$(GO) test ./test -v -timeout 30s

# 运行基准测试
bench:
	@echo "⚡ 运行基准测试..."
	@echo "事件循环性能测试..."
	@$(GO) test ./test -bench=BenchmarkEventLoop -benchmem -run=^$$ -benchtime=3s
	@echo ""
	@echo "异步操作性能测试..."
	@$(GO) test ./test -bench=BenchmarkRunnerAsync -benchmem -run=^$$ -benchtime=3s
	@echo ""
	@echo "内存优化性能测试..."
	@$(GO) test ./test -bench=BenchmarkPool -benchmem -run=^$$ -benchtime=3s

# 运行示例
run: dev
	@echo "🚀 运行示例程序..."
	@$(BIN_DIR)/$(PROJECT_NAME) examples/calculator-app.ts

# 安装到系统
install:
	@echo "📥 安装到 GOPATH/bin..."
	@$(GO) install -ldflags "$(LDFLAGS)" -trimpath .
	@echo "✅ 安装成功"

# 清理
clean:
	@echo "🧹 清理构建产物..."
	@rm -rf $(BUILD_DIR)
	@$(GO) clean -testcache
	@$(GO) clean -cache
	@echo "✅ 清理完成"

# 检查代码格式
fmt:
	@echo "🎨 格式化代码..."
	@$(GO) fmt ./...
	@echo "✅ 格式化完成"

# 代码检查
lint:
	@echo "🔍 运行代码检查..."
	@$(GO) vet ./...
	@echo "✅ 检查完成"

# 下载依赖
deps:
	@echo "📦 下载依赖..."
	@$(GO) mod download
	@$(GO) mod verify
	@echo "✅ 依赖下载完成"

# 更新依赖
update-deps:
	@echo "🔄 更新依赖..."
	@$(GO) get -u ./...
	@$(GO) mod tidy
	@echo "✅ 依赖更新完成"

# 完整的开发周期
dev-cycle: clean fmt lint test build
	@echo "✅ 开发周期完成"

# 生成测试覆盖率报告
coverage:
	@echo "📊 生成测试覆盖率报告..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) test ./test -coverprofile=$(BUILD_DIR)/coverage.out
	@$(GO) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "✅ 覆盖率报告: $(BUILD_DIR)/coverage.html"

# 性能分析
profile:
	@echo "🔬 性能分析..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) test ./test -bench=. -benchmem -cpuprofile=$(BUILD_DIR)/cpu.prof -memprofile=$(BUILD_DIR)/mem.prof
	@echo "✅ 性能分析完成"
	@echo "查看 CPU 分析: go tool pprof $(BUILD_DIR)/cpu.prof"
	@echo "查看内存分析: go tool pprof $(BUILD_DIR)/mem.prof"
