# Day 19 구현가이드 - Repository 저장 검증과 Idempotency

관련 Jira: SPN-36

Day19는 Day18에서 만든 `CreateTransaction` 저장 메서드를 실제 DB 기준으로 검증하는 날입니다.

Day18에서 한 일은 아래였습니다.

```text
Ledger Transaction 1건과 Ledger Entry 여러 건을
하나의 DB transaction으로 저장하는 Repository 메서드를 만든다.
```

Day19에서 할 일은 아래입니다.

```text
그 저장 메서드가 성공, 중복, 실패 상황에서
원장을 깨뜨리지 않는지 테스트로 확인한다.
```

## 오늘의 큰 그림

![Day19 Repository 저장 검증](https://raw.githubusercontent.com/HoBaeBang/2030-korea-stablepay-network/main/docs/confluence/diagrams/spn36-day19-repository-idempotency.png)

## 오늘 만들 것

```text
파일:
internal/ledger/repository_test.go

새 테스트:
1. transaction과 entries를 함께 저장한다
2. 같은 idempotency_key는 중복 저장되지 않는다
3. 존재하지 않는 account_id를 사용하면 저장이 실패하고 transaction도 남지 않는다
```

오늘의 핵심은 `Repository`를 믿을 수 있게 만드는 것입니다.

## 왜 이 기능이 필요한가?

Ledger는 돈의 이동 기록입니다.

그래서 단순히 코드가 컴파일되는 것만으로는 부족합니다.

아래 질문에 답할 수 있어야 합니다.

```text
1. 정상 입력이면 transaction row와 entry rows가 함께 저장되는가?
2. 같은 비즈니스 사건이 두 번 들어오면 중복 저장을 막는가?
3. entry 저장 중 실패하면 앞에서 넣은 transaction row도 rollback되는가?
```

이 세 가지가 확인되지 않으면 Ledger는 운영에서 위험합니다.

예를 들어 결제 확정 이벤트가 두 번 들어왔는데 Ledger가 두 번 credit을 기록하면, 가맹점에게 지급 가능한 금액이 실제보다 커질 수 있습니다.

반대로 transaction row만 저장되고 entries가 실패하면, "왜 발생했는지"는 있는데 "돈이 어떻게 움직였는지"가 없는 깨진 장부가 됩니다.

## 출퇴근 예습 포인트

출퇴근 시간에는 아래 질문에 답할 수 있을 정도로 읽습니다.

```text
1. idempotency는 왜 필요한가?
2. unique index는 중복을 어떻게 막는가?
3. foreign key 오류는 어떤 상황에서 발생하는가?
4. rollback 검증은 왜 transaction row count까지 확인해야 하는가?
5. integration test와 unit test는 무엇이 다른가?
```

## 오늘의 핵심 문장

```text
Repository 저장 검증은 "저장된다"만 보는 것이 아니라,
"중복 저장되지 않고, 실패할 때 깨끗하게 되돌아가는지"까지 확인하는 일이다.
```

## 핵심 용어

| 용어 | 한글 의미 | 오늘 문맥에서의 의미 |
| --- | --- | --- |
| Idempotency | 멱등성, 같은 요청을 여러 번 처리해도 결과가 한 번 처리된 것과 같은 성질 | 같은 결제 확정 이벤트가 두 번 와도 Ledger가 두 번 기록되지 않게 하는 기준 |
| idempotency_key | 멱등성 키 | `payment:pay_123:finalized`처럼 같은 비즈니스 사건을 식별하는 문자열 |
| Unique index | 고유 인덱스 | 같은 값이 두 번 들어가지 못하게 DB가 막는 제약 |
| Foreign key | 외래 키 | 한 테이블의 값이 다른 테이블의 row를 참조한다는 제약 |
| Rollback | 되돌리기 | DB transaction 중간에 실패하면 앞에서 했던 INSERT까지 취소하는 동작 |
| Integration test | 통합 테스트 | 실제 DB 같은 외부 구성요소까지 연결해서 확인하는 테스트 |

## `idempotency_key`를 왜 보는가?

`ledger_transactions`에는 아래 unique index가 있습니다.

```sql
CREATE UNIQUE INDEX idx_ledger_transactions_idempotency_key
    ON ledger_transactions (idempotency_key);
```

이 뜻은 아래와 같습니다.

```text
같은 idempotency_key를 가진 ledger_transaction은 두 번 저장할 수 없다.
```

예를 들어 아래 사건은 한 번만 기록되어야 합니다.

```text
payment pay_123 finalized
```

그래서 아래 키를 사용합니다.

```text
payment:pay_123:finalized
```

동일한 이벤트가 재시도되거나 중복 수신되어도 이 키가 같으면 DB가 중복 저장을 막습니다.

## 왜 `foreign key` 오류를 일부러 테스트하는가?

`ledger_entries.account_id`는 `ledger_accounts.id`를 참조합니다.

```sql
account_id TEXT NOT NULL REFERENCES ledger_accounts (id)
```

그래서 존재하지 않는 account를 참조하면 entry INSERT가 실패해야 합니다.

중요한 점은 여기서 끝이 아닙니다.

Day18의 `CreateTransaction`은 아래 순서로 동작합니다.

```text
1. BeginTx
2. ledger_transactions INSERT 성공
3. ledger_entries INSERT 실패
4. Rollback
```

그렇다면 테스트는 아래까지 확인해야 합니다.

```text
entry 저장은 실패했다.
그리고 앞에서 저장했던 transaction row도 남아 있지 않다.
```

이게 rollback 검증입니다.

## integration test를 왜 바로 일반 테스트처럼 실행하지 않는가?

Repository 저장 검증은 실제 PostgreSQL이 필요합니다.

하지만 모든 개발자가 항상 로컬 DB를 켜둔 상태는 아닙니다.

그래서 Day19 테스트는 아래 방식으로 작성합니다.

```text
TEST_DATABASE_URL이 있으면 실제 DB 테스트 실행
TEST_DATABASE_URL이 없으면 테스트 skip
```

이렇게 하면 평소에는 `go test ./...`가 깨지지 않고, DB 검증이 필요할 때만 아래처럼 실행할 수 있습니다.

```bash
TEST_DATABASE_URL="postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" go test ./internal/ledger -run Repository
```

## 사전 조건

Day19 실습 전에 Day18의 `CreateTransaction` 메서드가 먼저 구현되어 있어야 합니다.

확인 명령:

```bash
sed -n '1,260p' internal/ledger/repository.go
```

아래 메서드가 있어야 합니다.

```go
func (r *Repository) CreateTransaction(ctx context.Context, tx Transaction, entries []Entry) error
```

아직 없다면 Day18 구현가이드를 먼저 완료합니다.

## DB 준비

PostgreSQL을 실행합니다.

```bash
docker compose up -d postgres
```

migration이 아직 적용되지 않았다면 아래 SQL을 적용합니다.

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -f migrations/000002_create_ledger_core_tables.up.sql
```

이미 적용되어 있다면 중복 생성 오류가 날 수 있습니다. 그 경우에는 기존 DB에 테이블이 이미 있다는 뜻일 수 있으니, 지금 단계에서는 무리해서 다시 적용하지 않습니다.

## 오늘 수정/생성할 파일

오늘 새로 만드는 파일은 하나입니다.

```text
internal/ledger/repository_test.go
```

이 파일의 역할은 아래입니다.

```text
Repository가 실제 DB에 Ledger Transaction과 Entry를 안전하게 저장하는지 확인한다.
```

## Step 1. 테스트용 DB 연결 만들기

테스트에서 DB URL은 환경변수로 받습니다.

```go
dsn := os.Getenv("TEST_DATABASE_URL")
if dsn == "" {
    t.Skip("TEST_DATABASE_URL이 없어서 ledger repository integration test를 건너뜁니다")
}
```

`Skip`은 실패가 아닙니다.

```text
지금은 DB 환경이 준비되지 않았으니 이 테스트는 실행하지 않겠다는 뜻입니다.
```

## Step 2. 테스트용 account seed 만들기

`ledger_entries.account_id`는 반드시 `ledger_accounts.id`를 참조해야 합니다.

그래서 테스트 전에 account row를 넣어둡니다.

```sql
INSERT INTO ledger_accounts (id, type, owner_id, currency)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO NOTHING
```

`ON CONFLICT (id) DO NOTHING`은 같은 account id가 이미 있으면 그냥 넘어가라는 뜻입니다.

## Step 3. 정상 저장 테스트

검증할 것:

```text
1. CreateTransaction이 nil error를 반환한다.
2. ledger_transactions에 tx row가 1개 생긴다.
3. ledger_entries에 entry row가 3개 생긴다.
```

## Step 4. idempotency 중복 테스트

검증할 것:

```text
1. 첫 번째 저장은 성공한다.
2. 같은 idempotency_key로 두 번째 저장하면 실패한다.
3. 중복 저장 때문에 같은 비즈니스 사건이 두 번 기록되지 않는다.
```

여기서 중요한 점은 `tx.ID`가 달라도 `idempotency_key`가 같으면 실패해야 한다는 점입니다.

왜냐하면 중복 판단 기준은 DB primary key가 아니라 비즈니스 사건 키이기 때문입니다.

## Step 5. foreign key 실패와 rollback 테스트

검증할 것:

```text
1. 존재하지 않는 account_id를 사용하면 저장이 실패한다.
2. 실패 후 ledger_transactions에 tx row가 남아 있지 않다.
```

이 테스트가 통과하면 Day18에서 만든 rollback 구조가 실제로 의미 있음을 확인할 수 있습니다.

## `repository_test.go` 최종 완성본 전체

<details>
<summary><code>repository_test.go</code> 최종 완성본 전체 보기</summary>

```go
package ledger

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func newTestRepository(t *testing.T) (*Repository, *sql.DB, context.Context) {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL이 없어서 ledger repository integration test를 건너뜁니다")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("테스트 DB 연결 생성 실패: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("테스트 DB ping 실패: %v", err)
	}

	return NewRepository(db), db, ctx
}

func seedLedgerAccount(t *testing.T, ctx context.Context, db *sql.DB, account Account) {
	t.Helper()

	const query = `
INSERT INTO ledger_accounts (id, type, owner_id, currency)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO NOTHING
`

	if _, err := db.ExecContext(ctx, query, account.ID, account.Type, account.OwnerID, account.Currency); err != nil {
		t.Fatalf("테스트 원장 계정 생성 실패: %v", err)
	}
}

func cleanupLedgerTransaction(t *testing.T, ctx context.Context, db *sql.DB, transactionID string) {
	t.Helper()

	_, _ = db.ExecContext(ctx, "DELETE FROM ledger_entries WHERE transaction_id = $1", transactionID)
	_, _ = db.ExecContext(ctx, "DELETE FROM ledger_transactions WHERE id = $1", transactionID)
}

func countRows(t *testing.T, ctx context.Context, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		t.Fatalf("row count 조회 실패: %v", err)
	}

	return count
}

func TestRepositoryCreateTransaction(t *testing.T) {
	t.Run("transaction과 entries를 함께 저장한다", func(t *testing.T) {
		repo, db, ctx := newTestRepository(t)

		seedLedgerAccount(t, ctx, db, Account{ID: "acct_test_customer_1", Type: AccountTypeCustomer, OwnerID: "customer_test_1", Currency: "USDC"})
		seedLedgerAccount(t, ctx, db, Account{ID: "acct_test_merchant_pending_1", Type: AccountTypeMerchantPending, OwnerID: "merchant_test_1", Currency: "USDC"})
		seedLedgerAccount(t, ctx, db, Account{ID: "acct_test_platform_fee_1", Type: AccountTypePlatformFee, OwnerID: "platform", Currency: "USDC"})

		tx := Transaction{
			ID:             "led_tx_test_success_1",
			ReferenceType:  "PAYMENT",
			ReferenceID:    "pay_test_success_1",
			IdempotencyKey: "payment:pay_test_success_1:finalized",
		}
		entries := []Entry{
			{ID: "led_entry_test_success_1", TransactionID: tx.ID, AccountID: "acct_test_customer_1", Direction: EntryDirectionDebit, Amount: 10_000_000, Currency: "USDC"},
			{ID: "led_entry_test_success_2", TransactionID: tx.ID, AccountID: "acct_test_merchant_pending_1", Direction: EntryDirectionCredit, Amount: 9_800_000, Currency: "USDC"},
			{ID: "led_entry_test_success_3", TransactionID: tx.ID, AccountID: "acct_test_platform_fee_1", Direction: EntryDirectionCredit, Amount: 200_000, Currency: "USDC"},
		}
		t.Cleanup(func() { cleanupLedgerTransaction(t, ctx, db, tx.ID) })

		if err := repo.CreateTransaction(ctx, tx, entries); err != nil {
			t.Fatalf("원장 거래 저장이 성공해야 하는데 실패했습니다: %v", err)
		}

		transactionCount := countRows(t, ctx, db, "SELECT count(*) FROM ledger_transactions WHERE id = $1", tx.ID)
		if transactionCount != 1 {
			t.Fatalf("ledger_transactions row가 1개여야 하는데 %d개입니다", transactionCount)
		}

		entryCount := countRows(t, ctx, db, "SELECT count(*) FROM ledger_entries WHERE transaction_id = $1", tx.ID)
		if entryCount != len(entries) {
			t.Fatalf("ledger_entries row가 %d개여야 하는데 %d개입니다", len(entries), entryCount)
		}
	})

	t.Run("같은 idempotency_key는 중복 저장되지 않는다", func(t *testing.T) {
		repo, db, ctx := newTestRepository(t)

		seedLedgerAccount(t, ctx, db, Account{ID: "acct_test_customer_2", Type: AccountTypeCustomer, OwnerID: "customer_test_2", Currency: "USDC"})
		seedLedgerAccount(t, ctx, db, Account{ID: "acct_test_merchant_pending_2", Type: AccountTypeMerchantPending, OwnerID: "merchant_test_2", Currency: "USDC"})

		idempotencyKey := "payment:pay_test_duplicate:finalized"
		firstTx := Transaction{ID: "led_tx_test_duplicate_1", ReferenceType: "PAYMENT", ReferenceID: "pay_test_duplicate", IdempotencyKey: idempotencyKey}
		firstEntries := []Entry{
			{ID: "led_entry_test_duplicate_1", TransactionID: firstTx.ID, AccountID: "acct_test_customer_2", Direction: EntryDirectionDebit, Amount: 10_000_000, Currency: "USDC"},
			{ID: "led_entry_test_duplicate_2", TransactionID: firstTx.ID, AccountID: "acct_test_merchant_pending_2", Direction: EntryDirectionCredit, Amount: 10_000_000, Currency: "USDC"},
		}
		t.Cleanup(func() { cleanupLedgerTransaction(t, ctx, db, firstTx.ID) })

		if err := repo.CreateTransaction(ctx, firstTx, firstEntries); err != nil {
			t.Fatalf("첫 번째 저장은 성공해야 합니다: %v", err)
		}

		secondTx := Transaction{ID: "led_tx_test_duplicate_2", ReferenceType: "PAYMENT", ReferenceID: "pay_test_duplicate", IdempotencyKey: idempotencyKey}
		secondEntries := []Entry{
			{ID: "led_entry_test_duplicate_3", TransactionID: secondTx.ID, AccountID: "acct_test_customer_2", Direction: EntryDirectionDebit, Amount: 10_000_000, Currency: "USDC"},
			{ID: "led_entry_test_duplicate_4", TransactionID: secondTx.ID, AccountID: "acct_test_merchant_pending_2", Direction: EntryDirectionCredit, Amount: 10_000_000, Currency: "USDC"},
		}
		t.Cleanup(func() { cleanupLedgerTransaction(t, ctx, db, secondTx.ID) })

		if err := repo.CreateTransaction(ctx, secondTx, secondEntries); err == nil {
			t.Fatal("같은 idempotency_key는 중복 저장에 실패해야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("entry 저장 실패 시 transaction도 rollback된다", func(t *testing.T) {
		repo, db, ctx := newTestRepository(t)

		tx := Transaction{
			ID:             "led_tx_test_rollback_1",
			ReferenceType:  "PAYMENT",
			ReferenceID:    "pay_test_rollback_1",
			IdempotencyKey: "payment:pay_test_rollback_1:finalized",
		}
		entries := []Entry{
			{ID: "led_entry_test_rollback_1", TransactionID: tx.ID, AccountID: "acct_missing_for_rollback", Direction: EntryDirectionDebit, Amount: 10_000_000, Currency: "USDC"},
			{ID: "led_entry_test_rollback_2", TransactionID: tx.ID, AccountID: "acct_missing_for_rollback_2", Direction: EntryDirectionCredit, Amount: 10_000_000, Currency: "USDC"},
		}
		t.Cleanup(func() { cleanupLedgerTransaction(t, ctx, db, tx.ID) })

		if err := repo.CreateTransaction(ctx, tx, entries); err == nil {
			t.Fatal("존재하지 않는 account_id를 사용하면 저장이 실패해야 하는데 nil이 반환되었습니다")
		}

		transactionCount := countRows(t, ctx, db, "SELECT count(*) FROM ledger_transactions WHERE id = $1", tx.ID)
		if transactionCount != 0 {
			t.Fatalf("rollback 후 transaction row가 남아 있으면 안 되는데 %d개가 남았습니다", transactionCount)
		}
	})
}

func Example_idempotencyKey() {
	paymentID := "pay_123"
	key := fmt.Sprintf("payment:%s:finalized", paymentID)

	fmt.Println(key)

	// Output:
	// payment:pay_123:finalized
}
```

</details>

## 실행 명령

먼저 일반 테스트를 실행합니다.

```bash
go test ./...
```

DB integration test까지 실행하려면 아래처럼 실행합니다.

```bash
TEST_DATABASE_URL="postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" go test ./internal/ledger -run Repository
```

## 예상 결과

DB URL 없이 실행하면 repository integration test는 skip될 수 있습니다.

```text
TEST_DATABASE_URL이 없어서 ledger repository integration test를 건너뜁니다
```

DB URL을 넣고 실행하면 아래 세 케이스가 통과해야 합니다.

```text
transaction과 entries를 함께 저장한다
같은 idempotency_key는 중복 저장되지 않는다
entry 저장 실패 시 transaction도 rollback된다
```

## 자주 만나는 오류

### 1. `relation "ledger_transactions" does not exist`

의미:

```text
ledger migration이 DB에 적용되지 않았다.
```

해결:

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -f migrations/000002_create_ledger_core_tables.up.sql
```

### 2. `duplicate key value violates unique constraint`

의미:

```text
같은 idempotency_key 또는 같은 primary key가 이미 DB에 존재한다.
```

Day19에서는 이 오류가 완전히 나쁜 것은 아닙니다.

중복 저장 테스트에서는 이 오류가 발생하는 것이 정상입니다.

### 3. `violates foreign key constraint`

의미:

```text
ledger_entries.account_id가 ledger_accounts.id에 없는 값을 참조했다.
```

rollback 테스트에서는 이 오류가 발생하는 것이 정상입니다.

### 4. `sql: unknown driver "pgx"`

의미:

```text
pgx stdlib driver가 import되지 않았다.
```

해결:

```go
_ "github.com/jackc/pgx/v5/stdlib"
```

## 검증 방법

오늘의 검증 기준은 아래입니다.

```text
1. repository_test.go가 컴파일된다.
2. go test ./...가 통과하거나 DB 테스트가 skip된다.
3. TEST_DATABASE_URL을 넣으면 실제 DB 기준 repository 테스트가 실행된다.
4. 성공 저장, idempotency 중복, rollback 실패 케이스를 설명할 수 있다.
```

## 실습산출물 작성 포인트

오늘 산출물에서는 아래를 자기 말로 정리합니다.

```text
Repository 저장 검증의 목적
idempotency_key가 막는 중복
foreign key 오류가 필요한 이유
rollback을 어떻게 확인했는지
unit test와 integration test의 차이
```

## 커밋 메시지

구현까지 끝났다면 아래 형태로 작성합니다.

```text
feat: ledger repository 저장 검증 테스트 추가
```

문서만 먼저 반영하면 아래 형태를 사용합니다.

```text
docs: Day19 ledger repository 저장 검증 자료 추가
```

## 다음 작업 예고

Day20에서는 Ledger Service와 Repository를 연결합니다.

지금까지는 아래가 분리되어 있었습니다.

```text
Service: entry 균형을 검증한다.
Repository: 검증된 transaction과 entries를 저장한다.
```

Day20에서는 이 둘을 이어서 아래 흐름을 만듭니다.

```text
ValidateTransaction
-> CreateTransaction
```
