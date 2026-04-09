package controller

import (
	"encoding/json"
	"errors"
	"log/slog"
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
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("AuthController.RegisterUser started",
		slog.String("request_id", requestID),
	)

	var req dto.RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("AuthController.RegisterUser invalid request body",
			slog.String("request_id", requestID),
			slog.Any("error", err),
		)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := infrastructure.Validate(req); err != nil {
		slog.Warn("AuthController.RegisterUser validation failed",
			slog.String("request_id", requestID),
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

	slog.Debug("AuthController.RegisterUser calling service",
		slog.String("request_id", requestID),
		slog.String("email", req.Email),
	)

	output, err := h.service.RegisterUser(r.Context(), application.RegisterUserInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		slog.Error("AuthController.RegisterUser failed",
			slog.String("request_id", requestID),
			slog.String("email", req.Email),
			slog.Any("error", err),
		)
		h.handleError(w, err)
		return
	}

	slog.Info("AuthController.RegisterUser completed",
		slog.String("request_id", requestID),
		slog.String("auth_id", string(output.AuthID)),
		slog.String("account_id", string(output.AccountID)),
	)

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
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("AuthController.RegisterStore started",
		slog.String("request_id", requestID),
	)

	var req dto.RegisterStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("AuthController.RegisterStore invalid request body",
			slog.String("request_id", requestID),
			slog.Any("error", err),
		)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := infrastructure.Validate(req); err != nil {
		slog.Warn("AuthController.RegisterStore validation failed",
			slog.String("request_id", requestID),
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

	slog.Debug("AuthController.RegisterStore calling service",
		slog.String("request_id", requestID),
		slog.String("email", req.Email),
		slog.String("store_name", req.Name),
	)

	output, err := h.service.RegisterStore(r.Context(), &application.RegisterStoreInput{
		Email:       req.Email,
		Password:    req.Password,
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		PhoneNumber: req.PhoneNumber,
	})
	if err != nil {
		slog.Error("AuthController.RegisterStore failed",
			slog.String("request_id", requestID),
			slog.String("email", req.Email),
			slog.Any("error", err),
		)
		h.handleError(w, err)
		return
	}

	slog.Info("AuthController.RegisterStore completed",
		slog.String("request_id", requestID),
		slog.String("auth_id", string(output.AuthID)),
		slog.String("account_id", string(output.AccountID)),
	)

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
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("AuthController.Login started",
		slog.String("request_id", requestID),
	)

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("AuthController.Login invalid request body",
			slog.String("request_id", requestID),
			slog.Any("error", err),
		)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := infrastructure.Validate(req); err != nil {
		slog.Warn("AuthController.Login validation failed",
			slog.String("request_id", requestID),
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

	slog.Debug("AuthController.Login calling service",
		slog.String("request_id", requestID),
		slog.String("email", req.Email),
	)

	output, err := h.service.Login(r.Context(), application.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		slog.Error("AuthController.Login failed",
			slog.String("request_id", requestID),
			slog.String("email", req.Email),
			slog.Any("error", err),
		)
		h.handleError(w, err)
		return
	}

	slog.Info("AuthController.Login completed",
		slog.String("request_id", requestID),
		slog.String("account_id", string(output.AccountID)),
		slog.String("account_type", string(output.AccountType)),
	)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.LoginResponse{ //nolint:errcheck
		Token:       output.Token,
		AccountID:   string(output.AccountID),
		AccountType: string(output.AccountType),
	})
}

func (h *Controller) GetMe(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("AuthController.GetMe started",
		slog.String("request_id", requestID),
	)

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		slog.Warn("AuthController.GetMe unauthorized",
			slog.String("request_id", requestID),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	slog.Info("AuthController.GetMe completed",
		slog.String("request_id", requestID),
		slog.String("account_id", claims.AccountID),
		slog.String("account_type", string(claims.AccountType)),
	)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.MeResponse{ //nolint:errcheck
		AuthID:      claims.AuthID,
		AccountID:   claims.AccountID,
		AccountType: string(claims.AccountType),
	})
}

func (h *Controller) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apiDomain.ErrEmailExists):
		slog.Warn("AuthController handling error",
			slog.String("error_type", "email_exists"),
			slog.Any("error", err),
		)
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, apiDomain.ErrInvalidCredentials):
		slog.Warn("AuthController handling error",
			slog.String("error_type", "invalid_credentials"),
			slog.Any("error", err),
		)
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case errors.Is(err, apiDomain.ErrInvalidEmail), errors.Is(err, apiDomain.ErrInvalidPassword), errors.Is(err, apiDomain.ErrInvalidAccountID), errors.Is(err, apiDomain.ErrInvalidAccountType):
		slog.Warn("AuthController handling validation error",
			slog.String("error_type", "validation"),
			slog.Any("error", err),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		slog.Error("AuthController handling unexpected error",
			slog.String("error_type", "internal"),
			slog.Any("error", err),
		)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
