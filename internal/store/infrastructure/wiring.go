package infrastructure

import (
	"database/sql"
	"net/http"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	apiInfra "github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/application"
	storeController "github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure/controller"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure/repository"
	storeRoutes "github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure/routes"
)

type StoreModule struct {
	Repository *repository.StorePostgresRepository
	Service    *application.StoreService
	Controller *storeController.StoreController
}

func Wire(mux *http.ServeMux, db *sql.DB, idGen apiDomain.IDGenerator, authMw *apiInfra.AuthMiddleware) *StoreModule {
	repo := repository.NewStorePostgresRepository(db)
	service := application.NewService(repo, idGen)
	storeController := storeController.NewStoreController(service)

	storeRoutes.RegisterRoutes(mux, storeController, authMw)

	return &StoreModule{
		Repository: repo,
		Service:    service,
		Controller: storeController,
	}
}
