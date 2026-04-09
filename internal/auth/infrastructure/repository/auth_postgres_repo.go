package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
	"github.com/lib/pq"
)

type AuthPostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *AuthPostgresRepository {
	return &AuthPostgresRepository{db: db}
}

func (r *AuthPostgresRepository) Save(ctx context.Context, a *domain.Auth) error {
	query := `
		INSERT INTO auths (id, email, password, account_id, account_type)
		VALUES ($1, $2, $3, $4, $5)
	`
	slog.Debug("AuthPostgresRepository.Save executing",
		slog.String("query", query),
		slog.String("id", string(a.ID())),
		slog.String("email", a.Email()),
		slog.String("account_id", string(a.AccountID())),
		slog.String("account_type", string(a.AccountType())),
	)

	_, err := r.db.ExecContext(ctx, query,
		string(a.ID()),
		a.Email(),
		a.Password(),
		string(a.AccountID()),
		string(a.AccountType()),
	)
	if err != nil {
		slog.Error("AuthPostgresRepository.Save failed",
			slog.String("query", query),
			slog.String("id", string(a.ID())),
			slog.String("email", a.Email()),
			slog.Any("error", err),
		)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return apiDomain.ErrEmailExists
		}
		return err
	}
	return nil
}

func (r *AuthPostgresRepository) FindByID(ctx context.Context, id string) (*domain.Auth, error) {
	query := `
		SELECT id, email, password, account_id, account_type
		FROM auths
		WHERE id = $1
	`
	slog.Debug("AuthPostgresRepository.FindByID executing",
		slog.String("query", query),
		slog.String("id", id),
	)

	row := r.db.QueryRowContext(ctx, query, string(id))
	auth, err := r.scanAuth(row)
	if err != nil {
		slog.Error("AuthPostgresRepository.FindByID failed",
			slog.String("query", query),
			slog.String("id", id),
			slog.Any("error", err),
		)
		return nil, err
	}
	return auth, nil
}

func (r *AuthPostgresRepository) FindByEmail(ctx context.Context, email string) (*domain.Auth, error) {
	query := `
		SELECT id, email, password, account_id, account_type
		FROM auths
		WHERE email = $1
	`
	slog.Debug("AuthPostgresRepository.FindByEmail executing",
		slog.String("query", query),
		slog.String("email", email),
	)

	row := r.db.QueryRowContext(ctx, query, email)
	auth, err := r.scanAuth(row)
	if err != nil {
		slog.Error("AuthPostgresRepository.FindByEmail failed",
			slog.String("query", query),
			slog.String("email", email),
			slog.Any("error", err),
		)
		return nil, err
	}
	return auth, nil
}

func (r *AuthPostgresRepository) Update(ctx context.Context, a *domain.Auth) error {
	query := `
		UPDATE auths
		SET email = $2, password = $3
		WHERE id = $1
	`
	slog.Debug("AuthPostgresRepository.Update executing",
		slog.String("query", query),
		slog.String("id", string(a.ID())),
		slog.String("email", a.Email()),
	)

	result, err := r.db.ExecContext(ctx, query,
		string(a.ID()),
		a.Email(),
		a.Password(),
	)
	if err != nil {
		slog.Error("AuthPostgresRepository.Update failed",
			slog.String("query", query),
			slog.String("id", string(a.ID())),
			slog.String("email", a.Email()),
			slog.Any("error", err),
		)
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return apiDomain.ErrEmailExists
		}
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		slog.Error("AuthPostgresRepository.Update failed to get rows affected",
			slog.String("id", string(a.ID())),
			slog.Any("error", err),
		)
		return err
	}
	slog.Debug("AuthPostgresRepository.Update rows affected",
		slog.String("id", string(a.ID())),
		slog.Int64("rows", rows),
	)
	if rows == 0 {
		return apiDomain.ErrAuthNotFound
	}
	return nil
}

func (r *AuthPostgresRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM auths WHERE id = $1`
	slog.Debug("AuthPostgresRepository.Delete executing",
		slog.String("query", query),
		slog.String("id", id),
	)

	result, err := r.db.ExecContext(ctx, query, string(id))
	if err != nil {
		slog.Error("AuthPostgresRepository.Delete failed",
			slog.String("query", query),
			slog.String("id", id),
			slog.Any("error", err),
		)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		slog.Error("AuthPostgresRepository.Delete failed to get rows affected",
			slog.String("id", id),
			slog.Any("error", err),
		)
		return err
	}
	slog.Debug("AuthPostgresRepository.Delete rows affected",
		slog.String("id", id),
		slog.Int64("rows", rows),
	)
	if rows == 0 {
		return apiDomain.ErrAuthNotFound
	}
	return nil
}

func (r *AuthPostgresRepository) scanAuth(row *sql.Row) (*domain.Auth, error) {
	var id, email, password, accountID, accountType string
	if err := row.Scan(&id, &email, &password, &accountID, &accountType); err != nil {
		if err == sql.ErrNoRows {
			return nil, apiDomain.ErrAuthNotFound
		}
		return nil, err
	}
	return domain.NewAuth(
		id,
		email,
		password,
		accountID,
		domain.AccountType(accountType),
	)
}
