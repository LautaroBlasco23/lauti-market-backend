package repository

import (
	"context"
	"database/sql"
	"log/slog"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/product/domain"
)

type ProductPostgresRepository struct {
	db *sql.DB
}

func NewProductPostgresRepository(db *sql.DB) *ProductPostgresRepository {
	return &ProductPostgresRepository{db: db}
}

func (r *ProductPostgresRepository) Save(ctx context.Context, product *domain.Product) error {
	query := `
		INSERT INTO products (id, store_id, name, description, category, stock, price, image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	slog.Debug("ProductPostgresRepository.Save executing",
		slog.String("query", query),
		slog.String("product_id", product.ID()),
		slog.String("store_id", product.StoreID()),
		slog.String("name", product.Name()),
	)

	_, err := r.db.ExecContext(ctx, query,
		product.ID(),
		product.StoreID(),
		product.Name(),
		product.Description(),
		product.Category(),
		product.Stock(),
		product.Price(),
		product.ImageURL(),
	)
	if err != nil {
		slog.Error("ProductPostgresRepository.Save failed",
			slog.String("query", query),
			slog.String("product_id", product.ID()),
			slog.Any("error", err),
		)
		return err
	}

	return nil
}

func (r *ProductPostgresRepository) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	query := `
		SELECT id, store_id, name, description, category, stock, price, image_url
		FROM products
		WHERE id = $1
	`
	slog.Debug("ProductPostgresRepository.FindByID executing",
		slog.String("query", query),
		slog.String("id", id),
	)

	row := r.db.QueryRowContext(ctx, query, id)
	product, err := r.scanProduct(row)
	if err != nil {
		slog.Error("ProductPostgresRepository.FindByID failed",
			slog.String("query", query),
			slog.String("id", id),
			slog.Any("error", err),
		)
		return nil, err
	}

	return product, nil
}

func (r *ProductPostgresRepository) FindAll(ctx context.Context, limit, offset int, category *string) ([]*domain.Product, error) {
	var (
		rows  *sql.Rows
		err   error
		query string
	)

	if category != nil {
		query = `
			SELECT id, store_id, name, description, category, stock, price, image_url
			FROM products
			WHERE category = $1
			ORDER BY name
			LIMIT $2 OFFSET $3
		`
		slog.Debug("ProductPostgresRepository.FindAll executing with category filter",
			slog.String("query", query),
			slog.String("category", *category),
			slog.Int("limit", limit),
			slog.Int("offset", offset),
		)
		rows, err = r.db.QueryContext(ctx, query, *category, limit, offset)
	} else {
		query = `
			SELECT id, store_id, name, description, category, stock, price, image_url
			FROM products
			ORDER BY name
			LIMIT $1 OFFSET $2
		`
		slog.Debug("ProductPostgresRepository.FindAll executing",
			slog.String("query", query),
			slog.Int("limit", limit),
			slog.Int("offset", offset),
		)
		rows, err = r.db.QueryContext(ctx, query, limit, offset)
	}

	if err != nil {
		slog.Error("ProductPostgresRepository.FindAll query failed",
			slog.String("query", query),
			slog.Any("error", err),
		)
		return nil, err
	}
	defer func() {
		_ = rows.Close() //nolint:errcheck
	}()

	var products []*domain.Product
	for rows.Next() {
		product, err := r.scanProductFromRows(rows)
		if err != nil {
			slog.Error("ProductPostgresRepository.FindAll scan failed",
				slog.Any("error", err),
			)
			return nil, err
		}
		products = append(products, product)
	}

	slog.Debug("ProductPostgresRepository.FindAll completed",
		slog.Int("row_count", len(products)),
	)

	return products, rows.Err()
}

func (r *ProductPostgresRepository) FindByStoreID(ctx context.Context, storeID string, limit, offset int) ([]*domain.Product, error) {
	query := `
		SELECT id, store_id, name, description, category, stock, price, image_url
		FROM products
		WHERE store_id = $1
		ORDER BY name
		LIMIT $2 OFFSET $3
	`
	slog.Debug("ProductPostgresRepository.FindByStoreID executing",
		slog.String("query", query),
		slog.String("store_id", storeID),
		slog.Int("limit", limit),
		slog.Int("offset", offset),
	)

	rows, err := r.db.QueryContext(ctx, query, storeID, limit, offset)
	if err != nil {
		slog.Error("ProductPostgresRepository.FindByStoreID query failed",
			slog.String("query", query),
			slog.String("store_id", storeID),
			slog.Any("error", err),
		)
		return nil, err
	}
	defer func() {
		_ = rows.Close() //nolint:errcheck
	}()

	var products []*domain.Product
	for rows.Next() {
		product, err := r.scanProductFromRows(rows)
		if err != nil {
			slog.Error("ProductPostgresRepository.FindByStoreID scan failed",
				slog.String("store_id", storeID),
				slog.Any("error", err),
			)
			return nil, err
		}
		products = append(products, product)
	}

	slog.Debug("ProductPostgresRepository.FindByStoreID completed",
		slog.String("store_id", storeID),
		slog.Int("row_count", len(products)),
	)

	return products, rows.Err()
}

func (r *ProductPostgresRepository) Update(ctx context.Context, product *domain.Product) error {
	query := `
		UPDATE products
		SET name = $2, description = $3, category = $4, stock = $5, price = $6, image_url = $7
		WHERE id = $1
	`
	slog.Debug("ProductPostgresRepository.Update executing",
		slog.String("query", query),
		slog.String("product_id", product.ID()),
		slog.String("name", product.Name()),
	)

	result, err := r.db.ExecContext(ctx, query,
		product.ID(),
		product.Name(),
		product.Description(),
		product.Category(),
		product.Stock(),
		product.Price(),
		product.ImageURL(),
	)
	if err != nil {
		slog.Error("ProductPostgresRepository.Update exec failed",
			slog.String("query", query),
			slog.String("product_id", product.ID()),
			slog.Any("error", err),
		)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		slog.Error("ProductPostgresRepository.Update rows affected failed",
			slog.String("product_id", product.ID()),
			slog.Any("error", err),
		)
		return err
	}
	if rows == 0 {
		slog.Error("ProductPostgresRepository.Update product not found",
			slog.String("product_id", product.ID()),
		)
		return apiDomain.ErrProductNotFound
	}

	slog.Debug("ProductPostgresRepository.Update completed",
		slog.String("product_id", product.ID()),
		slog.Int64("rows_affected", rows),
	)

	return nil
}

func (r *ProductPostgresRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`
	slog.Debug("ProductPostgresRepository.Delete executing",
		slog.String("query", query),
		slog.String("id", id),
	)

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		slog.Error("ProductPostgresRepository.Delete exec failed",
			slog.String("query", query),
			slog.String("id", id),
			slog.Any("error", err),
		)
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		slog.Error("ProductPostgresRepository.Delete rows affected failed",
			slog.String("id", id),
			slog.Any("error", err),
		)
		return err
	}
	if rows == 0 {
		slog.Error("ProductPostgresRepository.Delete product not found",
			slog.String("id", id),
		)
		return apiDomain.ErrProductNotFound
	}

	slog.Debug("ProductPostgresRepository.Delete completed",
		slog.String("id", id),
		slog.Int64("rows_affected", rows),
	)

	return nil
}

func (r *ProductPostgresRepository) scanProduct(row *sql.Row) (*domain.Product, error) {
	var id, storeID, name, description, category string
	var stock int
	var price float64
	var imageURL *string

	if err := row.Scan(&id, &storeID, &name, &description, &category, &stock, &price, &imageURL); err != nil {
		if err == sql.ErrNoRows {
			return nil, apiDomain.ErrProductNotFound
		}
		return nil, err
	}

	return domain.NewProduct(id, storeID, name, description, category, stock, price, imageURL)
}

func (r *ProductPostgresRepository) scanProductFromRows(rows *sql.Rows) (*domain.Product, error) {
	var id, storeID, name, description, category string
	var stock int
	var price float64
	var imageURL *string

	if err := rows.Scan(&id, &storeID, &name, &description, &category, &stock, &price, &imageURL); err != nil {
		return nil, err
	}

	return domain.NewProduct(id, storeID, name, description, category, stock, price, imageURL)
}
