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
		return nil, fmt.Errorf("정산 수취인 id가 필요합니다.")
	}
	if currency == "" {
		return nil, fmt.Errorf("정산 통화가 필요합니다.")
	}
	if r.db == nil {
		return nil, fmt.Errorf("settlement repository db가 필요합니다.")
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

	if _, err = sqlTx.ExecContext(
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

// UpdateBatchStatus는 DB 상태가 예상한 현재 상태와 같을 때만 다음 상태로 변경한다.
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

// validateContext는 context가 존재하고 아직 취소되거나 만료되지 않았는지 확인한다.
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
			return fmt.Errorf("settlement item의 ledger entry id가 필요합니다.")
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
