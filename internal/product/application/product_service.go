package application

import (
	"context"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/product/domain"
	storeDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

type ProductService struct {
	repo      domain.Repository
	storeRepo storeDomain.Repository
	idGen     apiDomain.IDGenerator
}

func NewService(repo domain.Repository, storeRepo storeDomain.Repository, idGen apiDomain.IDGenerator) *ProductService {
	return &ProductService{
		repo:      repo,
		storeRepo: storeRepo,
		idGen:     idGen,
	}
}

type CreateProductInput struct {
	StoreID     string
	Name        string
	Description string
	Stock       int
	Price       float64
	ImageURL    *string
}

type UpdateProductInput struct {
	ID          string
	Name        string
	Description string
	Stock       int
	Price       float64
	ImageURL    *string
}

func (s *ProductService) Create(ctx context.Context, input CreateProductInput) (*domain.Product, error) {
	if _, err := s.storeRepo.FindByID(ctx, input.StoreID); err != nil {
		return nil, err
	}

	id := s.idGen.Generate()
	product, err := domain.NewProduct(id, input.StoreID, input.Name, input.Description, input.Stock, input.Price, input.ImageURL)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ProductService) GetByStoreID(ctx context.Context, storeID string, limit, offset int) ([]*domain.Product, error) {
	if _, err := s.storeRepo.FindByID(ctx, storeID); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.FindByStoreID(ctx, storeID, limit, offset)
}

func (s *ProductService) Update(ctx context.Context, input UpdateProductInput) (*domain.Product, error) {
	product, err := s.repo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if err := product.Update(input.Name, input.Description, input.Stock, input.Price, input.ImageURL); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
