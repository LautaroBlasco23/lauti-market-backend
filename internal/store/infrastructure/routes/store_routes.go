package routes

import (
	"net/http"

	apiInfra "github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	storeController "github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure/controller"
)

func RegisterRoutes(mux *http.ServeMux, controller *storeController.StoreController, authMw *apiInfra.AuthMiddleware) {
	mux.HandleFunc("GET /stores", controller.GetAll)
	mux.HandleFunc("GET /stores/{id}", controller.GetByID)
	mux.Handle("PUT /stores/{id}", authMw.Wrap(http.HandlerFunc(controller.Update)))
	mux.Handle("DELETE /stores/{id}", authMw.Wrap(http.HandlerFunc(controller.Delete)))
}
