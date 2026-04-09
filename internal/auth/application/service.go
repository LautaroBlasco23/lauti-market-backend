package application

import (
	"context"
	"log/slog"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
	storeApplication "github.com/LautaroBlasco23/lauti-market-backend/internal/store/application"
	userApplication "github.com/LautaroBlasco23/lauti-market-backend/internal/user/application"
)

type AuthService struct {
	repo     domain.Repository
	idGen    apiDomain.IDGenerator
	hasher   PasswordHasher
	tokenGen TokenGenerator
	userSvc  *userApplication.UserService
	storeSvc *storeApplication.StoreService
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashed, plain string) error
}

type TokenGenerator interface {
	Generate(authID string, accountType domain.AccountType, accountID string) (string, error)
}

func NewService(
	repo domain.Repository,
	idGen apiDomain.IDGenerator,
	hasher PasswordHasher,
	tokenGen TokenGenerator,
	userSvc *userApplication.UserService,
	storeSvc *storeApplication.StoreService,
) *AuthService {
	return &AuthService{
		repo:     repo,
		idGen:    idGen,
		hasher:   hasher,
		tokenGen: tokenGen,
		userSvc:  userSvc,
		storeSvc: storeSvc,
	}
}

type RegisterUserInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type RegisterStoreInput struct {
	Email       string
	Password    string
	Name        string
	Description string
	Address     string
	PhoneNumber string
}

type RegisterOutput struct {
	AuthID      string
	AccountID   string
	AccountType domain.AccountType
	Email       string
	Token       string
}

func (s *AuthService) RegisterUser(ctx context.Context, input RegisterUserInput) (*RegisterOutput, error) {
	slog.Debug("AuthService.RegisterUser started",
		slog.String("email", input.Email),
		slog.String("first_name", input.FirstName),
	)

	if err := s.checkEmailAvailable(ctx, input.Email); err != nil {
		slog.Error("AuthService.RegisterUser failed",
			slog.String("operation", "check_email_available"),
			slog.Any("error", err),
		)
		return nil, err
	}

	createUserInput := userApplication.CreateInput{
		FirstName: input.FirstName,
		LastName:  input.LastName,
	}

	slog.Debug("AuthService.RegisterUser creating user")
	user, err := s.userSvc.Create(ctx, createUserInput)
	if err != nil {
		slog.Error("AuthService.RegisterUser failed",
			slog.String("operation", "create_user"),
			slog.Any("error", err),
		)
		return nil, err
	}

	result, err := s.createAuth(ctx, input.Email, input.Password, user.ID, domain.AccountTypeUser)
	if err != nil {
		slog.Error("AuthService.RegisterUser failed",
			slog.String("operation", "create_auth"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("AuthService.RegisterUser completed",
		slog.String("auth_id", result.AuthID),
		slog.String("account_id", result.AccountID),
		slog.String("account_type", string(result.AccountType)),
	)
	return result, nil
}

func (s *AuthService) RegisterStore(ctx context.Context, input *RegisterStoreInput) (*RegisterOutput, error) {
	slog.Debug("AuthService.RegisterStore started",
		slog.String("email", input.Email),
		slog.String("store_name", input.Name),
	)

	if err := s.checkEmailAvailable(ctx, input.Email); err != nil {
		slog.Error("AuthService.RegisterStore failed",
			slog.String("operation", "check_email_available"),
			slog.Any("error", err),
		)
		return nil, err
	}

	createStoreInput := storeApplication.CreateStoreInput{
		Name:        input.Name,
		Description: input.Description,
		Address:     input.Address,
		PhoneNumber: input.PhoneNumber,
	}

	slog.Debug("AuthService.RegisterStore creating store")
	store, err := s.storeSvc.Create(ctx, createStoreInput)
	if err != nil {
		slog.Error("AuthService.RegisterStore failed",
			slog.String("operation", "create_store"),
			slog.Any("error", err),
		)
		return nil, err
	}

	result, err := s.createAuth(ctx, input.Email, input.Password, store.ID(), domain.AccountTypeStore)
	if err != nil {
		slog.Error("AuthService.RegisterStore failed",
			slog.String("operation", "create_auth"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("AuthService.RegisterStore completed",
		slog.String("auth_id", result.AuthID),
		slog.String("account_id", result.AccountID),
		slog.String("account_type", string(result.AccountType)),
	)
	return result, nil
}

func (s *AuthService) checkEmailAvailable(ctx context.Context, email string) error {
	slog.Debug("AuthService.checkEmailAvailable checking email")
	existing, err := s.repo.FindByEmail(ctx, email)
	if err == nil && existing != nil {
		slog.Error("AuthService.checkEmailAvailable failed",
			slog.String("operation", "find_by_email"),
			slog.Any("error", apiDomain.ErrEmailExists),
		)
		return apiDomain.ErrEmailExists
	}
	return nil
}

func (s *AuthService) createAuth(
	ctx context.Context,
	email, password string,
	accountID string,
	accountType domain.AccountType,
) (*RegisterOutput, error) {
	slog.Debug("AuthService.createAuth hashing password")
	hashedPassword, err := s.hasher.Hash(password)
	if err != nil {
		slog.Error("AuthService.createAuth failed",
			slog.String("operation", "hash_password"),
			slog.Any("error", err),
		)
		return nil, err
	}

	authID := s.idGen.Generate()
	slog.Debug("AuthService.createAuth creating auth entity",
		slog.String("auth_id", authID),
	)
	auth, err := domain.NewAuth(authID, email, hashedPassword, accountID, accountType)
	if err != nil {
		slog.Error("AuthService.createAuth failed",
			slog.String("operation", "new_auth_entity"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Debug("AuthService.createAuth saving auth to repository")
	err = s.repo.Save(ctx, auth)
	if err != nil {
		slog.Error("AuthService.createAuth failed",
			slog.String("operation", "save_auth"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Debug("AuthService.createAuth generating token")
	token, err := s.tokenGen.Generate(auth.ID(), auth.AccountType(), auth.AccountID())
	if err != nil {
		slog.Error("AuthService.createAuth failed",
			slog.String("operation", "generate_token"),
			slog.Any("error", err),
		)
		return nil, err
	}

	return &RegisterOutput{
		AuthID:      auth.ID(),
		AccountID:   accountID,
		AccountType: accountType,
		Email:       auth.Email(),
		Token:       token,
	}, nil
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	Token       string
	AccountID   string
	AccountType domain.AccountType
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	slog.Debug("AuthService.Login started",
		slog.String("email", input.Email),
	)

	slog.Debug("AuthService.Login finding auth by email")
	auth, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		slog.Error("AuthService.Login failed",
			slog.String("operation", "find_by_email"),
			slog.Any("error", apiDomain.ErrInvalidCredentials),
		)
		return nil, apiDomain.ErrInvalidCredentials
	}

	slog.Debug("AuthService.Login comparing password")
	if compareErr := s.hasher.Compare(auth.Password(), input.Password); compareErr != nil {
		slog.Error("AuthService.Login failed",
			slog.String("operation", "compare_password"),
			slog.Any("error", apiDomain.ErrInvalidCredentials),
		)
		return nil, apiDomain.ErrInvalidCredentials
	}

	slog.Debug("AuthService.Login generating token")
	token, err := s.tokenGen.Generate(auth.ID(), auth.AccountType(), auth.AccountID())
	if err != nil {
		slog.Error("AuthService.Login failed",
			slog.String("operation", "generate_token"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("AuthService.Login completed",
		slog.String("auth_id", auth.ID()),
		slog.String("account_id", auth.AccountID()),
		slog.String("account_type", string(auth.AccountType())),
	)
	return &LoginOutput{
		Token:       token,
		AccountID:   auth.AccountID(),
		AccountType: auth.AccountType(),
	}, nil
}
