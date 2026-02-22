package application

import (
	"context"

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
}

func (s *AuthService) RegisterUser(ctx context.Context, input RegisterUserInput) (*RegisterOutput, error) {
	if err := s.checkEmailAvailable(ctx, input.Email); err != nil {
		return nil, err
	}

	createUserInput := userApplication.CreateInput{
		FirstName: input.FirstName,
		LastName:  input.LastName,
	}
	user, err := s.userSvc.Create(ctx, createUserInput)
	if err != nil {
		return nil, err
	}

	return s.createAuth(ctx, input.Email, input.Password, user.ID, domain.AccountTypeUser)
}

func (s *AuthService) RegisterStore(ctx context.Context, input RegisterStoreInput) (*RegisterOutput, error) {
	if err := s.checkEmailAvailable(ctx, input.Email); err != nil {
		return nil, err
	}

	createStoreInput := storeApplication.CreateStoreInput{
		Name:        input.Name,
		Description: input.Description,
		Address:     input.Address,
		PhoneNumber: input.PhoneNumber,
	}

	store, err := s.storeSvc.Create(ctx, createStoreInput)
	if err != nil {
		return nil, err
	}

	return s.createAuth(ctx, input.Email, input.Password, store.ID(), domain.AccountTypeStore)
}

func (s *AuthService) checkEmailAvailable(ctx context.Context, email string) error {
	existing, err := s.repo.FindByEmail(ctx, email)
	if err == nil && existing != nil {
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
	hashedPassword, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	authID := s.idGen.Generate()
	auth, err := domain.NewAuth(authID, email, hashedPassword, accountID, accountType)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, auth); err != nil {
		return nil, err
	}

	return &RegisterOutput{
		AuthID:      auth.ID(),
		AccountID:   accountID,
		AccountType: accountType,
		Email:       auth.Email(),
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
	auth, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, apiDomain.ErrInvalidCredentials
	}

	if err := s.hasher.Compare(auth.Password(), input.Password); err != nil {
		return nil, apiDomain.ErrInvalidCredentials
	}

	token, err := s.tokenGen.Generate(auth.ID(), auth.AccountType(), auth.AccountID())
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		Token:       token,
		AccountID:   auth.AccountID(),
		AccountType: auth.AccountType(),
	}, nil
}
