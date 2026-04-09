package controller

import (
	"encoding/json"
	"errors"
	"log/slog"
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
	requestID := apiInfra.GetRequestID(r)
	slog.Debug("OrderController.Create started",
		slog.String("request_id", requestID),
	)

	claims, ok := apiInfra.GetClaims(r.Context())
	if !ok {
		slog.Warn("OrderController.Create failed: unauthorized",
			slog.String("request_id", requestID),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "user" {
		slog.Warn("OrderController.Create failed: forbidden - not a user account",
			slog.String("request_id", requestID),
			slog.String("account_type", string(claims.AccountType)),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	var req dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("OrderController.Create failed: invalid request body",
			slog.String("request_id", requestID),
			slog.Any("error", err),
		)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := apiInfra.Validate(req); err != nil {
		slog.Warn("OrderController.Create failed: validation error",
			slog.String("request_id", requestID),
			slog.Any("fields", apiInfra.FieldErrors(err)),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
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

	slog.Debug("OrderController.Create calling service",
		slog.String("request_id", requestID),
		slog.String("user_id", claims.AccountID),
		slog.String("store_id", req.StoreID),
		slog.Int("item_count", len(items)),
	)
	order, err := c.service.CreateOrder(r.Context(), application.CreateOrderInput{
		UserID:  claims.AccountID,
		StoreID: req.StoreID,
		Items:   items,
	})
	if err != nil {
		slog.Error("OrderController.Create failed",
			slog.String("request_id", requestID),
			slog.Any("error", err),
		)
		c.handleError(w, err)
		return
	}

	slog.Info("OrderController.Create completed",
		slog.String("request_id", requestID),
		slog.String("order_id", order.ID()),
		slog.String("user_id", claims.AccountID),
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(toOrderResponse(order)) //nolint:errcheck
}

func (c *OrderController) GetByID(w http.ResponseWriter, r *http.Request) {
	requestID := apiInfra.GetRequestID(r)
	slog.Debug("OrderController.GetByID started",
		slog.String("request_id", requestID),
	)

	id := r.PathValue("id")
	if id == "" {
		slog.Warn("OrderController.GetByID failed: missing order id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing order id", http.StatusBadRequest)
		return
	}

	claims, ok := apiInfra.GetClaims(r.Context())
	if !ok {
		slog.Warn("OrderController.GetByID failed: unauthorized",
			slog.String("request_id", requestID),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	slog.Debug("OrderController.GetByID calling service",
		slog.String("request_id", requestID),
		slog.String("order_id", id),
	)
	order, err := c.service.GetByID(r.Context(), id)
	if err != nil {
		slog.Error("OrderController.GetByID failed",
			slog.String("request_id", requestID),
			slog.String("order_id", id),
			slog.Any("error", err),
		)
		c.handleError(w, err)
		return
	}

	isOwner := (string(claims.AccountType) == "user" && claims.AccountID == order.UserID()) ||
		(string(claims.AccountType) == "store" && claims.AccountID == order.StoreID())
	if !isOwner {
		slog.Warn("OrderController.GetByID failed: forbidden - not order owner",
			slog.String("request_id", requestID),
			slog.String("account_id", claims.AccountID),
			slog.String("account_type", string(claims.AccountType)),
			slog.String("order_user_id", order.UserID()),
			slog.String("order_store_id", order.StoreID()),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	slog.Info("OrderController.GetByID completed",
		slog.String("request_id", requestID),
		slog.String("order_id", order.ID()),
	)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toOrderResponse(order)) //nolint:errcheck
}

func (c *OrderController) GetByUserID(w http.ResponseWriter, r *http.Request) {
	requestID := apiInfra.GetRequestID(r)
	slog.Debug("OrderController.GetByUserID started",
		slog.String("request_id", requestID),
	)

	userID := r.PathValue("user_id")
	if userID == "" {
		slog.Warn("OrderController.GetByUserID failed: missing user id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing user id", http.StatusBadRequest)
		return
	}

	claims, ok := apiInfra.GetClaims(r.Context())
	if !ok {
		slog.Warn("OrderController.GetByUserID failed: unauthorized",
			slog.String("request_id", requestID),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "user" || claims.AccountID != userID {
		slog.Warn("OrderController.GetByUserID failed: forbidden - not owner",
			slog.String("request_id", requestID),
			slog.String("account_id", claims.AccountID),
			slog.String("account_type", string(claims.AccountType)),
			slog.String("requested_user_id", userID),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	limit, offset := parsePagination(r)
	slog.Debug("OrderController.GetByUserID calling service",
		slog.String("request_id", requestID),
		slog.String("user_id", userID),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)
	orders, err := c.service.GetByUserID(r.Context(), userID, limit, offset)
	if err != nil {
		slog.Error("OrderController.GetByUserID failed",
			slog.String("request_id", requestID),
			slog.String("user_id", userID),
			slog.Any("error", err),
		)
		c.handleError(w, err)
		return
	}

	response := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		response[i] = toOrderResponse(o)
	}

	slog.Info("OrderController.GetByUserID completed",
		slog.String("request_id", requestID),
		slog.String("user_id", userID),
		slog.Int("count", len(orders)),
	)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response) //nolint:errcheck
}

func (c *OrderController) GetByStoreID(w http.ResponseWriter, r *http.Request) {
	requestID := apiInfra.GetRequestID(r)
	slog.Debug("OrderController.GetByStoreID started",
		slog.String("request_id", requestID),
	)

	storeID := r.PathValue("store_id")
	if storeID == "" {
		slog.Warn("OrderController.GetByStoreID failed: missing store id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing store id", http.StatusBadRequest)
		return
	}

	claims, ok := apiInfra.GetClaims(r.Context())
	if !ok {
		slog.Warn("OrderController.GetByStoreID failed: unauthorized",
			slog.String("request_id", requestID),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	if string(claims.AccountType) != "store" || claims.AccountID != storeID {
		slog.Warn("OrderController.GetByStoreID failed: forbidden - not owner",
			slog.String("request_id", requestID),
			slog.String("account_id", claims.AccountID),
			slog.String("account_type", string(claims.AccountType)),
			slog.String("requested_store_id", storeID),
		)
		http.Error(w, apiDomain.ErrForbidden.Error(), http.StatusForbidden)
		return
	}

	limit, offset := parsePagination(r)
	slog.Debug("OrderController.GetByStoreID calling service",
		slog.String("request_id", requestID),
		slog.String("store_id", storeID),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)
	orders, err := c.service.GetByStoreID(r.Context(), storeID, limit, offset)
	if err != nil {
		slog.Error("OrderController.GetByStoreID failed",
			slog.String("request_id", requestID),
			slog.String("store_id", storeID),
			slog.Any("error", err),
		)
		c.handleError(w, err)
		return
	}

	response := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		response[i] = toOrderResponse(o)
	}

	slog.Info("OrderController.GetByStoreID completed",
		slog.String("request_id", requestID),
		slog.String("store_id", storeID),
		slog.Int("count", len(orders)),
	)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response) //nolint:errcheck
}

func (c *OrderController) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	requestID := apiInfra.GetRequestID(r)
	slog.Debug("OrderController.UpdateStatus started",
		slog.String("request_id", requestID),
	)

	id := r.PathValue("id")
	if id == "" {
		slog.Warn("OrderController.UpdateStatus failed: missing order id",
			slog.String("request_id", requestID),
		)
		http.Error(w, "missing order id", http.StatusBadRequest)
		return
	}

	claims, ok := apiInfra.GetClaims(r.Context())
	if !ok {
		slog.Warn("OrderController.UpdateStatus failed: unauthorized",
			slog.String("request_id", requestID),
		)
		http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Warn("OrderController.UpdateStatus failed: invalid request body",
			slog.String("request_id", requestID),
			slog.Any("error", err),
		)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := apiInfra.Validate(req); err != nil {
		slog.Warn("OrderController.UpdateStatus failed: validation error",
			slog.String("request_id", requestID),
			slog.Any("fields", apiInfra.FieldErrors(err)),
		)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"error":  "invalid_payload",
			"fields": apiInfra.FieldErrors(err),
		})
		return
	}

	newStatus := domain.OrderStatus(req.Status)
	if !newStatus.IsValid() {
		slog.Warn("OrderController.UpdateStatus failed: invalid order status",
			slog.String("request_id", requestID),
			slog.String("status", req.Status),
		)
		http.Error(w, apiDomain.ErrInvalidOrderStatus.Error(), http.StatusBadRequest)
		return
	}

	slog.Debug("OrderController.UpdateStatus calling service",
		slog.String("request_id", requestID),
		slog.String("order_id", id),
		slog.String("new_status", req.Status),
		slog.String("account_id", claims.AccountID),
		slog.String("account_type", string(claims.AccountType)),
	)
	order, err := c.service.UpdateStatus(r.Context(), application.UpdateStatusInput{
		OrderID:     id,
		NewStatus:   newStatus,
		AccountType: string(claims.AccountType),
		AccountID:   claims.AccountID,
	})
	if err != nil {
		slog.Error("OrderController.UpdateStatus failed",
			slog.String("request_id", requestID),
			slog.String("order_id", id),
			slog.Any("error", err),
		)
		c.handleError(w, err)
		return
	}

	slog.Info("OrderController.UpdateStatus completed",
		slog.String("request_id", requestID),
		slog.String("order_id", order.ID()),
		slog.String("new_status", string(order.Status())),
	)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toOrderResponse(order)) //nolint:errcheck
}

func (c *OrderController) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, apiDomain.ErrOrderNotFound),
		errors.Is(err, apiDomain.ErrProductNotFound):
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
		slog.Error("unhandled order error", "error", err)
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
