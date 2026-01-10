package infrastructure

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	mux.HandleFunc("POST /stores", handler.Create)
	mux.HandleFunc("GET /stores", handler.GetAll)
	mux.HandleFunc("GET /stores/{id}", handler.GetByID)
	mux.HandleFunc("PUT /stores/{id}", handler.Update)
	mux.HandleFunc("DELETE /stores/{id}", handler.Delete)
}
