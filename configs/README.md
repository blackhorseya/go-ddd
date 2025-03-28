# Configuration

This directory contains configuration management for the application.

## Structure

- `config.go` - Configuration structs and loading logic
- `config.yaml` - Default configuration file (create as needed)

## Usage

```go
config, err := config.Load("configs/config.yaml")
if err != nil {
    log.Fatal(err)
}
```

## Configuration Options

The configuration supports:

- Server settings (host, port)
- Database connection details
- Additional settings as needed

## Environment Variables

Configuration can be overridden using environment variables:

- `APP_SERVER_HOST`
- `APP_SERVER_PORT`
- `APP_DB_HOST`
- `APP_DB_PORT`
- `APP_DB_NAME`
- `APP_DB_USER`
- `APP_DB_PASSWORD`
