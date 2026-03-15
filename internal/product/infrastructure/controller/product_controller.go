package controller

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

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
		Category:    p.Category(),
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

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "store" {
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}
	if claims.AccountID != storeID {
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	stock, err := strconv.Atoi(r.FormValue("stock"))
	if err != nil {
		http.Error(w, "invalid stock value", http.StatusBadRequest)
		return
	}
	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		http.Error(w, "invalid price value", http.StatusBadRequest)
		return
	}

	req := dto.CreateProductRequest{
		Name:        r.FormValue("name"),
		Description: r.FormValue("description"),
		Category:    r.FormValue("category"),
		Stock:       stock,
		Price:       price,
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

	input := application.CreateProductInput{
		StoreID:     storeID,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Stock:       req.Stock,
		Price:       req.Price,
	}

	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "failed to read image", http.StatusInternalServerError)
			return
		}
		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		input.ImageData = data
		input.ImageFilename = header.Filename
		input.ImageContentType = contentType
	}

	product, err := c.service.Create(r.Context(), input)
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

func (c *ProductController) GetAll(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := 10
	offset := 0

	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	var category *string
	if v := q.Get("category"); v != "" {
		category = &v
	}

	products, err := c.service.GetAll(r.Context(), application.GetAllProductsInput{
		Limit:    limit,
		Offset:   offset,
		Category: category,
	})
	if err != nil {
		c.handleError(w, err)
		return
	}

	responses := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		responses[i] = toProductResponse(p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.ProductListResponse{
		Products: responses,
		Limit:    limit,
		Offset:   offset,
	})
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
	storeID := r.PathValue("store_id")
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing product id", http.StatusBadRequest)
		return
	}

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "store" {
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}
	if claims.AccountID != storeID {
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
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
		Category:    req.Category,
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
	storeID := r.PathValue("store_id")
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing product id", http.StatusBadRequest)
		return
	}

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "store" {
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}
	if claims.AccountID != storeID {
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	if err := c.service.Delete(r.Context(), id); err != nil {
		c.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *ProductController) UploadImage(w http.ResponseWriter, r *http.Request) {
	storeID := r.PathValue("store_id")
	productID := r.PathValue("id")
	if storeID == "" || productID == "" {
		http.Error(w, "missing store_id or product id", http.StatusBadRequest)
		return
	}

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "store" {
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}
	if claims.AccountID != storeID {
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "image too large or invalid form", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "missing image field", http.StatusBadRequest)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read image", http.StatusInternalServerError)
		return
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	product, err := c.service.UploadImage(r.Context(), application.UploadProductImageInput{
		ProductID: productID, StoreID: storeID,
		Filename: header.Filename, ContentType: contentType, Data: data,
	})
	if err != nil {
		c.handleError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toProductResponse(product))
}

func (c *ProductController) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apiDomain.ErrProductNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, storeDomain.ErrStoreNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, apiDomain.ErrInvalidProductName), errors.Is(err, apiDomain.ErrInvalidProductDescription), errors.Is(err, apiDomain.ErrInvalidStock), errors.Is(err, apiDomain.ErrInvalidPrice), errors.Is(err, apiDomain.ErrInvalidStoreID), errors.Is(err, apiDomain.ErrInvalidCategory):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
