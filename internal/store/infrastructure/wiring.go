package infrastructure

import (
	"database/sql"
	"net/http"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/application"
	storeController "github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure/controller"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure/repository"
	storeRoutes "github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure/routes"
)

type Module struct {
	Repository *repository.StorePostgresRepository
	Service    *application.Service
	Controller *storeController.StoreController
}

func Wire(mux *http.ServeMux, db *sql.DB, idGen apiDomain.IDGenerator) *Module {
	repo := repository.NewStorePostgresRepository(db)
	service := application.NewService(repo, idGen)
	storeController := storeController.NewStoreController(service)

	storeRoutes.RegisterRoutes(mux, storeController)

	return &Module{
		Repository: repo,
		Service:    service,
		Controller: storeController,
	}
}
