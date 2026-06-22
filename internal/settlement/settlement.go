package settlement

import "github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"

// Status는 정산 묶음이 현재 어떤 처리 단계에 있는지를 나타낸다.
type Status string

const (
	// StatusDraft는 정산 후보를 계산해 묶음을 처음 만든 상태다.
	StatusDraft Status = "DRAFT"
	// StatusReady는 정산 항목 검증이 끝나 승인할 수 있는 상태다.
	StatusReady Status = "READY"
	// StatusApproved는 내부 정책이나 관리자의 지급 승인이 끝난 상태다.
	StatusApproved Status = "APPROVED"
	// StatusProcessing은 실제 지급 요청을 처리 중인 상태다.
	StatusProcessing Status = "PROCESSING"
	// StatusPaid는 지급 완료가 확인된 최종 상태다.
	StatusPaid Status = "PAID"
	// StatusFailed는 지급 시도가 실패해 재시도나 취소가 필요한 상태다.
	StatusFailed Status = "FAILED"
	// StatusCanceled는 지급 전에 정산 처리를 취소한 최종 상태다.
	StatusCanceled Status = "CANCELED"
)

// Candidate는 Ledger에서 조회한 정산 대상 후보를 나타낸다.
// 아직 Settlement에 포함된 것은 아니며 Calculator 검증을 통과해야 한다.
type Candidate struct {
	LedgerEntryID string
	RecipientID   string
	AccountType   ledger.AccountType
	Direction     ledger.EntryDirection
	Amount        int64
	Currency      string
}

// Batch는 같은 수취인과 통화의 정산 대상들을 하나로 묶은 결과다.
type Batch struct {
	ID          string
	RecipientID string
	Currency    string
	TotalAmount int64
	ItemCount   int
	Status      Status
}

// Item은 어떤 Ledger Entry가 정산 묶음에 포함되었는지를 나타내는 근거다.
type Item struct {
	BatchID       string
	LedgerEntryID string
	Amount        int64
	Currency      string
}
