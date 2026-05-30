package payment

import "time"

// Status는 payment가 현재 어떤 단계에 있는지 나타내는 전용 타입이다.
// string을 그대로 쓰지 않고 Status 타입을 만든 이유는 payment 상태라는 의미를 코드에 남기기 위해서다.
type Status string

const (
	// StatusPending은 payment가 생성되었지만 아직 온체인 결제가 감지되지 않은 상태다.
	StatusPending Status = "PENDING"

	// StatusOnchainDetected는 블록체인에서 결제 transaction이 감지된 상태다.
	StatusOnchainDetected Status = "ONCHAIN_DETECTED"

	// StatusFinalized는 블록체인 finality가 충분히 확보되어 결제가 확정된 상태다.
	StatusFinalized Status = "FINALIZED"

	// StatusSettled는 가맹점에게 정산까지 완료된 상태다.
	StatusSettled Status = "SETTLED"

	// StatusFailed는 결제가 실패한 상태다.
	StatusFailed Status = "FAILED"
)

// Payment는 invoice에 대해 실제로 발생한 결제 시도 또는 결제 결과를 나타낸다.
type Payment struct {
	ID        string
	InvoiceID string
	Amount    int64
	Currency  string
	Status    Status

	// TransactionHash는 온체인 transaction hash다.
	// PENDING 상태에서는 아직 없을 수 있으므로 *string으로 둔다.
	TransactionHash *string

	// FinalizedAt은 결제가 FINALIZED 상태가 된 시각이다.
	// 아직 확정되지 않았을 수 있으므로 *time.Time으로 둔다.
	FinalizedAt *time.Time

	CreatedAt time.Time
}

// CreatePaymentRequest는 payment 생성을 위해 service가 받는 입력 값이다.
type CreatePaymentRequest struct {
	InvoiceID string
	Amount    int64
	Currency  string
}

// UpdatePaymentStatusRequest는 payment 상태 변경을 위해 service가 받는 입력 값이다.
type UpdatePaymentStatusRequest struct {
	PaymentID  string
	NextStatus Status

	// ONCHAIN_DETECTED 상태로 변경할 때는 transaction hash가 필요하다.
	TransactionHash *string
}
