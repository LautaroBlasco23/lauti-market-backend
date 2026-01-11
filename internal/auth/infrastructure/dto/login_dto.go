package dto

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token       string `json:"token"`
	AccountID   string `json:"account_id"`
	AccountType string `json:"account_type"`
}
