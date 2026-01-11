package infrastructure

import (
	"context"
	"database/sql"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/user/domain"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO users (id, first_name, last_name)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.ExecContext(ctx, query,
		string(u.ID()),
		u.FirstName(),
		u.LastName(),
	)
	return err
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, first_name, last_name
		FROM users
		WHERE id = $1
	`
	var userID, firstName, lastName string
	err := r.db.QueryRowContext(ctx, query, string(id)).Scan(&userID, &firstName, &lastName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return domain.NewUser(string(userID), firstName, lastName)
}

func (r *PostgresRepository) Update(ctx context.Context, u *domain.User) error {
	query := `
		UPDATE users
		SET first_name = $2, last_name = $3
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query,
		string(u.ID()),
		u.FirstName(),
		u.LastName(),
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, string(id))
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
