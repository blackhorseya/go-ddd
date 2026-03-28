package user

import (
	"strings"
	"testing"
	"time"
)

func validEmail() Email {
	e, _ := NewEmail("test@example.com")
	return e
}

func validPassword() HashedPassword {
	p, _ := NewHashedPassword("secret123")
	return p
}

func activeUser(t *testing.T) *User {
	t.Helper()

	u, err := NewUser(NewUserParams{
		ID:       "u-1",
		Email:    validEmail(),
		Password: validPassword(),
		Name:     "Test User",
	})
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
		params  NewUserParams
		wantErr error
	}{
		{name: "valid", params: NewUserParams{ID: "u-1", Email: validEmail(), Password: validPassword(), Name: "Alice"}},
		{name: "empty id", params: NewUserParams{ID: "", Email: validEmail(), Password: validPassword(), Name: "Alice"}, wantErr: ErrEmptyID},
		{name: "empty name", params: NewUserParams{ID: "u-1", Email: validEmail(), Password: validPassword(), Name: ""}, wantErr: ErrEmptyName},
		{name: "name too long", params: NewUserParams{ID: "u-1", Email: validEmail(), Password: validPassword(), Name: strings.Repeat("a", 201)}, wantErr: ErrNameTooLong},
		{name: "zero email", params: NewUserParams{ID: "u-1", Email: Email{}, Password: validPassword(), Name: "Alice"}, wantErr: ErrEmailRequired},
		{name: "zero password", params: NewUserParams{ID: "u-1", Email: validEmail(), Password: HashedPassword{}, Name: "Alice"}, wantErr: ErrPasswordRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := NewUser(tt.params)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("NewUser() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewUser() unexpected error: %v", err)
			}

			if u.ID() != tt.params.ID {
				t.Errorf("ID() = %q, want %q", u.ID(), tt.params.ID)
			}

			if u.Name() != tt.params.Name {
				t.Errorf("Name() = %q, want %q", u.Name(), tt.params.Name)
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

func TestReconstituteUser(t *testing.T) {
	now := time.Now()

	t.Run("valid", func(t *testing.T) {
		u, err := ReconstituteUser(ReconstituteUserParams{
			ID:        "u-1",
			Email:     validEmail(),
			Password:  validPassword(),
			Name:      "Alice",
			Status:    StatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			t.Fatalf("ReconstituteUser() unexpected error: %v", err)
		}

		if u.ID() != "u-1" {
			t.Errorf("ID() = %q, want %q", u.ID(), "u-1")
		}

		if u.Status() != StatusActive {
			t.Errorf("Status() = %q, want %q", u.Status(), StatusActive)
		}

		if !u.CreatedAt().Equal(now) {
			t.Errorf("CreatedAt() = %v, want %v", u.CreatedAt(), now)
		}
	})

	t.Run("empty id", func(t *testing.T) {
		_, err := ReconstituteUser(ReconstituteUserParams{
			ID: "",
		})
		if err != ErrEmptyID {
			t.Errorf("ReconstituteUser() error = %v, want %v", err, ErrEmptyID)
		}
	})
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
				u, _ := NewUser(NewUserParams{ID: "u-1", Email: validEmail(), Password: validPassword(), Name: "Alice"})
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
				u, _ := NewUser(NewUserParams{ID: "u-1", Email: validEmail(), Password: validPassword(), Name: "Alice"})
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
			u, _ := NewUser(NewUserParams{ID: "u-1", Email: validEmail(), Password: validPassword(), Name: "Alice"})
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
	u, _ := NewUser(NewUserParams{ID: "u-1", Email: validEmail(), Password: validPassword(), Name: "Alice"})
	newEmail, _ := NewEmail("new@example.com")

	if err := u.UpdateEmail(newEmail); err != nil {
		t.Fatalf("UpdateEmail() unexpected error: %v", err)
	}

	if !u.Email().Equals(newEmail) {
		t.Errorf("Email() = %v, want %v", u.Email(), newEmail)
	}

	if err := u.UpdateEmail(Email{}); err != ErrEmailRequired {
		t.Errorf("UpdateEmail(zero) error = %v, want %v", err, ErrEmailRequired)
	}
}

func TestUser_ChangePassword(t *testing.T) {
	u, _ := NewUser(NewUserParams{ID: "u-1", Email: validEmail(), Password: validPassword(), Name: "Alice"})
	newPwd, _ := NewHashedPassword("newpass123")

	if err := u.ChangePassword(newPwd); err != nil {
		t.Fatalf("ChangePassword() unexpected error: %v", err)
	}

	if !u.Password().Verify("newpass123") {
		t.Error("Password should verify with new plaintext")
	}

	if err := u.ChangePassword(HashedPassword{}); err != ErrPasswordRequired {
		t.Errorf("ChangePassword(zero) error = %v, want %v", err, ErrPasswordRequired)
	}
}
