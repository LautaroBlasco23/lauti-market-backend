package application

import (
	"context"
	"fmt"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	imageDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/image/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/product/domain"
	storeDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

type ProductService struct {
	repo        domain.Repository
	storeRepo   storeDomain.Repository
	idGen       apiDomain.IDGenerator
	imageClient imageDomain.ImageClient
}

func NewService(repo domain.Repository, storeRepo storeDomain.Repository, idGen apiDomain.IDGenerator, imageClient imageDomain.ImageClient) *ProductService {
	return &ProductService{
		repo:        repo,
		storeRepo:   storeRepo,
		idGen:       idGen,
		imageClient: imageClient,
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

type UploadProductImageInput struct {
	ProductID   string
	StoreID     string
	Filename    string
	ContentType string
	Data        []byte
}

func (s *ProductService) UploadImage(ctx context.Context, input UploadProductImageInput) (*domain.Product, error) {
	product, err := s.repo.FindByID(ctx, input.ProductID)
	if err != nil {
		return nil, err
	}
	result, err := s.imageClient.UploadImage(ctx, imageDomain.UploadImageInput{
		UserID: input.ProductID, Filename: input.Filename,
		ContentType: input.ContentType, Data: input.Data,
	})
	if err != nil {
		return nil, fmt.Errorf("uploading image: %w", err)
	}
	if err := product.Update(product.Name(), product.Description(), product.Stock(), product.Price(), &result.URL); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}
