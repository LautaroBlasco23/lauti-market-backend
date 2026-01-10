package dto

type RegisterUserRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type RegisterStoreRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phone_number"`
}

type RegisterResponse struct {
	AuthID      string `json:"auth_id"`
	AccountID   string `json:"account_id"`
	AccountType string `json:"account_type"`
	Email       string `json:"email"`
}
