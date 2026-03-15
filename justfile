# feat — Justfile

# Default recipe shows available commands
default:
    @just --list

# Build the feat binary
build:
    go build -o feat ./cmd/feat

# Build and run parse on the sample manifest
run-parse: build
    ./feat parse

# Run parse with a specific manifest file
parse FILE=".feat.yml": build
    ./feat parse -f {{FILE}}

# Show feature tree
list: build
    ./feat list

# Show current status
status: build
    ./feat status

# Validate manifest
validate: build
    ./feat validate

# Clean build artifacts
clean:
    rm -f feat
    go clean

# Run go mod tidy
tidy:
    go mod tidy

# Run tests (when we have them)
test:
    go test ./...

# Format all Go files
fmt:
    go fmt ./...

# Run go vet for static analysis
vet:
    go vet ./...

# Install dependencies
deps:
    go mod download

# Show module info
info:
    go list -m all

# Run all checks
check: fmt vet test
    @echo "All checks passed!"
