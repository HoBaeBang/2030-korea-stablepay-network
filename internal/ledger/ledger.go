package ledger

import "time"

// AccountType은 원장에서 계정의 역할을 구분한다.
type AccountType string

const (
	AccountTypeCustomer        AccountType = "CUSTOMER"
	AccountTypeMerchantPending AccountType = "MERCHANT_PENDING"
	AccountTypePlatformFee     AccountType = "PLATFORM_FEE"
)

// EntryDirection은 원장 항목의 방향을 나타낸다.
type EntryDirection string

const (
	EntryDirectionDebit  EntryDirection = "DEBIT"
	EntryDirectionCredit EntryDirection = "CREDIT"
)

// Account는 원장에서 돈이 기록되는 주체이다.
type Account struct {
	ID       string
	Type     AccountType
	OwnerID  string
	Currency string
	CreateAt time.Time
}

// Transaction은 여러 Entry를 하나의 원장 거래로 묶는다.
type Transaction struct {
	ID             string
	ReferenceType  string
	ReferenceId    string
	IdempotencyKey string
	CreatedAt      time.Time
}

// Entry는 하나의 원장 거래 안에서 발생한 돈의 이동 한 줄이다.
type Entry struct {
	ID            string
	TransactionId string
	AccountId     string
	Direction     EntryDirection
	Amount        int64
	Currency      string
	CreateAt      time.Time
}
