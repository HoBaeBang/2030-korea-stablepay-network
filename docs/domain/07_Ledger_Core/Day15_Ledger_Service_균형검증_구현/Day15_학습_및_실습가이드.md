# Day 15 학습 및 실습가이드 - Ledger Service 균형 검증 구현

관련 Jira: SPN-32

Day15는 Ledger Core의 첫 번째 Service 규칙을 더 단단하게 만드는 날입니다.

Day12에서는 Ledger에서 사용할 도메인 타입을 만들었습니다.

```text
Account
Transaction
Entry
```

Day13에서는 Ledger Transaction의 debit/credit 합계가 맞는지 검증하는 기본 Service와 테스트를 만들었습니다.

Day15에서는 그 흐름을 이어서 아래 관점으로 점검하고 보강합니다.

```text
Ledger는 돈의 이동 기록이므로,
저장하기 전에 반드시 안전한 데이터인지 검증해야 한다.
```

## 오늘의 큰 그림

![Day15 Ledger Service 균형 검증 구현](../../../confluence/diagrams/spn32-day15-ledger-service-balance.png)

## 오늘 읽고 작업할 순서

이 문서는 기존의 `기초학습`, `개념학습`, `실습가이드`를 하나로 합친 문서입니다.

아래 순서대로 위에서 아래로 읽으면 됩니다.

```text
1. 오늘 왜 이 기능을 만드는지 이해한다.
2. Ledger Service가 맡는 책임을 이해한다.
3. debit/credit 균형 검증 규칙을 다시 확인한다.
4. context, map, 테스트 케이스를 코드와 연결한다.
5. service.go와 service_test.go를 점검하고 부족한 검증 조건을 보강한다.
6. gofmt와 go test로 결과를 확인한다.
7. 실습산출물을 작성한다.
```

## 1. 오늘 왜 이 기능을 만드는가?

Payment는 결제가 어떤 상태인지 알려줍니다.

예를 들면 아래와 같습니다.

```text
PENDING
ONCHAIN_DETECTED
FINALIZED
FAILED
```

하지만 Payment 상태만으로는 돈이 어떻게 이동했는지 충분히 설명하기 어렵습니다.

```text
누구의 돈이 줄었는가?
누구의 돈이 늘었는가?
플랫폼 수수료는 얼마인가?
가맹점에게 지급 예정인 금액은 얼마인가?
이 거래가 원장 규칙상 안전한가?
```

이 질문에 답하기 위해 Ledger가 필요합니다.

Ledger는 돈의 이동을 Entry 단위로 기록합니다.

그래서 Ledger에는 단순 CRUD보다 강한 규칙이 필요합니다.

```text
잘못된 Entry를 저장하지 않는다.
debit과 credit 합계가 맞지 않으면 실패시킨다.
테스트로 그 규칙을 계속 보장한다.
```

## 2. 오늘 만들거나 보강하는 범위

오늘 확인할 파일:

```text
internal/ledger/ledger.go
internal/ledger/service.go
internal/ledger/service_test.go
```

오늘 작업의 중심:

```text
ValidateTransaction
TestServiceValidateTransaction
```

오늘 수정하지 않는 것:

```text
DB migration
repository
HTTP API
Payment FINALIZED와 Ledger 자동 연결
Settlement
```

Day15의 목표는 저장 기능을 만드는 것이 아닙니다.

저장 전에 어떤 Ledger Transaction이 정상인지 판단하는 기준을 코드와 테스트로 고정하는 것입니다.

## 3. Ledger Service란 무엇인가?

Service는 도메인 규칙을 실행하는 영역입니다.

Java 백엔드 감각으로 보면 아래와 비슷합니다.

```java
public class LedgerService {
    public void validateTransaction(List<Entry> entries) {
        // 원장 거래가 유효한지 검증
    }
}
```

Go에서는 이렇게 표현합니다.

```go
type Service struct{}

func (s *Service) ValidateTransaction(ctx context.Context, entries []Entry) error {
    // 원장 거래가 유효한지 검증
}
```

`(s *Service)`는 receiver입니다.

Java의 instance method처럼 `Service`에 속한 메서드라고 생각하면 됩니다.

```go
svc := NewService()
err := svc.ValidateTransaction(ctx, entries)
```

## 4. 왜 DB 저장보다 검증이 먼저인가?

Ledger는 돈의 이동 기록입니다.

잘못된 Ledger Entry가 DB에 저장되면 이후에 정산, 대사, 장애 복구가 모두 흔들립니다.

그래서 순서는 아래가 되어야 합니다.

```text
Entry 생성
-> Service 검증
-> Repository 저장
-> DB 보존
```

Day15에서는 이 중 앞부분을 다룹니다.

```text
Entry 생성
-> Service 검증
```

저장소를 붙이기 전에 “어떤 데이터가 정상인지”를 먼저 고정해야 다음 단계가 흔들리지 않습니다.

## 5. Ledger Transaction이 균형을 맞춘다는 뜻

예를 들어 고객이 10 USDC를 결제하고, 플랫폼 수수료가 0.2 USDC라고 해봅니다.

Ledger Entry는 이렇게 나뉠 수 있습니다.

| 계정 | 방향 | 금액 |
| --- | --- | --- |
| 고객 계정 | DEBIT | 10 USDC |
| 가맹점 지급 예정 계정 | CREDIT | 9.8 USDC |
| 플랫폼 수수료 계정 | CREDIT | 0.2 USDC |

합계는 아래처럼 계산합니다.

```text
DEBIT  10.0
CREDIT  9.8
CREDIT  0.2

10.0 - 9.8 - 0.2 = 0
```

결과가 0이면 균형이 맞습니다.

이 프로젝트에서는 소수점 금액을 `float`으로 다루지 않습니다.

USDC의 최소 단위 정수로 다룹니다.

```text
10 USDC  = 10_000_000
9.8 USDC = 9_800_000
0.2 USDC = 200_000
```

그래서 Go 타입은 `int64`를 사용합니다.

## 6. `map[string]int64`를 쓰는 이유

오늘 구현에서는 통화별 합계를 계산합니다.

```go
totals := make(map[string]int64)
```

여기서 key와 value는 아래 의미입니다.

```text
key   = 통화, 예: "USDC"
value = 해당 통화의 debit/credit 계산 결과
```

왜 통화별로 나누냐면, 서로 다른 통화를 하나로 합치면 안 되기 때문입니다.

```text
10 USDC와 10 KRW는 같은 금액이 아니다.
```

따라서 `USDC`, `KRW`, `USDT` 같은 통화별로 각각 균형을 확인해야 합니다.

## 7. `context.Context`는 왜 받는가?

오늘 코드는 아직 DB나 외부 API를 직접 호출하지 않습니다.

그런데도 `context.Context`를 받습니다.

이유는 Service 메서드가 나중에 DB 저장, 트랜잭션 처리, 요청 취소 흐름과 연결될 수 있기 때문입니다.

```text
HTTP 요청이 취소된다.
-> context가 취소된다.
-> Service는 더 이상 작업하지 않아야 한다.
```

오늘 코드에서는 `ctx.Err()`를 확인합니다.

```go
if err := ctx.Err(); err != nil {
    return err
}
```

이 문장은 context가 이미 취소되었거나 timeout이 발생했다면 그 error를 그대로 반환한다는 뜻입니다.

## 8. 검증 순서

`ValidateTransaction`은 아래 순서로 검증합니다.

```text
1. context가 nil인지 확인한다.
2. context가 이미 취소되었는지 확인한다.
3. Entry가 최소 2개 이상인지 확인한다.
4. 각 Entry의 amount가 0보다 큰지 확인한다.
5. 각 Entry의 currency가 비어 있지 않은지 확인한다.
6. direction이 DEBIT 또는 CREDIT인지 확인한다.
7. 통화별 debit/credit 합계가 0인지 확인한다.
```

이 순서가 중요한 이유는 문제를 빨리 발견하기 위해서입니다.

예를 들어 Entry가 1개뿐이면 균형을 계산하기 전에 이미 잘못된 Ledger Transaction입니다.

## 9. 실습 전 현재 코드 확인

프로젝트 루트에서 시작합니다.

```bash
pwd
```

예상 위치:

```text
2030-korea-stablepay-network
```

아래 파일을 확인합니다.

```bash
ls internal/ledger
sed -n '1,240p' internal/ledger/service.go
sed -n '1,320p' internal/ledger/service_test.go
```

Day13을 완료했다면 아래 파일이 있어야 합니다.

```text
ledger.go
service.go
service_test.go
```

## 10. `service.go` 완성 기준

파일:

```text
internal/ledger/service.go
```

완성 기준 코드는 아래와 같습니다.

```go
package ledger

import (
	"context"
	"fmt"
)

// Service는 Ledger 도메인 규칙을 검증하고 실행한다.
type Service struct{}

// NewService는 Ledger Service 인스턴스를 만든다.
func NewService() *Service {
	return &Service{}
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
```

## 11. `service_test.go` 보강 기준

파일:

```text
internal/ledger/service_test.go
```

Day13에서 이미 기본 테스트를 작성했다면 Day15에서는 아래 케이스가 모두 있는지 확인합니다.

```text
debit과 credit 합계가 같으면 성공한다.
credit 합계가 부족하면 실패한다.
entry가 하나뿐이면 실패한다.
금액이 0이면 실패한다.
통화가 비어 있으면 실패한다.
알 수 없는 방향이면 실패한다.
context가 취소되었으면 실패한다.
```

테스트 파일의 완성 기준 코드는 아래와 같습니다.

```go
package ledger

import (
	"context"
	"testing"
)

func TestServiceValidateTransaction(t *testing.T) {
	t.Run("debit과 credit 합계가 같으면 성공한다", func(t *testing.T) {
		svc := NewService()

		entries := []Entry{
			{
				AccountID:  "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
			{
				AccountID:  "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    9_800_000,
				Currency:  "USDC",
			},
			{
				AccountID:  "acct_platform_fee",
				Direction: EntryDirectionCredit,
				Amount:    200_000,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(context.Background(), entries); err != nil {
			t.Fatalf("에러가 없어야 하는데 발생했습니다: %v", err)
		}
	})

	t.Run("credit 합계가 부족하면 실패한다", func(t *testing.T) {
		svc := NewService()

		entries := []Entry{
			{
				AccountID:  "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
			{
				AccountID:  "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    9_700_000,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(context.Background(), entries); err == nil {
			t.Fatal("에러가 발생해야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("entry가 하나뿐이면 실패한다", func(t *testing.T) {
		svc := NewService()

		entries := []Entry{
			{
				AccountID:  "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(context.Background(), entries); err == nil {
			t.Fatal("에러가 발생해야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("금액이 0이면 실패한다", func(t *testing.T) {
		svc := NewService()

		entries := []Entry{
			{
				AccountID:  "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    0,
				Currency:  "USDC",
			},
			{
				AccountID:  "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    0,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(context.Background(), entries); err == nil {
			t.Fatal("에러가 발생해야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("통화가 비어 있으면 실패한다", func(t *testing.T) {
		svc := NewService()

		entries := []Entry{
			{
				AccountID:  "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    10_000_000,
				Currency:  "",
			},
			{
				AccountID:  "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    10_000_000,
				Currency:  "",
			},
		}

		if err := svc.ValidateTransaction(context.Background(), entries); err == nil {
			t.Fatal("에러가 발생해야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("알 수 없는 방향이면 실패한다", func(t *testing.T) {
		svc := NewService()

		entries := []Entry{
			{
				AccountID:  "acct_customer_1",
				Direction: EntryDirection("UNKNOWN"),
				Amount:    10_000_000,
				Currency:  "USDC",
			},
			{
				AccountID:  "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(context.Background(), entries); err == nil {
			t.Fatal("에러가 발생해야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("context가 취소되었으면 실패한다", func(t *testing.T) {
		svc := NewService()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		entries := []Entry{
			{
				AccountID:  "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
			{
				AccountID:  "acct_merchant_pending_1",
				Direction: EntryDirectionCredit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
		}

		if err := svc.ValidateTransaction(ctx, entries); err == nil {
			t.Fatal("에러가 발생해야 하는데 nil이 반환되었습니다")
		}
	})
}
```

## 12. 테스트 코드 해석

### `t.Run`

하나의 테스트 함수 안에서 여러 케이스를 나누어 실행합니다.

```go
t.Run("금액이 0이면 실패한다", func(t *testing.T) {
    // ...
})
```

한글 이름을 쓰면 테스트 결과를 읽기 쉬워집니다.

### `context.Background()`

테스트에서 가장 기본으로 사용할 수 있는 빈 context입니다.

```go
svc.ValidateTransaction(context.Background(), entries)
```

### `context.WithCancel`

취소 가능한 context를 만듭니다.

```go
ctx, cancel := context.WithCancel(context.Background())
cancel()
```

`cancel()`을 호출하면 `ctx.Err()`가 nil이 아니게 됩니다.

그래서 Service가 취소된 요청을 계속 처리하지 않는지 테스트할 수 있습니다.

### `err == nil`

실패해야 하는 케이스에서는 error가 있어야 합니다.

따라서 아래처럼 검사합니다.

```go
if err := svc.ValidateTransaction(context.Background(), entries); err == nil {
	t.Fatal("에러가 발생해야 하는데 nil이 반환되었습니다")
}
```

뜻:

```text
검증했는데 에러가 없다면 테스트 실패
```

## 13. 포맷 정리

Ledger 패키지 파일만 포맷합니다.

```bash
gofmt -w internal/ledger/service.go internal/ledger/service_test.go
```

전체 Go 파일을 한 번에 포맷하려면 아래 명령도 가능합니다.

```bash
go fmt ./...
```

`gofmt -w`는 Go 코드의 들여쓰기, 공백, import 정렬을 Go 표준 스타일로 고쳐서 파일에 저장합니다.

## 14. 테스트 실행

Ledger 패키지만 테스트합니다.

```bash
go test ./internal/ledger -v
```

예상 결과:

```text
=== RUN   TestServiceValidateTransaction
=== RUN   TestServiceValidateTransaction/debit과_credit_합계가_같으면_성공한다
=== RUN   TestServiceValidateTransaction/credit_합계가_부족하면_실패한다
=== RUN   TestServiceValidateTransaction/entry가_하나뿐이면_실패한다
=== RUN   TestServiceValidateTransaction/금액이_0이면_실패한다
=== RUN   TestServiceValidateTransaction/통화가_비어_있으면_실패한다
=== RUN   TestServiceValidateTransaction/알_수_없는_방향이면_실패한다
=== RUN   TestServiceValidateTransaction/context가_취소되었으면_실패한다
--- PASS: TestServiceValidateTransaction
PASS
```

전체 테스트도 실행합니다.

```bash
go test ./...
```

전체 테스트가 성공하면 오늘 만든 Ledger Service가 기존 기능을 깨뜨리지 않은 것입니다.

## 15. 자주 만나는 오류

### `undefined: Entry`

`internal/ledger/ledger.go`에 `Entry` 타입이 없거나 패키지명이 다를 때 발생할 수 있습니다.

먼저 아래 파일을 확인합니다.

```bash
sed -n '1,240p' internal/ledger/ledger.go
```

### 테스트 이름이 깨져 보이는 경우

터미널 인코딩 문제일 수 있습니다.

테스트 자체가 실패한 것이 아니라면 우선 `PASS` 여부를 먼저 봅니다.

### `context가 취소되었으면 실패한다` 테스트가 실패하는 경우

`ValidateTransaction` 안에서 아래 코드가 있는지 확인합니다.

```go
if err := ctx.Err(); err != nil {
	return err
}
```

### `통화가 비어 있으면 실패한다` 테스트가 실패하는 경우

`ValidateTransaction` 안에서 아래 코드가 있는지 확인합니다.

```go
if entry.Currency == "" {
	return fmt.Errorf("원장 항목 통화가 필요합니다")
}
```

## 16. 완성본 확인

오늘 작업 후 파일 구조는 아래처럼 보여야 합니다.

```text
internal/ledger/
  ledger.go
  service.go
  service_test.go
```

핵심 함수:

```text
NewService
ValidateTransaction
TestServiceValidateTransaction
```

## 17. 완료 기준

아래를 모두 만족하면 Day15 완료입니다.

```text
service.go 검증 조건을 설명할 수 있다.
service_test.go에 7개 테스트 케이스가 있다.
go test ./internal/ledger -v가 성공한다.
go test ./...가 성공한다.
Day15 산출물 5문항을 작성한다.
```

## 18. 커밋 메시지

코드 작업을 완료했다면 아래 커밋 메시지를 사용합니다.

```bash
git add internal/ledger/service.go internal/ledger/service_test.go
git commit -m "test: Ledger 균형 검증 케이스 보강"
```

산출물 문서까지 함께 정리했다면 문서 커밋은 별도로 분리합니다.

```bash
git add docs/domain/07_Ledger_Core/Day15_Ledger_Service_균형검증_구현/Day15_실습산출물.md
git commit -m "docs: Day15 Ledger 실습 산출물 정리"
```

## 19. 다음 작업 예고

Day15가 끝나면 다음 단계는 Ledger 저장 구조입니다.

```text
Service가 검증한다.
Repository가 저장한다.
DB가 기록을 보존한다.
```

즉 Day16부터는 아래 DB 테이블 후보로 넘어갑니다.

```text
ledger_accounts
ledger_transactions
ledger_entries
```
