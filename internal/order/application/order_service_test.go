package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/order/application"
	orderDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/order/domain"
	productDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/product/domain"
)

// --- Mocks ---

type mockOrderRepo struct {
	SaveFn          func(ctx context.Context, order *orderDomain.Order) error
	FindByIDFn      func(ctx context.Context, id string) (*orderDomain.Order, error)
	FindByUserIDFn  func(ctx context.Context, userID string, limit, offset int) ([]*orderDomain.Order, error)
	FindByStoreIDFn func(ctx context.Context, storeID string, limit, offset int) ([]*orderDomain.Order, error)
	UpdateStatusFn  func(ctx context.Context, id string, status orderDomain.OrderStatus) error
}

func (m *mockOrderRepo) Save(ctx context.Context, order *orderDomain.Order) error {
	return m.SaveFn(ctx, order)
}

func (m *mockOrderRepo) FindByID(ctx context.Context, id string) (*orderDomain.Order, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *mockOrderRepo) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*orderDomain.Order, error) {
	return m.FindByUserIDFn(ctx, userID, limit, offset)
}

func (m *mockOrderRepo) FindByStoreID(ctx context.Context, storeID string, limit, offset int) ([]*orderDomain.Order, error) {
	return m.FindByStoreIDFn(ctx, storeID, limit, offset)
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, id string, status orderDomain.OrderStatus) error {
	return m.UpdateStatusFn(ctx, id, status)
}

type mockProductRepo struct {
	FindByIDFn      func(ctx context.Context, id string) (*productDomain.Product, error)
	FindAllFn       func(ctx context.Context, limit, offset int, category *string) ([]*productDomain.Product, error)
	FindByStoreIDFn func(ctx context.Context, storeID string, limit, offset int) ([]*productDomain.Product, error)
	SaveFn          func(ctx context.Context, product *productDomain.Product) error
	UpdateFn        func(ctx context.Context, product *productDomain.Product) error
	DeleteFn        func(ctx context.Context, id string) error
}

func (m *mockProductRepo) FindByID(ctx context.Context, id string) (*productDomain.Product, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *mockProductRepo) FindAll(ctx context.Context, limit, offset int, category *string) ([]*productDomain.Product, error) {
	return m.FindAllFn(ctx, limit, offset, category)
}

func (m *mockProductRepo) FindByStoreID(ctx context.Context, storeID string, limit, offset int) ([]*productDomain.Product, error) {
	return m.FindByStoreIDFn(ctx, storeID, limit, offset)
}

func (m *mockProductRepo) Save(ctx context.Context, product *productDomain.Product) error {
	return m.SaveFn(ctx, product)
}

func (m *mockProductRepo) Update(ctx context.Context, product *productDomain.Product) error {
	return m.UpdateFn(ctx, product)
}

func (m *mockProductRepo) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

type seqIDGen struct {
	ids []string
	i   int
}

func (s *seqIDGen) Generate() string {
	id := s.ids[s.i%len(s.ids)]
	s.i++
	return id
}

// --- Helpers ---

func mustProduct(id, storeID string, stock int, price float64) *productDomain.Product {
	p, err := productDomain.NewProduct(id, storeID, "Name", "Desc", "Cat", stock, price, nil)
	if err != nil {
		panic(err)
	}
	return p
}

func pendingOrderWith(id, userID, storeID string, items []*orderDomain.OrderItem, total float64) *orderDomain.Order {
	return orderDomain.NewOrderFromDB(id, userID, storeID, orderDomain.StatusPending, items, total, time.Now(), time.Now())
}

// --- CreateOrder ---

func TestCreateOrder_HappyPath(t *testing.T) {
	product := mustProduct("prod-1", "store-1", 5, 10.0)
	orderRepo := &mockOrderRepo{
		SaveFn: func(_ context.Context, _ *orderDomain.Order) error { return nil },
	}
	productRepo := &mockProductRepo{
		FindByIDFn: func(_ context.Context, _ string) (*productDomain.Product, error) {
			return product, nil
		},
	}
	svc := application.NewOrderService(orderRepo, productRepo, &seqIDGen{ids: []string{"order-1", "item-1", "item-2"}})

	order, err := svc.CreateOrder(context.Background(), application.CreateOrderInput{
		UserID:  "user-1",
		StoreID: "store-1",
		Items:   []application.OrderItemInput{{ProductID: "prod-1", Quantity: 2}},
	})

	require.NoError(t, err)
	assert.Equal(t, "user-1", order.UserID())
	assert.Equal(t, "store-1", order.StoreID())
	assert.Equal(t, orderDomain.StatusPending, order.Status())
	assert.Len(t, order.Items(), 1)
}

func TestCreateOrder_EmptyItems(t *testing.T) {
	svc := application.NewOrderService(&mockOrderRepo{}, &mockProductRepo{}, &seqIDGen{ids: []string{"id"}})
	_, err := svc.CreateOrder(context.Background(), application.CreateOrderInput{
		UserID:  "user-1",
		StoreID: "store-1",
		Items:   []application.OrderItemInput{},
	})
	assert.ErrorIs(t, err, apiDomain.ErrEmptyOrderItems)
}

func TestCreateOrder_InvalidQuantity(t *testing.T) {
	svc := application.NewOrderService(&mockOrderRepo{}, &mockProductRepo{}, &seqIDGen{ids: []string{"id"}})
	_, err := svc.CreateOrder(context.Background(), application.CreateOrderInput{
		UserID:  "user-1",
		StoreID: "store-1",
		Items:   []application.OrderItemInput{{ProductID: "prod-1", Quantity: 0}},
	})
	assert.ErrorIs(t, err, apiDomain.ErrInvalidQuantity)
}

func TestCreateOrder_ProductNotFound(t *testing.T) {
	productRepo := &mockProductRepo{
		FindByIDFn: func(_ context.Context, _ string) (*productDomain.Product, error) {
			return nil, apiDomain.ErrProductNotFound
		},
	}
	svc := application.NewOrderService(&mockOrderRepo{}, productRepo, &seqIDGen{ids: []string{"id"}})
	_, err := svc.CreateOrder(context.Background(), application.CreateOrderInput{
		UserID:  "user-1",
		StoreID: "store-1",
		Items:   []application.OrderItemInput{{ProductID: "prod-1", Quantity: 1}},
	})
	assert.ErrorIs(t, err, apiDomain.ErrProductNotFound)
}

func TestCreateOrder_ItemsFromMultipleStores(t *testing.T) {
	productRepo := &mockProductRepo{
		FindByIDFn: func(_ context.Context, _ string) (*productDomain.Product, error) {
			return mustProduct("prod-1", "other-store", 5, 10.0), nil
		},
	}
	svc := application.NewOrderService(&mockOrderRepo{}, productRepo, &seqIDGen{ids: []string{"id"}})
	_, err := svc.CreateOrder(context.Background(), application.CreateOrderInput{
		UserID:  "user-1",
		StoreID: "store-1",
		Items:   []application.OrderItemInput{{ProductID: "prod-1", Quantity: 1}},
	})
	assert.ErrorIs(t, err, apiDomain.ErrItemsFromMultipleStores)
}

func TestCreateOrder_InsufficientStock(t *testing.T) {
	product := mustProduct("prod-1", "store-1", 1, 10.0)
	productRepo := &mockProductRepo{
		FindByIDFn: func(_ context.Context, _ string) (*productDomain.Product, error) {
			return product, nil
		},
	}
	svc := application.NewOrderService(&mockOrderRepo{}, productRepo, &seqIDGen{ids: []string{"id"}})
	_, err := svc.CreateOrder(context.Background(), application.CreateOrderInput{
		UserID:  "user-1",
		StoreID: "store-1",
		Items:   []application.OrderItemInput{{ProductID: "prod-1", Quantity: 5}},
	})
	assert.ErrorIs(t, err, apiDomain.ErrInsufficientStock)
}

func TestCreateOrder_PriceCalculation(t *testing.T) {
	product := mustProduct("prod-1", "store-1", 10, 15.0)
	orderRepo := &mockOrderRepo{
		SaveFn: func(_ context.Context, _ *orderDomain.Order) error { return nil },
	}
	productRepo := &mockProductRepo{
		FindByIDFn: func(_ context.Context, _ string) (*productDomain.Product, error) {
			return product, nil
		},
	}
	svc := application.NewOrderService(orderRepo, productRepo, &seqIDGen{ids: []string{"order-1", "item-1"}})

	order, err := svc.CreateOrder(context.Background(), application.CreateOrderInput{
		UserID:  "user-1",
		StoreID: "store-1",
		Items:   []application.OrderItemInput{{ProductID: "prod-1", Quantity: 3}},
	})

	require.NoError(t, err)
	assert.Equal(t, 45.0, order.TotalPrice())
}

// --- UpdateStatus ---

func TestUpdateStatus_Confirm_StoreOwner(t *testing.T) {
	item, _ := orderDomain.NewOrderItem("item-1", "order-1", "prod-1", 2, 10.0)
	existing := pendingOrderWith("order-1", "user-1", "store-1", []*orderDomain.OrderItem{item}, 20.0)

	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return existing, nil
		},
		UpdateStatusFn: func(_ context.Context, _ string, _ orderDomain.OrderStatus) error { return nil },
	}
	svc := application.NewOrderService(orderRepo, &mockProductRepo{}, &seqIDGen{ids: []string{"id"}})

	order, err := svc.UpdateStatus(context.Background(), application.UpdateStatusInput{
		OrderID:     "order-1",
		NewStatus:   orderDomain.StatusConfirmed,
		AccountType: "store",
		AccountID:   "store-1",
	})

	require.NoError(t, err)
	assert.Equal(t, orderDomain.StatusConfirmed, order.Status())
}

func TestUpdateStatus_Cancel_RestoresStock(t *testing.T) {
	item, _ := orderDomain.NewOrderItem("item-1", "order-1", "prod-1", 3, 10.0)
	existing := pendingOrderWith("order-1", "user-1", "store-1", []*orderDomain.OrderItem{item}, 30.0)
	product := mustProduct("prod-1", "store-1", 2, 10.0)

	var updatedStock int
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return existing, nil
		},
		UpdateStatusFn: func(_ context.Context, _ string, _ orderDomain.OrderStatus) error { return nil },
	}
	productRepo := &mockProductRepo{
		FindByIDFn: func(_ context.Context, _ string) (*productDomain.Product, error) {
			return product, nil
		},
		UpdateFn: func(_ context.Context, p *productDomain.Product) error {
			updatedStock = p.Stock()
			return nil
		},
	}
	svc := application.NewOrderService(orderRepo, productRepo, &seqIDGen{ids: []string{"id"}})

	_, err := svc.UpdateStatus(context.Background(), application.UpdateStatusInput{
		OrderID:     "order-1",
		NewStatus:   orderDomain.StatusCancelled,
		AccountType: "user",
		AccountID:   "user-1",
	})

	require.NoError(t, err)
	assert.Equal(t, 5, updatedStock) // 2 original + 3 restored
}

func TestUpdateStatus_Cancel_MultipleItems(t *testing.T) {
	item1, _ := orderDomain.NewOrderItem("item-1", "order-1", "prod-1", 2, 10.0)
	item2, _ := orderDomain.NewOrderItem("item-2", "order-1", "prod-2", 1, 20.0)
	existing := pendingOrderWith("order-1", "user-1", "store-1", []*orderDomain.OrderItem{item1, item2}, 40.0)

	updateCalls := 0
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return existing, nil
		},
		UpdateStatusFn: func(_ context.Context, _ string, _ orderDomain.OrderStatus) error { return nil },
	}
	productRepo := &mockProductRepo{
		FindByIDFn: func(_ context.Context, id string) (*productDomain.Product, error) {
			if id == "prod-1" {
				return mustProduct("prod-1", "store-1", 5, 10.0), nil
			}
			return mustProduct("prod-2", "store-1", 3, 20.0), nil
		},
		UpdateFn: func(_ context.Context, _ *productDomain.Product) error {
			updateCalls++
			return nil
		},
	}
	svc := application.NewOrderService(orderRepo, productRepo, &seqIDGen{ids: []string{"id"}})

	_, err := svc.UpdateStatus(context.Background(), application.UpdateStatusInput{
		OrderID:     "order-1",
		NewStatus:   orderDomain.StatusCancelled,
		AccountType: "user",
		AccountID:   "user-1",
	})

	require.NoError(t, err)
	assert.Equal(t, 2, updateCalls)
}

func TestUpdateStatus_ForbiddenTransition(t *testing.T) {
	item, _ := orderDomain.NewOrderItem("item-1", "order-1", "prod-1", 1, 10.0)
	existing := pendingOrderWith("order-1", "user-1", "store-1", []*orderDomain.OrderItem{item}, 10.0)

	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return existing, nil
		},
	}
	svc := application.NewOrderService(orderRepo, &mockProductRepo{}, &seqIDGen{ids: []string{"id"}})

	_, err := svc.UpdateStatus(context.Background(), application.UpdateStatusInput{
		OrderID:     "order-1",
		NewStatus:   orderDomain.StatusShipped, // Pending -> Shipped is invalid
		AccountType: "store",
		AccountID:   "store-1",
	})
	assert.ErrorIs(t, err, apiDomain.ErrForbiddenTransition)
}

// --- Read operations ---

func TestGetByID_Found(t *testing.T) {
	item, _ := orderDomain.NewOrderItem("item-1", "order-1", "prod-1", 1, 10.0)
	expected := pendingOrderWith("order-1", "user-1", "store-1", []*orderDomain.OrderItem{item}, 10.0)
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return expected, nil
		},
	}
	svc := application.NewOrderService(orderRepo, &mockProductRepo{}, &seqIDGen{ids: []string{"id"}})

	order, err := svc.GetByID(context.Background(), "order-1")
	require.NoError(t, err)
	assert.Equal(t, "order-1", order.ID())
}

func TestGetByID_NotFound(t *testing.T) {
	orderRepo := &mockOrderRepo{
		FindByIDFn: func(_ context.Context, _ string) (*orderDomain.Order, error) {
			return nil, apiDomain.ErrOrderNotFound
		},
	}
	svc := application.NewOrderService(orderRepo, &mockProductRepo{}, &seqIDGen{ids: []string{"id"}})
	_, err := svc.GetByID(context.Background(), "order-1")
	assert.ErrorIs(t, err, apiDomain.ErrOrderNotFound)
}

func TestGetByUserID_DefaultPagination(t *testing.T) {
	var capturedLimit int
	orderRepo := &mockOrderRepo{
		FindByUserIDFn: func(_ context.Context, _ string, limit, _ int) ([]*orderDomain.Order, error) {
			capturedLimit = limit
			return nil, nil
		},
	}
	svc := application.NewOrderService(orderRepo, &mockProductRepo{}, &seqIDGen{ids: []string{"id"}})

	_, err := svc.GetByUserID(context.Background(), "user-1", -1, 0)
	require.NoError(t, err)
	assert.Equal(t, 10, capturedLimit)
}

func TestGetByStoreID_DefaultPagination(t *testing.T) {
	var capturedLimit int
	orderRepo := &mockOrderRepo{
		FindByStoreIDFn: func(_ context.Context, _ string, limit, _ int) ([]*orderDomain.Order, error) {
			capturedLimit = limit
			return nil, nil
		},
	}
	svc := application.NewOrderService(orderRepo, &mockProductRepo{}, &seqIDGen{ids: []string{"id"}})

	_, err := svc.GetByStoreID(context.Background(), "store-1", 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 10, capturedLimit)
}

// ensure unused import doesn't break compilation
var _ = errors.New
