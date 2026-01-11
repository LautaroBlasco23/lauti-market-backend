package dto

type RegisterUserRequest struct {
	Email     string `json:"email"       validate:"required,email"`
	Password  string `json:"password"    validate:"required,min=8,max=72"`
	FirstName string `json:"first_name"  validate:"required,min=2,max=50"`
	LastName  string `json:"last_name"   validate:"required,min=2,max=50"`
}

type RegisterStoreRequest struct {
	Email       string `json:"email"        validate:"required,email"`
	Password    string `json:"password"     validate:"required,min=8,max=72"`
	Name        string `json:"name"         validate:"required,min=2,max=100"`
	Description string `json:"description"  validate:"required,min=10,max=500"`
	Address     string `json:"address"      validate:"required,min=5,max=200"`
	PhoneNumber string `json:"phone_number" validate:"required,min=8,max=20"`
}

type RegisterResponse struct {
	AuthID      string `json:"auth_id"`
	AccountID   string `json:"account_id"`
	AccountType string `json:"account_type"`
	Email       string `json:"email"`
}
