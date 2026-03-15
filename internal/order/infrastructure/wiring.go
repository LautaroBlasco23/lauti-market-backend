package infrastructure

import (
	"database/sql"
	"net/http"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	apiInfra "github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/order/application"
	orderController "github.com/LautaroBlasco23/lauti-market-backend/internal/order/infrastructure/controller"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/order/infrastructure/repository"
	orderRoutes "github.com/LautaroBlasco23/lauti-market-backend/internal/order/infrastructure/routes"
	productDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/product/domain"
)

type OrderModule struct {
	Repository *repository.OrderPostgresRepository
	Service    *application.OrderService
	Controller *orderController.OrderController
}

func Wire(mux *http.ServeMux, db *sql.DB, idGen apiDomain.IDGenerator, productRepo productDomain.Repository, authMw *apiInfra.AuthMiddleware) *OrderModule {
	repo := repository.NewOrderPostgresRepository(db)
	service := application.NewOrderService(repo, productRepo, idGen)
	ctrl := orderController.NewOrderController(service)

	orderRoutes.RegisterRoutes(mux, ctrl, authMw)

	return &OrderModule{
		Repository: repo,
		Service:    service,
		Controller: ctrl,
	}
}
