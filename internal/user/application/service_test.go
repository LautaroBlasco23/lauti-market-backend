package application_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/application"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/domain"
)

// --- Mocks ---

type mockUserRepo struct {
	SaveFn     func(ctx context.Context, user *domain.User) error
	UpdateFn   func(ctx context.Context, user *domain.User) error
	FindByIDFn func(ctx context.Context, id string) (*domain.User, error)
	DeleteFn   func(ctx context.Context, id string) error
}

func (m *mockUserRepo) Save(ctx context.Context, user *domain.User) error {
	return m.SaveFn(ctx, user)
}

func (m *mockUserRepo) Update(ctx context.Context, user *domain.User) error {
	return m.UpdateFn(ctx, user)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *mockUserRepo) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

type mockIDGen struct{ id string }

func (m *mockIDGen) Generate() string { return m.id }

// --- Create ---

func TestCreate_HappyPath(t *testing.T) {
	repo := &mockUserRepo{
		SaveFn: func(_ context.Context, _ *domain.User) error { return nil },
	}
	svc := application.NewService(repo, &mockIDGen{"id-1"})

	out, err := svc.Create(context.Background(), application.CreateInput{
		FirstName: "John",
		LastName:  "Doe",
	})

	require.NoError(t, err)
	assert.Equal(t, "id-1", out.ID)
	assert.Equal(t, "John", out.FirstName)
	assert.Equal(t, "Doe", out.LastName)
}

func TestCreate_InvalidInput(t *testing.T) {
	repo := &mockUserRepo{}
	svc := application.NewService(repo, &mockIDGen{"id-1"})

	_, err := svc.Create(context.Background(), application.CreateInput{
		FirstName: "",
		LastName:  "Doe",
	})
	assert.ErrorIs(t, err, domain.ErrInvalidFirstName)
}

// --- GetByID ---

func TestGetByID_Found(t *testing.T) {
	u, _ := domain.NewUser("id-1", "John", "Doe")
	repo := &mockUserRepo{
		FindByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			return u, nil
		},
	}
	svc := application.NewService(repo, &mockIDGen{"id-1"})

	out, err := svc.GetByID(context.Background(), "id-1")
	require.NoError(t, err)
	assert.Equal(t, "id-1", out.ID)
}

func TestGetByID_NotFound(t *testing.T) {
	repo := &mockUserRepo{
		FindByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	svc := application.NewService(repo, &mockIDGen{"id-1"})

	_, err := svc.GetByID(context.Background(), "id-1")
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

// --- Update ---

func TestUpdate_HappyPath(t *testing.T) {
	u, _ := domain.NewUser("id-1", "John", "Doe")
	repo := &mockUserRepo{
		FindByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			return u, nil
		},
		UpdateFn: func(_ context.Context, _ *domain.User) error { return nil },
	}
	svc := application.NewService(repo, &mockIDGen{"id-1"})

	out, err := svc.Update(context.Background(), application.UpdateInput{
		ID:        "id-1",
		FirstName: "Jane",
		LastName:  "Smith",
	})

	require.NoError(t, err)
	assert.Equal(t, "Jane", out.FirstName)
	assert.Equal(t, "Smith", out.LastName)
}

func TestUpdate_NotFound(t *testing.T) {
	repo := &mockUserRepo{
		FindByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	svc := application.NewService(repo, &mockIDGen{"id-1"})

	_, err := svc.Update(context.Background(), application.UpdateInput{
		ID:        "id-1",
		FirstName: "Jane",
		LastName:  "Smith",
	})
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

// --- Delete ---

func TestDelete_HappyPath(t *testing.T) {
	repo := &mockUserRepo{
		DeleteFn: func(_ context.Context, _ string) error { return nil },
	}
	svc := application.NewService(repo, &mockIDGen{"id-1"})

	err := svc.Delete(context.Background(), "id-1")
	assert.NoError(t, err)
}

func TestDelete_NotFound(t *testing.T) {
	repo := &mockUserRepo{
		DeleteFn: func(_ context.Context, _ string) error {
			return domain.ErrUserNotFound
		},
	}
	svc := application.NewService(repo, &mockIDGen{"id-1"})

	err := svc.Delete(context.Background(), "id-1")
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}
