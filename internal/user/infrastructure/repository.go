package infrastructure

import (
	"context"
	"sync"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/domain"
)

type InMemoryRepository struct {
	mu    sync.RWMutex
	users map[domain.ID]*domain.User
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		users: make(map[domain.ID]*domain.User),
	}
}

func (r *InMemoryRepository) Save(_ context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.ID()] = u
	return nil
}

func (r *InMemoryRepository) FindByID(_ context.Context, id domain.ID) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	u, ok := r.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (r *InMemoryRepository) Delete(_ context.Context, id domain.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, id)
	return nil
}
