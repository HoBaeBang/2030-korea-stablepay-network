package ledger

import (
	"context"
	"fmt"
)

// Store는 Service가 필요로 하는 Ledger 저장 동작을 정의한다.
type Store interface {
	CreateTransaction(ctx context.Context, tx Transaction, entries []Entry) error
}

// Service는 Ledger 도메인 규칙을 검증하고 실행한다.
type Service struct {
	store Store
}

// NewService는 Ledger Service 인스턴스를 만든다.
func NewService(store Store) *Service {
	return &Service{store: store}
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

// RecordTransaction은 원장 거래를 검증한 뒤 저장소에 기록한다.
func (s *Service) RecordTransaction(ctx context.Context, tx Transaction, entries []Entry) error {
	if err := s.ValidateTransaction(ctx, entries); err != nil {
		return err
	}

	if s.store == nil {
		return fmt.Errorf("ledger store가 필요합니다")
	}

	return s.store.CreateTransaction(ctx, tx, entries)
}
