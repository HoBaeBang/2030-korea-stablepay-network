package ledger

import (
	"context"
	"fmt"
)

// Service는 Ledger 도메인 규칙을 검증하고 실행한다.
type Service struct{}

// NewService는 Ledger Service 인스턴스를 만든다.
func NewService() *Service {
	return &Service{}
}

// ValidateTransaction은 원장 거래의 기본 규칙을 검증한다.
func (s *Service) ValidateTransaction(ctx context.Context, entries []Entry) error {
	if ctx == nil {
		return fmt.Errorf("context가 필요합니다")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if len(entries) < 2 {
		return fmt.Errorf("원장 거래는 최소 2개 이상의 항목이 필요합니다")
	}

	totals := make(map[string]int64)

	for _, entry := range entries {
		if entry.Amount <= 0 {
			return fmt.Errorf("원장 항목 금액은 0보다 커야 합니다")
		}

		if entry.Currency == "" {
			return fmt.Errorf("원장 항목 통화가 필요합니다")
		}

		switch entry.Direction {
		case EntryDirectionDebit:
			totals[entry.Currency] += entry.Amount
		case EntryDirectionCredit:
			totals[entry.Currency] -= entry.Amount
		default:
			return fmt.Errorf("알 수 없는 원장 항목 방향입니다: %s", entry.Direction)
		}
	}

	for currency, total := range totals {
		if total != 0 {
			return fmt.Errorf("원장 거래의 debit과 credit 합계가 일치하지 않습니다: %s", currency)
		}
	}
	return nil
}
