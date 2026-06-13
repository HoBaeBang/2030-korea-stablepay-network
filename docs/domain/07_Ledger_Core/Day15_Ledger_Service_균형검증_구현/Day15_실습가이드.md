# Day 15 실습가이드 - Ledger Service 균형 검증 구현

관련 Jira: SPN-32

Day15의 퇴근 후 실습은 작은 코드 작업 하나입니다.

```text
Ledger Transaction의 Entry 목록을 받아
debit/credit 균형이 맞는지 검증하는 Service와 테스트를 작성한다.
```

## 실습 흐름

![Day15 Ledger Service 균형 검증 구현](../../../confluence/diagrams/spn32-day15-ledger-service-balance.png)

## 사전 조건

프로젝트 루트에서 시작합니다.

```bash
pwd
```

예상 위치:

```text
2030-korea-stablepay-network
```

아래 파일이 있어야 합니다.

```text
internal/ledger/ledger.go
```

확인 명령:

```bash
ls internal/ledger
```

현재 코드 기준으로 `ledger.go`만 있어도 Day15 실습을 시작할 수 있습니다.

## 오늘 만들 파일

새로 만들 파일:

```text
internal/ledger/service.go
internal/ledger/service_test.go
```

이미 파일이 있다면 먼저 내용을 확인합니다.

```bash
sed -n '1,240p' internal/ledger/service.go
sed -n '1,320p' internal/ledger/service_test.go
```

이미 비슷한 코드가 있다면 그대로 덮어쓰기보다, 오늘 문서와 비교해서 빠진 테스트나 검증 조건만 보강합니다.

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

## Step 2. `service.go` 코드 해석

### `type Service struct{}`

아직 repository나 DB 의존성이 없기 때문에 빈 구조체입니다.

Java로 비유하면 아직 필드가 없는 서비스 클래스와 비슷합니다.

```java
public class LedgerService {
}
```

### `func NewService() *Service`

Service 값을 만들고 그 주소를 반환합니다.

```go
return &Service{}
```

여기서 `&Service{}`는 “Service 구조체 값을 만들고, 그 값의 주소를 반환한다”는 뜻입니다.

### `func (s *Service) ValidateTransaction(...) error`

`(s *Service)`는 receiver입니다.

Java의 instance method처럼 `Service`에 속한 메서드라고 보면 됩니다.

```go
svc := NewService()
err := svc.ValidateTransaction(ctx, entries)
```

### `totals := make(map[string]int64)`

통화별 합계를 저장하는 map입니다.

```text
totals["USDC"] = debit 합계 - credit 합계
```

최종 값이 0이면 균형이 맞습니다.

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
	t.Run("debit과 credit 합계가 같으면 성공한다", func(t *testing.T) {
		svc := NewService()

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
				AccountID: "acct_platform_fee",
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
				AccountID: "acct_customer_1",
				Direction: EntryDirectionDebit,
				Amount:    10_000_000,
				Currency:  "USDC",
			},
			{
				AccountID: "acct_merchant_pending_1",
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
				AccountID: "acct_customer_1",
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

		if err := svc.ValidateTransaction(context.Background(), entries); err == nil {
			t.Fatal("에러가 발생해야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("통화가 비어 있으면 실패한다", func(t *testing.T) {
		svc := NewService()

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

		if err := svc.ValidateTransaction(context.Background(), entries); err == nil {
			t.Fatal("에러가 발생해야 하는데 nil이 반환되었습니다")
		}
	})

	t.Run("알 수 없는 방향이면 실패한다", func(t *testing.T) {
		svc := NewService()

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
			t.Fatal("에러가 발생해야 하는데 nil이 반환되었습니다")
		}
	})
}
```

## Step 4. `service_test.go` 코드 해석

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

## Step 5. 포맷 정리

Ledger 패키지 파일만 포맷합니다.

```bash
gofmt -w internal/ledger
```

`gofmt -w`는 Go 코드의 들여쓰기, 공백, import 정렬을 Go 표준 스타일로 고쳐서 파일에 저장합니다.

## Step 6. Ledger 테스트 실행

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

## Step 7. 전체 테스트 실행

```bash
go test ./...
```

전체 테스트가 성공하면 오늘 만든 Ledger Service가 기존 기능을 깨뜨리지 않은 것입니다.

## Step 8. 완성본 확인

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

## Step 9. 커밋 메시지

코드 작업을 완료했다면 아래 커밋 메시지를 사용합니다.

```bash
git add internal/ledger/service.go internal/ledger/service_test.go
git commit -m "feat: Ledger 균형 검증 서비스 추가"
```

산출물 문서까지 함께 정리했다면 문서 커밋은 별도로 분리합니다.

```bash
git add docs/domain/07_Ledger_Core/Day15_Ledger_Service_균형검증_구현/Day15_실습산출물.md
git commit -m "docs: Day15 Ledger 실습 산출물 정리"
```

## 오늘의 완료 기준

아래를 모두 만족하면 Day15 완료입니다.

```text
service.go 작성 완료
service_test.go 작성 완료
go test ./internal/ledger -v 성공
go test ./... 성공
Day15 산출물 5문항 작성
```
