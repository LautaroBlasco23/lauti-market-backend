package routes

import (
	"net/http"

	productController "github.com/LautaroBlasco23/lauti-market-backend/internal/product/infrastructure/controller"
)

func RegisterRoutes(mux *http.ServeMux, controller *productController.ProductController) {
	mux.HandleFunc("GET /products", controller.GetAll)
	mux.HandleFunc("POST /stores/{store_id}/products", controller.Create)
	mux.HandleFunc("GET /stores/{store_id}/products", controller.GetByStoreID)
	mux.HandleFunc("GET /stores/{store_id}/products/{id}", controller.GetByID)
	mux.HandleFunc("PUT /stores/{store_id}/products/{id}", controller.Update)
	mux.HandleFunc("DELETE /stores/{store_id}/products/{id}", controller.Delete)
	mux.HandleFunc("POST /stores/{store_id}/products/{id}/image", controller.UploadImage)
}
