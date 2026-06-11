# Day 13 실습가이드 - Ledger 균형 검증 테스트 작성

관련 Jira: [SPN-30](https://aslan0.atlassian.net/browse/SPN-30)

Day13의 퇴근 후 실습은 작은 코드 작업 하나입니다.

```text
Ledger Transaction의 debit/credit 합계가 맞는지 검증하는 Service 메서드와 테스트를 작성한다.
```

## 실습 흐름

![Day13 Ledger 균형 검증 흐름](../../../confluence/diagrams/spn30-day13-ledger-balance-test.png)

## 사전 조건

Day12 실습이 완료되어 있어야 합니다.

아래 파일이 있어야 합니다.

```text
internal/ledger/ledger.go
```

그리고 최소한 아래 타입이 있어야 합니다.

```text
Account
Transaction
Entry
EntryDirectionDebit
EntryDirectionCredit
```

## 오늘 만들 코드의 위치

새로 만들 파일:

```text
internal/ledger/service.go
internal/ledger/service_test.go
```

## Step 1. `service.go` 작성

파일:

```text
internal/ledger/service.go
```

작성할 코드:

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
		return fmt.Errorf("context is required")
	}

	if err := ctx.Err(); err != nil {
		return err
	}

	if len(entries) < 2 {
		return fmt.Errorf("ledger transaction requires at least two entries")
	}

	totals := make(map[string]int64)

	for _, entry := range entries {
		if entry.Amount <= 0 {
			return fmt.Errorf("entry amount must be greater than zero")
		}

		if entry.Currency == "" {
			return fmt.Errorf("entry currency is required")
		}

		switch entry.Direction {
		case EntryDirectionDebit:
			totals[entry.Currency] += entry.Amount
		case EntryDirectionCredit:
			totals[entry.Currency] -= entry.Amount
		default:
			return fmt.Errorf("unknown entry direction: %s", entry.Direction)
		}
	}

	for currency, total := range totals {
		if total != 0 {
			return fmt.Errorf("ledger transaction is not balanced for %s", currency)
		}
	}

	return nil
}
```

## Step 2. 코드 해석

### `type Service struct{}`

아직 저장소나 외부 의존성이 없기 때문에 빈 구조체입니다.

```go
type Service struct{}
```

Java로 비유하면 아직 필드가 없는 서비스 클래스와 비슷합니다.

```java
public class LedgerService {
}
```

### `func NewService() *Service`

새로운 Service 인스턴스의 포인터를 반환합니다.

```go
return &Service{}
```

여기서 `&Service{}`는 “Service 구조체 값을 만들고, 그 주소를 반환한다”는 뜻입니다.

### `func (s *Service) ValidateTransaction(...) error`

`(s *Service)`는 receiver입니다.

Java의 instance method처럼 `Service`에 속한 메서드라고 보면 됩니다.

```go
svc := NewService()
err := svc.ValidateTransaction(ctx, entries)
```

### `totals := make(map[string]int64)`

통화별 합계를 저장하기 위한 map입니다.

```text
key   = currency
value = debit과 credit을 반영한 합계
```

예시:

```text
USDC -> 0
```

최종 합계가 0이면 debit과 credit이 균형을 이룬 것입니다.

### `switch entry.Direction`

Entry의 방향에 따라 합계를 다르게 반영합니다.

```go
case EntryDirectionDebit:
	totals[entry.Currency] += entry.Amount
case EntryDirectionCredit:
	totals[entry.Currency] -= entry.Amount
```

오늘 실습에서는 debit은 더하고 credit은 뺍니다.

최종 결과가 0이면 균형이 맞습니다.

## Step 3. `service_test.go` 작성

파일:

```text
internal/ledger/service_test.go
```

작성할 코드:

```go
package ledger

import (
	"context"
	"testing"
)

func TestServiceValidateTransaction(t *testing.T) {
	svc := NewService()
	ctx := context.Background()

	t.Run("debit과 credit 합계가 같으면 성공한다", func(t *testing.T) {
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
			t.Fatalf("expected transaction to be balanced: %v", err)
		}
	})

	t.Run("credit 합계가 부족하면 실패한다", func(t *testing.T) {
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
			t.Fatal("expected transaction to be unbalanced")
		}
	})

	t.Run("금액이 0이면 실패한다", func(t *testing.T) {
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
			t.Fatal("expected zero amount entry to fail")
		}
	})

	t.Run("알 수 없는 방향이면 실패한다", func(t *testing.T) {
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
			t.Fatal("expected unknown direction to fail")
		}
	})
}
```

## Step 4. 테스트 코드 해석

### `svc := NewService()`

Ledger Service를 하나 만듭니다.

`:=`는 Go의 짧은 변수 선언입니다.

```text
타입은 오른쪽 값을 보고 Go가 추론한다.
```

### `ctx := context.Background()`

가장 기본적인 context를 만듭니다.

오늘은 DB나 API 요청이 없지만, Service 메서드 모양을 실제 백엔드처럼 유지하기 위해 context를 넘깁니다.

### `entries := []Entry{...}`

`Entry` 여러 개를 담은 slice를 만듭니다.

Java의 `List<Entry>`와 비슷하게 생각하면 됩니다.

### `if err := svc.ValidateTransaction(ctx, entries); err != nil`

Go에서 자주 쓰는 에러 처리 패턴입니다.

```text
ValidateTransaction을 실행한다.
반환된 error를 err 변수에 담는다.
err가 nil이 아니면 실패로 처리한다.
```

## Step 5. 포맷 실행

프로젝트 루트에서 실행합니다.

```bash
gofmt -w internal/ledger/service.go internal/ledger/service_test.go
```

전체 Go 파일을 한 번에 포맷하려면 아래 명령도 가능합니다.

```bash
go fmt ./...
```

## Step 6. 테스트 실행

Ledger 패키지만 테스트합니다.

```bash
go test ./internal/ledger -v
```

전체 테스트도 실행합니다.

```bash
go test ./...
```

예상 결과:

```text
PASS
ok   github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger
```

## Step 7. 완성 기준

오늘 완성 기준:

```text
internal/ledger/service.go 파일이 있다.
internal/ledger/service_test.go 파일이 있다.
debit과 credit 합계가 같으면 테스트가 성공한다.
합계가 맞지 않으면 테스트가 실패한다.
0 이하 금액은 실패한다.
알 수 없는 direction은 실패한다.
go test ./... 가 성공한다.
```

## Step 8. 실습산출물 작성

`Day13_실습산출물.md`에는 5개 질문만 답합니다.

```text
1. 오늘 만든 Service 메서드는 어떤 규칙을 검증하는가?
2. debit과 credit 합계가 같아야 하는 이유는 무엇인가?
3. `map[string]int64`는 어떤 역할을 하는가?
4. 오늘 테스트 4개는 각각 어떤 버그를 막는가?
5. 아직 헷갈리는 Go 문법 또는 Ledger 개념은 무엇인가?
```

## Step 9. 커밋 메시지

코드 작업까지 완료했다면 아래 커밋 메시지를 사용합니다.

```bash
git status
git add internal/ledger/service.go internal/ledger/service_test.go
git commit -m "test: Ledger 균형 검증 테스트 추가"
```

산출물 문서를 함께 작성했다면 문서 커밋을 분리하는 것이 좋습니다.

```bash
git add docs/domain/07_Ledger_Core/Day13_Ledger_균형검증_테스트/Day13_실습산출물.md
git commit -m "docs: Day13 Ledger 균형 검증 산출물 정리"
```
