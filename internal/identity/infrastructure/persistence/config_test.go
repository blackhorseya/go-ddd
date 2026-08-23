package persistence

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_DSN(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "no password omits the credential separator",
			cfg: Config{
				Host: "localhost", Port: 5432, User: "postgres",
				Name: "identity", SSLMode: "disable",
			},
			want: "postgres://postgres@localhost:5432/identity?sslmode=disable",
		},
		{
			name: "password is included",
			cfg: Config{
				Host: "db.example.com", Port: 5432, User: "app",
				Password: "secret", Name: "identity", SSLMode: "require",
			},
			want: "postgres://app:secret@db.example.com:5432/identity?sslmode=require",
		},
		{
			name: "reserved characters in the password are escaped",
			cfg: Config{
				Host: "localhost", Port: 5432, User: "app",
				Password: "p@ss:w/rd?", Name: "identity", SSLMode: "disable",
			},
			want: "postgres://app:p%40ss%3Aw%2Frd%3F@localhost:5432/identity?sslmode=disable",
		},
		{
			name: "empty ssl mode is omitted rather than emitted blank",
			cfg: Config{
				Host: "localhost", Port: 5432, User: "postgres",
				Name: "identity",
			},
			want: "postgres://postgres@localhost:5432/identity",
		},
		{
			name: "ipv6 host is bracketed",
			cfg: Config{
				Host: "::1", Port: 5432, User: "postgres",
				Name: "identity", SSLMode: "disable",
			},
			want: "postgres://postgres@[::1]:5432/identity?sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange / Act
			got := tt.cfg.DSN()

			// Assert
			assert.Equal(t, tt.want, got)
		})
	}
}
