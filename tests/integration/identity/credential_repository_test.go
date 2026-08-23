// Package identity holds integration tests for the Identity bounded context.
//
// They run against a real PostgreSQL instance. When none is reachable the tests
// skip instead of failing, so `task test` stays green without local infra.
// Connection settings come from TEST_DATABASE_* environment variables; the
// defaults match deployments/docker/docker-compose.yaml.
package identity

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/blackhorseya/go-ddd/internal/identity/domain/credential"
	"github.com/blackhorseya/go-ddd/internal/identity/infrastructure/persistence"
	"github.com/blackhorseya/go-ddd/internal/identity/infrastructure/persistence/postgres"
	domain "github.com/blackhorseya/go-ddd/internal/shared/domain"
)

// maintenanceDatabase is the database used to create the test database itself.
const maintenanceDatabase = "postgres"

var (
	testDB *sql.DB

	// skipReason is non-empty when no database is available; every test skips on it.
	skipReason string
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

// run owns the whole database lifecycle so every exit path still tears down.
func run(m *testing.M) int {
	cfg := testConfig()

	if err := ensureDatabase(cfg); err != nil {
		skipReason = fmt.Sprintf("no postgres at %s:%d (%v)", cfg.Host, cfg.Port, err)

		return m.Run()
	}

	// The schema is built by the real migrations, so this exercises them too.
	migrator, err := persistence.NewMigrator(cfg)
	if err != nil {
		skipReason = fmt.Sprintf("cannot open migrator: %v", err)

		return m.Run()
	}

	if err := migrator.Up(); err != nil {
		_ = migrator.Close()
		skipReason = fmt.Sprintf("cannot apply migrations: %v", err)

		return m.Run()
	}

	defer func() {
		_ = migrator.Down()
		_ = migrator.Close()
	}()

	db, cleanup, err := persistence.NewDB(cfg)
	if err != nil {
		skipReason = fmt.Sprintf("cannot open database: %v", err)

		return m.Run()
	}
	defer cleanup()

	testDB = db

	return m.Run()
}

func TestCredentialRepository_SaveAndFind(t *testing.T) {
	repo := newRepo(t)
	c := t.Context()

	// Arrange
	cred := newCredential(t, "cred-001", "Alice@Example.com")

	// Act
	require.NoError(t, repo.Save(c, cred))

	// Assert — found by id
	found, err := repo.FindByID(c, "cred-001")
	require.NoError(t, err)
	assert.Equal(t, cred.ID(), found.ID())
	// The Email VO lowercases on construction, so that is what round-trips.
	assert.Equal(t, "alice@example.com", found.Email().Address())
	assert.Equal(t, credential.StatusActive, found.Status())
	assert.True(t, found.Password().Verify("secret123"))
	assert.WithinDuration(t, cred.CreatedAt(), found.CreatedAt(), time.Millisecond)

	// Assert — found by email
	found, err = repo.FindByEmail(c, cred.Email())
	require.NoError(t, err)
	assert.Equal(t, cred.ID(), found.ID())
}

func TestCredentialRepository_SaveUpdatesExisting(t *testing.T) {
	repo := newRepo(t)
	c := t.Context()

	// Arrange
	cred := newCredential(t, "cred-002", "bob@example.com")
	require.NoError(t, repo.Save(c, cred))

	// Act — Save upserts, so a second Save of the same id updates in place
	require.NoError(t, cred.Suspend())
	require.NoError(t, repo.Save(c, cred))

	// Assert
	found, err := repo.FindByID(c, "cred-002")
	require.NoError(t, err)
	assert.Equal(t, credential.StatusSuspended, found.Status())

	page, err := repo.List(c, domain.NewPageRequestWithDefaults())
	require.NoError(t, err)
	assert.Equal(t, int64(1), page.TotalItems(), "upsert must not insert a second row")
}

func TestCredentialRepository_SaveDuplicateEmail(t *testing.T) {
	repo := newRepo(t)
	c := t.Context()

	// Arrange
	require.NoError(t, repo.Save(c, newCredential(t, "cred-003", "carol@example.com")))

	// Act — a different id carrying an email already in use
	err := repo.Save(c, newCredential(t, "cred-004", "carol@example.com"))

	// Assert — the driver error is translated into the domain error
	assert.ErrorIs(t, err, credential.ErrEmailDuplicated)
}

func TestCredentialRepository_NotFound(t *testing.T) {
	repo := newRepo(t)
	c := t.Context()

	// Arrange
	email, err := credential.NewEmail("nobody@example.com")
	require.NoError(t, err)

	// Act / Assert
	_, err = repo.FindByID(c, "missing")
	assert.ErrorIs(t, err, credential.ErrNotFound)

	_, err = repo.FindByEmail(c, email)
	assert.ErrorIs(t, err, credential.ErrNotFound)

	assert.ErrorIs(t, repo.Delete(c, "missing"), credential.ErrNotFound)
}

func TestCredentialRepository_Delete(t *testing.T) {
	repo := newRepo(t)
	c := t.Context()

	// Arrange
	require.NoError(t, repo.Save(c, newCredential(t, "cred-005", "dave@example.com")))

	// Act
	require.NoError(t, repo.Delete(c, "cred-005"))

	// Assert
	_, err := repo.FindByID(c, "cred-005")
	assert.ErrorIs(t, err, credential.ErrNotFound)
}

func TestCredentialRepository_ListPaginates(t *testing.T) {
	repo := newRepo(t)
	c := t.Context()

	// Arrange — ids are zero-padded so the ORDER BY id result is predictable
	for n := 1; n <= 5; n++ {
		id := fmt.Sprintf("cred-%03d", n)
		require.NoError(t, repo.Save(c, newCredential(t, id, fmt.Sprintf("user%d@example.com", n))))
	}

	req, err := domain.NewPageRequest(2, 2)
	require.NoError(t, err)

	// Act
	page, err := repo.List(c, req)
	require.NoError(t, err)

	// Assert
	assert.Equal(t, int64(5), page.TotalItems())
	assert.Equal(t, 3, page.TotalPages())
	require.Len(t, page.Items(), 2)
	assert.Equal(t, "cred-003", page.Items()[0].ID())
	assert.Equal(t, "cred-004", page.Items()[1].ID())
	assert.True(t, page.HasNext())
	assert.True(t, page.HasPrev())
}

// newRepo returns a repository over an empty credentials table.
func newRepo(t *testing.T) credential.Repository {
	t.Helper()

	if skipReason != "" {
		t.Skip(skipReason)
	}

	_, err := testDB.ExecContext(t.Context(), `TRUNCATE TABLE credentials`)
	require.NoError(t, err)

	return postgres.NewCredentialRepository(testDB)
}

// newCredential builds a valid aggregate for the tests to persist.
func newCredential(t *testing.T, id, address string) *credential.Credential {
	t.Helper()

	email, err := credential.NewEmail(address)
	require.NoError(t, err)

	password, err := credential.NewHashedPassword("secret123")
	require.NoError(t, err)

	cred, err := credential.NewCredential(credential.NewCredentialParams{
		ID:       id,
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)

	return cred
}

// testConfig reads the connection settings, defaulting to the local compose stack.
// The database name defaults to identity_test so the tests never touch dev data.
func testConfig() persistence.Config {
	port, err := strconv.Atoi(envOrDefault("TEST_DATABASE_PORT", "5432"))
	if err != nil {
		port = 5432
	}

	return persistence.Config{
		Driver:          "postgres",
		Host:            envOrDefault("TEST_DATABASE_HOST", "localhost"),
		Port:            port,
		User:            envOrDefault("TEST_DATABASE_USER", "postgres"),
		Password:        os.Getenv("TEST_DATABASE_PASSWORD"),
		Name:            envOrDefault("TEST_DATABASE_NAME", "identity_test"),
		SSLMode:         envOrDefault("TEST_DATABASE_SSL_MODE", "disable"),
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
	}
}

// ensureDatabase creates the test database when it does not exist yet.
func ensureDatabase(cfg persistence.Config) error {
	admin := cfg
	admin.Name = maintenanceDatabase

	db, cleanup, err := persistence.NewDB(admin)
	if err != nil {
		return err
	}
	defer cleanup()

	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var exists bool
	if err := db.QueryRowContext(
		c, `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, cfg.Name,
	).Scan(&exists); err != nil {
		return fmt.Errorf("check database %s: %w", cfg.Name, err)
	}

	if exists {
		return nil
	}

	// CREATE DATABASE takes no parameters, so the name is quoted as an identifier.
	if _, err := db.ExecContext(c, `CREATE DATABASE `+pgx.Identifier{cfg.Name}.Sanitize()); err != nil {
		return fmt.Errorf("create database %s: %w", cfg.Name, err)
	}

	return nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
