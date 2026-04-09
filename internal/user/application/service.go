package application

import (
	"context"
	"log/slog"

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
	slog.Debug("UserService.Create started",
		slog.String("first_name", input.FirstName),
		slog.String("last_name", input.LastName),
	)

	id := s.idGen.Generate()
	slog.Debug("UserService.Create creating user entity",
		slog.String("id", id),
	)
	u, err := domain.NewUser(id, input.FirstName, input.LastName)
	if err != nil {
		slog.Error("UserService.Create failed",
			slog.String("operation", "new_user_entity"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Debug("UserService.Create saving user to repository")
	if err := s.repo.Save(ctx, u); err != nil {
		slog.Error("UserService.Create failed",
			slog.String("operation", "save_user"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("UserService.Create completed",
		slog.String("id", u.ID()),
	)
	return &Output{
		ID:        u.ID(),
		FirstName: u.FirstName(),
		LastName:  u.LastName(),
	}, nil
}

func (s *UserService) GetByID(ctx context.Context, id string) (*Output, error) {
	slog.Debug("UserService.GetByID started",
		slog.String("id", id),
	)

	slog.Debug("UserService.GetByID finding user by ID")
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		slog.Error("UserService.GetByID failed",
			slog.String("operation", "find_by_id"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("UserService.GetByID completed",
		slog.String("id", u.ID()),
	)
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
	slog.Debug("UserService.Update started",
		slog.String("id", input.ID),
		slog.String("first_name", input.FirstName),
		slog.String("last_name", input.LastName),
	)

	slog.Debug("UserService.Update finding user by ID")
	u, err := s.repo.FindByID(ctx, input.ID)
	if err != nil {
		slog.Error("UserService.Update failed",
			slog.String("operation", "find_by_id"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Debug("UserService.Update updating user name")
	if err := u.UpdateName(input.FirstName, input.LastName); err != nil {
		slog.Error("UserService.Update failed",
			slog.String("operation", "update_name"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Debug("UserService.Update saving user to repository")
	if err := s.repo.Update(ctx, u); err != nil {
		slog.Error("UserService.Update failed",
			slog.String("operation", "update_user"),
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Info("UserService.Update completed",
		slog.String("id", u.ID()),
	)
	return &Output{
		ID:        u.ID(),
		FirstName: u.FirstName(),
		LastName:  u.LastName(),
	}, nil
}

func (s *UserService) Delete(ctx context.Context, id string) error {
	slog.Debug("UserService.Delete started",
		slog.String("id", id),
	)

	slog.Debug("UserService.Delete deleting user from repository")
	if err := s.repo.Delete(ctx, id); err != nil {
		slog.Error("UserService.Delete failed",
			slog.String("operation", "delete_user"),
			slog.Any("error", err),
		)
		return err
	}

	slog.Info("UserService.Delete completed",
		slog.String("id", id),
	)
	return nil
}
