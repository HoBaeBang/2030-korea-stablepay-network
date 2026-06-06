# Blockchain Event Indexer

이 문서는 Day 4 퇴근 후 실습에서 직접 채워 넣는 학습 산출물입니다.

관련 가이드: [Blockchain Event Indexer 실습 가이드](Blockchain_Event_Indexer_실습가이드.md)

## 한 문장 요약

Event Indexer는 블록체인에 이미 기록된 block, transaction, event/log를 주기적으로 읽고, 우리 서비스의 deposit, payment, ledger 상태로 안전하게 반영하는 off-chain worker다.

## Event Indexer의 역할

Event Indexer는 블록체인 안에서 실행되는 기능이 아니라, 우리 백엔드 쪽에서 따로 실행되는 worker 프로세스다.

사용자가 "입금했다"고 말하는 것만으로는 내부 DB의 결제 상태나 원장 상태를 바꾸면 안 된다. 실제 블록체인에 transaction이 존재하는지, 그 transaction이 성공했는지, 우리 주소와 관련된 event인지, 충분히 확정됐는지를 확인해야 한다.

Indexer는 이 확인 과정을 자동화한다. 블록체인 RPC를 통해 block과 transaction, event/log를 조회하고, 우리 서비스와 관련된 온체인 이벤트를 찾은 뒤 내부 DB에 반영한다.

즉, Indexer의 핵심 역할은 다음과 같다.

```text
온체인에 실제로 기록된 사실
-> 우리 서비스가 이해할 수 있는 내부 상태
```

예를 들어 USDC 입금 이벤트가 발견되면 payment 상태를 `ONCHAIN_DETECTED`로 바꾸고, finality 기준을 만족하면 `FINALIZED`로 바꾼 뒤 ledger entry를 만들 수 있다.

## Indexer가 읽는 데이터

| 데이터 | 의미 | 왜 필요한가 |
| --- | --- | --- |
| block | 여러 transaction이 묶여 블록체인에 기록되는 단위 | Indexer가 어디까지 읽었는지 block height 기준으로 판단하기 위해 필요 |
| transaction | 블록체인에 제출된 하나의 거래 또는 실행 요청 | tx hash, from/to, status, fee, 실행 결과를 확인하기 위해 필요 |
| event/log | transaction 실행 중 발생한 세부 기록 | ERC-20 `Transfer`처럼 입금/출금과 관련된 실제 이벤트를 찾기 위해 필요 |
| receipt | transaction 실행 결과를 담은 정보 | transaction이 성공했는지, 실패했는지, 어떤 logs가 발생했는지 확인하기 위해 필요 |
| finality | 해당 transaction이 되돌아가기 어렵다고 인정할 수 있는 확정 수준 | 내부 Ledger에 확정 반영해도 되는지 판단하기 위해 필요 |

## Indexer가 저장하거나 변경하는 데이터

| 내부 데이터 | 역할 |
| --- | --- |
| deposits | 온체인에서 감지한 입금 후보와 검증 상태를 저장한다. 예: 감지됨, 확인 중, 확정됨, ledger 반영 완료 |
| payments | 결제가 온체인에서 감지됐는지, 최종 확정됐는지 같은 결제 상태를 변경한다 |
| ledger_entries | 확정된 입금이나 결제로 인해 돈이 어느 계정에 증가/감소했는지 기록한다 |
| indexer_checkpoints | Indexer가 마지막으로 안전하게 처리한 block height를 저장한다 |
| processed_events | 이미 처리한 event를 기록해서 같은 tx/event를 두 번 반영하지 않게 막는다 |

## Polling과 Checkpoint 흐름

Polling은 우리 백엔드의 off-chain worker가 일정 주기로 블록체인 RPC에 질문해서 새 block, transaction, event를 조회하는 방식이다.

처음에는 WebSocket처럼 실시간 이벤트를 받는 방식보다 polling이 더 단순하고 안정적이다. 5초 또는 10초마다 "마지막으로 처리한 block 이후에 새로 생긴 block이 있는가?"를 조회하면 된다.

Checkpoint는 Indexer가 어디까지 처리했는지 저장하는 기준점이다. 예를 들어 `last_processed_height = 1000`이라면 다음 실행에서는 1001번 block부터 읽으면 된다.

Checkpoint가 없으면 장애 후 재시작했을 때 어디서부터 다시 읽어야 할지 모른다. 너무 앞에서 읽으면 중복 처리가 많아지고, 너무 뒤에서 읽으면 이벤트를 놓칠 수 있다.

```text
1. Scheduler가 Indexer를 주기적으로 실행한다.
2. Indexer가 indexer_checkpoints에서 last_processed_height를 조회한다.
3. Blockchain RPC로 다음 block range를 조회한다.
4. 각 block의 transaction receipt와 event/log를 확인한다.
5. 우리 서비스와 관련된 event인지 검증한다.
6. processed_events를 확인해서 이미 처리한 event인지 검사한다.
7. 처음 보는 유효한 event라면 deposits, payments, ledger_entries에 반영한다.
8. 안전하게 처리한 block height를 checkpoint로 저장한다.
```

## Idempotency와 중복 이벤트 처리

Idempotency는 같은 작업을 여러 번 실행해도 결과가 한 번 실행한 것과 같게 유지되는 성질이다.

Indexer는 같은 event를 여러 번 읽을 수 있다. 예를 들어 장애 후 재시작하면서 같은 block range를 다시 읽을 수 있고, reconciliation이나 backfill 과정에서 과거 block을 다시 확인할 수도 있다.

이때 같은 입금 event를 두 번 반영하면 내부 Ledger에 credit이 두 번 생길 수 있다. 온체인 입금은 한 번인데 내부 잔액이 두 번 증가하면 회사가 손실을 볼 수 있다.

그래서 event를 처리할 때는 다음과 같은 idempotency key가 필요하다.

```text
chain + tx_hash + log_index
```

예를 들어 Ethereum에서 하나의 transaction 안에 여러 log가 있을 수 있으므로 `tx_hash`만으로는 부족할 수 있다. `log_index`까지 함께 사용하면 특정 transaction 안의 특정 event를 더 정확하게 식별할 수 있다.

처리 흐름은 다음과 같다.

```text
1. event를 읽는다.
2. chain + tx_hash + log_index를 만든다.
3. processed_events에 이미 존재하는지 확인한다.
4. 이미 있으면 아무 것도 다시 반영하지 않는다.
5. 없으면 deposit/payment/ledger를 반영하고 processed_events에 기록한다.
```

## Reconciliation과 장애 복구

Reconciliation은 온체인 상태와 내부 DB 상태가 서로 맞는지 대조하는 작업이다. 한글로는 보통 `대사`라고 부르지만, 여기서는 "상태 대조 확인"이라고 이해하면 더 쉽다.

Polling과 checkpoint가 있어도 장애나 누락 가능성은 남는다. 예를 들어 checkpoint는 1000번까지 처리됐다고 저장됐지만, 실제로는 998번 block의 event 하나가 DB 반영 전에 누락됐을 수 있다. 또는 RPC 장애, DB 장애, 중간 프로세스 종료 때문에 온체인과 내부 상태가 어긋날 수 있다.

그래서 일정 범위를 다시 조회하면서 다음을 비교해야 한다.

```text
온체인 상태
= 실제 block, transaction, event/log, finality

내부 DB 상태
= deposits, payments, ledger_entries, processed_events, indexer_checkpoints
```

찾아야 하는 문제 예시는 다음과 같다.

| 문제 | 의미 | 조치 방향 |
| --- | --- | --- |
| 온체인에는 입금 event가 있는데 deposits가 없음 | 입금 감지가 누락됨 | deposit 후보를 생성하고 검증 재시작 |
| deposits는 CONFIRMED인데 ledger_entries가 없음 | 입금 확정 후 원장 반영이 누락됨 | ledger entry 생성 여부 재검증 |
| transaction은 실패했는데 payment가 FINALIZED임 | 잘못된 상태 반영 | payment 상태 보정 또는 운영 검토 |
| checkpoint는 앞으로 갔는데 중간 event가 없음 | 처리 범위 일부 누락 가능성 | 해당 block range backfill |

## Payment 상태 변경 흐름

Event Indexer는 payment 상태를 사람이 직접 바꾸는 대신, 온체인 이벤트를 근거로 자동 변경할 수 있다.

```text
PENDING
-> ONCHAIN_DETECTED
-> FINALIZED
```

| 상태 | 의미 | Indexer가 확인해야 하는 것 |
| --- | --- | --- |
| PENDING | 아직 온체인 결제나 입금이 감지되지 않은 상태 | 관련 tx/event가 아직 없음 |
| ONCHAIN_DETECTED | 온체인 transaction/event를 발견했지만 finality를 기다리는 상태 | tx_hash 존재, event/log 존재, 우리 주소와 token이 맞음 |
| FINALIZED | 충분히 확정되어 내부 Ledger 반영이 가능한 상태 | finality 기준 충족, 중복 이벤트 아님, ledger 반영 가능 |

현재 Phase 1에서는 API로 payment 상태를 직접 변경한다. Phase 2에서는 Indexer가 온체인 event를 읽고, 유효한 event라면 상태를 `ONCHAIN_DETECTED`로 바꾸고, 충분히 확정되면 `FINALIZED`로 바꾸는 방향으로 확장한다.

## Phase 2 구현 전 체크리스트

- [x] 지원할 chain과 token을 정해야 한다는 점을 이해했다.
- [x] RPC provider가 필요하다는 점을 이해했다.
- [x] checkpoint 저장 방식이 필요하다는 점을 이해했다.
- [x] idempotency key가 필요하다는 점을 이해했다.
- [x] finality 기준이 필요하다는 점을 이해했다.
- [x] reconciliation 방식이 필요하다는 점을 이해했다.
- [ ] retry 정책은 아직 더 학습이 필요하다.

## 아직 모르는 것과 다음 질문

- Event Indexer를 실제 Go 코드로 구현할 때 worker 프로세스를 API 서버와 같은 프로세스로 둘지, 별도 프로세스로 둘지 궁금하다.
- finality 기준을 체인마다 어떻게 정하는지 더 알고 싶다. 예: Ethereum, Solana, Cosmos 계열의 차이.
- processed_events와 ledger_entries를 같은 DB transaction 안에서 저장해야 하는지 궁금하다.
- Retry 정책에서 "다시 시도해도 되는 실패"와 "운영자가 봐야 하는 실패"를 어떻게 나누는지 궁금하다.

## 검증 체크리스트

- [x] Event Indexer가 무엇인지 설명할 수 있다.
- [x] Checkpoint가 왜 필요한지 설명할 수 있다.
- [x] Idempotency와 Reconciliation이 왜 필요한지 예시로 설명할 수 있다.
- [x] Payment 상태 변경 흐름을 설명할 수 있다.
