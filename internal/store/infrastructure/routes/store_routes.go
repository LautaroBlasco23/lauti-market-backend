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

	// MercadoPago OAuth endpoints
	mux.Handle("GET /stores/{id}/mercadopago/connect", authMw.Wrap(http.HandlerFunc(controller.GetOAuthConnectURL)))
	mux.Handle("POST /stores/{id}/mercadopago/callback", authMw.Wrap(http.HandlerFunc(controller.HandleOAuthCallback)))
	mux.Handle("GET /stores/{id}/mercadopago/status", authMw.Wrap(http.HandlerFunc(controller.GetMPConnectionStatus)))
	mux.Handle("DELETE /stores/{id}/mercadopago/disconnect", authMw.Wrap(http.HandlerFunc(controller.DisconnectMP)))

	// Public MercadoPago OAuth callback (MercadoPago redirects here directly)
	mux.HandleFunc("GET /mercadopago/oauth/callback", controller.HandlePublicOAuthCallback)
}
