package infrastructure

import (
	"database/sql"
	"net/http"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/application"
)

type Module struct {
	Repository *PostgresRepository
	Service    *application.Service
	Handler    *Handler
}

func Wire(mux *http.ServeMux, db *sql.DB, idGen apiDomain.IDGenerator) *Module {
	repo := NewPostgresRepository(db)
	service := application.NewService(repo, idGen)
	handler := NewHandler(service)

	RegisterRoutes(mux, handler)

	return &Module{
		Repository: repo,
		Service:    service,
		Handler:    handler,
	}
}
