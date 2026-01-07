package infrastructure

import (
	"net/http"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/api"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/domain"
)

type Module struct {
	Repository *InMemoryRepository
	Service    *application.Service
	Handler    *Handler
}

type idGenAdapter struct {
	gen *api.UUIDGenerator
}

func (a *idGenAdapter) Generate() domain.ID {
	return domain.ID(a.gen.Generate())
}

func Wire(mux *http.ServeMux, uuidGen *api.UUIDGenerator) *Module {
	repo := NewInMemoryRepository()
	service := application.NewService(repo, &idGenAdapter{gen: uuidGen})
	handler := NewHandler(service)

	RegisterRoutes(mux, handler)

	return &Module{
		Repository: repo,
		Service:    service,
		Handler:    handler,
	}
}
