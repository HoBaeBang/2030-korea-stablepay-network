package settlement

import "testing"

func TestRepositoryReconciliationFlow(t *testing.T) {
	repository, db, ctx := newTestRepository(t)

	accountID := "acct_reconciliation_1"
	transactionID := "led_tx_reconciliation_1"
	entryID := "led_entry_reconciliation_1"
	batchID := "stl_batch_reconciliation_repository_1"

	seedSettlementCandidate(
		t,
		ctx,
		db,
		accountID,
		"merchant_reconciliation_1",
		transactionID,
		entryID,
		9_800_000,
	)
	t.Cleanup(func() {
		cleanupSettlementFixture(
			t,
			ctx,
			db,
			[]string{batchID},
			[]string{entryID},
			[]string{transactionID},
			[]string{accountID},
		)
	})

	candidates, err := repository.FindCandidates(ctx, "merchant_reconciliation_1", "USDC")
	if err != nil {
		t.Fatalf("정산 후보 조회가 성공해야 합니다: %v", err)
	}
	batch, items, err := NewCalculator().BuildBatch(
		ctx,
		batchID,
		"merchant_reconciliation_1",
		"USDC",
		candidates,
	)
	if err != nil {
		t.Fatalf("정산 계산이 성공해야 합니다: %v", err)
	}
	if err := repository.CreateBatch(ctx, *batch, items); err != nil {
		t.Fatalf("정산 저장이 성공해야 합니다: %v", err)
	}

	t.Run("저장된 Settlement와 Ledger가 일치한다", func(t *testing.T) {
		report, err := NewReconciler(repository).Check(ctx, batchID)
		if err != nil {
			t.Fatalf("reconciliation이 성공해야 합니다: %v", err)
		}
		if !report.IsHealthy() {
			t.Fatalf("정상 데이터에는 issue가 없어야 합니다: %+v", report.Issues)
		}
	})

	t.Run("Settlement Item 금액이 바뀌면 불일치를 찾는다", func(t *testing.T) {
		if _, err := db.ExecContext(
			ctx,
			"UPDATE settlement_items SET amount = $1 WHERE batch_id = $2 AND ledger_entry_id = $3",
			9_700_000,
			batchID,
			entryID,
		); err != nil {
			t.Fatalf("불일치 fixture 생성 실패: %v", err)
		}

		report, err := NewReconciler(repository).Check(ctx, batchID)
		if err != nil {
			t.Fatalf("불일치 점검이 성공해야 합니다: %v", err)
		}
		if !reportHasIssue(report, MismatchAmount) {
			t.Fatalf("금액 불일치를 찾아야 합니다: %+v", report.Issues)
		}
		if !reportHasIssue(report, MismatchBatchTotal) {
			t.Fatalf("Batch 총액 불일치를 찾아야 합니다: %+v", report.Issues)
		}
	})
}
