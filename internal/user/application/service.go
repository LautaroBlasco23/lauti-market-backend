package application

import (
	"context"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/domain"
)

type Service struct {
	repo  domain.Repository
	idGen IDGenerator
}

type IDGenerator interface {
	Generate() domain.ID
}

func NewService(repo domain.Repository, idGen IDGenerator) *Service {
	return &Service{repo: repo, idGen: idGen}
}

type CreateInput struct {
	FirstName string
	LastName  string
}

type Output struct {
	ID        domain.ID
	FirstName string
	LastName  string
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Output, error) {
	id := s.idGen.Generate()
	u, err := domain.NewUser(id, input.FirstName, input.LastName)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, u); err != nil {
		return nil, err
	}
	return &Output{
		ID:        u.ID(),
		FirstName: u.FirstName(),
		LastName:  u.LastName(),
	}, nil
}

func (s *Service) GetByID(ctx context.Context, id domain.ID) (*Output, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &Output{
		ID:        u.ID(),
		FirstName: u.FirstName(),
		LastName:  u.LastName(),
	}, nil
}

type UpdateInput struct {
	ID        domain.ID
	FirstName string
	LastName  string
}

func (s *Service) Update(ctx context.Context, input UpdateInput) (*Output, error) {
	u, err := s.repo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if err := u.UpdateName(input.FirstName, input.LastName); err != nil {
		return nil, err
	}
	if err := s.repo.Save(ctx, u); err != nil {
		return nil, err
	}
	return &Output{
		ID:        u.ID(),
		FirstName: u.FirstName(),
		LastName:  u.LastName(),
	}, nil
}

func (s *Service) Delete(ctx context.Context, id domain.ID) error {
	return s.repo.Delete(ctx, id)
}
