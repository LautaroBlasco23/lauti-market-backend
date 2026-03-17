package domain

type PaymentStatus string

const (
	StatusPending   PaymentStatus = "pending"
	StatusApproved  PaymentStatus = "approved"
	StatusRejected  PaymentStatus = "rejected"
	StatusCancelled PaymentStatus = "cancelled"
	StatusInProcess PaymentStatus = "in_process"
)

func (s PaymentStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusApproved, StatusRejected, StatusCancelled, StatusInProcess:
		return true
	}
	return false
}

func (s PaymentStatus) IsTerminal() bool {
	switch s {
	case StatusApproved, StatusRejected, StatusCancelled:
		return true
	}
	return false
}
