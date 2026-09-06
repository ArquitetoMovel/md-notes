# Go Development Guidelines

## Project Stack

The following libraries were specified for reference in this project:

**Project Libraries**:
- **TUI Framework**: Charm Bubble Tea (v0.25.0) - Terminal user interface runtime based on The Elm Architecture - https://github.com/charmbracelet/bubbletea
- **TUI Styling**: Charm Lipgloss (v0.9.1) - Declarative style definitions and layout builder for terminal graphics - https://github.com/charmbracelet/lipgloss

**Auto-Populated Essential Tools**:
- **Testing**: Go Standard Testing (`testing`) - Built-in unit and benchmark test framework - https://pkg.go.dev/testing
- **Formatting**: `gofmt` / `go fmt` - Official standard Go source code formatter - https://pkg.go.dev/cmd/gofmt
- **Linting**: `go vet` - Official static analyzer for Go source code - https://pkg.go.dev/cmd/vet
- **Logging**: `log/slog` - Official high-performance structured logging library - https://pkg.go.dev/log/slog
- **Build Tool**: `go build` / `make` - Official compiler toolchain and automation - https://go.dev/cmd/go

> **Note**: This section lists libraries for quick reference.
> All code examples in this guideline use standard library or language-native features.
> Principles and patterns apply regardless of library choices.

---

## 1. Core Principles

### 1.1 Philosophy and Style
- Format all Go source files automatically using `gofmt` prior to commit.
- Embrace simplicity: clear, idiomatic Go beats clever abstractions and premature generalizations.
- Handle errors explicitly as values; never ignore errors or hide failures behind panic recoveries.
- Run `go vet` and static analyzers as part of every automated build pipeline.

### 1.2 Clarity over Brevity
- Variable and function names must communicate intent clearly without unnecessary verbosity.
- Scope determines name length: short names (`r`, `w`, `i`) belong in tight scopes; exported identifiers require descriptive names.
- Avoid premature optimization; prioritize clean, readable logic and profile before tuning.

---

## 2. Project Initialization

### 2.1 Creating New Project
Initialize a new Go module with a canonical module path matching the repository namespace:

```bash
# Initialize module with canonical path
go mod init md-notes

# Verify go.mod creation and toolchain version
go version
cat go.mod
```

### 2.2 Dependency Management
Manage dependencies deterministically using Go modules:

```bash
# Add a new dependency
go get github.com/charmbracelet/bubbletea@v0.25.0

# Prune unused dependencies and populate missing checksums
go mod tidy

# Verify dependency checksums against go.sum
go mod verify

# Vendor dependencies for isolated offline builds
go mod vendor
```

---

## 3. Project Structure

Adopt the standard Go application layout to clearly separate entry points, private implementation packages, and automation scripts:

```
md-notes/
├── cmd/
│   └── mdn/
│       ├── main.go            # Application entry point (flags, bootstrap)
│       └── main_test.go       # CLI acceptance and smoke tests
├── internal/                  # Private application packages (not importable externally)
│   ├── app/                   # Application orchestration and UI state model
│   ├── buffer/                # In-memory text buffer and file I/O
│   ├── config/                # Configuration parsing and schema validation
│   ├── render/                # Syntax highlighting and terminal rendering
│   └── theme/                 # Palettes and visual styling definitions
├── scripts/                   # Automation, build, and maintenance scripts
│   └── build.go               # Pure Go cross-compilation script
├── docs/                      # Technical specifications and architectural docs
├── bin/                       # Local build outputs (gitignored)
├── dist/                      # Packaged release archives and checksums (gitignored)
├── go.mod                     # Module declaration and direct dependencies
├── go.sum                     # Checksums for module verification
└── Makefile                   # Build automation shortcuts
```

### 3.1 Package Organization Rules
- Keep packages cohesive, focused on a single domain capability.
- Never create generic utility packages named `common`, `util`, or `helpers`.
- Enforce encapsulation using `internal/` for all packages not intended for public library consumption.

---

## 4. Container Development (Docker)

### 4.1 Container Philosophy
Containerized development isolates the Go compiler toolchain, CGO dependencies, and runtime environment, ensuring identical behavior across developer workstations and CI runners.

### 4.2 Docker File Structure
```
.
├── Dockerfile.dev             # Alpine-based development container definition
├── docker-compose.yaml        # Development orchestration with volume mounts
└── .dockerignore              # Files excluded from the build context
```

### 4.3 Dockerfile for Development
Use the official, minimal Alpine image pinned to the stable Go release with `sleep infinity` to maintain the container operational:

```dockerfile
FROM golang:1.24-alpine

RUN apk add --no-cache git make bash curl

WORKDIR /workspace

# Install development toolchain linters
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Keep container alive for interactive shell and compilation tasks
CMD ["sleep", "infinity"]
```

### 4.4 Docker Compose
Configure volume mounts to preserve module caches and synchronize workspace code:

```yaml
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile.dev
    container_name: md-notes-dev
    volumes:
      - .:/workspace
      - go-mod-cache:/go/pkg/mod
      - go-build-cache:/root/.cache/go-build
    working_dir: /workspace
    environment:
      - CGO_ENABLED=0
      - GOPROXY=https://proxy.golang.org,direct

volumes:
  go-mod-cache:
  go-build-cache:
```

### 4.5 .dockerignore
Exclude local binaries, git metadata, and build caches:

```
.git
.DS_Store
bin/
dist/
*.out
*.html
.cache/
```

### 4.6 Essential Commands
| Task | Command |
| :--- | :--- |
| Start environment | `docker compose up -d` |
| View container status | `docker compose ps` |
| Run application build | `docker compose exec app go build -o bin/mdn ./cmd/mdn` |
| Execute test suite | `docker compose exec app go test -v ./...` |
| Interactive shell | `docker compose exec app bash` |
| Stop environment | `docker compose down` |

### 4.7 Makefile Integration
```makefile
.PHONY: docker-up docker-down docker-shell docker-test

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-shell:
	docker compose exec app bash

docker-test:
	docker compose exec app go test -v ./...
```

### 4.8 Best Practices
- Mount both `/go/pkg/mod` and `/root/.cache/go-build` as persistent volumes to prevent re-downloading modules on container restart.
- Set `CGO_ENABLED=0` unless C bindings are strictly required.
- Pin the base image version (`golang:1.24-alpine`) to prevent unpredictable toolchain updates.

---

## 5. Naming Conventions

Follow official Go naming conventions to maintain codebase consistency:

| Category | Convention | Good Example | Bad Example |
| :--- | :--- | :--- | :--- |
| Packages | Lowercase single word, no underscores | `buffer`, `markdown` | `buf_mgr`, `commonUtil` |
| Structs / Types | PascalCase (MixedCaps) | `TextBuffer`, `FileOpener` | `text_buffer`, `Text_Buffer` |
| Interfaces | PascalCase, "-er" suffix for single-method | `Reader`, `Formatter` | `IReader`, `ReaderInterface` |
| Functions / Methods | PascalCase if exported, camelCase if unexported | `OpenFile`, `renderRow` | `open_file`, `Render_Row` |
| Getters | Omit "Get" prefix | `buffer.Length()` | `buffer.GetLength()` |
| Variables | Short in small scopes, descriptive elsewhere | `r`, `idx`, `filePath` | `theCurrentBufferIndex` |
| Constants | PascalCase or camelCase, never UPPER_CASE | `DefaultTimeout`, `maxRetries` | `DEFAULT_TIMEOUT`, `MAX_RETRIES` |
| Files | snake_case, ending in `.go` | `buffer_io.go` | `bufferIO.go`, `Buffer-Io.go` |

---

## 6. Types and Type System

### 6.1 Type Declaration
Define structs, custom type definitions, and enum-like constants with explicit types:

```go
package buffer

import "time"

// Mode represents the editing state of the editor.
type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeVisual
	ModeCommand
)

// DocumentMetadata stores file information and modification timestamps.
type DocumentMetadata struct {
	FilePath  string
	LineCount int
	ByteSize  int64
	UpdatedAt time.Time
}

// Container represents a generic collection using Go type parameters.
type Container[T any] struct {
	items []T
}

func (c *Container[T]) Push(item T) {
	c.items = append(c.items, item)
}
```

### 6.2 Type Safety and Zero Values
- Rely on the zero value of structs where possible to make instances usable immediately without constructors.
- Avoid raw `any` (`interface{}`); use strong domain types or type parameters (`[T any]`).
- Use value receivers when the struct is immutable or small; use pointer receivers when mutation or large structs are involved.

### 6.3 Allocation and Initialization
Initialize structs and slices with explicit allocation strategies:

```go
// Direct struct literal with named fields
doc := DocumentMetadata{
	FilePath:  "notes.md",
	LineCount: 1,
	ByteSize:  1024,
	UpdatedAt: time.Now().UTC(),
}

// Pre-allocate slices with make when capacity is known
lines := make([]string, 0, 128)

// Construct complex types with dedicated constructor functions
func NewContainer[T any](initialCap int) *Container[T] {
	return &Container[T]{
		items: make([]T, 0, initialCap),
	}
}
```

---

## 7. Functions and Methods

### 7.1 Signatures
Define explicit function signatures with parameter types, return values, and context propagation:

```go
package buffer

import (
	"context"
	"errors"
	"io"
	"os"
)

// ReadDocument reads the content from disk into memory up to maxBytes.
func ReadDocument(ctx context.Context, path string, maxBytes int64) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if path == "" {
		return nil, errors.New("path must not be empty")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	limitedReader := io.LimitReader(file, maxBytes)
	return io.ReadAll(limitedReader)
}
```

### 7.2 Returns and Errors
Return errors as the final return parameter. Never disguise errors with panic or magical status codes.

```go
// [GOOD] Explicit error return with clear semantics
func ParseIntSafe(s string) (int, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("parsing integer from %q: %w", s, err)
	}
	return val, nil
}

// [BAD] Magic error return value; disguises failure reason
func ParseIntBad(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return -1 // Ambiguous: is -1 an error or the actual value?
	}
	return val
}
```

### 7.3 Best Practices
- Keep parameter lists short (prefer 3 parameters or fewer). Use an options struct or functional options for complex configuration.
- Avoid hidden side effects; functions should either compute a value or perform an explicit I/O operation.
- Use named return values only when they significantly improve documentation for complex signatures.

---

## 8. Error Handling

### 8.1 Philosophy
Go treats errors as regular values. Errors are inspected, wrapped with operation context, and propagated upwards to execution boundaries.

```go
package buffer

import (
	"errors"
	"fmt"
)

// Sentinel error for known state conditions
var ErrBufferExceeded = errors.New("buffer capacity exceeded")

// ValidationError holds domain-specific error details.
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Reason)
}

// SaveDocument validates and persists the document.
func SaveDocument(path string, content []byte) error {
	if len(content) == 0 {
		return &ValidationError{Field: "content", Reason: "empty payload"}
	}
	if err := persist(path, content); err != nil {
		return fmt.Errorf("saving document to %s: %w", path, err)
	}
	return nil
}
```

### 8.2 Conventions
Check and handle errors immediately. Never discard errors with blank identifier (`_`).

```go
// [GOOD] Check error immediately and wrap with context
file, err := os.Open(filePath)
if err != nil {
	return fmt.Errorf("opening file %s for read: %w", filePath, err)
}
defer file.Close()

// [BAD] Discarding error silently, leading to silent state corruption
file, _ := os.Open(filePath) // Dangerous: file may be nil
content, _ := io.ReadAll(file) // Will panic on nil pointer
```

### 8.3 Best Practices
- Add context with `fmt.Errorf("operation failed: %w", err)` to preserve causal error chains for `errors.Is` and `errors.As`.
- Handle errors at the site of failure or propagate them up; do not log an error and also return it.
- Keep the happy path aligned to the left margin by handling errors with early return guards.

---

## 9. Concurrency and Parallelism

### 9.1 Concurrency Model
Go uses Communicating Sequential Processes (CSP). Do not communicate by sharing memory; instead, share memory by communicating via channels. Use goroutines for concurrent tasks and select statements for channel multiplexing.

```go
package worker

import (
	"context"
	"time"
)

// ProcessJob performs asynchronous work with cancellation support.
func ProcessJob(ctx context.Context, id int, out chan<- int) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Millisecond):
		select {
		case out <- id * 2:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
```

### 9.2 Synchronization
Use synchronization primitives from the `sync` package when in-memory state coordination is required:

- `sync.Mutex` / `sync.RWMutex`: Guard critical sections and shared maps/structs.
- `sync.WaitGroup`: Wait for a collection of goroutines to complete.
- `sync.Once`: Execute one-time initialization safely across multiple threads.
- `sync/atomic`: Perform lock-free primitive integer counters.

### 9.3 Best Practices
- Every goroutine must have an explicit owner and a guaranteed exit condition. Leaking goroutines causes permanent memory leaks.
- Always propagate `context.Context` to handle cancellation and deadlines across concurrent workers.
- Run tests regularly with the race detector enabled (`go test -race ./...`).

### 9.4 Common Pitfalls

```go
// [GOOD] Pass loop variable explicitly to goroutine to prevent data race
for _, item := range items {
	go func(val string) {
		process(val)
	}(item)
}

// [BAD] Capturing loop reference across concurrent goroutines in older Go versions
for _, item := range items {
	go func() {
		process(item) // Dangerous: value may mutate before execution
	}()
}
```

---

## 10. Interfaces and Abstractions

### 10.1 Interface Design
Keep interfaces small and focused. A single-method interface provides maximum composability:

```go
package storage

import "io"

// Writer abstracts file persistence.
type Writer interface {
	Write(p []byte) (n int, err error)
}

// Reader abstracts file content reading.
type Reader interface {
	Read(p []byte) (n int, err error)
}
```

### 10.2 Implementation
Interfaces are satisfied implicitly in Go. Define compile-time type assertions to guarantee compliance:

```go
package storage

import "os"

type DiskStorage struct {
	file *os.File
}

func (d *DiskStorage) Write(p []byte) (int, error) {
	return d.file.Write(p)
}

// Compile-time interface verification
var _ Writer = (*DiskStorage)(nil)
```

### 10.3 Composition
Combine smaller interfaces into composite contracts rather than declaring monolithic interfaces:

```go
// ReadWriter embeds both Reader and Writer
type ReadWriter interface {
	Reader
	Writer
}
```

---

## 11. Unit Tests

### 11.1 Structure
Organize unit tests in `_test.go` files in the same directory as the code under test. Use subtests via `t.Run` for clear isolation:

```go
package buffer_test

import (
	"testing"
)

func TestBufferBasic(t *testing.T) {
	t.Parallel()

	t.Run("creates buffer with initial capacity", func(t *testing.T) {
		buf := make([]byte, 0, 64)
		if cap(buf) != 64 {
			t.Fatalf("expected capacity 64, got %d", cap(buf))
		}
	})
}
```

### 11.2 Table-Driven Tests
Structure tests as tables with named test cases:

```go
package buffer_test

import (
	"strings"
	"testing"
)

func TestCleanPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "relative path", input: "docs/../notes.md", expected: "notes.md"},
		{name: "trailing slash", input: "notes.md/", expected: "notes.md"},
		{name: "clean path", input: "internal/buffer.go", expected: "internal/buffer.go"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := strings.TrimSuffix(strings.ReplaceAll(tc.input, "docs/../", ""), "/")
			if got != tc.expected {
				t.Errorf("CleanPath(%q) = %q; want %q", tc.input, got, tc.expected)
			}
		})
	}
}
```

### 11.3 Assertions
Use standard library control flow and `t.Errorf` or `t.Fatalf`. Prefer `t.Fatalf` when failure makes further execution invalid.

### 11.4 Essential Test Commands
| Command | Description |
| :--- | :--- |
| `go test ./...` | Run all unit tests in the module |
| `go test -v ./...` | Run all tests with verbose output |
| `go test -v -run TestBufferBasic ./...` | Run a specific test matching regex |
| `go test -race ./...` | Execute unit tests with race detection |
| `go test -coverprofile=coverage.out ./...` | Generate test coverage profile |
| `go tool cover -html=coverage.out -o coverage.html` | Inspect test coverage as HTML |

---

## 12. Mocks and Testability

### 12.1 Mock Strategies
Avoid heavy mock frameworks. Define consumer interfaces at call sites and implement lightweight stubs or fakes in test files:

```go
package app_test

import (
	"context"
	"errors"
)

// FileReaderStub implements the storage.Reader interface for tests.
type FileReaderStub struct {
	Content []byte
	Err     error
}

func (f *FileReaderStub) Read(p []byte) (int, error) {
	if f.Err != nil {
		return 0, f.Err
	}
	copy(p, f.Content)
	return len(f.Content), nil
}
```

### 12.2 Dependency Injection
Inject dependencies via constructor functions to maintain high testability:

```go
type DocumentService struct {
	reader storage.Reader
}

func NewDocumentService(r storage.Reader) *DocumentService {
	return &DocumentService{reader: r}
}
```

### 12.3 Test Doubles with Standard Library
Use `net/http/httptest` for HTTP clients, and `io.Pipe` or `bytes.Buffer` for simulating streaming I/O.

---

## 13. Integration Tests

### 13.1 Structure and Organization
Separate integration tests using Go build tags at the top of the test file:

```go
//go:build integration

package tests_test

import "testing"

func TestDatabaseIntegration(t *testing.T) {
	// Integration test logic connecting to real services
}
```

### 13.2 Selective Execution
Run unit tests and integration tests independently:

```bash
# Run unit tests only (excluding integration tags)
go test -v ./...

# Run integration tests specifically
go test -v -tags=integration ./...
```

### 13.3 Real Dependencies and Isolation
Use `t.TempDir()` to provide isolated filesystem directories for file-based tests. The directory is automatically cleaned up after test completion:

```go
func TestFileSystemPersistence(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "sample.md")
	if err := os.WriteFile(filePath, []byte("# Header"), 0600); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
}
```

---

## 14. Load and Stress Tests

### 14.1 Tools
Use Go native concurrency routines or external CLI benchmarks (`hey`, `k6`) to evaluate application limits.

### 14.2 Load Benchmarks
Execute concurrent load tests using `testing.B.RunParallel`:

```go
func BenchmarkConcurrentRead(b *testing.B) {
	data := []byte("frequent read content")
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = len(data)
		}
	})
}
```

### 14.3 Concurrency Tests
Verify absence of deadlocks and memory corruption under heavy goroutine load using `sync.WaitGroup` and race detection.

---

## 15. Profiling and Diagnostics

### 15.1 CPU and Memory Profiling
Capture profiling data using native Go profiling tools:

```bash
# Run tests and capture CPU profile
go test -cpuprofile=cpu.pprof ./internal/buffer

# Run tests and capture heap memory allocation profile
go test -memprofile=mem.pprof ./internal/buffer
```

### 15.2 Diagnostic Tools
Use standard Go toolchain commands for diagnostics:

```bash
# Analyze CPU profile interactively in CLI
go tool pprof cpu.pprof

# Launch interactive web visualizer (flamegraphs, top functions)
go tool pprof -http=:8080 mem.pprof

# Analyze execution trace
go tool trace trace.out
```

### 15.3 Performance Analysis
In the `pprof` shell, use `top 10` to identify hot spots and `list <function>` to view line-by-line machine instructions and cycle counts.

---

## 16. Benchmarks

### 16.1 Writing Benchmarks
Write benchmarks in `_test.go` files prefixed with `Benchmark`. Reset the timer after setup:

```go
package buffer_test

import (
	"bytes"
	"testing"
)

func BenchmarkBufferAppend(b *testing.B) {
	payload := []byte("line of markdown content\n")
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		buf.Write(payload)
	}
}
```

### 16.2 Sub-benchmarks
Parametrize benchmarks over variable payload sizes:

```go
func BenchmarkBufferScaling(b *testing.B) {
	sizes := []int{128, 1024, 8192}
	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			b.ReportAllocs()
			chunk := make([]byte, size)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = len(chunk)
			}
		})
	}
}
```

### 16.3 Benchmark Commands
| Command | Description |
| :--- | :--- |
| `go test -bench=. ./...` | Execute all benchmarks |
| `go test -bench=BenchmarkBufferAppend -benchmem ./...` | Run benchmark with memory allocations |
| `go test -bench=. -count=5 > old.txt` | Run multiple iterations for statistical analysis |
| `benchstat old.txt new.txt` | Compare performance differences between branches |

---

## 17. Optimization

### 17.1 Principles
Measure before optimizing. Use benchmarks (`testing.B`) and `pprof` to identify bottlenecks. Never sacrifice clarity for micro-optimizations outside verified hot paths.

### 17.2 Common Optimizations
Reduce allocations by preallocating slices and using `strings.Builder` for text concatenation:

```go
package render

import "strings"

// RenderLines concatenates lines efficiently with minimal allocations.
func RenderLines(lines []string) string {
	var builder strings.Builder
	// Preallocate estimate: avg 40 bytes per line
	builder.Grow(len(lines) * 40)

	for _, line := range lines {
		builder.WriteString(line)
		builder.WriteByte('\n')
	}
	return builder.String()
}
```

### 17.3 Memory Optimization
- Use `sync.Pool` to recycle frequently allocated temporary buffers in high-throughput loops.
- Order struct fields from largest to smallest to minimize memory padding caused by alignment rules.
- Inspect escape analysis to verify heap vs stack allocations:

```bash
# Check compiler escape analysis decisions
go build -gcflags="-m" ./internal/buffer
```

### 17.4 Basic Performance
- Avoid converting between `[]byte` and `string` inside hot loops; use `bytes` package functions directly.
- Avoid reflection (`reflect`) on critical request paths.

---

## 18. Security

### 18.1 Essential Practices
- Never commit secrets, API keys, or sensitive credentials to source control.
- Prevent directory traversal attacks by validating and sanitizing paths using `filepath.Clean`:

```go
package storage

import (
	"errors"
	"path/filepath"
	"strings"
)

// SafePath ensures the requested path remains within the sandbox root.
func SafePath(rootDir, userPath string) (string, error) {
	cleanPath := filepath.Clean(filepath.Join(rootDir, userPath))
	if !strings.HasPrefix(cleanPath, rootDir) {
		return "", errors.New("access denied: path escapes root")
	}
	return cleanPath, nil
}
```

### 18.2 Security Tools
Run official Go security analysis tools in continuous integration:

```bash
# Install and run official Go vulnerability scanner
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# Run static security linter
go run github.com/securego/gosec/v2/cmd/gosec@latest ./...
```

### 18.3 Security at API Boundaries
- Make defensive copies of mutable slices when exposing internal state.
- Bound slice reads and limit file streams with `io.LimitReader` to prevent denial-of-service via memory exhaustion.

---

## 19. Code Patterns

### 19.1 Early Return
Keep the happy path aligned to the left margin. Return early on validation errors to avoid deep indentation:

```go
// [GOOD] Early return keeps happy path aligned to left margin
func OpenDocument(path string) (*Document, error) {
	if path == "" {
		return nil, errors.New("empty path")
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}

	if info.IsDir() {
		return nil, errors.New("path is directory")
	}

	return loadDocument(path)
}

// [BAD] Deeply nested logic reduces readability
func OpenDocumentBad(path string) (*Document, error) {
	if path != "" {
		info, err := os.Stat(path)
		if err == nil {
			if !info.IsDir() {
				return loadDocument(path)
			} else {
				return nil, errors.New("path is directory")
			}
		} else {
			return nil, err
		}
	} else {
		return nil, errors.New("empty path")
	}
}
```

### 19.2 Separation of Concerns
Separate core domain logic from terminal rendering and disk I/O. Make buffer manipulations pure functions where practical to simplify testing.

### 19.3 DRY vs Premature Abstraction
Prefer a little copying over a premature abstraction. Avoid creating generic interfaces until at least two concrete use cases require polymorphism.

### 19.4 Variable Scope
Limit variable scope to the block where it is used. Use initialization statements in `if` conditions:

```go
if val, ok := cache[key]; ok {
	return val, nil
}
```

---

## 20. Dependency Management

### 20.1 Principles
- Standard library first: always prefer standard library packages (`net/http`, `os`, `encoding/json`) before adding third-party dependencies.
- Vet external dependencies for active maintenance, license compatibility (MIT, Apache 2.0, BSD), and security history.
- Pin dependency versions explicitly in `go.mod`.

### 20.2 Essential Commands
| Command | Description |
| :--- | :--- |
| `go list -m all` | List all direct and indirect dependencies |
| `go list -m -u all` | Check for available dependency updates |
| `go mod why -m <package>` | Explain why a specific dependency is imported |
| `go mod tidy` | Remove unreferenced dependencies and add missing ones |
| `go mod verify` | Verify dependencies against expected cryptographic hashes |

---

## 21. Comments and Documentation

### 21.1 Code Comments
Comment "why" a decision was made, not "what" the code does. The code itself explains the mechanics; comments record business constraints, non-obvious algorithms, or bug workarounds.

### 21.2 API Documentation
Document all exported packages, types, functions, and constants according to Go conventions:

```go
package buffer

// LineTracker maintains the mapping between byte offsets and line numbers.
// It is safe for concurrent read access after initial population.
type LineTracker struct {
	offsets []int
}

// LineAt returns the 1-indexed line number for the given byte offset.
// If offset exceeds the document bounds, it returns -1.
func (lt *LineTracker) LineAt(offset int) int {
	// Implementation logic
	return 0
}
```

### 21.3 Package Documentation
Every package should have a package-level doc comment at the top of its primary file or in a dedicated `doc.go`:

```go
// Package buffer provides in-memory text representations, undo history,
// and UTF-8 cursor navigation primitives for terminal editors.
package buffer
```

---

## 22. Database

### 22.1 Approach
Go standardizes database interactions through the `database/sql` package, which defines generic interfaces implemented by specific database drivers. Developers can choose between:

- Standard Library (`database/sql`): Direct, lightweight SQL with zero magic, explicit control, and optimal performance.
- Query Builders: Type-safe query composition without heavy runtime reflection.
- Object-Relational Mappers (ORMs): High-level entity abstractions suitable for rapid prototyping, but with potential overhead and query abstraction friction.

### 22.2 Connection and Driver
Always use `database/sql` with a registered driver. Configure connection pool parameters and timeouts explicitly:

```go
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// InitDatabase establishes a pooled database connection.
func InitDatabase(driverName, dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Configure connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(15 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	// Verify connectivity with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return db, nil
}
```

Execute queries using prepared statements or parameterized queries to eliminate SQL injection risks:

```go
// DocumentRecord represents a persisted document row.
type DocumentRecord struct {
	ID        int64
	Title     string
	Path      string
	UpdatedAt time.Time
}

// FindDocumentByPath queries a single document using parameterized binding.
func FindDocumentByPath(ctx context.Context, db *sql.DB, path string) (*DocumentRecord, error) {
	const query = `
		SELECT id, title, path, updated_at
		FROM documents
		WHERE path = $1
		LIMIT 1;
	`

	var doc DocumentRecord
	err := db.QueryRowContext(ctx, query, path).Scan(
		&doc.ID,
		&doc.Title,
		&doc.Path,
		&doc.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found: %w", err)
		}
		return nil, fmt.Errorf("executing document query: %w", err)
	}

	return &doc, nil
}
```

```go
// [GOOD] Parameterized query prevents SQL injection
query := "SELECT id, content FROM notes WHERE author = $1 AND status = $2"
rows, err := db.QueryContext(ctx, query, authorName, status)

// [BAD] Direct string concatenation leads to catastrophic SQL injection
queryBad := fmt.Sprintf("SELECT id, content FROM notes WHERE author = '%s' AND status = '%s'", authorName, status)
rowsBad, errBad := db.QueryContext(ctx, queryBad)
```

### 22.3 Migrations
Manage database schema changes with explicit versioned SQL migration files (`0001_initial_schema.up.sql`, `0001_initial_schema.down.sql`). Execute migrations during service startup or via dedicated deployment jobs using tools such as `golang-migrate` or `goose`.

### 22.4 Best Practices
- Always use parameterized placeholders (`$1`, `?`) rather than string interpolation.
- Close query results with `defer rows.Close()` and inspect `rows.Err()` after iteration.
- Manage transactional boundaries explicitly with `db.BeginTx(ctx, &sql.TxOptions{})` and defer rollbacks.

---

## 23. Logs and Observability

### 23.1 Log Levels
Go provides native structured logging via `log/slog` (introduced in Go 1.21). Four primary severity levels categorize system events:

- `slog.LevelDebug`: Fine-grained troubleshooting diagnostics.
- `slog.LevelInfo`: Standard confirmation of operational milestones.
- `slog.LevelWarn`: Unexpected occurrences that do not halt normal execution.
- `slog.LevelError`: Errors preventing normal completion of an operation.

### 23.2 Structured Logs
Configure structured logging using JSON output for machine parsing in production, and text output for human inspection during local development:

```go
package logger

import (
	"log/slog"
	"os"
)

// SetupLogger initializes the global structured logger.
func SetupLogger(isProduction bool) *slog.Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if isProduction {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}
```

### 23.3 Logging Implementation
Emit structured log records enriched with contextual key-value pairs and correlation attributes:

```go
package app

import (
	"context"
	"log/slog"
	"time"
)

// HandleFileSave records structured metrics and telemetry during file operations.
func HandleFileSave(ctx context.Context, logger *slog.Logger, path string, byteCount int) {
	start := time.Now()

	// Structured event log with contextual key-values
	logger.InfoContext(ctx, "document saved successfully",
		slog.String("file_path", path),
		slog.Int("bytes_written", byteCount),
		slog.Duration("duration_ms", time.Since(start)),
	)
}
```

### 23.4 Metrics and Observability
- Expose application health checks (`/healthz`) and readiness endpoints (`/ready`) using `net/http`.
- Instrument latency and error rates around I/O boundaries, external process invocations, and storage operations.
- Keep metric label cardinality low to prevent memory bloat in monitoring systems.

---

## 24. Golden Rules

1. **Simplicity Over Cleverness**: Write idiomatic, readable Go code. Reject premature abstractions, deep inheritance patterns, and reflection tricks.
2. **Handle Errors Explicitly**: Errors are values. Inspect errors immediately, add context with `%w`, and never swallow failures silently.
3. **Continuous Testing and Race Detection**: Write table-driven unit tests, verify concurrent execution with `go test -race`, and benchmark critical paths.
4. **Adhere to Formatting and Tooling**: Let `gofmt` eliminate stylistic debate. Run `go vet` and static analysis on every pull request.
5. **Measure Before Optimizing**: Profile with `pprof` and verify hypotheses with `testing.B` before modifying code for performance.

---

## 25. Pre-Commit Checklist

### Code Quality
- [ ] Source files formatted with `gofmt` or `go fmt ./...`.
- [ ] Static analyzer passes without warnings (`go vet ./...`).
- [ ] External linter reports no critical issues (`golangci-lint run`).
- [ ] Code compiles cleanly across targets (`go build -v ./...`).

### Tests
- [ ] All unit tests pass (`go test ./...`).
- [ ] Race detector executes cleanly (`go test -race ./...`).
- [ ] Code coverage meets or exceeds 70% threshold on core domain logic.
- [ ] Benchmarks executed and compared if performance-sensitive code was modified.

### Integrity & Security
- [ ] Errors handled explicitly and wrapped with operation context.
- [ ] Resources (files, network connections, mutexes) properly released via `defer`.
- [ ] No hardcoded passwords, private keys, or API tokens committed.
- [ ] Vulnerability scan passes without flagged issues (`govulncheck ./...`).

### Documentation
- [ ] Exported types, functions, and interfaces have complete Godoc comments.
- [ ] Package-level documentation present in `doc.go` or main package file.
- [ ] README and operational documentation updated to reflect changes.

### Docker (If Applicable)
- [ ] Docker development environment boots cleanly (`docker compose up -d`).
- [ ] Application compiles and tests pass inside the container environment.

---

## 26. References

### Official Documentation
- [The Go Programming Language](https://go.dev/): Official project documentation and specifications.
- [Effective Go](https://go.dev/doc/effective_go): Essential guide to writing clear, idiomatic Go.
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments): Common code review guidelines from the Go team.
- [Go Memory Model](https://go.dev/ref/mem): Formal specification of concurrent memory synchronization in Go.

### Essential Tools
- [gofmt](https://pkg.go.dev/cmd/gofmt): Standard Go source code formatter.
- [go vet](https://pkg.go.dev/cmd/vet): Official Go static analysis tool.
- [golangci-lint](https://golangci-lint.run/): Fast, comprehensive Go linter aggregator.
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck): Official Go vulnerability database scanner.

### Testing and Observability
- [Go Testing Package](https://pkg.go.dev/testing): Standard testing, table-driven test patterns, and benchmarks.
- [log/slog](https://pkg.go.dev/log/slog): Structured logging for Go standard library.
- [database/sql](https://pkg.go.dev/database/sql): Standard generic database connectivity.
- [pprof](https://pkg.go.dev/runtime/pprof): Performance profiling data analysis tool.

### Community & Industry Standards
- [Google Go Style Guide](https://google.github.io/styleguide/go/): Best practices from Google engineering.
- [Uber Go Style Guide](https://github.com/uber-go/guide): Opinionated conventions and common patterns from Uber.
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout): Widely adopted directory layout for Go applications.
