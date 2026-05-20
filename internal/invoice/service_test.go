package invoice

import (
	"context"
	"testing"
)

type fakeStore struct {
	created Invoice
}

func (f *fakeStore) Create(ctx context.Context, inv Invoice) (*Invoice, error) {
	f.created = inv
	return &inv, nil
}

func TestServiceCreateInvoice(t *testing.T) {
	t.Run("정상 입력이면 통화를 정규화하고 PENDING invoice를 생성한다", func(t *testing.T) {
		service := NewService(&fakeStore{})

		got, err := service.CreateInvoice(context.Background(), CreateInvoiceRequest{
			MerchantID: " m_123 ",
			Amount:     10000000,
			Currency:   "usdc",
		})
		if err != nil {
			t.Fatalf("CreateInvoice returned error: %v", err)
		}

		if got.MerchantID != "m_123" {
			t.Fatalf("expected trimmed merchant id, got %q", got.MerchantID)
		}
		if got.Currency != "USDC" {
			t.Fatalf("expected normalized currency, got %q", got.Currency)
		}
		if got.Status != StatusPending {
			t.Fatalf("expected status %q, got %q", StatusPending, got.Status)
		}
		if got.ID == "" {
			t.Fatal("expected invoice id to be generated")
		}
	})

	t.Run("amount가 0이면 에러를 반환한다", func(t *testing.T) {
		service := NewService(&fakeStore{})

		_, err := service.CreateInvoice(context.Background(), CreateInvoiceRequest{
			MerchantID: "m_123",
			Amount:     0,
			Currency:   "USDC",
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
