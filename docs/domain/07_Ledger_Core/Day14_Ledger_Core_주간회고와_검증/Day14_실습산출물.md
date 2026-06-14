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
Account는 원장에서 돈이 기록되는 계정 또는 주체다.
Transaction은 여러 Entry를 하나로 묶는 원장 거래 단위다.
Entry는 실제 돈의 이동 정보를 한 줄로 기록한 항목이다.
```

Codex 점검:

```text
방향은 맞다.
다만 여기서 Account는 실제 은행 계좌라기보다 StablePay 내부 원장에서 돈을 분류해 기록하기 위한 계정이다.
Transaction은 블록체인 transaction hash와 같은 뜻이 아니라, Ledger 내부에서 여러 Entry를 묶는 단위다.
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
entries 개수가 2보다 작지 않은지 확인한다.
각 Entry의 Amount가 0보다 큰지 확인한다.
Currency가 빈 값이 아닌지 확인한다.
Direction이 DEBIT 또는 CREDIT인지 확인한다.
totals map에 currency별 합계를 저장한다.
DEBIT이면 amount를 더하고 CREDIT이면 amount를 뺀다.
최종 합계가 0인지 확인하고, 0이 아니면 균형이 맞지 않으므로 error를 반환한다.
```

Codex 점검:

```text
기존 답변에서 "Amount가 비어 있지 않은지"라고 표현한 부분은 "Amount가 0보다 큰지"가 더 정확하다.
Currency도 현재 코드에서는 실제 지원 통화 목록까지 검증하지 않고, 빈 문자열인지 여부만 확인한다.
즉 지금 단계의 검증은 "USDC가 맞는 통화인가?"까지가 아니라 "통화 값이 아예 비어 있지 않은가?"까지다.
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
credit과 debit의 합계가 0이 되는지 확인하는 테스트가 가장 중요하다고 생각한다.
Ledger는 돈의 이동 기록이기 때문에 이 균형이 깨지면 서비스 신뢰성에 큰 문제가 생길 수 있다.
특히 credit 합계가 부족하면 실패하는 테스트는 돈이 사라진 것처럼 기록되는 불균형 거래를 막아준다.
```

Codex 점검:

```text
좋은 답변이다.
Ledger에서 가장 중요한 것은 "잘못된 돈의 이동이 통과하지 않는 것"이다.
성공 케이스도 중요하지만, 실제 운영에서는 실패해야 하는 데이터가 통과하지 않는 것이 더 중요하다.
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
ledger_accounts는 Account와 연결된다.
ledger_transactions는 Transaction과 연결된다.
ledger_entries는 Entry와 연결된다.

Entry.TransactionID는 ledger_entries.transaction_id가 될 수 있다.
Entry.AccountID는 ledger_entries.account_id가 될 수 있다.
Entry.Amount는 ledger_entries.amount가 될 수 있다.
Entry.Currency는 ledger_entries.currency가 될 수 있다.
Entry.Direction은 ledger_entries.direction이 될 수 있다.
```

Codex 점검:

```text
기본 연결은 정확하다.
추가로 Entry는 혼자 존재하기보다 어떤 Transaction에 속하고 어떤 Account에 기록되는지가 중요하다.
그래서 다음 단계의 DB 설계에서는 ledger_entries.transaction_id와 ledger_entries.account_id가 중요한 연결점이 된다.
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
ledger_accounts, ledger_transactions, ledger_entries를 만든다고 할 때 이 테이블들을 아우르는 부모 테이블이 따로 필요한지 궁금하다.
```

Codex 점검:

```text
좋은 질문이다.
현재 구조에서는 세 테이블 전체를 감싸는 별도의 parents 테이블이 꼭 필요하지 않다.
ledger_transactions가 여러 ledger_entries를 묶는 부모 역할을 하고, ledger_accounts는 각 entry가 어느 계정에 기록되는지 연결되는 기준 테이블 역할을 한다.
즉 전체 구조는 "최상위 부모 테이블 1개 + 하위 테이블들"이라기보다, 원장 거래와 원장 계정이 entry를 통해 연결되는 구조에 가깝다.
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
internal/ledger/ledger.go
6:type AccountType string
15:type EntryDirection string
23:type Account struct {
32:type Transaction struct {
41:type Entry struct {

internal/ledger/service.go
16:// ValidateTransaction은 원장 거래의 기본 규칙을 검증한다.
17:func (s *Service) ValidateTransaction(ctx context.Context, entries []Entry) error {

internal/ledger/service_test.go
8:func TestServiceValidateTransaction(t *testing.T) {
34:             if err := svc.ValidateTransaction(ctx, entries); err != nil {
55:             if err := svc.ValidateTransaction(ctx, entries); err == nil {
76:             if err := svc.ValidateTransaction(ctx, entries); err == nil {
97:             if err := svc.ValidateTransaction(ctx, entries); err == nil {

go test ./internal/ledger -v 성공
go test ./... 성공
```

## 오늘의 결론

```text
Day14에서 확인한 결론:
Ledger Core는 Account, Transaction, Entry 타입을 바탕으로 돈의 이동을 기록한다.
ValidateTransaction은 저장 전에 Entry 목록이 기본 규칙과 debit/credit 균형을 만족하는지 확인한다.
Day13의 테스트 4개는 정상 균형 거래와 대표적인 실패 케이스를 고정한다.
현재 코드 기준으로 Day14 회고 범위는 잘 반영되어 있으며, 추가 코드 수정은 필요하지 않다.

다음 구현으로 넘어가기 전에 남은 질문:
Ledger 테이블 간의 부모/자식 관계를 DB foreign key 관점에서 더 익숙하게 볼 필요가 있다.
Account와 실제 계좌의 차이, Ledger Transaction과 블록체인 transaction의 차이는 계속 구분해서 복습해야 한다.
```

## 추가 보충 정리

### 1. Day14 기준 코드 수정이 필요한가?

```text
현재 Day14 범위에서는 코드 수정이 필요하지 않다.
internal/ledger/ledger.go에는 Account, Transaction, Entry 타입이 있다.
internal/ledger/service.go에는 ValidateTransaction이 있다.
internal/ledger/service_test.go에는 Day13 기준 테스트 4개가 있다.
go test ./internal/ledger -v와 go test ./... 모두 성공했다.
```

Day15 이후에는 테스트 케이스를 더 늘릴 수 있습니다.

예:

```text
entry가 하나뿐이면 실패한다.
통화가 비어 있으면 실패한다.
context가 취소되었으면 실패한다.
```

하지만 이것은 Day14 회고 범위의 누락이라기보다 Day15 이후 보강 범위입니다.

### 2. 부모 테이블이 따로 필요한가?

```text
지금 구조에서는 모든 Ledger 테이블을 감싸는 별도의 parents 테이블은 필요하지 않다.
```

관계는 이렇게 보는 것이 좋습니다.

```text
ledger_transactions
  └─ ledger_entries

ledger_accounts
  └─ ledger_entries
```

`ledger_entries`는 두 방향으로 연결됩니다.

```text
transaction_id -> ledger_transactions.id
account_id     -> ledger_accounts.id
```

즉 `ledger_transactions`는 여러 Entry를 묶는 부모 역할을 하고, `ledger_accounts`는 각 Entry가 어느 계정에 기록되는지 알려주는 기준 테이블 역할을 합니다.

### 3. 지금 단계에서 꼭 구분해야 하는 것

```text
Payment는 결제 상태를 관리한다.
Ledger는 돈의 이동 기록을 관리한다.
Ledger Transaction은 내부 원장 거래 묶음이다.
Blockchain transaction은 온체인에 기록되는 트랜잭션이다.
Account는 실제 은행 계좌가 아니라 원장 기록을 위한 계정이다.
Entry는 실제 돈의 이동 한 줄이다.
```
