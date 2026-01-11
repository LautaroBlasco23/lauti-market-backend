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
