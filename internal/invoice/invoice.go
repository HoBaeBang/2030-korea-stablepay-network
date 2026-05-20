package invoice

import "time"

const StatusPending = "PENDING"

type Invoice struct {
	ID         string
	MerchantID string
	Amount     int64
	Currency   string
	Status     string
	ExpiresAt  *time.Time
	CreatedAt  time.Time
}

type CreateInvoiceRequest struct {
	MerchantID string
	Amount     int64
	Currency   string
	ExpiresAt  *time.Time
}
