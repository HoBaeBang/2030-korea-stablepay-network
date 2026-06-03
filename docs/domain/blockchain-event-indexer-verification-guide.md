# Blockchain Event Indexer 검증문제와 답변가이드

관련 Jira: [SPN-21](https://aslan0.atlassian.net/browse/SPN-21)

이 문서는 Day 4 학습 후 스스로 확인할 검증문제와 답변가이드입니다.

## 검증문제

1. Event Indexer는 블록체인 안에서 실행되는가, 우리 백엔드에서 실행되는가?
2. Polling 방식은 무엇인가?
3. Checkpoint는 왜 필요한가?
4. `chain + tx_hash + log_index`가 idempotency key 후보가 되는 이유는 무엇인가?
5. 같은 event를 두 번 읽었을 때 Ledger credit을 두 번 만들면 어떤 문제가 생기는가?
6. Reconciliation은 무엇을 비교하는 작업인가?
7. Payment가 `PENDING -> ONCHAIN_DETECTED -> FINALIZED`로 바뀌는 흐름을 설명해보라.
8. Indexer가 장애로 중간에 멈췄다가 다시 시작할 때 어떤 정보가 필요할까?
9. WebSocket subscription만 믿으면 왜 위험할 수 있을까?
10. Phase 2 Indexer 구현 전 최소 체크리스트에는 어떤 항목이 들어가야 할까?

## 답변가이드

### 1. Event Indexer는 어디에서 실행되는가

Event Indexer는 블록체인 안에서 실행되는 것이 아니라, 우리 백엔드의 off-chain worker/indexer layer에서 실행되는 별도 프로세스입니다.

### 2. Polling 방식

Polling은 일정 주기로 블록체인 RPC를 조회해서 새 block, transaction, event를 읽는 방식입니다.

### 3. Checkpoint가 필요한 이유

Checkpoint는 Indexer가 어디까지 처리했는지 저장하는 기준점입니다. 장애 후 재시작할 때 어디서부터 다시 읽어야 하는지 판단하기 위해 필요합니다.

### 4. Idempotency key 후보

같은 chain 안에서 하나의 transaction hash와 log index 조합은 특정 event를 식별할 수 있습니다. 그래서 `chain + tx_hash + log_index`를 기준으로 이미 처리한 event인지 확인할 수 있습니다.

### 5. 중복 Ledger credit 문제

온체인 입금은 한 번인데 내부 Ledger에 credit을 두 번 만들면 사용자 잔액이 실제보다 커집니다. 이는 정산 손실과 회계 불일치로 이어질 수 있습니다.

### 6. Reconciliation이 비교하는 것

Reconciliation은 온체인 상태와 내부 DB 상태를 비교합니다. 예를 들어 블록체인의 tx/event/finality와 내부 deposits, payments, ledger_entries, checkpoints가 맞는지 확인합니다.

### 7. Payment 상태 변경 흐름

`PENDING`은 아직 온체인 이벤트가 감지되지 않은 상태입니다. `ONCHAIN_DETECTED`는 transaction/event는 발견했지만 finality를 기다리는 상태입니다. `FINALIZED`는 finality 기준을 만족해서 내부 Ledger 반영까지 가능한 상태입니다.

### 8. 장애 후 재시작에 필요한 정보

마지막 처리 block height, 처리된 event key, 처리 중이던 block range, 실패한 event의 retry 상태가 필요합니다.

### 9. WebSocket만 믿으면 위험한 이유

연결이 끊기거나 이벤트를 놓칠 수 있습니다. 그래서 WebSocket을 쓰더라도 주기적 backfill이나 reconciliation이 필요합니다.

### 10. 최소 체크리스트 예시

- 지원할 chain과 token을 정했는가?
- RPC provider와 장애 대응 전략을 정했는가?
- checkpoint 저장 방식을 정했는가?
- idempotency key를 정했는가?
- finality 기준을 정했는가?
- 누락 이벤트를 찾는 reconciliation 방식을 정했는가?
- 실패 이벤트 retry 정책을 정했는가?
