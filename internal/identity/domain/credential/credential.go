// Package credential contains the Credential aggregate root for the Identity bounded context.
//
// The Credential aggregate manages authentication state:
//   - Private fields with getters (no public struct fields)
//   - Constructor validation ensures objects are always valid
//   - Behavior-oriented methods (Activate, Suspend) instead of setters
//   - State transition guards prevent invalid status changes
//
// Status transitions:
//
//	[NewCredential] ──► active ──► suspended
//	[Reconstitute]  ──► inactive ──► active ──► suspended
package credential

import (
	"errors"
	"time"
)

// Status represents the lifecycle state of a credential.
type Status string

const (
	// StatusInactive is the initial state for newly created credentials.
	StatusInactive Status = "inactive"

	// StatusActive indicates the credential has been activated.
	StatusActive Status = "active"

	// StatusSuspended indicates the credential has been suspended.
	StatusSuspended Status = "suspended"
)

// Domain errors for the Credential aggregate.
var (
	ErrEmptyID             = errors.New("credential id must not be empty")
	ErrEmailRequired       = errors.New("email is required")
	ErrAlreadyActive       = errors.New("credential is already active")
	ErrAlreadySuspended    = errors.New("credential is already suspended")
	ErrCannotActivate      = errors.New("only inactive credentials can be activated")
	ErrCannotSuspend       = errors.New("only active credentials can be suspended")
	ErrNotFound            = errors.New("credential not found")
	ErrEmailDuplicated     = errors.New("email already in use")
	ErrCredentialSuspended = errors.New("credential is suspended")
)

// Credential is the aggregate root for authentication identity.
type Credential struct {
	id        string
	email     Email
	password  HashedPassword
	status    Status
	createdAt time.Time
	updatedAt time.Time
}

// NewCredentialParams holds the required parameters for creating a new Credential.
type NewCredentialParams struct {
	ID       string
	Email    Email
	Password HashedPassword
}

// NewCredential creates a new Credential in active status.
func NewCredential(params NewCredentialParams) (*Credential, error) {
	if params.ID == "" {
		return nil, ErrEmptyID
	}

	if params.Email.IsZero() {
		return nil, ErrEmailRequired
	}

	if params.Password.IsZero() {
		return nil, ErrPasswordRequired
	}

	now := time.Now()

	return &Credential{
		id:        params.ID,
		email:     params.Email,
		password:  params.Password,
		status:    StatusActive,
		createdAt: now,
		updatedAt: now,
	}, nil
}

// ReconstituteCredentialParams holds all fields needed to restore a Credential from persistence.
type ReconstituteCredentialParams struct {
	ID        string
	Email     Email
	Password  HashedPassword
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ReconstituteCredential restores a Credential from persistence without applying business rules.
func ReconstituteCredential(params ReconstituteCredentialParams) (*Credential, error) {
	if params.ID == "" {
		return nil, ErrEmptyID
	}

	return &Credential{
		id:        params.ID,
		email:     params.Email,
		password:  params.Password,
		status:    params.Status,
		createdAt: params.CreatedAt,
		updatedAt: params.UpdatedAt,
	}, nil
}

// Activate transitions the credential from inactive to active.
func (x *Credential) Activate() error {
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

// Suspend transitions the credential from active to suspended.
func (x *Credential) Suspend() error {
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

// ChangePassword replaces the credential's password.
func (x *Credential) ChangePassword(password HashedPassword) error {
	if password.IsZero() {
		return ErrPasswordRequired
	}

	x.password = password
	x.updatedAt = time.Now()

	return nil
}

// UpdateEmail changes the credential's email address.
func (x *Credential) UpdateEmail(email Email) error {
	if email.IsZero() {
		return ErrEmailRequired
	}

	x.email = email
	x.updatedAt = time.Now()

	return nil
}

// Getters

// ID returns the credential's unique identifier.
func (x *Credential) ID() string { return x.id }

// Email returns the credential's email address.
func (x *Credential) Email() Email { return x.email }

// Password returns the credential's hashed password.
func (x *Credential) Password() HashedPassword { return x.password }

// Status returns the credential's current lifecycle status.
func (x *Credential) Status() Status { return x.status }

// CreatedAt returns the time the credential was created.
func (x *Credential) CreatedAt() time.Time { return x.createdAt }

// UpdatedAt returns the time the credential was last modified.
func (x *Credential) UpdatedAt() time.Time { return x.updatedAt }

// IsActive returns true if the credential is in active status.
func (x *Credential) IsActive() bool { return x.status == StatusActive }

// IsSuspended returns true if the credential is in suspended status.
func (x *Credential) IsSuspended() bool { return x.status == StatusSuspended }
