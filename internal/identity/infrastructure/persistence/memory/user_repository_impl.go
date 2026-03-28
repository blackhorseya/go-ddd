package memory

import (
	"context"
	"sort"
	"sync"

	domain "github.com/blackhorseya/go-ddd/internal/shared/domain"

	"github.com/blackhorseya/go-ddd/internal/identity/domain/user"
)

var _ user.Repository = (*userRepoImpl)(nil)

type userRepoImpl struct {
	mu    sync.RWMutex
	users map[string]*user.User
}

// NewUserRepository creates an in-memory user repository for testing and development.
func NewUserRepository() user.Repository {
	return &userRepoImpl{
		users: make(map[string]*user.User),
	}
}

func (i *userRepoImpl) Save(c context.Context, u *user.User) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Check email uniqueness (skip if updating the same user).
	for _, existing := range i.users {
		if existing.Email().Equals(u.Email()) && existing.ID() != u.ID() {
			return user.ErrEmailDuplicated
		}
	}

	i.users[u.ID()] = u

	return nil
}

func (i *userRepoImpl) FindByID(c context.Context, id string) (*user.User, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	u, ok := i.users[id]
	if !ok {
		return nil, user.ErrNotFound
	}

	return u, nil
}

func (i *userRepoImpl) FindByEmail(c context.Context, email user.Email) (*user.User, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	for _, u := range i.users {
		if u.Email().Equals(email) {
			return u, nil
		}
	}

	return nil, user.ErrNotFound
}

func (i *userRepoImpl) Delete(c context.Context, id string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if _, ok := i.users[id]; !ok {
		return user.ErrNotFound
	}

	delete(i.users, id)

	return nil
}

func (i *userRepoImpl) List(c context.Context, req domain.PageRequest) (domain.PageResult[*user.User], error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	all := make([]*user.User, 0, len(i.users))
	for _, u := range i.users {
		all = append(all, u)
	}

	// Stable sort by ID for deterministic output.
	sort.Slice(all, func(a, b int) bool {
		return all[a].ID() < all[b].ID()
	})

	total := int64(len(all))
	start := req.Offset()
	if start >= len(all) {
		return domain.NewPageResult([]*user.User{}, req.Page(), req.PageSize(), total), nil
	}

	end := min(start+req.Limit(), len(all))

	return domain.NewPageResult(all[start:end], req.Page(), req.PageSize(), total), nil
}
