package application

import (
	"context"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
)

type Service struct {
	repo     domain.Repository
	idGen    IDGenerator
	hasher   PasswordHasher
	tokenGen TokenGenerator
	userSvc  UserService
}

type IDGenerator interface {
	GenerateAuthID() domain.ID
	GenerateUserID() domain.UserID
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashed, plain string) error
}

type TokenGenerator interface {
	Generate(authID domain.ID, userID domain.UserID) (string, error)
}

type UserService interface {
	Create(ctx context.Context, firstName, lastName string, id domain.UserID) error
}

func NewService(
	repo domain.Repository,
	idGen IDGenerator,
	hasher PasswordHasher,
	tokenGen TokenGenerator,
	userSvc UserService,
) *Service {
	return &Service{
		repo:     repo,
		idGen:    idGen,
		hasher:   hasher,
		tokenGen: tokenGen,
		userSvc:  userSvc,
	}
}

type RegisterInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type RegisterOutput struct {
	AuthID    domain.ID
	UserID    domain.UserID
	Email     string
	FirstName string
	LastName  string
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*RegisterOutput, error) {
	existing, err := s.repo.FindByEmail(ctx, input.Email)
	if err == nil && existing != nil {
		return nil, domain.ErrEmailExists
	}

	userID := s.idGen.GenerateUserID()

	if err := s.userSvc.Create(ctx, input.FirstName, input.LastName, userID); err != nil {
		return nil, err
	}

	hashedPassword, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	authID := s.idGen.GenerateAuthID()
	a, err := domain.NewAuth(authID, input.Email, hashedPassword, userID)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, a); err != nil {
		return nil, err
	}

	return &RegisterOutput{
		AuthID:    a.ID(),
		UserID:    userID,
		Email:     a.Email(),
		FirstName: input.FirstName,
		LastName:  input.LastName,
	}, nil
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	Token  string
	UserID domain.UserID
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginOutput, error) {
	a, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := s.hasher.Compare(a.Password(), input.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	token, err := s.tokenGen.Generate(a.ID(), a.UserID())
	if err != nil {
		return nil, err
	}

	return &LoginOutput{
		Token:  token,
		UserID: a.UserID(),
	}, nil
}
