package settlement

import "github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"

// Status는 정산 묶음이 현재 어떤 처리 단계에 있는지를 나타낸다.
type Status string

const (
	// StatusDraft는 지급 가능 금액을 계산해 정산 묶음을 처음 만든 상태다.
	StatusDraft Status = "DRAFT"
)

// Candidate는 Ledger에서 조회한 정산 대상 후보를 나타낸다.
// 아직 Settlement에 포함된것은 아니며 Calculator검증을 통과해야 한다.
type Candidate struct {
	LedgerEntryID string
	RecipientID   string
	AccountType   ledger.AccountType
	Direction     ledger.EntryDirection
	Amount        int64
	Currency      string
}

// Batch는 같은 수취인과 토오하의 정산 대상들을 하나로 묶은 결과이다.
type Batch struct {
	ID          string
	RecipientID string
	Currency    string
	TotalAmount int64
	ItemCount   int
	Status      Status
}

// Item은 어떤 Ledger Entry가 정산 묶음에 포함되었는지를 나타내는 근거이다.
type Item struct {
	BatchID       string
	LedgerEntryID string
	Amount        int64
	Currency      string
}
