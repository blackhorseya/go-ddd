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
)

// User is the aggregate root for user identity.
type User struct {
	id        string
	email     Email
	name      string
	status    Status
	createdAt time.Time
	updatedAt time.Time
}

// NewUser creates a new User in inactive status.
// The caller must provide a pre-generated ID, a validated Email, and a non-empty name.
func NewUser(id string, email Email, name string) (*User, error) {
	if id == "" {
		return nil, ErrEmptyID
	}

	if email.IsZero() {
		return nil, ErrEmailRequired
	}

	if name == "" {
		return nil, ErrEmptyName
	}

	if len(name) > maxNameLength {
		return nil, ErrNameTooLong
	}

	now := time.Now()

	return &User{
		id:        id,
		email:     email,
		name:      name,
		status:    StatusInactive,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// Activate transitions the user from inactive to active.
func (u *User) Activate() error {
	if u.status == StatusActive {
		return ErrAlreadyActive
	}

	if u.status != StatusInactive {
		return ErrCannotActivate
	}

	u.status = StatusActive
	u.updatedAt = time.Now()

	return nil
}

// Suspend transitions the user from active to suspended.
func (u *User) Suspend() error {
	if u.status == StatusSuspended {
		return ErrAlreadySuspended
	}

	if u.status != StatusActive {
		return ErrCannotSuspend
	}

	u.status = StatusSuspended
	u.updatedAt = time.Now()

	return nil
}

// UpdateName changes the user's display name.
func (u *User) UpdateName(name string) error {
	if name == "" {
		return ErrEmptyName
	}

	if len(name) > maxNameLength {
		return ErrNameTooLong
	}

	u.name = name
	u.updatedAt = time.Now()

	return nil
}

// UpdateEmail changes the user's email address.
// The Email value object is already validated by its constructor.
func (u *User) UpdateEmail(email Email) {
	u.email = email
	u.updatedAt = time.Now()
}

// Getters

// ID returns the user's unique identifier.
func (u *User) ID() string { return u.id }

// Email returns the user's email address.
func (u *User) Email() Email { return u.email }

// Name returns the user's display name.
func (u *User) Name() string { return u.name }

// Status returns the user's current lifecycle status.
func (u *User) Status() Status { return u.status }

// CreatedAt returns the time the user was created.
func (u *User) CreatedAt() time.Time { return u.createdAt }

// UpdatedAt returns the time the user was last modified.
func (u *User) UpdatedAt() time.Time { return u.updatedAt }

// IsActive returns true if the user is in active status.
func (u *User) IsActive() bool { return u.status == StatusActive }

// IsSuspended returns true if the user is in suspended status.
func (u *User) IsSuspended() bool { return u.status == StatusSuspended }
