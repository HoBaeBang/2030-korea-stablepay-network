package ledger

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	// pgx의 init 함수가 database/sql에 "pgx" 드라이버를 등록하도록 불러온다.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// newTestRepository는 각 통합 테스트가 사용할 Repository, DB 연결, Context를 준비한다.
// DB 준비 코드를 테스트마다 반복하지 않고, 테스트 본문이 검증할 동작에 집중하도록 만든 helper다.
func newTestRepository(t *testing.T) (*Repository, *sql.DB, context.Context) {
	// helper 내부에서 실패해도 이 함수를 호출한 테스트의 위치가 오류 위치로 표시된다.
	t.Helper()

	// 테스트 DB 주소는 코드에 고정하지 않고 실행 환경에서 주입받는다.
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL이 없어서 ledger repository integration test를 건너뜁니다")
	}

	// DB 작업이 무한히 기다리지 않도록 테스트 작업에 5초 제한을 둔다.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 테스트 종료 시 Context가 가진 타이머 자원을 해제한다.
	t.Cleanup(cancel)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("테스트 DB 연결 생성 실패: %v", err)
	}
	// 테스트 종료 시 DB 연결 풀을 닫아 자원이 다음 테스트에 남지 않게 한다.
	t.Cleanup(func() { _ = db.Close() })

	// sql.Open은 실제 연결 성공을 보장하지 않으므로 Ping으로 접속 가능 여부를 확인한다.
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("테스트 DB ping 실패: %v", err)
	}

	return NewRepository(db), db, ctx
}

// seedLedgerAccount는 ledger_entries의 외래 키가 참조할 원장 계정을 미리 저장한다.
// Repository의 거래 저장을 검증하는 테스트가 선행 계정 준비 코드로 복잡해지지 않게 분리한 helper다.
func seedLedgerAccount(t *testing.T, ctx context.Context, db *sql.DB, account Account) {
	t.Helper()

	const query = `
		INSERT INTO ledger_accounts (id, type, owner_id, currency)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO NOTHING
	`
	if _, err := db.ExecContext(ctx, query, account.ID, account.Type, account.OwnerID, account.Currency); err != nil {
		t.Fatalf("테스트 원장 계정 생성 실패: %v", err)
	}
}

// cleanupLedgerTransaction은 테스트를 반복 실행할 때 이전 거래 데이터가 결과에 영향을 주지 않도록 정리한다.
// 외래 키 제약을 지키기 위해 자식인 ledger_entries를 먼저 삭제한 뒤 ledger_transactions를 삭제한다.
func cleanupLedgerTransaction(t *testing.T, ctx context.Context, db *sql.DB, transactionID string) {
	t.Helper()
	_, _ = db.ExecContext(ctx, "DELETE FROM ledger_entries WHERE transaction_id = $1", transactionID)
	_, _ = db.ExecContext(ctx, "DELETE FROM ledger_transactions WHERE id = $1", transactionID)
}

// countRows는 저장 또는 rollback 이후 DB에 남은 행 개수를 조회해 테스트 결과를 검증한다.
// 조회 SQL과 조건값을 인자로 받아 여러 테이블의 행 개수 검증에 재사용한다.
func countRows(t *testing.T, ctx context.Context, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		t.Fatalf("row count 조회 실패: %v", err)
	}
	return count
}

func TestRepositoryCreateTransactionValidation(t *testing.T) {
	t.Run("원장 거래 ID가 비어 있으면 DB 저장 전에 실패한다", func(t *testing.T) {
		repo := NewRepository(nil)

		err := repo.CreateTransaction(context.Background(), Transaction{}, []Entry{{ID: "entry_1"}})
		if err == nil {
			t.Fatal("원장 거래 ID가 비어 있으면 검증 오류가 발생해야 합니다")
		}
	})

	t.Run("원장 항목이 없으면 DB 저장 전에 실패한다", func(t *testing.T) {
		repo := NewRepository(nil)
		tx := Transaction{
			ID:             "led_tx_validation_1",
			ReferenceType:  "PAYMENT",
			ReferenceID:    "pay_validation_1",
			IdempotencyKey: "payment:pay_validation_1:finalized",
		}

		err := repo.CreateTransaction(context.Background(), tx, nil)
		if err == nil {
			t.Fatal("원장 항목이 없으면 검증 오류가 발생해야 합니다")
		}
	})
}

func TestRepositoryCreateTransaction(t *testing.T) {
	t.Run("transaction과 entries를 함께 저장한다.", func(t *testing.T) {
		repo, db, ctx := newTestRepository(t)

		seedLedgerAccount(t, ctx, db, Account{ID: "acct_test_customer_1", Type: AccountTypeCustomer, OwnerID: "customer_test_1", Currency: "USDC"})
		seedLedgerAccount(t, ctx, db, Account{ID: "acct_test_merchant_pending_1", Type: AccountTypeMerchantPending, OwnerID: "merchant_test_1", Currency: "USDC"})
		seedLedgerAccount(t, ctx, db, Account{ID: "acct_test_platform_fee_1", Type: AccountTypePlatformFee, OwnerID: "platform", Currency: "USDC"})

		tx := Transaction{
			ID:             "led_tx_test_success_1",
			ReferenceType:  "PAYMENT",
			ReferenceID:    "pay_test_success_1",
			IdempotencyKey: "payment:pay_test_success_1:finalized",
		}
		entries := []Entry{
			{ID: "led_entry_test_success_1", TransactionID: tx.ID, AccountID: "acct_test_customer_1", Direction: EntryDirectionDebit, Amount: 10_000_000, Currency: "USDC"},
			{ID: "led_entry_test_success_2", TransactionID: tx.ID, AccountID: "acct_test_merchant_pending_1", Direction: EntryDirectionCredit, Amount: 9_800_000, Currency: "USDC"},
			{ID: "led_entry_test_success_3", TransactionID: tx.ID, AccountID: "acct_test_platform_fee_1", Direction: EntryDirectionCredit, Amount: 200_000, Currency: "USDC"},
		}
		t.Cleanup(func() { cleanupLedgerTransaction(t, ctx, db, tx.ID) })

		if err := repo.CreateTransaction(ctx, tx, entries); err != nil {
			t.Fatalf("원장 거래 저장이 성공해야 하는데 실패했습니다: %v", err)
		}

		transactionCount := countRows(t, ctx, db, "SELECT count(*) FROM ledger_transactions WHERE id = $1", tx.ID)
		if transactionCount != 1 {
			t.Fatalf("ledger_transactions row가 1개여야 하는데 %d개입니다", transactionCount)
		}

		entryCount := countRows(t, ctx, db, "SELECT count(*) FROM ledger_entries WHERE transaction_id = $1", tx.ID)
		if entryCount != len(entries) {
			t.Fatalf("ledger_entries row가 %d개여야 하는데 %d개입니다", len(entries), entryCount)
		}
	})

	t.Run("같은 idempotency_key는 중복 저장되지 않는다", func(t *testing.T) {
		repo, db, ctx := newTestRepository(t)
		seedLedgerAccount(t, ctx, db, Account{ID: "acct_test_customer_2", Type: AccountTypeCustomer, OwnerID: "customer_test_2", Currency: "USDC"})
		seedLedgerAccount(t, ctx, db, Account{ID: "acct_test_merchant_pending_2", Type: AccountTypeMerchantPending, OwnerID: "merchant_test_2", Currency: "USDC"})
		idempotencyKey := "payment:pay_test_duplicate:finalized"
		firstTx := Transaction{ID: "led_tx_test_duplicate_1", ReferenceType: "PAYMENT", ReferenceID: "pay_test_duplicate", IdempotencyKey: idempotencyKey}
		firstEntries := []Entry{
			{ID: "led_entry_test_duplicate_1", TransactionID: firstTx.ID, AccountID: "acct_test_customer_2", Direction: EntryDirectionDebit, Amount: 10_000_000, Currency: "USDC"},
			{ID: "led_entry_test_duplicate_2", TransactionID: firstTx.ID, AccountID: "acct_test_merchant_pending_2", Direction: EntryDirectionCredit, Amount: 10_000_000, Currency: "USDC"},
		}
		t.Cleanup(func() { cleanupLedgerTransaction(t, ctx, db, firstTx.ID) })
		if err := repo.CreateTransaction(ctx, firstTx, firstEntries); err != nil {
			t.Fatalf("첫 번째 저장은 성공해야 합니다: %v", err)
		}
		secondTx := Transaction{ID: "led_tx_test_duplicate_2", ReferenceType: "PAYMENT", ReferenceID: "pay_test_duplicate", IdempotencyKey: idempotencyKey}
		secondEntries := []Entry{
			{ID: "led_entry_test_duplicate_3", TransactionID: secondTx.ID, AccountID: "acct_test_customer_2", Direction: EntryDirectionDebit, Amount: 10_000_000, Currency: "USDC"},
			{ID: "led_entry_test_duplicate_4", TransactionID: secondTx.ID, AccountID: "acct_test_merchant_pending_2", Direction: EntryDirectionCredit, Amount: 10_000_000, Currency: "USDC"},
		}
		t.Cleanup(func() { cleanupLedgerTransaction(t, ctx, db, secondTx.ID) })
		if err := repo.CreateTransaction(ctx, secondTx, secondEntries); err == nil {
			t.Fatal("같은 idempotency_key는 중복 저장에 실패해야 하는데 nil이 반환되었습니다")
		}
	})
	t.Run("entry 저장 실패 시 transaction도 rollback된다", func(t *testing.T) {
		repo, db, ctx := newTestRepository(t)
		tx := Transaction{
			ID:             "led_tx_test_rollback_1",
			ReferenceType:  "PAYMENT",
			ReferenceID:    "pay_test_rollback_1",
			IdempotencyKey: "payment:pay_test_rollback_1:finalized",
		}
		entries := []Entry{
			{ID: "led_entry_test_rollback_1", TransactionID: tx.ID, AccountID: "acct_missing_for_rollback", Direction: EntryDirectionDebit, Amount: 10_000_000, Currency: "USDC"},
			{ID: "led_entry_test_rollback_2", TransactionID: tx.ID, AccountID: "acct_missing_for_rollback_2", Direction: EntryDirectionCredit, Amount: 10_000_000, Currency: "USDC"},
		}
		t.Cleanup(func() { cleanupLedgerTransaction(t, ctx, db, tx.ID) })
		if err := repo.CreateTransaction(ctx, tx, entries); err == nil {
			t.Fatal("존재하지 않는 account_id를 사용하면 저장이 실패해야 하는데 nil이 반환되었습니다")
		}
		transactionCount := countRows(t, ctx, db, "SELECT count(*) FROM ledger_transactions WHERE id = $1", tx.ID)
		if transactionCount != 0 {
			t.Fatalf("rollback 후 transaction row가 남아 있으면 안 되는데 %d개가 남았습니다", transactionCount)
		}
	})
}
func Example_idempotencyKey() {
	paymentID := "pay_123"
	key := fmt.Sprintf("payment:%s:finalized", paymentID)
	fmt.Println(key)
	// Output:
	// payment:pay_123:finalized
}
