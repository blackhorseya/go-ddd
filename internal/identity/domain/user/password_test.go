package user

import (
	"strings"
	"testing"
)

func TestNewHashedPassword(t *testing.T) {
	tests := []struct {
		name    string
		plain   string
		wantErr error
	}{
		{name: "valid password", plain: "secureP@ss1"},
		{name: "exactly 6 chars", plain: "123456"},
		{name: "empty", plain: "", wantErr: ErrPasswordRequired},
		{name: "too short", plain: "12345", wantErr: ErrPasswordTooShort},
		{name: "exceeds 72 bytes", plain: strings.Repeat("a", 73), wantErr: ErrPasswordTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hp, err := NewHashedPassword(tt.plain)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("NewHashedPassword() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewHashedPassword() unexpected error: %v", err)
			}

			if hp.IsZero() {
				t.Error("HashedPassword should not be zero after construction")
			}

			if hp.Hash() == tt.plain {
				t.Error("Hash() should not return plaintext")
			}
		})
	}
}

func TestHashedPassword_Verify(t *testing.T) {
	plain := "secureP@ss1"
	hp, err := NewHashedPassword(plain)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "correct password", input: plain, want: true},
		{name: "wrong password", input: "wrongpassword", want: false},
		{name: "empty string", input: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hp.Verify(tt.input); got != tt.want {
				t.Errorf("Verify(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewHashedPasswordFromHash(t *testing.T) {
	original, _ := NewHashedPassword("secureP@ss1")
	restored := NewHashedPasswordFromHash(original.Hash())

	if !restored.Verify("secureP@ss1") {
		t.Error("restored password should verify the original plaintext")
	}
}

func TestHashedPassword_IsZero(t *testing.T) {
	var zero HashedPassword
	if !zero.IsZero() {
		t.Error("zero value HashedPassword.IsZero() should be true")
	}
}
