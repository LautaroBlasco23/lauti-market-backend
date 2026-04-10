package routes

import (
	"net/http"

	apiInfra "github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/payment/infrastructure/controller"
)

func RegisterRoutes(mux *http.ServeMux, ctrl *controller.PaymentController, authMw *apiInfra.AuthMiddleware) {
	mux.Handle("POST /payments", authMw.Wrap(http.HandlerFunc(ctrl.Create)))
	mux.Handle("POST /payments/cart", authMw.Wrap(http.HandlerFunc(ctrl.CreateCartPreference)))
	mux.Handle("GET /payments/{id}", authMw.Wrap(http.HandlerFunc(ctrl.GetByID)))
	mux.Handle("GET /orders/{order_id}/payment", authMw.Wrap(http.HandlerFunc(ctrl.GetByOrderID)))
	mux.HandleFunc("POST /webhooks/mercadopago", ctrl.HandleWebhook)
}
