package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/auth/domain"
	"github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(ctx context.Context, a *domain.Auth) error {
	query := `
		INSERT INTO auths (id, email, password, account_id, account_type)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		string(a.ID()),
		a.Email(),
		a.Password(),
		string(a.AccountID()),
		string(a.AccountType()),
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return domain.ErrEmailExists
		}
		return err
	}
	return nil
}

func (r *PostgresRepository) FindByID(ctx context.Context, id domain.ID) (*domain.Auth, error) {
	query := `
		SELECT id, email, password, account_id, account_type
		FROM auths
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, string(id))
	return r.scanAuth(row)
}

func (r *PostgresRepository) FindByEmail(ctx context.Context, email string) (*domain.Auth, error) {
	query := `
		SELECT id, email, password, account_id, account_type
		FROM auths
		WHERE email = $1
	`
	row := r.db.QueryRowContext(ctx, query, email)
	return r.scanAuth(row)
}

func (r *PostgresRepository) Update(ctx context.Context, a *domain.Auth) error {
	query := `
		UPDATE auths
		SET email = $2, password = $3
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query,
		string(a.ID()),
		a.Email(),
		a.Password(),
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return domain.ErrEmailExists
		}
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrAuthNotFound
	}
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id domain.ID) error {
	query := `DELETE FROM auths WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, string(id))
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrAuthNotFound
	}
	return nil
}

func (r *PostgresRepository) scanAuth(row *sql.Row) (*domain.Auth, error) {
	var id, email, password, accountID, accountType string
	if err := row.Scan(&id, &email, &password, &accountID, &accountType); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrAuthNotFound
		}
		return nil, err
	}
	return domain.NewAuth(
		domain.ID(id),
		email,
		password,
		domain.AccountID(accountID),
		domain.AccountType(accountType),
	)
}
