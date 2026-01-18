package repository

import (
	"context"
	"database/sql"

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
		INSERT INTO products (id, store_id, name, description, stock, price, image_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(ctx, query,
		product.ID(),
		product.StoreID(),
		product.Name(),
		product.Description(),
		product.Stock(),
		product.Price(),
		product.ImageURL(),
	)
	return err
}

func (r *ProductPostgresRepository) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	query := `
		SELECT id, store_id, name, description, stock, price, image_url
		FROM products
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanProduct(row)
}

func (r *ProductPostgresRepository) FindByStoreID(ctx context.Context, storeID string, limit, offset int) ([]*domain.Product, error) {
	query := `
		SELECT id, store_id, name, description, stock, price, image_url
		FROM products
		WHERE store_id = $1
		ORDER BY name
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, storeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		product, err := r.scanProductFromRows(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, rows.Err()
}

func (r *ProductPostgresRepository) Update(ctx context.Context, product *domain.Product) error {
	query := `
		UPDATE products
		SET name = $2, description = $3, stock = $4, price = $5, image_url = $6
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query,
		product.ID(),
		product.Name(),
		product.Description(),
		product.Stock(),
		product.Price(),
		product.ImageURL(),
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return apiDomain.ErrProductNotFound
	}

	return nil
}

func (r *ProductPostgresRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM products WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return apiDomain.ErrProductNotFound
	}

	return nil
}

func (r *ProductPostgresRepository) scanProduct(row *sql.Row) (*domain.Product, error) {
	var id, storeID, name, description string
	var stock int
	var price float64
	var imageURL *string

	if err := row.Scan(&id, &storeID, &name, &description, &stock, &price, &imageURL); err != nil {
		if err == sql.ErrNoRows {
			return nil, apiDomain.ErrProductNotFound
		}
		return nil, err
	}

	return domain.NewProduct(id, storeID, name, description, stock, price, imageURL)
}

func (r *ProductPostgresRepository) scanProductFromRows(rows *sql.Rows) (*domain.Product, error) {
	var id, storeID, name, description string
	var stock int
	var price float64
	var imageURL *string

	if err := rows.Scan(&id, &storeID, &name, &description, &stock, &price, &imageURL); err != nil {
		return nil, err
	}

	return domain.NewProduct(id, storeID, name, description, stock, price, imageURL)
}
