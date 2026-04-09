package routes

import (
	"net/http"

	apiInfra "github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/controller"
)

func RegisterRoutes(mux *http.ServeMux, controller *controller.Controller, authMw *apiInfra.AuthMiddleware) {
	mux.HandleFunc("POST /auth/register/user", controller.RegisterUser)
	mux.HandleFunc("POST /auth/register/store", controller.RegisterStore)
	mux.HandleFunc("POST /auth/login", controller.Login)
	mux.Handle("GET /auth/me", authMw.Wrap(http.HandlerFunc(controller.GetMe)))
}
