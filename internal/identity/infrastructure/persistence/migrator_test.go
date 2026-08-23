package persistence

import (
	"errors"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/stretchr/testify/require"

	identitymigrations "github.com/blackhorseya/go-ddd/scripts/migrations/identity"
)

// TestEmbeddedMigrations guards the migration files without needing a database:
// every version must be readable and ship both an up and a down script.
func TestEmbeddedMigrations(t *testing.T) {
	// Arrange
	driver, err := iofs.New(identitymigrations.FS, migrationsRoot)
	require.NoError(t, err, "embedded migrations must be parseable")

	t.Cleanup(func() {
		require.NoError(t, driver.Close())
	})

	// Act / Assert
	version, err := driver.First()
	require.NoError(t, err, "at least one migration must exist")

	for {
		up, _, err := driver.ReadUp(version)
		require.NoErrorf(t, err, "version %d is missing its up migration", version)
		require.NoError(t, up.Close())

		down, _, err := driver.ReadDown(version)
		require.NoErrorf(t, err, "version %d is missing its down migration", version)
		require.NoError(t, down.Close())

		next, err := driver.Next(version)
		if errors.Is(err, os.ErrNotExist) {
			break
		}

		require.NoErrorf(t, err, "failed to walk past version %d", version)
		version = next
	}
}
