.PHONY: build test install clean fmt lint run help bench

# Default target
all: build

# Build the binary
build:
	go build -o bin/nestjs-module-lint .

# Run tests
test:
	go test -v ./...

# Run benchmarks
bench:
	go test -bench=. -benchmem ./internal/parser ./internal/path-resolver

# Install the binary to GOPATH/bin
install:
	go install .

# Clean build artifacts
clean:
	rm -rf bin/
	go clean -cache

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run ./...

# Run the tool with test file
run: build
	./bin/nestjs-module-lint import-lint test.ts

# Run with JSON output
run-json: build
	./bin/nestjs-module-lint import-lint --json test.ts

# Run in check mode (good for CI)
check: build
	./bin/nestjs-module-lint import-lint --check test.ts

# Show help
help:
	@echo "Available targets:"
	@echo "  make build     - Build the binary"
	@echo "  make test      - Run tests"
	@echo "  make bench     - Run benchmarks"
	@echo "  make install   - Install to GOPATH/bin"
	@echo "  make clean     - Remove build artifacts"
	@echo "  make fmt       - Format code"
	@echo "  make lint      - Run linter (requires golangci-lint)"
	@echo "  make run       - Build and run with test.ts"
	@echo "  make run-json  - Build and run with JSON output"
	@echo "  make check     - Build and run in check mode (CI-friendly)"