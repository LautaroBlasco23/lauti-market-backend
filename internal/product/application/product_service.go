package application

import (
	"context"
	"fmt"
	"log/slog"

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
	StoreID          string
	Name             string
	Description      string
	Category         string
	Stock            int
	Price            float64
	ImageData        []byte
	ImageFilename    string
	ImageContentType string
}

type UpdateProductInput struct {
	ID          string
	Name        string
	Description string
	Category    string
	Stock       int
	Price       float64
	ImageURL    *string
}

type GetAllProductsInput struct {
	Limit    int
	Offset   int
	Category *string
}

func (s *ProductService) Create(ctx context.Context, input *CreateProductInput) (*domain.Product, error) {
	slog.Debug("ProductService.Create started",
		slog.String("store_id", input.StoreID),
		slog.String("name", input.Name),
		slog.String("category", input.Category),
	)

	store, err := s.storeRepo.FindByID(ctx, input.StoreID)
	if err != nil {
		slog.Error("ProductService.Create failed",
			slog.String("operation", "find_store_by_id"),
			slog.String("store_id", input.StoreID),
			slog.Any("error", err),
		)
		return nil, err
	}

	if !store.IsMPConnected() {
		slog.Error("ProductService.Create failed",
			slog.String("operation", "check_mp_connected"),
			slog.String("store_id", input.StoreID),
			slog.String("error", "store does not have a connected MercadoPago account"),
		)
		return nil, apiDomain.ErrStoreMPNotConnected
	}

	if !store.IsMPTokenValid() {
		slog.Error("ProductService.Create failed",
			slog.String("operation", "check_mp_token_valid"),
			slog.String("store_id", input.StoreID),
			slog.String("error", "store MercadoPago token is invalid or expired"),
		)
		return nil, apiDomain.ErrStoreMPTokenExpired
	}

	id := s.idGen.Generate()
	slog.Debug("ProductService.Create generated product ID", slog.String("product_id", id))

	var imageURL *string
	if len(input.ImageData) > 0 {
		slog.Debug("ProductService.Create uploading image",
			slog.String("filename", input.ImageFilename),
			slog.Int("data_size", len(input.ImageData)),
		)
		result, uploadErr := s.imageClient.UploadImage(ctx, imageDomain.UploadImageInput{
			UserID:      id,
			Filename:    input.ImageFilename,
			ContentType: input.ImageContentType,
			Data:        input.ImageData,
		})
		if uploadErr != nil {
			slog.Error("ProductService.Create failed",
				slog.String("operation", "upload_image"),
				slog.String("filename", input.ImageFilename),
				slog.Any("error", uploadErr),
			)
			return nil, fmt.Errorf("uploading image: %w", uploadErr)
		}
		imageURL = &result.URL
	}

	product, err := domain.NewProduct(id, input.StoreID, input.Name, input.Description, input.Category, input.Stock, input.Price, imageURL)
	if err != nil {
		slog.Error("ProductService.Create failed",
			slog.String("operation", "create_product_domain"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Debug("ProductService.Create saving product to repository", slog.String("product_id", id))
	if err := s.repo.Save(ctx, product); err != nil {
		slog.Error("ProductService.Create failed",
			slog.String("operation", "save_product"),
			slog.String("product_id", id),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("ProductService.Create completed", slog.String("product_id", product.ID()))
	return product, nil
}

func (s *ProductService) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	slog.Debug("ProductService.GetByID started", slog.String("product_id", id))

	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.Error("ProductService.GetByID failed",
			slog.String("operation", "find_by_id"),
			slog.String("product_id", id),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("ProductService.GetByID completed", slog.String("product_id", id))
	return product, nil
}

func (s *ProductService) GetAll(ctx context.Context, input GetAllProductsInput) ([]*domain.Product, error) {
	slog.Debug("ProductService.GetAll started",
		slog.Int("limit", input.Limit),
		slog.Int("offset", input.Offset),
	)

	if input.Limit <= 0 {
		input.Limit = 10
	}
	if input.Offset < 0 {
		input.Offset = 0
	}

	slog.Debug("ProductService.GetAll querying repository",
		slog.Int("limit", input.Limit),
		slog.Int("offset", input.Offset),
	)
	products, err := s.repo.FindAll(ctx, input.Limit, input.Offset, input.Category)
	if err != nil {
		slog.Error("ProductService.GetAll failed",
			slog.String("operation", "find_all"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("ProductService.GetAll completed", slog.Int("count", len(products)))
	return products, nil
}

func (s *ProductService) GetByStoreID(ctx context.Context, storeID string, limit, offset int) ([]*domain.Product, error) {
	slog.Debug("ProductService.GetByStoreID started",
		slog.String("store_id", storeID),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	if _, err := s.storeRepo.FindByID(ctx, storeID); err != nil {
		slog.Error("ProductService.GetByStoreID failed",
			slog.String("operation", "find_store_by_id"),
			slog.String("store_id", storeID),
			slog.Any("error", err),
		)
		return nil, err
	}

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	slog.Debug("ProductService.GetByStoreID querying repository",
		slog.String("store_id", storeID),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)
	products, err := s.repo.FindByStoreID(ctx, storeID, limit, offset)
	if err != nil {
		slog.Error("ProductService.GetByStoreID failed",
			slog.String("operation", "find_by_store_id"),
			slog.String("store_id", storeID),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("ProductService.GetByStoreID completed",
		slog.String("store_id", storeID),
		slog.Int("count", len(products)),
	)
	return products, nil
}

func (s *ProductService) Update(ctx context.Context, input *UpdateProductInput) (*domain.Product, error) {
	slog.Debug("ProductService.Update started", slog.String("product_id", input.ID))

	product, err := s.repo.FindByID(ctx, input.ID)
	if err != nil {
		slog.Error("ProductService.Update failed",
			slog.String("operation", "find_by_id"),
			slog.String("product_id", input.ID),
			slog.Any("error", err),
		)
		return nil, err
	}

	if err := product.Update(input.Name, input.Description, input.Category, input.Stock, input.Price, input.ImageURL); err != nil {
		slog.Error("ProductService.Update failed",
			slog.String("operation", "update_product_domain"),
			slog.String("product_id", input.ID),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Debug("ProductService.Update saving to repository", slog.String("product_id", input.ID))
	if err := s.repo.Update(ctx, product); err != nil {
		slog.Error("ProductService.Update failed",
			slog.String("operation", "update_repository"),
			slog.String("product_id", input.ID),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("ProductService.Update completed", slog.String("product_id", product.ID()))
	return product, nil
}

func (s *ProductService) Delete(ctx context.Context, id string) error {
	slog.Debug("ProductService.Delete started", slog.String("product_id", id))

	if err := s.repo.Delete(ctx, id); err != nil {
		slog.Error("ProductService.Delete failed",
			slog.String("operation", "delete"),
			slog.String("product_id", id),
			slog.Any("error", err),
		)
		return err
	}

	slog.Info("ProductService.Delete completed", slog.String("product_id", id))
	return nil
}

type UploadProductImageInput struct {
	ProductID   string
	StoreID     string
	Filename    string
	ContentType string
	Data        []byte
}

func (s *ProductService) UploadImage(ctx context.Context, input *UploadProductImageInput) (*domain.Product, error) {
	slog.Debug("ProductService.UploadImage started",
		slog.String("product_id", input.ProductID),
		slog.String("filename", input.Filename),
	)

	product, err := s.repo.FindByID(ctx, input.ProductID)
	if err != nil {
		slog.Error("ProductService.UploadImage failed",
			slog.String("operation", "find_by_id"),
			slog.String("product_id", input.ProductID),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Debug("ProductService.UploadImage uploading to image service",
		slog.String("product_id", input.ProductID),
		slog.Int("data_size", len(input.Data)),
	)
	result, err := s.imageClient.UploadImage(ctx, imageDomain.UploadImageInput{
		UserID: input.ProductID, Filename: input.Filename,
		ContentType: input.ContentType, Data: input.Data,
	})
	if err != nil {
		slog.Error("ProductService.UploadImage failed",
			slog.String("operation", "upload_image"),
			slog.String("product_id", input.ProductID),
			slog.String("filename", input.Filename),
			slog.Any("error", err),
		)
		return nil, fmt.Errorf("uploading image: %w", err)
	}

	if err := product.Update(product.Name(), product.Description(), product.Category(), product.Stock(), product.Price(), &result.URL); err != nil {
		slog.Error("ProductService.UploadImage failed",
			slog.String("operation", "update_product"),
			slog.String("product_id", input.ProductID),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Debug("ProductService.UploadImage saving product with new image", slog.String("product_id", input.ProductID))
	if err := s.repo.Update(ctx, product); err != nil {
		slog.Error("ProductService.UploadImage failed",
			slog.String("operation", "update_repository"),
			slog.String("product_id", input.ProductID),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("ProductService.UploadImage completed", slog.String("product_id", product.ID()))
	return product, nil
}
