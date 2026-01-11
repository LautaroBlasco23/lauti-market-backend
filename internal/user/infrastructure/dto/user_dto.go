package dto

type UpdateUserRequest struct {
	FirstName string `json:"first_name" validate:"omitempty,min=2,max=50"`
	LastName  string `json:"last_name"  validate:"omitempty,min=2,max=50"`
}
