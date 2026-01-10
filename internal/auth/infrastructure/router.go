package infrastructure

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("POST /auth/register/user", handler.RegisterUser)
	mux.HandleFunc("POST /auth/register/store", handler.RegisterStore)
	mux.HandleFunc("POST /auth/login", handler.Login)
}
