package persistence

import "time"

// Config holds database configuration for the Identity BC's persistence layer.
// This struct is independent of shared/infrastructure/config — Wire maps the values.
type Config struct {
	Driver          string
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}
