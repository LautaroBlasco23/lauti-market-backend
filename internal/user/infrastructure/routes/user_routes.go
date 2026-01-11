package routes

import (
	"net/http"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure/controller"
)

func RegisterUserRoutes(mux *http.ServeMux, h *controller.UserController) {
	mux.HandleFunc("GET /users/{id}", h.GetByID)
	mux.HandleFunc("PUT /users/{id}", h.Update)
	mux.HandleFunc("DELETE /users/{id}", h.Delete)
}
