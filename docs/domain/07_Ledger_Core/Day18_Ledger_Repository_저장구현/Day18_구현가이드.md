# Day 18 구현가이드 - Ledger Repository 저장 구현

관련 Jira: [SPN-35](https://aslan0.atlassian.net/browse/SPN-35)

Day18은 Day18 이후 새 리듬을 처음 적용하는 문서입니다.

이 문서 하나 안에 출퇴근 예습 자료와 퇴근 후 실습 가이드를 함께 넣습니다.

오늘의 퇴근 후 실습은 작은 코드 작업 하나입니다.

```text
Ledger Transaction 1건과 Ledger Entry 여러 건을
하나의 DB transaction으로 저장하는 Repository 메서드를 만든다.
```

## 오늘의 큰 그림

![Day18 Ledger Repository 저장 구현](https://raw.githubusercontent.com/HoBaeBang/2030-korea-stablepay-network/main/docs/confluence/diagrams/spn35-day18-ledger-repository-write.png)

## 오늘 만들 것

```text
파일:
internal/ledger/repository.go

새 메서드:
func (r *Repository) CreateTransaction(ctx context.Context, tx Transaction, entries []Entry) error
```

오늘 메서드의 책임은 아래 한 줄입니다.

```text
하나의 원장 거래와 그에 속한 여러 원장 항목을
중간 실패 없이 함께 저장한다.
```

## 왜 이 기능이 필요한가?

Day15에서는 Ledger 규칙을 검증했습니다.

Day16에서는 Ledger를 저장할 테이블을 만들었습니다.

Day17에서는 Repository라는 DB 경계를 만들었습니다.

이제 Day18에서는 그 경계 안에 "실제 저장"을 넣습니다.

Payment 상태만으로는 돈의 이동 기록이 남지 않습니다.

Ledger가 진짜 역할을 하려면 아래 두 가지가 함께 저장되어야 합니다.

```text
1. ledger_transactions
   -> 이 돈의 이동이 어떤 비즈니스 사건에서 발생했는가

2. ledger_entries
   -> 실제로 어느 계정에 얼마가 debit/credit 되었는가
```

둘 중 하나만 저장되고 다른 하나가 실패하면 원장 기록이 깨집니다.

그래서 Day18의 핵심은 단순 INSERT가 아니라 "같이 저장하고 같이 실패하는 흐름"을 만드는 것입니다.

## 출퇴근 예습 포인트

출퇴근 시간에는 아래 질문에 답할 수 있을 정도로 읽습니다.

```text
1. *sql.DB는 어디서 만들어지고 Repository까지 어떻게 전달되는가?
2. sql.Tx는 무엇이고 왜 필요한가?
3. 왜 ledger_transactions와 ledger_entries를 따로 저장하면 위험한가?
4. 오늘 ledger_accounts INSERT까지 같이 만들지 않는 이유는 무엇인가?
5. Day19에서는 무엇을 검증하게 되는가?
```

## 오늘의 핵심 문장

```text
Ledger 저장은 "row를 여러 개 넣는 작업"이 아니라,
"하나의 비즈니스 사건을 원자적으로 기록하는 작업"이다.
```

## 저장 흐름을 먼저 이해하기

예를 들어 아래 결제 사건이 있다고 생각해 봅니다.

```text
payment pay_123 finalized
```

이 사건은 Ledger에서 아래처럼 저장됩니다.

```text
ledger_transactions row 1개
-> 이 기록이 왜 생겼는지 설명

ledger_entries row 여러 개
-> 실제 돈이 어떻게 이동했는지 설명
```

예시:

```text
ledger_transactions
  id              = led_tx_123
  reference_type  = PAYMENT
  reference_id    = pay_123
  idempotency_key = payment:pay_123:finalized

ledger_entries
  1. 고객 계정              DEBIT  10_000_000 USDC
  2. 가맹점 지급 예정 계정  CREDIT  9_800_000 USDC
  3. 플랫폼 수수료 계정      CREDIT    200_000 USDC
```

이 세 Entry는 하나의 Ledger Transaction에 속합니다.

그래서 저장도 아래처럼 움직여야 합니다.

```text
BeginTx
-> ledger_transactions INSERT
-> ledger_entries 여러 건 INSERT
-> Commit
```

중간에 하나라도 실패하면:

```text
Rollback
```

## `sql.Tx`는 무엇인가?

`sql.Tx`는 database/sql이 제공하는 DB transaction 객체입니다.

쉽게 말하면 아래 뜻입니다.

```text
지금부터 하는 여러 SQL 작업을
하나의 묶음으로 성공시키거나,
하나의 묶음으로 취소하겠다.
```

## `*sql.DB`는 어디서 오는가?

Day17에서 궁금해했던 부분을 먼저 짚고 갑니다.

`*sql.DB`는 Repository가 스스로 만드는 것이 아닙니다.

현재 프로젝트에서는 `main.go`와 `database.Open`이 그 역할을 합니다.

### `cmd/api/main.go`

확인할 부분:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

db, err := database.Open(ctx, cfg.DatabaseURL)
if err != nil {
	log.Fatalf("database connection failed: %v", err)
}
defer db.Close()
```

의미:

```text
main.go가 먼저 DB 연결 풀을 만든다.
그 다음 Repository 생성자에 넣어준다.
```

### `internal/platform/database/database.go`

확인할 부분:

```go
func Open(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	...
	return db, nil
}
```

의미:

```text
sql.Open("pgx", dsn)이 *sql.DB를 만든다.
즉 Repository가 받는 db는 여기서 만들어진 연결 풀이다.
```

## 오늘 확인할 기존 파일

오늘 수정하지 않고 확인만 하는 파일입니다.

### `internal/ledger/ledger.go`

확인할 타입:

```go
type Transaction struct {
	ID             string
	ReferenceType  string
	ReferenceID    string
	IdempotencyKey string
	CreatedAt      time.Time
}

type Entry struct {
	ID            string
	TransactionID string
	AccountID     string
	Direction     EntryDirection
	Amount        int64
	Currency      string
	CreatedAt     time.Time
}
```

확인 포인트:

```text
오늘 Repository 메서드는 Transaction 1건과 Entry 여러 건을 저장한다.
```

### `migrations/000002_create_ledger_core_tables.up.sql`

확인할 부분:

```sql
CREATE TABLE ledger_transactions
(
    id              TEXT PRIMARY KEY,
    reference_type  TEXT        NOT NULL,
    reference_id    TEXT        NOT NULL,
    idempotency_key TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ledger_entries
(
    id             TEXT PRIMARY KEY,
    transaction_id TEXT        NOT NULL REFERENCES ledger_transactions (id),
    account_id     TEXT        NOT NULL REFERENCES ledger_accounts (id),
    direction      TEXT        NOT NULL,
    amount         BIGINT      NOT NULL,
    currency       TEXT        NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

확인 포인트:

```text
entry는 transaction을 참조한다.
즉 transaction이 먼저 저장되어야 한다.
```

### `internal/ledger/repository.go`

현재 상태:

```go
type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}
```

확인 포인트:

```text
오늘은 이 파일에 "저장 메서드"를 추가한다.
```

## 오늘 만들 메서드의 범위

오늘은 아래 범위까지만 합니다.

```text
1. Repository 메서드 추가
2. BeginTx / Commit / Rollback 흐름 추가
3. ledger_transactions INSERT
4. ledger_entries loop INSERT
5. 최소한의 입력 검증 추가
```

오늘 하지 않는 것:

```text
1. ledger_accounts 생성 로직
2. API나 main.go 연결
3. 실제 HTTP 요청으로 호출하는 흐름
4. idempotency_key 세부 오류 분기
5. 저장 integration test
```

이 다섯 가지는 Day19 이후에 이어집니다.

## 왜 `ledger_accounts` INSERT는 오늘 안 하는가?

`ledger_accounts`는 계정 마스터 데이터에 가깝습니다.

반면 오늘 저장하는 것은 특정 비즈니스 사건에 대한 원장 거래입니다.

즉 오늘 저장 대상은 아래 두 가지입니다.

```text
ledger_transactions
ledger_entries
```

오늘 전제는 아래와 같습니다.

```text
entry가 참조하는 account_id는 이미 존재한다고 가정한다.
```

## Step 1. `repository.go` import 보강

Day18에서는 아래 import가 필요합니다.

```go
import (
	"context"
	"database/sql"
	"fmt"
)
```

## Step 2. 메서드 시그니처 추가

```go
func (r *Repository) CreateTransaction(ctx context.Context, tx Transaction, entries []Entry) error
```

## Step 3. 저장 전 최소 입력 검증

Repository는 Service처럼 비즈니스 규칙 전체를 검증하지는 않습니다.

하지만 저장 자체가 불가능한 입력은 초반에 막는 편이 읽기 쉽습니다.

예:

```text
ctx가 nil이다
db가 nil이다
entries가 비어 있다
tx.ID가 비어 있다
entry.TransactionID가 tx.ID와 다르다
entry.AccountID가 비어 있다
```

## Step 4. DB transaction 시작

```go
sqlTx, err := r.db.BeginTx(ctx, nil)
if err != nil {
	return fmt.Errorf("원장 저장 transaction 시작 실패: %w", err)
}
```

## Step 5. rollback 보장

```go
committed := false
defer func() {
	if !committed {
		_ = sqlTx.Rollback()
	}
}()
```

의미:

```text
중간에 에러가 나면 commit되지 않았으므로 rollback이 실행된다.
```

## Step 6. `ledger_transactions` 먼저 저장

SQL:

```sql
INSERT INTO ledger_transactions (id, reference_type, reference_id, idempotency_key)
VALUES ($1, $2, $3, $4)
```

왜 먼저 저장하나?

`ledger_entries.transaction_id`가 `ledger_transactions.id`를 참조하는 foreign key이기 때문입니다.

## Step 7. `ledger_entries` 여러 건 저장

SQL:

```sql
INSERT INTO ledger_entries (id, transaction_id, account_id, direction, amount, currency)
VALUES ($1, $2, $3, $4, $5, $6)
```

중요한 점:

```text
Entry는 여러 건일 수 있으므로 loop가 필요하다.
```

## Step 8. commit

```go
if err := sqlTx.Commit(); err != nil {
	return fmt.Errorf("원장 저장 commit 실패: %w", err)
}

committed = true
return nil
```

## `repository.go` 최종 완성본 전체

<details>
<summary><code>repository.go</code> 최종 완성본 전체 보기</summary>

```go
package ledger

import (
	"context"
	"database/sql"
	"fmt"
)

// Repository는 Ledger 데이터를 DB에 저장하고 조회하는 경계이다.
type Repository struct {
	db *sql.DB
}

// NewRepository는 Ledger Repository 인스턴스를 만든다.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateTransaction은 Ledger Transaction 1건과 Entry 여러 건을 하나의 DB transaction으로 저장한다.
func (r *Repository) CreateTransaction(ctx context.Context, tx Transaction, entries []Entry) error {
	if ctx == nil {
		return fmt.Errorf("context가 필요합니다")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if r.db == nil {
		return fmt.Errorf("ledger repository db가 필요합니다")
	}

	if tx.ID == "" {
		return fmt.Errorf("원장 거래 id가 필요합니다")
	}

	if tx.ReferenceType == "" {
		return fmt.Errorf("원장 거래 reference type이 필요합니다")
	}

	if tx.ReferenceID == "" {
		return fmt.Errorf("원장 거래 reference id가 필요합니다")
	}

	if tx.IdempotencyKey == "" {
		return fmt.Errorf("원장 거래 idempotency key가 필요합니다")
	}

	if len(entries) == 0 {
		return fmt.Errorf("원장 항목이 필요합니다")
	}

	for _, entry := range entries {
		if entry.ID == "" {
			return fmt.Errorf("원장 항목 id가 필요합니다")
		}

		if entry.TransactionID != tx.ID {
			return fmt.Errorf("원장 항목의 transaction id가 원장 거래 id와 다릅니다")
		}

		if entry.AccountID == "" {
			return fmt.Errorf("원장 항목 account id가 필요합니다")
		}
	}

	sqlTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("원장 저장 transaction 시작 실패: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = sqlTx.Rollback()
		}
	}()

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
		return fmt.Errorf("원장 거래 저장 실패: %w", err)
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
			return fmt.Errorf("원장 항목 저장 실패: %w", err)
		}
	}

	if err := sqlTx.Commit(); err != nil {
		return fmt.Errorf("원장 저장 commit 실패: %w", err)
	}

	committed = true
	return nil
}
```

</details>

## 실습 순서

```text
1. repository.go 현재 코드 확인
2. import 보강
3. CreateTransaction 메서드 추가
4. gofmt 실행
5. go test ./... 실행
6. 실습산출물 작성
```

## 실행 명령

```bash
gofmt -w internal/ledger/repository.go
go test ./...
```

## 검증 방법

오늘은 아직 HTTP/API 연결과 DB integration test를 붙이지 않습니다.

그래서 검증 기준은 아래입니다.

```text
1. repository.go가 컴파일된다.
2. go test ./...가 통과한다.
3. BeginTx -> transaction insert -> entry insert -> Commit 흐름이 보인다.
4. 실패 시 rollback되는 구조가 있다.
```

## 자주 만나는 오류

### 1. `undefined: context`

원인:

```text
context import를 추가하지 않았다.
```

### 2. `undefined: fmt`

원인:

```text
fmt.Errorf를 쓰는데 fmt import가 없다.
```

### 3. `foreign key violation`

의미:

```text
entry.account_id가 ledger_accounts에 없는 계정을 가리킬 수 있다.
또는 transaction insert 전에 entry insert를 시도했을 수 있다.
```

## 오늘 실습이 끝나면 설명할 수 있어야 하는 것

```text
1. 왜 sql.Tx가 필요한가?
2. 왜 ledger_transactions와 ledger_entries를 한 번에 저장해야 하는가?
3. 왜 entry는 transaction보다 나중에 저장해야 하는가?
4. *sql.DB는 어디서 만들어져 Repository로 들어오는가?
5. 왜 오늘은 ledger_accounts 생성까지 하지 않는가?
```

## 실습산출물 작성 포인트

오늘 산출물에서는 아래를 자기 말로 정리하면 됩니다.

```text
Repository 저장 메서드의 책임
sql.Tx가 필요한 이유
commit / rollback 흐름
transaction row와 entry row의 관계
다음 날 더 검증해야 할 부분
```

## 커밋 메시지

구현까지 끝났다면 아래 형태로 작성합니다.

```text
feat: ledger repository 저장 메서드 구현
```

문서만 먼저 반영하면 아래 형태를 사용합니다.

```text
docs: Day18 ledger repository 저장 구현 자료 추가
```

## 다음 작업 예고

Day19에서는 아래를 이어서 봅니다.

```text
1. 저장 성공/실패를 실제로 더 검증하는 방법
2. idempotency_key 중복이 발생할 때 어떤 에러가 나오는가
3. 저장 테스트를 어떤 형태로 붙이면 좋은가
4. ledger_accounts fixture 또는 seed를 어떻게 준비할 것인가
```
