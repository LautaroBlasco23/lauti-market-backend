package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/dto"
)

type Controller struct {
	service *application.AuthService
}

func NewController(service *application.AuthService) *Controller {
	return &Controller{service: service}
}

func (h *Controller) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterUserRequest
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

	output, err := h.service.RegisterUser(r.Context(), application.RegisterUserInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dto.RegisterResponse{
		AuthID:      string(output.AuthID),
		AccountID:   string(output.AccountID),
		AccountType: string(output.AccountType),
		Email:       output.Email,
	})
}

func (h *Controller) RegisterStore(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterStoreRequest
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

	output, err := h.service.RegisterStore(r.Context(), application.RegisterStoreInput{
		Email:       req.Email,
		Password:    req.Password,
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
	json.NewEncoder(w).Encode(dto.RegisterResponse{
		AuthID:      string(output.AuthID),
		AccountID:   string(output.AccountID),
		AccountType: string(output.AccountType),
		Email:       output.Email,
	})
}

func (h *Controller) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
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

	output, err := h.service.Login(r.Context(), application.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.LoginResponse{
		Token:       output.Token,
		AccountID:   string(output.AccountID),
		AccountType: string(output.AccountType),
	})
}

func (h *Controller) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrEmailExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, domain.ErrInvalidCredentials):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case errors.Is(err, domain.ErrInvalidEmail), errors.Is(err, domain.ErrInvalidPassword), errors.Is(err, domain.ErrInvalidAccountID), errors.Is(err, domain.ErrInvalidAccountType):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
