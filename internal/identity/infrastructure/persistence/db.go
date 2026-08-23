package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver
)

// pingTimeout bounds the startup connectivity check.
const pingTimeout = 5 * time.Second

// NewDB opens the Identity BC's connection pool and verifies it is reachable.
// The returned cleanup closes the pool — Wire propagates it up to the caller.
func NewDB(cfg Config) (*sql.DB, func(), error) {
	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, nil, fmt.Errorf("open database %s: %w", cfg.Name, err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// sql.Open never dials, so ping here to fail at startup rather than on the
	// first request that happens to need the database.
	c, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	if err := db.PingContext(c); err != nil {
		_ = db.Close()

		return nil, nil, fmt.Errorf("connect database %s: %w", cfg.Name, err)
	}

	cleanup := func() {
		_ = db.Close()
	}

	return db, cleanup, nil
}
