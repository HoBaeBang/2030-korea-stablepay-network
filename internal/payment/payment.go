package payment

import "time"

type Status string

const (
	StatusPending         Status = "PENDING"
	StatusOnchainDetected Status = "ONCHAIN_DETECTED"
	StatusFinalized       Status = "FINALIZED"
	StatusSettled         Status = "SETTLED"
	StatusFailed          Status = "FAILED"
)

type Payment struct {
	ID              string
	InvoiceID       string
	Amount          int64
	Currency        string
	Status          Status
	TransactionHash *string
	FinalizedAt     *time.Time
	CreatedAt       time.Time
}

type CreatePaymentRequest struct {
	InvoiceID string
	Amount    int64
	Currency  string
}

type UpdatePaymentStatusRequest struct {
	PaymentID       string
	NextStatus      Status
	TransactionHash *string
}
