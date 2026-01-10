package infrastructure

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/api"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
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

func (a *idGenAdapter) GenerateAccountID() domain.AccountID {
	return domain.AccountID(a.gen.Generate())
}

type userServiceAdapter struct {
	repo *userinfra.PostgresRepository
}

func (a *userServiceAdapter) Create(ctx context.Context, firstName, lastName string, id domain.AccountID) error {
	u, err := userdom.NewUser(userdom.ID(id), firstName, lastName)
	if err != nil {
		return err
	}
	return a.repo.Save(ctx, u)
}

type storeServiceAdapter struct {
	repo *storeinfra.PostgresRepository
}

func (a *storeServiceAdapter) Create(ctx context.Context, name, description, address, phoneNumber string, id domain.AccountID) error {
	s, err := storedom.NewStore(storedom.ID(id), name, description, address, phoneNumber)
	if err != nil {
		return err
	}
	return a.repo.Save(ctx, s)
}

func Wire(
	mux *http.ServeMux,
	db *sql.DB,
	uuidGen *api.UUIDGenerator,
	userModule *userinfra.Module,
	storeModule *storeinfra.Module,
	cfg Config,
) *Module {
	repo := repository.NewPostgresRepository(db)
	hasher := utils.NewBcryptHasher()
	jwtGen := utils.NewJWTGenerator(cfg.JWTSecret, cfg.JWTExpiration)

	service := application.NewService(
		repo,
		&idGenAdapter{gen: uuidGen},
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
