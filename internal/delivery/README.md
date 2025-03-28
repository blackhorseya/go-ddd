# Delivery Layer

This package contains interface adapters that convert data between the format most convenient for use cases and entities, and the format most convenient for external agencies (such as HTTP or gRPC).

## Structure

- `http/` - HTTP handlers and routes
- `grpc/` - gRPC service implementations
- `middleware/` - Common middleware functions

## Guidelines

- Handle HTTP/gRPC specific logic
- Convert between DTOs and transport formats
- Implement input validation
- Handle authentication/authorization
- Define API routes and endpoints

## Example

```go
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    // Parse request
    // Call use case
    // Format response
}
```
