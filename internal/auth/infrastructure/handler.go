package infrastructure

import (
	"encoding/json"
	"net/http"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

type registerUserRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type registerStoreRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	AuthID      string `json:"auth_id"`
	AccountID   string `json:"account_id"`
	AccountType string `json:"account_type"`
	Email       string `json:"email"`
}

type loginResponse struct {
	Token       string `json:"token"`
	AccountID   string `json:"account_id"`
	AccountType string `json:"account_type"`
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req registerUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
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
	json.NewEncoder(w).Encode(registerResponse{
		AuthID:      string(output.AuthID),
		AccountID:   string(output.AccountID),
		AccountType: string(output.AccountType),
		Email:       output.Email,
	})
}

func (h *Handler) RegisterStore(w http.ResponseWriter, r *http.Request) {
	var req registerStoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
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
	json.NewEncoder(w).Encode(registerResponse{
		AuthID:      string(output.AuthID),
		AccountID:   string(output.AccountID),
		AccountType: string(output.AccountType),
		Email:       output.Email,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
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
	json.NewEncoder(w).Encode(loginResponse{
		Token:       output.Token,
		AccountID:   string(output.AccountID),
		AccountType: string(output.AccountType),
	})
}

func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrEmailExists:
		http.Error(w, err.Error(), http.StatusConflict)
	case domain.ErrInvalidCredentials:
		http.Error(w, err.Error(), http.StatusUnauthorized)
	case domain.ErrInvalidEmail, domain.ErrInvalidPassword,
		domain.ErrInvalidAccountID, domain.ErrInvalidAccountType:
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
