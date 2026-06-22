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
