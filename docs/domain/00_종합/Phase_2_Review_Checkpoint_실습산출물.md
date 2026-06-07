# Phase 2 Review Checkpoint

관련 Jira: [SPN-23](https://aslan0.atlassian.net/browse/SPN-23)

이 문서는 Day 6 퇴근 후 직접 작성하는 통합 복습 산출물입니다.

## 한 문장 요약

> 여기에 Day 1~5를 통합해서 이해한 내용을 한 문장으로 작성한다.

## 종합 교재 핵심 질문 답변

### 1. Phase 1에서 만든 결제 백엔드는 무엇이 부족했는가?

Phase 1에서 만든 결제 백엔드는 `Merchant`, `Invoice`, `Payment`를 중심으로 결제 요청과 결제 상태를 관리할 수 있는 MVP였다.

하지만 실제 블록체인 결제 서비스라고 보기에는 다음 기능들이 부족했다.

- 블록체인에서 발생한 transaction이나 event를 읽는 기능
- 온체인 입금을 감지해서 payment 상태를 자동으로 바꾸는 기능
- 돈이 어느 계정에서 어느 계정으로 이동했는지 기록하는 ledger
- 가맹점에게 지급 가능한 금액을 계산하는 settlement
- deposit, withdrawal, wallet, private key를 안전하게 다루는 경계

즉 Phase 1은 결제 백엔드의 기본 흐름을 만든 단계이고, Phase 2는 이 흐름을 블록체인 금융 백엔드에 가깝게 확장하는 단계다.

### 2. Phase 2에서 Ledger와 Settlement가 왜 필요한가?

`Ledger`는 돈의 이동을 기록하기 위해 필요하다.

Payment는 결제가 `PENDING`, `FINALIZED`, `SETTLED` 중 어디까지 진행됐는지 보여주는 상태 정보에 가깝다. 하지만 Payment만으로는 돈이 왜 이동했는지, 어느 계정에서 어느 계정으로 이동했는지, 중복 반영이 발생하지 않았는지 설명하기 어렵다.

그래서 Ledger에는 돈의 증가와 감소를 짝으로 기록하는 entry가 필요하다. 예를 들어 고객이 10 USDC를 결제했다면 고객 계정에는 `-10 USDC`, 가맹점 pending 계정에는 `+10 USDC`가 기록될 수 있다.

`Settlement`는 가맹점에게 실제로 지급 가능한 금액을 계산하고 묶기 위해 필요하다.

결제가 확정됐다고 해서 무조건 바로 지급할 수 있는 것은 아니다. finality가 충분한지, 중복 ledger entry가 없는지, 실패/환불 상태와 충돌하지 않는지, 수수료와 정산 정책을 반영했는지 확인해야 한다. 이 과정을 통해 지급 가능한 묶음을 만드는 것이 settlement다.

### 3. Deposit과 Withdrawal은 왜 일반 CRUD처럼 만들면 안 되는가?

Deposit과 Withdrawal은 단순히 DB row를 생성, 조회, 수정, 삭제하는 CRUD가 아니다.

`Deposit`은 이미 온체인에서 발생한 입금을 감지하고, 그 입금이 실제로 유효한지 확인한 뒤 내부 상태와 Ledger에 안전하게 반영하는 흐름이다.

확인해야 할 것들은 다음과 같다.

- transaction hash가 실제로 존재하는가?
- 받는 주소가 우리 시스템의 주소인가?
- 토큰 종류와 금액이 맞는가?
- transaction이 성공했는가?
- 충분한 confirmation 또는 finality가 확보됐는가?
- 이미 처리한 transaction은 아닌가?

`Withdrawal`은 우리 시스템이 외부 주소로 자산을 내보내는 흐름이다.

그래서 주소 검증, 출금 승인, 잔액 확인, 중복 출금 방지, transaction signing, broadcast, 온체인 확정 확인이 필요하다. 단순히 `withdrawal` row를 만들었다고 출금이 끝나는 것이 아니다.

### 4. Wallet과 Key Security는 왜 별도 경계로 분리해야 하는가?

Wallet과 Key Security는 자산을 실제로 움직일 수 있는 민감한 영역이기 때문에 일반 백엔드 로직과 분리해야 한다.

특히 private key는 노출되면 자산 탈취로 이어질 수 있다. 그래서 wallet 주소 관리, private key 보관, transaction signing은 일반 API/service 로직에서 마음대로 접근하지 못하도록 별도 경계로 두는 것이 좋다.

분리했을 때 얻는 이점은 다음과 같다.

- private key 접근 범위를 최소화할 수 있다.
- 실수로 일반 API에서 signing 기능을 호출할 가능성을 줄일 수 있다.
- 보안 정책, 감사 로그, 권한 제어를 별도로 적용하기 쉽다.
- 나중에 Rust signer, HSM, KMS 같은 보안 모듈로 확장하기 좋다.

### 5. Event Indexer는 어디에서 실행되고 무엇을 안전하게 처리해야 하는가?

Event Indexer는 블록체인 내부에서 실행되는 기능이 아니라, 우리 백엔드의 off-chain worker layer에서 실행된다.

보통은 worker가 RPC를 통해 블록체인 노드에 접근하고, block, transaction, event를 조회한다. 구현 방식은 polling으로 시작할 수 있고, 나중에 websocket이나 message queue 기반 구조로 확장할 수 있다.

Event Indexer가 안전하게 처리해야 하는 것은 다음과 같다.

- 어디 블록까지 읽었는지 checkpoint로 저장하기
- 같은 event를 여러 번 읽어도 한 번만 반영되도록 idempotency 보장하기
- transaction hash, log index, chain 정보를 기준으로 event 중복 처리 방지하기
- finality가 충분하지 않은 event를 너무 빨리 확정 처리하지 않기
- 장애 후 재시작해도 누락 없이 다시 이어서 처리하기

### 6. 왜 Sprint 2에서는 화려한 블록체인 연결보다 Backend Core를 먼저 정리하려 하는가?

Sprint 2에서 먼저 해야 할 일은 실제 체인 연결보다 Backend Core의 정책과 데이터 구조를 작게 검증하는 것이다.

Ledger, Settlement, Payment finalized 흐름이 정리되지 않은 상태에서 블록체인 연결부터 붙이면, 온체인 event를 읽어도 내부에서 어떻게 기록하고 정산할지 결정하기 어렵다.

그래서 먼저 Backend Core vertical slice를 구현한다.

```text
Payment FINALIZED
        |
        v
Ledger Transaction / Entry 생성
        |
        v
Settlement 후보 계산
```

이 흐름이 작게라도 동작하면, 이후 Event Indexer, Deposit, Withdrawal, Wallet Security를 붙였을 때 어디에 연결해야 하는지 명확해진다.

## Day 1~5 핵심 요약

| Day | 주제 | 내가 이해한 한 문장 |
| --- | --- | --- |
| Day 1 | Phase 2 Domain Map | 기존 결제 백엔드 시스템에서 블록체인 백엔드 결제 시스템으로의 확장 |
| Day 2 | Ledger & Settlement | 원장과 정산 왜 블록체인의 결제 상태만으로 관리가 불가능한가? |
| Day 3 | Deposit / Withdrawal / Wallet / Key Security | 입금, 출금, 지갑, 비밀키의 관계 입금,출금 시 벌어지게 되는 flow 및 발생하는 중복 에러 등에 대한 위험에따라 설계상 key security를 확보 할 수 있도록 레이어 분리 및 각 트랜젝션에 대해 멱등성 유지 |
| Day 4 | Blockchain Event Indexer | 블록체인 네트워크에서 트렌젝션을 조회하여 입금되는 이벤트를 감지하고 그에 대한 작업을 진행하는데 우선적으로는 polling방식으로 진행한다. |
| Day 5 | First Implementation Scope | 첫번째 구현 범위는 백엔드 코어를 기준으로 삼는다. |

## 도메인별 이해 상태

| 도메인 | 이해도(상/중/하) | 이유 |
| --- | --- | --- |
| Payment |  |  |
| Ledger |  |  |
| Settlement |  |  |
| Deposit |  |  |
| Withdrawal |  |  |
| Wallet |  |  |
| Key Security |  |  |
| Event Indexer |  |  |
| Idempotency |  |  |
| Reconciliation |  |  |
| Finality |  |  |

## 아직 약한 개념과 질문

### 약한 개념 1

- 현재 이해:
- 헷갈리는 지점:
- 구현 전에 확인할 질문:

### 약한 개념 2

- 현재 이해:
- 헷갈리는 지점:
- 구현 전에 확인할 질문:

### 약한 개념 3

- 현재 이해:
- 헷갈리는 지점:
- 구현 전에 확인할 질문:

## Sprint 2 구현 전 체크리스트

| 체크 항목 | 결과 | 메모 |
| --- | --- | --- |
| Payment와 Ledger의 책임 차이를 설명할 수 있다 | 가능 | Payment는 결제 진행 상태를 요약하고, Ledger는 돈이 왜 누구에게서 누구에게 얼마 이동했는지 기록한다. |
| Ledger entry에 `+`, `-` 금액이 왜 필요한지 설명할 수 있다 | 가능 | 하나의 거래 사건에서 감소 entry와 증가 entry를 함께 남겨 돈이 임의로 생기거나 사라지지 않게 한다. |
| Settlement가 단순 합계가 아닌 이유를 설명할 수 있다 | 가능 | finality, 중복 ledger entry, 실패/환불 상태, 수수료, 정산 정책을 확인한 뒤 지급 가능한 묶음을 만들어야 한다. |
| Deposit과 Withdrawal의 위험 차이를 설명할 수 있다 | 가능 | Deposit은 이미 발생한 온체인 입금을 감지하고 중복 반영을 막는 것이 중요하고, Withdrawal은 주소 검증, 승인, 서명, broadcast, 중복 출금 방지가 중요하다. |
| Event Indexer가 off-chain worker라는 점을 설명할 수 있다 | 가능 | Indexer는 블록체인 내부 코드가 아니라 우리 백엔드 worker가 RPC로 block, transaction, event/log를 읽는 구조다. |
| Idempotency key 후보를 말할 수 있다 | 가능 | `chain + tx_hash + log_index`로 특정 체인의 특정 transaction 안에 있는 특정 event/log를 식별할 수 있다. |
| Reconciliation이 무엇을 비교하는지 설명할 수 있다 | 가능 | 온체인 상태와 내부 DB 상태를 비교해 누락, 중복, 실패 tx의 잘못된 반영, ledger 불일치를 찾는다. |
| Backend Core vertical slice를 먼저 구현해야 하는 이유를 설명할 수 있다 | 가능 | 실제 체인 연결보다 먼저 Payment, Ledger, Settlement의 최소 연결과 공통 에러/검증/테스트 구조를 작게 검증해야 이후 기능이 흔들리지 않는다. |

## Sprint 2 진입 판단

선택:

- [ ] 구현 시작 가능
- [ ] 부분 보강 후 시작
- [ ] 하루 더 복습 필요

판단 이유:

```text
여기에 판단 이유를 작성한다.
```

## 다음 작업 후보

| 후보 작업 | 해야 하는 이유 | 예상 난이도 | 먼저 확인할 질문 |
| --- | --- | --- | --- |
| Backend Core migration 설계 | Phase 2 기능이 붙기 전에 공통 설정, 에러 응답, validation, logging, 테스트 패턴을 정리해야 한다. | 중 | 현재 `merchant`, `invoice`, `payment` 패키지에서 공통화할 수 있는 에러/응답/검증 패턴은 무엇인가? |
| Ledger account/entry 모델 설계 | Payment 상태만으로는 돈의 이동을 설명할 수 없으므로, 계정과 entry 구조를 먼저 잡아야 한다. | 중상 | `ledger_accounts`, `ledger_transactions`, `ledger_entries`를 어떤 관계로 둘 것인가? |
| Payment finalized 이후 ledger 연결 | `FINALIZED` 상태가 된 결제를 Ledger transaction과 entry로 연결해야 Phase 2 핵심 흐름이 시작된다. | 중상 | 어떤 Payment 상태 전이에서 ledger entry를 생성하고, 중복 생성을 어떤 key로 막을 것인가? |
| Settlement skeleton 작성 | Ledger에 쌓인 확정 금액을 지급 가능한 묶음으로 계산하는 첫 구조가 필요하다. | 중 | Settlement 대상은 merchant로 시작하되, 추후 사용자/파트너 계정으로 확장 가능하게 설계할 것인가? |

## 오늘의 회고

### 오늘 가장 잘 이해된 개념

### 아직 가장 약한 개념

### 다음 구현에서 조심할 점
