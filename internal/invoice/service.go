package invoice

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Store interface {
	Create(ctx context.Context, inv Invoice) (*Invoice, error)
}

type Service struct {
	store Store
	now   func() time.Time
}

func NewService(store Store) *Service {
	return &Service{store: store, now: time.Now}
}

func (s *Service) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*Invoice, error) {
	merchantID := strings.TrimSpace(req.MerchantID)
	currency := strings.ToUpper(strings.TrimSpace(req.Currency))

	if merchantID == "" {
		return nil, fmt.Errorf("merchant id is required")
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}
	if currency == "" {
		return nil, fmt.Errorf("currency is required")
	}
	if currency != "USDC" {
		return nil, fmt.Errorf("currency is not supported")
	}

	inv := Invoice{
		ID:         newInvoiceID(s.now()),
		MerchantID: merchantID,
		Amount:     req.Amount,
		Currency:   currency,
		Status:     StatusPending,
		ExpiresAt:  req.ExpiresAt,
	}

	saved, err := s.store.Create(ctx, inv)
	if err != nil {
		return nil, fmt.Errorf("create invoice: %w", err)
	}

	return saved, nil
}

func newInvoiceID(now time.Time) string {
	return fmt.Sprintf("inv_%d", now.UnixNano())
}
