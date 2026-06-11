# Day 11 검증문제와 답변가이드

관련 Jira: [SPN-28](https://aslan0.atlassian.net/browse/SPN-28)

Day11 검증은 많은 문제를 푸는 방식이 아닙니다.

아래 질문에 답할 수 있으면 Day11의 목적은 충분합니다.

## 먼저 풀어볼 문제

1. Day10에서 추가한 payment 테스트는 어떤 버그를 막는가?
2. `ONCHAIN_DETECTED` 상태에 `transaction_hash`가 필요한 이유는 무엇인가?
3. 테스트와 로그의 차이는 무엇인가?
4. Payment와 Ledger의 차이는 무엇인가?
5. Ledger를 단순 CRUD처럼 만들면 위험한 이유는 무엇인가?
6. Ledger 첫 구현을 작게 나누면 어떤 장점이 있는가?

## 답변가이드

### 1. Day10 payment 테스트가 막는 버그

`transaction_hash` 없이 `ONCHAIN_DETECTED` 상태가 저장되는 버그를 막습니다.

이 상태는 블록체인 transaction을 감지했다는 의미이므로, 어떤 transaction을 감지했는지 추적할 수 있어야 합니다.

### 2. transaction_hash가 필요한 이유

`transaction_hash`는 블록체인 transaction을 식별하는 값입니다.

이 값이 없으면 DB의 payment 상태와 실제 온체인 거래를 연결하기 어렵습니다.

### 3. 테스트와 로그의 차이

테스트는 개발 시점에 규칙이 깨지지 않도록 고정하는 장치입니다.

로그는 운영 시점에 이미 발생한 일을 추적하기 위한 기록입니다.

### 4. Payment와 Ledger의 차이

Payment는 결제의 상태를 관리합니다.

Ledger는 돈의 이동 기록을 관리합니다.

Payment만 있으면 결제가 `FINALIZED`인지 알 수 있지만, 누가 누구에게 어떤 통화로 얼마를 이동시켰는지 충분히 설명하기 어렵습니다.

### 5. Ledger를 단순 CRUD처럼 만들면 위험한 이유

Ledger는 돈의 이동 기록이므로 중복 저장, 부분 저장, debit/credit 불균형이 치명적입니다.

단순 CRUD처럼 만들면 같은 payment가 두 번 반영되거나, transaction은 저장되고 entry 일부만 저장되는 문제가 생길 수 있습니다.

### 6. Ledger 첫 구현을 작게 나누는 장점

작게 나누면 어떤 개념에서 막히는지 빨리 알 수 있습니다.

예를 들어 `ledger.go` 타입 정의, `service_test.go` 테스트, `service.go` 검증 로직, migration을 분리하면 각 작업의 목적이 분명해집니다.

## Day11 통과 기준

```text
Day10 payment 테스트를 설명할 수 있다.
Payment와 Ledger의 차이를 설명할 수 있다.
Ledger 첫 구현을 작게 시작해야 하는 이유를 설명할 수 있다.
```
