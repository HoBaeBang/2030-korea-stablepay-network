package settlement

import (
	"context"
	"testing"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"
)

func newTestCalculator(t *testing.T) (*Calculator, context.Context) {
	t.Helper()

	return NewCalculator(), context.Background()
}

func validCandidates() []Candidate {
	return []Candidate{
		{
			LedgerEntryID: "led_entry_payment_1",
			RecipientID:   "merchant_1",
			AccountType:   ledger.AccountTypeMerchantPending,
			Direction:     ledger.EntryDirectionCredit,
			Amount:        9_800_000,
			Currency:      "USDC",
		},
		{
			LedgerEntryID: "led_entry_payment_2",
			RecipientID:   "merchant_1",
			AccountType:   ledger.AccountTypeMerchantPending,
			Direction:     ledger.EntryDirectionCredit,
			Amount:        5_000_000,
			Currency:      "USDC",
		},
	}
}

func TestCalculatorBuildBatch(t *testing.T) {
	t.Run("같은 수취인과 통화의 후보를 정산 묶음으로 계산한다", func(t *testing.T) {
		calculator, ctx := newTestCalculator(t)

		batch, items, err := calculator.BuildBatch(
			ctx,
			"stl_batch_1",
			"merchant_1",
			"USDC",
			validCandidates(),
		)
		if err != nil {
			t.Fatalf("정상 정산 후보 계산에 실패했습니다: %v", err)
		}

		if batch.TotalAmount != 14_800_000 {
			t.Fatalf("정산 총액은 14_800_000이어야 하는데 %d입니다", batch.TotalAmount)
		}

		if batch.ItemCount != 2 || len(items) != 2 {
			t.Fatalf("정산 항목은 2개여야 합니다: batch=%d, items=%d", batch.ItemCount, len(items))
		}

		if batch.Status != StatusDraft {
			t.Fatalf("새 정산 묶음 상태는 DRAFT여야 하는데 %s입니다", batch.Status)
		}
	})

	t.Run("merchant pending 계정이 아니면 실패한다", func(t *testing.T) {
		calculator, ctx := newTestCalculator(t)
		candidates := validCandidates()
		candidates[0].AccountType = ledger.AccountTypeCustomer

		_, _, err := calculator.BuildBatch(
			ctx,
			"stl_batch_invalid_account",
			"merchant_1",
			"USDC",
			candidates,
		)
		if err == nil {
			t.Fatal("merchant pending 계정이 아니면 실패해야 합니다")
		}
	})

	t.Run("credit 원장 항목이 아니면 실패한다", func(t *testing.T) {
		calculator, ctx := newTestCalculator(t)
		candidates := validCandidates()
		candidates[0].Direction = ledger.EntryDirectionDebit

		_, _, err := calculator.BuildBatch(
			ctx,
			"stl_batch_invalid_direction",
			"merchant_1",
			"USDC",
			candidates,
		)
		if err == nil {
			t.Fatal("credit 원장 항목이 아니면 실패해야 합니다")
		}
	})

	t.Run("같은 ledger entry가 두 번 포함되면 실패한다", func(t *testing.T) {
		calculator, ctx := newTestCalculator(t)
		candidates := validCandidates()
		candidates[1].LedgerEntryID = candidates[0].LedgerEntryID

		_, _, err := calculator.BuildBatch(
			ctx,
			"stl_batch_duplicate",
			"merchant_1",
			"USDC",
			candidates,
		)
		if err == nil {
			t.Fatal("같은 ledger entry가 두 번 포함되면 실패해야 합니다")
		}
	})

	t.Run("다른 수취인의 후보가 섞이면 실패한다", func(t *testing.T) {
		calculator, ctx := newTestCalculator(t)
		candidates := validCandidates()
		candidates[1].RecipientID = "merchant_2"

		_, _, err := calculator.BuildBatch(
			ctx,
			"stl_batch_mixed_recipient",
			"merchant_1",
			"USDC",
			candidates,
		)
		if err == nil {
			t.Fatal("다른 수취인의 후보가 섞이면 실패해야 합니다")
		}
	})
}
