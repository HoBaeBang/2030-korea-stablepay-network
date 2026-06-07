# Phase 2 통합 복습 검증문제와 답변가이드

관련 Jira: [SPN-23](https://aslan0.atlassian.net/browse/SPN-23)

이 문서는 Day 6 학습 후 스스로 확인할 검증문제와 답변가이드입니다.

질문을 먼저 풀어보고, 아래 답변가이드와 비교합니다.

## 진행 방법

이 문서는 위에서 아래로 한 번에 읽는 문서가 아닙니다.

1. 먼저 `검증문제`만 보고 답을 작성한다.
2. 답을 작성할 때는 문장을 완벽하게 만들기보다, 본인이 이해한 흐름을 적는다.
3. 그 다음 `답변가이드`를 본다.
4. 내가 쓴 답과 답변가이드의 차이를 표시한다.
5. 부족한 개념은 [Phase 2 Review Checkpoint 실습산출물](Phase_2_Review_Checkpoint_실습산출물.md)에 다시 적는다.

중요한 것은 정답을 맞히는 것이 아니라, 구현 전에 어떤 개념이 아직 약한지 발견하는 것입니다.

## 검증문제

1. Phase 1과 Phase 2의 가장 큰 차이는 무엇인가?
2. Payment 상태와 Ledger 기록은 왜 분리해야 하는가?
3. Ledger Account, Ledger Transaction, Ledger Entry의 차이를 설명해보라.
4. Settlement는 왜 단순히 모든 결제 금액을 더하는 작업이 아닌가?
5. Deposit과 Withdrawal의 가장 큰 차이는 무엇인가?
6. Withdrawal에서 transaction signing은 Ledger에 기록하는 것과 같은 일인가?
7. Event Indexer는 어디에서 실행되는가?
8. `chain + tx_hash + log_index`가 idempotency key 후보가 되는 이유는 무엇인가?
9. Reconciliation은 어떤 상태들을 비교하는가?
10. Sprint 2에서 Backend Core vertical slice를 먼저 구현하는 이유는 무엇인가?
11. Day 6에서 약한 개념 목록을 작성하는 이유는 무엇인가?
12. 지금 바로 구현을 시작하기 전에 반드시 확인해야 하는 질문 3가지를 적어보라.

## 답변가이드

### 1. Phase 1과 Phase 2의 가장 큰 차이

Phase 1은 Merchant, Invoice, Payment 중심의 결제 백엔드 MVP입니다.

Phase 2는 여기에 Ledger, Settlement, Blockchain Event Indexer, Deposit, Withdrawal, Wallet, Key Security를 붙여 블록체인 금융 백엔드로 확장하는 단계입니다.

### 2. Payment와 Ledger를 분리하는 이유

Payment는 결제의 진행 상태를 나타냅니다.

Ledger는 돈이 어느 계정에서 어느 계정으로 왜 이동했는지 기록합니다.

상태와 돈의 이동 기록을 분리해야 중복 결제, 장애 복구, 정산 검증, 회계 추적이 가능합니다.

### 3. Ledger Account, Transaction, Entry의 차이

Ledger Account는 돈이 귀속되는 주체입니다.

Ledger Transaction은 하나의 돈 이동 사건을 묶는 단위입니다.

Ledger Entry는 특정 account에 금액이 증가하거나 감소한 개별 기록입니다.

예를 들어 10 USDC 결제라면 customer 쪽에는 `-10`, merchant 쪽에는 `+10` entry가 생길 수 있고, 이 둘이 하나의 ledger transaction으로 묶입니다.

### 4. Settlement가 단순 합계가 아닌 이유

Settlement는 가맹점에게 지급 가능한 금액을 계산하는 과정입니다.

Finality가 부족하거나, 중복 ledger entry 의심이 있거나, 환불/실패 상태와 충돌하거나, 정산 정책을 만족하지 않으면 바로 지급 대상이 아닐 수 있습니다.

그래서 Settlement는 단순 합계가 아니라 지급 가능성 검증과 묶음 생성 과정입니다.

### 5. Deposit과 Withdrawal의 차이

Deposit은 외부에서 우리 시스템으로 이미 들어온 자산을 감지하는 흐름입니다.

Withdrawal은 우리 시스템에서 외부 주소로 자산을 내보내는 흐름입니다.

Deposit은 감지와 중복 반영 방지가 중요하고, Withdrawal은 승인, 주소 검증, 키 보안, 중복 출금 방지가 중요합니다.

### 6. Withdrawal signing과 Ledger 기록의 차이

둘은 다른 일입니다.

Signing은 블록체인에 전송할 transaction에 개인키로 서명하는 일입니다.

Ledger 기록은 내부 DB에 돈의 이동을 장부로 남기는 일입니다.

서명은 온체인 전송을 가능하게 하는 암호학적 작업이고, Ledger는 내부 회계/정산/복구를 위한 기록입니다.

### 7. Event Indexer가 실행되는 위치

Event Indexer는 블록체인 안에서 실행되지 않습니다.

우리 백엔드의 off-chain worker layer에서 실행되며, RPC를 통해 블록체인의 block, transaction, event를 조회합니다.

### 8. Idempotency key 후보

같은 chain에서 transaction hash와 log index 조합은 특정 event를 식별할 수 있습니다.

따라서 `chain + tx_hash + log_index`를 저장해두면 같은 이벤트를 다시 읽어도 이미 처리한 이벤트인지 확인할 수 있습니다.

### 9. Reconciliation이 비교하는 상태

Reconciliation은 내부 DB 상태와 온체인 상태를 비교합니다.

예를 들어 blockchain event에는 입금이 있는데 내부 deposit이나 payment가 아직 반영되지 않았는지, ledger entry가 중복으로 생기지 않았는지, checkpoint가 실제 처리 상태와 맞는지 확인합니다.

### 10. Backend Core vertical slice를 먼저 구현하는 이유

Phase 2에는 많은 도메인이 있지만 모든 것을 한 번에 구현하면 위험합니다.

Backend Core vertical slice를 먼저 구현하면 Payment, Ledger, Settlement의 최소 연결을 작게 검증할 수 있고, 이후 Indexer, Deposit, Withdrawal을 붙일 기반을 만들 수 있습니다.

### 11. 약한 개념 목록을 작성하는 이유

구현 전에 모르는 것을 드러내기 위해서입니다.

약한 개념을 목록화하면 다음 학습이나 구현 티켓에서 무엇을 보강해야 하는지 명확해집니다.

### 12. 구현 전 확인 질문 예시

- Ledger entry는 어떤 idempotency key를 기준으로 중복 생성을 막을 것인가?
- Payment가 `FINALIZED`가 되는 기준은 mock으로 둘 것인가, 실제 chain confirmation 기준으로 둘 것인가?
- Settlement는 어떤 payment 상태부터 포함할 수 있는가?
- Backend Core 첫 구현 범위는 API, service, repository, migration 중 어디까지 포함할 것인가?
