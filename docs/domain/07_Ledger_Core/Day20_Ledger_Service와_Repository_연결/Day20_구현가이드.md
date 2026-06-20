# Day 20 구현가이드 - Ledger Service와 Repository 연결

관련 Jira: [SPN-37](https://aslan0.atlassian.net/browse/SPN-37)

Day20은 Day15에서 만든 Ledger 검증 흐름과 Day18에서 만든 Repository 저장 흐름을 하나의 Service 유스케이스로 연결하는 날입니다.

지금까지 흐름은 나뉘어 있었습니다.

```text
Day15 Service
-> 원장 entry의 debit / credit 균형을 검증한다.

Day18 Repository
-> 검증된 transaction과 entries를 DB transaction으로 저장한다.
```

Day20에서는 이 둘을 아래처럼 연결합니다.

```text
ValidateTransaction
-> CreateTransaction
```

## 오늘의 큰 그림

![Day20 Ledger Service Repository 연결](https://raw.githubusercontent.com/HoBaeBang/2030-korea-stablepay-network/main/docs/confluence/diagrams/spn37-day20-service-repository-link.png)

## 오늘 만들 것

```text
수정 파일:
internal/ledger/service.go
internal/ledger/service_test.go

새 개념:
Store interface
RecordTransaction 메서드
fakeStore 테스트 대역
```

오늘의 핵심은 Service가 Repository를 직접 세부 구현으로 아는 것이 아니라, `Store`라는 작은 interface를 통해 저장 기능만 알고 있게 만드는 것입니다.

## 먼저 작성할 전체 코드

아래 두 파일의 전체 완성본을 먼저 작성합니다. 코드를 작성한 뒤 하단의 해설을 읽으면서 각 타입과 메서드가 필요한 이유를 확인합니다.

## `service.go` 최종 완성본 전체

<details>
<summary><code>service.go</code> 최종 완성본 전체 보기</summary>

```go
package ledger

import (
	"context"
	"fmt"
)

// Store는 Service가 필요로 하는 Ledger 저장 동작을 정의한다.
type Store interface {
	CreateTransaction(ctx context.Context, tx Transaction, entries []Entry) error
}

// Service는 Ledger 도메인 규칙을 검증하고 실행한다.
type Service struct {
	store Store
}

// NewService는 Ledger Service 인스턴스를 만든다.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// ValidateTransaction은 원장 거래의 기본 규칙을 검증한다.
func (s *Service) ValidateTransaction(ctx context.Context, entries []Entry) error {
	if ctx == nil {
		return fmt.Errorf("context가 필요합니다")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if len(entries) < 2 {
		return fmt.Errorf("원장 거래는 최소 2개 이상의 항목이 필요합니다")
	}

	totals := make(map[string]int64)

	for _, entry := range entries {
		if entry.Amount <= 0 {
			return fmt.Errorf("원장 항목 금액은 0보다 커야 합니다")
		}

		if entry.Currency == "" {
			return fmt.Errorf("원장 항목 통화가 필요합니다")
		}

		switch entry.Direction {
		case EntryDirectionDebit:
			totals[entry.Currency] += entry.Amount
		case EntryDirectionCredit:
			totals[entry.Currency] -= entry.Amount
		default:
			return fmt.Errorf("알 수 없는 원장 항목 방향입니다: %s", entry.Direction)
		}
	}

	for currency, total := range totals {
		if total != 0 {
			return fmt.Errorf("원장 거래의 debit과 credit 합계가 일치하지 않습니다: %s", currency)
		}
	}
	return nil
}

// RecordTransaction은 원장 거래를 검증한 뒤 저장소에 기록한다.
func (s *Service) RecordTransaction(ctx context.Context, tx Transaction, entries []Entry) error {
	if err := s.ValidateTransaction(ctx, entries); err != nil {
		return err
	}

	if s.store == nil {
		return fmt.Errorf("ledger store가 필요합니다")
	}

	return s.store.CreateTransaction(ctx, tx, entries)
}
```

</details>

## `service_test.go` 최종 완성본 전체

<details>
<summary><code>service_test.go</code> 최종 완성본 전체 보기</summary>

```go
package ledger

import (
	"context"
	"errors"
	"testing"
)

func newTestService(t *testing.T) (*Service, context.Context) {
	t.Helper()

	return NewService(nil), context.Background()
}

func newTestServiceWithStore(t *testing.T, store Store) (*Service, context.Context) {
	t.Helper()

	return NewService(store), context.Background()
}

type fakeStore struct {
	calls   int
	tx      Transaction
	entries []Entry
	err     error
}

func (f *fakeStore) CreateTransaction(ctx context.Context, tx Transaction, entries []Entry) error {
	f.calls++
	f.tx = tx
	f.entries = append([]Entry(nil), entries...)

	return f.err
}

func TestServiceValidateTransaction(t *testing.T) {
	t.Run("debit과 credit 합계가 같으면 성공한다", func(t *testing.T) {
		svc, ctx := newTestService(t)

		entries := []Entry{
			{
				AccountID: "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
			{
				AccountID: "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    9_800_000,
				Currency:  "USDC",
			},
			{
				AccountID: "acct_platform_fee_1",
				Direction: EntryDirectionCredit,
				Amount:    200_000,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(ctx, entries); err != nil {
			t.Fatalf("원장 거래의 균형이 맞아야 하는데 에러가 발생했습니다: %v", err)
		}
	})

	t.Run("credit 합계가 부족하면 실패한다", func(t *testing.T) {
		svc, ctx := newTestService(t)

		entries := []Entry{
			{
				AccountID: "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
			{
				AccountID: "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    9_000_000,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(ctx, entries); err == nil {
			t.Fatal("원장 거래의 균형이 맞지 않아야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("entry가 하나뿐이면 실패한다", func(t *testing.T) {
		svc, ctx := newTestService(t)

		entries := []Entry{
			{
				AccountID: "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(ctx, entries); err == nil {
			t.Fatal("에러가 발생해야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("통화가 비어 있으면 실패한다", func(t *testing.T) {
		svc, ctx := newTestService(t)

		entries := []Entry{
			{
				AccountID: "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    10_000_000,
				Currency:  "",
			},
			{
				AccountID: "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    10_000_000,
				Currency:  "",
			},
		}

		if err := svc.ValidateTransaction(ctx, entries); err == nil {
			t.Fatal("통화가 비어 있으면 실패해야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("금액이 0이면 실패한다", func(t *testing.T) {
		svc, ctx := newTestService(t)

		entries := []Entry{
			{
				AccountID: "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    0,
				Currency:  "USDC",
			},
			{
				AccountID: "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    0,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(ctx, entries); err == nil {
			t.Fatal("금액이 0인 원장 항목은 실패해야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("알 수 없는 방향이면 실패한다", func(t *testing.T) {
		svc, ctx := newTestService(t)

		entries := []Entry{
			{
				AccountID: "acct_customer_1",
				Direction: EntryDirection("UNKNOWN"),
				Amount:    10_000_000,
				Currency:  "USDC",
			},
			{
				AccountID: "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(ctx, entries); err == nil {
			t.Fatal("알 수 없는 방향은 실패해야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("context가 취소되었으면 실패한다", func(t *testing.T) {
		svc, _ := newTestService(t)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		entries := []Entry{
			{
				AccountID: "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
			{
				AccountID: "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(ctx, entries); err == nil {
			t.Fatal("context가 취소되었으면 실패해야 하는데 nil이 반환되었습니다")
		}
	})
}

func TestServiceRecordTransaction(t *testing.T) {
	t.Run("검증에 성공하면 저장소를 호출한다", func(t *testing.T) {
		store := &fakeStore{}
		svc, ctx := newTestServiceWithStore(t, store)

		tx := Transaction{
			ID:             "led_tx_service_1",
			ReferenceType:  "PAYMENT",
			ReferenceID:    "pay_service_1",
			IdempotencyKey: "payment:pay_service_1:finalized",
		}
		entries := []Entry{
			{
				ID:            "led_entry_service_1",
				TransactionID: tx.ID,
				AccountID:     "acct_customer_1",
				Direction:     EntryDirectionDebit,
				Amount:        10_000_000,
				Currency:      "USDC",
			},
			{
				ID:            "led_entry_service_2",
				TransactionID: tx.ID,
				AccountID:     "acct_merchant_pending_1",
				Direction:     EntryDirectionCredit,
				Amount:        10_000_000,
				Currency:      "USDC",
			},
		}

		if err := svc.RecordTransaction(ctx, tx, entries); err != nil {
			t.Fatalf("검증에 성공한 원장 거래는 저장되어야 합니다: %v", err)
		}

		if store.calls != 1 {
			t.Fatalf("저장소는 1번 호출되어야 하는데 %d번 호출되었습니다", store.calls)
		}

		if store.tx.ID != tx.ID {
			t.Fatalf("저장소에 전달된 transaction id가 다릅니다: %s", store.tx.ID)
		}

		if len(store.entries) != len(entries) {
			t.Fatalf("저장소에 전달된 entries 개수가 다릅니다: %d", len(store.entries))
		}
	})

	t.Run("검증에 실패하면 저장소를 호출하지 않는다", func(t *testing.T) {
		store := &fakeStore{}
		svc, ctx := newTestServiceWithStore(t, store)

		tx := Transaction{
			ID:             "led_tx_service_invalid",
			ReferenceType:  "PAYMENT",
			ReferenceID:    "pay_service_invalid",
			IdempotencyKey: "payment:pay_service_invalid:finalized",
		}
		entries := []Entry{
			{
				ID:            "led_entry_service_invalid_1",
				TransactionID: tx.ID,
				AccountID:     "acct_customer_1",
				Direction:     EntryDirectionDebit,
				Amount:        10_000_000,
				Currency:      "USDC",
			},
			{
				ID:            "led_entry_service_invalid_2",
				TransactionID: tx.ID,
				AccountID:     "acct_merchant_pending_1",
				Direction:     EntryDirectionCredit,
				Amount:        9_000_000,
				Currency:      "USDC",
			},
		}

		if err := svc.RecordTransaction(ctx, tx, entries); err == nil {
			t.Fatal("균형이 맞지 않는 원장 거래는 실패해야 하는데 nil이 반환되었습니다")
		}

		if store.calls != 0 {
			t.Fatalf("검증 실패 시 저장소는 호출되면 안 되는데 %d번 호출되었습니다", store.calls)
		}
	})

	t.Run("저장소 에러를 반환한다", func(t *testing.T) {
		storeErr := errors.New("저장소 실패")
		store := &fakeStore{err: storeErr}
		svc, ctx := newTestServiceWithStore(t, store)

		tx := Transaction{
			ID:             "led_tx_service_store_error",
			ReferenceType:  "PAYMENT",
			ReferenceID:    "pay_service_store_error",
			IdempotencyKey: "payment:pay_service_store_error:finalized",
		}
		entries := []Entry{
			{
				ID:            "led_entry_service_store_error_1",
				TransactionID: tx.ID,
				AccountID:     "acct_customer_1",
				Direction:     EntryDirectionDebit,
				Amount:        10_000_000,
				Currency:      "USDC",
			},
			{
				ID:            "led_entry_service_store_error_2",
				TransactionID: tx.ID,
				AccountID:     "acct_merchant_pending_1",
				Direction:     EntryDirectionCredit,
				Amount:        10_000_000,
				Currency:      "USDC",
			},
		}

		err := svc.RecordTransaction(ctx, tx, entries)
		if !errors.Is(err, storeErr) {
			t.Fatalf("저장소 에러를 그대로 반환해야 합니다: %v", err)
		}

		if store.calls != 1 {
			t.Fatalf("저장소는 1번 호출되어야 하는데 %d번 호출되었습니다", store.calls)
		}
	})
}
```

</details>

## 코드 해설과 개념 이해

## 왜 이 기능이 필요한가?

Ledger 저장 흐름은 아무 값이나 DB에 넣으면 안 됩니다.

먼저 Service가 도메인 규칙을 검증해야 합니다.

```text
1. entry가 2개 이상인가?
2. amount가 0보다 큰가?
3. currency가 비어 있지 않은가?
4. direction이 DEBIT 또는 CREDIT인가?
5. debit 합계와 credit 합계가 같은가?
```

이 검증을 통과한 뒤에만 Repository 저장으로 넘어가야 합니다.

즉 Day20의 목표는 아래 한 문장입니다.

```text
잘못된 원장 거래는 저장하지 않고,
올바른 원장 거래만 Repository로 넘긴다.
```

## 출퇴근 예습 포인트

출퇴근 시간에는 아래 질문에 답할 수 있을 정도로 읽습니다.

```text
1. Service와 Repository의 책임 차이는 무엇인가?
2. Service가 Repository 구조체가 아니라 Store interface에 의존하면 어떤 장점이 있는가?
3. fakeStore는 왜 필요한가?
4. RecordTransaction은 왜 ValidateTransaction을 먼저 호출해야 하는가?
5. 잘못된 entries가 들어오면 저장 메서드가 호출되면 안 되는 이유는 무엇인가?
```

## 오늘의 핵심 문장

```text
Service는 "검증 후 저장"이라는 업무 흐름을 담당하고,
Repository는 "DB에 저장"이라는 기술 세부사항을 담당한다.
```

## 핵심 용어

| 용어 | 한글 의미 | 오늘 문맥에서의 의미 |
| --- | --- | --- |
| Use case | 사용 사례, 하나의 업무 흐름 | 원장 거래를 검증하고 저장하는 `RecordTransaction` 흐름 |
| Dependency | 의존성 | Service가 저장을 위해 필요로 하는 Store |
| Interface | 동작 약속 | Service가 필요로 하는 저장 동작만 정의한 `Store` |
| Fake | 테스트 대역 | 실제 DB 대신 테스트에서 저장 호출 여부를 기록하는 객체 |
| Boundary | 경계 | Service와 Repository의 책임을 나누는 선 |

## 왜 interface를 쓰는가?

Java식으로 생각하면 아래와 비슷합니다.

```java
interface LedgerStore {
    void createTransaction(Transaction tx, List<Entry> entries);
}

class LedgerService {
    private final LedgerStore store;
}
```

Go에서는 이렇게 표현합니다.

```go
type Store interface {
	CreateTransaction(ctx context.Context, tx Transaction, entries []Entry) error
}
```

이렇게 하면 Service는 실제 구현체가 PostgreSQL Repository인지, 테스트용 fakeStore인지 알 필요가 없습니다.

Service가 아는 것은 딱 하나입니다.

```text
내가 검증한 transaction과 entries를 저장할 수 있는 객체가 필요하다.
```

## 오늘 확인할 기존 파일

### `internal/ledger/repository.go`

확인할 메서드:

```go
func (r *Repository) CreateTransaction(ctx context.Context, tx Transaction, entries []Entry) error
```

확인 포인트:

```text
Repository는 Store interface를 따로 선언하지 않아도 자동으로 만족한다.
Go에서는 어떤 타입이 interface의 메서드를 가지고 있으면 그 interface를 구현한 것으로 본다.
```

### `internal/ledger/service.go`

현재 Service는 store를 가지고 있지 않습니다.

```go
type Service struct{}
```

Day20 이후에는 아래처럼 바뀝니다.

```go
type Service struct {
	store Store
}
```

## 오늘 만들 메서드의 범위

오늘은 아래 범위까지만 합니다.

```text
1. Store interface 추가
2. Service가 Store를 필드로 가지도록 변경
3. NewService(store Store) 생성자 변경
4. RecordTransaction 메서드 추가
5. fakeStore를 이용한 Service 테스트 추가
```

오늘 하지 않는 것:

```text
1. HTTP API 연결
2. Payment FINALIZED와 Ledger 자동 연결
3. 실제 DB integration test 확장
4. Settlement 계산
```

이 내용은 Day21 이후에 이어집니다.

## Step 1. `Store` interface 추가

`Service`가 필요로 하는 저장 동작을 interface로 정의합니다.

```go
type Store interface {
	CreateTransaction(ctx context.Context, tx Transaction, entries []Entry) error
}
```

여기서 중요한 점은 interface가 Repository를 설명하는 것이 아니라 Service의 필요를 설명한다는 것입니다.

```text
Service 입장에서는 저장소가 DB인지 fake인지 중요하지 않다.
CreateTransaction을 할 수 있으면 된다.
```

## Step 2. `Service` 구조체 변경

기존:

```go
type Service struct{}
```

변경:

```go
type Service struct {
	store Store
}
```

`store Store`는 Service가 저장 작업을 맡길 대상입니다.

## Step 3. 생성자 변경

기존:

```go
func NewService() *Service {
	return &Service{}
}
```

변경:

```go
func NewService(store Store) *Service {
	return &Service{store: store}
}
```

이제 실제 서버에서는 나중에 아래처럼 연결할 수 있습니다.

```go
repo := ledger.NewRepository(db)
svc := ledger.NewService(repo)
```

테스트에서는 아래처럼 연결합니다.

```go
store := &fakeStore{}
svc := NewService(store)
```

## Step 4. `RecordTransaction` 추가

`RecordTransaction`은 오늘의 핵심 메서드입니다.

흐름:

```text
1. ValidateTransaction으로 entries 검증
2. store가 없으면 실패
3. store.CreateTransaction으로 저장
```

코드:

```go
func (s *Service) RecordTransaction(ctx context.Context, tx Transaction, entries []Entry) error {
	if err := s.ValidateTransaction(ctx, entries); err != nil {
		return err
	}

	if s.store == nil {
		return fmt.Errorf("ledger store가 필요합니다")
	}

	return s.store.CreateTransaction(ctx, tx, entries)
}
```

## Step 5. fakeStore 만들기

실제 DB를 쓰지 않고 Service 흐름만 테스트하기 위해 fakeStore를 만듭니다.

fakeStore는 아래를 기록합니다.

```text
1. 저장 메서드가 몇 번 호출되었는가?
2. 어떤 transaction이 넘어왔는가?
3. 어떤 entries가 넘어왔는가?
4. 저장소 에러를 일부러 반환할 수 있는가?
```

## 실행 명령

```bash
gofmt -w internal/ledger/service.go internal/ledger/service_test.go
go test ./internal/ledger -v
go test ./...
```

## 예상 결과

`go test ./internal/ledger -v`에서 아래 테스트가 보여야 합니다.

```text
TestServiceRecordTransaction/검증에_성공하면_저장소를_호출한다
TestServiceRecordTransaction/검증에_실패하면_저장소를_호출하지_않는다
TestServiceRecordTransaction/저장소_에러를_반환한다
```

## 자주 만나는 오류

### 1. `not enough arguments in call to NewService`

원인:

```text
NewService 생성자가 store를 받도록 바뀌었는데 기존 테스트에서 NewService()로 호출하고 있다.
```

해결:

```go
NewService(nil)
```

또는 fake store를 넣습니다.

```go
NewService(&fakeStore{})
```

### 2. `cannot use repo as Store`

원인:

```text
Repository에 CreateTransaction(ctx, tx, entries) 메서드가 없거나 시그니처가 다르다.
```

해결:

Day18의 `CreateTransaction` 메서드 시그니처를 확인합니다.

### 3. 검증 실패인데 fakeStore가 호출된다

원인:

```text
RecordTransaction에서 ValidateTransaction보다 store.CreateTransaction을 먼저 호출했을 가능성이 있다.
```

해결:

반드시 아래 순서를 지킵니다.

```text
ValidateTransaction
-> Store nil 체크
-> CreateTransaction
```

## 검증 방법

오늘의 검증 기준은 아래입니다.

```text
1. Service가 Store interface를 가진다.
2. RecordTransaction이 ValidateTransaction을 먼저 호출한다.
3. 검증 실패 시 fakeStore.calls가 0이다.
4. 검증 성공 시 fakeStore.calls가 1이다.
5. 저장소 에러가 Service 밖으로 반환된다.
```

## 실습산출물 작성 포인트

오늘 산출물에서는 아래를 자기 말로 정리합니다.

```text
Service와 Repository의 책임 차이
Store interface를 둔 이유
fakeStore가 필요한 이유
RecordTransaction의 실행 순서
검증 실패 시 저장소를 호출하면 안 되는 이유
```

## 커밋 메시지

구현까지 끝났다면 아래 형태로 작성합니다.

```text
feat: ledger service 저장 흐름 연결
```

문서만 먼저 반영하면 아래 형태를 사용합니다.

```text
docs: Day20 ledger service repository 연결 자료 추가
```

## 다음 작업 예고

Day21에서는 Payment `FINALIZED` 상태와 Ledger 기록을 어떻게 연결할지 설계합니다.

```text
Payment FINALIZED
-> Ledger Transaction 생성
-> Ledger Entries 생성
-> RecordTransaction 호출
```
