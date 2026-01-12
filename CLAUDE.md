# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SW Runtime is an enterprise-grade JavaScript/TypeScript runtime built in Go, using the `goja` JavaScript engine. It provides a Node.js-like runtime environment with zero Node.js dependency, featuring a comprehensive module system, built-in modules for networking, databases, encryption, and file operations.

**Key differentiator:** This is a standalone Go binary that executes JavaScript/TypeScript without requiring Node.js, making it ideal for embedded use cases, microservices, and cross-platform scripting.

## Build Commands

```bash
# Standard build (optimized release binary)
make build
# or: make release

# Development build (with debug symbols)
make dev

# Build for all platforms (Windows/Linux/macOS, AMD64/ARM64)
make all-platforms

# Run tests
make test

# Run benchmarks
make bench

# Test coverage report
make coverage

# Install to GOPATH/bin
make install

# Clean build artifacts
make clean

# Lint and format
make lint
make fmt
```

The built binary is placed at `build/bin/sw_runtime` (or `.exe` on Windows).

## High-Level Architecture

### Core Components

```
main.go (entry point)
    └── cmd/ (Cobra CLI framework)
        ├── root.go    - Base command, global flags
        ├── run.go     - Execute JS/TS files
        ├── eval.go    - Evaluate code strings
        ├── bundle.go  - Script bundler
        ├── version.go - Version info
        └── info.go    - Runtime info

    └── internal/
        ├── runtime/   - Core runtime engine
        │   ├── runner.go     - Main script runner (Runner type)
        │   ├── eventloop.go  - Event loop for async operations
        │   └── transpiler.go - TypeScript compiler (esbuild wrapper)
        │
        ├── modules/   - Module system (CommonJS + ES6 import)
        │   ├── system.go       - Module loading, resolution, caching
        │   └── transpiler.go   - Module compilation
        │
        ├── builtins/  - Built-in JavaScript modules
        │   ├── manager.go   - Module registry (BuiltinModule interface)
        │   ├── path.go      - Path manipulation
        │   ├── fs.go        - File system operations
        │   ├── crypto.go    - Hashing, encryption, encoding
        │   ├── compression.go - Gzip/zlib compression
        │   ├── http.go      - HTTP client (http/client)
        │   ├── httpserver.go - HTTP/HTTPS server (http/server)
        │   ├── websocket.go - WebSocket client/server
        │   ├── net.go       - TCP/UDP networking
        │   ├── proxy.go     - HTTP/TCP proxy
        │   ├── redis.go     - Redis client
        │   ├── sqlite.go    - SQLite database
        │   ├── sqlite.go    - SQLite database
        │   ├── time.go      - Time utilities
        │   ├── os.go        - Operating System info
        │   ├── util.go      - Utility functions
        │   ├── process.go   - Process info and control
        │   └── exec.go      - Command execution (process/exec)
        │
        ├── pool/      - Memory pool monitoring
        └── bundler/   - Script bundler for distribution
```

### Module System Architecture

The module system is hybrid, supporting both CommonJS (`require`) and ES6 dynamic `import()`:

1. **Module Resolution** (`internal/modules/system.go`):

   - Builtin modules are checked first
   - Relative paths (`./`, `../`) resolved from current file directory
   - Absolute paths used directly
   - `node_modules/` directories searched
   - Extensions tried: `.js`, `.ts`, `.json`, then `index.js`

2. **Module Loading**:

   - TypeScript files are transpiled via esbuild before execution
   - JSON files are parsed directly
   - Modules are cached in `System.cache` (map[string]\*Module)
   - Thread-safe with `sync.RWMutex`

3. **Builtin Modules** (`internal/builtins/manager.go`):
   - Registered via `BuiltinModule` interface
   - Aliases supported (e.g., `server` → `httpserver`, `ws` → `websocket`)
   - Add new modules by implementing `BuiltinModule` and calling `RegisterModule()`

### Event Loop

The `SimpleEventLoop` in `internal/runtime/eventloop.go` handles async operations:

- `setTimeout`/`clearTimeout`
- `setInterval`/`clearInterval`
- Promise resolution
- Goroutine-based async bridging to JavaScript

### Important Concurrency Notes

**goja.Runtime is NOT concurrent-safe:** Each `Runner` creates its own `goja.Runtime` instance. The runtime is single-threaded, and async operations use goroutines that bridge back to the JS runtime via the event loop.

When working with `httpserver`, ensure each request handler doesn't block the event loop. Use async patterns with Promises for long-running operations.

## Running the CLI

```bash
# Run a TypeScript/JavaScript file
sw_runtime run app.ts
sw_runtime run app.js

# Clear module cache before running
sw_runtime run app.ts --clear-cache

# Evaluate code directly
sw_runtime eval "console.log('Hello')"

# Bundle scripts (excludes builtins)
sw_runtime bundle app.js -o dist/bundle.js --minify

# Show version/info
sw_runtime version
sw_runtime info
```

## Adding a New Builtin Module

1. Create a new file in `internal/builtins/` (e.g., `mymodule.go`)
2. Implement the `BuiltinModule` interface:

```go
package builtins

import "github.com/dop251/goja"

type MyModule struct {
    vm *goja.Runtime
}

func NewMyModule(vm *goja.Runtime) *MyModule {
    return &MyModule{vm: vm}
}

func (m *MyModule) GetModule() *goja.Object {
    obj := m.vm.NewObject()
    obj.Set("myFunction", func(call goja.FunctionCall) goja.Value {
        // Implementation
        return goja.Undefined()
    })
    return obj
}
```

3. Register in `internal/builtins/manager.go`:

```go
func (m *Manager) registerBuiltinModules() {
    // ... existing modules
    m.modules["mymodule"] = NewMyModule(m.vm)
}
```

## Module Aliases

The runtime provides Node.js-style compatibility aliases:

- `http/client` (Native HTTP Client)
- `http/server` (Native HTTP Server)
- `zlib` → `compression`
- `ws` → `websocket`
- `child_process` → `process/exec`
- `os` → `os` (Native)
- `util` → `util` (Native)
- `process` → `process` (Native)

## Testing

Tests are located in the `test/` directory. Test file naming follows Go conventions (`*_test.go`).

```bash
# Run specific test
go test ./test -run TestHTTP

# Run with coverage
go test ./test -cover

# Benchmarks
go test ./test -bench=. -benchmem
```

## Key Patterns

### Error Handling in Go→JS Bridge

```go
// Use vm.NewGoError() to convert Go errors to JavaScript
if err != nil {
    return r.vm.NewGoError(err)
}
```

### Creating Promises from Async Go Operations

```go
promise, resolve, reject := r.vm.NewPromise()
go func() {
    result, err := someAsyncOperation()
    if err != nil {
        reject(r.vm.NewGoError(err))
    } else {
        resolve(result)
    }
}()
return r.vm.ToValue(promise)
```

### Thread-Module Loading

The module cache is protected by `sync.RWMutex`. Always use locking when accessing `cache` map.

## Hot Reloading

SW Runtime supports hot reloading of JavaScript/TypeScript applications. When enabled, the runtime monitors source files for changes and automatically restarts the application.

### Usage

```bash
# Monitor file changes and reload on modification
sw_runtime run --watch app.ts
sw_runtime run -w server.js

# Combined with other options
sw_runtime run --watch --clear-cache app.js
```

### Features

- **File monitoring**: Automatically detects changes to the entry script
- **Graceful restart**: Stops the current runner and starts a new instance
- **Debounced events**: Multiple rapid changes are coalesced into a single restart
- **Signal handling**: Supports Ctrl+C for graceful shutdown

### Limitations

- Encrypted bundle files are not supported in watch mode
- Only monitors the entry file (not dependent modules)
- Application state is not preserved across restarts

### Example

See `example_hotreload.js` for a complete example with HTTP server.

## Documentation

- Main README: Comprehensive feature overview and examples
- API_REFERENCE.md: Full API documentation
- docs/BUNDLE_GUIDE.md: Bundling functionality
- docs/QUICK_START_BUNDLE.md: Quick start for bundling

## Dependencies

Key Go dependencies:

- `github.com/dop251/goja` - JavaScript engine (ECMAScript 5.1+)
- `github.com/evanw/esbuild` - TypeScript compiler
- `github.com/go-redis/redis/v8` - Redis client
- `github.com/spf13/cobra` - CLI framework
- `modernc.org/sqlite` - SQLite database driver
