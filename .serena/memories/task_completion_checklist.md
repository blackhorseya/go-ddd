# Task Completion Checklist

When completing a task in this Go DDD project, ensure you:

## Before Committing Code

### 1. Code Quality
- [ ] Run linting: `golangci-lint run`
- [ ] Fix any linting issues: `golangci-lint run --fix`
- [ ] Format code: `go fmt ./...`

### 2. Testing
- [ ] Run all tests: `go test ./...`
- [ ] Ensure all tests pass
- [ ] Add tests for new functionality
- [ ] Check test coverage if needed: `go test -cover ./...`

### 3. Dependencies
- [ ] Run `go mod tidy` to clean up dependencies
- [ ] Verify dependencies: `go mod verify`

### 4. Documentation
- [ ] Update godoc comments for exported functions/types
- [ ] Update README files if architecture/usage changes
- [ ] Generate Swagger docs if API changes: `swag init -g cmd/service/main.go`

### 5. Code Generation (if applicable)
- [ ] Regenerate mocks if interfaces changed: `go generate ./...`

## Final Checks
- [ ] Ensure code follows DDD principles (domain at center, dependencies flow inward)
- [ ] Verify no sensitive information in code
- [ ] Check that error handling is appropriate
- [ ] Confirm code follows project structure conventions