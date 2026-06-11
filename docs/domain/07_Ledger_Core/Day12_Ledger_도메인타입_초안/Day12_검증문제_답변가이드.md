# Day 12 검증문제와 답변가이드

관련 Jira: [SPN-29](https://aslan0.atlassian.net/browse/SPN-29)

먼저 문제를 풀어보고, 그 다음 답변가이드와 비교합니다.

## 먼저 풀어볼 문제

1. Payment와 Ledger는 무엇이 다른가?
2. Ledger Account는 무엇인가?
3. Ledger Transaction은 무엇인가?
4. Ledger Entry는 무엇인가?
5. amount를 `float64`가 아니라 `int64`로 둔 이유는 무엇인가?
6. Day12에서 DB migration이나 repository를 만들지 않는 이유는 무엇인가?
7. Day12 타입 초안은 다음 구현에서 어디로 이어지는가?

## 답변가이드

### 1. Payment와 Ledger의 차이

Payment는 결제의 상태를 관리합니다.

예를 들어 `PENDING`, `ONCHAIN_DETECTED`, `FINALIZED`처럼 결제가 어디까지 진행되었는지 표현합니다.

Ledger는 돈의 이동 기록을 관리합니다.

누가 누구에게 어떤 통화로 얼마를 이동했는지 기록합니다.

### 2. Ledger Account

Account는 돈이 기록되는 주체입니다.

예를 들어 고객 계정, 가맹점 지급 예정 계정, 플랫폼 수수료 계정 같은 것이 있을 수 있습니다.

### 3. Ledger Transaction

Transaction은 여러 Entry를 하나로 묶는 원장 거래 단위입니다.

블록체인 transaction hash와는 다른 개념입니다.

블록체인 transaction은 온체인 거래이고, Ledger Transaction은 우리 내부 원장의 기록 묶음입니다.

### 4. Ledger Entry

Entry는 실제 돈의 이동 한 줄입니다.

예를 들어 고객 계정에서 10 USDC가 빠져나가는 줄, 가맹점 계정에 9.8 USDC가 들어가는 줄, 플랫폼 수수료 계정에 0.2 USDC가 들어가는 줄이 각각 Entry가 될 수 있습니다.

### 5. amount를 int64로 둔 이유

돈을 `float64`로 다루면 소수점 오차가 생길 수 있습니다.

금융 시스템에서는 보통 최소 단위 정수로 저장합니다.

예를 들어 USDC가 소수점 6자리라면 10 USDC는 `10_000_000`으로 저장할 수 있습니다.

### 6. Day12에서 migration/repository를 만들지 않는 이유

Day12의 목적은 Ledger 전체 구현이 아니라 도메인 언어를 코드로 만드는 것입니다.

Account, Transaction, Entry 타입을 먼저 이해해야 이후 DB 테이블과 저장 로직도 더 안전하게 설계할 수 있습니다.

### 7. 다음 구현으로 이어지는 지점

Day12 타입 초안은 다음 작업에서 service test, service validation, migration, repository로 이어집니다.

특히 `EntryDirection`과 `Amount`는 debit/credit 합계 검증 테스트에서 중요하게 사용됩니다.

## Day12 통과 기준

```text
Account, Transaction, Entry의 차이를 설명할 수 있다.
금액을 int64로 두는 이유를 설명할 수 있다.
Day12가 전체 Ledger 구현이 아니라 타입 초안 작성일임을 이해한다.
```
