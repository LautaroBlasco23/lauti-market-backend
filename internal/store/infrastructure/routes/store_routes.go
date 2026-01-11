package routes

import (
	"net/http"

	storeController "github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure/controller"
)

func RegisterRoutes(mux *http.ServeMux, controller *storeController.StoreController) {
	mux.HandleFunc("GET /stores", controller.GetAll)
	mux.HandleFunc("GET /stores/{id}", controller.GetByID)
	mux.HandleFunc("PUT /stores/{id}", controller.Update)
	mux.HandleFunc("DELETE /stores/{id}", controller.Delete)
}
