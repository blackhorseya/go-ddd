package user

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = bcrypt.DefaultCost

// Password errors.
var (
	ErrPasswordTooShort = errors.New("password must be at least 6 characters")
	ErrPasswordTooLong  = errors.New("password must not exceed 72 bytes")
	ErrPasswordRequired = errors.New("password is required")
)

const minPasswordLength = 6

// HashedPassword is a value object representing a bcrypt-hashed password.
// It never stores or exposes the plaintext password.
type HashedPassword struct {
	hash string
}

// NewHashedPassword creates a HashedPassword by hashing the given plaintext.
func NewHashedPassword(plain string) (HashedPassword, error) {
	if plain == "" {
		return HashedPassword{}, ErrPasswordRequired
	}

	if len(plain) < minPasswordLength {
		return HashedPassword{}, ErrPasswordTooShort
	}

	if len([]byte(plain)) > 72 {
		return HashedPassword{}, ErrPasswordTooLong
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return HashedPassword{}, err
	}

	return HashedPassword{hash: string(hashed)}, nil
}

// NewHashedPasswordFromHash reconstructs a HashedPassword from an existing hash.
// Used when loading from persistence. No validation is performed on the hash.
func NewHashedPasswordFromHash(hash string) HashedPassword {
	return HashedPassword{hash: hash}
}

// Verify checks if the given plaintext matches this hashed password.
func (x HashedPassword) Verify(plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(x.hash), []byte(plain)) == nil
}

// Hash returns the bcrypt hash string for persistence.
func (x HashedPassword) Hash() string { return x.hash }

// IsZero returns true if the password is uninitialized.
func (x HashedPassword) IsZero() bool { return x.hash == "" }
