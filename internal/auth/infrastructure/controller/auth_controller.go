package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/application"
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
		_ = json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
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
	_ = json.NewEncoder(w).Encode(dto.RegisterResponse{ //nolint:errcheck
		AuthID:      string(output.AuthID),
		AccountID:   string(output.AccountID),
		AccountType: string(output.AccountType),
		Email:       output.Email,
		Token:       output.Token,
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
		_ = json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"error":  "invalid_payload",
			"fields": infrastructure.FieldErrors(err),
		})
		return
	}

	output, err := h.service.RegisterStore(r.Context(), &application.RegisterStoreInput{
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
	_ = json.NewEncoder(w).Encode(dto.RegisterResponse{ //nolint:errcheck
		AuthID:      string(output.AuthID),
		AccountID:   string(output.AccountID),
		AccountType: string(output.AccountType),
		Email:       output.Email,
		Token:       output.Token,
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
		_ = json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
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
	_ = json.NewEncoder(w).Encode(dto.LoginResponse{ //nolint:errcheck
		Token:       output.Token,
		AccountID:   string(output.AccountID),
		AccountType: string(output.AccountType),
	})
}

func (h *Controller) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apiDomain.ErrEmailExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, apiDomain.ErrInvalidCredentials):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case errors.Is(err, apiDomain.ErrInvalidEmail), errors.Is(err, apiDomain.ErrInvalidPassword), errors.Is(err, apiDomain.ErrInvalidAccountID), errors.Is(err, apiDomain.ErrInvalidAccountType):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
