# Day 13 개념학습 - Ledger 균형 검증과 테스트 우선 사고

관련 Jira: [SPN-30](https://aslan0.atlassian.net/browse/SPN-30)

## 1. 오늘 왜 균형 검증을 먼저 하는가?

Ledger는 단순히 “결제가 완료됐다”를 저장하는 기능이 아닙니다.

Ledger는 돈의 이동을 기록합니다.

돈의 이동을 기록하는 시스템에서 가장 위험한 버그는 아래 두 가지입니다.

```text
1. 돈이 실제보다 더 생긴 것처럼 기록되는 버그
2. 돈이 실제보다 사라진 것처럼 기록되는 버그
```

그래서 Ledger Transaction은 내부적으로 균형이 맞아야 합니다.

```text
debit 합계 = credit 합계
```

이 규칙을 먼저 테스트로 고정해두면 이후 DB 저장, Payment 연결, Settlement 연결을 확장할 때 기본 안전장치가 생깁니다.

## 2. debit과 credit을 어떻게 이해할까?

처음에는 회계 용어로 접근하면 어렵습니다.

우리 프로젝트에서는 일단 아래처럼 이해합니다.

| 용어 | 한글 감각 | 오늘의 단순 이해 |
| --- | --- | --- |
| debit | 차변 | 어떤 계정에서 빠지는 방향으로 먼저 이해한다 |
| credit | 대변 | 어떤 계정에 들어가는 방향으로 먼저 이해한다 |

정확한 회계에서는 계정 종류에 따라 debit/credit의 의미가 더 정교해집니다.

하지만 지금은 Ledger Core 첫 구현 단계이므로 아래 규칙을 먼저 고정합니다.

```text
하나의 Ledger Transaction 안에서 debit 총액과 credit 총액은 같아야 한다.
```

## 3. 예시로 보기

고객이 10 USDC를 결제하고, 플랫폼 수수료가 0.2 USDC라면 원장은 이렇게 쌓일 수 있습니다.

```text
Transaction: tx_ledger_001

Entry 1:
  account = customer_account
  direction = DEBIT
  amount = 10_000_000
  currency = USDC

Entry 2:
  account = merchant_pending_account
  direction = CREDIT
  amount = 9_800_000
  currency = USDC

Entry 3:
  account = platform_fee_account
  direction = CREDIT
  amount = 200_000
  currency = USDC
```

합계를 계산하면 아래와 같습니다.

```text
DEBIT  합계 = 10_000_000
CREDIT 합계 = 9_800_000 + 200_000 = 10_000_000
```

따라서 이 Ledger Transaction은 균형이 맞습니다.

## 4. 왜 `map[string]int64`를 쓰는가?

오늘 실습에서는 통화별 합계를 계산합니다.

USDC만 있을 때는 단순한 변수 두 개로도 가능합니다.

```go
var debitTotal int64
var creditTotal int64
```

하지만 Ledger는 나중에 여러 통화를 다룰 수 있습니다.

```text
USDC
KRW stablecoin
platform point
```

그래서 오늘은 통화별로 합계를 모으기 위해 `map[string]int64`를 사용합니다.

```go
totals := make(map[string]int64)
```

이 구조는 아래처럼 이해하면 됩니다.

```text
key   = currency
value = debit과 credit을 반영한 합계
```

예를 들어 USDC debit은 더하고, USDC credit은 빼면 최종 값이 0이어야 합니다.

```text
USDC: +10_000_000 - 9_800_000 - 200_000 = 0
```

## 5. 오늘 테스트가 잡는 버그

Day13 테스트는 최소한 아래 상황을 잡습니다.

| 테스트 상황 | 잡고 싶은 문제 |
| --- | --- |
| debit과 credit 합계가 같으면 성공 | 정상 거래가 막히지 않아야 한다 |
| credit이 부족하면 실패 | 돈이 사라진 것처럼 기록되는 문제를 막는다 |
| 금액이 0이면 실패 | 의미 없는 Entry를 막는다 |
| 알 수 없는 direction이면 실패 | 잘못된 문자열이 원장에 들어가는 것을 막는다 |

## 6. Go 테스트 이름을 한글로 쓰는 이유

Go 테스트에서는 `t.Run`을 사용할 수 있습니다.

```go
t.Run("debit과 credit 합계가 같으면 성공한다", func(t *testing.T) {
    // ...
})
```

테스트 이름을 한글로 쓰면 지금 학습 단계에서 장점이 큽니다.

```text
테스트가 문서처럼 읽힌다.
실패했을 때 어떤 규칙이 깨졌는지 바로 알 수 있다.
영어 도메인 용어를 몰라도 의도를 파악할 수 있다.
```

실무에서도 팀 컨벤션에 따라 한글 테스트명을 사용하는 경우가 있습니다.

우리 프로젝트에서는 학습 효과를 위해 한글 테스트명을 적극적으로 사용합니다.

## 7. context.Context는 왜 또 나오는가?

Day13 코드에서는 `ValidateTransaction(ctx, entries)`처럼 `context.Context`를 인자로 받습니다.

아직 DB나 외부 API를 호출하지 않기 때문에 꼭 필요해 보이지 않을 수 있습니다.

그래도 Service 메서드에 `context.Context`를 받게 해두는 이유는 이후 확장 때문입니다.

나중에 Ledger Service는 아래 일을 하게 됩니다.

```text
DB transaction 시작
ledger_transactions 저장
ledger_entries 저장
중복 요청 idempotency key 확인
요청 취소 또는 timeout 처리
```

이때 `context.Context`는 요청 취소, timeout, trace 정보를 전달하는 통로가 됩니다.

오늘은 작은 코드지만, 이후 실제 백엔드 서비스 형태로 확장될 수 있게 메서드 모양을 잡아둡니다.

## 8. 오늘의 결론

```text
Ledger의 첫 번째 안전장치는 균형 검증이다.
debit과 credit 합계가 맞지 않으면 원장 거래를 저장하면 안 된다.
이 규칙은 테스트로 먼저 고정하는 것이 좋다.
```
