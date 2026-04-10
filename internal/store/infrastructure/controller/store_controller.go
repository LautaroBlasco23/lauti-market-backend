package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

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

func (h *StoreController) GetOAuthConnectURL(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("StoreController.GetOAuthConnectURL started",
		slog.String("request_id", requestID),
	)

	id := r.PathValue("id")
	if id == "" {
		slog.Warn("StoreController.GetOAuthConnectURL missing store id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing store id", http.StatusBadRequest)
		return
	}

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		slog.Warn("StoreController.GetOAuthConnectURL unauthorized",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "store" {
		slog.Warn("StoreController.GetOAuthConnectURL forbidden - wrong account type",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.String("account_type", string(claims.AccountType)),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}
	if claims.AccountID != id {
		slog.Warn("StoreController.GetOAuthConnectURL forbidden - account mismatch",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.String("account_id", claims.AccountID),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	slog.Debug("StoreController.GetOAuthConnectURL calling service",
		slog.String("request_id", requestID),
		slog.String("store_id", id),
	)

	authURL, err := h.service.GetOAuthConnectURL(r.Context(), id)
	if err != nil {
		slog.Error("StoreController.GetOAuthConnectURL service error",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.Any("error", err),
		)
		h.handleError(w, err)
		return
	}

	slog.Info("StoreController.GetOAuthConnectURL completed",
		slog.String("request_id", requestID),
		slog.String("store_id", id),
	)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.OAuthConnectResponse{AuthURL: authURL}) //nolint:errcheck
}

func (h *StoreController) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("StoreController.HandleOAuthCallback started",
		slog.String("request_id", requestID),
	)

	id := r.PathValue("id")
	if id == "" {
		slog.Warn("StoreController.HandleOAuthCallback missing store id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing store id", http.StatusBadRequest)
		return
	}

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		slog.Warn("StoreController.HandleOAuthCallback unauthorized",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "store" {
		slog.Warn("StoreController.HandleOAuthCallback forbidden - wrong account type",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.String("account_type", string(claims.AccountType)),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}
	if claims.AccountID != id {
		slog.Warn("StoreController.HandleOAuthCallback forbidden - account mismatch",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.String("account_id", claims.AccountID),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	var req dto.OAuthCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("StoreController.HandleOAuthCallback invalid request body",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.Any("error", err),
		)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := infrastructure.Validate(req); err != nil {
		slog.Warn("StoreController.HandleOAuthCallback validation failed",
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

	slog.Debug("StoreController.HandleOAuthCallback calling service",
		slog.String("request_id", requestID),
		slog.String("store_id", id),
	)

	if err := h.service.HandleOAuthCallback(r.Context(), id, req.Code); err != nil {
		slog.Error("StoreController.HandleOAuthCallback service error",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.Any("error", err),
		)
		handleOAuthError(w, err)
		return
	}

	slog.Info("StoreController.HandleOAuthCallback completed",
		slog.String("request_id", requestID),
		slog.String("store_id", id),
	)

	w.WriteHeader(http.StatusOK)
}

func (h *StoreController) GetMPConnectionStatus(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("StoreController.GetMPConnectionStatus started",
		slog.String("request_id", requestID),
	)

	id := r.PathValue("id")
	if id == "" {
		slog.Warn("StoreController.GetMPConnectionStatus missing store id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing store id", http.StatusBadRequest)
		return
	}

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		slog.Warn("StoreController.GetMPConnectionStatus unauthorized",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "store" {
		slog.Warn("StoreController.GetMPConnectionStatus forbidden - wrong account type",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.String("account_type", string(claims.AccountType)),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}
	if claims.AccountID != id {
		slog.Warn("StoreController.GetMPConnectionStatus forbidden - account mismatch",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.String("account_id", claims.AccountID),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	slog.Debug("StoreController.GetMPConnectionStatus calling service",
		slog.String("request_id", requestID),
		slog.String("store_id", id),
	)

	status, err := h.service.GetMPConnectionStatus(r.Context(), id)
	if err != nil {
		slog.Error("StoreController.GetMPConnectionStatus service error",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.Any("error", err),
		)
		handleOAuthError(w, err)
		return
	}

	slog.Info("StoreController.GetMPConnectionStatus completed",
		slog.String("request_id", requestID),
		slog.String("store_id", id),
	)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status) //nolint:errcheck
}

func (h *StoreController) DisconnectMP(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("StoreController.DisconnectMP started",
		slog.String("request_id", requestID),
	)

	id := r.PathValue("id")
	if id == "" {
		slog.Warn("StoreController.DisconnectMP missing store id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing store id", http.StatusBadRequest)
		return
	}

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		slog.Warn("StoreController.DisconnectMP unauthorized",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "store" {
		slog.Warn("StoreController.DisconnectMP forbidden - wrong account type",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.String("account_type", string(claims.AccountType)),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}
	if claims.AccountID != id {
		slog.Warn("StoreController.DisconnectMP forbidden - account mismatch",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.String("account_id", claims.AccountID),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	slog.Debug("StoreController.DisconnectMP calling service",
		slog.String("request_id", requestID),
		slog.String("store_id", id),
	)

	if err := h.service.DisconnectMP(r.Context(), id); err != nil {
		slog.Error("StoreController.DisconnectMP service error",
			slog.String("request_id", requestID),
			slog.String("store_id", id),
			slog.Any("error", err),
		)
		handleOAuthError(w, err)
		return
	}

	slog.Info("StoreController.DisconnectMP completed",
		slog.String("request_id", requestID),
		slog.String("store_id", id),
	)

	w.WriteHeader(http.StatusNoContent)
}

func handleOAuthError(w http.ResponseWriter, err error) {
	slog.Warn("StoreController handling OAuth error",
		slog.String("error_type", "oauth_error"),
		slog.Any("error", err),
	)
	http.Error(w, err.Error(), http.StatusBadRequest)
}

// HandlePublicOAuthCallback handles the MercadoPago OAuth redirect callback
// This endpoint is PUBLIC (no auth required) - MercadoPago redirects here directly
func (h *StoreController) HandlePublicOAuthCallback(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("StoreController.HandlePublicOAuthCallback started",
		slog.String("request_id", requestID),
	)

	// Get query parameters from MercadoPago redirect
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorMsg := r.URL.Query().Get("error")
	errorDesc := r.URL.Query().Get("error_description")

	// Handle error from MercadoPago
	if errorMsg != "" {
		slog.Warn("StoreController.HandlePublicOAuthCallback received error from MercadoPago",
			slog.String("request_id", requestID),
			slog.String("error", errorMsg),
			slog.String("error_description", errorDesc),
		)
		redirectURL := fmt.Sprintf("/seller?mp_error=%s&mp_error_description=%s",
			url.QueryEscape(errorMsg),
			url.QueryEscape(errorDesc))
		http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
		return
	}

	// Validate required parameters
	if code == "" {
		slog.Warn("StoreController.HandlePublicOAuthCallback missing code parameter",
			slog.String("request_id", requestID),
		)
		http.Redirect(w, r, "/seller?mp_error=missing_code", http.StatusTemporaryRedirect)
		return
	}

	if state == "" {
		slog.Warn("StoreController.HandlePublicOAuthCallback missing state parameter",
			slog.String("request_id", requestID),
		)
		http.Redirect(w, r, "/seller?mp_error=missing_state", http.StatusTemporaryRedirect)
		return
	}

	slog.Debug("StoreController.HandlePublicOAuthCallback calling service",
		slog.String("request_id", requestID),
		slog.String("store_id", state),
	)

	// Exchange code for tokens and save connection
	if err := h.service.HandleOAuthCallback(r.Context(), state, code); err != nil {
		slog.Error("StoreController.HandlePublicOAuthCallback service error",
			slog.String("request_id", requestID),
			slog.String("store_id", state),
			slog.Any("error", err),
		)
		http.Redirect(w, r, "/seller?mp_error=connection_failed", http.StatusTemporaryRedirect)
		return
	}

	slog.Info("StoreController.HandlePublicOAuthCallback completed successfully",
		slog.String("request_id", requestID),
		slog.String("store_id", state),
	)

	// Redirect to seller dashboard with success message
	http.Redirect(w, r, "/seller?mp_connected=true", http.StatusTemporaryRedirect)
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
