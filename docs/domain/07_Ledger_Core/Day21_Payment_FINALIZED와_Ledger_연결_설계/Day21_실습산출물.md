# Day 21 실습산출물 - Payment FINALIZED와 Ledger 연결 설계

관련 Jira: [SPN-38](https://aslan0.atlassian.net/browse/SPN-38)

Day21 산출물은 `Payment FINALIZED`가 어떤 Ledger 기록으로 바뀌어야 하는지 자기 말로 정리하는 문서입니다.

오늘은 코드를 많이 작성하는 날이 아니라, 다음 구현 전에 연결 규칙을 정확히 이해하는 날입니다.

## 작성 전 확인

아래 파일을 먼저 확인합니다.

```bash
sed -n '1,220p' internal/payment/payment.go
sed -n '1,260p' internal/payment/service.go
sed -n '1,220p' internal/ledger/ledger.go
```

특히 아래를 확인합니다.

```text
Payment.StatusFinalized
Payment.Amount
Payment.Currency
Ledger Transaction.ReferenceType
Ledger Transaction.ReferenceID
Ledger Transaction.IdempotencyKey
Ledger Entry.Direction
Ledger Entry.Amount
```

## 예시 결제

아래 예시를 기준으로 답변을 작성합니다.

```text
payment_id = pay_123
amount = 10_000_000
currency = USDC
platform_fee = 200_000
merchant_amount = 9_800_000
```

## 1. Payment FINALIZED만으로 부족한 이유는 무엇인가?

작성 힌트:

```text
FINALIZED는 결제 상태를 말하지만,
돈이 어느 계정에서 어느 계정으로 어떤 의미로 이동했는지는 Ledger가 기록해야 한다는 점을 적는다.
```

내 답변:

```text

```

## 2. 위 예시 결제의 Ledger Transaction은 어떻게 만들어야 하는가?

작성 힌트:

```text
reference_type, reference_id, idempotency_key를 직접 적는다.
```

내 답변:

```text
reference_type =
reference_id =
idempotency_key =
```

## 3. 위 예시 결제의 Ledger Entries를 작성해보자.

작성 힌트:

```text
고객, 가맹점 지급 예정, 플랫폼 수수료 계정이 각각 어떤 방향과 금액을 가져야 하는지 적는다.
```

내 답변:

```text
1.
2.
3.
```

## 4. debit 합계와 credit 합계가 같은지 검증해보자.

작성 힌트:

```text
DEBIT total과 CREDIT total을 숫자로 계산한다.
```

내 답변:

```text
DEBIT total =
CREDIT total =
검증 결과 =
```

## 5. 같은 Payment FINALIZED가 두 번 처리되면 어떤 문제가 생기고, 무엇으로 막을 수 있는가?

작성 힌트:

```text
같은 결제가 Ledger에 두 번 기록되면 지급 예정 금액이 두 번 생길 수 있다는 점,
그리고 idempotency_key로 막는다는 점을 적는다.
```

내 답변:

```text

```

## 오늘 실행 결과

코드 변경이 없다면 실행하지 않아도 됩니다.

실행했다면 아래 결과를 적습니다.

```bash
go test ./internal/payment -v
go test ./internal/ledger -v
```

기록:

```text

```

## 아직 헷갈리는 부분

아래 후보 중 헷갈리는 것을 적습니다.

```text
FINALIZED와 SETTLED 차이
Payment와 Ledger 책임 차이
reference_type / reference_id
idempotency_key
DEBIT / CREDIT 방향
MerchantPending 계정
PlatformFee 계정
```

메모:

```text

```

## 정답/점검 가이드

먼저 `내 답변`을 작성한 뒤 아래 내용을 펼쳐서 비교합니다.

### 1. Payment FINALIZED만으로 부족한 이유는 무엇인가?

<details>
<summary>정답/점검 가이드 보기</summary>

`FINALIZED`는 결제가 확정되었다는 상태입니다.

하지만 그 상태만으로는 고객이 얼마를 지불했고, 가맹점에게 얼마가 지급 예정이며, 플랫폼 수수료가 얼마인지 명확히 분리되지 않습니다.

따라서 확정된 결제를 Ledger Transaction과 Ledger Entry로 기록해야 나중에 정산, 대사, 중복 방지에 사용할 수 있습니다.

</details>

### 2. 위 예시 결제의 Ledger Transaction은 어떻게 만들어야 하는가?

<details>
<summary>정답/점검 가이드 보기</summary>

```text
reference_type = PAYMENT
reference_id = pay_123
idempotency_key = payment:pay_123:finalized
```

`reference_type`과 `reference_id`는 이 원장 거래가 어떤 업무 사건에서 생겼는지 알려줍니다.

`idempotency_key`는 같은 결제를 두 번 원장에 반영하지 않기 위한 중복 방지 키입니다.

</details>

### 3. 위 예시 결제의 Ledger Entries를 작성해보자.

<details>
<summary>정답/점검 가이드 보기</summary>

```text
1. Customer 계정
   Direction = DEBIT
   Amount = 10_000_000
   Currency = USDC

2. MerchantPending 계정
   Direction = CREDIT
   Amount = 9_800_000
   Currency = USDC

3. PlatformFee 계정
   Direction = CREDIT
   Amount = 200_000
   Currency = USDC
```

고객이 낸 전체 금액이 DEBIT으로 기록되고, 그 금액이 가맹점 지급 예정 금액과 플랫폼 수수료로 CREDIT 분리됩니다.

</details>

### 4. debit 합계와 credit 합계가 같은지 검증해보자.

<details>
<summary>정답/점검 가이드 보기</summary>

```text
DEBIT total = 10_000_000
CREDIT total = 9_800_000 + 200_000 = 10_000_000
검증 결과 = 일치한다
```

Ledger는 항상 debit 합계와 credit 합계가 같아야 합니다.

그래야 돈이 어디선가 갑자기 생기거나 사라지지 않습니다.

</details>

### 5. 같은 Payment FINALIZED가 두 번 처리되면 어떤 문제가 생기고, 무엇으로 막을 수 있는가?

<details>
<summary>정답/점검 가이드 보기</summary>

같은 Payment FINALIZED가 두 번 처리되면 같은 결제 금액이 Ledger에 두 번 기록될 수 있습니다.

예를 들어 가맹점 지급 예정 금액 9.8 USDC가 두 번 생기면 시스템은 실제보다 더 많은 금액을 정산 대상으로 볼 수 있습니다.

이 문제는 `idempotency_key = payment:pay_123:finalized` 같은 고유 키로 막습니다.

DB에서는 이 값을 unique 제약으로 관리해 같은 업무 사건이 중복 저장되지 않게 할 수 있습니다.

</details>

## 추가 보충 정리

### Codex 점검

오늘 산출물에서 가장 중요한 문장은 아래입니다.

```text
Payment FINALIZED는 결제 상태의 확정이고,
Ledger 기록은 그 확정된 결제를 돈의 이동으로 해석하는 과정이다.
```

### 코드 확인 포인트

다음 구현 전에 아래를 확인합니다.

```text
- Payment가 FINALIZED 되는 위치는 어디인가?
- LedgerService.RecordTransaction이 준비되어 있는가?
- Ledger Transaction의 idempotency_key를 어떤 규칙으로 만들 것인가?
- Platform fee 계산은 어디에서 할 것인가?
- Payment 상태 변경과 Ledger 기록 실패를 어떻게 처리할 것인가?
```

### 다음 학습 포인트

Day22에서는 Ledger Core 전체를 중간 회고합니다.

```text
Day15 균형 검증
Day16 DB migration
Day17 Repository 초안
Day18 저장 구현
Day19 idempotency
Day20 Service -> Repository
Day21 Payment FINALIZED -> Ledger 설계
```
