package controller

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
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
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("ProductController.Create started",
		slog.String("request_id", requestID),
	)

	storeID := r.PathValue("store_id")
	if storeID == "" {
		slog.Warn("ProductController.Create failed: missing store id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing store id", http.StatusBadRequest)
		return
	}

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		slog.Warn("ProductController.Create failed: unauthorized",
			slog.String("request_id", requestID),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "store" {
		slog.Warn("ProductController.Create failed: forbidden - not a store account",
			slog.String("request_id", requestID),
			slog.String("account_type", string(claims.AccountType)),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}
	if claims.AccountID != storeID {
		slog.Warn("ProductController.Create failed: forbidden - account ID mismatch",
			slog.String("request_id", requestID),
			slog.String("account_id", claims.AccountID),
			slog.String("store_id", storeID),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		slog.Warn("ProductController.Create failed: invalid form",
			slog.String("request_id", requestID),
			slog.Any("error", err),
		)
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	stock, stockErr := strconv.Atoi(r.FormValue("stock"))
	if stockErr != nil {
		slog.Warn("ProductController.Create failed: invalid stock value",
			slog.String("request_id", requestID),
			slog.String("stock_value", r.FormValue("stock")),
		)
		http.Error(w, "invalid stock value", http.StatusBadRequest)
		return
	}
	price, priceErr := strconv.ParseFloat(r.FormValue("price"), 64)
	if priceErr != nil {
		slog.Warn("ProductController.Create failed: invalid price value",
			slog.String("request_id", requestID),
			slog.String("price_value", r.FormValue("price")),
		)
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

	if validateErr := infrastructure.Validate(req); validateErr != nil {
		slog.Warn("ProductController.Create failed: validation error",
			slog.String("request_id", requestID),
			slog.Any("fields", infrastructure.FieldErrors(validateErr)),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"error":  "invalid_payload",
			"fields": infrastructure.FieldErrors(validateErr),
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
		defer func() {
			_ = file.Close() //nolint:errcheck
		}()
		data, readErr := io.ReadAll(file)
		if readErr != nil {
			slog.Error("ProductController.Create failed: failed to read image",
				slog.String("request_id", requestID),
				slog.Any("error", readErr),
			)
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

	slog.Debug("ProductController.Create calling service",
		slog.String("request_id", requestID),
		slog.String("store_id", storeID),
		slog.String("product_name", req.Name),
	)
	product, err := c.service.Create(r.Context(), &input)
	if err != nil {
		slog.Error("ProductController.Create failed",
			slog.String("request_id", requestID),
			slog.Any("error", err),
		)
		c.handleError(w, err)
		return
	}

	slog.Info("ProductController.Create completed",
		slog.String("request_id", requestID),
		slog.String("product_id", product.ID()),
		slog.String("store_id", storeID),
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(toProductResponse(product)) //nolint:errcheck
}

func (c *ProductController) GetByID(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("ProductController.GetByID started",
		slog.String("request_id", requestID),
	)

	id := r.PathValue("id")
	if id == "" {
		slog.Warn("ProductController.GetByID failed: missing product id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing product id", http.StatusBadRequest)
		return
	}

	slog.Debug("ProductController.GetByID calling service",
		slog.String("request_id", requestID),
		slog.String("product_id", id),
	)
	product, err := c.service.GetByID(r.Context(), id)
	if err != nil {
		slog.Error("ProductController.GetByID failed",
			slog.String("request_id", requestID),
			slog.String("product_id", id),
			slog.Any("error", err),
		)
		c.handleError(w, err)
		return
	}

	slog.Info("ProductController.GetByID completed",
		slog.String("request_id", requestID),
		slog.String("product_id", product.ID()),
	)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toProductResponse(product)) //nolint:errcheck
}

func (c *ProductController) GetAll(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("ProductController.GetAll started",
		slog.String("request_id", requestID),
	)

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

	slog.Debug("ProductController.GetAll calling service",
		slog.String("request_id", requestID),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)
	products, err := c.service.GetAll(r.Context(), application.GetAllProductsInput{
		Limit:    limit,
		Offset:   offset,
		Category: category,
	})
	if err != nil {
		slog.Error("ProductController.GetAll failed",
			slog.String("request_id", requestID),
			slog.Any("error", err),
		)
		c.handleError(w, err)
		return
	}

	responses := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		responses[i] = toProductResponse(p)
	}

	slog.Info("ProductController.GetAll completed",
		slog.String("request_id", requestID),
		slog.Int("count", len(products)),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(dto.ProductListResponse{ //nolint:errcheck
		Products: responses,
		Limit:    limit,
		Offset:   offset,
	})
}

func (c *ProductController) GetByStoreID(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("ProductController.GetByStoreID started",
		slog.String("request_id", requestID),
	)

	storeID := r.PathValue("store_id")
	if storeID == "" {
		slog.Warn("ProductController.GetByStoreID failed: missing store id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing store id", http.StatusBadRequest)
		return
	}

	slog.Debug("ProductController.GetByStoreID calling service",
		slog.String("request_id", requestID),
		slog.String("store_id", storeID),
	)
	products, err := c.service.GetByStoreID(r.Context(), storeID, 100, 0)
	if err != nil {
		slog.Error("ProductController.GetByStoreID failed",
			slog.String("request_id", requestID),
			slog.String("store_id", storeID),
			slog.Any("error", err),
		)
		c.handleError(w, err)
		return
	}

	response := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		response[i] = toProductResponse(p)
	}

	slog.Info("ProductController.GetByStoreID completed",
		slog.String("request_id", requestID),
		slog.String("store_id", storeID),
		slog.Int("count", len(products)),
	)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response) //nolint:errcheck
}

func (c *ProductController) Update(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("ProductController.Update started",
		slog.String("request_id", requestID),
	)

	storeID := r.PathValue("store_id")
	id := r.PathValue("id")
	if id == "" {
		slog.Warn("ProductController.Update failed: missing product id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing product id", http.StatusBadRequest)
		return
	}

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		slog.Warn("ProductController.Update failed: unauthorized",
			slog.String("request_id", requestID),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "store" {
		slog.Warn("ProductController.Update failed: forbidden - not a store account",
			slog.String("request_id", requestID),
			slog.String("account_type", string(claims.AccountType)),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}
	if claims.AccountID != storeID {
		slog.Warn("ProductController.Update failed: forbidden - account ID mismatch",
			slog.String("request_id", requestID),
			slog.String("account_id", claims.AccountID),
			slog.String("store_id", storeID),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	var req dto.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("ProductController.Update failed: invalid request body",
			slog.String("request_id", requestID),
			slog.Any("error", err),
		)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := infrastructure.Validate(req); err != nil {
		slog.Warn("ProductController.Update failed: validation error",
			slog.String("request_id", requestID),
			slog.Any("fields", infrastructure.FieldErrors(err)),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"error":  "invalid_payload",
			"fields": infrastructure.FieldErrors(err),
		})
		return
	}

	slog.Debug("ProductController.Update calling service",
		slog.String("request_id", requestID),
		slog.String("product_id", id),
		slog.String("store_id", storeID),
	)
	product, err := c.service.Update(r.Context(), &application.UpdateProductInput{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Stock:       req.Stock,
		Price:       req.Price,
		ImageURL:    req.ImageURL,
	})
	if err != nil {
		slog.Error("ProductController.Update failed",
			slog.String("request_id", requestID),
			slog.String("product_id", id),
			slog.Any("error", err),
		)
		c.handleError(w, err)
		return
	}

	slog.Info("ProductController.Update completed",
		slog.String("request_id", requestID),
		slog.String("product_id", product.ID()),
	)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toProductResponse(product)) //nolint:errcheck
}

func (c *ProductController) Delete(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("ProductController.Delete started",
		slog.String("request_id", requestID),
	)

	storeID := r.PathValue("store_id")
	id := r.PathValue("id")
	if id == "" {
		slog.Warn("ProductController.Delete failed: missing product id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing product id", http.StatusBadRequest)
		return
	}

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		slog.Warn("ProductController.Delete failed: unauthorized",
			slog.String("request_id", requestID),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "store" {
		slog.Warn("ProductController.Delete failed: forbidden - not a store account",
			slog.String("request_id", requestID),
			slog.String("account_type", string(claims.AccountType)),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}
	if claims.AccountID != storeID {
		slog.Warn("ProductController.Delete failed: forbidden - account ID mismatch",
			slog.String("request_id", requestID),
			slog.String("account_id", claims.AccountID),
			slog.String("store_id", storeID),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	slog.Debug("ProductController.Delete calling service",
		slog.String("request_id", requestID),
		slog.String("product_id", id),
		slog.String("store_id", storeID),
	)
	if err := c.service.Delete(r.Context(), id); err != nil {
		slog.Error("ProductController.Delete failed",
			slog.String("request_id", requestID),
			slog.String("product_id", id),
			slog.Any("error", err),
		)
		c.handleError(w, err)
		return
	}

	slog.Info("ProductController.Delete completed",
		slog.String("request_id", requestID),
		slog.String("product_id", id),
	)
	w.WriteHeader(http.StatusNoContent)
}

func (c *ProductController) UploadImage(w http.ResponseWriter, r *http.Request) {
	requestID := infrastructure.GetRequestID(r)
	slog.Debug("ProductController.UploadImage started",
		slog.String("request_id", requestID),
	)

	storeID := r.PathValue("store_id")
	productID := r.PathValue("id")
	if storeID == "" || productID == "" {
		slog.Warn("ProductController.UploadImage failed: missing store_id or product id",
			slog.String("request_id", requestID),
			slog.String("store_id", storeID),
			slog.String("product_id", productID),
		)
		http.Error(w, "missing store_id or product id", http.StatusBadRequest)
		return
	}

	claims, ok := infrastructure.GetClaims(r.Context())
	if !ok {
		slog.Warn("ProductController.UploadImage failed: unauthorized",
			slog.String("request_id", requestID),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "store" {
		slog.Warn("ProductController.UploadImage failed: forbidden - not a store account",
			slog.String("request_id", requestID),
			slog.String("account_type", string(claims.AccountType)),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}
	if claims.AccountID != storeID {
		slog.Warn("ProductController.UploadImage failed: forbidden - account ID mismatch",
			slog.String("request_id", requestID),
			slog.String("account_id", claims.AccountID),
			slog.String("store_id", storeID),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		slog.Warn("ProductController.UploadImage failed: invalid form",
			slog.String("request_id", requestID),
			slog.Any("error", err),
		)
		http.Error(w, "image too large or invalid form", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		slog.Warn("ProductController.UploadImage failed: missing image field",
			slog.String("request_id", requestID),
			slog.Any("error", err),
		)
		http.Error(w, "missing image field", http.StatusBadRequest)
		return
	}
	defer func() {
		_ = file.Close() //nolint:errcheck
	}()
	data, err := io.ReadAll(file)
	if err != nil {
		slog.Error("ProductController.UploadImage failed: failed to read image",
			slog.String("request_id", requestID),
			slog.Any("error", err),
		)
		http.Error(w, "failed to read image", http.StatusInternalServerError)
		return
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	slog.Debug("ProductController.UploadImage calling service",
		slog.String("request_id", requestID),
		slog.String("product_id", productID),
		slog.String("store_id", storeID),
		slog.String("filename", header.Filename),
	)
	product, err := c.service.UploadImage(r.Context(), &application.UploadProductImageInput{
		ProductID: productID, StoreID: storeID,
		Filename: header.Filename, ContentType: contentType, Data: data,
	})
	if err != nil {
		slog.Error("ProductController.UploadImage failed",
			slog.String("request_id", requestID),
			slog.String("product_id", productID),
			slog.Any("error", err),
		)
		c.handleError(w, err)
		return
	}

	slog.Info("ProductController.UploadImage completed",
		slog.String("request_id", requestID),
		slog.String("product_id", product.ID()),
	)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toProductResponse(product)) //nolint:errcheck
}

func (c *ProductController) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apiDomain.ErrProductNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, storeDomain.ErrStoreNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, apiDomain.ErrStoreMPNotConnected):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, apiDomain.ErrStoreMPTokenExpired):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, apiDomain.ErrInvalidProductName), errors.Is(err, apiDomain.ErrInvalidProductDescription), errors.Is(err, apiDomain.ErrInvalidStock), errors.Is(err, apiDomain.ErrInvalidPrice), errors.Is(err, apiDomain.ErrInvalidStoreID), errors.Is(err, apiDomain.ErrInvalidCategory):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		slog.Error("ProductController: internal server error", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
