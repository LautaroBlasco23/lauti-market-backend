package application

import (
	"context"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
)

type Service struct {
	repo     domain.Repository
	idGen    apiDomain.IDGenerator
	hasher   PasswordHasher
	tokenGen TokenGenerator
	userSvc  UserService
	storeSvc StoreService
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashed, plain string) error
}

type TokenGenerator interface {
	Generate(authID string, accountType domain.AccountType, accountID string) (string, error)
}

type UserService interface {
	Create(ctx context.Context, firstName, lastName string, id string) error
}

type StoreService interface {
	Create(ctx context.Context, name, description, address, phoneNumber string, id string) error
}

func NewService(
	repo domain.Repository,
	idGen apiDomain.IDGenerator,
	hasher PasswordHasher,
	tokenGen TokenGenerator,
	userSvc UserService,
	storeSvc StoreService,
) *Service {
	return &Service{
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

func (s *Service) RegisterUser(ctx context.Context, input RegisterUserInput) (*RegisterOutput, error) {
	if err := s.checkEmailAvailable(ctx, input.Email); err != nil {
		return nil, err
	}

	accountID := s.idGen.Generate()
	if err := s.userSvc.Create(ctx, input.FirstName, input.LastName, accountID); err != nil {
		return nil, err
	}

	return s.createAuth(ctx, input.Email, input.Password, accountID, domain.AccountTypeUser)
}

func (s *Service) RegisterStore(ctx context.Context, input RegisterStoreInput) (*RegisterOutput, error) {
	if err := s.checkEmailAvailable(ctx, input.Email); err != nil {
		return nil, err
	}

	accountID := s.idGen.Generate()
	if err := s.storeSvc.Create(ctx, input.Name, input.Description, input.Address, input.PhoneNumber, accountID); err != nil {
		return nil, err
	}

	return s.createAuth(ctx, input.Email, input.Password, accountID, domain.AccountTypeStore)
}

func (s *Service) checkEmailAvailable(ctx context.Context, email string) error {
	existing, err := s.repo.FindByEmail(ctx, email)
	if err == nil && existing != nil {
		return domain.ErrEmailExists
	}
	return nil
}

func (s *Service) createAuth(
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

func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	auth, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := s.hasher.Compare(auth.Password(), input.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
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
