package credential

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
func (x Email) Address() string { return x.address }

// String implements fmt.Stringer.
func (x Email) String() string { return x.address }

// Equals checks if two emails are the same.
func (x Email) Equals(other Email) bool { return x.address == other.address }

// IsZero returns true if the email is uninitialized.
func (x Email) IsZero() bool { return x.address == "" }
