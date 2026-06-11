# Day 12 실습산출물 - Ledger 도메인 타입 초안

관련 Jira: [SPN-29](https://aslan0.atlassian.net/browse/SPN-29)

Day12 산출물은 5개 질문만 작성합니다.

## 1. 오늘 만든 타입은 무엇인가?

작성 예시:

```text
오늘은 `internal/ledger/ledger.go`에 Account, Transaction, Entry 타입 초안을 만들었다.
이 타입들은 Ledger를 구현하기 전에 돈의 이동 기록을 코드로 표현하기 위한 기본 언어다.
```

내 답변:

```text

```

## 2. Account, Transaction, Entry는 각각 무엇인가?

작성 예시:

```text
Account는 돈이 기록되는 주체다.
Transaction은 여러 Entry를 하나로 묶는 원장 거래다.
Entry는 실제 돈의 이동 한 줄이다.
```

내 답변:

```text

```

## 3. Amount를 int64로 둔 이유는 무엇인가?

작성 예시:

```text
돈을 float으로 다루면 소수점 오차가 생길 수 있다.
그래서 USDC의 최소 단위 같은 정수 단위로 금액을 저장하기 위해 int64를 사용한다.
```

내 답변:

```text

```

## 4. 이 타입들이 다음 구현에서 어디로 이어지는가?

작성 예시:

```text
Account, Transaction, Entry 타입은 다음에 Ledger service 테스트와 DB migration으로 이어진다.
특히 Entry의 Direction과 Amount는 debit/credit 합계 검증 테스트에서 사용될 수 있다.
```

내 답변:

```text

```

## 5. 아직 헷갈리는 개념은 무엇인가?

작성 예시:

```text
debit과 credit의 방향이 아직 헷갈린다.
또한 고객 결제, 가맹점 지급 예정, 플랫폼 수수료가 각각 어떤 account에 기록되는지 예시가 더 필요하다.
```

내 답변:

```text

```

## 실행 결과

실행한 명령:

```bash
gofmt -w internal/ledger/ledger.go
go test ./...
```

결과:

```text

```

## 오늘의 결론

```text
Day12에서 확인한 결론:

다음 구현으로 넘어가기 전에 남은 질문:
```
