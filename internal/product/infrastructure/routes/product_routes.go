package routes

import (
	"net/http"

	apiInfra "github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	productController "github.com/LautaroBlasco23/lauti-market-backend/internal/product/infrastructure/controller"
)

func RegisterRoutes(mux *http.ServeMux, controller *productController.ProductController, authMw *apiInfra.AuthMiddleware) {
	mux.HandleFunc("GET /products", controller.GetAll)
	mux.HandleFunc("GET /stores/{store_id}/products", controller.GetByStoreID)
	mux.HandleFunc("GET /stores/{store_id}/products/{id}", controller.GetByID)
	mux.Handle("POST /stores/{store_id}/products", authMw.Wrap(http.HandlerFunc(controller.Create)))
	mux.Handle("PUT /stores/{store_id}/products/{id}", authMw.Wrap(http.HandlerFunc(controller.Update)))
	mux.Handle("DELETE /stores/{store_id}/products/{id}", authMw.Wrap(http.HandlerFunc(controller.Delete)))
	mux.Handle("POST /stores/{store_id}/products/{id}/image", authMw.Wrap(http.HandlerFunc(controller.UploadImage)))
}
