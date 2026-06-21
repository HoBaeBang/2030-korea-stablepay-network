# Day 22 구현가이드 - Settlement 도메인 타입과 계산 서비스

관련 Jira: [SPN-39](https://aslan0.atlassian.net/browse/SPN-39)

새 Day22는 기존 Day23의 Settlement 도메인 타입과 기존 Day24의 계산 서비스를 하나의 구현 흐름으로 합친 일정입니다.

오늘은 Ledger에 기록된 `MERCHANT_PENDING` CREDIT 후보를 수취인과 통화 기준으로 검증하고, 하나의 Settlement Batch와 여러 Settlement Item으로 계산합니다.

## 오늘 만들 것

```text
새 패키지: internal/settlement

새 파일:
internal/settlement/settlement.go
internal/settlement/calculator.go
internal/settlement/calculator_test.go
```

## 먼저 작성할 전체 코드

아래 세 파일을 먼저 작성합니다. 코드를 작성한 뒤 하단 설명을 읽으면서 각 타입과 검증 규칙이 필요한 이유를 확인합니다.

## `settlement.go` 최종 완성본 전체

<details>
<summary><code>settlement.go</code> 최종 완성본 전체 보기</summary>

```go
package settlement

import "github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"

// Status는 정산 묶음이 현재 어떤 처리 단계에 있는지 나타낸다.
type Status string

const (
	// StatusDraft는 지급 가능 금액을 계산해 정산 묶음을 처음 만든 상태다.
	StatusDraft Status = "DRAFT"
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

// Item은 어떤 Ledger Entry가 정산 묶음에 포함되었는지 나타내는 근거다.
type Item struct {
	BatchID       string
	LedgerEntryID string
	Amount        int64
	Currency      string
}
```

</details>

## `calculator.go` 최종 완성본 전체

<details>
<summary><code>calculator.go</code> 최종 완성본 전체 보기</summary>

```go
package settlement

import (
	"context"
	"fmt"
	"strings"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"
)

// Calculator는 Ledger 후보를 검증하고 Settlement Batch와 Items를 계산한다.
type Calculator struct{}

// NewCalculator는 Settlement Calculator 인스턴스를 만든다.
func NewCalculator() *Calculator {
	return &Calculator{}
}

// BuildBatch는 같은 수취인과 통화의 정산 후보를 하나의 묶음으로 계산한다.
func (c *Calculator) BuildBatch(
	ctx context.Context,
	batchID string,
	recipientID string,
	currency string,
	candidates []Candidate,
) (*Batch, []Item, error) {
	if ctx == nil {
		return nil, nil, fmt.Errorf("context가 필요합니다")
	}

	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}

	batchID = strings.TrimSpace(batchID)
	recipientID = strings.TrimSpace(recipientID)
	currency = strings.ToUpper(strings.TrimSpace(currency))

	if batchID == "" {
		return nil, nil, fmt.Errorf("settlement batch id가 필요합니다")
	}

	if recipientID == "" {
		return nil, nil, fmt.Errorf("settlement recipient id가 필요합니다")
	}

	if currency == "" {
		return nil, nil, fmt.Errorf("settlement currency가 필요합니다")
	}

	if len(candidates) == 0 {
		return nil, nil, fmt.Errorf("settlement candidate가 필요합니다")
	}

	seenEntryIDs := make(map[string]struct{}, len(candidates))
	items := make([]Item, 0, len(candidates))
	var totalAmount int64

	for _, candidate := range candidates {
		if candidate.LedgerEntryID == "" {
			return nil, nil, fmt.Errorf("ledger entry id가 필요합니다")
		}

		if _, exists := seenEntryIDs[candidate.LedgerEntryID]; exists {
			return nil, nil, fmt.Errorf("중복된 ledger entry입니다: %s", candidate.LedgerEntryID)
		}
		seenEntryIDs[candidate.LedgerEntryID] = struct{}{}

		if candidate.RecipientID != recipientID {
			return nil, nil, fmt.Errorf("다른 수취인의 settlement candidate가 포함되어 있습니다")
		}

		if candidate.AccountType != ledger.AccountTypeMerchantPending {
			return nil, nil, fmt.Errorf("merchant pending 계정만 정산할 수 있습니다")
		}

		if candidate.Direction != ledger.EntryDirectionCredit {
			return nil, nil, fmt.Errorf("credit 원장 항목만 정산할 수 있습니다")
		}

		if candidate.Amount <= 0 {
			return nil, nil, fmt.Errorf("정산 금액은 0보다 커야 합니다")
		}

		candidateCurrency := strings.ToUpper(strings.TrimSpace(candidate.Currency))
		if candidateCurrency != currency {
			return nil, nil, fmt.Errorf("다른 통화의 settlement candidate가 포함되어 있습니다")
		}

		totalAmount += candidate.Amount
		items = append(items, Item{
			BatchID:       batchID,
			LedgerEntryID: candidate.LedgerEntryID,
			Amount:        candidate.Amount,
			Currency:      currency,
		})
	}

	batch := &Batch{
		ID:          batchID,
		RecipientID: recipientID,
		Currency:    currency,
		TotalAmount: totalAmount,
		ItemCount:   len(items),
		Status:      StatusDraft,
	}

	return batch, items, nil
}
```

</details>

## `calculator_test.go` 최종 완성본 전체

<details>
<summary><code>calculator_test.go</code> 최종 완성본 전체 보기</summary>

```go
package settlement

import (
	"context"
	"testing"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"
)

func newTestCalculator(t *testing.T) (*Calculator, context.Context) {
	t.Helper()

	return NewCalculator(), context.Background()
}

func validCandidates() []Candidate {
	return []Candidate{
		{
			LedgerEntryID: "led_entry_payment_1",
			RecipientID:   "merchant_1",
			AccountType:   ledger.AccountTypeMerchantPending,
			Direction:     ledger.EntryDirectionCredit,
			Amount:        9_800_000,
			Currency:      "USDC",
		},
		{
			LedgerEntryID: "led_entry_payment_2",
			RecipientID:   "merchant_1",
			AccountType:   ledger.AccountTypeMerchantPending,
			Direction:     ledger.EntryDirectionCredit,
			Amount:        5_000_000,
			Currency:      "USDC",
		},
	}
}

func TestCalculatorBuildBatch(t *testing.T) {
	t.Run("같은 수취인과 통화의 후보를 정산 묶음으로 계산한다", func(t *testing.T) {
		calculator, ctx := newTestCalculator(t)

		batch, items, err := calculator.BuildBatch(
			ctx,
			"stl_batch_1",
			"merchant_1",
			"USDC",
			validCandidates(),
		)
		if err != nil {
			t.Fatalf("정상 정산 후보 계산에 실패했습니다: %v", err)
		}

		if batch.TotalAmount != 14_800_000 {
			t.Fatalf("정산 총액은 14_800_000이어야 하는데 %d입니다", batch.TotalAmount)
		}

		if batch.ItemCount != 2 || len(items) != 2 {
			t.Fatalf("정산 항목은 2개여야 합니다: batch=%d, items=%d", batch.ItemCount, len(items))
		}

		if batch.Status != StatusDraft {
			t.Fatalf("새 정산 묶음 상태는 DRAFT여야 하는데 %s입니다", batch.Status)
		}
	})

	t.Run("merchant pending 계정이 아니면 실패한다", func(t *testing.T) {
		calculator, ctx := newTestCalculator(t)
		candidates := validCandidates()
		candidates[0].AccountType = ledger.AccountTypeCustomer

		_, _, err := calculator.BuildBatch(
			ctx,
			"stl_batch_invalid_account",
			"merchant_1",
			"USDC",
			candidates,
		)
		if err == nil {
			t.Fatal("merchant pending 계정이 아니면 실패해야 합니다")
		}
	})

	t.Run("credit 원장 항목이 아니면 실패한다", func(t *testing.T) {
		calculator, ctx := newTestCalculator(t)
		candidates := validCandidates()
		candidates[0].Direction = ledger.EntryDirectionDebit

		_, _, err := calculator.BuildBatch(
			ctx,
			"stl_batch_invalid_direction",
			"merchant_1",
			"USDC",
			candidates,
		)
		if err == nil {
			t.Fatal("credit 원장 항목이 아니면 실패해야 합니다")
		}
	})

	t.Run("같은 ledger entry가 두 번 포함되면 실패한다", func(t *testing.T) {
		calculator, ctx := newTestCalculator(t)
		candidates := validCandidates()
		candidates[1].LedgerEntryID = candidates[0].LedgerEntryID

		_, _, err := calculator.BuildBatch(
			ctx,
			"stl_batch_duplicate",
			"merchant_1",
			"USDC",
			candidates,
		)
		if err == nil {
			t.Fatal("같은 ledger entry가 두 번 포함되면 실패해야 합니다")
		}
	})

	t.Run("다른 수취인의 후보가 섞이면 실패한다", func(t *testing.T) {
		calculator, ctx := newTestCalculator(t)
		candidates := validCandidates()
		candidates[1].RecipientID = "merchant_2"

		_, _, err := calculator.BuildBatch(
			ctx,
			"stl_batch_mixed_recipient",
			"merchant_1",
			"USDC",
			candidates,
		)
		if err == nil {
			t.Fatal("다른 수취인의 후보가 섞이면 실패해야 합니다")
		}
	})
}
```

</details>

## 코드 해설과 개념 이해

## Settlement란 무엇인가?

`Settlement`는 한글로 **정산**입니다. StablePay에서는 Ledger에 확정된 가맹점 지급 예정 금액을 모아 실제 지급 가능한 묶음으로 계산하는 과정입니다.

```text
Payment FINALIZED
-> Ledger에 돈의 이동 기록
-> MerchantPending CREDIT 후보 조회
-> Settlement Calculator 검증
-> Batch와 Items 계산
-> 이후 승인·지급 처리
```

![Day22 Settlement 계산 흐름](https://raw.githubusercontent.com/HoBaeBang/2030-korea-stablepay-network/main/docs/confluence/diagrams/spn39-day22-settlement-calculation.png)

오늘은 실제 지급이나 DB 저장까지 만들지 않습니다. **어떤 Ledger Entry를 하나의 정산 묶음으로 계산할 수 있는가**에 집중합니다.

## 핵심 용어

| 영어 | 한글 의미 | 오늘 코드에서의 역할 |
| --- | --- | --- |
| Settlement | 정산 | 지급 가능한 금액을 계산하고 묶는 과정 |
| Candidate | 후보 | 아직 정산에 포함되기 전의 Ledger Entry 정보 |
| Recipient | 수취인 | 정산금을 받을 주체. 현재 예시는 가맹점 |
| Batch | 묶음 | 같은 수취인·통화의 정산 결과 전체 |
| Item | 항목 | Batch 금액의 근거가 된 개별 Ledger Entry |
| DRAFT | 초안 | 계산은 되었지만 아직 승인·지급되지 않은 상태 |

## 왜 `Candidate`가 필요한가?

Ledger의 `Entry`에는 `AccountID`가 있지만 그 계정의 실제 소유자 `OwnerID`는 `Account`에 있습니다. 정산 후보를 조회할 때는 DB에서 Ledger Entry와 Account를 함께 조회해야 합니다.

```text
ledger_entries
JOIN ledger_accounts
-> Entry ID
-> Account Type
-> Owner/Recipient ID
-> Direction
-> Amount
-> Currency
```

`Candidate`는 이 조회 결과를 Settlement 계산기가 받기 좋은 입력 형태로 표현합니다. 오늘은 Repository 조회를 만들지 않고 테스트에서 후보를 직접 전달합니다.

## 왜 `MERCHANT_PENDING` CREDIT만 받는가?

Day21 결제 예시에서 가맹점 몫은 다음과 같이 기록됐습니다.

```text
Customer          DEBIT  10.0 USDC
MerchantPending   CREDIT  9.8 USDC
PlatformFee       CREDIT  0.2 USDC
```

`MerchantPending CREDIT`은 가맹점에게 귀속되었지만 아직 지급되지 않은 금액입니다. Customer DEBIT이나 PlatformFee CREDIT까지 가맹점 정산에 포함하면 잘못된 지급액이 만들어집니다.

## `Batch`와 `Item`을 분리하는 이유

```text
Batch stl_batch_1
recipient = merchant_1
currency = USDC
total = 14.8 USDC
item_count = 2

Items
- led_entry_payment_1: 9.8 USDC
- led_entry_payment_2: 5.0 USDC
```

Batch만 저장하면 총액은 알 수 있지만 어떤 Ledger Entry가 근거인지 알 수 없습니다. Item을 함께 관리해야 금액 근거를 추적하고 같은 Entry가 다른 정산에 중복 포함되는 것을 막을 수 있습니다.

## `BuildBatch` 실행 순서

```text
Context 확인
-> Batch ID, Recipient ID, Currency 정규화·검증
-> 후보 존재 여부 확인
-> Ledger Entry ID 중복 확인
-> 수취인 일치 확인
-> MERCHANT_PENDING 계정 확인
-> CREDIT 방향 확인
-> 양수 금액 확인
-> 통화 일치 확인
-> Item 생성과 TotalAmount 합계
-> DRAFT Batch 반환
```

### `seenEntryIDs`는 무엇인가?

```go
seenEntryIDs := make(map[string]struct{}, len(candidates))
```

Ledger Entry ID를 key로 사용하는 Set 역할의 map입니다. 이미 처리한 ID가 다시 나오면 같은 돈의 이동을 두 번 합산할 위험이 있으므로 즉시 실패합니다.

```go
if _, exists := seenEntryIDs[candidate.LedgerEntryID]; exists {
	return nil, nil, fmt.Errorf("중복된 ledger entry입니다: %s", candidate.LedgerEntryID)
}
```

`struct{}`는 저장할 데이터가 없는 빈 구조체입니다. 여기서는 값보다 “이 key를 이미 봤는가”만 필요하므로 사용합니다.

## 오늘 구현 경계

오늘 포함:

```text
Settlement 타입
정산 후보 검증
총액 계산
Batch와 Items 생성
단위 테스트
```

오늘 제외:

```text
Settlement DB migration과 Repository
Ledger 후보 조회 SQL
READY / APPROVED / PAID 상태 전이
실제 은행 또는 온체인 지급
Reconciliation
```

## 실습 순서

### Step 1. 패키지 생성

```bash
mkdir -p internal/settlement
```

### Step 2. 위 전체 코드 작성

```text
internal/settlement/settlement.go
internal/settlement/calculator.go
internal/settlement/calculator_test.go
```

### Step 3. 전체 Go 파일 정리

```bash
gofmt -w ./internal/settlement
```

### Step 4. 집중 테스트

```bash
go test ./internal/settlement -v
```

예상 테스트:

```text
같은 수취인과 통화의 후보를 정산 묶음으로 계산한다
merchant pending 계정이 아니면 실패한다
credit 원장 항목이 아니면 실패한다
같은 ledger entry가 두 번 포함되면 실패한다
다른 수취인의 후보가 섞이면 실패한다
```

### Step 5. 전체 회귀 테스트

```bash
go test ./...
```

## 자주 만나는 오류

### `use of internal package not allowed`

프로젝트 루트가 아닌 별도 디렉터리에서 Settlement 파일만 떼어 실행하면 발생할 수 있습니다. 반드시 현재 Go module의 `internal/settlement`에서 실행합니다.

### `undefined: ledger.AccountTypeMerchantPending`

import 경로와 기존 상수 이름을 확인합니다.

```go
import "github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"
```

### 정상 후보인데 수취인 오류가 발생함

`BuildBatch`에 전달한 `recipientID`와 각 Candidate의 `RecipientID`가 같은지 확인합니다.

## 검증 기준

```text
1. Settlement와 Ledger의 책임 차이를 설명할 수 있다.
2. Candidate, Batch, Item의 역할을 설명할 수 있다.
3. MerchantPending CREDIT만 정산 후보가 되는 이유를 설명할 수 있다.
4. 9.8 USDC와 5.0 USDC를 14.8 USDC Batch로 계산할 수 있다.
5. 중복 Ledger Entry를 차단해야 하는 이유를 설명할 수 있다.
6. Settlement 테스트와 전체 테스트가 통과한다.
```

## 커밋 메시지

```bash
git add internal/settlement
git commit -m "feat: settlement 도메인 타입과 계산 서비스 추가"
```

## 다음 작업

새 Day23에서는 기존 Day25~26을 합쳐 Settlement DB migration과 상태 흐름을 구현합니다.

