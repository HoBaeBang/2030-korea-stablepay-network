# Phase 2 첫 구현 범위 검증문제와 답변가이드

관련 Jira: [SPN-22](https://aslan0.atlassian.net/browse/SPN-22)

이 문서는 Day 5 학습 후 스스로 확인할 검증문제와 답변가이드입니다.

## 검증문제

1. Day 5의 목표는 새로운 도메인을 추가로 많이 배우는 것인가, 구현 범위를 결정하는 것인가?
2. Backend Core를 Ledger나 Indexer보다 먼저 추천하는 이유는 무엇인가?
3. Vertical slice란 무엇인가?
4. Sprint 2 Backend Core에 들어갈 수 있는 백로그 후보 3개를 말해보라.
5. Ledger Core를 바로 시작할 때의 위험은 무엇인가?
6. Indexer Skeleton을 바로 시작할 때의 위험은 무엇인가?
7. Day 6에서 다시 점검해야 할 개념은 어떤 것들이 있는가?

## 답변가이드

### 1. Day 5의 목표

Day 5의 목표는 새로운 도메인을 많이 추가하는 것이 아니라, Day 1~4 학습 내용을 바탕으로 Phase 2 첫 구현 범위를 결정하는 것입니다.

### 2. Backend Core를 먼저 추천하는 이유

Backend Core는 이후 Ledger, Settlement, Indexer, Deposit, Withdrawal이 따라갈 공통 패턴을 만듭니다. 에러 응답, validation, config, logging, 테스트 패턴이 먼저 정리되면 후속 기능이 더 안정적으로 붙습니다.

### 3. Vertical slice

Vertical slice는 하나의 기능을 migration, domain, repository, service, API/test까지 얇게 끝까지 연결하는 방식입니다.

### 4. Sprint 2 백로그 후보

예시는 공통 에러 응답, 요청 validation 정리, 설정 구조 정리, logging 정리, API boundary 정리, 테스트 패턴 정리입니다.

### 5. Ledger Core를 바로 시작할 때의 위험

공통 에러/검증/테스트 구조가 약하면 Ledger 구현이 되더라도 코드가 일관되지 않고, 후속 정산/입출금과 연결할 때 흔들릴 수 있습니다.

### 6. Indexer Skeleton을 바로 시작할 때의 위험

idempotency, checkpoint, finality, payment 상태 전이, ledger 반영 기준이 약하면 이벤트를 읽어도 중복 반영이나 누락 처리를 제대로 설명하기 어렵습니다.

### 7. Day 6 점검 대상

Phase 2 도메인 지도, Ledger/Settlement, Deposit/Withdrawal/Wallet/Key Security, Event Indexer, Idempotency, Reconciliation, Finality, Sprint 2 첫 구현 범위를 다시 확인해야 합니다.
