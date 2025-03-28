# Use Case Layer

This package implements application-specific business rules and orchestrates the flow of data between the domain layer and external interfaces.

## Structure

- `interfaces/` - Use case interfaces
- `interactors/` - Use case implementations
- `dto/` - Data Transfer Objects

## Guidelines

- Use cases should depend on domain interfaces
- Keep business flow logic here
- Transform domain objects to DTOs
- Handle transaction boundaries
- Coordinate between multiple domains if needed

## Example

```go
type CreateUserUseCase interface {
    Execute(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error)
}

type CreateUserInput struct {
    Email string
}

type CreateUserOutput struct {
    ID    string
    Email string
}
```
