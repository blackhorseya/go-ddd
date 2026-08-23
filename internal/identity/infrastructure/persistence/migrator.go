package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	pgxdriver "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib" // registers the "pgx" database/sql driver

	identitymigrations "github.com/blackhorseya/go-ddd/scripts/migrations/identity"
)

// ErrNoVersion is returned by Version when no migration has been applied yet.
var ErrNoVersion = migrate.ErrNilVersion

// migrationsRoot is the path inside the embedded FS holding the .sql files.
// //go:embed is rooted at the embedding package's own directory, so they sit at the top.
const migrationsRoot = "."

// Migrator applies the Identity BC's schema migrations to PostgreSQL.
// The SQL lives in scripts/migrations/identity and is embedded into the binary,
// so a deployment carries its own schema history.
type Migrator struct {
	runner *migrate.Migrate
}

// NewMigrator opens a connection using cfg and prepares the embedded migrations.
// The caller must call Close to release the connection.
func NewMigrator(cfg Config) (*Migrator, error) {
	source, err := iofs.New(identitymigrations.FS, migrationsRoot)
	if err != nil {
		return nil, fmt.Errorf("load embedded migrations: %w", err)
	}

	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open database %s: %w", cfg.Name, err)
	}

	// WithInstance pings and inspects the connection, so an unreachable database fails here.
	driver, err := pgxdriver.WithInstance(db, &pgxdriver.Config{})
	if err != nil {
		_ = db.Close()

		return nil, fmt.Errorf("connect database %s: %w", cfg.Name, err)
	}

	runner, err := migrate.NewWithInstance("iofs", source, cfg.Name, driver)
	if err != nil {
		_ = db.Close()

		return nil, fmt.Errorf("create migrator: %w", err)
	}

	return &Migrator{runner: runner}, nil
}

// Up applies every pending migration. An already up-to-date schema is not an error.
func (i *Migrator) Up() error {
	if err := i.runner.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

// Down rolls back every applied migration. An already empty schema is not an error.
func (i *Migrator) Down() error {
	if err := i.runner.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("roll back migrations: %w", err)
	}

	return nil
}

// Steps moves n migrations forward (n > 0) or backward (n < 0).
func (i *Migrator) Steps(n int) error {
	if err := i.runner.Steps(n); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate %d steps: %w", n, err)
	}

	return nil
}

// Version reports the applied version and whether the schema is dirty.
// It returns ErrNoVersion when no migration has been applied yet.
func (i *Migrator) Version() (version uint, dirty bool, err error) {
	return i.runner.Version()
}

// Force pins the schema version and clears the dirty flag, for recovering from a
// failed migration. It runs no SQL — the schema must already match the given version.
func (i *Migrator) Force(version int) error {
	if err := i.runner.Force(version); err != nil {
		return fmt.Errorf("force version %d: %w", version, err)
	}

	return nil
}

// Close releases the underlying database connection.
func (i *Migrator) Close() error {
	sourceErr, dbErr := i.runner.Close()

	return errors.Join(sourceErr, dbErr)
}
