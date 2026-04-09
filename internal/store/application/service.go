package application

import (
	"context"
	"log/slog"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

type StoreService struct {
	repo  domain.Repository
	idGen apiDomain.IDGenerator
}

func NewService(repo domain.Repository, idGen apiDomain.IDGenerator) *StoreService {
	return &StoreService{
		repo:  repo,
		idGen: idGen,
	}
}

type CreateStoreInput struct {
	Name        string
	Description string
	Address     string
	PhoneNumber string
}

type UpdateStoreInput struct {
	ID          string
	Name        string
	Description string
	Address     string
	PhoneNumber string
}

func (s *StoreService) Create(ctx context.Context, input CreateStoreInput) (*domain.Store, error) {
	slog.Debug("StoreService.Create started",
		slog.String("name", input.Name),
	)

	id := s.idGen.Generate()
	slog.Debug("StoreService.Create creating store entity",
		slog.String("id", id),
	)
	store, err := domain.NewStore(id, input.Name, input.Description, input.Address, input.PhoneNumber)
	if err != nil {
		slog.Error("StoreService.Create failed",
			slog.String("operation", "new_store_entity"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Debug("StoreService.Create saving store to repository")
	if err := s.repo.Save(ctx, store); err != nil {
		slog.Error("StoreService.Create failed",
			slog.String("operation", "save_store"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("StoreService.Create completed",
		slog.String("id", store.ID()),
	)
	return store, nil
}

func (s *StoreService) GetByID(ctx context.Context, id string) (*domain.Store, error) {
	slog.Debug("StoreService.GetByID started",
		slog.String("id", id),
	)

	slog.Debug("StoreService.GetByID finding store by ID")
	store, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.Error("StoreService.GetByID failed",
			slog.String("operation", "find_by_id"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("StoreService.GetByID completed",
		slog.String("id", store.ID()),
	)
	return store, nil
}

func (s *StoreService) GetAll(ctx context.Context, limit, offset int) ([]*domain.Store, error) {
	slog.Debug("StoreService.GetAll started",
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	slog.Debug("StoreService.GetAll finding all stores")
	stores, err := s.repo.FindAll(ctx, limit, offset)
	if err != nil {
		slog.Error("StoreService.GetAll failed",
			slog.String("operation", "find_all"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("StoreService.GetAll completed",
		slog.Int("count", len(stores)),
	)
	return stores, nil
}

func (s *StoreService) Update(ctx context.Context, input *UpdateStoreInput) (*domain.Store, error) {
	slog.Debug("StoreService.Update started",
		slog.String("id", input.ID),
		slog.String("name", input.Name),
	)

	slog.Debug("StoreService.Update finding store by ID")
	store, err := s.repo.FindByID(ctx, input.ID)
	if err != nil {
		slog.Error("StoreService.Update failed",
			slog.String("operation", "find_by_id"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Debug("StoreService.Update updating store fields")
	if err := store.Update(input.Name, input.Description, input.Address, input.PhoneNumber); err != nil {
		slog.Error("StoreService.Update failed",
			slog.String("operation", "update_fields"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Debug("StoreService.Update saving store to repository")
	if err := s.repo.Update(ctx, store); err != nil {
		slog.Error("StoreService.Update failed",
			slog.String("operation", "update_store"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("StoreService.Update completed",
		slog.String("id", store.ID()),
	)
	return store, nil
}

func (s *StoreService) Delete(ctx context.Context, id string) error {
	slog.Debug("StoreService.Delete started",
		slog.String("id", id),
	)

	slog.Debug("StoreService.Delete deleting store from repository")
	if err := s.repo.Delete(ctx, id); err != nil {
		slog.Error("StoreService.Delete failed",
			slog.String("operation", "delete_store"),
			slog.Any("error", err),
		)
		return err
	}

	slog.Info("StoreService.Delete completed",
		slog.String("id", id),
	)
	return nil
}
