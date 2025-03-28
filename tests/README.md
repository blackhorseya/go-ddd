# Testing Guide

This directory contains various types of tests for the project. We follow a comprehensive testing strategy that includes unit tests, integration tests, and end-to-end tests.

## Test Types

### Unit Tests
- Located alongside the code being tested (`*_test.go`)
- Test individual components in isolation
- Use mocks/stubs for external dependencies

### Integration Tests
- Located in `tests/integration`
- Test interaction between multiple components
- May require test containers or local services

### End-to-End Tests
- Located in `tests/e2e`
- Test complete user scenarios
- Require a full system setup

## Running Tests

```bash
# Run all tests
go test ./...

# Run specific test suite
go test ./tests/integration
go test ./tests/e2e

# Run with coverage
go test -cover ./...
```

## Writing Tests

1. Follow table-driven test patterns
2. Use meaningful test names
3. Follow the Arrange-Act-Assert pattern
4. Clean up test resources in defer statements
