package infrastructure

import (
	"context"
	"net/http"
	"time"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/api"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
	userdom "github.com/LautaroBlasco23/lauti-market-backend/internal/user/domain"
	userinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure"
)

type Module struct {
	Repository *InMemoryRepository
	Service    *application.Service
	Handler    *Handler
}

type Config struct {
	JWTSecret     string
	JWTExpiration time.Duration
}

type idGenAdapter struct {
	gen *api.UUIDGenerator
}

func (a *idGenAdapter) GenerateAuthID() domain.ID {
	return domain.ID(a.gen.Generate())
}

func (a *idGenAdapter) GenerateUserID() domain.UserID {
	return domain.UserID(a.gen.Generate())
}

type userServiceAdapter struct {
	repo *userinfra.InMemoryRepository
}

func (a *userServiceAdapter) Create(ctx context.Context, firstName, lastName string, id domain.UserID) error {
	u, err := userdom.NewUser(userdom.ID(id), firstName, lastName)
	if err != nil {
		return err
	}
	return a.repo.Save(ctx, u)
}

func Wire(mux *http.ServeMux, uuidGen *api.UUIDGenerator, userModule *userinfra.Module, cfg Config) *Module {
	repo := NewInMemoryRepository()
	hasher := NewBcryptHasher()
	jwtGen := NewJWTGenerator(cfg.JWTSecret, cfg.JWTExpiration)

	service := application.NewService(
		repo,
		&idGenAdapter{gen: uuidGen},
		hasher,
		jwtGen,
		&userServiceAdapter{repo: userModule.Repository},
	)
	handler := NewHandler(service)

	RegisterRoutes(mux, handler)

	return &Module{
		Repository: repo,
		Service:    service,
		Handler:    handler,
	}
}
