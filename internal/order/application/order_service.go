package application

import (
	"context"
	"log/slog"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/order/domain"
	productDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/product/domain"
)

type OrderItemInput struct {
	ProductID string
	Quantity  int
}

type CreateOrderInput struct {
	UserID  string
	StoreID string
	Items   []OrderItemInput
}

type UpdateStatusInput struct {
	OrderID     string
	NewStatus   domain.OrderStatus
	AccountType string
	AccountID   string
}

type OrderService struct {
	repo        domain.Repository
	productRepo productDomain.Repository
	idGen       apiDomain.IDGenerator
}

func NewOrderService(repo domain.Repository, productRepo productDomain.Repository, idGen apiDomain.IDGenerator) *OrderService {
	return &OrderService{
		repo:        repo,
		productRepo: productRepo,
		idGen:       idGen,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, input CreateOrderInput) (*domain.Order, error) {
	slog.Debug("OrderService.CreateOrder started",
		slog.String("user_id", input.UserID),
		slog.String("store_id", input.StoreID),
		slog.Int("item_count", len(input.Items)),
	)

	if len(input.Items) == 0 {
		slog.Error("OrderService.CreateOrder failed",
			slog.String("operation", "validate_items"),
			slog.String("error", "empty order items"),
		)
		return nil, apiDomain.ErrEmptyOrderItems
	}

	var totalPrice float64
	orderItems := make([]*domain.OrderItem, 0, len(input.Items))
	orderID := s.idGen.Generate()
	slog.Debug("OrderService.CreateOrder generated order ID", slog.String("order_id", orderID))

	for i, itemInput := range input.Items {
		if itemInput.Quantity <= 0 {
			slog.Error("OrderService.CreateOrder failed",
				slog.String("operation", "validate_quantity"),
				slog.Int("item_index", i),
				slog.Int("quantity", itemInput.Quantity),
				slog.String("error", "invalid quantity"),
			)
			return nil, apiDomain.ErrInvalidQuantity
		}

		slog.Debug("OrderService.CreateOrder fetching product",
			slog.String("product_id", itemInput.ProductID),
		)
		product, err := s.productRepo.FindByID(ctx, itemInput.ProductID)
		if err != nil {
			slog.Error("OrderService.CreateOrder failed",
				slog.String("operation", "find_product_by_id"),
				slog.String("product_id", itemInput.ProductID),
				slog.Any("error", err),
			)
			return nil, err
		}

		if product.StoreID() != input.StoreID {
			slog.Error("OrderService.CreateOrder failed",
				slog.String("operation", "validate_store_match"),
				slog.String("expected_store_id", input.StoreID),
				slog.String("actual_store_id", product.StoreID()),
				slog.String("error", "items from multiple stores"),
			)
			return nil, apiDomain.ErrItemsFromMultipleStores
		}

		if product.Stock() < itemInput.Quantity {
			slog.Error("OrderService.CreateOrder failed",
				slog.String("operation", "check_stock"),
				slog.String("product_id", itemInput.ProductID),
				slog.Int("requested_quantity", itemInput.Quantity),
				slog.Int("available_stock", product.Stock()),
				slog.String("error", "insufficient stock"),
			)
			return nil, apiDomain.ErrInsufficientStock
		}

		item, err := domain.NewOrderItem(s.idGen.Generate(), orderID, itemInput.ProductID, itemInput.Quantity, product.Price())
		if err != nil {
			slog.Error("OrderService.CreateOrder failed",
				slog.String("operation", "create_order_item"),
				slog.String("product_id", itemInput.ProductID),
				slog.Any("error", err),
			)
			return nil, err
		}

		orderItems = append(orderItems, item)
		totalPrice += product.Price() * float64(itemInput.Quantity)
	}

	order, err := domain.NewOrder(orderID, input.UserID, input.StoreID, orderItems, totalPrice)
	if err != nil {
		slog.Error("OrderService.CreateOrder failed",
			slog.String("operation", "create_order_domain"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Debug("OrderService.CreateOrder saving order to repository", slog.String("order_id", orderID))
	if err := s.repo.Save(ctx, order); err != nil {
		slog.Error("OrderService.CreateOrder failed",
			slog.String("operation", "save_order"),
			slog.String("order_id", orderID),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("OrderService.CreateOrder completed",
		slog.String("order_id", order.ID()),
		slog.String("user_id", order.UserID()),
		slog.Float64("total_price", order.TotalPrice()),
	)
	return order, nil
}

func (s *OrderService) UpdateStatus(ctx context.Context, input UpdateStatusInput) (*domain.Order, error) {
	slog.Debug("OrderService.UpdateStatus started",
		slog.String("order_id", input.OrderID),
		slog.String("new_status", string(input.NewStatus)),
		slog.String("account_type", input.AccountType),
	)

	order, err := s.repo.FindByID(ctx, input.OrderID)
	if err != nil {
		slog.Error("OrderService.UpdateStatus failed",
			slog.String("operation", "find_by_id"),
			slog.String("order_id", input.OrderID),
			slog.Any("error", err),
		)
		return nil, err
	}

	if err := order.TransitionStatus(input.NewStatus, input.AccountType, input.AccountID); err != nil {
		slog.Error("OrderService.UpdateStatus failed",
			slog.String("operation", "transition_status"),
			slog.String("order_id", input.OrderID),
			slog.String("current_status", string(order.Status())),
			slog.String("requested_status", string(input.NewStatus)),
			slog.Any("error", err),
		)
		return nil, err
	}

	if input.NewStatus == domain.StatusCancelled {
		slog.Debug("OrderService.UpdateStatus restoring stock for cancelled order", slog.String("order_id", input.OrderID))
		for _, item := range order.Items() {
			product, err := s.productRepo.FindByID(ctx, item.ProductID())
			if err != nil {
				slog.Error("OrderService.UpdateStatus failed",
					slog.String("operation", "find_product_for_stock_restore"),
					slog.String("product_id", item.ProductID()),
					slog.Any("error", err),
				)
				return nil, err
			}
			err = product.Update(
				product.Name(),
				product.Description(),
				product.Category(),
				product.Stock()+item.Quantity(),
				product.Price(),
				product.ImageURL(),
			)
			if err != nil {
				slog.Error("OrderService.UpdateStatus failed",
					slog.String("operation", "update_product_stock"),
					slog.String("product_id", item.ProductID()),
					slog.Any("error", err),
				)
				return nil, err
			}
			if err := s.productRepo.Update(ctx, product); err != nil {
				slog.Error("OrderService.UpdateStatus failed",
					slog.String("operation", "save_product_stock"),
					slog.String("product_id", item.ProductID()),
					slog.Any("error", err),
				)
				return nil, err
			}
		}
	}

	slog.Debug("OrderService.UpdateStatus saving status to repository",
		slog.String("order_id", order.ID()),
		slog.String("new_status", string(input.NewStatus)),
	)
	if err := s.repo.UpdateStatus(ctx, order.ID(), input.NewStatus); err != nil {
		slog.Error("OrderService.UpdateStatus failed",
			slog.String("operation", "update_status_repository"),
			slog.String("order_id", order.ID()),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("OrderService.UpdateStatus completed",
		slog.String("order_id", order.ID()),
		slog.String("new_status", string(input.NewStatus)),
	)
	return order, nil
}

func (s *OrderService) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	slog.Debug("OrderService.GetByID started", slog.String("order_id", id))

	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.Error("OrderService.GetByID failed",
			slog.String("operation", "find_by_id"),
			slog.String("order_id", id),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("OrderService.GetByID completed", slog.String("order_id", id))
	return order, nil
}

func (s *OrderService) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Order, error) {
	slog.Debug("OrderService.GetByUserID started",
		slog.String("user_id", userID),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	if limit <= 0 {
		limit = 10
	}

	orders, err := s.repo.FindByUserID(ctx, userID, limit, offset)
	if err != nil {
		slog.Error("OrderService.GetByUserID failed",
			slog.String("operation", "find_by_user_id"),
			slog.String("user_id", userID),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("OrderService.GetByUserID completed",
		slog.String("user_id", userID),
		slog.Int("count", len(orders)),
	)
	return orders, nil
}

func (s *OrderService) GetByStoreID(ctx context.Context, storeID string, limit, offset int) ([]*domain.Order, error) {
	slog.Debug("OrderService.GetByStoreID started",
		slog.String("store_id", storeID),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	if limit <= 0 {
		limit = 10
	}

	orders, err := s.repo.FindByStoreID(ctx, storeID, limit, offset)
	if err != nil {
		slog.Error("OrderService.GetByStoreID failed",
			slog.String("operation", "find_by_store_id"),
			slog.String("store_id", storeID),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("OrderService.GetByStoreID completed",
		slog.String("store_id", storeID),
		slog.Int("count", len(orders)),
	)
	return orders, nil
}
