package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure/dto"
)

type StoreController struct {
	service *application.StoreService
}

func NewStoreController(service *application.StoreService) *StoreController {
	return &StoreController{service: service}
}

func toStoreResponse(s *domain.Store) dto.StoreResponse {
	return dto.StoreResponse{
		ID:          string(s.ID()),
		Name:        s.Name(),
		Description: s.Description(),
		Address:     s.Address(),
		PhoneNumber: s.PhoneNumber(),
	}
}

func (h *StoreController) GetByID(w http.ResponseWriter, r *http.Request) {
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

func (h *StoreController) GetAll(w http.ResponseWriter, r *http.Request) {
	stores, err := h.service.GetAll(r.Context(), 100, 0)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response := make([]dto.StoreResponse, len(stores))
	for i, s := range stores {
		response[i] = toStoreResponse(s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *StoreController) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing store id", http.StatusBadRequest)
		return
	}

	var req dto.UpdateStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := infrastructure.Validate(req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error":  "invalid_payload",
			"fields": infrastructure.FieldErrors(err),
		})
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

func (h *StoreController) Delete(w http.ResponseWriter, r *http.Request) {
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

func (h *StoreController) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrStoreNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrInvalidName), errors.Is(err, domain.ErrInvalidDescription), errors.Is(err, domain.ErrInvalidAddress), errors.Is(err, domain.ErrInvalidPhoneNumber):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
