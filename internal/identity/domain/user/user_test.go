package user

import (
	"strings"
	"testing"
)

func validEmail() Email {
	e, _ := NewEmail("test@example.com")
	return e
}

func activeUser(t *testing.T) *User {
	t.Helper()

	u, err := NewUser("u-1", validEmail(), "Test User")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if err := u.Activate(); err != nil {
		t.Fatalf("failed to activate user: %v", err)
	}

	return u
}

func TestNewUser(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		email   Email
		uname   string
		wantErr error
	}{
		{name: "valid", id: "u-1", email: validEmail(), uname: "Alice"},
		{name: "empty id", id: "", email: validEmail(), uname: "Alice", wantErr: ErrEmptyID},
		{name: "empty name", id: "u-1", email: validEmail(), uname: "", wantErr: ErrEmptyName},
		{name: "name too long", id: "u-1", email: validEmail(), uname: strings.Repeat("a", 201), wantErr: ErrNameTooLong},
		{name: "zero email", id: "u-1", email: Email{}, uname: "Alice", wantErr: ErrEmailRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := NewUser(tt.id, tt.email, tt.uname)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("NewUser() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewUser() unexpected error: %v", err)
			}

			if u.ID() != tt.id {
				t.Errorf("ID() = %q, want %q", u.ID(), tt.id)
			}

			if u.Name() != tt.uname {
				t.Errorf("Name() = %q, want %q", u.Name(), tt.uname)
			}

			if u.Status() != StatusInactive {
				t.Errorf("Status() = %q, want %q", u.Status(), StatusInactive)
			}

			if u.CreatedAt().IsZero() {
				t.Error("CreatedAt() should not be zero")
			}
		})
	}
}

func TestUser_Activate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) *User
		wantErr error
	}{
		{
			name: "from inactive",
			setup: func(t *testing.T) *User {
				t.Helper()
				u, _ := NewUser("u-1", validEmail(), "Alice")
				return u
			},
		},
		{
			name: "already active",
			setup: func(t *testing.T) *User {
				t.Helper()
				return activeUser(t)
			},
			wantErr: ErrAlreadyActive,
		},
		{
			name: "from suspended",
			setup: func(t *testing.T) *User {
				t.Helper()
				u := activeUser(t)
				_ = u.Suspend()
				return u
			},
			wantErr: ErrCannotActivate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.setup(t)
			err := u.Activate()

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Activate() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Activate() unexpected error: %v", err)
			}

			if !u.IsActive() {
				t.Error("IsActive() should be true after Activate()")
			}
		})
	}
}

func TestUser_Suspend(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) *User
		wantErr error
	}{
		{
			name: "from active",
			setup: func(t *testing.T) *User {
				t.Helper()
				return activeUser(t)
			},
		},
		{
			name: "already suspended",
			setup: func(t *testing.T) *User {
				t.Helper()
				u := activeUser(t)
				_ = u.Suspend()
				return u
			},
			wantErr: ErrAlreadySuspended,
		},
		{
			name: "from inactive",
			setup: func(t *testing.T) *User {
				t.Helper()
				u, _ := NewUser("u-1", validEmail(), "Alice")
				return u
			},
			wantErr: ErrCannotSuspend,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := tt.setup(t)
			err := u.Suspend()

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Suspend() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Suspend() unexpected error: %v", err)
			}

			if !u.IsSuspended() {
				t.Error("IsSuspended() should be true after Suspend()")
			}
		})
	}
}

func TestUser_UpdateName(t *testing.T) {
	tests := []struct {
		name    string
		newName string
		wantErr error
	}{
		{name: "valid name", newName: "Bob"},
		{name: "empty name", newName: "", wantErr: ErrEmptyName},
		{name: "too long", newName: strings.Repeat("a", 201), wantErr: ErrNameTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := NewUser("u-1", validEmail(), "Alice")
			before := u.UpdatedAt()

			err := u.UpdateName(tt.newName)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("UpdateName() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("UpdateName() unexpected error: %v", err)
			}

			if u.Name() != tt.newName {
				t.Errorf("Name() = %q, want %q", u.Name(), tt.newName)
			}

			if u.UpdatedAt().Before(before) {
				t.Error("UpdatedAt() should not be before the previous value")
			}
		})
	}
}

func TestUser_UpdateEmail(t *testing.T) {
	u, _ := NewUser("u-1", validEmail(), "Alice")
	newEmail, _ := NewEmail("new@example.com")

	u.UpdateEmail(newEmail)

	if !u.Email().Equals(newEmail) {
		t.Errorf("Email() = %v, want %v", u.Email(), newEmail)
	}
}
