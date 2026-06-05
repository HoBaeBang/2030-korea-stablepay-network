# Blockchain Event Indexer 개념 학습

관련 Jira: [SPN-21](https://aslan0.atlassian.net/browse/SPN-21)

이 문서는 출퇴근 시간에 읽는 Day 4 개념 학습자료입니다.

## 1. Event Indexer란 무엇인가

`Event Indexer`는 블록체인에 이미 기록된 block, transaction, event/log를 읽고, 우리 서비스의 DB 상태로 변환하는 off-chain worker입니다.

![Blockchain Event Indexer 실행 레이어](../../confluence/diagrams/spn21-indexer-runtime-layer.png)

중요한 점은 Event Indexer가 블록체인 네트워크 내부에서 실행되는 것이 아니라는 점입니다. Event Indexer는 우리 백엔드에 속한 별도 worker 프로세스입니다.

초기 구현에서는 일정 주기로 블록체인 RPC를 조회하는 polling 방식이 가장 현실적입니다. 이후에는 WebSocket subscription, backfill, reconciliation을 섞은 hybrid 방식으로 확장할 수 있습니다.

## 2. Indexer가 읽는 데이터

| 데이터 | 의미 | 우리 프로젝트에서 보는 이유 |
| --- | --- | --- |
| Block | transaction들이 묶여 기록되는 단위 | 어떤 높이까지 처리했는지 판단하기 위해 필요 |
| Transaction | 블록체인에 제출된 거래 | tx hash, status, from/to, fee, 실행 결과를 확인하기 위해 필요 |
| Event / Log | transaction 실행 중 발생한 세부 이벤트 | ERC-20 `Transfer` 같은 입금 이벤트를 찾기 위해 필요 |
| Receipt | transaction 실행 결과 | 성공/실패, gas, logs를 확인하기 위해 필요 |
| Finality | 되돌릴 수 없다고 인정할 수 있는 상태 | 내부 Ledger에 확정 반영해도 되는지 판단하기 위해 필요 |

## 3. Polling 방식의 실행 흐름

`Polling`은 일정 주기로 블록체인 RPC를 조회하는 방식입니다.

![Indexer Polling Sequence](../../confluence/diagrams/spn21-indexer-polling-sequence.png)

예를 들면 다음처럼 동작합니다.

```text
1. Scheduler가 5초마다 Indexer를 실행한다.
2. Indexer가 마지막 처리 block height를 checkpoint DB에서 읽는다.
3. Blockchain RPC로 다음 block range를 조회한다.
4. 각 block의 transaction/event를 파싱한다.
5. 우리 deposit address와 관련된 event인지 검증한다.
6. 중복 이벤트가 아니면 deposit/payment/ledger 상태를 반영한다.
7. 안전하게 처리한 block height를 checkpoint로 저장한다.
```

## 4. Checkpoint가 왜 필요한가

`Checkpoint`는 Indexer가 어디까지 처리했는지 기록하는 기준점입니다.

예를 들어 `last_processed_height = 1000`이라면, Indexer는 다음 실행 때 1001번 블록부터 읽을 수 있습니다.

Checkpoint가 없으면 장애 후에 어디서부터 다시 읽어야 하는지 알 수 없습니다. 너무 앞에서 다시 읽으면 중복 처리가 늘어나고, 너무 뒤에서 시작하면 이벤트를 놓칠 수 있습니다.

## 5. Idempotency가 왜 필요한가

`Idempotency`는 같은 작업을 여러 번 실행해도 결과가 한 번 실행한 것과 같게 만드는 성질입니다.

![Idempotency and Reconciliation](../../confluence/diagrams/spn21-idempotency-reconciliation.png)

Indexer에서는 같은 event를 여러 번 읽을 수 있습니다.

예를 들어 장애가 나서 같은 block range를 다시 읽거나, RPC 응답이 중복되거나, backfill 작업이 과거 블록을 다시 읽을 수 있습니다.

그래서 다음 같은 idempotency key가 필요합니다.

```text
chain + tx_hash + log_index
```

이 키가 이미 처리된 이벤트라면 Ledger credit을 다시 만들면 안 됩니다.

## 6. Reconciliation이 왜 필요한가

`Reconciliation`은 한글로 보통 `대사`라고 부릅니다. 여기서 대사는 대화의 대사가 아니라, 서로 다른 장부나 시스템의 상태가 맞는지 대조하는 일을 의미합니다.

우리 프로젝트에서는 다음을 비교합니다.

```text
온체인 상태
= 실제 블록체인에 기록된 tx/event/finality

내부 DB 상태
= deposits, payments, ledger_entries, checkpoints
```

둘이 어긋나면 다음 문제가 있을 수 있습니다.

- 온체인에는 입금이 있는데 내부 DB에는 deposit이 없다.
- 내부 DB에는 deposit이 있는데 ledger credit이 없다.
- transaction은 실패했는데 payment가 finalized로 되어 있다.
- checkpoint는 앞으로 갔는데 중간 block event가 누락됐다.

## 7. Payment 상태 변경 흐름

Event Indexer는 Payment 상태를 자동으로 바꾸는 역할도 할 수 있습니다.

```text
PENDING
-> ONCHAIN_DETECTED
-> FINALIZED
```

| 상태 | 의미 |
| --- | --- |
| PENDING | 아직 온체인 입금이 감지되지 않은 상태 |
| ONCHAIN_DETECTED | 온체인 transaction/event를 발견했지만 finality 기준을 기다리는 상태 |
| FINALIZED | 충분히 확정되어 내부 Ledger 반영까지 가능한 상태 |

## 8. 오늘 기억할 요약

```text
Event Indexer는 블록체인 이벤트를 읽는 off-chain worker다.
Checkpoint는 어디까지 읽었는지 기록한다.
Idempotency는 같은 이벤트를 두 번 반영하지 않게 만든다.
Reconciliation은 온체인 상태와 내부 DB 상태가 맞는지 확인한다.
Phase 2에서는 이 개념들이 Deposit 자동화와 장애 복구의 핵심이 된다.
```
