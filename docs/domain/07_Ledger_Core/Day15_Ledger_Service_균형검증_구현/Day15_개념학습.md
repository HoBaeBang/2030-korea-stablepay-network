# Day 15 개념학습 - Ledger Service와 균형 검증

관련 Jira: SPN-32

## 1. Day15에서 만드는 것은 무엇인가?

Day15에서는 Ledger의 첫 번째 Service를 만듭니다.

```text
Service = 도메인 규칙을 실행하는 영역
```

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

## 2. 왜 DB 저장보다 검증이 먼저인가?

Ledger는 돈의 이동 기록입니다.

잘못된 Ledger Entry가 DB에 저장되면 이후에 정산, 대사, 장애 복구가 모두 흔들립니다.

그래서 순서는 아래가 되어야 합니다.

```text
Entry 생성
-> Service 검증
-> Repository 저장
-> DB 보존
```

Day15에서는 이 중 앞부분만 구현합니다.

```text
Entry 생성
-> Service 검증
```

아직 DB에 저장하지 않는 이유는, 저장소를 붙이기 전에 “어떤 데이터가 정상인지”를 먼저 고정해야 하기 때문입니다.

## 3. Ledger Transaction이 균형을 맞춘다는 뜻

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

## 4. `map[string]int64`를 쓰는 이유

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

## 5. 검증 순서

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

## 6. `context.Context`는 왜 받는가?

오늘 코드는 DB나 외부 API를 호출하지 않습니다.

그런데도 `context.Context`를 받습니다.

이유는 Service 메서드가 나중에 DB 저장, 트랜잭션 처리, 요청 취소 흐름과 연결될 수 있기 때문입니다.

```text
HTTP 요청이 취소된다.
-> context가 취소된다.
-> Service는 더 이상 작업하지 않아야 한다.
```

오늘은 `ctx.Err()`만 확인합니다.

```go
if err := ctx.Err(); err != nil {
    return err
}
```

이 문장은 context가 이미 취소되었거나 timeout이 발생했다면 그 error를 그대로 반환한다는 뜻입니다.

## 7. 테스트는 무엇을 보장해야 하는가?

Day15 테스트는 성공 케이스보다 실패 케이스가 더 중요합니다.

Ledger는 “좋은 데이터가 잘 통과한다”도 중요하지만, “나쁜 데이터가 절대 통과하지 않는다”가 훨씬 중요합니다.

오늘 테스트 후보:

```text
균형이 맞으면 성공한다.
credit 합계가 부족하면 실패한다.
Entry가 1개뿐이면 실패한다.
금액이 0이면 실패한다.
통화가 비어 있으면 실패한다.
알 수 없는 방향이면 실패한다.
취소된 context면 실패한다.
```

테스트 이름은 한글 `t.Run`으로 작성합니다.

```go
t.Run("debit과 credit 합계가 같으면 성공한다", func(t *testing.T) {
    // ...
})
```

이렇게 하면 테스트 실행 결과만 봐도 어떤 규칙을 검증하는지 알 수 있습니다.

## 8. 오늘의 결론

```text
Day15의 핵심은 Ledger를 저장하는 것이 아니다.

저장하기 전에 Ledger Transaction이 안전한지 검증하는
첫 번째 Service 규칙을 코드와 테스트로 고정하는 것이다.
```

이 규칙이 있어야 다음 단계에서 DB migration과 repository를 만들 때 기준이 생깁니다.
