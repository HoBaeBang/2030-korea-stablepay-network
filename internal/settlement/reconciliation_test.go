package settlement

import (
	"context"
	"errors"
	"testing"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"
)

type fakeReconciliationStore struct {
	snapshot *ReconciliationSnapshot
	err      error
	calls    int
	batchID  string
}

func (f *fakeReconciliationStore) FindReconciliationSnapshot(
	ctx context.Context,
	batchID string,
) (*ReconciliationSnapshot, error) {
	f.calls++
	f.batchID = batchID
	return f.snapshot, f.err
}

func healthyReconciliationSnapshot() *ReconciliationSnapshot {
	return &ReconciliationSnapshot{
		Batch: Batch{
			ID:          "stl_batch_reconciliation_1",
			RecipientID: "merchant_1",
			Currency:    "USDC",
			TotalAmount: 14_800_000,
			ItemCount:   2,
			Status:      StatusReady,
		},
		Items: []ReconciliationItem{
			{
				LedgerEntryID:      "led_entry_1",
				SettlementAmount:   9_800_000,
				SettlementCurrency: "USDC",
				LedgerEntryFound:   true,
				LedgerAmount:       9_800_000,
				LedgerCurrency:     "USDC",
				LedgerRecipientID:  "merchant_1",
				LedgerAccountType:  ledger.AccountTypeMerchantPending,
				LedgerDirection:    ledger.EntryDirectionCredit,
			},
			{
				LedgerEntryID:      "led_entry_2",
				SettlementAmount:   5_000_000,
				SettlementCurrency: "USDC",
				LedgerEntryFound:   true,
				LedgerAmount:       5_000_000,
				LedgerCurrency:     "USDC",
				LedgerRecipientID:  "merchant_1",
				LedgerAccountType:  ledger.AccountTypeMerchantPending,
				LedgerDirection:    ledger.EntryDirectionCredit,
			},
		},
	}
}

func TestReconcilerCheck(t *testing.T) {
	t.Run("Ledger와 Settlement가 일치하면 정상 리포트를 반환한다", func(t *testing.T) {
		store := &fakeReconciliationStore{snapshot: healthyReconciliationSnapshot()}
		reconciler := NewReconciler(store)

		report, err := reconciler.Check(context.Background(), "stl_batch_reconciliation_1")
		if err != nil {
			t.Fatalf("reconciliation이 성공해야 합니다: %v", err)
		}
		if !report.IsHealthy() {
			t.Fatalf("정상 데이터에는 issue가 없어야 합니다: %+v", report.Issues)
		}
		if report.CheckedItemCount != 2 {
			t.Fatalf("검사한 Item은 2개여야 합니다: %d", report.CheckedItemCount)
		}
		if store.calls != 1 || store.batchID != "stl_batch_reconciliation_1" {
			t.Fatal("요청한 Batch ID로 Snapshot을 한 번 조회해야 합니다")
		}
	})

	t.Run("요약과 Ledger 정보가 다르면 모든 불일치를 수집한다", func(t *testing.T) {
		snapshot := healthyReconciliationSnapshot()
		snapshot.Batch.ItemCount = 3
		snapshot.Batch.TotalAmount = 20_000_000
		snapshot.Items[0].SettlementAmount = 9_700_000
		snapshot.Items[0].SettlementCurrency = "KRW"
		snapshot.Items[0].LedgerRecipientID = "merchant_other"
		snapshot.Items[0].LedgerAccountType = ledger.AccountTypePlatformFee
		snapshot.Items[0].LedgerDirection = ledger.EntryDirectionDebit

		store := &fakeReconciliationStore{snapshot: snapshot}
		report, err := NewReconciler(store).Check(context.Background(), snapshot.Batch.ID)
		if err != nil {
			t.Fatalf("불일치는 오류가 아니라 리포트로 반환해야 합니다: %v", err)
		}

		expectedTypes := []MismatchType{
			MismatchItemCount,
			MismatchAmount,
			MismatchCurrency,
			MismatchRecipient,
			MismatchAccountType,
			MismatchEntryDirection,
			MismatchBatchTotal,
		}
		for _, mismatchType := range expectedTypes {
			if !reportHasIssue(report, mismatchType) {
				t.Fatalf("%s issue가 있어야 합니다: %+v", mismatchType, report.Issues)
			}
		}
	})

	t.Run("원본 Ledger Entry가 없으면 누락 issue를 반환한다", func(t *testing.T) {
		snapshot := healthyReconciliationSnapshot()
		snapshot.Items[0].LedgerEntryFound = false

		report, err := NewReconciler(&fakeReconciliationStore{snapshot: snapshot}).Check(
			context.Background(),
			snapshot.Batch.ID,
		)
		if err != nil {
			t.Fatalf("누락은 리포트로 반환해야 합니다: %v", err)
		}
		if !reportHasIssue(report, MismatchLedgerMissing) {
			t.Fatalf("Ledger 누락 issue가 있어야 합니다: %+v", report.Issues)
		}
	})

	t.Run("Snapshot 조회 실패를 전달한다", func(t *testing.T) {
		store := &fakeReconciliationStore{err: errors.New("DB 조회 실패")}
		if _, err := NewReconciler(store).Check(context.Background(), "stl_batch_error"); err == nil {
			t.Fatal("Snapshot 조회 실패는 error를 반환해야 합니다")
		}
	})
}

func reportHasIssue(report *ReconciliationReport, mismatchType MismatchType) bool {
	for _, issue := range report.Issues {
		if issue.Type == mismatchType {
			return true
		}
	}
	return false
}
