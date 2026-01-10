package application

import (
	"context"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

type IDGenerator interface {
	GenerateStoreID() domain.ID
}

type Service struct {
	repo  domain.Repository
	idGen IDGenerator
}

func NewService(repo domain.Repository, idGen IDGenerator) *Service {
	return &Service{
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
	ID          domain.ID
	Name        string
	Description string
	Address     string
	PhoneNumber string
}

func (s *Service) Create(ctx context.Context, input CreateStoreInput) (*domain.Store, error) {
	id := s.idGen.GenerateStoreID()
	store, err := domain.NewStore(id, input.Name, input.Description, input.Address, input.PhoneNumber)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, store); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Service) GetByID(ctx context.Context, id domain.ID) (*domain.Store, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) GetAll(ctx context.Context, limit, offset int) ([]*domain.Store, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.FindAll(ctx, limit, offset)
}

func (s *Service) Update(ctx context.Context, input UpdateStoreInput) (*domain.Store, error) {
	store, err := s.repo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if err := store.Update(input.Name, input.Description, input.Address, input.PhoneNumber); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, store); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Service) Delete(ctx context.Context, id domain.ID) error {
	return s.repo.Delete(ctx, id)
}
