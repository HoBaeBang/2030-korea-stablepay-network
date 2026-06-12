# Day 14 검증문제와 답변가이드

관련 Jira: [SPN-31](https://aslan0.atlassian.net/browse/SPN-31)

먼저 문제를 풀어보고, 그 다음 답변가이드와 비교합니다.

## 먼저 풀어볼 문제

1. Payment와 Ledger를 분리하는 이유는 무엇인가?
2. Account, Transaction, Entry의 관계를 한 문장으로 설명하면?
3. `ValidateTransaction`에서 debit은 더하고 credit은 빼는 이유는 무엇인가?
4. Day13 테스트가 없었다면 다음 DB 저장 구현에서 어떤 위험이 생길 수 있는가?
5. `ledger_accounts`, `ledger_transactions`, `ledger_entries`는 각각 어떤 타입과 연결되는가?
6. Day14에서 새 기능을 크게 추가하지 않는 이유는 무엇인가?
7. 다음 구현 후보는 무엇인가?
8. Day14를 시작하기 전에 `service.go`, `service_test.go`가 있어야 하는 이유는 무엇인가?

## 답변가이드

### 1. Payment와 Ledger를 분리하는 이유

Payment는 결제 상태를 관리합니다.

Ledger는 돈의 이동 기록을 관리합니다.

두 책임을 섞으면 결제 상태 변경, 돈의 이동 기록, 정산, 장애 복구가 한곳에 엉키기 쉽습니다.

그래서 결제 상태와 원장 기록을 분리합니다.

### 2. Account, Transaction, Entry의 관계

Account는 돈이 기록되는 주체입니다.

Transaction은 여러 Entry를 하나로 묶는 원장 거래입니다.

Entry는 특정 Account에 대해 발생한 돈의 이동 한 줄입니다.

한 문장으로 쓰면:

```text
하나의 Transaction은 여러 Entry를 가지고, 각 Entry는 하나의 Account에 돈의 이동을 기록한다.
```

### 3. debit은 더하고 credit은 빼는 이유

오늘 구현에서는 균형 검증을 단순하게 하기 위해 debit을 plus, credit을 minus로 계산합니다.

최종 합계가 0이면 debit 총액과 credit 총액이 같다는 뜻입니다.

이 방식은 회계 전체를 완벽하게 모델링한 것이 아니라, 첫 Ledger Core 단계에서 균형 검증을 쉽게 구현하기 위한 방법입니다.

### 4. 테스트가 없었다면 생길 위험

DB 저장을 먼저 만들면 불균형 Ledger Transaction도 저장될 수 있습니다.

그렇게 되면 정산 결과가 틀리거나, 장애 복구 때 어떤 기록이 맞는지 판단하기 어려워질 수 있습니다.

테스트는 이런 핵심 규칙을 먼저 고정해줍니다.

### 5. DB 테이블 후보와 타입 연결

```text
ledger_accounts      -> Account
ledger_transactions  -> Transaction
ledger_entries       -> Entry
```

이 연결은 다음 migration 설계의 출발점입니다.

### 6. Day14에서 새 기능을 크게 추가하지 않는 이유

Day12와 Day13에서 새 개념과 새 코드가 이미 추가되었습니다.

이해가 흔들린 상태에서 기능을 계속 쌓으면 코드만 늘고 학습 효과가 떨어질 수 있습니다.

그래서 Day14는 회고와 검증으로 잡습니다.

### 7. 다음 구현 후보

다음 구현 후보는 Ledger DB migration과 repository 초안입니다.

구체적으로는 아래 테이블을 만들 가능성이 큽니다.

```text
ledger_accounts
ledger_transactions
ledger_entries
```

### 8. Day14 전에 `service.go`, `service_test.go`가 있어야 하는 이유

Day14는 Day13에서 만든 Ledger 균형 검증 로직과 테스트를 회고하는 날입니다.

따라서 `service.go`, `service_test.go`가 없다면 아직 회고할 Day13 코드가 없는 상태입니다.

이 경우 Day14 산출물을 억지로 작성하기보다 Day13 실습을 먼저 완료해야 합니다.

## Day14 통과 기준

```text
Payment와 Ledger의 차이를 설명할 수 있다.
Account, Transaction, Entry 관계를 설명할 수 있다.
ValidateTransaction의 균형 검증 흐름을 설명할 수 있다.
다음 DB migration 후보를 말할 수 있다.
Day13 코드가 없으면 Day14를 진행하지 않고 Day13으로 돌아가야 한다는 점을 안다.
```
