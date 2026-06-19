package ledger

import (
	"context"
	"database/sql"
	"fmt"
)

// Repository는 Ledger 데이터를 DB에 저장하고 조회하는 경계이다.
type Repository struct {
	db *sql.DB
}

// NewRepository는 Ledger Repository 인스턴스를 만든다.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateTransaction(ctx context.Context, tx Transaction, entries []Entry) error {
	if ctx == nil {
		return fmt.Errorf("context가 필요합니다.")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if err := validateTransaction(tx); err != nil {
		return err
	}

	if err := validateEntries(tx, entries); err != nil {
		return err
	}

	if r.db == nil {
		return fmt.Errorf("ledger repository db가 필요합니다")
	}

	sqlTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("원장 저장 transaction 시작 실패: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = sqlTx.Rollback()
		}
	}()

	const insertTransactionQuery = `
		INSERT INTO ledger_transactions (id, reference_type, reference_id, idempotency_key) 
		VALUES ($1, $2, $3, $4)
	`

	if _, err := sqlTx.ExecContext(
		ctx,
		insertTransactionQuery,
		tx.ID,
		tx.ReferenceType,
		tx.ReferenceID,
		tx.IdempotencyKey,
	); err != nil {
		return fmt.Errorf("원장 거래 저장 실패: %w", err)
	}

	const insertEntryQuery = `
		INSERT INTO ledger_entries (id, transaction_id, account_id, direction, amount, currency)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	for _, entry := range entries {
		if _, err := sqlTx.ExecContext(
			ctx,
			insertEntryQuery,
			entry.ID,
			entry.TransactionID,
			entry.AccountID,
			entry.Direction,
			entry.Amount,
			entry.Currency,
		); err != nil {
			return fmt.Errorf("원장 항목 저장 실패: %w", err)
		}
	}

	if err := sqlTx.Commit(); err != nil {
		return fmt.Errorf("원장 저장 commit 실패: %w", err)
	}
	committed = true
	return nil

}

func validateTransaction(tx Transaction) error {
	if tx.ID == "" {
		return fmt.Errorf("원장 거래 id가 필요합니다")
	}

	if tx.ReferenceType == "" {
		return fmt.Errorf("원장 거래 reference type이 필요합니다.")
	}

	if tx.ReferenceID == "" {
		return fmt.Errorf("원장 거래 reference id가 필요합니다.")
	}

	if tx.IdempotencyKey == "" {
		return fmt.Errorf("원장 거래 idempotency key가 필요합니다")
	}

	return nil
}

func validateEntries(tx Transaction, entries []Entry) error {
	if len(entries) == 0 {
		return fmt.Errorf("원장 항목이 필요합니다")
	}

	for _, entry := range entries {
		if entry.ID == "" {
			return fmt.Errorf("원장 항목 id가 필요합니다")
		}
		if entry.TransactionID != tx.ID {
			return fmt.Errorf("원장 항목의 transaction id가 원장 거래 id와 다릅니다")
		}
		if entry.AccountID == "" {
			return fmt.Errorf("원장 항목 account id가 필요합니다")
		}
	}
	return nil
}
