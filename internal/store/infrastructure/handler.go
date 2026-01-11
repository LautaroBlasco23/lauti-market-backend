package infrastructure

import (
	"encoding/json"
	"net/http"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

type createStoreRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
}

type updateStoreRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
}

type storeResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
}

func toStoreResponse(s *domain.Store) storeResponse {
	return storeResponse{
		ID:          string(s.ID()),
		Name:        s.Name(),
		Description: s.Description(),
		Address:     s.Address(),
		PhoneNumber: s.PhoneNumber(),
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	store, err := h.service.Create(r.Context(), application.CreateStoreInput{
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		PhoneNumber: req.PhoneNumber,
	})
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toStoreResponse(store))
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing store id", http.StatusBadRequest)
		return
	}

	store, err := h.service.GetByID(r.Context(), string(id))
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toStoreResponse(store))
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	stores, err := h.service.GetAll(r.Context(), 100, 0)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response := make([]storeResponse, len(stores))
	for i, s := range stores {
		response[i] = toStoreResponse(s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing store id", http.StatusBadRequest)
		return
	}

	var req updateStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	store, err := h.service.Update(r.Context(), application.UpdateStoreInput{
		ID:          string(id),
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		PhoneNumber: req.PhoneNumber,
	})
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toStoreResponse(store))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing store id", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), string(id)); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrStoreNotFound:
		http.Error(w, err.Error(), http.StatusNotFound)
	case domain.ErrInvalidName, domain.ErrInvalidDescription,
		domain.ErrInvalidAddress, domain.ErrInvalidPhoneNumber:
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
