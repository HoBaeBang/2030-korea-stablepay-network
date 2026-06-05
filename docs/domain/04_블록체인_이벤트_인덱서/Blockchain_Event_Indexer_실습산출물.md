# Blockchain Event Indexer

이 문서는 Day 4 퇴근 후 실습에서 직접 채워 넣는 학습 산출물입니다.

관련 가이드: [Blockchain Event Indexer 실습 가이드](Blockchain_Event_Indexer_실습가이드.md)

## 한 문장 요약

TODO: Event Indexer를 본인 말로 한 문장으로 정리합니다.

## Event Indexer의 역할

TODO: Event Indexer가 왜 필요한지 작성합니다.
블록체인에서 이벤트 인덱서는 블록체인에 외부에서 내부로 인입되는 트렌젝션에 대해서 내용을 확인해서 조회합니다.

## Indexer가 읽는 데이터

| 데이터 | 의미                         | 왜 필요한가                        |
| --- |----------------------------|-------------------------------|
| block | 거래 정보가 저장된 묶음 단위?          | 저장을 하는 단위로 블록체인으로 저장하기 위해서 필요 |
| transaction | 거래 정보에 대한 일련의 과정을 기록       | 거래 정보에 대해서 블록에 기록하기 위해서 필요    |
| event/log | 거래 정보에 대해서 입금, 출금에 대한 이벤트? |                               |
| receipt | transaction 실행 결과 | 성공, 실패 확인을 위해서                |
| finality | 해당 거래 내용이 얼마나 믿을수 있는지를 확인  | 해당 거래가 신뢰 할 수 있는지 확인하기 위함     |

## Indexer가 저장하거나 변경하는 데이터

| 내부 데이터 | 역할                               |
| --- |----------------------------------|
| deposits | 입금 내용을 확인함                       |
| payments | 결제 상태 확인을 위함                     |
| ledger_entries | 모르겠음                                 |
| indexer_checkpoints | 마지막 반영한 거래거 어디인지 확인 해서 중복을 막기 위함 |
| processed_events | 모르겠음                             |

## Polling과 Checkpoint 흐름

TODO: polling과 checkpoint 흐름을 작성합니다.
polling은 off chain 환경의 Go 백엔드 서버가 on chain 환경에다가 지속적으로 내용을 확인해서 이벤트를 확인하는 방식으로 진행하기 위해
정보를 주기적으로 받아오는 것을 의미하고 checkpoint는 마지막 확인했던 정보를 기록함으로써 마지막으로 기록된 부분을 확인

```text
last_processed_height 조회
-> 다음 block range 조회
-> event 파싱
-> DB 반영
-> checkpoint 갱신
```

## Idempotency와 중복 이벤트 처리

TODO: 같은 이벤트를 두 번 읽어도 한 번만 반영해야 하는 이유를 작성합니다.
같은 이벤트를 두번 반영하게되면 사용자의 계좌에서 출금이 2번 일어나거나 입금이 2번 일어나는 등의 문제가 발생하기 때문에 한번만 반영해야 합니다.

## Reconciliation과 장애 복구

TODO: 온체인 상태와 내부 DB 상태를 비교해야 하는 이유를 작성합니다.
온체인 상태와 내부 상태가 동일하다면 이미 정상 반영된것이고, 만약 다르다면 온체인 상테에 맞게 내용을 확인하고 반영해야 하기 때문이다.

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
