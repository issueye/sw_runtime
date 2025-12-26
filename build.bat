@echo off
REM SW Runtime 构建脚本 (Windows Batch)
REM 用法: build.bat [选项]

setlocal enabledelayedexpansion

set PROJECT_NAME=sw_runtime
set VERSION=1.0.0
set BUILD_DIR=build
set BIN_DIR=%BUILD_DIR%\bin

REM 解析参数
set CLEAN=0
set TEST=0
set BENCH=0
set RELEASE=0
set ALL=0

:parse_args
if "%1"=="" goto end_parse_args
if /i "%1"=="clean" set CLEAN=1
if /i "%1"=="test" set TEST=1
if /i "%1"=="bench" set BENCH=1
if /i "%1"=="release" set RELEASE=1
if /i "%1"=="all" set ALL=1
if /i "%1"=="help" goto show_help
shift
goto parse_args
:end_parse_args

REM 显示帮助信息
if "%1"=="help" goto show_help

echo.
echo ========================================
echo   SW Runtime 构建系统
echo ========================================
echo   项目名称: %PROJECT_NAME%
echo   版本号:   %VERSION%
echo ========================================
echo.

REM 清理
if %CLEAN%==1 (
    echo [*] 清理构建产物...
    if exist %BUILD_DIR% (
        rmdir /s /q %BUILD_DIR%
        echo [√] 已删除 %BUILD_DIR% 目录
    )
    go clean -testcache
    echo [√] 已清理测试缓存
)

REM 检查环境
echo [*] 检查构建环境...
go version >nul 2>&1
if errorlevel 1 (
    echo [X] 未找到 Go 环境，请先安装 Go 1.24+
    exit /b 1
)
echo [√] Go 环境检查通过

REM 下载依赖
echo [*] 检查项目依赖...
go mod download
go mod verify
echo [√] 依赖检查完成

REM 运行测试
if %TEST%==1 (
    echo [*] 运行测试套件...
    go test ./test -v -timeout 30s
    if errorlevel 1 (
        echo [X] 测试失败
        exit /b 1
    )
    echo [√] 所有测试通过
)

REM 运行基准测试
if %BENCH%==1 (
    echo [*] 运行基准测试...
    go test ./test -bench=BenchmarkEventLoop -benchmem -run=^$ -benchtime=3s
    go test ./test -bench=BenchmarkRunnerAsync -benchmem -run=^$ -benchtime=3s
    echo [√] 基准测试完成
)

REM 创建构建目录
if not exist %BIN_DIR% mkdir %BIN_DIR%

REM 设置构建标志
set LDFLAGS=-s -w -X main.Version=%VERSION%

REM 构建
if %ALL%==1 (
    echo [*] 构建所有平台版本...
    
    REM Windows AMD64
    echo [*] 构建 Windows/AMD64...
    set GOOS=windows
    set GOARCH=amd64
    go build -ldflags "%LDFLAGS%" -trimpath -o %BIN_DIR%\windows-amd64\%PROJECT_NAME%.exe .
    if errorlevel 1 goto build_error
    echo [√] Windows/AMD64 构建成功
    
    REM Windows ARM64
    echo [*] 构建 Windows/ARM64...
    set GOOS=windows
    set GOARCH=arm64
    go build -ldflags "%LDFLAGS%" -trimpath -o %BIN_DIR%\windows-arm64\%PROJECT_NAME%.exe .
    if errorlevel 1 goto build_error
    echo [√] Windows/ARM64 构建成功
    
    REM Linux AMD64
    echo [*] 构建 Linux/AMD64...
    set GOOS=linux
    set GOARCH=amd64
    go build -ldflags "%LDFLAGS%" -trimpath -o %BIN_DIR%\linux-amd64\%PROJECT_NAME% .
    if errorlevel 1 goto build_error
    echo [√] Linux/AMD64 构建成功
    
    REM macOS AMD64
    echo [*] 构建 macOS/AMD64...
    set GOOS=darwin
    set GOARCH=amd64
    go build -ldflags "%LDFLAGS%" -trimpath -o %BIN_DIR%\darwin-amd64\%PROJECT_NAME% .
    if errorlevel 1 goto build_error
    echo [√] macOS/AMD64 构建成功
    
    echo [√] 所有平台构建完成
) else (
    REM 只构建当前平台
    echo [*] 构建 Windows/AMD64...
    if %RELEASE%==1 (
        go build -ldflags "%LDFLAGS%" -trimpath -o %BIN_DIR%\%PROJECT_NAME%.exe .
    ) else (
        go build -o %BIN_DIR%\%PROJECT_NAME%.exe .
    )
    if errorlevel 1 goto build_error
    echo [√] 构建成功: %BIN_DIR%\%PROJECT_NAME%.exe
)

echo.
echo [√] 构建完成！
echo.
echo 构建产物:
dir /b /s %BIN_DIR%\*.exe 2>nul
echo.

goto end

:build_error
echo [X] 构建失败
exit /b 1

:show_help
echo.
echo SW Runtime 构建脚本
echo.
echo 用法: build.bat [选项]
echo.
echo 选项:
echo   clean      清理构建产物
echo   test       运行测试
echo   bench      运行基准测试
echo   release    构建发布版本
echo   all        构建所有平台版本
echo   help       显示帮助信息
echo.
echo 示例:
echo   build.bat                # 构建当前平台版本
echo   build.bat test           # 运行测试
echo   build.bat release        # 构建优化版本
echo   build.bat clean all      # 清理后构建所有平台
echo.
goto end

:end
endlocal
