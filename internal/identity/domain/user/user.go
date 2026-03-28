// Package user contains the User aggregate root for the Identity bounded context.
//
// The User aggregate demonstrates core DDD patterns:
//   - Private fields with getters (no public struct fields)
//   - Constructor validation (NewUser) ensures objects are always valid
//   - Behavior-oriented methods (Activate, Suspend) instead of setters
//   - State transition guards prevent invalid status changes
//
// Status transitions:
//
//	inactive ──► active ──► suspended
package user

import (
	"errors"
	"time"
)

// Status represents the lifecycle state of a user.
type Status string

const (
	// StatusInactive is the initial state for newly created users.
	StatusInactive Status = "inactive"

	// StatusActive indicates the user has been activated.
	StatusActive Status = "active"

	// StatusSuspended indicates the user account has been suspended.
	StatusSuspended Status = "suspended"
)

const maxNameLength = 200

// Domain errors for the User aggregate.
var (
	ErrEmptyID          = errors.New("user id must not be empty")
	ErrEmptyName        = errors.New("user name must not be empty")
	ErrNameTooLong      = errors.New("user name must not exceed 200 characters")
	ErrEmailRequired    = errors.New("email is required")
	ErrAlreadyActive    = errors.New("user is already active")
	ErrAlreadySuspended = errors.New("user is already suspended")
	ErrCannotActivate   = errors.New("only inactive users can be activated")
	ErrCannotSuspend    = errors.New("only active users can be suspended")
	ErrNotFound         = errors.New("user not found")
	ErrEmailDuplicated  = errors.New("email already in use")
)

// User is the aggregate root for user identity.
type User struct {
	id        string
	email     Email
	password  HashedPassword
	name      string
	status    Status
	createdAt time.Time
	updatedAt time.Time
}

// NewUserParams holds the required parameters for creating a new User.
type NewUserParams struct {
	ID       string
	Email    Email
	Password HashedPassword
	Name     string
}

// NewUser creates a new User in inactive status.
func NewUser(params NewUserParams) (*User, error) {
	if params.ID == "" {
		return nil, ErrEmptyID
	}

	if params.Email.IsZero() {
		return nil, ErrEmailRequired
	}

	if params.Password.IsZero() {
		return nil, ErrPasswordRequired
	}

	if params.Name == "" {
		return nil, ErrEmptyName
	}

	if len(params.Name) > maxNameLength {
		return nil, ErrNameTooLong
	}

	now := time.Now()

	return &User{
		id:        params.ID,
		email:     params.Email,
		password:  params.Password,
		name:      params.Name,
		status:    StatusInactive,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// ReconstituteUserParams holds all fields needed to restore a User from persistence.
type ReconstituteUserParams struct {
	ID        string
	Email     Email
	Password  HashedPassword
	Name      string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ReconstituteUser restores a User from persistence without applying business rules.
func ReconstituteUser(params ReconstituteUserParams) (*User, error) {
	if params.ID == "" {
		return nil, ErrEmptyID
	}

	return &User{
		id:        params.ID,
		email:     params.Email,
		password:  params.Password,
		name:      params.Name,
		status:    params.Status,
		createdAt: params.CreatedAt,
		updatedAt: params.UpdatedAt,
	}, nil
}

// Activate transitions the user from inactive to active.
func (x *User) Activate() error {
	if x.status == StatusActive {
		return ErrAlreadyActive
	}

	if x.status != StatusInactive {
		return ErrCannotActivate
	}

	x.status = StatusActive
	x.updatedAt = time.Now()

	return nil
}

// Suspend transitions the user from active to suspended.
func (x *User) Suspend() error {
	if x.status == StatusSuspended {
		return ErrAlreadySuspended
	}

	if x.status != StatusActive {
		return ErrCannotSuspend
	}

	x.status = StatusSuspended
	x.updatedAt = time.Now()

	return nil
}

// UpdateName changes the user's display name.
func (x *User) UpdateName(name string) error {
	if name == "" {
		return ErrEmptyName
	}

	if len(name) > maxNameLength {
		return ErrNameTooLong
	}

	x.name = name
	x.updatedAt = time.Now()

	return nil
}

// ChangePassword replaces the user's password.
func (x *User) ChangePassword(password HashedPassword) error {
	if password.IsZero() {
		return ErrPasswordRequired
	}

	x.password = password
	x.updatedAt = time.Now()

	return nil
}

// UpdateEmail changes the user's email address.
func (x *User) UpdateEmail(email Email) error {
	if email.IsZero() {
		return ErrEmailRequired
	}

	x.email = email
	x.updatedAt = time.Now()

	return nil
}

// Getters

// ID returns the user's unique identifier.
func (x *User) ID() string { return x.id }

// Email returns the user's email address.
func (x *User) Email() Email { return x.email }

// Password returns the user's hashed password.
func (x *User) Password() HashedPassword { return x.password }

// Name returns the user's display name.
func (x *User) Name() string { return x.name }

// Status returns the user's current lifecycle status.
func (x *User) Status() Status { return x.status }

// CreatedAt returns the time the user was created.
func (x *User) CreatedAt() time.Time { return x.createdAt }

// UpdatedAt returns the time the user was last modified.
func (x *User) UpdatedAt() time.Time { return x.updatedAt }

// IsActive returns true if the user is in active status.
func (x *User) IsActive() bool { return x.status == StatusActive }

// IsSuspended returns true if the user is in suspended status.
func (x *User) IsSuspended() bool { return x.status == StatusSuspended }
