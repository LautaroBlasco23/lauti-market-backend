package routes

import (
	"net/http"

	apiInfra "github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure/controller"
)

func RegisterUserRoutes(mux *http.ServeMux, h *controller.UserController, authMw *apiInfra.AuthMiddleware) {
	mux.HandleFunc("GET /users/{id}", h.GetByID)
	mux.Handle("PUT /users/{id}", authMw.Wrap(http.HandlerFunc(h.Update)))
	mux.Handle("DELETE /users/{id}", authMw.Wrap(http.HandlerFunc(h.Delete)))
}
