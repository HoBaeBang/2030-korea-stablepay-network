# Day 13 검증문제와 답변가이드

관련 Jira: [SPN-30](https://aslan0.atlassian.net/browse/SPN-30)

먼저 문제를 풀어보고, 그 다음 답변가이드와 비교합니다.

## 먼저 풀어볼 문제

1. Ledger Transaction에서 debit과 credit 합계가 같아야 하는 이유는 무엇인가?
2. `ValidateTransaction`이 `error`만 반환하는 이유는 무엇인가?
3. `map[string]int64`에서 key와 value는 각각 무엇을 의미하는가?
4. `switch entry.Direction`은 어떤 문제를 막기 위해 필요한가?
5. `context.Context`를 오늘 코드에 넣은 이유는 무엇인가?
6. `t.Run`에 한글 테스트 이름을 쓰면 어떤 장점이 있는가?
7. Day13에서 repository나 DB migration을 만들지 않는 이유는 무엇인가?

## 답변가이드

### 1. debit과 credit 합계가 같아야 하는 이유

Ledger는 돈의 이동 기록입니다.

하나의 Ledger Transaction 안에서 debit과 credit 합계가 다르면 돈이 생기거나 사라진 것처럼 기록될 수 있습니다.

그래서 균형이 맞지 않는 거래는 저장하면 안 됩니다.

### 2. `ValidateTransaction`이 `error`만 반환하는 이유

오늘 메서드는 새로운 값을 만들어내는 것이 아니라, 입력된 entries가 유효한지 검사합니다.

성공하면 `nil` error를 반환하고, 실패하면 실패 이유가 담긴 error를 반환합니다.

즉 반환값의 핵심은 “통과 또는 실패”입니다.

### 3. `map[string]int64`의 key와 value

key는 통화입니다.

예를 들어 `USDC`가 key가 될 수 있습니다.

value는 해당 통화의 합계입니다.

오늘 코드에서는 debit을 더하고 credit을 빼서 최종 합계가 0인지 확인합니다.

### 4. `switch entry.Direction`이 막는 문제

Entry의 방향은 `DEBIT` 또는 `CREDIT`이어야 합니다.

그런데 잘못된 문자열이 들어올 수 있습니다.

예를 들어 `UNKNOWN` 같은 값입니다.

`switch`의 `default`에서 이런 값을 error로 막습니다.

### 5. `context.Context`를 넣은 이유

오늘은 아직 DB나 외부 API를 호출하지 않지만, Service 메서드는 나중에 DB transaction, timeout, 요청 취소와 연결될 수 있습니다.

그래서 실제 백엔드 서비스 형태에 맞게 `context.Context`를 받도록 만들어둡니다.

### 6. 한글 테스트 이름의 장점

테스트가 문서처럼 읽힙니다.

실패했을 때 어떤 비즈니스 규칙이 깨졌는지 바로 알 수 있습니다.

학습 단계에서는 영어 도메인 용어보다 한글 설명이 이해를 빠르게 도와줍니다.

### 7. Day13에서 DB를 만들지 않는 이유

Day13의 목적은 저장이 아니라 핵심 도메인 규칙을 테스트로 고정하는 것입니다.

균형 검증 규칙이 먼저 안정되어야 이후 repository나 DB migration을 붙였을 때도 안전하게 확장할 수 있습니다.

## Day13 통과 기준

```text
debit과 credit 합계가 같아야 하는 이유를 설명할 수 있다.
ValidateTransaction의 흐름을 읽을 수 있다.
map을 이용해 통화별 합계를 계산하는 이유를 설명할 수 있다.
테스트 4개가 어떤 문제를 막는지 설명할 수 있다.
```
