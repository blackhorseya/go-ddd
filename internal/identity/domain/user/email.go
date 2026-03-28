package user

import (
	"errors"
	"regexp"
	"strings"
)

// ErrInvalidEmail is returned when an email address is not valid.
var ErrInvalidEmail = errors.New("invalid email address")

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email is a value object representing a validated email address.
// It normalizes the address to lowercase on construction.
type Email struct {
	address string
}

// NewEmail creates a new Email after validating the format.
func NewEmail(address string) (Email, error) {
	normalized := strings.TrimSpace(strings.ToLower(address))
	if !emailRegex.MatchString(normalized) {
		return Email{}, ErrInvalidEmail
	}

	return Email{address: normalized}, nil
}

// Address returns the normalized email address.
func (e Email) Address() string { return e.address }

// String implements fmt.Stringer.
func (e Email) String() string { return e.address }

// Equals checks if two emails are the same.
func (e Email) Equals(other Email) bool { return e.address == other.address }

// IsZero returns true if the email is uninitialized.
func (e Email) IsZero() bool { return e.address == "" }
