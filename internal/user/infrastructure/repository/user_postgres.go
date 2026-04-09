package repository

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/domain"
)

type UserPostgresRepository struct {
	db *sql.DB
}

func NewUserPostgresRepository(db *sql.DB) *UserPostgresRepository {
	return &UserPostgresRepository{db: db}
}

func (r *UserPostgresRepository) Save(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (id, first_name, last_name)
		VALUES ($1, $2, $3)
	`
	slog.Debug("UserPostgresRepository.Save executing",
		slog.String("query", query),
		slog.String("id", string(u.ID())),
		slog.String("first_name", u.FirstName()),
		slog.String("last_name", u.LastName()),
	)

	_, err := r.db.ExecContext(ctx, query,
		string(u.ID()),
		u.FirstName(),
		u.LastName(),
	)
	if err != nil {
		slog.Error("UserPostgresRepository.Save failed",
			slog.String("query", query),
			slog.String("id", string(u.ID())),
			slog.Any("error", err),
		)
		return err
	}
	return nil
}

func (r *UserPostgresRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, first_name, last_name
		FROM users
		WHERE id = $1
	`
	slog.Debug("UserPostgresRepository.FindByID executing",
		slog.String("query", query),
		slog.String("id", id),
	)

	var userID, firstName, lastName string
	err := r.db.QueryRowContext(ctx, query, string(id)).Scan(&userID, &firstName, &lastName)
	if err != nil {
		slog.Error("UserPostgresRepository.FindByID failed",
			slog.String("query", query),
			slog.String("id", id),
			slog.Any("error", err),
		)
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return domain.NewUser(string(userID), firstName, lastName)
}

func (r *UserPostgresRepository) Update(ctx context.Context, u *domain.User) error {
	query := `
		UPDATE users
		SET first_name = $2, last_name = $3
		WHERE id = $1
	`
	slog.Debug("UserPostgresRepository.Update executing",
		slog.String("query", query),
		slog.String("id", string(u.ID())),
		slog.String("first_name", u.FirstName()),
		slog.String("last_name", u.LastName()),
	)

	result, err := r.db.ExecContext(ctx, query,
		string(u.ID()),
		u.FirstName(),
		u.LastName(),
	)
	if err != nil {
		slog.Error("UserPostgresRepository.Update failed",
			slog.String("query", query),
			slog.String("id", string(u.ID())),
			slog.Any("error", err),
		)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		slog.Error("UserPostgresRepository.Update failed to get rows affected",
			slog.String("id", string(u.ID())),
			slog.Any("error", err),
		)
		return err
	}
	slog.Debug("UserPostgresRepository.Update rows affected",
		slog.String("id", string(u.ID())),
		slog.Int64("rows", rows),
	)
	if rows == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *UserPostgresRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	slog.Debug("UserPostgresRepository.Delete executing",
		slog.String("query", query),
		slog.String("id", id),
	)

	result, err := r.db.ExecContext(ctx, query, string(id))
	if err != nil {
		slog.Error("UserPostgresRepository.Delete failed",
			slog.String("query", query),
			slog.String("id", id),
			slog.Any("error", err),
		)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		slog.Error("UserPostgresRepository.Delete failed to get rows affected",
			slog.String("id", id),
			slog.Any("error", err),
		)
		return err
	}
	slog.Debug("UserPostgresRepository.Delete rows affected",
		slog.String("id", id),
		slog.Int64("rows", rows),
	)
	if rows == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
