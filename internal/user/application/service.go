package application

import (
	"context"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/domain"
)

type UserService struct {
	repo  domain.Repository
	idGen apiDomain.IDGenerator
}

func NewService(repo domain.Repository, idGen apiDomain.IDGenerator) *UserService {
	return &UserService{repo: repo, idGen: idGen}
}

type CreateInput struct {
	FirstName string
	LastName  string
}

type Output struct {
	ID        string
	FirstName string
	LastName  string
}

func (s *UserService) Create(ctx context.Context, input CreateInput) (*Output, error) {
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

func (s *UserService) GetByID(ctx context.Context, id string) (*Output, error) {
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
	ID        string
	FirstName string
	LastName  string
}

func (s *UserService) Update(ctx context.Context, input UpdateInput) (*Output, error) {
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

func (s *UserService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
