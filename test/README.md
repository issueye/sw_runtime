# SW Runtime 测试套件

这个测试套件为 SW Runtime 项目提供了全面的测试覆盖，包括单元测试、集成测试和性能基准测试。

## 测试文件结构

### 核心功能测试

- **`runner_test.go`** - 运行器核心功能测试
  - 基本 JavaScript 代码执行
  - TypeScript 支持和编译
  - 控制台输出功能
  - 全局变量管理
  - 异步操作支持
  - Promise 执行
  - 模块系统集成
  - 错误处理
  - 模块缓存管理

- **`builtins_test.go`** - 内置模块测试
  - 内置模块管理器功能
  - Path 模块功能
  - 文件系统模块
  - 加密模块
  - HTTP 模块
  - 压缩模块
  - 自定义模块注册

- **`eventloop_test.go`** - 事件循环测试
  - 基本定时器功能
  - 间隔定时器
  - 定时器清除
  - 多个定时器协调
  - 嵌套定时器
  - Promise 集成
  - 错误处理

- **`modules_test.go`** - 模块系统测试
  - 基本 require 功能
  - 文件模块加载
  - TypeScript 模块支持
  - 循环依赖处理
  - 模块缓存机制
  - 动态 import 支持

### 集成测试

- **`integration_test.go`** - 集成测试
  - 基本应用程序集成
  - 异步操作集成
  - 模块间交互
  - 错误恢复机制
  - 复杂应用程序场景

### 性能测试

- **`benchmark_test.go`** - 性能基准测试
  - 基本代码执行性能
  - TypeScript 编译性能
  - 模块加载性能
  - 异步操作性能
  - Promise 执行性能
  - 复杂计算性能
  - 对象操作性能
  - 字符串操作性能
  - 内存使用测试
  - 并发操作性能

## 运行测试

### 运行所有测试
```bash
go test ./test/...
```

### 运行特定测试文件
```bash
go test ./test/runner_test.go
go test ./test/builtins_test.go
go test ./test/eventloop_test.go
go test ./test/modules_test.go
go test ./test/integration_test.go
```

### 运行性能基准测试
```bash
go test -bench=. ./test/benchmark_test.go
```

### 运行特定的基准测试
```bash
go test -bench=BenchmarkRunnerBasicExecution ./test/benchmark_test.go
go test -bench=BenchmarkRunnerTypeScriptCompilation ./test/benchmark_test.go
```

### 详细输出
```bash
go test -v ./test/...
```

### 测试覆盖率
```bash
go test -cover ./test/...
go test -coverprofile=coverage.out ./test/...
go tool cover -html=coverage.out
```

## 测试特性

### 功能测试覆盖
- ✅ JavaScript/TypeScript 代码执行
- ✅ 模块系统（require/import）
- ✅ 异步操作（setTimeout/setInterval）
- ✅ Promise 支持
- ✅ 事件循环机制
- ✅ 内置模块功能
- ✅ 错误处理和恢复
- ✅ 模块缓存机制
- ✅ 循环依赖处理
- ✅ 动态模块加载

### 性能测试覆盖
- ✅ 代码执行速度
- ✅ TypeScript 编译性能
- ✅ 模块加载效率
- ✅ 异步操作开销
- ✅ 内存使用情况
- ✅ 并发处理能力

### 集成测试覆盖
- ✅ 完整应用程序流程
- ✅ 多模块协作
- ✅ 复杂异步场景
- ✅ 错误恢复流程
- ✅ 实际使用场景模拟

## 测试数据和临时文件

测试过程中会创建临时文件和目录，这些都会在测试完成后自动清理。测试使用 Go 的 `t.TempDir()` 来创建临时目录，确保测试环境的隔离和清洁。

## 持续集成

这些测试可以轻松集成到 CI/CD 流水线中：

```yaml
# GitHub Actions 示例
- name: Run Tests
  run: |
    go test -v ./test/...
    go test -bench=. ./test/benchmark_test.go

- name: Generate Coverage Report
  run: |
    go test -coverprofile=coverage.out ./test/...
    go tool cover -html=coverage.out -o coverage.html
```

## 贡献指南

添加新测试时请遵循以下原则：

1. **测试命名**：使用描述性的测试函数名
2. **测试隔离**：每个测试应该独立运行
3. **清理资源**：使用 `defer` 或 `t.Cleanup()` 清理资源
4. **错误处理**：使用 `t.Fatalf()` 处理致命错误，`t.Errorf()` 处理非致命错误
5. **文档**：为复杂测试添加注释说明

## 故障排除

### 常见问题

1. **测试超时**：增加异步操作的等待时间
2. **文件权限**：确保测试有权限创建临时文件
3. **模块路径**：检查模块导入路径是否正确
4. **依赖缺失**：确保所有依赖都已安装

### 调试技巧

1. 使用 `t.Logf()` 添加调试输出
2. 使用 `-v` 标志查看详细输出
3. 单独运行失败的测试进行调试
4. 检查临时文件内容（在清理前）