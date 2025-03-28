# Domain Layer

This package contains the core business logic and domain models of the application.

## Structure

- `entities/` - Domain entities and value objects
- `repositories/` - Repository interfaces
- `services/` - Domain services

## Guidelines

- Keep domain logic pure and free from external dependencies
- Use interfaces to define repository contracts
- Domain entities should enforce business rules
- Avoid external dependencies (database, HTTP, etc.)

## Example

```go
type User struct {
    ID        string
    Email     string
    CreatedAt time.Time
}

type UserRepository interface {
    FindByID(ctx context.Context, id string) (*User, error)
    Save(ctx context.Context, user *User) error
}
```
