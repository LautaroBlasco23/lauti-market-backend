package infrastructure

import (
	"context"
	"sync"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
)

type InMemoryRepository struct {
	mu      sync.RWMutex
	auths   map[domain.ID]*domain.Auth
	byEmail map[string]domain.ID
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		auths:   make(map[domain.ID]*domain.Auth),
		byEmail: make(map[string]domain.ID),
	}
}

func (r *InMemoryRepository) Save(_ context.Context, a *domain.Auth) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.auths[a.ID()] = a
	r.byEmail[a.Email()] = a.ID()
	return nil
}

func (r *InMemoryRepository) FindByID(_ context.Context, id domain.ID) (*domain.Auth, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.auths[id]
	if !ok {
		return nil, domain.ErrAuthNotFound
	}
	return a, nil
}

func (r *InMemoryRepository) FindByEmail(_ context.Context, email string) (*domain.Auth, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.byEmail[email]
	if !ok {
		return nil, domain.ErrAuthNotFound
	}
	return r.auths[id], nil
}

func (r *InMemoryRepository) Delete(_ context.Context, id domain.ID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if a, ok := r.auths[id]; ok {
		delete(r.byEmail, a.Email())
	}
	delete(r.auths, id)
	return nil
}
