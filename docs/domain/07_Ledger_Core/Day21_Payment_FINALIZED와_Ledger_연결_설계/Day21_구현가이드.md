# Day 21 구현가이드 - Payment FINALIZED와 Ledger 연결 설계

관련 Jira: [SPN-38](https://aslan0.atlassian.net/browse/SPN-38)

Day21은 `Payment`가 `FINALIZED` 되었을 때 이 결제 결과를 `Ledger`에 어떻게 기록할지 설계하는 날입니다.

중요한 전제부터 확인합니다.

```text
Payment는 결제의 상태를 말한다.
Ledger는 돈의 이동을 말한다.
```

Phase 1에서는 `Payment` 상태를 만들었습니다.

```text
PENDING
-> ONCHAIN_DETECTED
-> FINALIZED
-> SETTLED
```

하지만 `Payment` 상태만으로는 아래 질문에 충분히 답하기 어렵습니다.

```text
누가 돈을 냈는가?
누구에게 지급 예정인가?
수수료는 얼마인가?
동일 결제를 두 번 원장에 반영하지 않았는가?
나중에 정산할 금액은 어디서 계산할 것인가?
```

그래서 Day21에서는 `FINALIZED` 결제가 Ledger 기록으로 바뀌는 흐름을 설계합니다.

## 오늘의 큰 그림

![Day21 Payment FINALIZED와 Ledger 연결](https://raw.githubusercontent.com/HoBaeBang/2030-korea-stablepay-network/main/docs/confluence/diagrams/spn38-day21-payment-ledger-link.png)

## 오늘 만들 것

오늘은 코드 파일을 직접 수정하는 날이 아니라, 다음 구현을 위한 연결 설계와 테스트 후보를 정리하는 날입니다.

```text
작성 문서:
docs/domain/07_Ledger_Core/Day21_Payment_FINALIZED와_Ledger_연결_설계/Day21_구현가이드.md
docs/domain/07_Ledger_Core/Day21_Payment_FINALIZED와_Ledger_연결_설계/Day21_실습산출물.md

확인할 코드:
internal/payment/payment.go
internal/payment/service.go
internal/ledger/ledger.go
internal/ledger/service.go
```

오늘 직접 구현하지 않는 이유는 현재 실제 코드에는 아직 Day20의 `RecordTransaction` 연결이 반영되어 있지 않기 때문입니다.

따라서 Day21의 목표는 아래처럼 잡습니다.

```text
Day20 구현이 끝난 뒤,
Payment FINALIZED 이벤트를 어떤 Ledger Transaction과 Entries로 바꿀지 설계한다.
```

## 왜 이 기능이 필요한가?

`Payment`가 `FINALIZED` 되었다는 것은 “결제가 확정되었다”는 뜻입니다.

하지만 결제가 확정되었다고 해서 자동으로 정산 가능한 금액이 계산되는 것은 아닙니다.

정산하려면 먼저 돈의 이동이 원장에 기록되어야 합니다.

예를 들어 고객이 10 USDC를 결제했고, 플랫폼 수수료가 0.2 USDC라고 해보겠습니다.

```text
Payment:
amount = 10 USDC
status = FINALIZED
```

이 정보만 보면 전체 결제 금액은 알 수 있지만, 수수료와 가맹점 지급 예정 금액은 명확히 분리되어 있지 않습니다.

Ledger에는 아래처럼 기록해야 합니다.

```text
DEBIT  Customer           10.0 USDC
CREDIT MerchantPending     9.8 USDC
CREDIT PlatformFee         0.2 USDC
```

이렇게 해야 나중에 Settlement가 아래 질문에 답할 수 있습니다.

```text
이 가맹점에게 지급 가능한 금액은 얼마인가?
플랫폼 수수료는 얼마인가?
이미 원장에 반영된 결제를 또 반영하지 않았는가?
```

## 출퇴근 예습 포인트

출퇴근 시간에는 아래 질문을 생각하면서 읽습니다.

```text
1. Payment FINALIZED는 왜 Ledger 기록의 시작점인가?
2. Payment와 Ledger는 왜 같은 테이블로 합치면 안 되는가?
3. Ledger Transaction의 reference_type, reference_id는 왜 필요한가?
4. idempotency_key는 왜 payment id 기반으로 만들 수 있는가?
5. 고객, 가맹점 지급 예정, 플랫폼 수수료 계정은 각각 어떤 방향의 entry가 되는가?
```

## 핵심 용어

| 용어 | 한글 의미 | 오늘 문맥에서의 의미 |
| --- | --- | --- |
| FINALIZED | 확정됨 | 온체인 결제가 충분히 확정되어 더 이상 결제 실패로 보기 어려운 상태 |
| Trigger | 방아쇠, 시작 조건 | Payment가 FINALIZED 되었을 때 Ledger 기록을 시작하는 조건 |
| Reference | 참조 | Ledger Transaction이 어떤 원인에서 생겼는지 가리키는 정보 |
| Idempotency Key | 멱등성 키 | 같은 결제를 두 번 Ledger에 기록하지 않기 위한 중복 방지 키 |
| Merchant Pending | 가맹점 지급 예정 | 아직 정산 전이지만 가맹점에게 지급될 금액을 모아두는 원장 계정 |
| Platform Fee | 플랫폼 수수료 | 결제 금액 중 서비스가 가져갈 수수료 |

## 현재 코드에서 확인할 부분

### `internal/payment/payment.go`

확인할 타입:

```go
type Payment struct {
	ID              string
	InvoiceID       string
	Amount          int64
	Currency        string
	Status          Status
	TransactionHash *string
	FinalizedAt     *time.Time
	CreatedAt       time.Time
}
```

확인 포인트:

```text
Payment에는 amount, currency, status가 있다.
하지만 수수료 금액, 가맹점 지급 예정 계정, ledger transaction id는 아직 없다.
```

### `internal/payment/service.go`

확인할 상태 전이:

```go
case StatusOnchainDetected:
	return to == StatusFinalized || to == StatusFailed
case StatusFinalized:
	return to == StatusSettled
```

확인 포인트:

```text
ONCHAIN_DETECTED에서 FINALIZED로 갈 수 있다.
FINALIZED 이후에는 SETTLED로 갈 수 있다.
즉 FINALIZED는 Settlement 직전의 중요한 경계다.
```

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
Ledger Transaction은 여러 Entry를 하나로 묶는다.
ReferenceType과 ReferenceID는 이 Ledger가 어떤 업무 사건에서 생겼는지 알려준다.
```

## Payment FINALIZED를 Ledger로 바꾸는 규칙

Day21에서는 아래 규칙을 후보로 둡니다.

### 1. Ledger Transaction

```text
ReferenceType  = "PAYMENT"
ReferenceID    = payment.ID
IdempotencyKey = "payment:{payment.ID}:finalized"
```

예:

```text
payment.ID = pay_123

ReferenceType  = PAYMENT
ReferenceID    = pay_123
IdempotencyKey = payment:pay_123:finalized
```

이렇게 하면 나중에 Ledger를 볼 때 아래 질문에 답할 수 있습니다.

```text
이 원장 거래는 어떤 결제 때문에 생겼는가?
```

그리고 idempotency key로 아래 질문에도 답할 수 있습니다.

```text
이 결제를 이미 원장에 반영했는가?
```

### 2. Ledger Entries

예시 결제:

```text
payment amount = 10_000_000
currency = USDC
platform fee = 200_000
merchant net amount = 9_800_000
```

Ledger Entry 후보:

| 계정 | 방향 | 금액 | 의미 |
| --- | --- | ---: | --- |
| Customer | DEBIT | 10_000_000 | 고객이 결제 금액만큼 지불했다 |
| MerchantPending | CREDIT | 9_800_000 | 가맹점에게 지급 예정 금액이 생겼다 |
| PlatformFee | CREDIT | 200_000 | 플랫폼 수수료 수익이 생겼다 |

검증:

```text
DEBIT 합계  = 10_000_000
CREDIT 합계 = 9_800_000 + 200_000 = 10_000_000
```

따라서 Ledger 균형이 맞습니다.

## 왜 Payment 테이블에 다 넣지 않는가?

Payment에 아래 필드를 계속 추가할 수도 있습니다.

```text
merchant_amount
platform_fee
settlement_amount
ledger_recorded_at
ledger_transaction_id
```

하지만 이렇게 하면 Payment가 너무 많은 책임을 갖게 됩니다.

```text
Payment:
- 결제 상태
- 온체인 감지 상태
- 수수료 분리
- 원장 기록
- 정산 기준
```

이러면 나중에 Deposit, Withdrawal, Refund 같은 돈의 이동도 모두 Payment에 억지로 넣게 됩니다.

Ledger를 따로 두면 책임이 분리됩니다.

```text
Payment는 결제 상태를 담당한다.
Ledger는 모든 돈의 이동 기록을 담당한다.
Settlement는 Ledger를 기반으로 지급 가능 금액을 계산한다.
```

## 다음 구현 후보

Day21 자체에서는 실제 코드를 수정하지 않습니다.

다만 다음 구현에서는 아래 중 하나의 방향을 선택할 수 있습니다.

### 후보 A. Payment Service가 Ledger Service를 직접 호출

```text
PaymentService.UpdatePaymentStatus(FINALIZED)
-> LedgerService.RecordTransaction(...)
```

장점:

```text
구현이 단순하다.
한 요청 흐름 안에서 Payment 상태 변경과 Ledger 기록을 이해하기 쉽다.
```

주의점:

```text
Payment DB update와 Ledger DB insert가 같은 transaction으로 묶이지 않으면 중간 실패 처리가 복잡해질 수 있다.
```

### 후보 B. Payment FINALIZED 이벤트를 발행하고 별도 handler가 Ledger를 기록

```text
Payment FINALIZED
-> PaymentFinalized event
-> Ledger handler
-> RecordTransaction
```

장점:

```text
Payment와 Ledger의 결합이 낮아진다.
나중에 알림, 정산 준비, 분석 이벤트를 붙이기 쉽다.
```

주의점:

```text
이벤트 저장, 재처리, 중복 처리 방지가 더 중요해진다.
```

현재 학습 프로젝트에서는 먼저 후보 A로 이해하고, 나중에 이벤트 방식으로 확장하는 것이 좋습니다.

## 다음 구현에서 만들 수 있는 테스트 후보

아래 테스트는 Day21 이후 구현 후보입니다.

```text
1. Payment가 FINALIZED 되면 Ledger RecordTransaction이 호출된다.
2. Payment가 FINALIZED가 아니면 Ledger는 호출되지 않는다.
3. 같은 Payment FINALIZED를 두 번 처리해도 idempotency_key 때문에 중복 저장되지 않는다.
4. Ledger 기록 실패 시 Payment 상태를 어떻게 할지 정책을 정한다.
```

특히 4번은 아직 결정이 필요합니다.

```text
Payment를 FINALIZED로 바꾼 뒤 Ledger 기록에 실패하면?
Payment 상태 변경과 Ledger 기록을 하나의 DB transaction으로 묶을 것인가?
Ledger 기록 실패 시 재시도 큐를 둘 것인가?
```

Day21에서는 이 문제를 “앞으로 설계해야 할 위험”으로 기억합니다.

## 실습 절차

오늘은 코드를 직접 수정하지 않고 아래 순서로 실습합니다.

### Step 1. 현재 Payment 상태 전이 확인

```bash
sed -n '1,220p' internal/payment/service.go
```

확인할 부분:

```text
CanTransition에서 ONCHAIN_DETECTED -> FINALIZED가 허용되는가?
FINALIZED -> SETTLED가 허용되는가?
```

### Step 2. 현재 Ledger 타입 확인

```bash
sed -n '1,220p' internal/ledger/ledger.go
```

확인할 부분:

```text
Transaction에 ReferenceType, ReferenceID, IdempotencyKey가 있는가?
Entry에 AccountID, Direction, Amount, Currency가 있는가?
```

### Step 3. 예시 결제를 Ledger로 손으로 변환

아래 결제를 기준으로 산출물에 직접 작성합니다.

```text
payment_id = pay_123
amount = 10_000_000
currency = USDC
platform_fee = 200_000
merchant_amount = 9_800_000
```

작성할 것:

```text
Ledger Transaction:
- reference_type
- reference_id
- idempotency_key

Ledger Entries:
- customer debit
- merchant pending credit
- platform fee credit
```

### Step 4. 균형 검증

아래 식이 맞는지 확인합니다.

```text
DEBIT total = CREDIT total
```

위 예시에서는 아래처럼 되어야 합니다.

```text
10_000_000 = 9_800_000 + 200_000
```

## 실행 명령

오늘은 코드 변경이 없으므로 필수 실행 명령은 없습니다.

다만 현재 테스트가 깨지지 않는지 확인하고 싶다면 아래를 실행합니다.

```bash
go test ./internal/payment -v
go test ./internal/ledger -v
```

## 검증 방법

오늘의 검증 기준은 아래입니다.

```text
1. Payment FINALIZED가 왜 Ledger 기록의 시작점인지 설명할 수 있다.
2. Ledger Transaction의 reference_type, reference_id, idempotency_key를 직접 만들 수 있다.
3. 고객, 가맹점 지급 예정, 플랫폼 수수료 entry의 debit/credit 방향을 설명할 수 있다.
4. debit 합계와 credit 합계가 같아야 하는 이유를 설명할 수 있다.
5. Day22 이후 어떤 구현 위험을 조심해야 하는지 말할 수 있다.
```

## 자주 헷갈리는 부분

### 1. FINALIZED가 되면 바로 SETTLED인가?

아닙니다.

```text
FINALIZED:
결제가 확정되었다.

SETTLED:
가맹점에게 정산까지 완료되었다.
```

Day21은 `FINALIZED -> Ledger 기록`을 다루고, Settlement는 이후 단계입니다.

### 2. Customer DEBIT은 고객 돈이 늘어난다는 뜻인가?

아닙니다.

여기서는 “고객 쪽에서 금액이 빠져나간다”는 결제 시스템 관점의 기록으로 이해합니다.

중요한 것은 항상 아래 균형입니다.

```text
DEBIT 합계 = CREDIT 합계
```

### 3. idempotency_key는 왜 payment id만 쓰지 않는가?

`payment.ID`만 써도 중복 방지는 가능해 보입니다.

하지만 같은 payment에서 나중에 다른 목적의 원장 기록이 생길 수 있습니다.

예:

```text
payment:pay_123:finalized
payment:pay_123:refund
payment:pay_123:chargeback
```

그래서 업무 사건까지 포함한 문자열이 더 안전합니다.

## 커밋 메시지

문서만 반영하는 경우:

```text
docs: Day21 payment finalized ledger 연결 설계 자료 추가
```

나중에 실제 구현까지 하는 경우:

```text
feat: payment finalized ledger 기록 설계 반영
```

## 다음 작업 예고

Day22는 Ledger Core 중간 회고입니다.

Day15부터 Day21까지 이어진 흐름을 아래 순서로 다시 점검합니다.

```text
Ledger 타입
-> 균형 검증
-> DB migration
-> Repository 저장
-> Service 연결
-> Payment FINALIZED와 Ledger 연결 설계
```
