package repository

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/LautaroBlasco23/lauti-market-backend/internal/store/domain"
)

type StorePostgresRepository struct {
	db *sql.DB
}

func NewStorePostgresRepository(db *sql.DB) *StorePostgresRepository {
	return &StorePostgresRepository{db: db}
}

func (r *StorePostgresRepository) Save(ctx context.Context, store *domain.Store) error {
	query := `
		INSERT INTO stores (id, name, description, address, phone_number)
		VALUES ($1, $2, $3, $4, $5)
	`
	slog.Debug("StorePostgresRepository.Save executing",
		slog.String("query", query),
		slog.String("id", string(store.ID())),
		slog.String("name", store.Name()),
	)

	_, err := r.db.ExecContext(ctx, query,
		string(store.ID()),
		store.Name(),
		store.Description(),
		store.Address(),
		store.PhoneNumber(),
	)
	if err != nil {
		slog.Error("StorePostgresRepository.Save failed",
			slog.String("query", query),
			slog.String("id", string(store.ID())),
			slog.String("name", store.Name()),
			slog.Any("error", err),
		)
		return err
	}
	return nil
}

func (r *StorePostgresRepository) FindByID(ctx context.Context, id string) (*domain.Store, error) {
	query := `
		SELECT id, name, description, address, phone_number
		FROM stores
		WHERE id = $1
	`
	slog.Debug("StorePostgresRepository.FindByID executing",
		slog.String("query", query),
		slog.String("id", id),
	)

	row := r.db.QueryRowContext(ctx, query, string(id))
	store, err := r.scanStore(row)
	if err != nil {
		slog.Error("StorePostgresRepository.FindByID failed",
			slog.String("query", query),
			slog.String("id", id),
			slog.Any("error", err),
		)
		return nil, err
	}
	return store, nil
}

func (r *StorePostgresRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.Store, error) {
	query := `
		SELECT id, name, description, address, phone_number
		FROM stores
		ORDER BY name
		LIMIT $1 OFFSET $2
	`
	slog.Debug("StorePostgresRepository.FindAll executing",
		slog.String("query", query),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		slog.Error("StorePostgresRepository.FindAll failed",
			slog.String("query", query),
			slog.Int("limit", limit),
			slog.Int("offset", offset),
			slog.Any("error", err),
		)
		return nil, err
	}
	defer func() {
		_ = rows.Close() //nolint:errcheck
	}()

	var stores []*domain.Store
	for rows.Next() {
		store, err := r.scanStoreFromRows(rows)
		if err != nil {
			slog.Error("StorePostgresRepository.FindAll failed to scan row",
				slog.Any("error", err),
			)
			return nil, err
		}
		stores = append(stores, store)
	}

	if err := rows.Err(); err != nil {
		slog.Error("StorePostgresRepository.FindAll rows iteration error",
			slog.Any("error", err),
		)
		return nil, err
	}

	slog.Debug("StorePostgresRepository.FindAll rows returned",
		slog.Int("count", len(stores)),
	)
	return stores, nil
}

func (r *StorePostgresRepository) Update(ctx context.Context, store *domain.Store) error {
	query := `
		UPDATE stores
		SET name = $2, description = $3, address = $4, phone_number = $5
		WHERE id = $1
	`
	slog.Debug("StorePostgresRepository.Update executing",
		slog.String("query", query),
		slog.String("id", string(store.ID())),
		slog.String("name", store.Name()),
	)

	result, err := r.db.ExecContext(ctx, query,
		string(store.ID()),
		store.Name(),
		store.Description(),
		store.Address(),
		store.PhoneNumber(),
	)
	if err != nil {
		slog.Error("StorePostgresRepository.Update failed",
			slog.String("query", query),
			slog.String("id", string(store.ID())),
			slog.String("name", store.Name()),
			slog.Any("error", err),
		)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		slog.Error("StorePostgresRepository.Update failed to get rows affected",
			slog.String("id", string(store.ID())),
			slog.Any("error", err),
		)
		return err
	}
	slog.Debug("StorePostgresRepository.Update rows affected",
		slog.String("id", string(store.ID())),
		slog.Int64("rows", rows),
	)
	if rows == 0 {
		return domain.ErrStoreNotFound
	}
	return nil
}

func (r *StorePostgresRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM stores WHERE id = $1`
	slog.Debug("StorePostgresRepository.Delete executing",
		slog.String("query", query),
		slog.String("id", id),
	)

	result, err := r.db.ExecContext(ctx, query, string(id))
	if err != nil {
		slog.Error("StorePostgresRepository.Delete failed",
			slog.String("query", query),
			slog.String("id", id),
			slog.Any("error", err),
		)
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		slog.Error("StorePostgresRepository.Delete failed to get rows affected",
			slog.String("id", id),
			slog.Any("error", err),
		)
		return err
	}
	slog.Debug("StorePostgresRepository.Delete rows affected",
		slog.String("id", id),
		slog.Int64("rows", rows),
	)
	if rows == 0 {
		return domain.ErrStoreNotFound
	}
	return nil
}

func (r *StorePostgresRepository) scanStore(row *sql.Row) (*domain.Store, error) {
	var id, name, description, address, phoneNumber string
	if err := row.Scan(&id, &name, &description, &address, &phoneNumber); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrStoreNotFound
		}
		return nil, err
	}
	return domain.NewStore(string(id), name, description, address, phoneNumber)
}

func (r *StorePostgresRepository) scanStoreFromRows(rows *sql.Rows) (*domain.Store, error) {
	var id, name, description, address, phoneNumber string
	if err := rows.Scan(&id, &name, &description, &address, &phoneNumber); err != nil {
		return nil, err
	}
	return domain.NewStore(string(id), name, description, address, phoneNumber)
}
