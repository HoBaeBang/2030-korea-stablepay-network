package payment

import (
	"context"
	"testing"
	"time"
)

type fakeStore struct {
	created Payment
	current Payment
	updated StatusUpdate
}

func (f *fakeStore) Create(ctx context.Context, p Payment) (*Payment, error) {
	f.created = p
	return &p, nil
}

func (f *fakeStore) FindByID(ctx context.Context, id string) (*Payment, error) {
	return &f.current, nil
}

func (f *fakeStore) UpdateStatus(ctx context.Context, update StatusUpdate) (*Payment, error) {
	f.updated = update
	f.current.Status = update.NextStatus
	f.current.TransactionHash = update.TransactionHash
	f.current.FinalizedAt = update.FinalizedAt
	return &f.current, nil
}

func TestService_CreatePayment(t *testing.T) {
	fixedNow := time.Date(2026, 5, 21, 9, 0, 0, 0, time.UTC)
	store := &fakeStore{}
	service := NewService(store)
	service.now = func() time.Time { return fixedNow }

	t.Run("정상 요청이면 PENDING 상태의 payment를 생성한다", func(t *testing.T) {
		got, err := service.CreatePayment(context.Background(), CreatePaymentRequest{
			InvoiceID: "inv_123",
			Amount:    10000,
			Currency:  "usdc",
		})
		if err != nil {
			t.Fatalf("CreatePayment returned error: %v", err)
		}
		if got.Status != StatusPending {
			t.Fatalf("status = %s, want %s", got.Status, StatusPending)
		}
		if got.Currency != "USDC" {
			t.Fatalf("currency = %s, want USDC", got.Currency)
		}
	})

	t.Run("amount가 0 이하이면 에러를 반환한다", func(t *testing.T) {
		_, err := service.CreatePayment(context.Background(), CreatePaymentRequest{
			InvoiceID: "inv_123",
			Amount:    0,
			Currency:  "USDC",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestService_UpdatePaymentStatus(t *testing.T) {
	fixedNow := time.Date(2026, 5, 21, 9, 0, 0, 0, time.UTC)

	t.Run("PENDING에서 ONCHAIN_DETECTED로 변경할 수 있다", func(t *testing.T) {
		store := &fakeStore{current: Payment{ID: "pay_123", Status: StatusPending}}
		service := NewService(store)
		service.now = func() time.Time { return fixedNow }

		txHash := "0xabc"
		got, err := service.UpdatePaymentStatus(context.Background(), UpdatePaymentStatusRequest{
			PaymentID:       "pay_123",
			NextStatus:      StatusOnchainDetected,
			TransactionHash: &txHash,
		})
		if err != nil {
			t.Fatalf("UpdatePaymentStatus returned error: %v", err)
		}
		if got.Status != StatusOnchainDetected {
			t.Fatalf("status = %s, want %s", got.Status, StatusOnchainDetected)
		}
		if got.TransactionHash == nil || *got.TransactionHash != txHash {
			t.Fatalf("transaction hash was not saved")
		}
	})

	t.Run("FINALIZED에서 PENDING으로 되돌릴 수 없다", func(t *testing.T) {
		store := &fakeStore{current: Payment{ID: "pay_123", Status: StatusFinalized}}
		service := NewService(store)
		service.now = func() time.Time { return fixedNow }

		_, err := service.UpdatePaymentStatus(context.Background(), UpdatePaymentStatusRequest{
			PaymentID:  "pay_123",
			NextStatus: StatusPending,
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("FINALIZED가 되면 finalized_at을 저장한다", func(t *testing.T) {
		store := &fakeStore{current: Payment{ID: "pay_123", Status: StatusOnchainDetected}}
		service := NewService(store)
		service.now = func() time.Time { return fixedNow }

		got, err := service.UpdatePaymentStatus(context.Background(), UpdatePaymentStatusRequest{
			PaymentID:  "pay_123",
			NextStatus: StatusFinalized,
		})
		if err != nil {
			t.Fatalf("UpdatePaymentStatus returned error: %v", err)
		}
		if got.FinalizedAt == nil {
			t.Fatal("finalized_at is nil")
		}
		if !got.FinalizedAt.Equal(fixedNow) {
			t.Fatalf("finalized_at = %v, want %v", got.FinalizedAt, fixedNow)
		}
	})
}

func TestCanTransition(t *testing.T) {
	tests := []struct {
		name string
		from Status
		to   Status
		want bool
	}{
		{name: "PENDING에서 ONCHAIN_DETECTED는 허용된다", from: StatusPending, to: StatusOnchainDetected, want: true},
		{name: "ONCHAIN_DETECTED에서 FINALIZED는 허용된다", from: StatusOnchainDetected, to: StatusFinalized, want: true},
		{name: "FINALIZED에서 SETTLED는 허용된다", from: StatusFinalized, to: StatusSettled, want: true},
		{name: "FINALIZED에서 PENDING은 차단된다", from: StatusFinalized, to: StatusPending, want: false},
		{name: "SETTLED에서 FAILED는 차단된다", from: StatusSettled, to: StatusFailed, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanTransition(tt.from, tt.to)
			if got != tt.want {
				t.Fatalf("CanTransition(%s, %s) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}
