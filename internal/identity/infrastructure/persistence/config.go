package persistence

import (
	"net"
	"net/url"
	"strconv"
	"time"
)

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

// DSN builds a PostgreSQL connection string from the config.
// net/url handles escaping, so passwords containing reserved characters stay intact.
func (c Config) DSN() string {
	u := &url.URL{
		Scheme: "postgres",
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:   "/" + c.Name,
	}

	if c.Password != "" {
		u.User = url.UserPassword(c.User, c.Password)
	} else {
		u.User = url.User(c.User)
	}

	// An empty sslmode= would make the driver's config parsing fail, so only set it
	// when a value is present and let the driver apply its own default otherwise.
	if c.SSLMode != "" {
		query := u.Query()
		query.Set("sslmode", c.SSLMode)
		u.RawQuery = query.Encode()
	}

	return u.String()
}
