# Day 23 구현가이드 - Settlement DB 저장과 상태 전이

관련 Jira: [SPN-40](https://aslan0.atlassian.net/browse/SPN-40)

Day22에서는 정산 후보를 메모리에서 검증하고 Batch와 Items로 계산했습니다. Day23에서는 계산 결과를 PostgreSQL에 보존하고, 정산 묶음이 허용된 순서로만 상태를 변경하도록 연결합니다.

## 오늘 완성할 흐름

```text
미정산 MERCHANT_PENDING CREDIT 조회
-> Calculator가 Batch와 Items 계산
-> 하나의 DB transaction으로 함께 저장
-> DRAFT부터 허용된 순서로 상태 변경
```

오늘 새로 작성하거나 수정할 파일:

```text
migrations/000003_create_settlement_tables.up.sql
migrations/000003_create_settlement_tables.down.sql
internal/settlement/settlement.go
internal/settlement/repository.go
internal/settlement/repository_test.go
internal/settlement/service.go
internal/settlement/service_test.go
```

`calculator.go`와 `calculator_test.go`는 Day22 코드를 그대로 사용합니다.

## 먼저 작성할 전체 코드

아래 완성본을 파일별로 먼저 작성합니다. 모든 코드는 별도 복사본과 임시 PostgreSQL에서 컴파일, 단위 테스트, 통합 테스트를 검증한 상태입니다.

## `000003_create_settlement_tables.up.sql` 최종 완성본 전체

<details>
<summary><code>000003_create_settlement_tables.up.sql</code> 최종 완성본 전체 보기</summary>

```sql
CREATE TABLE settlement_batches
(
    id           TEXT PRIMARY KEY,
    recipient_id TEXT        NOT NULL,
    currency     TEXT        NOT NULL,
    total_amount BIGINT      NOT NULL CHECK (total_amount > 0),
    item_count   INTEGER     NOT NULL CHECK (item_count > 0),
    status       TEXT        NOT NULL CHECK
        (status IN ('DRAFT', 'READY', 'APPROVED', 'PROCESSING', 'PAID', 'FAILED', 'CANCELED')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_settlement_batches_recipient_currency_status
    ON settlement_batches (recipient_id, currency, status);

CREATE TABLE settlement_items
(
    batch_id       TEXT   NOT NULL REFERENCES settlement_batches (id),
    ledger_entry_id TEXT  NOT NULL REFERENCES ledger_entries (id),
    amount          BIGINT NOT NULL CHECK (amount > 0),
    currency        TEXT   NOT NULL,
    PRIMARY KEY (batch_id, ledger_entry_id)
);

CREATE UNIQUE INDEX idx_settlement_items_ledger_entry_id
    ON settlement_items (ledger_entry_id);
```

</details>

## `000003_create_settlement_tables.down.sql` 최종 완성본 전체

<details>
<summary><code>000003_create_settlement_tables.down.sql</code> 최종 완성본 전체 보기</summary>

```sql
DROP TABLE IF EXISTS settlement_items;
DROP TABLE IF EXISTS settlement_batches;
```

</details>

## `settlement.go` 최종 완성본 전체

<details>
<summary><code>settlement.go</code> 최종 완성본 전체 보기</summary>

```go
package settlement

import "github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"

// Status는 정산 묶음이 현재 어떤 처리 단계에 있는지를 나타낸다.
type Status string

const (
	// StatusDraft는 정산 후보를 계산해 묶음을 처음 만든 상태다.
	StatusDraft Status = "DRAFT"
	// StatusReady는 정산 항목 검증이 끝나 승인할 수 있는 상태다.
	StatusReady Status = "READY"
	// StatusApproved는 내부 정책이나 관리자의 지급 승인이 끝난 상태다.
	StatusApproved Status = "APPROVED"
	// StatusProcessing은 실제 지급 요청을 처리 중인 상태다.
	StatusProcessing Status = "PROCESSING"
	// StatusPaid는 지급 완료가 확인된 최종 상태다.
	StatusPaid Status = "PAID"
	// StatusFailed는 지급 시도가 실패해 재시도나 취소가 필요한 상태다.
	StatusFailed Status = "FAILED"
	// StatusCanceled는 지급 전에 정산 처리를 취소한 최종 상태다.
	StatusCanceled Status = "CANCELED"
)

// Candidate는 Ledger에서 조회한 정산 대상 후보를 나타낸다.
// 아직 Settlement에 포함된 것은 아니며 Calculator 검증을 통과해야 한다.
type Candidate struct {
	LedgerEntryID string
	RecipientID   string
	AccountType   ledger.AccountType
	Direction     ledger.EntryDirection
	Amount        int64
	Currency      string
}

// Batch는 같은 수취인과 통화의 정산 대상들을 하나로 묶은 결과다.
type Batch struct {
	ID          string
	RecipientID string
	Currency    string
	TotalAmount int64
	ItemCount   int
	Status      Status
}

// Item은 어떤 Ledger Entry가 정산 묶음에 포함되었는지를 나타내는 근거다.
type Item struct {
	BatchID       string
	LedgerEntryID string
	Amount        int64
	Currency      string
}
```

</details>

## `repository.go` 최종 완성본 전체

<details>
<summary><code>repository.go</code> 최종 완성본 전체 보기</summary>

```go
package settlement

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Repository는 Settlement 데이터를 DB에서 조회하고 저장하는 경계다.
type Repository struct {
	db *sql.DB
}

// NewRepository는 Settlement Repository 인스턴스를 만든다.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// FindCandidates는 아직 정산되지 않은 가맹점 지급 예정 CREDIT 항목을 조회한다.
func (r *Repository) FindCandidates(
	ctx context.Context,
	recipientID string,
	currency string,
) ([]Candidate, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}

	recipientID = strings.TrimSpace(recipientID)
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if recipientID == "" {
		return nil, fmt.Errorf("정산 수취인 id가 필요합니다")
	}
	if currency == "" {
		return nil, fmt.Errorf("정산 통화가 필요합니다")
	}
	if r.db == nil {
		return nil, fmt.Errorf("settlement repository db가 필요합니다")
	}

	const query = `
		SELECT
			le.id,
			la.owner_id,
			la.type,
			le.direction,
			le.amount,
			le.currency
		FROM ledger_entries le
		JOIN ledger_accounts la ON la.id = le.account_id
		LEFT JOIN settlement_items si ON si.ledger_entry_id = le.id
		WHERE la.owner_id = $1
		  AND la.type = 'MERCHANT_PENDING'
		  AND le.direction = 'CREDIT'
		  AND le.currency = $2
		  AND si.ledger_entry_id IS NULL
		ORDER BY le.created_at, le.id
	`

	rows, err := r.db.QueryContext(ctx, query, recipientID, currency)
	if err != nil {
		return nil, fmt.Errorf("정산 후보 조회 실패: %w", err)
	}
	defer rows.Close()

	candidates := make([]Candidate, 0)
	for rows.Next() {
		var candidate Candidate
		if err := rows.Scan(
			&candidate.LedgerEntryID,
			&candidate.RecipientID,
			&candidate.AccountType,
			&candidate.Direction,
			&candidate.Amount,
			&candidate.Currency,
		); err != nil {
			return nil, fmt.Errorf("정산 후보 변환 실패: %w", err)
		}
		candidates = append(candidates, candidate)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("정산 후보 순회 실패: %w", err)
	}

	return candidates, nil
}

// CreateBatch는 Batch와 Items를 하나의 DB transaction으로 저장한다.
func (r *Repository) CreateBatch(ctx context.Context, batch Batch, items []Item) error {
	if err := validateContext(ctx); err != nil {
		return err
	}
	if err := validateBatchAndItems(batch, items); err != nil {
		return err
	}
	if r.db == nil {
		return fmt.Errorf("settlement repository db가 필요합니다")
	}

	sqlTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("정산 저장 transaction 시작 실패: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = sqlTx.Rollback()
		}
	}()

	const insertBatchQuery = `
		INSERT INTO settlement_batches
			(id, recipient_id, currency, total_amount, item_count, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	if _, err := sqlTx.ExecContext(
		ctx,
		insertBatchQuery,
		batch.ID,
		batch.RecipientID,
		batch.Currency,
		batch.TotalAmount,
		batch.ItemCount,
		batch.Status,
	); err != nil {
		return fmt.Errorf("정산 묶음 저장 실패: %w", err)
	}

	const insertItemQuery = `
		INSERT INTO settlement_items (batch_id, ledger_entry_id, amount, currency)
		VALUES ($1, $2, $3, $4)
	`

	for _, item := range items {
		if _, err := sqlTx.ExecContext(
			ctx,
			insertItemQuery,
			item.BatchID,
			item.LedgerEntryID,
			item.Amount,
			item.Currency,
		); err != nil {
			return fmt.Errorf("정산 항목 저장 실패: %w", err)
		}
	}

	if err := sqlTx.Commit(); err != nil {
		return fmt.Errorf("정산 저장 commit 실패: %w", err)
	}
	committed = true

	return nil
}

// UpdateBatchStatus는 예상한 현재 상태일 때만 다음 상태로 변경한다.
func (r *Repository) UpdateBatchStatus(
	ctx context.Context,
	batchID string,
	currentStatus Status,
	nextStatus Status,
) error {
	if err := validateContext(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(batchID) == "" {
		return fmt.Errorf("settlement batch id가 필요합니다")
	}
	if r.db == nil {
		return fmt.Errorf("settlement repository db가 필요합니다")
	}

	const query = `
		UPDATE settlement_batches
		SET status = $1, updated_at = now()
		WHERE id = $2 AND status = $3
	`

	result, err := r.db.ExecContext(ctx, query, nextStatus, batchID, currentStatus)
	if err != nil {
		return fmt.Errorf("정산 상태 변경 실패: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("정산 상태 변경 결과 확인 실패: %w", err)
	}
	if rowsAffected != 1 {
		return fmt.Errorf("정산 상태가 변경되지 않았습니다: batch=%s, expected=%s", batchID, currentStatus)
	}

	return nil
}

func validateContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context가 필요합니다")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func validateBatchAndItems(batch Batch, items []Item) error {
	if strings.TrimSpace(batch.ID) == "" {
		return fmt.Errorf("settlement batch id가 필요합니다")
	}
	if strings.TrimSpace(batch.RecipientID) == "" {
		return fmt.Errorf("settlement recipient id가 필요합니다")
	}
	if strings.TrimSpace(batch.Currency) == "" {
		return fmt.Errorf("settlement currency가 필요합니다")
	}
	if batch.TotalAmount <= 0 {
		return fmt.Errorf("settlement total amount는 0보다 커야 합니다")
	}
	if batch.Status != StatusDraft {
		return fmt.Errorf("새 settlement batch 상태는 DRAFT여야 합니다")
	}
	if len(items) == 0 || batch.ItemCount != len(items) {
		return fmt.Errorf("settlement item 개수가 batch item count와 일치해야 합니다")
	}

	var totalAmount int64
	seenEntryIDs := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item.BatchID != batch.ID {
			return fmt.Errorf("settlement item의 batch id가 다릅니다")
		}
		if strings.TrimSpace(item.LedgerEntryID) == "" {
			return fmt.Errorf("settlement item의 ledger entry id가 필요합니다")
		}
		if _, exists := seenEntryIDs[item.LedgerEntryID]; exists {
			return fmt.Errorf("중복된 settlement ledger entry입니다: %s", item.LedgerEntryID)
		}
		seenEntryIDs[item.LedgerEntryID] = struct{}{}
		if item.Amount <= 0 {
			return fmt.Errorf("settlement item amount는 0보다 커야 합니다")
		}
		if item.Currency != batch.Currency {
			return fmt.Errorf("settlement item 통화가 batch 통화와 다릅니다")
		}
		totalAmount += item.Amount
	}

	if totalAmount != batch.TotalAmount {
		return fmt.Errorf("settlement item 합계가 batch total amount와 다릅니다")
	}

	return nil
}
```

</details>

## `repository_test.go` 최종 완성본 전체

<details>
<summary><code>repository_test.go</code> 최종 완성본 전체 보기</summary>

```go
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
```

</details>

## `service.go` 최종 완성본 전체

<details>
<summary><code>service.go</code> 최종 완성본 전체 보기</summary>

```go
package settlement

import (
	"context"
	"fmt"
)

// Store는 Service가 필요로 하는 Settlement 저장소 동작을 정의한다.
type Store interface {
	FindCandidates(ctx context.Context, recipientID string, currency string) ([]Candidate, error)
	CreateBatch(ctx context.Context, batch Batch, items []Item) error
	UpdateBatchStatus(ctx context.Context, batchID string, currentStatus Status, nextStatus Status) error
}

// Service는 후보 조회, 계산, 저장과 상태 전이 흐름을 조정한다.
type Service struct {
	store      Store
	calculator *Calculator
}

// NewService는 Settlement Service 인스턴스를 만든다.
func NewService(store Store, calculator *Calculator) *Service {
	return &Service{store: store, calculator: calculator}
}

// CreateBatch는 정산 후보를 조회해 계산하고 DRAFT Batch와 Items를 저장한다.
func (s *Service) CreateBatch(
	ctx context.Context,
	batchID string,
	recipientID string,
	currency string,
) (*Batch, []Item, error) {
	if err := validateContext(ctx); err != nil {
		return nil, nil, err
	}
	if s.store == nil {
		return nil, nil, fmt.Errorf("settlement store가 필요합니다")
	}
	if s.calculator == nil {
		return nil, nil, fmt.Errorf("settlement calculator가 필요합니다")
	}

	candidates, err := s.store.FindCandidates(ctx, recipientID, currency)
	if err != nil {
		return nil, nil, fmt.Errorf("정산 후보 준비 실패: %w", err)
	}

	batch, items, err := s.calculator.BuildBatch(
		ctx,
		batchID,
		recipientID,
		currency,
		candidates,
	)
	if err != nil {
		return nil, nil, err
	}

	if err := s.store.CreateBatch(ctx, *batch, items); err != nil {
		return nil, nil, fmt.Errorf("정산 묶음 저장 실패: %w", err)
	}

	return batch, items, nil
}

// TransitionStatus는 허용된 상태 전이인지 확인한 뒤 DB 상태를 변경한다.
func (s *Service) TransitionStatus(
	ctx context.Context,
	batch Batch,
	nextStatus Status,
) (*Batch, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	if s.store == nil {
		return nil, fmt.Errorf("settlement store가 필요합니다")
	}
	if !canTransition(batch.Status, nextStatus) {
		return nil, fmt.Errorf("허용되지 않은 정산 상태 전이입니다: %s -> %s", batch.Status, nextStatus)
	}

	if err := s.store.UpdateBatchStatus(ctx, batch.ID, batch.Status, nextStatus); err != nil {
		return nil, err
	}

	batch.Status = nextStatus
	return &batch, nil
}

func canTransition(currentStatus Status, nextStatus Status) bool {
	allowed := map[Status]map[Status]struct{}{
		StatusDraft: {
			StatusReady:    {},
			StatusCanceled: {},
		},
		StatusReady: {
			StatusApproved: {},
			StatusCanceled: {},
		},
		StatusApproved: {
			StatusProcessing: {},
			StatusCanceled:   {},
		},
		StatusProcessing: {
			StatusPaid:   {},
			StatusFailed: {},
		},
		StatusFailed: {
			StatusProcessing: {},
			StatusCanceled:   {},
		},
	}

	nextStatuses, exists := allowed[currentStatus]
	if !exists {
		return false
	}
	_, exists = nextStatuses[nextStatus]
	return exists
}
```

</details>

## `service_test.go` 최종 완성본 전체

<details>
<summary><code>service_test.go</code> 최종 완성본 전체 보기</summary>

```go
package settlement

import (
	"context"
	"errors"
	"testing"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"
)

type fakeStore struct {
	candidates        []Candidate
	findErr           error
	createErr         error
	updateErr         error
	findCalls         int
	createCalls       int
	updateCalls       int
	savedBatch        Batch
	savedItems        []Item
	updatedBatchID    string
	updatedFromStatus Status
	updatedToStatus   Status
}

func (f *fakeStore) FindCandidates(
	ctx context.Context,
	recipientID string,
	currency string,
) ([]Candidate, error) {
	f.findCalls++
	return append([]Candidate(nil), f.candidates...), f.findErr
}

func (f *fakeStore) CreateBatch(ctx context.Context, batch Batch, items []Item) error {
	f.createCalls++
	f.savedBatch = batch
	f.savedItems = append([]Item(nil), items...)
	return f.createErr
}

func (f *fakeStore) UpdateBatchStatus(
	ctx context.Context,
	batchID string,
	currentStatus Status,
	nextStatus Status,
) error {
	f.updateCalls++
	f.updatedBatchID = batchID
	f.updatedFromStatus = currentStatus
	f.updatedToStatus = nextStatus
	return f.updateErr
}

func validServiceCandidates() []Candidate {
	return []Candidate{
		{
			LedgerEntryID: "led_entry_service_1",
			RecipientID:   "merchant_1",
			AccountType:   ledger.AccountTypeMerchantPending,
			Direction:     ledger.EntryDirectionCredit,
			Amount:        9_800_000,
			Currency:      "USDC",
		},
		{
			LedgerEntryID: "led_entry_service_2",
			RecipientID:   "merchant_1",
			AccountType:   ledger.AccountTypeMerchantPending,
			Direction:     ledger.EntryDirectionCredit,
			Amount:        5_000_000,
			Currency:      "USDC",
		},
	}
}

func TestServiceCreateBatch(t *testing.T) {
	t.Run("후보를 조회하고 계산한 Batch와 Items를 저장한다", func(t *testing.T) {
		store := &fakeStore{candidates: validServiceCandidates()}
		service := NewService(store, NewCalculator())

		batch, items, err := service.CreateBatch(
			context.Background(),
			"stl_batch_service_1",
			"merchant_1",
			"USDC",
		)
		if err != nil {
			t.Fatalf("정산 묶음 생성이 성공해야 합니다: %v", err)
		}

		if store.findCalls != 1 || store.createCalls != 1 {
			t.Fatalf("후보 조회와 저장은 각각 1번이어야 합니다: find=%d create=%d", store.findCalls, store.createCalls)
		}
		if batch.TotalAmount != 14_800_000 || len(items) != 2 {
			t.Fatalf("정산 계산 결과가 다릅니다: total=%d items=%d", batch.TotalAmount, len(items))
		}
		if store.savedBatch.ID != batch.ID || len(store.savedItems) != len(items) {
			t.Fatal("계산된 Batch와 Items가 저장소에 전달되어야 합니다")
		}
	})

	t.Run("후보 조회에 실패하면 저장하지 않는다", func(t *testing.T) {
		store := &fakeStore{findErr: errors.New("후보 조회 실패")}
		service := NewService(store, NewCalculator())

		_, _, err := service.CreateBatch(
			context.Background(),
			"stl_batch_service_error",
			"merchant_1",
			"USDC",
		)
		if err == nil {
			t.Fatal("후보 조회 실패는 에러를 반환해야 합니다")
		}
		if store.createCalls != 0 {
			t.Fatal("후보 조회에 실패하면 Batch를 저장하면 안 됩니다")
		}
	})
}

func TestServiceTransitionStatus(t *testing.T) {
	t.Run("DRAFT에서 READY로 전이한다", func(t *testing.T) {
		store := &fakeStore{}
		service := NewService(store, NewCalculator())
		batch := Batch{ID: "stl_batch_status_1", Status: StatusDraft}

		updated, err := service.TransitionStatus(context.Background(), batch, StatusReady)
		if err != nil {
			t.Fatalf("DRAFT에서 READY 전이는 성공해야 합니다: %v", err)
		}
		if updated.Status != StatusReady {
			t.Fatalf("변경된 상태는 READY여야 하는데 %s입니다", updated.Status)
		}
		if store.updateCalls != 1 || store.updatedFromStatus != StatusDraft || store.updatedToStatus != StatusReady {
			t.Fatal("저장소에 DRAFT -> READY 상태 변경이 전달되어야 합니다")
		}
	})

	t.Run("DRAFT에서 PAID로 바로 전이할 수 없다", func(t *testing.T) {
		store := &fakeStore{}
		service := NewService(store, NewCalculator())
		batch := Batch{ID: "stl_batch_status_invalid", Status: StatusDraft}

		if _, err := service.TransitionStatus(context.Background(), batch, StatusPaid); err == nil {
			t.Fatal("DRAFT에서 PAID로 직접 전이하면 실패해야 합니다")
		}
		if store.updateCalls != 0 {
			t.Fatal("허용되지 않은 전이는 저장소를 호출하면 안 됩니다")
		}
	})

	t.Run("PAID는 다른 상태로 전이할 수 없다", func(t *testing.T) {
		store := &fakeStore{}
		service := NewService(store, NewCalculator())
		batch := Batch{ID: "stl_batch_paid", Status: StatusPaid}

		if _, err := service.TransitionStatus(context.Background(), batch, StatusFailed); err == nil {
			t.Fatal("PAID는 최종 상태이므로 전이하면 실패해야 합니다")
		}
	})
}
```

</details>

## 코드 작성 후 이해할 전체 구조

![Day23 Settlement DB 저장과 상태 전이](https://raw.githubusercontent.com/HoBaeBang/2030-korea-stablepay-network/main/docs/confluence/diagrams/spn40-day23-settlement-db-status.png)

```text
Settlement Service
-> Repository.FindCandidates
-> Calculator.BuildBatch
-> Repository.CreateBatch

상태 변경 요청
-> Service.TransitionStatus
-> 허용된 전이인지 검사
-> Repository.UpdateBatchStatus
```

## 출퇴근 예습 1. 왜 Settlement DB가 필요한가?

Day22의 Batch와 Items는 함수가 반환한 Go 값이었습니다. 프로세스가 종료되면 사라지고, 다음 요청에서 어떤 Ledger Entry가 이미 정산됐는지 알 수 없습니다.

DB에 저장하면 다음 질문에 답할 수 있습니다.

```text
누구에게 어떤 통화로 얼마를 정산하는가?
어떤 Ledger Entry들이 총액의 근거인가?
현재 승인·지급 단계는 어디인가?
같은 Ledger Entry를 이미 정산했는가?
오류가 발생했을 때 어디서부터 복구해야 하는가?
```

## 출퇴근 예습 2. 두 테이블의 역할

### `settlement_batches`

이 테이블은 **정산 묶음의 요약과 현재 처리 상태**를 저장합니다.

| 컬럼 | 의미 |
| --- | --- |
| `id` | 정산 묶음 고유 ID |
| `recipient_id` | 정산금을 받을 주체 ID |
| `currency` | 정산 통화 |
| `total_amount` | Items 금액 합계 |
| `item_count` | 포함된 Item 개수 |
| `status` | DRAFT부터 시작하는 처리 상태 |
| `created_at` | Batch 생성 시각 |
| `updated_at` | 마지막 상태 변경 시각 |

이 테이블이 답하는 질문:

> `merchant_1`에게 지급할 USDC 정산 묶음은 얼마이며 현재 어느 단계인가?

### `settlement_items`

이 테이블은 **Batch 총액의 근거가 된 개별 Ledger Entry**를 저장합니다.

| 컬럼 | 의미 |
| --- | --- |
| `batch_id` | 어떤 Settlement Batch에 포함됐는가 |
| `ledger_entry_id` | 어떤 Ledger Entry가 근거인가 |
| `amount` | 정산에 포함된 금액 |
| `currency` | 정산 통화 |

이 테이블이 답하는 질문:

> 이 Batch의 14.8 USDC는 어떤 결제 원장 항목들로 구성됐는가?

## 왜 `ledger_entry_id`가 UNIQUE인가?

```sql
CREATE UNIQUE INDEX idx_settlement_items_ledger_entry_id
    ON settlement_items (ledger_entry_id);
```

`PRIMARY KEY (batch_id, ledger_entry_id)`만 있으면 같은 Ledger Entry가 다른 Batch ID와 함께 다시 저장될 수 있습니다.

```text
허용하면 안 되는 데이터

batch_1 + ledger_entry_1
batch_2 + ledger_entry_1
```

별도의 UNIQUE 제약은 Ledger Entry 하나가 전체 Settlement 시스템에서 한 번만 정산되도록 보장합니다. Calculator의 `seenEntryIDs`는 한 번의 계산 안에서 중복을 막고, DB UNIQUE는 여러 요청과 여러 서버 사이의 중복을 최종적으로 막습니다.

## 왜 Batch와 Items를 하나의 DB transaction으로 저장하는가?

```text
Batch 저장 성공
-> Item 1 저장 성공
-> Item 2 저장 실패
```

transaction이 없다면 총액은 14.8 USDC인데 Item은 9.8 USDC 하나만 남는 불완전한 정산이 만들어질 수 있습니다.

`CreateBatch`는 다음 원칙을 지킵니다.

```text
모든 INSERT 성공 -> COMMIT
하나라도 실패    -> ROLLBACK
```

## 미정산 후보 조회 SQL 해석

핵심 조건은 다음과 같습니다.

```sql
LEFT JOIN settlement_items si ON si.ledger_entry_id = le.id
WHERE la.owner_id = $1
  AND la.type = 'MERCHANT_PENDING'
  AND le.direction = 'CREDIT'
  AND le.currency = $2
  AND si.ledger_entry_id IS NULL
```

`LEFT JOIN` 후 `si.ledger_entry_id IS NULL`인 항목은 `settlement_items`에서 연결된 기록을 찾지 못한 Ledger Entry입니다. 즉 아직 정산되지 않은 후보입니다.

```text
Ledger Entry 있음 + Settlement Item 없음 -> 정산 후보
Ledger Entry 있음 + Settlement Item 있음 -> 이미 정산됨, 제외
```

## 상태와 전이 순서

| 상태 | 언제 이 상태가 되는가 |
| --- | --- |
| `DRAFT` | 후보를 계산해 Batch를 처음 저장했을 때 |
| `READY` | 항목·총액·정책 검증을 마쳤을 때 |
| `APPROVED` | 내부 정책 또는 관리자가 지급을 승인했을 때 |
| `PROCESSING` | 은행이나 온체인 지급 요청을 시작했을 때 |
| `PAID` | 실제 지급 완료를 확인했을 때 |
| `FAILED` | 지급 시도가 실패했을 때 |
| `CANCELED` | 지급 전 또는 실패 후 정산을 취소했을 때 |

정상 지급 흐름:

```text
DRAFT -> READY -> APPROVED -> PROCESSING -> PAID
```

실패와 취소:

```text
PROCESSING -> FAILED -> PROCESSING  재시도
DRAFT / READY / APPROVED / FAILED -> CANCELED
```

`PAID`와 `CANCELED`는 최종 상태이므로 더 이상 다른 상태로 변경하지 않습니다.

## `UPDATE ... WHERE status = current`가 필요한 이유

```sql
UPDATE settlement_batches
SET status = $1, updated_at = now()
WHERE id = $2 AND status = $3
```

두 요청이 동시에 `DRAFT` Batch를 변경한다고 가정합니다.

```text
요청 A: DRAFT -> READY
요청 B: DRAFT -> CANCELED
```

요청 A가 먼저 성공하면 DB 상태는 READY가 됩니다. 요청 B의 `WHERE status = 'DRAFT'` 조건은 더 이상 맞지 않으므로 변경된 row가 0개가 되고 실패합니다.

이 방식은 마지막 요청이 앞선 변경을 조용히 덮어쓰는 것을 막는 **낙관적 동시성 제어**입니다.

## Service, Calculator, Repository 책임

| 구성 요소 | 책임 |
| --- | --- |
| `Service` | 후보 조회 → 계산 → 저장 순서를 조정하고 상태 전이 정책을 검사한다 |
| `Calculator` | Candidate를 검증하고 Batch와 Items를 계산한다 |
| `Repository` | PostgreSQL에서 후보를 조회하고 Batch·Items·상태를 저장한다 |

Service가 SQL을 직접 실행하지 않고, Repository가 상태 전이 정책을 결정하지 않는 것이 핵심입니다.

## 퇴근 후 실습 순서

### Step 1. Migration 파일 작성

앞의 완성본대로 다음 파일을 작성합니다.

```text
migrations/000003_create_settlement_tables.up.sql
migrations/000003_create_settlement_tables.down.sql
```

Docker PostgreSQL을 실행합니다.

```bash
docker compose up -d postgres
docker compose ps
```

기존 Payment와 Ledger Migration이 적용된 DB에서 Settlement Migration을 적용합니다.

```bash
docker compose exec -T postgres psql -U stablepay -d stablepay < migrations/000003_create_settlement_tables.up.sql
```

테이블과 제약을 확인합니다.

```bash
docker compose exec -T postgres psql -U stablepay -d stablepay -c "\d settlement_batches"
docker compose exec -T postgres psql -U stablepay -d stablepay -c "\d settlement_items"
```

### Step 2. `settlement.go` 상태 확장

앞의 `settlement.go` 전체 코드로 교체합니다. Day22의 `DRAFT`에 READY부터 CANCELED까지 상태가 추가됩니다.

### Step 3. Repository 작성

다음 파일을 새로 만듭니다.

```text
internal/settlement/repository.go
```

앞의 전체 코드에는 후보 조회, Batch·Items transaction 저장, 현재 상태 조건부 갱신이 모두 포함되어 있습니다.

### Step 4. Service 작성

다음 파일을 새로 만듭니다.

```text
internal/settlement/service.go
```

Service는 `Store` interface를 통해 Repository 구현에 직접 결합하지 않고, 단위 테스트에서는 `fakeStore`를 받을 수 있습니다.

### Step 5. 테스트 작성

```text
internal/settlement/service_test.go
internal/settlement/repository_test.go
```

`service_test.go`는 DB 없이 흐름과 상태 규칙을 검증합니다. `repository_test.go`는 `TEST_DATABASE_URL`이 있을 때 실제 PostgreSQL을 검증합니다.

### Step 6. Go 코드 전체 정리

```bash
gofmt -w ./internal/settlement
```

### Step 7. 단위 테스트

```bash
go test ./internal/settlement -run 'TestCalculator|TestService|Example_statusFlow' -v
```

### Step 8. Repository 통합 테스트

```bash
TEST_DATABASE_URL="postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" \
go test ./internal/settlement -run TestRepositorySettlementFlow -v
```

다음 세 흐름이 통과해야 합니다.

```text
미정산 후보를 조회해 Batch와 Items를 저장한다.
같은 Ledger Entry는 다른 Batch에 중복 저장되지 않는다.
예상한 현재 상태일 때만 상태를 변경한다.
```

### Step 9. 전체 회귀 테스트

```bash
go test ./...
```

## 실습 검증 SQL

```bash
docker compose exec -T postgres psql -U stablepay -d stablepay -c "SELECT id, recipient_id, currency, total_amount, item_count, status FROM settlement_batches ORDER BY created_at DESC;"
```

```bash
docker compose exec -T postgres psql -U stablepay -d stablepay -c "SELECT batch_id, ledger_entry_id, amount, currency FROM settlement_items ORDER BY batch_id, ledger_entry_id;"
```

통합 테스트는 종료 시 테스트 데이터를 정리하므로 조회 결과가 0건이어도 정상입니다.

## 자주 만나는 오류

### `relation "settlement_batches" does not exist`

000003 up migration이 적용되지 않았습니다.

```bash
docker compose exec -T postgres psql -U stablepay -d stablepay < migrations/000003_create_settlement_tables.up.sql
```

### Repository 테스트가 모두 SKIP됨

`TEST_DATABASE_URL`이 없으면 의도적으로 통합 테스트를 건너뜁니다. 환경변수를 포함한 Step 8 명령으로 실행합니다.

### `duplicate key value violates unique constraint "idx_settlement_items_ledger_entry_id"`

이미 다른 Settlement Batch에 포함된 Ledger Entry를 다시 저장하려 했다는 뜻입니다. 중복 정산 방지가 정상 작동한 것입니다.

### `정산 상태가 변경되지 않았습니다`

Batch가 없거나 DB의 현재 상태가 Service가 예상한 상태와 다릅니다. 먼저 DB 상태를 확인해야 합니다.

### `settlement item 합계가 batch total amount와 다릅니다`

Items의 금액 합계와 Batch의 `TotalAmount`가 일치하지 않습니다. Calculator 결과를 임의로 수정하지 않았는지 확인합니다.

## 오늘 완료 기준

```text
1. settlement_batches와 settlement_items 역할을 설명할 수 있다.
2. ledger_entry_id UNIQUE가 중복 정산을 막는 이유를 설명할 수 있다.
3. Batch와 Items를 한 transaction으로 저장하는 이유를 설명할 수 있다.
4. Repository, Calculator, Service 책임을 구분할 수 있다.
5. 정상 상태 전이와 금지 상태 전이를 설명할 수 있다.
6. 단위 테스트, Repository 통합 테스트, 전체 테스트가 통과한다.
```

## 커밋 메시지

```bash
git add migrations/000003_create_settlement_tables.up.sql \
  migrations/000003_create_settlement_tables.down.sql \
  internal/settlement

git commit -m "feat: settlement DB 저장과 상태 전이 구현"
```

## 다음 작업

Day24에서는 Ledger와 Settlement DB를 비교해 누락, 중복, 금액 불일치를 찾는 Reconciliation과 Settlement 종합 검증을 구현합니다.
