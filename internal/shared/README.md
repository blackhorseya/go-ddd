# Shared Layer

This package contains shared utilities and common code used across different layers of the application.

## Structure

- `errors/` - Custom error definitions
- `logger/` - Logging utilities
- `utils/` - Common helper functions
- `types/` - Shared types and constants

## Guidelines

- Keep shared code minimal
- Avoid circular dependencies
- Use interfaces for cross-cutting concerns
- Define common error types
- Implement shared utilities

## Example

```go
type Error struct {
    Code    string
    Message string
}

func NewLogger() *Logger {
    // Implementation
}
```
