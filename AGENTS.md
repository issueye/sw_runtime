# AGENTS.md

This file provides guidance for AI agents working in this repository.

## Build Commands

```bash
# Standard release build (optimized binary)
make build
make release

# Development build (with debug symbols)
make dev

# Build for all platforms (Windows/Linux/macOS, AMD64/ARM64)
make all-platforms

# Run all tests
make test

# Run a single test by name
go test ./test -run TestHTTPServer -v
go test ./test -run TestEventLoop -v -timeout 30s

# Run tests with coverage
make coverage

# Run benchmarks
make bench

# Clean build artifacts
make clean

# Lint and format
make fmt
make lint

# Install to GOPATH/bin
make install

# Full dev cycle: clean, fmt, lint, test, build
make dev-cycle
```

## Code Style Guidelines

### Imports
- Group imports: stdlib first, then internal packages, then external deps
- Use the module path prefix `sw_runtime/` for internal imports
- Example:
```go
import (
    "fmt"
    "os"
    "path/filepath"
    "sync"

    "sw_runtime/internal/modules"
    "sw_runtime/internal/pool"

    "github.com/dop251/goja"
)
```

### Formatting
- Run `make fmt` before committing
- Use tabs for indentation, not spaces
- No trailing whitespace
- Put `}` on same line as closing `)` for control structures

### Types and Declarations
- Use explicit types for exported types (e.g., `type EventLoopType int`)
- Use iota for enum-like constants with comment documentation
- Use receiver names like `m` for Manager, `r` for Runner
- Interface names should be descriptive (e.g., `eventLoopInterface`)

### Naming Conventions
- **Variables**: camelCase for local, PascalCase for exported
- **Constants**: PascalCase with helpful comments
- **Files**: single lowercase word or underscore-separated (e.g., `httpserver.go`)
- **Acronyms**: Keep as-is (e.g., `HTTP`, `URL` in names)
- **Abbreviations**: Use full words when unclear

### Error Handling
- Return errors to callers; avoid `panic` except for unrecoverable states
- Use `vm.NewGoError(err)` to convert Go errors to JavaScript exceptions
- When bridging async Go operations to JS:
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

### Concurrency
- `goja.Runtime` is NOT thread-safe; each `Runner` has its own instance
- Use `sync.RWMutex` for protecting shared state like module cache
- Use goroutines for async operations that bridge back via event loop

### Module System
- Builtin modules implement `BuiltinModule` interface with `GetModule()` method
- Register modules in `internal/builtins/manager.go:registerBuiltinModules()`
- Add aliases for Node.js compatibility (e.g., `server` → `httpserver`)

### Testing
- Test files in `test/` directory, named `*_test.go`
- Use table-driven tests when testing multiple cases
- Set reasonable timeouts on tests (e.g., `-timeout 30s`)

## Key Patterns

**Creating Promises from Async Go:**
```go
promise, resolve, reject := r.vm.NewPromise()
go func() {
    if err := operation(); err != nil {
        reject(r.vm.NewGoError(err))
    } else {
        resolve(result)
    }
}()
return r.vm.ToValue(promise)
```

**Thread-Safe Module Cache:**
```go
type System struct {
    cache map[string]*Module
    mu    sync.RWMutex
}
```
