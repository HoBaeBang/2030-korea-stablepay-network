package settlement

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func newTestRepository(t *testing.T) (*Repository, *sql.DB, context.Context) {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL이 없어서 settlement repository 통합 테스트를 건너뜁니다")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("테스트 DB 연결 생성 실패: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("테스트 DB ping 실패: %v", err)
	}

	return NewRepository(db), db, ctx
}

func seedSettlementCandidate(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	accountID string,
	recipientID string,
	transactionID string,
	entryID string,
	amount int64,
) {
	t.Helper()

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO ledger_accounts (id, type, owner_id, currency)
		 VALUES ($1, $2, $3, 'USDC') ON CONFLICT (id) DO NOTHING`,
		accountID,
		ledger.AccountTypeMerchantPending,
		recipientID,
	); err != nil {
		t.Fatalf("테스트 원장 계정 생성 실패: %v", err)
	}

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO ledger_transactions (id, reference_type, reference_id, idempotency_key)
		 VALUES ($1, 'PAYMENT', $2, $3) ON CONFLICT (id) DO NOTHING`,
		transactionID,
		"pay_"+transactionID,
		"payment:"+transactionID+":finalized",
	); err != nil {
		t.Fatalf("테스트 원장 거래 생성 실패: %v", err)
	}

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO ledger_entries (id, transaction_id, account_id, direction, amount, currency)
		 VALUES ($1, $2, $3, 'CREDIT', $4, 'USDC') ON CONFLICT (id) DO NOTHING`,
		entryID,
		transactionID,
		accountID,
		amount,
	); err != nil {
		t.Fatalf("테스트 원장 항목 생성 실패: %v", err)
	}
}

func cleanupSettlementFixture(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	batchIDs []string,
	entryIDs []string,
	transactionIDs []string,
	accountIDs []string,
) {
	t.Helper()

	for _, batchID := range batchIDs {
		_, _ = db.ExecContext(ctx, "DELETE FROM settlement_items WHERE batch_id = $1", batchID)
		_, _ = db.ExecContext(ctx, "DELETE FROM settlement_batches WHERE id = $1", batchID)
	}
	for _, entryID := range entryIDs {
		_, _ = db.ExecContext(ctx, "DELETE FROM settlement_items WHERE ledger_entry_id = $1", entryID)
		_, _ = db.ExecContext(ctx, "DELETE FROM ledger_entries WHERE id = $1", entryID)
	}
	for _, transactionID := range transactionIDs {
		_, _ = db.ExecContext(ctx, "DELETE FROM ledger_transactions WHERE id = $1", transactionID)
	}
	for _, accountID := range accountIDs {
		_, _ = db.ExecContext(ctx, "DELETE FROM ledger_accounts WHERE id = $1", accountID)
	}
}

func settlementRowCount(t *testing.T, ctx context.Context, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		t.Fatalf("row count 조회 실패: %v", err)
	}
	return count
}

func TestRepositorySettlementFlow(t *testing.T) {
	t.Run("미정산 후보를 조회해 Batch와 Items를 저장한다", func(t *testing.T) {
		repository, db, ctx := newTestRepository(t)

		accountID := "acct_settlement_candidate_1"
		transactionID := "led_tx_settlement_candidate_1"
		entryID := "led_entry_settlement_candidate_1"
		batchID := "stl_batch_repository_1"
		seedSettlementCandidate(t, ctx, db, accountID, "merchant_repository_1", transactionID, entryID, 9_800_000)
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

		candidates, err := repository.FindCandidates(ctx, "merchant_repository_1", "USDC")
		if err != nil {
			t.Fatalf("정산 후보 조회가 성공해야 합니다: %v", err)
		}
		if len(candidates) != 1 || candidates[0].LedgerEntryID != entryID {
			t.Fatalf("정산 후보가 1개여야 합니다: %+v", candidates)
		}

		batch, items, err := NewCalculator().BuildBatch(
			ctx,
			batchID,
			"merchant_repository_1",
			"USDC",
			candidates,
		)
		if err != nil {
			t.Fatalf("정산 계산이 성공해야 합니다: %v", err)
		}
		if err := repository.CreateBatch(ctx, *batch, items); err != nil {
			t.Fatalf("정산 저장이 성공해야 합니다: %v", err)
		}

		batchCount := settlementRowCount(t, ctx, db, "SELECT count(*) FROM settlement_batches WHERE id = $1", batchID)
		itemCount := settlementRowCount(t, ctx, db, "SELECT count(*) FROM settlement_items WHERE batch_id = $1", batchID)
		if batchCount != 1 || itemCount != 1 {
			t.Fatalf("Batch와 Item이 각각 1개여야 합니다: batch=%d item=%d", batchCount, itemCount)
		}

		remaining, err := repository.FindCandidates(ctx, "merchant_repository_1", "USDC")
		if err != nil {
			t.Fatalf("저장 후 후보 재조회가 성공해야 합니다: %v", err)
		}
		if len(remaining) != 0 {
			t.Fatalf("이미 정산한 Entry는 후보에서 제외되어야 합니다: %+v", remaining)
		}
	})

	t.Run("같은 Ledger Entry는 다른 Batch에 중복 저장되지 않는다", func(t *testing.T) {
		repository, db, ctx := newTestRepository(t)

		accountID := "acct_settlement_duplicate_1"
		transactionID := "led_tx_settlement_duplicate_1"
		entryID := "led_entry_settlement_duplicate_1"
		firstBatchID := "stl_batch_duplicate_1"
		secondBatchID := "stl_batch_duplicate_2"
		seedSettlementCandidate(t, ctx, db, accountID, "merchant_duplicate_1", transactionID, entryID, 5_000_000)
		t.Cleanup(func() {
			cleanupSettlementFixture(
				t,
				ctx,
				db,
				[]string{firstBatchID, secondBatchID},
				[]string{entryID},
				[]string{transactionID},
				[]string{accountID},
			)
		})

		firstBatch := Batch{
			ID:          firstBatchID,
			RecipientID: "merchant_duplicate_1",
			Currency:    "USDC",
			TotalAmount: 5_000_000,
			ItemCount:   1,
			Status:      StatusDraft,
		}
		firstItems := []Item{
			{
				BatchID:       firstBatchID,
				LedgerEntryID: entryID,
				Amount:        5_000_000,
				Currency:      "USDC",
			},
		}
		if err := repository.CreateBatch(ctx, firstBatch, firstItems); err != nil {
			t.Fatalf("첫 번째 정산 저장은 성공해야 합니다: %v", err)
		}

		secondBatch := firstBatch
		secondBatch.ID = secondBatchID
		secondItems := []Item{
			{
				BatchID:       secondBatchID,
				LedgerEntryID: entryID,
				Amount:        5_000_000,
				Currency:      "USDC",
			},
		}
		if err := repository.CreateBatch(ctx, secondBatch, secondItems); err == nil {
			t.Fatal("같은 Ledger Entry의 두 번째 정산 저장은 실패해야 합니다")
		}

		secondBatchCount := settlementRowCount(t, ctx, db, "SELECT count(*) FROM settlement_batches WHERE id = $1", secondBatchID)
		if secondBatchCount != 0 {
			t.Fatal("Item 저장 실패 시 두 번째 Batch도 rollback되어야 합니다")
		}
	})

	t.Run("예상한 현재 상태일 때만 상태를 변경한다", func(t *testing.T) {
		repository, db, ctx := newTestRepository(t)
		batchID := "stl_batch_status_repository_1"
		t.Cleanup(func() {
			_, _ = db.ExecContext(ctx, "DELETE FROM settlement_batches WHERE id = $1", batchID)
		})

		if _, err := db.ExecContext(
			ctx,
			`INSERT INTO settlement_batches
			 (id, recipient_id, currency, total_amount, item_count, status)
			 VALUES ($1, 'merchant_status_1', 'USDC', 1000000, 1, 'DRAFT')`,
			batchID,
		); err != nil {
			t.Fatalf("상태 테스트 Batch 생성 실패: %v", err)
		}

		if err := repository.UpdateBatchStatus(ctx, batchID, StatusDraft, StatusReady); err != nil {
			t.Fatalf("DRAFT에서 READY DB 변경은 성공해야 합니다: %v", err)
		}
		if err := repository.UpdateBatchStatus(ctx, batchID, StatusDraft, StatusApproved); err == nil {
			t.Fatal("DB 현재 상태가 READY이므로 DRAFT 조건의 변경은 실패해야 합니다")
		}

		var status Status
		if err := db.QueryRowContext(ctx, "SELECT status FROM settlement_batches WHERE id = $1", batchID).Scan(&status); err != nil {
			t.Fatalf("정산 상태 조회 실패: %v", err)
		}
		if status != StatusReady {
			t.Fatalf("DB 상태는 READY여야 하는데 %s입니다", status)
		}
	})
}

func Example_statusFlow() {
	fmt.Println("DRAFT -> READY -> APPROVED -> PROCESSING -> PAID")
	// Output:
	// DRAFT -> READY -> APPROVED -> PROCESSING -> PAID
}
