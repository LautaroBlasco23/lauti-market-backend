package controller

import (
	"encoding/json"
	"net/http"

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
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing user id")
		return
	}

	callerID, _ := infrastructure.AccountIDFromContext(r)
	if callerID != id {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	output, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, userDto.UserResponse{
		ID:        string(output.ID),
		FirstName: output.FirstName,
		LastName:  output.LastName,
	})
}

func (h *UserController) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing user id")
		return
	}

	callerID, _ := infrastructure.AccountIDFromContext(r)
	if callerID != id {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	var req userDto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
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

	output, err := h.service.Update(r.Context(), application.UpdateInput{
		ID:        id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, userDto.UserResponse{
		ID:        string(output.ID),
		FirstName: output.FirstName,
		LastName:  output.LastName,
	})
}

func (h *UserController) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing user id")
		return
	}

	callerID, _ := infrastructure.AccountIDFromContext(r)
	if callerID != id {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
