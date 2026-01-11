package infrastructure

import (
	"database/sql"
	"net/http"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/application"
)

type Module struct {
	Repository *PostgresRepository
	Service    *application.UserService
	Handler    *Handler
}

func Wire(mux *http.ServeMux, db *sql.DB, uuidGen apiDomain.IDGenerator) *Module {
	repo := NewPostgresRepository(db)
	service := application.NewService(repo, uuidGen)
	handler := NewHandler(service)

	RegisterRoutes(mux, handler)

	return &Module{
		Repository: repo,
		Service:    service,
		Handler:    handler,
	}
}
