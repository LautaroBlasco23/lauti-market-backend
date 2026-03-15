package application

import (
	"context"

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
	if len(input.Items) == 0 {
		return nil, apiDomain.ErrEmptyOrderItems
	}

	var totalPrice float64
	orderItems := make([]*domain.OrderItem, 0, len(input.Items))
	orderID := s.idGen.Generate()

	for _, itemInput := range input.Items {
		if itemInput.Quantity <= 0 {
			return nil, apiDomain.ErrInvalidQuantity
		}

		product, err := s.productRepo.FindByID(ctx, itemInput.ProductID)
		if err != nil {
			return nil, err
		}

		if product.StoreID() != input.StoreID {
			return nil, apiDomain.ErrItemsFromMultipleStores
		}

		if product.Stock() < itemInput.Quantity {
			return nil, apiDomain.ErrInsufficientStock
		}

		item, err := domain.NewOrderItem(s.idGen.Generate(), orderID, itemInput.ProductID, itemInput.Quantity, product.Price())
		if err != nil {
			return nil, err
		}

		orderItems = append(orderItems, item)
		totalPrice += product.Price() * float64(itemInput.Quantity)
	}

	order, err := domain.NewOrder(orderID, input.UserID, input.StoreID, orderItems, totalPrice)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) UpdateStatus(ctx context.Context, input UpdateStatusInput) (*domain.Order, error) {
	order, err := s.repo.FindByID(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}

	if err := order.TransitionStatus(input.NewStatus, input.AccountType, input.AccountID); err != nil {
		return nil, err
	}

	if input.NewStatus == domain.StatusCancelled {
		for _, item := range order.Items() {
			product, err := s.productRepo.FindByID(ctx, item.ProductID())
			if err != nil {
				return nil, err
			}
			err = product.Update(
				product.Name(),
				product.Description(),
				product.Stock()+item.Quantity(),
				product.Price(),
				product.ImageURL(),
			)
			if err != nil {
				return nil, err
			}
			if err := s.productRepo.Update(ctx, product); err != nil {
				return nil, err
			}
		}
	}

	if err := s.repo.UpdateStatus(ctx, order.ID(), input.NewStatus); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *OrderService) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Order, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.FindByUserID(ctx, userID, limit, offset)
}

func (s *OrderService) GetByStoreID(ctx context.Context, storeID string, limit, offset int) ([]*domain.Order, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.FindByStoreID(ctx, storeID, limit, offset)
}
