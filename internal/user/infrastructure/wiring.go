package infrastructure

import (
	"database/sql"
	"net/http"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure/controller"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure/repository"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure/routes"
)

type UserModule struct {
	Repository *repository.UserPostgresRepository
	Service    *application.UserService
	Controller *controller.UserController
}

func Wire(mux *http.ServeMux, db *sql.DB, idGen apiDomain.IDGenerator, validate func(string) (string, error)) *UserModule {
	repo := repository.NewUserPostgresRepository(db)
	service := application.NewService(repo, idGen)
	userController := controller.NewUserController(service)

	routes.RegisterUserRoutes(mux, userController, validate)

	return &UserModule{
		Repository: repo,
		Service:    service,
		Controller: userController,
	}
}
