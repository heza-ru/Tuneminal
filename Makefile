# Tuneminal Makefile

.PHONY: build run clean test install deps demo

# Build the application
build:
	go build -o tuneminal cmd/tuneminal/main.go

# Run the application
run: build
	./tuneminal

# Clean build artifacts
clean:
	rm -f tuneminal
	go clean

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Install dependencies
deps:
	go mod tidy
	go mod download

# Build for release (optimized)
build-release:
	go build -ldflags="-s -w" -o tuneminal cmd/tuneminal/main.go

# Install the binary to GOPATH/bin
install: build
	go install ./cmd/tuneminal

# Run the demo
demo: build
	@echo "Starting Tuneminal demo..."
	@echo "Make sure you have audio files in uploads/demo/ directory"
	./tuneminal

# Development mode with auto-reload (requires entr)
dev:
	@echo "Watching for changes..."
	find . -name "*.go" | entr -r make run

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Check for security issues
security:
	gosec ./...

# Generate documentation
docs:
	godoc -http=:6060

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  run           - Build and run the application"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  deps          - Install dependencies"
	@echo "  build-release - Build optimized release version"
	@echo "  install       - Install binary to GOPATH/bin"
	@echo "  demo          - Run the demo"
	@echo "  dev           - Development mode with auto-reload"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  security      - Check for security issues"
	@echo "  docs          - Generate documentation"
	@echo "  help          - Show this help"


