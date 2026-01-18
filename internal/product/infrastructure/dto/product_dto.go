package dto

type CreateProductRequest struct {
	Name        string  `json:"name"        validate:"required,min=2,max=100"`
	Description string  `json:"description" validate:"required,min=10,max=500"`
	Stock       int     `json:"stock"       validate:"gte=0"`
	Price       float64 `json:"price"       validate:"required,gt=0"`
	ImageURL    *string `json:"image_url"   validate:"omitempty,url"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name"        validate:"required,min=2,max=100"`
	Description string  `json:"description" validate:"required,min=10,max=500"`
	Stock       int     `json:"stock"       validate:"gte=0"`
	Price       float64 `json:"price"       validate:"required,gt=0"`
	ImageURL    *string `json:"image_url"   validate:"omitempty,url"`
}

type ProductResponse struct {
	ID          string  `json:"id"`
	StoreID     string  `json:"store_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Stock       int     `json:"stock"`
	Price       float64 `json:"price"`
	ImageURL    *string `json:"image_url,omitempty"`
}
