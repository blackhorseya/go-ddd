package credential

import (
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

func activeCredential(t *testing.T) *Credential {
	t.Helper()

	c, err := NewCredential(NewCredentialParams{
		ID:       "c-1",
		Email:    validEmail(),
		Password: validPassword(),
	})
	if err != nil {
		t.Fatalf("failed to create credential: %v", err)
	}

	return c
}

func inactiveCredential(t *testing.T) *Credential {
	t.Helper()

	c, err := ReconstituteCredential(ReconstituteCredentialParams{
		ID:       "c-1",
		Email:    validEmail(),
		Password: validPassword(),
		Status:   StatusInactive,
	})
	if err != nil {
		t.Fatalf("failed to reconstitute credential: %v", err)
	}

	return c
}

func TestNewCredential(t *testing.T) {
	tests := []struct {
		name    string
		params  NewCredentialParams
		wantErr error
	}{
		{name: "valid", params: NewCredentialParams{ID: "c-1", Email: validEmail(), Password: validPassword()}},
		{name: "empty id", params: NewCredentialParams{ID: "", Email: validEmail(), Password: validPassword()}, wantErr: ErrEmptyID},
		{name: "zero email", params: NewCredentialParams{ID: "c-1", Email: Email{}, Password: validPassword()}, wantErr: ErrEmailRequired},
		{name: "zero password", params: NewCredentialParams{ID: "c-1", Email: validEmail(), Password: HashedPassword{}}, wantErr: ErrPasswordRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewCredential(tt.params)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("NewCredential() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewCredential() unexpected error: %v", err)
			}

			if c.ID() != tt.params.ID {
				t.Errorf("ID() = %q, want %q", c.ID(), tt.params.ID)
			}

			if c.Status() != StatusActive {
				t.Errorf("Status() = %q, want %q", c.Status(), StatusActive)
			}

			if c.CreatedAt().IsZero() {
				t.Error("CreatedAt() should not be zero")
			}
		})
	}
}

func TestReconstituteCredential(t *testing.T) {
	now := time.Now()

	t.Run("valid", func(t *testing.T) {
		c, err := ReconstituteCredential(ReconstituteCredentialParams{
			ID:        "c-1",
			Email:     validEmail(),
			Password:  validPassword(),
			Status:    StatusActive,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err != nil {
			t.Fatalf("ReconstituteCredential() unexpected error: %v", err)
		}

		if c.Status() != StatusActive {
			t.Errorf("Status() = %q, want %q", c.Status(), StatusActive)
		}
	})

	t.Run("empty id", func(t *testing.T) {
		_, err := ReconstituteCredential(ReconstituteCredentialParams{ID: ""})
		if err != ErrEmptyID {
			t.Errorf("error = %v, want %v", err, ErrEmptyID)
		}
	})
}

func TestCredential_Activate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) *Credential
		wantErr error
	}{
		{
			name:  "from inactive",
			setup: func(t *testing.T) *Credential { t.Helper(); return inactiveCredential(t) },
		},
		{
			name:    "already active",
			setup:   func(t *testing.T) *Credential { t.Helper(); return activeCredential(t) },
			wantErr: ErrAlreadyActive,
		},
		{
			name: "from suspended",
			setup: func(t *testing.T) *Credential {
				t.Helper()
				c := activeCredential(t)
				_ = c.Suspend()
				return c
			},
			wantErr: ErrCannotActivate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup(t)
			err := c.Activate()

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Activate() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Activate() unexpected error: %v", err)
			}

			if !c.IsActive() {
				t.Error("IsActive() should be true")
			}
		})
	}
}

func TestCredential_Suspend(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) *Credential
		wantErr error
	}{
		{
			name:  "from active",
			setup: func(t *testing.T) *Credential { t.Helper(); return activeCredential(t) },
		},
		{
			name: "already suspended",
			setup: func(t *testing.T) *Credential {
				t.Helper()
				c := activeCredential(t)
				_ = c.Suspend()
				return c
			},
			wantErr: ErrAlreadySuspended,
		},
		{
			name:    "from inactive",
			setup:   func(t *testing.T) *Credential { t.Helper(); return inactiveCredential(t) },
			wantErr: ErrCannotSuspend,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup(t)
			err := c.Suspend()

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Suspend() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Suspend() unexpected error: %v", err)
			}

			if !c.IsSuspended() {
				t.Error("IsSuspended() should be true")
			}
		})
	}
}

func TestCredential_UpdateEmail(t *testing.T) {
	c, _ := NewCredential(NewCredentialParams{ID: "c-1", Email: validEmail(), Password: validPassword()})
	newEmail, _ := NewEmail("new@example.com")

	if err := c.UpdateEmail(newEmail); err != nil {
		t.Fatalf("UpdateEmail() unexpected error: %v", err)
	}

	if !c.Email().Equals(newEmail) {
		t.Errorf("Email() = %v, want %v", c.Email(), newEmail)
	}

	if err := c.UpdateEmail(Email{}); err != ErrEmailRequired {
		t.Errorf("UpdateEmail(zero) error = %v, want %v", err, ErrEmailRequired)
	}
}

func TestCredential_ChangePassword(t *testing.T) {
	c, _ := NewCredential(NewCredentialParams{ID: "c-1", Email: validEmail(), Password: validPassword()})
	newPwd, _ := NewHashedPassword("newpass123")

	if err := c.ChangePassword(newPwd); err != nil {
		t.Fatalf("ChangePassword() unexpected error: %v", err)
	}

	if !c.Password().Verify("newpass123") {
		t.Error("Password should verify with new plaintext")
	}

	if err := c.ChangePassword(HashedPassword{}); err != ErrPasswordRequired {
		t.Errorf("ChangePassword(zero) error = %v, want %v", err, ErrPasswordRequired)
	}
}
