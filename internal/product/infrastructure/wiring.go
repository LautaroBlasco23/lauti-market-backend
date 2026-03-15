package infrastructure

import (
	"database/sql"
	"net/http"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	imageDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/image/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/product/application"
	productController "github.com/LautaroBlasco23/lauti-market-backend/internal/product/infrastructure/controller"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/product/infrastructure/repository"
	productRoutes "github.com/LautaroBlasco23/lauti-market-backend/internal/product/infrastructure/routes"
	storeDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

type ProductModule struct {
	Repository *repository.ProductPostgresRepository
	Service    *application.ProductService
	Controller *productController.ProductController
}

func Wire(mux *http.ServeMux, db *sql.DB, idGen apiDomain.IDGenerator, storeRepo storeDomain.Repository, imageClient imageDomain.ImageClient) *ProductModule {
	repo := repository.NewProductPostgresRepository(db)
	service := application.NewService(repo, storeRepo, idGen, imageClient)
	controller := productController.NewProductController(service)

	productRoutes.RegisterRoutes(mux, controller)

	return &ProductModule{
		Repository: repo,
		Service:    service,
		Controller: controller,
	}
}
