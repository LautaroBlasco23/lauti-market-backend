package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

// --- Mocks ---

type mockStoreRepo struct {
	SaveFn     func(ctx context.Context, store *domain.Store) error
	FindByIDFn func(ctx context.Context, id string) (*domain.Store, error)
	FindAllFn  func(ctx context.Context, limit, offset int) ([]*domain.Store, error)
	UpdateFn   func(ctx context.Context, store *domain.Store) error
	DeleteFn   func(ctx context.Context, id string) error
}

func (m *mockStoreRepo) Save(ctx context.Context, store *domain.Store) error {
	return m.SaveFn(ctx, store)
}

func (m *mockStoreRepo) FindByID(ctx context.Context, id string) (*domain.Store, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *mockStoreRepo) FindAll(ctx context.Context, limit, offset int) ([]*domain.Store, error) {
	return m.FindAllFn(ctx, limit, offset)
}

func (m *mockStoreRepo) Update(ctx context.Context, store *domain.Store) error {
	return m.UpdateFn(ctx, store)
}

func (m *mockStoreRepo) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

type mockIDGen struct{ id string }

func (m *mockIDGen) Generate() string { return m.id }

// --- Create ---

func TestCreate_HappyPath(t *testing.T) {
	repo := &mockStoreRepo{
		SaveFn: func(_ context.Context, _ *domain.Store) error { return nil },
	}
	svc := application.NewService(repo, &mockIDGen{"id-1"})

	store, err := svc.Create(context.Background(), application.CreateStoreInput{
		Name:        "My Store",
		Description: "Best store",
		Address:     "123 Main St",
		PhoneNumber: "555-0100",
	})

	require.NoError(t, err)
	assert.Equal(t, "My Store", store.Name())
	assert.Equal(t, "id-1", store.ID())
}

func TestCreate_InvalidInput(t *testing.T) {
	repo := &mockStoreRepo{}
	svc := application.NewService(repo, &mockIDGen{"id-1"})

	_, err := svc.Create(context.Background(), application.CreateStoreInput{
		Name:        "",
		Description: "Best store",
		Address:     "123 Main St",
		PhoneNumber: "555-0100",
	})
	assert.ErrorIs(t, err, domain.ErrInvalidName)
}

// --- GetByID ---

func TestGetByID_Found(t *testing.T) {
	s, _ := domain.NewStore("id-1", "My Store", "Best store", "123 Main St", "555-0100")
	repo := &mockStoreRepo{
		FindByIDFn: func(_ context.Context, _ string) (*domain.Store, error) {
			return s, nil
		},
	}
	svc := application.NewService(repo, &mockIDGen{"id-1"})

	store, err := svc.GetByID(context.Background(), "id-1")
	require.NoError(t, err)
	assert.Equal(t, "id-1", store.ID())
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mockStoreRepo{
		FindByIDFn: func(_ context.Context, _ string) (*domain.Store, error) {
			return nil, domain.ErrStoreNotFound
		},
	}
	svc := application.NewService(repo, &mockIDGen{"id-1"})

	_, err := svc.GetByID(context.Background(), "id-1")
	assert.ErrorIs(t, err, domain.ErrStoreNotFound)
}

// --- GetAll ---

func TestGetAll_DefaultPagination(t *testing.T) {
	var capturedLimit, capturedOffset int
	repo := &mockStoreRepo{
		FindAllFn: func(_ context.Context, limit, offset int) ([]*domain.Store, error) {
			capturedLimit = limit
			capturedOffset = offset
			return nil, nil
		},
	}
	svc := application.NewService(repo, &mockIDGen{"id-1"})

	_, err := svc.GetAll(context.Background(), -1, -1)
	require.NoError(t, err)
	assert.Equal(t, 10, capturedLimit)
	assert.Equal(t, 0, capturedOffset)
}

// --- Update ---

func TestUpdate_HappyPath(t *testing.T) {
	s, _ := domain.NewStore("id-1", "My Store", "Best store", "123 Main St", "555-0100")
	repo := &mockStoreRepo{
		FindByIDFn: func(_ context.Context, _ string) (*domain.Store, error) {
			return s, nil
		},
		UpdateFn: func(_ context.Context, _ *domain.Store) error { return nil },
	}
	svc := application.NewService(repo, &mockIDGen{"id-1"})

	store, err := svc.Update(context.Background(), &application.UpdateStoreInput{
		ID:          "id-1",
		Name:        "Updated Store",
		Description: "Updated desc",
		Address:     "456 Oak Ave",
		PhoneNumber: "555-9999",
	})

	require.NoError(t, err)
	assert.Equal(t, "Updated Store", store.Name())
}

// --- Delete ---

func TestDelete_HappyPath(t *testing.T) {
	repo := &mockStoreRepo{
		DeleteFn: func(_ context.Context, _ string) error { return nil },
	}
	svc := application.NewService(repo, &mockIDGen{"id-1"})

	err := svc.Delete(context.Background(), "id-1")
	assert.NoError(t, err)
}
