package repository

import (
	"context"
	"database/sql"
	"time"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/order/domain"
)

type OrderPostgresRepository struct {
	db *sql.DB
}

func NewOrderPostgresRepository(db *sql.DB) *OrderPostgresRepository {
	return &OrderPostgresRepository{db: db}
}

func (r *OrderPostgresRepository) Save(ctx context.Context, order *domain.Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback() //nolint:errcheck
	}()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders (id, user_id, store_id, status, total_price)
		VALUES ($1, $2, $3, $4, $5)
	`, order.ID(), order.UserID(), order.StoreID(), string(order.Status()), order.TotalPrice())
	if err != nil {
		return err
	}

	for _, item := range order.Items() {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO order_items (id, order_id, product_id, quantity, unit_price)
			VALUES ($1, $2, $3, $4, $5)
		`, item.ID(), item.OrderID(), item.ProductID(), item.Quantity(), item.UnitPrice())
		if err != nil {
			return err
		}

		result, err := tx.ExecContext(ctx, `
			UPDATE products SET stock = stock - $1 WHERE id = $2 AND stock >= $1
		`, item.Quantity(), item.ProductID())
		if err != nil {
			return err
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return apiDomain.ErrInsufficientStock
		}
	}

	return tx.Commit()
}

func (r *OrderPostgresRepository) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	orderRow := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, store_id, status, total_price, created_at, updated_at
		FROM orders WHERE id = $1
	`, id)

	var orderID, userID, storeID, status string
	var totalPrice float64
	var createdAt, updatedAt time.Time

	if err := orderRow.Scan(&orderID, &userID, &storeID, &status, &totalPrice, &createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, apiDomain.ErrOrderNotFound
		}
		return nil, err
	}

	items, err := r.findItemsByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	return domain.NewOrderFromDB(orderID, userID, storeID, domain.OrderStatus(status), items, totalPrice, createdAt, updatedAt), nil
}

func (r *OrderPostgresRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*domain.Order, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, store_id, status, total_price, created_at, updated_at
		FROM orders WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close() //nolint:errcheck
	}()

	return r.scanOrders(ctx, rows)
}

func (r *OrderPostgresRepository) FindByStoreID(ctx context.Context, storeID string, limit, offset int) ([]*domain.Order, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, store_id, status, total_price, created_at, updated_at
		FROM orders WHERE store_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, storeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close() //nolint:errcheck
	}()

	return r.scanOrders(ctx, rows)
}

func (r *OrderPostgresRepository) UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2
	`, string(status), id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return apiDomain.ErrOrderNotFound
	}

	return nil
}

func (r *OrderPostgresRepository) findItemsByOrderID(ctx context.Context, orderID string) ([]*domain.OrderItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, order_id, product_id, quantity, unit_price
		FROM order_items WHERE order_id = $1
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close() //nolint:errcheck
	}()

	var items []*domain.OrderItem
	for rows.Next() {
		var id, oid, productID string
		var quantity int
		var unitPrice float64

		if err := rows.Scan(&id, &oid, &productID, &quantity, &unitPrice); err != nil {
			return nil, err
		}

		item, err := domain.NewOrderItem(id, oid, productID, quantity, unitPrice)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *OrderPostgresRepository) scanOrders(ctx context.Context, rows *sql.Rows) ([]*domain.Order, error) {
	var orders []*domain.Order

	for rows.Next() {
		var orderID, userID, storeID, status string
		var totalPrice float64
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&orderID, &userID, &storeID, &status, &totalPrice, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		items, err := r.findItemsByOrderID(ctx, orderID)
		if err != nil {
			return nil, err
		}

		orders = append(orders, domain.NewOrderFromDB(orderID, userID, storeID, domain.OrderStatus(status), items, totalPrice, createdAt, updatedAt))
	}

	return orders, rows.Err()
}
