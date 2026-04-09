package controller

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
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
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("StoreController.GetByID started",
		slog.String("request_id", requestID),
	)

	id := r.PathValue("id")
	if id == "" {
		slog.Warn("StoreController.GetByID missing store id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing store id", http.StatusBadRequest)
		return
	}

	slog.Debug("StoreController.GetByID calling service",
		slog.String("request_id", requestID),
		slog.String("store_id", id),
	)

	store, err := h.service.GetByID(r.Context(), string(id))
	if err != nil {
		slog.Error("StoreController.GetByID service error",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.Any("error", err),
		)
		h.handleError(w, err)
		return
	}

	slog.Info("StoreController.GetByID completed",
		slog.String("request_id", requestID),
		slog.String("store_id", id),
	)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toStoreResponse(store)) //nolint:errcheck
}

func (h *StoreController) GetAll(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("StoreController.GetAll started",
		slog.String("request_id", requestID),
	)

	stores, err := h.service.GetAll(r.Context(), 100, 0)
	if err != nil {
		slog.Error("StoreController.GetAll service error",
			slog.String("request_id", requestID),
			slog.Any("error", err),
		)
		h.handleError(w, err)
		return
	}

	slog.Info("StoreController.GetAll completed",
		slog.String("request_id", requestID),
		slog.Int("count", len(stores)),
	)

	response := make([]dto.StoreResponse, len(stores))
	for i, s := range stores {
		response[i] = toStoreResponse(s)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response) //nolint:errcheck
}

func (h *StoreController) Update(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("StoreController.Update started",
		slog.String("request_id", requestID),
	)

	id := r.PathValue("id")
	if id == "" {
		slog.Warn("StoreController.Update missing store id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing store id", http.StatusBadRequest)
		return
	}

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		slog.Warn("StoreController.Update unauthorized",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "store" {
		slog.Warn("StoreController.Update forbidden - wrong account type",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.String("account_type", string(claims.AccountType)),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}
	if claims.AccountID != id {
		slog.Warn("StoreController.Update forbidden - account mismatch",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.String("account_id", claims.AccountID),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	var req dto.UpdateStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("StoreController.Update invalid request body",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.Any("error", err),
		)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := infrastructure.Validate(req); err != nil {
		slog.Warn("StoreController.Update validation failed",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.Any("fields", infrastructure.FieldErrors(err)),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"error":  "invalid_payload",
			"fields": infrastructure.FieldErrors(err),
		})
		return
	}

	slog.Debug("StoreController.Update calling service",
		slog.String("request_id", requestID),
		slog.String("store_id", id),
	)

	store, err := h.service.Update(r.Context(), &application.UpdateStoreInput{
		ID:          string(id),
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		PhoneNumber: req.PhoneNumber,
	})
	if err != nil {
		slog.Error("StoreController.Update service error",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.Any("error", err),
		)
		h.handleError(w, err)
		return
	}

	slog.Info("StoreController.Update completed",
		slog.String("request_id", requestID),
		slog.String("store_id", id),
	)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toStoreResponse(store)) //nolint:errcheck
}

func (h *StoreController) Delete(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("StoreController.Delete started",
		slog.String("request_id", requestID),
	)

	id := r.PathValue("id")
	if id == "" {
		slog.Warn("StoreController.Delete missing store id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing store id", http.StatusBadRequest)
		return
	}

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		slog.Warn("StoreController.Delete unauthorized",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "store" {
		slog.Warn("StoreController.Delete forbidden - wrong account type",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.String("account_type", string(claims.AccountType)),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}
	if claims.AccountID != id {
		slog.Warn("StoreController.Delete forbidden - account mismatch",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.String("account_id", claims.AccountID),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	slog.Debug("StoreController.Delete calling service",
		slog.String("request_id", requestID),
		slog.String("store_id", id),
	)

	if err := h.service.Delete(r.Context(), string(id)); err != nil {
		slog.Error("StoreController.Delete service error",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.Any("error", err),
		)
		h.handleError(w, err)
		return
	}

	slog.Info("StoreController.Delete completed",
		slog.String("request_id", requestID),
		slog.String("store_id", id),
	)

	w.WriteHeader(http.StatusNoContent)
}

func (h *StoreController) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrStoreNotFound):
		slog.Warn("StoreController handling error",
			slog.String("error_type", "store_not_found"),
			slog.Any("error", err),
		)
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrInvalidName), errors.Is(err, domain.ErrInvalidDescription), errors.Is(err, domain.ErrInvalidAddress), errors.Is(err, domain.ErrInvalidPhoneNumber):
		slog.Warn("StoreController handling validation error",
			slog.String("error_type", "validation"),
			slog.Any("error", err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		slog.Error("StoreController handling unexpected error",
			slog.String("error_type", "internal"),
			slog.Any("error", err),
		)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
