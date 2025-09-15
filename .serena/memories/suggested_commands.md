# Development Commands for Go DDD Project

## Running the Application
```bash
# Run the main service
go run cmd/service/main.go
```

## Testing
```bash
# Run all tests
go test ./...

# Run specific test suite
go test ./tests/integration
go test ./tests/e2e

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run a single test
go test -run TestName ./path/to/package
```

## Linting and Code Quality
```bash
# Install golangci-lint (if not installed)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linting
golangci-lint run

# Run linting with auto-fix
golangci-lint run --fix
```

## Code Generation
```bash
# Generate mocks
go generate ./...

# Generate Swagger documentation
swag init -g cmd/service/main.go
```

## Dependency Management
```bash
# Download dependencies
go mod download

# Tidy up dependencies
go mod tidy

# Verify dependencies
go mod verify

# Update dependencies
go get -u ./...
```

## Build
```bash
# Build the application
go build -o bin/service cmd/service/main.go

# Build with race detector
go build -race -o bin/service cmd/service/main.go
```

## Format Code
```bash
# Format all Go files
go fmt ./...

# Use gofmt with simplification
gofmt -s -w .
```