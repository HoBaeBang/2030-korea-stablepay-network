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
