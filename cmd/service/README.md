# Service Entry Point

This directory contains the main application entry points. The `main.go` file is responsible for:

1. Loading configuration
2. Setting up dependency injection
3. Initializing infrastructure components
4. Starting service endpoints (HTTP/gRPC)
5. Handling graceful shutdown

## Main Components

- Configuration loading
- Dependency injection setup
- Infrastructure initialization
- Service lifecycle management
- Signal handling for graceful shutdown

## Usage

To run the service:

```bash
go run cmd/service/main.go
```
