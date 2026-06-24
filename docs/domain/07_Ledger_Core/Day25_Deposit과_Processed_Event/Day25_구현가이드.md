# Day 25 구현가이드 - Deposit과 Processed Event

관련 Jira: [SPN-42](https://aslan0.atlassian.net/browse/SPN-42)  
관련 Confluence:
[Day25 메인](https://aslan0.atlassian.net/wiki/spaces/SPN/pages/14876674/Deposit+Processed+Event),
[구현가이드](https://aslan0.atlassian.net/wiki/spaces/SPN/pages/14909442),
[실습산출물](https://aslan0.atlassian.net/wiki/spaces/SPN/pages/14942209)

Day25에서는 온체인 입금 이벤트를 내부 Ledger에 안전하게 반영하는 첫 흐름을 만듭니다.

이번 문서는 Day24 이후 정한 Flow-first 방식으로 작성합니다. 먼저 실제 운영 시나리오와 대표 진입점 함수를 보고, 그 흐름 안에서 왜 Deposit, Processed Event, Repository, Ledger Entry가 필요한지 연결합니다.

## 오늘의 실제 운영 시나리오

```text
Blockchain Event Indexer가 USDC Transfer event를 읽었다.
그 event는 우리 서비스가 관리하는 입금 주소로 들어온 입금이다.
같은 event는 RPC 재시도, worker 재시작, polling 범위 중복 때문에 여러 번 들어올 수 있다.

목표:
같은 온체인 event가 여러 번 들어와도
deposits row와 ledger credit은 한 번만 생성한다.
```

오늘은 실제 Ethereum RPC를 붙이지 않습니다.

```text
오늘 구현:
이미 감지된 ChainTransferEvent를 DepositProcessor가 처리한다.

오늘 구현하지 않음:
실제 RPC polling
block checkpoint
reorg 대응
withdrawal
HTTP API
```

## 대표 진입점 함수와 전체 호출 흐름

오늘 가장 먼저 이해해야 하는 코드는 아래 한 줄입니다.

```go
result, err := processor.ProcessEvent(ctx, event)
```

전체 흐름은 아래처럼 읽습니다.

```text
Indexer 또는 테스트
-> DepositProcessor.ProcessEvent(ctx, event)
-> event 필수값 검증
-> event_key 생성: chain + tx_hash + log_index
-> Deposit row 생성 준비
-> Ledger Transaction/Entry 생성 준비
-> Repository.SaveDepositCredit(...)
   -> processed_events 먼저 INSERT
   -> 이미 있으면 duplicate result 반환
   -> 없으면 deposits INSERT
   -> ledger_transactions INSERT
   -> ledger_entries INSERT
   -> processed_events PROCESSED로 변경
-> ProcessResult 반환
```

핵심은 `processed_events`를 먼저 저장한다는 점입니다.

```text
processed_events insert 성공
= 이 event는 이번 worker가 처음 처리한다.

processed_events unique 충돌
= 이미 처리했거나 처리 중인 event다.
  deposits와 ledger_entries를 다시 만들면 안 된다.
```

## 왜 이 기능이 필요한가

온체인 이벤트 처리에서는 같은 이벤트가 여러 번 들어오는 것이 정상입니다.

```text
RPC 호출 재시도
Indexer worker 재시작
마지막 처리 block부터 넉넉하게 다시 polling
여러 worker가 같은 구간을 읽음
장애 복구를 위해 과거 event 재처리
```

이때 중복 방어가 없으면 내부 Ledger에 credit이 두 번 생깁니다.

```text
온체인 실제 입금:
100 USDC 1번

내부 Ledger 오류:
100 USDC credit 2번

결과:
사용자 잔액 또는 정산 가능 금액이 실제보다 커진다.
```

그래서 Day25의 핵심은 Deposit 자체보다 **event idempotency**, 즉 같은 이벤트를 여러 번 처리해도 결과가 한 번만 반영되게 만드는 것입니다.

## 기존 코드 어디에 붙는가

Day25는 기존 Ledger 이후에 붙습니다.

```text
Payment FINALIZED
-> Ledger 기록

Settlement
-> Ledger 중 지급 가능한 CREDIT을 묶음

Reconciliation
-> Settlement와 Ledger가 맞는지 검사

Deposit
-> 온체인 입금 event를 Ledger CREDIT으로 반영
```

Day25가 끝나면 이후 Day26 Event Indexer가 아래처럼 연결됩니다.

```text
Event Indexer
-> ChainTransferEvent 생성
-> DepositProcessor.ProcessEvent(ctx, event)
-> Deposit + Ledger 반영
```

## 오늘 만들 것

오늘 새로 작성할 파일:

```text
migrations/000004_create_deposit_tables.up.sql
migrations/000004_create_deposit_tables.down.sql
internal/deposit/deposit.go
internal/deposit/service.go
internal/deposit/repository.go
internal/deposit/service_test.go
internal/deposit/repository_test.go
```

오늘 수정할 파일:

```text
internal/ledger/ledger.go
```

`ledger.go`에는 입금 반영을 위해 `DEPOSIT_CLEARING` 계정 타입을 추가합니다.

```text
CUSTOMER
= 사용자 내부 잔액 계정

DEPOSIT_CLEARING
= 온체인에서 들어온 돈을 내부 Ledger로 반영할 때 반대편 debit으로 사용하는 정리 계정
```

Ledger는 항상 debit과 credit 합계가 같아야 하므로, 입금 credit만 단독으로 저장하지 않습니다.

## 핵심 용어

| 영어 | 한글 의미 | 프로젝트에서의 뜻 |
| --- | --- | --- |
| Deposit | 입금 | 온체인에서 들어온 돈을 내부 DB와 Ledger에 반영한 기록 |
| Processed Event | 처리 완료 이벤트 | 이미 처리한 blockchain event를 기록해 중복 반영을 막는 row |
| Event Key | 이벤트 고유 키 | `chain + tx_hash + log_index` 조합 |
| Idempotency | 멱등성 | 같은 요청/event를 여러 번 처리해도 결과가 한 번만 반영되는 성질 |
| Clearing Account | 정리 계정 | 외부 온체인 자금 유입을 내부 Ledger의 반대편 entry로 맞추는 계정 |
| Processor | 처리기 | 감지된 event를 검증하고 저장 흐름을 조정하는 service |

## 핵심 다이어그램

![Day25 Deposit과 Processed Event 흐름](https://raw.githubusercontent.com/HoBaeBang/2030-korea-stablepay-network/main/docs/confluence/diagrams/spn42-day25-deposit-processed-event-flow.png)

## 확인할 기존 파일

```text
internal/ledger/ledger.go
internal/ledger/service.go
internal/ledger/repository.go
migrations/000002_create_ledger_core_tables.up.sql
docs/domain/04_블록체인_이벤트_인덱서/Blockchain_Event_Indexer_실습산출물.md
```

특히 Ledger Service의 원칙을 다시 확인합니다.

```text
Ledger transaction은 최소 2개 이상의 entry가 필요하다.
Debit 합계와 Credit 합계가 같아야 한다.
```

입금은 사용자 입장에서는 돈이 증가하는 일이지만, Ledger에서는 아래처럼 두 줄로 기록합니다.

```text
DEBIT  deposit clearing account  100 USDC
CREDIT customer account          100 USDC
```

## 실습 전 확인

프로젝트 루트에서 실행합니다.

```bash
git status --short
go test ./internal/ledger ./internal/settlement
docker compose ps
```

PostgreSQL 통합 테스트를 실행하려면 아래 환경변수를 사용합니다.

```bash
TEST_DATABASE_URL="postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable"
```

## Step 1. Ledger Account Type을 추가한다

파일 위치:

```text
internal/ledger/ledger.go
```

`AccountType` 상수에 아래 값을 추가합니다.

```go
const (
	AccountTypeCustomer        AccountType = "CUSTOMER"
	AccountTypeMerchantPending AccountType = "MERCHANT_PENDING"
	AccountTypePlatformFee     AccountType = "PLATFORM_FEE"
	AccountTypeDepositClearing AccountType = "DEPOSIT_CLEARING"
)
```

### 왜 필요한가

입금 event를 Ledger에 반영하려면 사용자 계정에 credit을 기록해야 합니다. 그런데 Ledger 거래는 한쪽 entry만으로는 저장하지 않습니다.

```text
외부 온체인 자금 유입을 내부 Ledger에 반영하는 반대편 계정
= DEPOSIT_CLEARING
```

이 계정은 실제 고객 잔액 계정이 아니라, 내부 회계 균형을 맞추기 위한 시스템성 계정입니다.

## Step 2. Deposit DB migration을 작성한다

파일 위치:

```text
migrations/000004_create_deposit_tables.up.sql
migrations/000004_create_deposit_tables.down.sql
```

<details>
<summary>000004_create_deposit_tables.up.sql 전체 보기</summary>

```sql
CREATE TABLE processed_events
(
    event_key    text primary key,
    chain        text        not null,
    tx_hash      text        not null,
    log_index    integer     not null,
    status       text        not null check ( status in ('PROCESSING', 'PROCESSED', 'FAILED') ),
    created_at   timestamptz not null default now(),
    processed_at timestamptz
);

CREATE UNIQUE INDEX idx_processed_events_chain_tx_log
    ON processed_events (chain, tx_hash, log_index);

CREATE TABLE deposits
(
    id                    text primary key,
    event_key             text        not null unique references processed_events (event_key),
    chain                 text        not null,
    tx_hash               text        not null,
    log_index             integer     not null,
    block_number          bigint      not null,
    from_address          text        not null,
    to_address            text        not null,
    token_address         text        not null,
    recipient_id          text        not null,
    recipient_account_id  text        not null references ledger_accounts (id),
    clearing_account_id   text        not null references ledger_accounts (id),
    amount                bigint      not null check ( amount > 0 ),
    currency              text        not null,
    status                text        not null check ( status in ('DETECTED', 'CREDITED', 'FAILED') ),
    ledger_transaction_id text        references ledger_transactions (id),
    created_at            timestamptz not null default now(),
    updated_at            timestamptz not null default now()
);

CREATE INDEX idx_deposits_recipient_status
    ON deposits (recipient_id, status);

CREATE INDEX idx_deposits_tx_hash
    ON deposits (tx_hash);
```

</details>

<details>
<summary>000004_create_deposit_tables.down.sql 전체 보기</summary>

```sql
DROP TABLE IF EXISTS deposits;
DROP TABLE IF EXISTS processed_events;
```

</details>

### 왜 `processed_events`와 `deposits`를 분리하는가

```text
processed_events
= 이 blockchain event를 처리했는지 여부를 판단하는 멱등성 장치

deposits
= 이 event가 입금으로 해석되어 내부 서비스에 반영된 업무 기록
```

모든 processed event가 deposit일 필요는 없습니다. 나중에는 withdrawal confirmation, payment event, contract event도 같은 processed event 패턴을 쓸 수 있습니다.

## Step 3. Deposit 도메인 타입을 작성한다

파일 위치:

```text
internal/deposit/deposit.go
```

<details>
<summary>deposit.go 전체 보기</summary>

```go
package deposit

import (
	"fmt"
	"strings"
)

type Status string

const (
	StatusDetected Status = "DETECTED"
	StatusCredited Status = "CREDITED"
	StatusFailed   Status = "FAILED"
)

type ProcessedEventStatus string

const (
	ProcessedEventStatusProcessing ProcessedEventStatus = "PROCESSING"
	ProcessedEventStatusProcessed  ProcessedEventStatus = "PROCESSED"
	ProcessedEventStatusFailed     ProcessedEventStatus = "FAILED"
)

type ChainTransferEvent struct {
	Chain              string
	TxHash             string
	LogIndex           int
	BlockNumber        int64
	FromAddress        string
	ToAddress          string
	TokenAddress       string
	RecipientID        string
	RecipientAccountID string
	ClearingAccountID  string
	Amount             int64
	Currency           string
}

func (e ChainTransferEvent) EventKey() string {
	return fmt.Sprintf(
		"%s:%s:%d",
		strings.ToLower(strings.TrimSpace(e.Chain)),
		strings.ToLower(strings.TrimSpace(e.TxHash)),
		e.LogIndex,
	)
}

type Deposit struct {
	ID                  string
	EventKey            string
	Chain               string
	TxHash              string
	LogIndex            int
	BlockNumber         int64
	FromAddress         string
	ToAddress           string
	TokenAddress        string
	RecipientID         string
	RecipientAccountID  string
	ClearingAccountID   string
	Amount              int64
	Currency            string
	Status              Status
	LedgerTransactionID string
}

type ProcessedEvent struct {
	EventKey string
	Chain    string
	TxHash   string
	LogIndex int
	Status   ProcessedEventStatus
}

type ProcessResult struct {
	EventKey  string
	DepositID string
	Duplicate bool
	Status    Status
}
```

</details>

### 타입이 전체 흐름에서 쓰이는 위치

```text
ChainTransferEvent
= Indexer가 읽어온 원본 event를 내부 처리용으로 바꾼 입력값

Deposit
= event를 입금 업무 기록으로 저장할 값

ProcessedEvent
= event 중복 처리 여부를 판단할 값

ProcessResult
= 호출자에게 "처리됨" 또는 "이미 처리됨"을 알려줄 결과
```

## Step 4. DepositProcessor 흐름을 먼저 작성한다

파일 위치:

```text
internal/deposit/service.go
```

이 파일이 Day25의 중심입니다. Repository와 DB보다 먼저 `ProcessEvent`의 흐름을 읽습니다.

<details>
<summary>service.go 전체 보기</summary>

```go
package deposit

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"
)

type Store interface {
	SaveDepositCredit(
		ctx context.Context,
		event ProcessedEvent,
		deposit Deposit,
		tx ledger.Transaction,
		entries []ledger.Entry,
	) (bool, error)
}

type Processor struct {
	store Store
}

func NewProcessor(store Store) *Processor {
	return &Processor{store: store}
}

func (p *Processor) ProcessEvent(
	ctx context.Context,
	event ChainTransferEvent,
) (*ProcessResult, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	if p.store == nil {
		return nil, fmt.Errorf("deposit store가 필요합니다")
	}
	if err := validateEvent(event); err != nil {
		return nil, err
	}

	eventKey := event.EventKey()
	depositID := buildDepositID(event)
	ledgerTxID := "led_tx_" + depositID

	processedEvent := ProcessedEvent{
		EventKey: eventKey,
		Chain:    strings.ToUpper(strings.TrimSpace(event.Chain)),
		TxHash:   strings.ToLower(strings.TrimSpace(event.TxHash)),
		LogIndex: event.LogIndex,
		Status:   ProcessedEventStatusProcessing,
	}

	deposit := Deposit{
		ID:                  depositID,
		EventKey:            eventKey,
		Chain:               processedEvent.Chain,
		TxHash:              processedEvent.TxHash,
		LogIndex:            event.LogIndex,
		BlockNumber:         event.BlockNumber,
		FromAddress:         strings.ToLower(strings.TrimSpace(event.FromAddress)),
		ToAddress:           strings.ToLower(strings.TrimSpace(event.ToAddress)),
		TokenAddress:        strings.ToLower(strings.TrimSpace(event.TokenAddress)),
		RecipientID:         strings.TrimSpace(event.RecipientID),
		RecipientAccountID:  strings.TrimSpace(event.RecipientAccountID),
		ClearingAccountID:   strings.TrimSpace(event.ClearingAccountID),
		Amount:              event.Amount,
		Currency:            strings.ToUpper(strings.TrimSpace(event.Currency)),
		Status:              StatusCredited,
		LedgerTransactionID: ledgerTxID,
	}

	tx := ledger.Transaction{
		ID:             ledgerTxID,
		ReferenceType:  "DEPOSIT",
		ReferenceID:    deposit.ID,
		IdempotencyKey: "deposit:" + eventKey,
	}

	entries := []ledger.Entry{
		{
			ID:            "led_entry_" + deposit.ID + "_debit",
			TransactionID: tx.ID,
			AccountID:     deposit.ClearingAccountID,
			Direction:     ledger.EntryDirectionDebit,
			Amount:        deposit.Amount,
			Currency:      deposit.Currency,
		},
		{
			ID:            "led_entry_" + deposit.ID + "_credit",
			TransactionID: tx.ID,
			AccountID:     deposit.RecipientAccountID,
			Direction:     ledger.EntryDirectionCredit,
			Amount:        deposit.Amount,
			Currency:      deposit.Currency,
		},
	}

	created, err := p.store.SaveDepositCredit(ctx, processedEvent, deposit, tx, entries)
	if err != nil {
		return nil, fmt.Errorf("deposit event 저장 실패: %w", err)
	}
	if !created {
		return &ProcessResult{
			EventKey:  eventKey,
			DepositID: depositID,
			Duplicate: true,
			Status:    StatusCredited,
		}, nil
	}

	return &ProcessResult{
		EventKey:  eventKey,
		DepositID: depositID,
		Duplicate: false,
		Status:    StatusCredited,
	}, nil
}

func validateContext(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("context가 필요합니다")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func validateEvent(event ChainTransferEvent) error {
	if strings.TrimSpace(event.Chain) == "" {
		return fmt.Errorf("chain이 필요합니다")
	}
	if strings.TrimSpace(event.TxHash) == "" {
		return fmt.Errorf("tx hash가 필요합니다")
	}
	if event.LogIndex < 0 {
		return fmt.Errorf("log index는 0 이상이어야 합니다")
	}
	if event.BlockNumber <= 0 {
		return fmt.Errorf("block number는 0보다 커야 합니다")
	}
	if strings.TrimSpace(event.FromAddress) == "" {
		return fmt.Errorf("from address가 필요합니다")
	}
	if strings.TrimSpace(event.ToAddress) == "" {
		return fmt.Errorf("to address가 필요합니다")
	}
	if strings.TrimSpace(event.TokenAddress) == "" {
		return fmt.Errorf("token address가 필요합니다")
	}
	if strings.TrimSpace(event.RecipientID) == "" {
		return fmt.Errorf("recipient id가 필요합니다")
	}
	if strings.TrimSpace(event.RecipientAccountID) == "" {
		return fmt.Errorf("recipient account id가 필요합니다")
	}
	if strings.TrimSpace(event.ClearingAccountID) == "" {
		return fmt.Errorf("clearing account id가 필요합니다")
	}
	if event.Amount <= 0 {
		return fmt.Errorf("amount는 0보다 커야 합니다")
	}
	if strings.TrimSpace(event.Currency) == "" {
		return fmt.Errorf("currency가 필요합니다")
	}
	return nil
}

func buildDepositID(event ChainTransferEvent) string {
	chain := strings.ToLower(strings.TrimSpace(event.Chain))
	txHash := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(event.TxHash), "0x"))
	return "dep_" + chain + "_" + txHash + "_" + strconv.Itoa(event.LogIndex)
}
```

</details>

### 이 흐름에서 interface가 필요한 이유

`Processor`는 업무 흐름을 담당합니다.

```text
event 검증
deposit id 생성
ledger transaction 생성
duplicate 결과 해석
```

`Repository`는 DB 저장을 담당합니다.

```text
processed_events INSERT
deposits INSERT
ledger_transactions INSERT
ledger_entries INSERT
commit 또는 rollback
```

그래서 `Processor`는 구체적인 PostgreSQL 코드를 직접 알 필요가 없습니다. 대신 아래 능력을 가진 저장소만 요구합니다.

```go
type Store interface {
	SaveDepositCredit(...) (bool, error)
}
```

테스트에서는 fake store를 넣고, 실제 실행에서는 repository를 넣습니다.

## Step 5. Processor 단위 테스트를 작성한다

파일 위치:

```text
internal/deposit/service_test.go
```

<details>
<summary>service_test.go 전체 보기</summary>

```go
package deposit

import (
	"context"
	"errors"
	"testing"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"
)

type fakeStore struct {
	created bool
	err     error
	calls   int
	event   ProcessedEvent
	deposit Deposit
	tx      ledger.Transaction
	entries []ledger.Entry
}

func (f *fakeStore) SaveDepositCredit(
	ctx context.Context,
	event ProcessedEvent,
	deposit Deposit,
	tx ledger.Transaction,
	entries []ledger.Entry,
) (bool, error) {
	f.calls++
	f.event = event
	f.deposit = deposit
	f.tx = tx
	f.entries = append([]ledger.Entry(nil), entries...)
	return f.created, f.err
}

func validEvent() ChainTransferEvent {
	return ChainTransferEvent{
		Chain:              "ethereum-sepolia",
		TxHash:             "0xabc123",
		LogIndex:           7,
		BlockNumber:        1_234_567,
		FromAddress:        "0xsender",
		ToAddress:          "0xstablepay",
		TokenAddress:       "0xusdc",
		RecipientID:        "customer_1",
		RecipientAccountID: "acct_customer_1",
		ClearingAccountID:  "acct_deposit_clearing_usdc",
		Amount:             10_000_000,
		Currency:           "USDC",
	}
}

func TestProcessorProcessEvent(t *testing.T) {
	t.Run("새 입금 이벤트를 Deposit과 Ledger로 저장한다", func(t *testing.T) {
		store := &fakeStore{created: true}
		processor := NewProcessor(store)

		result, err := processor.ProcessEvent(context.Background(), validEvent())
		if err != nil {
			t.Fatalf("입금 event 처리가 성공해야 합니다: %v", err)
		}
		if result.Duplicate {
			t.Fatal("처음 처리한 event는 duplicate가 아니어야 합니다")
		}
		if result.Status != StatusCredited {
			t.Fatalf("입금은 CREDITED 상태여야 합니다: %s", result.Status)
		}
		if store.calls != 1 {
			t.Fatalf("store가 한 번 호출되어야 합니다: %d", store.calls)
		}
		if store.event.EventKey != "ethereum-sepolia:0xabc123:7" {
			t.Fatalf("event key가 예상과 다릅니다: %s", store.event.EventKey)
		}
		if len(store.entries) != 2 {
			t.Fatalf("ledger entry는 debit/credit 2개여야 합니다: %+v", store.entries)
		}
	})

	t.Run("이미 처리한 event는 duplicate result를 반환한다", func(t *testing.T) {
		store := &fakeStore{created: false}
		result, err := NewProcessor(store).ProcessEvent(context.Background(), validEvent())
		if err != nil {
			t.Fatalf("중복 event는 error가 아니라 result로 반환해야 합니다: %v", err)
		}
		if !result.Duplicate {
			t.Fatal("이미 처리한 event는 duplicate여야 합니다")
		}
	})

	t.Run("필수값이 없으면 store를 호출하지 않는다", func(t *testing.T) {
		store := &fakeStore{created: true}
		event := validEvent()
		event.TxHash = ""

		if _, err := NewProcessor(store).ProcessEvent(context.Background(), event); err == nil {
			t.Fatal("tx hash가 없으면 실패해야 합니다")
		}
		if store.calls != 0 {
			t.Fatal("검증 실패 시 store를 호출하면 안 됩니다")
		}
	})

	t.Run("store 실패는 error로 반환한다", func(t *testing.T) {
		store := &fakeStore{err: errors.New("db down")}
		if _, err := NewProcessor(store).ProcessEvent(context.Background(), validEvent()); err == nil {
			t.Fatal("store 실패는 error여야 합니다")
		}
	})
}
```

</details>

### 테스트가 먼저 보여주는 큰 그림

```text
정상 event
-> Deposit + Ledger 저장 요청
-> CREDITED result

중복 event
-> error가 아니라 Duplicate result

필수값 누락
-> 실행 불가 error

DB 실패
-> 실행 실패 error
```

이 구분은 Day24의 `error`와 `ReconciliationIssue` 구분과 비슷합니다.

```text
error
= 처리 자체를 할 수 없음

Duplicate result
= 처리는 성공했고, 이미 반영된 event임을 확인함
```

## Step 6. Repository를 작성한다

파일 위치:

```text
internal/deposit/repository.go
```

<details>
<summary>repository.go 전체 보기</summary>

```go
package deposit

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveDepositCredit(
	ctx context.Context,
	event ProcessedEvent,
	deposit Deposit,
	tx ledger.Transaction,
	entries []ledger.Entry,
) (bool, error) {
	if err := validateContext(ctx); err != nil {
		return false, err
	}
	if r.db == nil {
		return false, fmt.Errorf("deposit repository db가 필요합니다")
	}
	if len(entries) != 2 {
		return false, fmt.Errorf("deposit ledger entries는 debit/credit 2개여야 합니다")
	}

	sqlTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("deposit 저장 transaction 시작 실패: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = sqlTx.Rollback()
		}
	}()

	const insertEventQuery = `
		INSERT INTO processed_events (event_key, chain, tx_hash, log_index, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (event_key) DO NOTHING
	`

	result, err := sqlTx.ExecContext(
		ctx,
		insertEventQuery,
		event.EventKey,
		event.Chain,
		event.TxHash,
		event.LogIndex,
		ProcessedEventStatusProcessing,
	)
	if err != nil {
		return false, fmt.Errorf("processed event 저장 실패: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("processed event 저장 결과 확인 실패: %w", err)
	}
	if rowsAffected == 0 {
		if err := sqlTx.Commit(); err != nil {
			return false, fmt.Errorf("중복 event 확인 commit 실패: %w", err)
		}
		committed = true
		return false, nil
	}

	const insertDepositQuery = `
		INSERT INTO deposits
			(id, event_key, chain, tx_hash, log_index, block_number,
			 from_address, to_address, token_address,
			 recipient_id, recipient_account_id, clearing_account_id,
			 amount, currency, status, ledger_transaction_id)
		VALUES
			($1, $2, $3, $4, $5, $6,
			 $7, $8, $9,
			 $10, $11, $12,
			 $13, $14, $15, $16)
	`

	if _, err := sqlTx.ExecContext(
		ctx,
		insertDepositQuery,
		deposit.ID,
		deposit.EventKey,
		deposit.Chain,
		deposit.TxHash,
		deposit.LogIndex,
		deposit.BlockNumber,
		deposit.FromAddress,
		deposit.ToAddress,
		deposit.TokenAddress,
		deposit.RecipientID,
		deposit.RecipientAccountID,
		deposit.ClearingAccountID,
		deposit.Amount,
		deposit.Currency,
		deposit.Status,
		deposit.LedgerTransactionID,
	); err != nil {
		return false, fmt.Errorf("deposit 저장 실패: %w", err)
	}

	const insertTransactionQuery = `
		INSERT INTO ledger_transactions (id, reference_type, reference_id, idempotency_key)
		VALUES ($1, $2, $3, $4)
	`

	if _, err := sqlTx.ExecContext(
		ctx,
		insertTransactionQuery,
		tx.ID,
		tx.ReferenceType,
		tx.ReferenceID,
		tx.IdempotencyKey,
	); err != nil {
		return false, fmt.Errorf("deposit ledger transaction 저장 실패: %w", err)
	}

	const insertEntryQuery = `
		INSERT INTO ledger_entries (id, transaction_id, account_id, direction, amount, currency)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	for _, entry := range entries {
		if _, err := sqlTx.ExecContext(
			ctx,
			insertEntryQuery,
			entry.ID,
			entry.TransactionID,
			entry.AccountID,
			entry.Direction,
			entry.Amount,
			entry.Currency,
		); err != nil {
			return false, fmt.Errorf("deposit ledger entry 저장 실패: %w", err)
		}
	}

	const markProcessedQuery = `
		UPDATE processed_events
		SET status = $1, processed_at = now()
		WHERE event_key = $2
	`

	if _, err := sqlTx.ExecContext(
		ctx,
		markProcessedQuery,
		ProcessedEventStatusProcessed,
		event.EventKey,
	); err != nil {
		return false, fmt.Errorf("processed event 완료 표시 실패: %w", err)
	}

	if err := sqlTx.Commit(); err != nil {
		return false, fmt.Errorf("deposit 저장 commit 실패: %w", err)
	}
	committed = true

	return true, nil
}
```

</details>

### 왜 Repository가 Ledger 테이블까지 직접 저장하는가

Day25의 저장은 반드시 한 transaction이어야 합니다.

```text
processed_events만 저장되고 deposits가 실패하면?
-> event는 처리된 것처럼 보이는데 입금은 없음

deposits만 저장되고 ledger_entries가 실패하면?
-> 입금은 CREDITED처럼 보이는데 Ledger 잔액은 반영되지 않음

ledger_entries만 저장되고 processed_events가 실패하면?
-> 재처리 때 중복 ledger entry 위험
```

기존 `ledger.Repository.CreateTransaction`을 그대로 호출하면 DB transaction 경계를 공유하기 어렵습니다. 그래서 Day25에서는 Deposit Repository가 같은 DB transaction 안에서 필요한 테이블을 함께 저장합니다.

나중에 코드가 커지면 아래처럼 리팩터링할 수 있습니다.

```text
LedgerRepository.CreateTransactionTx(sqlTx, tx, entries)
DepositRepository.SaveDepositCreditTx(sqlTx, ...)
UnitOfWork 또는 TransactionManager
```

하지만 Day25에서는 먼저 안전한 흐름을 눈에 보이게 만드는 것이 우선입니다.

## Step 7. Repository 통합 테스트를 작성한다

파일 위치:

```text
internal/deposit/repository_test.go
```

테스트 목표:

```text
1. 새 event는 processed_events, deposits, ledger_transactions, ledger_entries를 만든다.
2. 같은 event를 다시 처리하면 duplicate로 판단하고 row를 늘리지 않는다.
3. ledger entry는 debit/credit 2개만 생긴다.
```

테스트 helper는 기존 `internal/settlement/repository_test.go` 패턴을 참고합니다.

통합 테스트에서 준비할 fixture:

```text
ledger_accounts row 2개

acct_deposit_clearing_usdc
= type DEPOSIT_CLEARING

acct_customer_deposit_1
= type CUSTOMER
```

실행 명령:

```bash
TEST_DATABASE_URL="postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" \
go test ./internal/deposit -run TestRepositorySaveDepositCredit -v
```

## Step 8. 포맷과 테스트를 실행한다

```bash
go fmt ./internal/ledger ./internal/deposit
go test ./internal/deposit -run TestProcessorProcessEvent -v
TEST_DATABASE_URL="postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" \
go test ./internal/deposit -run TestRepositorySaveDepositCredit -v
go test ./...
go vet ./...
```

예상 결과:

```text
TestProcessorProcessEvent 통과
TestRepositorySaveDepositCredit 통과
go test ./... 통과
go vet ./... 통과
```

## 자주 만나는 오류

### `duplicate key value violates unique constraint`

같은 event를 두 번 처리할 때 이 오류가 밖으로 나오면 안 됩니다.

```text
원인:
processed_events INSERT에서 ON CONFLICT DO NOTHING을 사용하지 않았거나,
RowsAffected를 확인하지 않았다.

기대:
중복 event는 error가 아니라 Duplicate result로 반환한다.
```

### `ledger_entries account_id foreign key`

```text
원인:
recipient_account_id 또는 clearing_account_id에 해당하는 ledger_accounts fixture가 없다.

해결:
통합 테스트에서 ledger_accounts를 먼저 insert한다.
```

### `debit과 credit 합계가 일치하지 않는다`

```text
원인:
입금 credit entry만 만들었거나 debit entry 금액과 다르다.

해결:
DEPOSIT_CLEARING debit과 CUSTOMER credit을 같은 금액/통화로 만든다.
```

### `append`와 slice가 헷갈린다

테스트 fake store에서 아래 코드가 나옵니다.

```go
f.entries = append([]ledger.Entry(nil), entries...)
```

이 코드는 테스트가 받은 entries를 복사해 저장합니다. 원본 slice가 나중에 바뀌어도 fake store의 기록이 흔들리지 않게 하기 위한 테스트 패턴입니다.

## 완성 기준

```text
1. ProcessEvent 흐름을 먼저 설명할 수 있다.
2. 같은 blockchain event가 여러 번 들어오는 이유를 설명할 수 있다.
3. event_key가 chain + tx_hash + log_index인 이유를 설명할 수 있다.
4. processed_events와 deposits 테이블의 역할 차이를 설명할 수 있다.
5. 입금 Ledger가 debit/credit 두 entry로 저장되는 이유를 설명할 수 있다.
6. 중복 event가 error가 아니라 Duplicate result인 이유를 설명할 수 있다.
7. PostgreSQL 통합 테스트에서 중복 처리 방지가 확인된다.
```

## 커밋 메시지

```bash
git add migrations/000004_create_deposit_tables.* internal/ledger/ledger.go internal/deposit
git commit -m "feat: Deposit 이벤트 멱등 처리 구현"
```

## 다음 작업 예고

Day26에서는 Event Indexer mock을 만들어 여러 block range에서 transfer event를 읽는 흐름을 붙입니다.

```text
Day25:
이미 들어온 event를 안전하게 처리한다.

Day26:
event를 어디서 어떻게 읽어올지 만든다.
```
