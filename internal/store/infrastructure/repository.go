package infrastructure

import (
	"context"
	"database/sql"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(ctx context.Context, store *domain.Store) error {
	query := `
		INSERT INTO stores (id, name, description, address, phone_number)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		string(store.ID()),
		store.Name(),
		store.Description(),
		store.Address(),
		store.PhoneNumber(),
	)
	return err
}

func (r *PostgresRepository) FindByID(ctx context.Context, id domain.ID) (*domain.Store, error) {
	query := `
		SELECT id, name, description, address, phone_number
		FROM stores
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, string(id))
	return r.scanStore(row)
}

func (r *PostgresRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.Store, error) {
	query := `
		SELECT id, name, description, address, phone_number
		FROM stores
		ORDER BY name
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []*domain.Store
	for rows.Next() {
		store, err := r.scanStoreFromRows(rows)
		if err != nil {
			return nil, err
		}
		stores = append(stores, store)
	}
	return stores, rows.Err()
}

func (r *PostgresRepository) Update(ctx context.Context, store *domain.Store) error {
	query := `
		UPDATE stores
		SET name = $2, description = $3, address = $4, phone_number = $5
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query,
		string(store.ID()),
		store.Name(),
		store.Description(),
		store.Address(),
		store.PhoneNumber(),
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrStoreNotFound
	}
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id domain.ID) error {
	query := `DELETE FROM stores WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, string(id))
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrStoreNotFound
	}
	return nil
}

func (r *PostgresRepository) scanStore(row *sql.Row) (*domain.Store, error) {
	var id, name, description, address, phoneNumber string
	if err := row.Scan(&id, &name, &description, &address, &phoneNumber); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrStoreNotFound
		}
		return nil, err
	}
	return domain.NewStore(domain.ID(id), name, description, address, phoneNumber)
}

func (r *PostgresRepository) scanStoreFromRows(rows *sql.Rows) (*domain.Store, error) {
	var id, name, description, address, phoneNumber string
	if err := rows.Scan(&id, &name, &description, &address, &phoneNumber); err != nil {
		return nil, err
	}
	return domain.NewStore(domain.ID(id), name, description, address, phoneNumber)
}
