package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	apiInfra "github.com/LautaroBlasco23/lauti-market-backend/internal/api/infrastructure"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/order/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/order/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/order/infrastructure/dto"
)

type OrderController struct {
	service *application.OrderService
}

func NewOrderController(service *application.OrderService) *OrderController {
	return &OrderController{service: service}
}

func toOrderResponse(o *domain.Order) dto.OrderResponse {
	items := make([]dto.OrderItemResponse, len(o.Items()))
	for i, item := range o.Items() {
		items[i] = dto.OrderItemResponse{
			ID:        item.ID(),
			ProductID: item.ProductID(),
			Quantity:  item.Quantity(),
			UnitPrice: item.UnitPrice(),
		}
	}
	return dto.OrderResponse{
		ID:         o.ID(),
		UserID:     o.UserID(),
		StoreID:    o.StoreID(),
		Status:     string(o.Status()),
		Items:      items,
		TotalPrice: o.TotalPrice(),
		CreatedAt:  o.CreatedAt(),
		UpdatedAt:  o.UpdatedAt(),
	}
}

func (c *OrderController) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := apiInfra.GetClaims(r.Context())
	if !ok {
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "user" {
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	var req dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := apiInfra.Validate(req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error":  "invalid_payload",
			"fields": apiInfra.FieldErrors(err),
		})
		return
	}

	items := make([]application.OrderItemInput, len(req.Items))
	for i, item := range req.Items {
		items[i] = application.OrderItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	order, err := c.service.CreateOrder(r.Context(), application.CreateOrderInput{
		UserID:  claims.AccountID,
		StoreID: req.StoreID,
		Items:   items,
	})
	if err != nil {
		c.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toOrderResponse(order))
}

func (c *OrderController) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing order id", http.StatusBadRequest)
		return
	}

	order, err := c.service.GetByID(r.Context(), id)
	if err != nil {
		c.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toOrderResponse(order))
}

func (c *OrderController) GetByUserID(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	if userID == "" {
		http.Error(w, "missing user id", http.StatusBadRequest)
		return
	}

	limit, offset := parsePagination(r)
	orders, err := c.service.GetByUserID(r.Context(), userID, limit, offset)
	if err != nil {
		c.handleError(w, err)
		return
	}

	response := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		response[i] = toOrderResponse(o)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *OrderController) GetByStoreID(w http.ResponseWriter, r *http.Request) {
	storeID := r.PathValue("store_id")
	if storeID == "" {
		http.Error(w, "missing store id", http.StatusBadRequest)
		return
	}

	limit, offset := parsePagination(r)
	orders, err := c.service.GetByStoreID(r.Context(), storeID, limit, offset)
	if err != nil {
		c.handleError(w, err)
		return
	}

	response := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		response[i] = toOrderResponse(o)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *OrderController) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing order id", http.StatusBadRequest)
		return
	}

	claims, ok := apiInfra.GetClaims(r.Context())
	if !ok {
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := apiInfra.Validate(req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error":  "invalid_payload",
			"fields": apiInfra.FieldErrors(err),
		})
		return
	}

	newStatus := domain.OrderStatus(req.Status)
	if !newStatus.IsValid() {
		http.Error(w, apiDomain.ErrInvalidOrderStatus.Error(), http.StatusBadRequest)
		return
	}

	order, err := c.service.UpdateStatus(r.Context(), application.UpdateStatusInput{
		OrderID:     id,
		NewStatus:   newStatus,
		AccountType: string(claims.AccountType),
		AccountID:   claims.AccountID,
	})
	if err != nil {
		c.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toOrderResponse(order))
}

func (c *OrderController) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apiDomain.ErrOrderNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, apiDomain.ErrInsufficientStock),
		errors.Is(err, apiDomain.ErrEmptyOrderItems),
		errors.Is(err, apiDomain.ErrInvalidQuantity),
		errors.Is(err, apiDomain.ErrInvalidOrderStatus),
		errors.Is(err, apiDomain.ErrItemsFromMultipleStores):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, apiDomain.ErrForbiddenTransition),
		errors.Is(err, apiDomain.ErrForbidden):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, apiDomain.ErrUnauthorized):
		http.Error(w, err.Error(), http.StatusUnauthorized)
	default:
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func parsePagination(r *http.Request) (limit, offset int) {
	limit = 10
	offset = 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	return
}
