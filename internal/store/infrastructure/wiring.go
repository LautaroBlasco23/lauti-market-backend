package infrastructure

import (
	"database/sql"
	"net/http"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/api"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

type Module struct {
	Repository *PostgresRepository
	Service    *application.Service
	Handler    *Handler
}

type idGenAdapter struct {
	gen *api.UUIDGenerator
}

func (a *idGenAdapter) GenerateStoreID() domain.ID {
	return domain.ID(a.gen.Generate())
}

func Wire(mux *http.ServeMux, db *sql.DB, uuidGen *api.UUIDGenerator) *Module {
	repo := NewPostgresRepository(db)
	service := application.NewService(repo, &idGenAdapter{gen: uuidGen})
	handler := NewHandler(service)

	RegisterRoutes(mux, handler)

	return &Module{
		Repository: repo,
		Service:    service,
		Handler:    handler,
	}
}
