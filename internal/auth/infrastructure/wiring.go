package infrastructure

import (
	"database/sql"
	"net/http"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/controller"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/repository"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/routes"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/utils"
	storeinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure"
	userinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure"
)

type Module struct {
	Repository *repository.AuthPostgresRepository
	Service    *application.AuthService
	Controller *controller.Controller
}

func Wire(
	mux *http.ServeMux,
	db *sql.DB,
	idGen apiDomain.IDGenerator,
	userModule *userinfra.Module,
	storeModule *storeinfra.Module,
	cfg utils.JwtConfig,
) *Module {
	repo := repository.NewPostgresRepository(db)
	hasher := utils.NewBcryptHasher()
	jwtGen := utils.NewJWTGenerator(cfg.JWTSecret, cfg.JWTExpiration)

	service := application.NewService(
		repo,
		idGen,
		hasher,
		jwtGen,
		userModule.Service,
		storeModule.Service,
	)

	authController := controller.NewController(service)
	routes.RegisterRoutes(mux, authController)

	return &Module{
		Repository: repo,
		Service:    service,
		Controller: authController,
	}
}
