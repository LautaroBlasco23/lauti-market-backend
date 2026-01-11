package infrastructure

import (
	"context"
	"database/sql"
	"net/http"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/controller"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/repository"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/routes"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/utils"
	storedom "github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
	storeinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure"
	userdom "github.com/LautaroBlasco23/lauti-market-backend/internal/user/domain"
	userinfra "github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure"
)

type Module struct {
	Repository *repository.AuthPostgresRepository
	Service    *application.Service
	Controller *controller.Controller
}

type userServiceAdapter struct {
	repo *userinfra.PostgresRepository
}

func (a *userServiceAdapter) Create(ctx context.Context, firstName, lastName string, id string) error {
	u, err := userdom.NewUser(id, firstName, lastName)
	if err != nil {
		return err
	}
	return a.repo.Save(ctx, u)
}

type storeServiceAdapter struct {
	repo *storeinfra.PostgresRepository
}

func (a *storeServiceAdapter) Create(ctx context.Context, name, description, address, phoneNumber string, id string) error {
	s, err := storedom.NewStore(id, name, description, address, phoneNumber)
	if err != nil {
		return err
	}
	return a.repo.Save(ctx, s)
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
		&userServiceAdapter{repo: userModule.Repository},
		&storeServiceAdapter{repo: storeModule.Repository},
	)

	authController := controller.NewController(service)
	routes.RegisterRoutes(mux, authController)

	return &Module{
		Repository: repo,
		Service:    service,
		Controller: authController,
	}
}
