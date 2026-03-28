package memory

import (
	"context"
	"sort"
	"sync"

	"github.com/blackhorseya/go-ddd/internal/identity/domain/credential"
	domain "github.com/blackhorseya/go-ddd/internal/shared/domain"
)

var _ credential.Repository = (*credentialRepoImpl)(nil)

type credentialRepoImpl struct {
	mu    sync.RWMutex
	creds map[string]*credential.Credential
}

// NewCredentialRepository creates an in-memory credential repository.
func NewCredentialRepository() credential.Repository {
	return &credentialRepoImpl{
		creds: make(map[string]*credential.Credential),
	}
}

func (i *credentialRepoImpl) Save(c context.Context, cred *credential.Credential) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	for _, existing := range i.creds {
		if existing.Email().Equals(cred.Email()) && existing.ID() != cred.ID() {
			return credential.ErrEmailDuplicated
		}
	}

	i.creds[cred.ID()] = cred

	return nil
}

func (i *credentialRepoImpl) FindByID(c context.Context, id string) (*credential.Credential, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	cred, ok := i.creds[id]
	if !ok {
		return nil, credential.ErrNotFound
	}

	return cred, nil
}

func (i *credentialRepoImpl) FindByEmail(c context.Context, email credential.Email) (*credential.Credential, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	for _, cred := range i.creds {
		if cred.Email().Equals(email) {
			return cred, nil
		}
	}

	return nil, credential.ErrNotFound
}

func (i *credentialRepoImpl) Delete(c context.Context, id string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if _, ok := i.creds[id]; !ok {
		return credential.ErrNotFound
	}

	delete(i.creds, id)

	return nil
}

func (i *credentialRepoImpl) List(c context.Context, req domain.PageRequest) (domain.PageResult[*credential.Credential], error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	all := make([]*credential.Credential, 0, len(i.creds))
	for _, cred := range i.creds {
		all = append(all, cred)
	}

	sort.Slice(all, func(a, b int) bool {
		return all[a].ID() < all[b].ID()
	})

	total := int64(len(all))
	start := req.Offset()
	if start >= len(all) {
		return domain.NewPageResult([]*credential.Credential{}, req.Page(), req.PageSize(), total), nil
	}

	end := min(start+req.Limit(), len(all))

	return domain.NewPageResult(all[start:end], req.Page(), req.PageSize(), total), nil
}
