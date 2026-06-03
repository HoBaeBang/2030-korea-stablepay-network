# Blockchain Event Indexer

이 문서는 Day 4 퇴근 후 실습에서 직접 채워 넣는 학습 산출물입니다.

관련 가이드: [Blockchain Event Indexer 실습 가이드](blockchain-event-indexer-practice-guide.md)

## 한 문장 요약

TODO: Event Indexer를 본인 말로 한 문장으로 정리합니다.

## Event Indexer의 역할

TODO: Event Indexer가 왜 필요한지 작성합니다.

## Indexer가 읽는 데이터

| 데이터 | 의미 | 왜 필요한가 |
| --- | --- | --- |
| block | TODO | TODO |
| transaction | TODO | TODO |
| event/log | TODO | TODO |
| receipt | TODO | TODO |
| finality | TODO | TODO |

## Indexer가 저장하거나 변경하는 데이터

| 내부 데이터 | 역할 |
| --- | --- |
| deposits | TODO |
| payments | TODO |
| ledger_entries | TODO |
| indexer_checkpoints | TODO |
| processed_events | TODO |

## Polling과 Checkpoint 흐름

TODO: polling과 checkpoint 흐름을 작성합니다.

```text
last_processed_height 조회
-> 다음 block range 조회
-> event 파싱
-> DB 반영
-> checkpoint 갱신
```

## Idempotency와 중복 이벤트 처리

TODO: 같은 이벤트를 두 번 읽어도 한 번만 반영해야 하는 이유를 작성합니다.

## Reconciliation과 장애 복구

TODO: 온체인 상태와 내부 DB 상태를 비교해야 하는 이유를 작성합니다.

## Payment 상태 변경 흐름

TODO: `PENDING -> ONCHAIN_DETECTED -> FINALIZED` 흐름을 작성합니다.

## Phase 2 구현 전 체크리스트

- [ ] 지원할 chain과 token을 정한다.
- [ ] RPC provider를 정한다.
- [ ] checkpoint 저장 방식을 정한다.
- [ ] idempotency key를 정한다.
- [ ] finality 기준을 정한다.
- [ ] reconciliation 방식을 정한다.
- [ ] retry 정책을 정한다.

## 아직 모르는 것과 다음 질문

- TODO
- TODO

## 검증 체크리스트

- [ ] Event Indexer가 무엇인지 설명할 수 있다.
- [ ] Checkpoint가 왜 필요한지 설명할 수 있다.
- [ ] Idempotency와 Reconciliation이 왜 필요한지 예시로 설명할 수 있다.
- [ ] Payment 상태 변경 흐름을 설명할 수 있다.
