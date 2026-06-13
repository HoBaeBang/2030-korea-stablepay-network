# Day 14 실습산출물 - Ledger Core 주간 회고와 검증

관련 Jira: [SPN-31](https://aslan0.atlassian.net/browse/SPN-31)

Day14 산출물은 5개 질문만 작성합니다.

작성 전에 먼저 확인합니다.

```bash
ls internal/ledger
```

아래 파일이 없다면 Day14 산출물 작성 전에 Day13 실습을 먼저 끝냅니다.

```text
service.go
service_test.go
```

오늘 산출물은 “새로운 지식을 많이 쓰는 문서”가 아니라, Day12~13 코드가 내 머릿속에서 연결되었는지 확인하는 문서입니다.

## 1. Day12에서 만든 타입 3개는 각각 무엇인가?

작성할 때 볼 파일:

```text
internal/ledger/ledger.go
```

특히 아래 타입을 확인합니다.

```text
Account
Transaction
Entry
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

아래 키워드를 순서대로 포함해보면 좋습니다.

```text
entries 개수
Amount
Currency
Direction
totals map
최종 합계 0
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

정답이 하나로 고정된 질문은 아닙니다.

중요하다고 생각한 이유를 코드와 연결해서 설명하면 됩니다.

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

추가로 여유가 있다면 아래처럼 필드도 한 줄씩 연결해봅니다.

```text
Entry.TransactionID는 ledger_entries.transaction_id가 될 수 있다.
Entry.AccountID는 ledger_entries.account_id가 될 수 있다.
Entry.Amount는 ledger_entries.amount가 될 수 있다.
```

내 답변:

```text

```

## 5. 다음 구현으로 넘어가기 전에 아직 헷갈리는 부분은 무엇인가?

아래 후보 중 실제로 헷갈리는 것을 골라도 됩니다.

```text
Payment와 Ledger의 차이
Account와 실제 계좌의 차이
Ledger Transaction과 블록체인 transaction의 차이
debit / credit 방향
map으로 합계를 계산하는 방식
DB 테이블로 나누는 기준
```

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
grep -RnE "type Account|type Transaction|type Entry|ValidateTransaction" internal/ledger
go test ./internal/ledger -v
go test ./...
```

`rg`가 설치되어 있다면 아래 명령으로 바꿔 사용할 수 있습니다.

```bash
rg -n "type Account|type Transaction|type Entry|ValidateTransaction" internal/ledger
```

결과:

```text

```

## 오늘의 결론

```text
Day14에서 확인한 결론:

다음 구현으로 넘어가기 전에 남은 질문:
```
