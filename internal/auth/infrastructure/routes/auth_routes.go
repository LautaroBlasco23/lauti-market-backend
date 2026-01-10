package routes

import (
	"net/http"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/controller"
)

func RegisterRoutes(mux *http.ServeMux, controller *controller.Controller) {
	mux.HandleFunc("POST /auth/register/user", controller.RegisterUser)
	mux.HandleFunc("POST /auth/register/store", controller.RegisterStore)
	mux.HandleFunc("POST /auth/login", controller.Login)
}
