# Blockchain Event Indexer와 운영 개념 기초 학습

관련 Jira: [SPN-21](https://aslan0.atlassian.net/browse/SPN-21)

이 문서는 Phase 2 Day 4 학습 허브입니다.

Day 4의 목표는 Blockchain Event Indexer가 무엇을 읽고, 어디에서 실행되고, 중복 이벤트와 장애 복구를 어떻게 다뤄야 하는지 이해하는 것입니다.

## 오늘의 큰 그림

![SPN-21 Day 4 학습 흐름](../../confluence/diagrams/spn21-day4-learning-flow.png)

Day 3에서는 Deposit과 Withdrawal의 도메인 흐름을 이해했습니다. Day 4에서는 그중 Deposit 쪽을 실제로 움직이게 만드는 핵심 컴포넌트인 Event Indexer를 학습합니다.

## 오늘의 목표

1. Event Indexer가 블록체인 안이 아니라 우리 백엔드의 off-chain worker layer에서 실행된다는 점을 설명할 수 있다.
2. block, transaction, event/log, RPC의 의미를 구분할 수 있다.
3. checkpoint가 왜 필요한지 설명할 수 있다.
4. idempotency가 중복 이벤트 반영을 어떻게 막는지 설명할 수 있다.
5. reconciliation이 온체인 상태와 내부 DB 상태를 비교하는 작업임을 이해한다.

## 읽기 순서

| 순서 | 문서 | 목적 |
| --- | --- | --- |
| 1 | [Blockchain Event Indexer 개념 학습](Blockchain_Event_Indexer_개념학습.md) | 출퇴근 시간에 읽을 핵심 개념 자료 |
| 2 | [Blockchain Event Indexer 실습 가이드](Blockchain_Event_Indexer_실습가이드.md) | 퇴근 후 직접 작성할 문서 가이드 |
| 3 | [Blockchain Event Indexer 검증문제와 답변가이드](Blockchain_Event_Indexer_검증문제_답변가이드.md) | 학습 후 스스로 확인할 문제와 답변 기준 |

## 오늘 꼭 잡아야 하는 문장

```text
Event Indexer는 블록체인에 이미 기록된 block, transaction, event를 읽어서
우리 내부 DB의 deposit/payment/ledger 상태를 안전하게 바꾸는 off-chain worker다.

Indexer에서 중요한 것은 빨리 읽는 것만이 아니라,
놓치지 않고, 두 번 반영하지 않고, 장애 후에도 다시 맞출 수 있게 만드는 것이다.
```

## 퇴근 후 작업의 원칙

퇴근 후 작업은 사용자가 직접 진행합니다.

1. GitHub repo에서 `docs/domain/04_블록체인_이벤트_인덱서/Blockchain_Event_Indexer_실습산출물.md` 파일을 만든다.
2. Event Indexer가 읽는 데이터와 저장해야 하는 데이터를 정리한다.
3. polling, checkpoint, idempotency, reconciliation의 의미를 본인 말로 작성한다.
4. Payment 상태가 `ONCHAIN_DETECTED` 또는 `FINALIZED`로 바뀌는 흐름을 정리한다.
5. 검증문제를 풀고 답변가이드와 비교한다.

## 완료 기준

- [ ] Event Indexer가 Payment를 `ONCHAIN_DETECTED` 또는 `FINALIZED`로 바꾸는 흐름을 설명할 수 있다.
- [ ] Idempotency와 Reconciliation이 왜 필요한지 예시로 설명한다.
- [ ] Phase 2 Indexer 구현 전 최소 체크리스트를 작성한다.
