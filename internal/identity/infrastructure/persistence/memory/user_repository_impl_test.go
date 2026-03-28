package memory

import (
	"context"
	"testing"

	"github.com/blackhorseya/go-ddd/internal/identity/domain/user"
	domain "github.com/blackhorseya/go-ddd/internal/shared/domain"
)

func seedUser(t *testing.T, id, email, name string) *user.User {
	t.Helper()

	e, err := user.NewEmail(email)
	if err != nil {
		t.Fatalf("NewEmail(%q): %v", email, err)
	}

	p, err := user.NewHashedPassword("secret123")
	if err != nil {
		t.Fatalf("NewHashedPassword: %v", err)
	}

	u, err := user.NewUser(user.NewUserParams{
		ID:       id,
		Email:    e,
		Password: p,
		Name:     name,
	})
	if err != nil {
		t.Fatalf("NewUser: %v", err)
	}

	return u
}

func TestUserRepoImpl_Save(t *testing.T) {
	repo := NewUserRepository()
	c := context.Background()

	t.Run("create new user", func(t *testing.T) {
		u := seedUser(t, "u-1", "alice@example.com", "Alice")

		if err := repo.Save(c, u); err != nil {
			t.Fatalf("Save() unexpected error: %v", err)
		}

		got, err := repo.FindByID(c, "u-1")
		if err != nil {
			t.Fatalf("FindByID() unexpected error: %v", err)
		}

		if got.Name() != "Alice" {
			t.Errorf("Name() = %q, want %q", got.Name(), "Alice")
		}
	})

	t.Run("upsert existing user", func(t *testing.T) {
		u := seedUser(t, "u-1", "alice@example.com", "Alice Updated")

		if err := repo.Save(c, u); err != nil {
			t.Fatalf("Save() unexpected error: %v", err)
		}

		got, err := repo.FindByID(c, "u-1")
		if err != nil {
			t.Fatalf("FindByID() unexpected error: %v", err)
		}

		if got.Name() != "Alice Updated" {
			t.Errorf("Name() = %q, want %q", got.Name(), "Alice Updated")
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		u := seedUser(t, "u-2", "alice@example.com", "Bob")

		if err := repo.Save(c, u); err != user.ErrEmailDuplicated {
			t.Errorf("Save() error = %v, want %v", err, user.ErrEmailDuplicated)
		}
	})
}

func TestUserRepoImpl_FindByID(t *testing.T) {
	repo := NewUserRepository()
	c := context.Background()

	t.Run("not found", func(t *testing.T) {
		_, err := repo.FindByID(c, "non-existent")
		if err != user.ErrNotFound {
			t.Errorf("FindByID() error = %v, want %v", err, user.ErrNotFound)
		}
	})
}

func TestUserRepoImpl_FindByEmail(t *testing.T) {
	repo := NewUserRepository()
	c := context.Background()

	u := seedUser(t, "u-1", "alice@example.com", "Alice")
	_ = repo.Save(c, u)

	t.Run("found", func(t *testing.T) {
		email, _ := user.NewEmail("alice@example.com")

		got, err := repo.FindByEmail(c, email)
		if err != nil {
			t.Fatalf("FindByEmail() unexpected error: %v", err)
		}

		if got.ID() != "u-1" {
			t.Errorf("ID() = %q, want %q", got.ID(), "u-1")
		}
	})

	t.Run("not found", func(t *testing.T) {
		email, _ := user.NewEmail("nobody@example.com")

		_, err := repo.FindByEmail(c, email)
		if err != user.ErrNotFound {
			t.Errorf("FindByEmail() error = %v, want %v", err, user.ErrNotFound)
		}
	})
}

func TestUserRepoImpl_Delete(t *testing.T) {
	repo := NewUserRepository()
	c := context.Background()

	u := seedUser(t, "u-1", "alice@example.com", "Alice")
	_ = repo.Save(c, u)

	t.Run("delete existing", func(t *testing.T) {
		if err := repo.Delete(c, "u-1"); err != nil {
			t.Fatalf("Delete() unexpected error: %v", err)
		}

		_, err := repo.FindByID(c, "u-1")
		if err != user.ErrNotFound {
			t.Errorf("FindByID() after delete error = %v, want %v", err, user.ErrNotFound)
		}
	})

	t.Run("delete non-existent", func(t *testing.T) {
		if err := repo.Delete(c, "non-existent"); err != user.ErrNotFound {
			t.Errorf("Delete() error = %v, want %v", err, user.ErrNotFound)
		}
	})
}

func TestUserRepoImpl_List(t *testing.T) {
	repo := NewUserRepository()
	c := context.Background()

	// Seed 3 users.
	for _, s := range []struct{ id, email, name string }{
		{"u-1", "a@example.com", "Alice"},
		{"u-2", "b@example.com", "Bob"},
		{"u-3", "c@example.com", "Charlie"},
	} {
		_ = repo.Save(c, seedUser(t, s.id, s.email, s.name))
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

		if !result.HasNext() {
			t.Error("HasNext() should be true")
		}
	})

	t.Run("second page", func(t *testing.T) {
		req, _ := domain.NewPageRequest(2, 2)

		result, err := repo.List(c, req)
		if err != nil {
			t.Fatalf("List() unexpected error: %v", err)
		}

		if len(result.Items()) != 1 {
			t.Errorf("Items() len = %d, want 1", len(result.Items()))
		}

		if result.HasNext() {
			t.Error("HasNext() should be false")
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
