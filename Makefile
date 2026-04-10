.PHONY: help build build-all install clean test test-coverage fmt lint dev-setup run-daemon run-server

# Default target
help:
	@echo "Ravel - Containers as microVMs orchestrator"
	@echo ""
	@echo "Build targets:"
	@echo "  build-all      Build all binaries (ravel, initd, jailer)"
	@echo "  build-ravel    Build main ravel binary"
	@echo "  build-initd    Build initd binary"
	@echo "  build-jailer   Build jailer binary"
	@echo ""
	@echo "Install targets:"
	@echo "  install        Install all binaries to /usr/bin and /opt/ravel"
	@echo "  install-ravel  Install only ravel binary"
	@echo "  uninstall      Remove installed binaries"
	@echo ""
	@echo "Development targets:"
	@echo "  dev-setup      Setup development environment"
	@echo "  run-daemon     Run daemon in debug mode"
	@echo "  run-server     Run server in debug mode"
	@echo ""
	@echo "Testing targets:"
	@echo "  test           Run all tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  test-verbose   Run tests with verbose output"
	@echo ""
	@echo "Code quality targets:"
	@echo "  fmt            Format code with gofmt"
	@echo "  lint           Run golangci-lint"
	@echo "  vet            Run go vet"
	@echo ""
	@echo "Utility targets:"
	@echo "  clean          Remove built binaries"
	@echo "  deps           Download dependencies"
	@echo "  tidy           Tidy go modules"

# Build targets
build-all: build-ravel build-initd build-jailer
	@echo "All binaries built successfully"

build-ravel:
	@echo "Building ravel..."
	@CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/ravel cmd/ravel/ravel.go

build-initd:
	@echo "Building initd..."
	@CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/initd cmd/initd/initd.go

build-jailer:
	@echo "Building jailer..."
	@CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/jailer cmd/jailer/jailer.go

# Install targets
install: build-all
	@echo "Installing binaries..."
	@sudo mkdir -p /opt/ravel
	@sudo cp ./bin/ravel /usr/bin/ravel
	@sudo cp ./bin/initd /opt/ravel/initd
	@sudo cp ./bin/jailer /opt/ravel/jailer
	@sudo chmod +x /usr/bin/ravel
	@sudo chmod +x /opt/ravel/initd
	@sudo chmod +x /opt/ravel/jailer
	@echo "Installation complete"

install-ravel: build-ravel
	@echo "Installing ravel binary..."
	@sudo cp ./bin/ravel /usr/bin/ravel
	@sudo chmod +x /usr/bin/ravel
	@echo "Ravel installed to /usr/bin/ravel"

uninstall:
	@echo "Uninstalling Ravel..."
	@sudo rm -f /usr/bin/ravel
	@sudo rm -rf /opt/ravel
	@echo "Uninstall complete"

# Development targets
dev-setup:
	@echo "Setting up development environment..."
	@go mod download
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development environment ready"

run-daemon:
	@echo "Starting daemon in debug mode..."
	@sudo go run cmd/ravel/ravel.go daemon -c ravel.toml --debug

run-server:
	@echo "Starting server in debug mode..."
	@go run cmd/ravel/ravel.go server -c ravel.toml --debug

# Testing targets
test:
	@echo "Running tests..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-verbose:
	@echo "Running tests with verbose output..."
	@go test -v -race ./...

test-short:
	@echo "Running short tests..."
	@go test -short ./...

# Code quality targets
fmt:
	@echo "Formatting code..."
	@gofmt -s -w .
	@echo "Code formatted"

lint:
	@echo "Running linter..."
	@golangci-lint run ./...

vet:
	@echo "Running go vet..."
	@go vet ./...

check: fmt vet lint
	@echo "All checks passed"

# Utility targets
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@echo "Dependencies downloaded"

tidy:
	@echo "Tidying go modules..."
	@go mod tidy
	@echo "Modules tidied"

# Legacy aliases for compatibility
run-raveld: run-daemon
run-api: run-server