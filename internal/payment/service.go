package payment

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Store는 Service가 필요로 하는 저장소 동작만 정의한다.
// 실제 서비스에서는 Repository가 이 interface를 만족하고,
// 테스트에서는 fakeStore가 이 interface를 만족한다.
type Store interface {
	Create(ctx context.Context, p Payment) (*Payment, error)
	FindByID(ctx context.Context, id string) (*Payment, error)
	UpdateStatus(ctx context.Context, update StatusUpdate) (*Payment, error)
}

// Service는 payment 비즈니스 규칙을 처리한다.
type Service struct {
	store Store

	// now는 현재 시간을 가져오는 함수다.
	// 테스트에서 시간을 고정하기 위해 time.Now를 직접 호출하지 않고 함수 필드로 둔다.
	now func() time.Time
}

func NewService(store Store) *Service {
	return &Service{store: store, now: time.Now}
}

// CreatePayment는 invoice에 대한 payment를 생성한다.
func (s *Service) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*Payment, error) {
	// 앞뒤 공백 제거와 대문자 변환으로 입력값을 정규화한다.
	invoiceID := strings.TrimSpace(req.InvoiceID)
	currency := strings.ToUpper(strings.TrimSpace(req.Currency))

	if invoiceID == "" {
		return nil, fmt.Errorf("invoice id is required")
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}
	if currency != "USDC" {
		return nil, fmt.Errorf("currency must be USDC")
	}

	// 새 payment는 항상 PENDING 상태로 시작한다.
	p := Payment{
		ID:        newPaymentID(s.now()),
		InvoiceID: invoiceID,
		Amount:    req.Amount,
		Currency:  currency,
		Status:    StatusPending,
	}

	return s.store.Create(ctx, p)
}

// UpdatePaymentStatus는 payment 상태 전이 규칙을 검사한 뒤 상태를 변경한다.
func (s *Service) UpdatePaymentStatus(ctx context.Context, req UpdatePaymentStatusRequest) (*Payment, error) {
	paymentID := strings.TrimSpace(req.PaymentID)

	// 요청으로 들어온 status를 대문자로 정규화한다.
	// 예: finalized -> FINALIZED
	nextStatus := Status(strings.ToUpper(strings.TrimSpace(string(req.NextStatus))))

	if paymentID == "" {
		return nil, fmt.Errorf("payment id is required")
	}
	if !IsKnownStatus(nextStatus) {
		return nil, fmt.Errorf("unknown payment status: %s", nextStatus)
	}

	// 상태 전이를 검사하려면 먼저 현재 payment 상태를 알아야 한다.
	current, err := s.store.FindByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}

	// 현재 상태에서 다음 상태로 이동할 수 있는지 검사한다.
	if !CanTransition(current.Status, nextStatus) {
		return nil, fmt.Errorf("invalid payment status transition: %s -> %s", current.Status, nextStatus)
	}

	// 온체인 결제가 감지되었다면 어떤 transaction인지 식별할 hash가 필요하다.
	if nextStatus == StatusOnchainDetected && req.TransactionHash == nil {
		return nil, fmt.Errorf("transaction hash is required when status is ONCHAIN_DETECTED")
	}

	// FINALIZED가 아닌 상태에서는 finalized_at을 수정하지 않기 위해 nil로 둔다.
	var finalizedAt *time.Time
	if nextStatus == StatusFinalized {
		now := s.now()
		finalizedAt = &now
	}

	update := StatusUpdate{
		PaymentID:       paymentID,
		NextStatus:      nextStatus,
		TransactionHash: req.TransactionHash,
		FinalizedAt:     finalizedAt,
	}

	return s.store.UpdateStatus(ctx, update)
}

// IsKnownStatus는 입력된 status가 우리 시스템이 아는 상태인지 확인한다.
func IsKnownStatus(status Status) bool {
	switch status {
	case StatusPending, StatusOnchainDetected, StatusFinalized, StatusSettled, StatusFailed:
		return true
	default:
		return false
	}
}

// CanTransition은 현재 상태에서 다음 상태로 이동할 수 있는지 검사한다.
func CanTransition(from Status, to Status) bool {
	switch from {
	case StatusPending:
		return to == StatusOnchainDetected || to == StatusFailed
	case StatusOnchainDetected:
		return to == StatusFinalized || to == StatusFailed
	case StatusFinalized:
		return to == StatusSettled
	default:
		// SETTLED, FAILED 같은 종료 상태에서는 더 이상 상태 변경을 허용하지 않는다.
		return false
	}
}

// newPaymentID는 현재 시간을 이용해 학습용 payment id를 만든다.
// 실무에서는 UUID, ULID, Snowflake 같은 id 전략을 고려할 수 있다.
func newPaymentID(t time.Time) string {
	return fmt.Sprintf("pay_%d", t.UnixNano())
}
