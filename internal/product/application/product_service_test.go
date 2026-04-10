package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	imageDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/image/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/product/application"
	productDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/product/domain"
	storeDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

// --- Mocks ---

type mockProductRepo struct {
	SaveFn          func(ctx context.Context, product *productDomain.Product) error
	FindByIDFn      func(ctx context.Context, id string) (*productDomain.Product, error)
	FindAllFn       func(ctx context.Context, limit, offset int, category *string) ([]*productDomain.Product, error)
	FindByStoreIDFn func(ctx context.Context, storeID string, limit, offset int) ([]*productDomain.Product, error)
	UpdateFn        func(ctx context.Context, product *productDomain.Product) error
	DeleteFn        func(ctx context.Context, id string) error
}

func (m *mockProductRepo) Save(ctx context.Context, product *productDomain.Product) error {
	return m.SaveFn(ctx, product)
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

func (m *mockProductRepo) Update(ctx context.Context, product *productDomain.Product) error {
	return m.UpdateFn(ctx, product)
}

func (m *mockProductRepo) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

type mockStoreRepo struct {
	FindByIDFn func(ctx context.Context, id string) (*storeDomain.Store, error)
	SaveFn     func(ctx context.Context, store *storeDomain.Store) error
	FindAllFn  func(ctx context.Context, limit, offset int) ([]*storeDomain.Store, error)
	UpdateFn   func(ctx context.Context, store *storeDomain.Store) error
	DeleteFn   func(ctx context.Context, id string) error
}

func (m *mockStoreRepo) FindByID(ctx context.Context, id string) (*storeDomain.Store, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *mockStoreRepo) Save(ctx context.Context, store *storeDomain.Store) error {
	return m.SaveFn(ctx, store)
}

func (m *mockStoreRepo) FindAll(ctx context.Context, limit, offset int) ([]*storeDomain.Store, error) {
	return m.FindAllFn(ctx, limit, offset)
}

func (m *mockStoreRepo) Update(ctx context.Context, store *storeDomain.Store) error {
	return m.UpdateFn(ctx, store)
}

func (m *mockStoreRepo) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *mockStoreRepo) UpdateMPConnection(ctx context.Context, storeID string, fields storeDomain.MPFields) error {
	return nil
}

type mockImageClient struct {
	UploadImageFn func(ctx context.Context, input imageDomain.UploadImageInput) (*imageDomain.UploadImageResult, error)
}

func (m *mockImageClient) UploadImage(ctx context.Context, input imageDomain.UploadImageInput) (*imageDomain.UploadImageResult, error) {
	return m.UploadImageFn(ctx, input)
}

type mockIDGen struct{ id string }

func (m *mockIDGen) Generate() string { return m.id }

// --- Helpers ---

func existingStore(id string) *storeDomain.Store {
	s, err := storeDomain.NewStore(id, storeDomain.CreateStoreInput{
		Name:        "Store",
		Description: "Desc",
		Address:     "Address",
		PhoneNumber: "555-0000",
	})
	if err != nil {
		panic(err)
	}
	return s
}

func existingStoreWithMP(id string) *storeDomain.Store {
	s := existingStore(id)
	// Connect MP with a far future expiration to ensure token is valid
	s.ConnectMP("mp-user-123", "access-token", "refresh-token", time.Now().Add(24*time.Hour))
	return s
}

func storeFoundRepo(id string) *mockStoreRepo {
	return &mockStoreRepo{
		FindByIDFn: func(_ context.Context, _ string) (*storeDomain.Store, error) {
			return existingStoreWithMP(id), nil
		},
	}
}

func storeNotFoundRepo() *mockStoreRepo {
	return &mockStoreRepo{
		FindByIDFn: func(_ context.Context, _ string) (*storeDomain.Store, error) {
			return nil, storeDomain.ErrStoreNotFound
		},
	}
}

func noImageClient() *mockImageClient {
	return &mockImageClient{}
}

// --- Create ---

func TestCreate_HappyPath(t *testing.T) {
	productRepo := &mockProductRepo{
		SaveFn: func(_ context.Context, _ *productDomain.Product) error { return nil },
	}
	svc := application.NewService(productRepo, storeFoundRepo("store-1"), &mockIDGen{"prod-1"}, noImageClient())

	p, err := svc.Create(context.Background(), &application.CreateProductInput{
		StoreID:     "store-1",
		Name:        "Widget",
		Description: "A widget",
		Category:    "tools",
		Stock:       10,
		Price:       9.99,
	})

	require.NoError(t, err)
	assert.Equal(t, "Widget", p.Name())
	assert.Equal(t, "store-1", p.StoreID())
}

func TestCreate_StoreNotFound(t *testing.T) {
	svc := application.NewService(&mockProductRepo{}, storeNotFoundRepo(), &mockIDGen{"prod-1"}, noImageClient())

	_, err := svc.Create(context.Background(), &application.CreateProductInput{
		StoreID:     "store-1",
		Name:        "Widget",
		Description: "A widget",
		Category:    "tools",
		Stock:       10,
		Price:       9.99,
	})
	assert.ErrorIs(t, err, storeDomain.ErrStoreNotFound)
}

func TestCreate_WithImage(t *testing.T) {
	imageClient := &mockImageClient{
		UploadImageFn: func(_ context.Context, _ imageDomain.UploadImageInput) (*imageDomain.UploadImageResult, error) {
			return &imageDomain.UploadImageResult{URL: "https://cdn.example.com/img.jpg"}, nil
		},
	}
	productRepo := &mockProductRepo{
		SaveFn: func(_ context.Context, _ *productDomain.Product) error { return nil },
	}
	svc := application.NewService(productRepo, storeFoundRepo("store-1"), &mockIDGen{"prod-1"}, imageClient)

	p, err := svc.Create(context.Background(), &application.CreateProductInput{
		StoreID:          "store-1",
		Name:             "Widget",
		Description:      "A widget",
		Category:         "tools",
		Stock:            10,
		Price:            9.99,
		ImageData:        []byte("fakeimage"),
		ImageFilename:    "img.jpg",
		ImageContentType: "image/jpeg",
	})

	require.NoError(t, err)
	require.NotNil(t, p.ImageURL())
	assert.Equal(t, "https://cdn.example.com/img.jpg", *p.ImageURL())
}

func TestCreate_InvalidInput(t *testing.T) {
	svc := application.NewService(&mockProductRepo{}, storeFoundRepo("store-1"), &mockIDGen{"prod-1"}, noImageClient())

	_, err := svc.Create(context.Background(), &application.CreateProductInput{
		StoreID:     "store-1",
		Name:        "", // invalid
		Description: "A widget",
		Category:    "tools",
		Stock:       10,
		Price:       9.99,
	})
	assert.ErrorIs(t, err, apiDomain.ErrInvalidProductName)
}

// --- GetByID ---

func TestGetByID_Found(t *testing.T) {
	existing, _ := productDomain.NewProduct("prod-1", "store-1", "Widget", "A widget", "tools", 10, 9.99, nil)
	productRepo := &mockProductRepo{
		FindByIDFn: func(_ context.Context, _ string) (*productDomain.Product, error) {
			return existing, nil
		},
	}
	svc := application.NewService(productRepo, storeFoundRepo("store-1"), &mockIDGen{"prod-1"}, noImageClient())

	p, err := svc.GetByID(context.Background(), "prod-1")
	require.NoError(t, err)
	assert.Equal(t, "prod-1", p.ID())
}

func TestGetByID_NotFound(t *testing.T) {
	productRepo := &mockProductRepo{
		FindByIDFn: func(_ context.Context, _ string) (*productDomain.Product, error) {
			return nil, apiDomain.ErrProductNotFound
		},
	}
	svc := application.NewService(productRepo, storeFoundRepo("store-1"), &mockIDGen{"prod-1"}, noImageClient())

	_, err := svc.GetByID(context.Background(), "prod-1")
	assert.ErrorIs(t, err, apiDomain.ErrProductNotFound)
}

// --- GetAll ---

func TestGetAll_DefaultPagination(t *testing.T) {
	var capturedLimit, capturedOffset int
	productRepo := &mockProductRepo{
		FindAllFn: func(_ context.Context, limit, offset int, _ *string) ([]*productDomain.Product, error) {
			capturedLimit = limit
			capturedOffset = offset
			return nil, nil
		},
	}
	svc := application.NewService(productRepo, storeFoundRepo("store-1"), &mockIDGen{"prod-1"}, noImageClient())

	_, err := svc.GetAll(context.Background(), application.GetAllProductsInput{Limit: -1, Offset: -1})
	require.NoError(t, err)
	assert.Equal(t, 10, capturedLimit)
	assert.Equal(t, 0, capturedOffset)
}

// --- GetByStoreID ---

func TestGetByStoreID_StoreNotFound(t *testing.T) {
	svc := application.NewService(&mockProductRepo{}, storeNotFoundRepo(), &mockIDGen{"prod-1"}, noImageClient())

	_, err := svc.GetByStoreID(context.Background(), "store-1", 10, 0)
	assert.ErrorIs(t, err, storeDomain.ErrStoreNotFound)
}

// --- Update ---

func TestUpdate_HappyPath(t *testing.T) {
	existing, _ := productDomain.NewProduct("prod-1", "store-1", "Widget", "A widget", "tools", 10, 9.99, nil)
	productRepo := &mockProductRepo{
		FindByIDFn: func(_ context.Context, _ string) (*productDomain.Product, error) {
			return existing, nil
		},
		UpdateFn: func(_ context.Context, _ *productDomain.Product) error { return nil },
	}
	svc := application.NewService(productRepo, storeFoundRepo("store-1"), &mockIDGen{"prod-1"}, noImageClient())

	p, err := svc.Update(context.Background(), &application.UpdateProductInput{
		ID:          "prod-1",
		Name:        "Gadget",
		Description: "A gadget",
		Category:    "electronics",
		Stock:       5,
		Price:       19.99,
	})

	require.NoError(t, err)
	assert.Equal(t, "Gadget", p.Name())
}

// --- Delete ---

func TestDelete_HappyPath(t *testing.T) {
	productRepo := &mockProductRepo{
		DeleteFn: func(_ context.Context, _ string) error { return nil },
	}
	svc := application.NewService(productRepo, storeFoundRepo("store-1"), &mockIDGen{"prod-1"}, noImageClient())

	err := svc.Delete(context.Background(), "prod-1")
	assert.NoError(t, err)
}

// suppress unused import
var _ = errors.New
