package routes

import (
	"net/http"

	apiInfra "github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	orderController "github.com/LautaroBlasco23/lauti-market-backend/internal/order/infrastructure/controller"
)

func RegisterRoutes(mux *http.ServeMux, ctrl *orderController.OrderController, authMw *apiInfra.AuthMiddleware) {
	mux.Handle("POST /orders", authMw.Wrap(http.HandlerFunc(ctrl.Create)))
	mux.Handle("GET /orders/{id}", authMw.Wrap(http.HandlerFunc(ctrl.GetByID)))
	mux.Handle("GET /users/{user_id}/orders", authMw.Wrap(http.HandlerFunc(ctrl.GetByUserID)))
	mux.Handle("GET /stores/{store_id}/orders", authMw.Wrap(http.HandlerFunc(ctrl.GetByStoreID)))
	mux.Handle("PUT /orders/{id}/status", authMw.Wrap(http.HandlerFunc(ctrl.UpdateStatus)))
}
