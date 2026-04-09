package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/application"
	userDto "github.com/LautaroBlasco23/lauti-market-backend/internal/user/infrastructure/dto"
)

type UserController struct {
	service *application.UserService
}

func NewUserController(service *application.UserService) *UserController {
	return &UserController{service: service}
}

func (h *UserController) GetByID(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("UserController.GetByID started",
		slog.String("request_id", requestID),
	)

	id := r.PathValue("id")
	if id == "" {
		slog.Warn("UserController.GetByID missing user id",
			slog.String("request_id", requestID),
		)
		writeError(w, http.StatusBadRequest, "missing user id")
		return
	}

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		slog.Warn("UserController.GetByID unauthorized",
			slog.String("request_id", requestID),
			slog.String("user_id", id),
		)
		writeError(w, http.StatusUnauthorized, apiDomain.ErrUnauthorized.Error())
		return
	}
	if claims.AccountID != id {
		slog.Warn("UserController.GetByID forbidden - account mismatch",
			slog.String("request_id", requestID),
			slog.String("user_id", id),
			slog.String("account_id", claims.AccountID),
		)
		writeError(w, http.StatusForbidden, apiDomain.ErrForbidden.Error())
		return
	}

	slog.Debug("UserController.GetByID calling service",
		slog.String("request_id", requestID),
		slog.String("user_id", id),
	)

	output, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		slog.Error("UserController.GetByID service error",
			slog.String("request_id", requestID),
			slog.String("user_id", id),
			slog.Any("error", err),
		)
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	slog.Info("UserController.GetByID completed",
		slog.String("request_id", requestID),
		slog.String("user_id", id),
	)

	writeJSON(w, http.StatusOK, userDto.UserResponse{
		ID:        string(output.ID),
		FirstName: output.FirstName,
		LastName:  output.LastName,
	})
}

func (h *UserController) Update(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("UserController.Update started",
		slog.String("request_id", requestID),
	)

	id := r.PathValue("id")
	if id == "" {
		slog.Warn("UserController.Update missing user id",
			slog.String("request_id", requestID),
		)
		writeError(w, http.StatusBadRequest, "missing user id")
		return
	}

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		slog.Warn("UserController.Update unauthorized",
			slog.String("request_id", requestID),
			slog.String("user_id", id),
		)
		writeError(w, http.StatusUnauthorized, apiDomain.ErrUnauthorized.Error())
		return
	}
	if string(claims.AccountType) != "user" {
		slog.Warn("UserController.Update forbidden - wrong account type",
			slog.String("request_id", requestID),
			slog.String("user_id", id),
			slog.String("account_type", string(claims.AccountType)),
		)
		writeError(w, http.StatusForbidden, apiDomain.ErrForbidden.Error())
		return
	}
	if claims.AccountID != id {
		slog.Warn("UserController.Update forbidden - account mismatch",
			slog.String("request_id", requestID),
			slog.String("user_id", id),
			slog.String("account_id", claims.AccountID),
		)
		writeError(w, http.StatusForbidden, apiDomain.ErrForbidden.Error())
		return
	}

	var req userDto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("UserController.Update invalid request body",
			slog.String("request_id", requestID),
			slog.String("user_id", id),
			slog.Any("error", err),
		)
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := infrastructure.Validate(req); err != nil {
		slog.Warn("UserController.Update validation failed",
			slog.String("request_id", requestID),
			slog.String("user_id", id),
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

	slog.Debug("UserController.Update calling service",
		slog.String("request_id", requestID),
		slog.String("user_id", id),
	)

	output, err := h.service.Update(r.Context(), application.UpdateInput{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		slog.Error("UserController.Update service error",
			slog.String("request_id", requestID),
			slog.String("user_id", id),
			slog.Any("error", err),
		)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	slog.Info("UserController.Update completed",
		slog.String("request_id", requestID),
		slog.String("user_id", id),
	)

	writeJSON(w, http.StatusOK, userDto.UserResponse{
		ID:        string(output.ID),
		FirstName: output.FirstName,
		LastName:  output.LastName,
	})
}

func (h *UserController) Delete(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("UserController.Delete started",
		slog.String("request_id", requestID),
	)

	id := r.PathValue("id")
	if id == "" {
		slog.Warn("UserController.Delete missing user id",
			slog.String("request_id", requestID),
		)
		writeError(w, http.StatusBadRequest, "missing user id")
		return
	}

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		slog.Warn("UserController.Delete unauthorized",
			slog.String("request_id", requestID),
			slog.String("user_id", id),
		)
		writeError(w, http.StatusUnauthorized, apiDomain.ErrUnauthorized.Error())
		return
	}
	if string(claims.AccountType) != "user" {
		slog.Warn("UserController.Delete forbidden - wrong account type",
			slog.String("request_id", requestID),
			slog.String("user_id", id),
			slog.String("account_type", string(claims.AccountType)),
		)
		writeError(w, http.StatusForbidden, apiDomain.ErrForbidden.Error())
		return
	}
	if claims.AccountID != id {
		slog.Warn("UserController.Delete forbidden - account mismatch",
			slog.String("request_id", requestID),
			slog.String("user_id", id),
			slog.String("account_id", claims.AccountID),
		)
		writeError(w, http.StatusForbidden, apiDomain.ErrForbidden.Error())
		return
	}

	slog.Debug("UserController.Delete calling service",
		slog.String("request_id", requestID),
		slog.String("user_id", id),
	)

	if err := h.service.Delete(r.Context(), id); err != nil {
		slog.Error("UserController.Delete service error",
			slog.String("request_id", requestID),
			slog.String("user_id", id),
			slog.Any("error", err),
		)
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	slog.Info("UserController.Delete completed",
		slog.String("request_id", requestID),
		slog.String("user_id", id),
	)

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data) //nolint:errcheck
}

func writeError(w http.ResponseWriter, status int, message string) {
	slog.Debug("UserController writing error response",
		slog.Int("status", status),
		slog.String("message", message),
	)
	writeJSON(w, status, map[string]string{"error": message})
}
