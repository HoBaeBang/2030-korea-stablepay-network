package ledger

import (
	"context"
	"testing"
)

func newTestService(t *testing.T) (*Service, context.Context) {
	t.Helper()

	return NewService(), context.Background()
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

	t.Run("entry가 하나뿐이면 실패한다.", func(t *testing.T) {
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
			t.Fatal("에러가 발생해야 하는데 nil이 반환되었습니다.")
		}
	})

	t.Run("통화가 비어 있으면 실패한다.", func(t *testing.T) {
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
