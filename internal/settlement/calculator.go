package settlement

import (
	"context"
	"fmt"
	"strings"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"
)

// Calculator는 Ledger 후보를 검증하고 Settlement Batch와 Items를 계산한다.
type Calculator struct {
}

// NewCalculator는 Settlement Calculator 인스턴스를 만든다.
func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) BuildBatch(
	ctx context.Context,
	batchID string,
	recipientID string,
	currency string,
	candidates []Candidate,
) (*Batch, []Item, error) {
	if ctx == nil {
		return nil, nil, fmt.Errorf("context가 필요합니다")
	}

	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	batchID = strings.TrimSpace(batchID)
	recipientID = strings.TrimSpace(recipientID)
	currency = strings.ToUpper(strings.TrimSpace(currency))

	if batchID == "" {
		return nil, nil, fmt.Errorf("settlement batch id가 필요합니다")
	}
	if recipientID == "" {
		return nil, nil, fmt.Errorf("settlement recipient id가 필요합니다")
	}
	if currency == "" {
		return nil, nil, fmt.Errorf("settlement currency가 필요합니다")
	}
	if len(candidates) == 0 {
		return nil, nil, fmt.Errorf("settlement candidate가 필요합니다")
	}

	seenEntryIDs := make(map[string]struct{}, len(candidates))
	items := make([]Item, 0, len(candidates))
	var totalAmount int64

	for _, candidate := range candidates {
		if candidate.LedgerEntryID == "" {
			return nil, nil, fmt.Errorf("ledger entry id가 필요합니다")
		}
		if _, exists := seenEntryIDs[candidate.LedgerEntryID]; exists {
			return nil, nil, fmt.Errorf("중복된 ledger entry입니다: %s", candidate.LedgerEntryID)
		}
		seenEntryIDs[candidate.LedgerEntryID] = struct{}{}
		if candidate.RecipientID != recipientID {
			return nil, nil, fmt.Errorf("다른 수취인의 settlement candidate가 포함되어 있습니다")
		}
		if candidate.AccountType != ledger.AccountTypeMerchantPending {
			return nil, nil, fmt.Errorf("merchant pending 계정만 정산할 수 있습니다")
		}
		if candidate.Direction != ledger.EntryDirectionCredit {
			return nil, nil, fmt.Errorf("credit 원장 항목만 정산할 수 있습니다")
		}
		if candidate.Amount <= 0 {
			return nil, nil, fmt.Errorf("정산 금액은 0보다 커야 합니다")
		}
		candidateCurrency := strings.ToUpper(strings.TrimSpace(candidate.Currency))
		if candidateCurrency != currency {
			return nil, nil, fmt.Errorf("다른 통화의 settlement candidate가 포함되어 있습니다")
		}
		totalAmount += candidate.Amount
		items = append(items, Item{
			BatchID:       batchID,
			LedgerEntryID: candidate.LedgerEntryID,
			Amount:        candidate.Amount,
			Currency:      currency,
		})
	}
	batch := &Batch{
		ID:          batchID,
		RecipientID: recipientID,
		Currency:    currency,
		TotalAmount: totalAmount,
		ItemCount:   len(items),
		Status:      StatusDraft,
	}
	return batch, items, nil
}
