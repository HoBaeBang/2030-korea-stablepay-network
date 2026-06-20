package ledger

import (
	"context"
	"errors"
	"testing"
)

func newTestService(t *testing.T) (*Service, context.Context) {
	t.Helper()

	return NewService(nil), context.Background()
}

func newTestServiceWithStore(t *testing.T, store Store) (*Service, context.Context) {
	t.Helper()

	return NewService(store), context.Background()
}

type fakeStore struct {
	calls   int
	tx      Transaction
	entries []Entry
	err     error
}

func (f *fakeStore) CreateTransaction(ctx context.Context, tx Transaction, entries []Entry) error {
	f.calls++
	f.tx = tx
	f.entries = append([]Entry(nil), entries...)

	return f.err
}

func TestServiceValidateTransaction(t *testing.T) {
	t.Run("debit과 credit 합계가 같으면 성공한다", func(t *testing.T) {
		svc, ctx := newTestService(t)

		entries := []Entry{
			{
				AccountID: "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
			{
				AccountID: "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    9_800_000,
				Currency:  "USDC",
			},
			{
				AccountID: "acct_platform_fee_1",
				Direction: EntryDirectionCredit,
				Amount:    200_000,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(ctx, entries); err != nil {
			t.Fatalf("원장 거래의 균형이 맞아야 하는데 에러가 발생했습니다: %v", err)
		}
	})

	t.Run("credit 합계가 부족하면 실패한다", func(t *testing.T) {
		svc, ctx := newTestService(t)

		entries := []Entry{
			{
				AccountID: "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
			{
				AccountID: "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    9_000_000,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(ctx, entries); err == nil {
			t.Fatal("원장 거래의 균형이 맞지 않아야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("entry가 하나뿐이면 실패한다", func(t *testing.T) {
		svc, ctx := newTestService(t)

		entries := []Entry{
			{
				AccountID: "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(ctx, entries); err == nil {
			t.Fatal("에러가 발생해야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("통화가 비어 있으면 실패한다", func(t *testing.T) {
		svc, ctx := newTestService(t)

		entries := []Entry{
			{
				AccountID: "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    10_000_000,
				Currency:  "",
			},
			{
				AccountID: "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    10_000_000,
				Currency:  "",
			},
		}

		if err := svc.ValidateTransaction(ctx, entries); err == nil {
			t.Fatal("통화가 비어 있으면 실패해야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("금액이 0이면 실패한다", func(t *testing.T) {
		svc, ctx := newTestService(t)

		entries := []Entry{
			{
				AccountID: "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    0,
				Currency:  "USDC",
			},
			{
				AccountID: "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    0,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(ctx, entries); err == nil {
			t.Fatal("금액이 0인 원장 항목은 실패해야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("알 수 없는 방향이면 실패한다", func(t *testing.T) {
		svc, ctx := newTestService(t)

		entries := []Entry{
			{
				AccountID: "acct_customer_1",
				Direction: EntryDirection("UNKNOWN"),
				Amount:    10_000_000,
				Currency:  "USDC",
			},
			{
				AccountID: "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(ctx, entries); err == nil {
			t.Fatal("알 수 없는 방향은 실패해야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("context가 취소되었으면 실패한다", func(t *testing.T) {
		svc, _ := newTestService(t)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		entries := []Entry{
			{
				AccountID: "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
			{
				AccountID: "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(ctx, entries); err == nil {
			t.Fatal("context가 취소되었으면 실패해야 하는데 nil이 반환되었습니다")
		}
	})
}

func TestServiceRecordTransaction(t *testing.T) {
	t.Run("검증에 성공하면 저장소를 호출한다", func(t *testing.T) {
		store := &fakeStore{}
		svc, ctx := newTestServiceWithStore(t, store)

		tx := Transaction{
			ID:             "led_tx_service_1",
			ReferenceType:  "PAYMENT",
			ReferenceID:    "pay_service_1",
			IdempotencyKey: "payment:pay_service_1:finalized",
		}
		entries := []Entry{
			{
				ID:            "led_entry_service_1",
				TransactionID: tx.ID,
				AccountID:     "acct_customer_1",
				Direction:     EntryDirectionDebit,
				Amount:        10_000_000,
				Currency:      "USDC",
			},
			{
				ID:            "led_entry_service_2",
				TransactionID: tx.ID,
				AccountID:     "acct_merchant_pending_1",
				Direction:     EntryDirectionCredit,
				Amount:        10_000_000,
				Currency:      "USDC",
			},
		}

		if err := svc.RecordTransaction(ctx, tx, entries); err != nil {
			t.Fatalf("검증에 성공한 원장 거래는 저장되어야 합니다: %v", err)
		}

		if store.calls != 1 {
			t.Fatalf("저장소는 1번 호출되어야 하는데 %d번 호출되었습니다", store.calls)
		}

		if store.tx.ID != tx.ID {
			t.Fatalf("저장소에 전달된 transaction id가 다릅니다: %s", store.tx.ID)
		}

		if len(store.entries) != len(entries) {
			t.Fatalf("저장소에 전달된 entries 개수가 다릅니다: %d", len(store.entries))
		}
	})

	t.Run("검증에 실패하면 저장소를 호출하지 않는다", func(t *testing.T) {
		store := &fakeStore{}
		svc, ctx := newTestServiceWithStore(t, store)

		tx := Transaction{
			ID:             "led_tx_service_invalid",
			ReferenceType:  "PAYMENT",
			ReferenceID:    "pay_service_invalid",
			IdempotencyKey: "payment:pay_service_invalid:finalized",
		}
		entries := []Entry{
			{
				ID:            "led_entry_service_invalid_1",
				TransactionID: tx.ID,
				AccountID:     "acct_customer_1",
				Direction:     EntryDirectionDebit,
				Amount:        10_000_000,
				Currency:      "USDC",
			},
			{
				ID:            "led_entry_service_invalid_2",
				TransactionID: tx.ID,
				AccountID:     "acct_merchant_pending_1",
				Direction:     EntryDirectionCredit,
				Amount:        9_000_000,
				Currency:      "USDC",
			},
		}

		if err := svc.RecordTransaction(ctx, tx, entries); err == nil {
			t.Fatal("균형이 맞지 않는 원장 거래는 실패해야 하는데 nil이 반환되었습니다")
		}

		if store.calls != 0 {
			t.Fatalf("검증 실패 시 저장소는 호출되면 안 되는데 %d번 호출되었습니다", store.calls)
		}
	})

	t.Run("저장소 에러를 반환한다", func(t *testing.T) {
		storeErr := errors.New("저장소 실패")
		store := &fakeStore{err: storeErr}
		svc, ctx := newTestServiceWithStore(t, store)

		tx := Transaction{
			ID:             "led_tx_service_store_error",
			ReferenceType:  "PAYMENT",
			ReferenceID:    "pay_service_store_error",
			IdempotencyKey: "payment:pay_service_store_error:finalized",
		}
		entries := []Entry{
			{
				ID:            "led_entry_service_store_error_1",
				TransactionID: tx.ID,
				AccountID:     "acct_customer_1",
				Direction:     EntryDirectionDebit,
				Amount:        10_000_000,
				Currency:      "USDC",
			},
			{
				ID:            "led_entry_service_store_error_2",
				TransactionID: tx.ID,
				AccountID:     "acct_merchant_pending_1",
				Direction:     EntryDirectionCredit,
				Amount:        10_000_000,
				Currency:      "USDC",
			},
		}

		err := svc.RecordTransaction(ctx, tx, entries)
		if !errors.Is(err, storeErr) {
			t.Fatalf("저장소 에러를 그대로 반환해야 합니다: %v", err)
		}

		if store.calls != 1 {
			t.Fatalf("저장소는 1번 호출되어야 하는데 %d번 호출되었습니다", store.calls)
		}
	})
}
