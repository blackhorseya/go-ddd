package user

//go:generate go tool mockgen -destination=mock_${GOFILE} -package=${GOPACKAGE} -source=${GOFILE}

import (
	"context"

	domain "github.com/blackhorseya/go-ddd/internal/shared/domain"
)

// Repository defines the persistence contract for the User aggregate.
// Implementations are provided by the infrastructure layer (e.g. postgres, memory).
type Repository interface {
	// Save persists a user. It should upsert (create or update).
	Save(c context.Context, user *User) error

	// FindByID retrieves a user by their unique identifier.
	// Returns an error if the user is not found.
	FindByID(c context.Context, id string) (*User, error)

	// FindByEmail retrieves a user by their email address.
	// Returns an error if no user is found with the given email.
	FindByEmail(c context.Context, email Email) (*User, error)

	// Delete removes a user by their unique identifier.
	Delete(c context.Context, id string) error

	// List retrieves a paginated list of users.
	List(c context.Context, req domain.PageRequest) (domain.PageResult[*User], error)
}
