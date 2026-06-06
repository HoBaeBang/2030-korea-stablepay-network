# Phase 2 Study Summary

이 문서는 Day 5 퇴근 후 실습에서 직접 채워 넣는 학습 산출물입니다.

관련 가이드: [Phase 2 Study Summary 실습 가이드](Phase_2_Study_Summary_실습가이드.md)

## 한 문장 요약

Phase 2는 Phase 1의 결제 백엔드를 Ledger, Settlement, Indexer, Deposit, Withdrawal이 안정적으로 붙을 수 있는 블록체인 금융 백엔드로 확장하는 단계이고, 첫 구현 범위는 작고 검증 가능한 Backend Core를 만드는 것이다.

## Day 1~4 학습 요약

| Day | 주제 | 내가 이해한 핵심 |
| --- | --- | --- |
| Day 1 | Phase 2 도메인 지도 | Phase 1의 Merchant, Invoice, Payment 중심 백엔드를 Ledger, Settlement, Indexer, Deposit, Withdrawal, Wallet이 붙을 수 있는 블록체인 금융 백엔드로 확장한다. |
| Day 2 | Ledger와 Settlement | Payment는 결제 상태를 나타내고, Ledger는 돈의 이동을 기록하며, Settlement는 확정된 결제 중 가맹점에게 지급 가능한 금액을 묶어 계산하고 처리한다. |
| Day 3 | Deposit, Withdrawal, Wallet, Key Security | Deposit은 이미 발생한 온체인 입금을 감지하고 반영하는 흐름이고, Withdrawal은 우리가 직접 transaction을 만들고 서명해 전송하는 흐름이라 더 강한 보안 경계가 필요하다. |
| Day 4 | Blockchain Event Indexer | Event Indexer는 온체인 block, transaction, event/log를 읽어 내부 deposit, payment, ledger 상태로 반영하는 off-chain worker이며, checkpoint, idempotency, reconciliation이 중요하다. |

## 아직 약한 개념

- 각 영어 용어의 한글 의미와 실제 역할을 더 익숙하게 만들어야 한다. 예: ledger, settlement, finality, reconciliation, idempotency.
- Phase 1과 Phase 2의 근본적 차이를 더 자연스럽게 설명할 수 있어야 한다.
- 우리가 만드는 서비스가 기존 결제/금융 서비스와 무엇이 같고 무엇이 다른지 더 정리해야 한다.
- 전체 흐름에서 어떤 기능이 어느 시점에 발생하고, 어떤 데이터가 어떤 도메인으로 이어지는지 더 반복해서 확인해야 한다.
- 상태 전이 정책을 누가 결정하고, 어떤 조건에서 상태가 바뀌는지 아직 더 학습이 필요하다.

## Phase 2 첫 구현 후보 비교

| 후보 | 지금 시작해도 되는 이유 | 아직 위험한 이유 | 내 판단 |
| --- | --- | --- | --- |
| Backend Core | 공통 에러 응답, validation, config, logging, 테스트 패턴은 지금도 정리할 수 있고 이후 모든 도메인이 이 기반을 사용한다. | 상태 변경 정책과 도메인별 세부 규칙이 아직 완전히 정리되지 않았다. | 가장 먼저 진행한다. 단순 공통 코드 정리가 아니라 이후 상태 전이와 도메인 기능이 붙을 수 있는 기반으로 만든다. |
| Ledger Core | Payment와 Settlement를 연결하려면 돈의 이동 기록이 필요하므로 중요한 구현 후보이다. | 공통 에러/검증/테스트 구조가 약하면 Ledger가 먼저 생겨도 코드 일관성과 검증 기준이 흔들릴 수 있다. | Backend Core 이후에 진행하는 것이 좋다. 다만 Backend Core 작업 중 Ledger에 필요한 요구사항을 계속 같이 확인한다. |
| Indexer Skeleton | polling 방식으로 온체인 이벤트를 읽는 구조를 작게 설계해볼 수 있다. | Ledger, idempotency, finality, payment 상태 전이 기준이 약하면 이벤트를 읽어도 안전하게 반영하기 어렵다. | 지금 당장 구현하기보다는 설계 후보로 남기고, Backend Core와 Ledger 기반이 생긴 뒤 구현한다. |

## 추천 첫 구현 작업

Sprint 2에서 가장 먼저 구현할 작업은 Backend Core 정리다.

이유는 Phase 2의 핵심 기능들이 모두 공통 기반 위에서 동작하기 때문이다. Ledger, Settlement, Indexer, Deposit, Withdrawal을 각각 따로 구현하면 기능은 생길 수 있지만 에러 응답, 요청 검증, 설정 관리, 로그, 테스트 방식이 흩어질 수 있다.

따라서 먼저 API로 들어오는 요청이 어떻게 검증되고, 실패했을 때 어떤 응답을 반환하고, service/repository에서 발생한 오류를 handler가 어떻게 변환할지 정리해야 한다. 이 기반이 있어야 이후 Ledger나 Indexer를 붙일 때도 같은 방식으로 개발하고 테스트할 수 있다.

다만 Backend Core 작업은 단순히 공통 코드만 만드는 것이 아니라, 앞으로 상태 변경 정책과 도메인 기능이 붙을 수 있는 방향으로 작게 vertical slice를 구성해야 한다.

## Sprint 2 백로그 후보

| 백로그 후보 | 필요한 이유 | 완료 기준 |
| --- | --- | --- |
| 공통 에러 응답 | API 실패 응답 형식을 통일해야 클라이언트와 테스트가 예측 가능해진다. | 모든 handler가 같은 error envelope를 반환한다. |
| 요청 validation 정리 | request body, path variable 검증을 일관되게 해야 잘못된 요청을 초기에 차단할 수 있다. | 잘못된 요청에 대해 명확한 400 응답을 준다. |
| 설정 구조 정리 | PORT, DATABASE_URL 같은 환경 변수를 config 구조로 모아야 실행 환경을 안전하게 관리할 수 있다. | main에서 config를 읽고 의존성에 전달한다. |
| logging 정리 | 요청, 실패, 상태 변경 로그가 있어야 장애 분석과 운영 추적이 가능하다. | 주요 상태 변경 시 로그가 남는다. |
| API boundary 정리 | 외부 공개 API와 내부 처리 책임을 구분해야 Phase 2 기능 확장 시 책임이 섞이지 않는다. | 외부 공개 API와 내부 처리 책임이 문서화된다. |
| 테스트 패턴 정리 | handler/service/repository 테스트의 기본 형태가 있어야 새 기능을 추가할 때 빠르게 검증할 수 있다. | 새 기능 추가 시 따라갈 테스트 예시가 생긴다. |

## Day 6에서 다시 확인할 질문

- Phase 1과 Phase 2의 차이를 내 말로 설명할 수 있는가?
- Payment, Ledger, Settlement의 책임 차이를 설명할 수 있는가?
- Deposit과 Withdrawal의 방향과 위험 차이를 설명할 수 있는가?
- Event Indexer가 어디에서 실행되고 무엇을 반영하는지 설명할 수 있는가?
- Backend Core를 Ledger나 Indexer보다 먼저 구현해야 하는 이유를 설명할 수 있는가?
- Sprint 2에서 바로 구현할 수 있는 백로그와 아직 더 설계가 필요한 백로그를 구분할 수 있는가?
    
## 검증 체크리스트

- [x] Phase 2의 핵심 도메인을 설명할 수 있다.
- [x] Backend Core를 먼저 추천하는 이유를 설명할 수 있다.
- [x] Sprint 2 백로그 후보를 3개 이상 말할 수 있다.
- [x] Day 6에서 점검할 약한 개념을 정리했다.
