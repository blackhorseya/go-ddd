package memory

import (
	"context"
	"testing"

	"github.com/blackhorseya/go-ddd/internal/identity/domain/credential"
	domain "github.com/blackhorseya/go-ddd/internal/shared/domain"
)

func seedCredential(t *testing.T, id, email string) *credential.Credential {
	t.Helper()

	e, err := credential.NewEmail(email)
	if err != nil {
		t.Fatalf("NewEmail(%q): %v", email, err)
	}

	p, err := credential.NewHashedPassword("secret123")
	if err != nil {
		t.Fatalf("NewHashedPassword: %v", err)
	}

	c, err := credential.NewCredential(credential.NewCredentialParams{
		ID:       id,
		Email:    e,
		Password: p,
	})
	if err != nil {
		t.Fatalf("NewCredential: %v", err)
	}

	return c
}

func TestCredentialRepoImpl_Save(t *testing.T) {
	repo := NewCredentialRepository()
	c := context.Background()

	t.Run("create new", func(t *testing.T) {
		cred := seedCredential(t, "c-1", "alice@example.com")

		if err := repo.Save(c, cred); err != nil {
			t.Fatalf("Save() unexpected error: %v", err)
		}

		got, err := repo.FindByID(c, "c-1")
		if err != nil {
			t.Fatalf("FindByID() unexpected error: %v", err)
		}

		if got.Email().Address() != "alice@example.com" {
			t.Errorf("Email() = %q, want %q", got.Email().Address(), "alice@example.com")
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		cred := seedCredential(t, "c-2", "alice@example.com")

		if err := repo.Save(c, cred); err != credential.ErrEmailDuplicated {
			t.Errorf("Save() error = %v, want %v", err, credential.ErrEmailDuplicated)
		}
	})
}

func TestCredentialRepoImpl_FindByID(t *testing.T) {
	repo := NewCredentialRepository()

	t.Run("not found", func(t *testing.T) {
		_, err := repo.FindByID(context.Background(), "non-existent")
		if err != credential.ErrNotFound {
			t.Errorf("FindByID() error = %v, want %v", err, credential.ErrNotFound)
		}
	})
}

func TestCredentialRepoImpl_FindByEmail(t *testing.T) {
	repo := NewCredentialRepository()
	c := context.Background()

	cred := seedCredential(t, "c-1", "alice@example.com")
	_ = repo.Save(c, cred)

	t.Run("found", func(t *testing.T) {
		email, _ := credential.NewEmail("alice@example.com")

		got, err := repo.FindByEmail(c, email)
		if err != nil {
			t.Fatalf("FindByEmail() unexpected error: %v", err)
		}

		if got.ID() != "c-1" {
			t.Errorf("ID() = %q, want %q", got.ID(), "c-1")
		}
	})

	t.Run("not found", func(t *testing.T) {
		email, _ := credential.NewEmail("nobody@example.com")

		_, err := repo.FindByEmail(c, email)
		if err != credential.ErrNotFound {
			t.Errorf("FindByEmail() error = %v, want %v", err, credential.ErrNotFound)
		}
	})
}

func TestCredentialRepoImpl_Delete(t *testing.T) {
	repo := NewCredentialRepository()
	c := context.Background()

	cred := seedCredential(t, "c-1", "alice@example.com")
	_ = repo.Save(c, cred)

	t.Run("delete existing", func(t *testing.T) {
		if err := repo.Delete(c, "c-1"); err != nil {
			t.Fatalf("Delete() unexpected error: %v", err)
		}

		_, err := repo.FindByID(c, "c-1")
		if err != credential.ErrNotFound {
			t.Errorf("FindByID() after delete error = %v, want %v", err, credential.ErrNotFound)
		}
	})

	t.Run("delete non-existent", func(t *testing.T) {
		if err := repo.Delete(c, "non-existent"); err != credential.ErrNotFound {
			t.Errorf("Delete() error = %v, want %v", err, credential.ErrNotFound)
		}
	})
}

func TestCredentialRepoImpl_List(t *testing.T) {
	repo := NewCredentialRepository()
	c := context.Background()

	for _, s := range []struct{ id, email string }{
		{"c-1", "a@example.com"},
		{"c-2", "b@example.com"},
		{"c-3", "c@example.com"},
	} {
		_ = repo.Save(c, seedCredential(t, s.id, s.email))
	}

	t.Run("first page", func(t *testing.T) {
		req, _ := domain.NewPageRequest(1, 2)

		result, err := repo.List(c, req)
		if err != nil {
			t.Fatalf("List() unexpected error: %v", err)
		}

		if len(result.Items()) != 2 {
			t.Errorf("Items() len = %d, want 2", len(result.Items()))
		}

		if result.TotalItems() != 3 {
			t.Errorf("TotalItems() = %d, want 3", result.TotalItems())
		}
	})

	t.Run("page beyond range", func(t *testing.T) {
		req, _ := domain.NewPageRequest(99, 2)

		result, err := repo.List(c, req)
		if err != nil {
			t.Fatalf("List() unexpected error: %v", err)
		}

		if len(result.Items()) != 0 {
			t.Errorf("Items() len = %d, want 0", len(result.Items()))
		}
	})
}
