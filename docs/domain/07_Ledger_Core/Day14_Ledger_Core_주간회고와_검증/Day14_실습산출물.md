# Day 14 실습산출물 - Ledger Core 주간 회고와 검증

관련 Jira: [SPN-31](https://aslan0.atlassian.net/browse/SPN-31)

Day14 산출물은 5개 질문만 작성합니다.

## 1. Day12에서 만든 타입 3개는 각각 무엇인가?

작성할 때 볼 파일:

```text
internal/ledger/ledger.go
```

작성 예시:

```text
Account는 돈이 기록되는 주체다.
Transaction은 여러 Entry를 하나로 묶는 원장 거래다.
Entry는 실제 돈의 이동 한 줄이다.
```

내 답변:

```text

```

## 2. Day13에서 만든 ValidateTransaction은 어떤 흐름으로 동작하는가?

작성할 때 볼 파일:

```text
internal/ledger/service.go
```

작성 예시:

```text
entries 개수를 확인하고, amount와 currency를 검증한다.
그 다음 direction에 따라 통화별 합계를 계산하고, 최종 합계가 0인지 확인한다.
합계가 0이 아니면 균형이 맞지 않으므로 error를 반환한다.
```

내 답변:

```text

```

## 3. Day13 테스트 4개 중 가장 중요한 테스트는 무엇이라고 생각하는가?

작성할 때 볼 파일:

```text
internal/ledger/service_test.go
```

작성 예시:

```text
나는 credit 합계가 부족하면 실패하는 테스트가 가장 중요하다고 생각한다.
왜냐하면 이 테스트가 돈이 사라진 것처럼 기록되는 불균형 원장 거래를 막아주기 때문이다.
```

내 답변:

```text

```

## 4. 다음 DB 테이블 후보 3개는 각각 어떤 타입과 연결되는가?

작성 예시:

```text
ledger_accounts는 Account와 연결된다.
ledger_transactions는 Transaction과 연결된다.
ledger_entries는 Entry와 연결된다.
```

내 답변:

```text

```

## 5. 다음 구현으로 넘어가기 전에 아직 헷갈리는 부분은 무엇인가?

작성 예시:

```text
타입과 테스트 흐름은 이해했지만, 이것이 DB 테이블로 나뉘면 어떤 컬럼이 필요한지는 아직 더 학습이 필요하다.
또한 debit과 credit 방향은 결제 예시를 더 보면서 익숙해져야 할 것 같다.
```

내 답변:

```text

```

## 실행 결과

실행한 명령:

```bash
ls internal/ledger
rg -n "type Account|type Transaction|type Entry|ValidateTransaction" internal/ledger
go test ./internal/ledger -v
go test ./...
```

결과:

```text

```

## 오늘의 결론

```text
Day14에서 확인한 결론:

다음 구현으로 넘어가기 전에 남은 질문:
```
