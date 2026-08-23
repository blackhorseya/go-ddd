// Package postgres implements the Identity BC's repositories on PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/blackhorseya/go-ddd/internal/identity/domain/credential"
	domain "github.com/blackhorseya/go-ddd/internal/shared/domain"
)

// emailUniqueConstraint is the unique index guarding credentials.email, declared in
// scripts/migrations/identity/000001_create_credentials.up.sql.
const emailUniqueConstraint = "credentials_email_key"

// credentialColumns is the column list shared by every read query, kept in the
// order scanCredential expects.
const credentialColumns = `id, email, password_hash, status, created_at, updated_at`

const saveCredentialStmt = `
INSERT INTO credentials (id, email, password_hash, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    password_hash = EXCLUDED.password_hash,
    status = EXCLUDED.status,
    updated_at = EXCLUDED.updated_at`

// credentialPage shortens the paginated return type at every use site.
type credentialPage = domain.PageResult[*credential.Credential]

var _ credential.Repository = (*credentialRepoImpl)(nil)

type credentialRepoImpl struct {
	db *sql.DB
}

// NewCredentialRepository creates a PostgreSQL-backed credential repository.
func NewCredentialRepository(db *sql.DB) credential.Repository {
	return &credentialRepoImpl{db: db}
}

func (i *credentialRepoImpl) Save(c context.Context, cred *credential.Credential) error {
	_, err := i.db.ExecContext(c, saveCredentialStmt,
		cred.ID(),
		cred.Email().Address(),
		cred.Password().Hash(),
		string(cred.Status()),
		cred.CreatedAt(),
		cred.UpdatedAt(),
	)
	if err != nil {
		if isEmailConflict(err) {
			return credential.ErrEmailDuplicated
		}

		return fmt.Errorf("save credential %s: %w", cred.ID(), err)
	}

	return nil
}

func (i *credentialRepoImpl) FindByID(c context.Context, id string) (*credential.Credential, error) {
	row := i.db.QueryRowContext(c, `SELECT `+credentialColumns+` FROM credentials WHERE id = $1`, id)

	cred, err := scanCredential(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, credential.ErrNotFound
		}

		return nil, fmt.Errorf("find credential %s: %w", id, err)
	}

	return cred, nil
}

func (i *credentialRepoImpl) FindByEmail(
	c context.Context,
	email credential.Email,
) (*credential.Credential, error) {
	row := i.db.QueryRowContext(c, `SELECT `+credentialColumns+` FROM credentials WHERE email = $1`, email.Address())

	cred, err := scanCredential(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, credential.ErrNotFound
		}

		return nil, fmt.Errorf("find credential by email: %w", err)
	}

	return cred, nil
}

func (i *credentialRepoImpl) Delete(c context.Context, id string) error {
	result, err := i.db.ExecContext(c, `DELETE FROM credentials WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete credential %s: %w", id, err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete credential %s: %w", id, err)
	}

	if affected == 0 {
		return credential.ErrNotFound
	}

	return nil
}

func (i *credentialRepoImpl) List(c context.Context, req domain.PageRequest) (credentialPage, error) {
	var total int64
	if err := i.db.QueryRowContext(c, `SELECT COUNT(*) FROM credentials`).Scan(&total); err != nil {
		return credentialPage{}, fmt.Errorf("count credentials: %w", err)
	}

	// req.Sort() is ignored on purpose: ordering by id matches the in-memory
	// repository, and a caller-supplied column cannot be parameterized — it would
	// have to be interpolated, which is exactly the SQL injection path to avoid.
	rows, err := i.db.QueryContext(c,
		`SELECT `+credentialColumns+` FROM credentials ORDER BY id LIMIT $1 OFFSET $2`,
		req.Limit(), req.Offset(),
	)
	if err != nil {
		return credentialPage{}, fmt.Errorf("list credentials: %w", err)
	}

	defer func() {
		_ = rows.Close()
	}()

	creds := make([]*credential.Credential, 0, req.Limit())

	for rows.Next() {
		cred, err := scanCredential(rows)
		if err != nil {
			return credentialPage{}, fmt.Errorf("list credentials: %w", err)
		}

		creds = append(creds, cred)
	}

	if err := rows.Err(); err != nil {
		return credentialPage{}, fmt.Errorf("list credentials: %w", err)
	}

	return domain.NewPageResult(creds, req.Page(), req.PageSize(), total), nil
}

// rowScanner is satisfied by both *sql.Row and *sql.Rows, so single-row and
// multi-row reads share one rebuild path.
type rowScanner interface {
	Scan(dest ...any) error
}

// scanCredential rebuilds the aggregate from a row. Scan errors are returned
// unwrapped so callers can test them with errors.Is(err, sql.ErrNoRows).
func scanCredential(scanner rowScanner) (*credential.Credential, error) {
	var (
		id        string
		address   string
		hash      string
		status    string
		createdAt time.Time
		updatedAt time.Time
	)

	if err := scanner.Scan(&id, &address, &hash, &status, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	email, err := credential.NewEmail(address)
	if err != nil {
		return nil, fmt.Errorf("rebuild email of credential %s: %w", id, err)
	}

	return credential.ReconstituteCredential(credential.ReconstituteCredentialParams{
		ID:        id,
		Email:     email,
		Password:  credential.NewHashedPasswordFromHash(hash),
		Status:    credential.Status(status),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	})
}

// isEmailConflict reports whether err is the unique violation on credentials.email.
// Matching the constraint name rather than the SQLSTATE alone keeps the translation
// correct once further unique constraints exist on the table.
func isEmailConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == pgerrcode.UniqueViolation && pgErr.ConstraintName == emailUniqueConstraint
}
