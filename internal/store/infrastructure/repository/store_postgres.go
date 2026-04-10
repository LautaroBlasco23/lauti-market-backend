package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

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
		INSERT INTO stores (id, name, description, address, phone_number, mp_user_id, mp_access_token, mp_refresh_token, mp_token_expires_at, mp_connected_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	slog.Debug("StorePostgresRepository.Save executing",
		slog.String("query", query),
		slog.String("id", string(store.ID())),
		slog.String("name", store.Name()),
	)

	_, err := r.db.ExecContext(ctx, query,
		store.ID(),
		store.Name(),
		store.Description(),
		store.Address(),
		store.PhoneNumber(),
		store.MPUserID(),
		store.MPAccessToken(),
		store.MPRefreshToken(),
		store.MPTokenExpiresAt(),
		store.MPConnectedAt(),
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
		SELECT id, name, description, address, phone_number, mp_user_id, mp_access_token, mp_refresh_token, mp_token_expires_at, mp_connected_at
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
		SELECT id, name, description, address, phone_number, mp_user_id, mp_access_token, mp_refresh_token, mp_token_expires_at, mp_connected_at
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
		SET name = $2, description = $3, address = $4, phone_number = $5,
		    mp_user_id = $6, mp_access_token = $7, mp_refresh_token = $8,
		    mp_token_expires_at = $9, mp_connected_at = $10
		WHERE id = $1
	`
	slog.Debug("StorePostgresRepository.Update executing",
		slog.String("query", query),
		slog.String("id", string(store.ID())),
		slog.String("name", store.Name()),
	)

	result, err := r.db.ExecContext(ctx, query,
		store.ID(),
		store.Name(),
		store.Description(),
		store.Address(),
		store.PhoneNumber(),
		store.MPUserID(),
		store.MPAccessToken(),
		store.MPRefreshToken(),
		store.MPTokenExpiresAt(),
		store.MPConnectedAt(),
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
	var mpUserID, mpAccessToken, mpRefreshToken sql.NullString
	var mpTokenExpiresAt, mpConnectedAt sql.NullTime

	if err := row.Scan(&id, &name, &description, &address, &phoneNumber,
		&mpUserID, &mpAccessToken, &mpRefreshToken, &mpTokenExpiresAt, &mpConnectedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrStoreNotFound
		}
		return nil, err
	}

	return r.buildStore(id, name, description, address, phoneNumber,
		mpUserID, mpAccessToken, mpRefreshToken, mpTokenExpiresAt, mpConnectedAt)
}

func (r *StorePostgresRepository) scanStoreFromRows(rows *sql.Rows) (*domain.Store, error) {
	var id, name, description, address, phoneNumber string
	var mpUserID, mpAccessToken, mpRefreshToken sql.NullString
	var mpTokenExpiresAt, mpConnectedAt sql.NullTime

	if err := rows.Scan(&id, &name, &description, &address, &phoneNumber,
		&mpUserID, &mpAccessToken, &mpRefreshToken, &mpTokenExpiresAt, &mpConnectedAt); err != nil {
		return nil, err
	}

	return r.buildStore(id, name, description, address, phoneNumber,
		mpUserID, mpAccessToken, mpRefreshToken, mpTokenExpiresAt, mpConnectedAt)
}

func (r *StorePostgresRepository) buildStore(
	id, name, description, address, phoneNumber string,
	mpUserID, mpAccessToken, mpRefreshToken sql.NullString,
	mpTokenExpiresAt, mpConnectedAt sql.NullTime,
) (*domain.Store, error) {
	input := domain.CreateStoreInput{
		Name:        name,
		Description: description,
		Address:     address,
		PhoneNumber: phoneNumber,
	}
	store, err := domain.NewStore(id, input)
	if err != nil {
		return nil, err
	}

	// Restore MP fields using reflection or package-private setters
	// Since we need to hydrate from DB, we'll use the HydrateStore function
	return domain.HydrateStore(store, domain.MPFields{
		MPUserID:         mpUserID.String,
		MPAccessToken:    mpAccessToken.String,
		MPRefreshToken:   mpRefreshToken.String,
		MPTokenExpiresAt: nullTimeToPtr(mpTokenExpiresAt),
		MPConnectedAt:    nullTimeToPtr(mpConnectedAt),
	}), nil
}

func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func (r *StorePostgresRepository) UpdateMPConnection(ctx context.Context, storeID string, fields domain.MPFields) error {
	query := `
		UPDATE stores
		SET mp_user_id = $2,
		    mp_access_token = $3,
		    mp_refresh_token = $4,
		    mp_token_expires_at = $5,
		    mp_connected_at = $6
		WHERE id = $1
	`
	slog.Debug("StorePostgresRepository.UpdateMPConnection executing",
		slog.String("store_id", storeID),
		slog.String("mp_user_id", fields.MPUserID),
	)

	result, err := r.db.ExecContext(ctx, query,
		storeID,
		fields.MPUserID,
		fields.MPAccessToken,
		fields.MPRefreshToken,
		fields.MPTokenExpiresAt,
		fields.MPConnectedAt,
	)
	if err != nil {
		slog.Error("StorePostgresRepository.UpdateMPConnection failed",
			slog.String("store_id", storeID),
			slog.Any("error", err),
		)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		slog.Error("StorePostgresRepository.UpdateMPConnection failed to get rows affected",
			slog.String("store_id", storeID),
			slog.Any("error", err),
		)
		return err
	}

	slog.Debug("StorePostgresRepository.UpdateMPConnection rows affected",
		slog.String("store_id", storeID),
		slog.Int64("rows", rows),
	)

	if rows == 0 {
		return domain.ErrStoreNotFound
	}
	return nil
}
