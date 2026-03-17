package routes

import (
	"net/http"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure/controller"
)

func RegisterUserRoutes(mux *http.ServeMux, h *controller.UserController, validate func(string) (string, error)) {
	auth := infrastructure.RequireAuth(validate)
	mux.Handle("GET /users/{id}", auth(http.HandlerFunc(h.GetByID)))
	mux.Handle("PUT /users/{id}", auth(http.HandlerFunc(h.Update)))
	mux.Handle("DELETE /users/{id}", auth(http.HandlerFunc(h.Delete)))
}
