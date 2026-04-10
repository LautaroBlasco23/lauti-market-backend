package dto

type UpdateStoreRequest struct {
	Name        string `json:"name"         validate:"omitempty,min=2,max=100"`
	Description string `json:"description"  validate:"omitempty,min=10,max=500"`
	Address     string `json:"address"      validate:"omitempty,min=5,max=200"`
	PhoneNumber string `json:"phone_number" validate:"omitempty,min=8,max=20"`
}

type StoreResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
}

// OAuth DTOs

type OAuthCallbackRequest struct {
	Code string `json:"code" validate:"required"`
}

type OAuthConnectResponse struct {
	AuthURL string `json:"auth_url"`
}

type MPConnectionStatusResponse struct {
	Connected    bool   `json:"connected"`
	ConnectedAt  string `json:"connected_at,omitempty"`
	ExpiresAt    string `json:"expires_at,omitempty"`
	IsTokenValid bool   `json:"is_token_valid"`
}
