package settlement

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"
)

// FindReconciliationSnapshot은 Settlement Batch, Items, 원본 Ledger 정보를 함께 조회한다.
func (r *Repository) FindReconciliationSnapshot(
	ctx context.Context,
	batchID string,
) (*ReconciliationSnapshot, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}

	batchID = strings.TrimSpace(batchID)
	if batchID == "" {
		return nil, fmt.Errorf("settlement batch id가 필요합니다")
	}
	if r.db == nil {
		return nil, fmt.Errorf("settlement repository db가 필요합니다")
	}

	const query = `
		SELECT
			sb.id,
			sb.recipient_id,
			sb.currency,
			sb.total_amount,
			sb.item_count,
			sb.status,
			si.ledger_entry_id,
			si.amount,
			si.currency,
			le.id,
			le.amount,
			le.currency,
			la.owner_id,
			la.type,
			le.direction
		FROM settlement_batches sb
		JOIN settlement_items si ON si.batch_id = sb.id
		LEFT JOIN ledger_entries le ON le.id = si.ledger_entry_id
		LEFT JOIN ledger_accounts la ON la.id = le.account_id
		WHERE sb.id = $1
		ORDER BY si.ledger_entry_id
	`

	rows, err := r.db.QueryContext(ctx, query, batchID)
	if err != nil {
		return nil, fmt.Errorf("reconciliation snapshot 조회 실패: %w", err)
	}
	defer rows.Close()

	snapshot := &ReconciliationSnapshot{
		Items: make([]ReconciliationItem, 0),
	}
	initialized := false

	for rows.Next() {
		var batch Batch
		var item ReconciliationItem
		var ledgerEntryID sql.NullString
		var ledgerAmount sql.NullInt64
		var ledgerCurrency sql.NullString
		var ledgerRecipientID sql.NullString
		var ledgerAccountType sql.NullString
		var ledgerDirection sql.NullString

		if err := rows.Scan(
			&batch.ID,
			&batch.RecipientID,
			&batch.Currency,
			&batch.TotalAmount,
			&batch.ItemCount,
			&batch.Status,
			&item.LedgerEntryID,
			&item.SettlementAmount,
			&item.SettlementCurrency,
			&ledgerEntryID,
			&ledgerAmount,
			&ledgerCurrency,
			&ledgerRecipientID,
			&ledgerAccountType,
			&ledgerDirection,
		); err != nil {
			return nil, fmt.Errorf("reconciliation snapshot 변환 실패: %w", err)
		}

		if !initialized {
			snapshot.Batch = batch
			initialized = true
		}

		item.LedgerEntryFound = ledgerEntryID.Valid
		if ledgerAmount.Valid {
			item.LedgerAmount = ledgerAmount.Int64
		}
		if ledgerCurrency.Valid {
			item.LedgerCurrency = ledgerCurrency.String
		}
		if ledgerRecipientID.Valid {
			item.LedgerRecipientID = ledgerRecipientID.String
		}
		if ledgerAccountType.Valid {
			item.LedgerAccountType = ledger.AccountType(ledgerAccountType.String)
		}
		if ledgerDirection.Valid {
			item.LedgerDirection = ledger.EntryDirection(ledgerDirection.String)
		}

		snapshot.Items = append(snapshot.Items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reconciliation snapshot 순회 실패: %w", err)
	}
	if !initialized {
		return nil, fmt.Errorf("settlement batch를 찾을 수 없습니다: %s", batchID)
	}

	return snapshot, nil
}
