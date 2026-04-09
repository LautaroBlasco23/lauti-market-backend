package domain

import "time"

type Payment struct {
	id            string
	orderID       string
	userID        string
	mpPaymentID   int64
	amount        float64
	currency      string
	status        PaymentStatus
	statusDetail  string
	paymentMethod string
	preferenceID  string
	createdAt     time.Time
	updatedAt     time.Time
}

func NewPayment(id, orderID, userID, preferenceID string, amount float64) *Payment {
	now := time.Now()
	return &Payment{
		id:           id,
		orderID:      orderID,
		userID:       userID,
		preferenceID: preferenceID,
		amount:       amount,
		currency:     "ARS",
		status:       StatusPending,
		createdAt:    now,
		updatedAt:    now,
	}
}

func NewPaymentFromDB(
	id, orderID, userID string,
	mpPaymentID int64,
	amount float64,
	currency string,
	status PaymentStatus,
	statusDetail, paymentMethod, preferenceID string,
	createdAt, updatedAt time.Time,
) *Payment {
	return &Payment{
		id:            id,
		orderID:       orderID,
		userID:        userID,
		mpPaymentID:   mpPaymentID,
		amount:        amount,
		currency:      currency,
		status:        status,
		statusDetail:  statusDetail,
		paymentMethod: paymentMethod,
		preferenceID:  preferenceID,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

func (p *Payment) UpdateFromMP(mpPaymentID int64, status PaymentStatus, statusDetail, paymentMethod string) {
	p.mpPaymentID = mpPaymentID
	p.status = status
	p.statusDetail = statusDetail
	p.paymentMethod = paymentMethod
	p.updatedAt = time.Now()
}

func (p *Payment) ID() string            { return p.id }
func (p *Payment) OrderID() string       { return p.orderID }
func (p *Payment) UserID() string        { return p.userID }
func (p *Payment) MPPaymentID() int64    { return p.mpPaymentID }
func (p *Payment) Amount() float64       { return p.amount }
func (p *Payment) Currency() string      { return p.currency }
func (p *Payment) Status() PaymentStatus { return p.status }
func (p *Payment) StatusDetail() string  { return p.statusDetail }
func (p *Payment) PaymentMethod() string { return p.paymentMethod }
func (p *Payment) PreferenceID() string  { return p.preferenceID }
func (p *Payment) CreatedAt() time.Time  { return p.createdAt }
func (p *Payment) UpdatedAt() time.Time  { return p.updatedAt }
