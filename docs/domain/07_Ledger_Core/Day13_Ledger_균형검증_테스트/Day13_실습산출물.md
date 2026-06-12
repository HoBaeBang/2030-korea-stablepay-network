# Day 13 실습산출물 - Ledger 균형 검증 테스트

관련 Jira: [SPN-30](https://aslan0.atlassian.net/browse/SPN-30)

Day13 산출물은 5개 질문만 작성합니다.

작성 전에 아래 파일을 먼저 확인합니다.

```text
internal/ledger/service.go
internal/ledger/service_test.go
```

오늘 산출물은 “외워서 쓰는 문서”가 아니라, 방금 작성한 코드와 테스트를 읽고 내 말로 정리하는 문서입니다.

## 1. 오늘 만든 Service 메서드는 어떤 규칙을 검증하는가?

작성할 때 볼 파일:

```text
internal/ledger/service.go
```

특히 아래 부분을 봅니다.

```go
func (s *Service) ValidateTransaction(ctx context.Context, entries []Entry) error
```

작성 예시:

```text
오늘 만든 ValidateTransaction은 하나의 Ledger Transaction 안에서 debit과 credit 합계가 같은지 검증한다.
합계가 맞지 않으면 원장 거래가 불균형하므로 error를 반환한다.
```

내 답변:

```text

```

## 2. debit과 credit 합계가 같아야 하는 이유는 무엇인가?

작성할 때 떠올릴 예시:

```text
고객 DEBIT 10 USDC
가맹점 CREDIT 9.8 USDC
플랫폼 수수료 CREDIT 0.2 USDC
```

이 질문은 회계 이론을 완벽히 설명하라는 뜻이 아닙니다.

오늘 코드 기준으로 “왜 합계가 0이어야 하는지”를 설명하면 됩니다.

작성 예시:

```text
Ledger는 돈의 이동 기록이기 때문에 돈이 갑자기 생기거나 사라지면 안 된다.
그래서 하나의 거래 안에서 debit 총액과 credit 총액이 같아야 한다.
```

내 답변:

```text

```

## 3. `map[string]int64`는 어떤 역할을 하는가?

작성할 때 볼 코드:

```go
totals := make(map[string]int64)
```

작성 예시:

```text
map의 key는 currency이고 value는 해당 통화의 합계다.
USDC debit은 더하고 USDC credit은 빼서 최종 값이 0인지 확인한다.
```

내 답변:

```text

```

## 4. 오늘 테스트 4개는 각각 어떤 버그를 막는가?

작성할 때 볼 파일:

```text
internal/ledger/service_test.go
```

테스트 이름을 먼저 읽고, 그 테스트가 실패해야 하는 상황인지 성공해야 하는 상황인지 구분합니다.

작성 예시:

```text
정상 균형 거래 테스트는 정상 케이스가 통과하는지 확인한다.
credit 부족 테스트는 불균형 거래가 저장되는 문제를 막는다.
0원 테스트는 의미 없는 Entry가 들어오는 문제를 막는다.
알 수 없는 direction 테스트는 잘못된 문자열이 원장에 들어오는 문제를 막는다.
```

내 답변:

```text

```

## 5. 아직 헷갈리는 Go 문법 또는 Ledger 개념은 무엇인가?

작성할 때 아래 후보 중에서 실제로 헷갈린 것을 골라도 됩니다.

```text
receiver: func (s *Service) ...
포인터: *Service, &Service{}
짧은 변수 선언: :=
slice: []Entry{...}
map: map[string]int64
에러 처리: if err := ...; err != nil
context.Context
debit / credit 방향
```

작성 예시:

```text
map에 값을 더하고 빼는 부분은 이해했지만, debit과 credit 방향은 아직 예시를 더 봐야 할 것 같다.
또한 `if err := ...; err != nil` 문법을 반복해서 봐야 할 것 같다.
```

내 답변:

```text

```

## 실행 결과

실행한 명령:

```bash
gofmt -w internal/ledger/service.go internal/ledger/service_test.go
go test ./internal/ledger -v
go test ./...
```

결과:

```text

```

## 오늘의 결론

```text
Day13에서 확인한 결론:

다음 구현으로 넘어가기 전에 남은 질문:
```
