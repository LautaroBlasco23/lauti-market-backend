package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/product/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/product/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/product/infrastructure/dto"
	storeDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

type ProductController struct {
	service *application.ProductService
}

func NewProductController(service *application.ProductService) *ProductController {
	return &ProductController{service: service}
}

func toProductResponse(p *domain.Product) dto.ProductResponse {
	return dto.ProductResponse{
		ID:          p.ID(),
		StoreID:     p.StoreID(),
		Name:        p.Name(),
		Description: p.Description(),
		Stock:       p.Stock(),
		Price:       p.Price(),
		ImageURL:    p.ImageURL(),
	}
}

func (c *ProductController) Create(w http.ResponseWriter, r *http.Request) {
	storeID := r.PathValue("store_id")
	if storeID == "" {
		http.Error(w, "missing store id", http.StatusBadRequest)
		return
	}

	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
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

	product, err := c.service.Create(r.Context(), application.CreateProductInput{
		StoreID:     storeID,
		Name:        req.Name,
		Description: req.Description,
		Stock:       req.Stock,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
	})
	if err != nil {
		c.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toProductResponse(product))
}

func (c *ProductController) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing product id", http.StatusBadRequest)
		return
	}

	product, err := c.service.GetByID(r.Context(), id)
	if err != nil {
		c.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toProductResponse(product))
}

func (c *ProductController) GetByStoreID(w http.ResponseWriter, r *http.Request) {
	storeID := r.PathValue("store_id")
	if storeID == "" {
		http.Error(w, "missing store id", http.StatusBadRequest)
		return
	}

	products, err := c.service.GetByStoreID(r.Context(), storeID, 100, 0)
	if err != nil {
		c.handleError(w, err)
		return
	}

	response := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		response[i] = toProductResponse(p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *ProductController) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing product id", http.StatusBadRequest)
		return
	}

	var req dto.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
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

	product, err := c.service.Update(r.Context(), application.UpdateProductInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Stock:       req.Stock,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
	})
	if err != nil {
		c.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toProductResponse(product))
}

func (c *ProductController) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing product id", http.StatusBadRequest)
		return
	}

	if err := c.service.Delete(r.Context(), id); err != nil {
		c.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *ProductController) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apiDomain.ErrProductNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, storeDomain.ErrStoreNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, apiDomain.ErrInvalidProductName), errors.Is(err, apiDomain.ErrInvalidProductDescription), errors.Is(err, apiDomain.ErrInvalidStock), errors.Is(err, apiDomain.ErrInvalidPrice), errors.Is(err, apiDomain.ErrInvalidStoreID):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
