package credential

import (
	"testing"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "valid email", input: "user@example.com", want: "user@example.com"},
		{name: "uppercase normalized", input: "User@Example.COM", want: "user@example.com"},
		{name: "with plus tag", input: "user+tag@example.com", want: "user+tag@example.com"},
		{name: "trimmed spaces", input: "  user@example.com  ", want: "user@example.com"},
		{name: "empty string", input: "", wantErr: true},
		{name: "missing @", input: "userexample.com", wantErr: true},
		{name: "missing domain", input: "user@", wantErr: true},
		{name: "missing local", input: "@example.com", wantErr: true},
		{name: "no TLD", input: "user@example", wantErr: true},
		{name: "spaces only", input: "   ", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEmail(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewEmail(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("NewEmail(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got.Address() != tt.want {
				t.Errorf("NewEmail(%q).Address() = %q, want %q", tt.input, got.Address(), tt.want)
			}
		})
	}
}

func TestEmail_Equals(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want bool
	}{
		{name: "same address", a: "user@example.com", b: "user@example.com", want: true},
		{name: "case insensitive", a: "User@Example.com", b: "user@example.com", want: true},
		{name: "different address", a: "a@example.com", b: "b@example.com", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emailA, _ := NewEmail(tt.a)
			emailB, _ := NewEmail(tt.b)
			if got := emailA.Equals(emailB); got != tt.want {
				t.Errorf("Email(%q).Equals(Email(%q)) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestEmail_IsZero(t *testing.T) {
	var zero Email
	if !zero.IsZero() {
		t.Error("zero value Email.IsZero() should be true")
	}

	valid, _ := NewEmail("user@example.com")
	if valid.IsZero() {
		t.Error("constructed Email.IsZero() should be false")
	}
}

func TestEmail_String(t *testing.T) {
	email, _ := NewEmail("user@example.com")
	if email.String() != "user@example.com" {
		t.Errorf("Email.String() = %q, want %q", email.String(), "user@example.com")
	}
}
