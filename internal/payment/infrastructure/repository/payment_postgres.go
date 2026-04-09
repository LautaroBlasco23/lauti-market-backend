package repository

import (
	"context"
	"database/sql"
	"time"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	"github.com/LautaroBlasco23/lauti-market-backend/internal/payment/domain"
)

type PaymentPostgresRepository struct {
	db *sql.DB
}

func NewPaymentPostgresRepository(db *sql.DB) *PaymentPostgresRepository {
	return &PaymentPostgresRepository{db: db}
}

func (r *PaymentPostgresRepository) Save(ctx context.Context, p *domain.Payment) error {
	query := `
		INSERT INTO payments (id, order_id, user_id, mp_payment_id, amount, currency, status, status_detail, payment_method, idempotency_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.ExecContext(ctx, query,
		p.ID(),
		p.OrderID(),
		p.UserID(),
		p.MPPaymentID(),
		p.Amount(),
		p.Currency(),
		string(p.Status()),
		p.StatusDetail(),
		p.PaymentMethod(),
		p.IdempotencyKey(),
		p.CreatedAt(),
		p.UpdatedAt(),
	)
	return err
}

func (r *PaymentPostgresRepository) FindByID(ctx context.Context, id string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, user_id, mp_payment_id, amount, currency, status, status_detail, payment_method, idempotency_key, created_at, updated_at
		FROM payments
		WHERE id = $1
	`
	return r.scanPayment(r.db.QueryRowContext(ctx, query, id))
}

func (r *PaymentPostgresRepository) FindByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, user_id, mp_payment_id, amount, currency, status, status_detail, payment_method, idempotency_key, created_at, updated_at
		FROM payments
		WHERE order_id = $1
	`
	return r.scanPayment(r.db.QueryRowContext(ctx, query, orderID))
}

func (r *PaymentPostgresRepository) FindByMPPaymentID(ctx context.Context, mpPaymentID int64) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, user_id, mp_payment_id, amount, currency, status, status_detail, payment_method, idempotency_key, created_at, updated_at
		FROM payments
		WHERE mp_payment_id = $1
	`
	return r.scanPayment(r.db.QueryRowContext(ctx, query, mpPaymentID))
}

func (r *PaymentPostgresRepository) UpdateFromMP(ctx context.Context, p *domain.Payment) error {
	query := `
		UPDATE payments
		SET status = $2, status_detail = $3, payment_method = $4, mp_payment_id = $5, updated_at = $6
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query,
		p.ID(),
		string(p.Status()),
		p.StatusDetail(),
		p.PaymentMethod(),
		p.MPPaymentID(),
		p.UpdatedAt(),
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return apiDomain.ErrPaymentNotFound
	}
	return nil
}

func (r *PaymentPostgresRepository) scanPayment(row *sql.Row) (*domain.Payment, error) {
	var (
		id, orderID, userID  string
		mpPaymentID          int64
		amount               float64
		currency, status     string
		statusDetail         sql.NullString
		paymentMethod        sql.NullString
		idempotencyKey       string
		createdAt, updatedAt time.Time
	)

	err := row.Scan(
		&id,
		&orderID,
		&userID,
		&mpPaymentID,
		&amount,
		&currency,
		&status,
		&statusDetail,
		&paymentMethod,
		&idempotencyKey,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apiDomain.ErrPaymentNotFound
		}
		return nil, err
	}

	var statusDetailStr, paymentMethodStr string
	if statusDetail.Valid {
		statusDetailStr = statusDetail.String
	}
	if paymentMethod.Valid {
		paymentMethodStr = paymentMethod.String
	}

	return domain.NewPaymentFromDB(
		id, orderID, userID,
		mpPaymentID,
		amount,
		currency,
		domain.PaymentStatus(status),
		statusDetailStr, paymentMethodStr, idempotencyKey,
		createdAt, updatedAt,
	), nil
}
