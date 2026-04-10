package application

import (
	"context"
	"log/slog"
	"time"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/infrastructure/mercadopago"
)

type StoreService struct {
	repo    domain.Repository
	idGen   apiDomain.IDGenerator
	mpOAuth *mercadopago.OAuthClient
}

func NewService(repo domain.Repository, idGen apiDomain.IDGenerator, mpOAuth *mercadopago.OAuthClient) *StoreService {
	return &StoreService{
		repo:    repo,
		idGen:   idGen,
		mpOAuth: mpOAuth,
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
	storeInput := domain.CreateStoreInput{
		Name:        input.Name,
		Description: input.Description,
		Address:     input.Address,
		PhoneNumber: input.PhoneNumber,
	}
	store, err := domain.NewStore(id, storeInput)
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

// GetOAuthConnectURL generates a MercadoPago OAuth authorization URL for the store
func (s *StoreService) GetOAuthConnectURL(ctx context.Context, storeID string) (string, error) {
	slog.Debug("StoreService.GetOAuthConnectURL started",
		slog.String("store_id", storeID),
	)

	slog.Debug("StoreService.GetOAuthConnectURL finding store by ID")
	store, err := s.repo.FindByID(ctx, storeID)
	if err != nil {
		slog.Error("StoreService.GetOAuthConnectURL failed",
			slog.String("operation", "find_by_id"),
			slog.Any("error", err),
		)
		return "", err
	}

	state := store.ID()
	authURL := s.mpOAuth.GetAuthorizationURL(state)

	slog.Info("StoreService.GetOAuthConnectURL completed",
		slog.String("store_id", storeID),
	)
	return authURL, nil
}

// HandleOAuthCallback exchanges the authorization code for tokens and saves them
func (s *StoreService) HandleOAuthCallback(ctx context.Context, storeID string, code string) error {
	slog.Debug("StoreService.HandleOAuthCallback started",
		slog.String("store_id", storeID),
	)

	slog.Debug("StoreService.HandleOAuthCallback finding store by ID")
	store, err := s.repo.FindByID(ctx, storeID)
	if err != nil {
		slog.Error("StoreService.HandleOAuthCallback failed",
			slog.String("operation", "find_by_id"),
			slog.Any("error", err),
		)
		return err
	}

	slog.Debug("StoreService.HandleOAuthCallback exchanging code for tokens")
	tokenResp, err := s.mpOAuth.ExchangeCode(ctx, code)
	if err != nil {
		slog.Error("StoreService.HandleOAuthCallback failed",
			slog.String("operation", "exchange_code"),
			slog.Any("error", err),
		)
		return err
	}

	expiresAt := tokenResp.CalculateExpiryTime()
	fields := domain.MPFields{
		MPUserID:         tokenResp.MPUserID,
		MPAccessToken:    tokenResp.AccessToken,
		MPRefreshToken:   tokenResp.RefreshToken,
		MPTokenExpiresAt: &expiresAt,
		MPConnectedAt:    func() *time.Time { now := time.Now(); return &now }(),
	}

	slog.Debug("StoreService.HandleOAuthCallback updating store MP connection")
	if err := s.repo.UpdateMPConnection(ctx, store.ID(), fields); err != nil {
		slog.Error("StoreService.HandleOAuthCallback failed",
			slog.String("operation", "update_mp_connection"),
			slog.Any("error", err),
		)
		return err
	}

	slog.Info("StoreService.HandleOAuthCallback completed",
		slog.String("store_id", store.ID()),
		slog.String("mp_user_id", tokenResp.MPUserID),
	)
	return nil
}

// RefreshAccessToken refreshes the MercadoPago access token if expired
func (s *StoreService) RefreshAccessToken(ctx context.Context, storeID string) error {
	slog.Debug("StoreService.RefreshAccessToken started",
		slog.String("store_id", storeID),
	)

	slog.Debug("StoreService.RefreshAccessToken finding store by ID")
	store, err := s.repo.FindByID(ctx, storeID)
	if err != nil {
		slog.Error("StoreService.RefreshAccessToken failed",
			slog.String("operation", "find_by_id"),
			slog.Any("error", err),
		)
		return err
	}

	if store.IsMPTokenValid() {
		slog.Debug("StoreService.RefreshAccessToken token is still valid, skipping refresh",
			slog.String("store_id", storeID),
		)
		return nil
	}

	refreshToken := store.MPRefreshToken()
	if refreshToken == "" {
		slog.Error("StoreService.RefreshAccessToken failed",
			slog.String("operation", "check_refresh_token"),
			slog.String("error", "no refresh token available"),
		)
		return domain.ErrMPInvalidToken
	}

	slog.Debug("StoreService.RefreshAccessToken refreshing token")
	tokenResp, err := s.mpOAuth.RefreshToken(ctx, refreshToken)
	if err != nil {
		slog.Error("StoreService.RefreshAccessToken failed",
			slog.String("operation", "refresh_token"),
			slog.Any("error", err),
		)
		return err
	}

	expiresAt := tokenResp.CalculateExpiryTime()
	fields := domain.MPFields{
		MPUserID:         tokenResp.MPUserID,
		MPAccessToken:    tokenResp.AccessToken,
		MPRefreshToken:   tokenResp.RefreshToken,
		MPTokenExpiresAt: &expiresAt,
		MPConnectedAt:    store.MPConnectedAt(),
	}

	slog.Debug("StoreService.RefreshAccessToken updating store MP connection")
	if err := s.repo.UpdateMPConnection(ctx, store.ID(), fields); err != nil {
		slog.Error("StoreService.RefreshAccessToken failed",
			slog.String("operation", "update_mp_connection"),
			slog.Any("error", err),
		)
		return err
	}

	slog.Info("StoreService.RefreshAccessToken completed",
		slog.String("store_id", store.ID()),
	)
	return nil
}

// GetMPConnectionStatus returns the MercadoPago connection status for a store
func (s *StoreService) GetMPConnectionStatus(ctx context.Context, storeID string) (map[string]interface{}, error) {
	slog.Debug("StoreService.GetMPConnectionStatus started",
		slog.String("store_id", storeID),
	)

	slog.Debug("StoreService.GetMPConnectionStatus finding store by ID")
	store, err := s.repo.FindByID(ctx, storeID)
	if err != nil {
		slog.Error("StoreService.GetMPConnectionStatus failed",
			slog.String("operation", "find_by_id"),
			slog.Any("error", err),
		)
		return nil, err
	}

	status := map[string]interface{}{
		"connected":      store.IsMPConnected(),
		"connected_at":   store.MPConnectedAt(),
		"expires_at":     store.MPTokenExpiresAt(),
		"is_token_valid": store.IsMPTokenValid(),
	}

	slog.Info("StoreService.GetMPConnectionStatus completed",
		slog.String("store_id", storeID),
	)
	return status, nil
}

// DisconnectMP removes the MercadoPago connection from a store
func (s *StoreService) DisconnectMP(ctx context.Context, storeID string) error {
	slog.Debug("StoreService.DisconnectMP started",
		slog.String("store_id", storeID),
	)

	slog.Debug("StoreService.DisconnectMP finding store by ID")
	store, err := s.repo.FindByID(ctx, storeID)
	if err != nil {
		slog.Error("StoreService.DisconnectMP failed",
			slog.String("operation", "find_by_id"),
			slog.Any("error", err),
		)
		return err
	}

	// Clear all MP connection fields
	fields := domain.MPFields{
		MPUserID:         "",
		MPAccessToken:    "",
		MPRefreshToken:   "",
		MPTokenExpiresAt: nil,
		MPConnectedAt:    nil,
	}

	slog.Debug("StoreService.DisconnectMP clearing store MP connection")
	if err := s.repo.UpdateMPConnection(ctx, store.ID(), fields); err != nil {
		slog.Error("StoreService.DisconnectMP failed",
			slog.String("operation", "update_mp_connection"),
			slog.Any("error", err),
		)
		return err
	}

	slog.Info("StoreService.DisconnectMP completed",
		slog.String("store_id", store.ID()),
	)
	return nil
}
