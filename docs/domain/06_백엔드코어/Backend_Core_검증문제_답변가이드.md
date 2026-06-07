# Backend Core 검증문제와 답변가이드

관련 Jira: [SPN-24](https://aslan0.atlassian.net/browse/SPN-24)

이 문서는 Day 7 학습 후 스스로 확인할 검증문제와 답변가이드입니다.

먼저 `검증문제`만 보고 답을 작성한 뒤, 아래 답변가이드와 비교합니다.

## 검증문제

1. Backend Core는 특정 도메인인가, 공통 기반인가?
2. Phase 2에서 Backend Core를 먼저 정리해야 하는 이유는 무엇인가?
3. 공통 에러 응답이 없으면 어떤 문제가 생기는가?
4. Handler, Service, Repository의 validation 책임은 어떻게 나눌 수 있는가?
5. Config를 구조체로 모으는 이유는 무엇인가?
6. 블록체인 결제 백엔드에서 logging이 중요한 이유는 무엇인가?
7. Go 테스트에서 한글 subtest를 쓰면 어떤 장점이 있는가?
8. Day 7에서 Ledger 전체 구현을 바로 하지 않는 이유는 무엇인가?
9. Backend Core 이후 Ledger 구현으로 넘어갈 때 가장 먼저 확인해야 할 것은 무엇인가?
10. Sprint 2 첫 구현 후보 중 가장 먼저 할 만한 작업은 무엇이고, 이유는 무엇인가?

## 답변가이드

### 1. Backend Core의 성격

Backend Core는 특정 도메인 하나가 아니라 공통 기반입니다.

Merchant, Invoice, Payment, Ledger, Settlement, Indexer, Deposit, Withdrawal이 공통으로 사용하는 에러 처리, 요청 검증, 설정, 로그, 테스트 패턴을 정리하는 영역입니다.

### 2. Backend Core를 먼저 정리하는 이유

Phase 2에서는 돈의 이동과 상태 변경이 많아집니다.

공통 에러 응답, validation, config, logging, test pattern이 정리되지 않은 상태에서 Ledger나 Indexer를 붙이면 코드 스타일이 흔들리고 장애 추적이 어려워집니다.

### 3. 공통 에러 응답이 없을 때의 문제

API마다 실패 응답 형식이 달라집니다.

사용자도 실패 원인을 이해하기 어렵고, 테스트도 매번 다른 응답 형태를 검증해야 합니다.

나중에 프론트엔드나 외부 클라이언트가 붙을 때도 에러 처리 비용이 커집니다.

### 4. Validation 책임 분리

Handler는 JSON 파싱, path variable, 필수 필드 같은 요청 형식 검증을 담당합니다.

Service는 지원 통화, 상태 전이, 금액 정책 같은 도메인 규칙을 검증합니다.

Repository는 DB 저장/조회에 집중하고, 도메인 규칙 판단의 중심이 되지 않도록 합니다.

### 5. Config 구조체가 필요한 이유

설정이 여러 파일에 흩어지면 실행 환경을 이해하기 어렵습니다.

Phase 2에서는 `BLOCKCHAIN_RPC_URL`, `FINALITY_CONFIRMATIONS`, `SIGNER_BASE_URL` 같은 설정이 늘어날 수 있으므로, 설정을 구조체로 모아 의존성 주입하기 쉽게 만드는 것이 좋습니다.

### 6. Logging이 중요한 이유

블록체인 결제 백엔드에서는 상태 변경과 돈의 이동을 추적해야 합니다.

예를 들어 payment가 왜 `FINALIZED` 되었는지, ledger entry가 언제 생성됐는지, indexer가 어디 block까지 처리했는지 로그가 있어야 장애 분석과 reconciliation이 가능합니다.

### 7. 한글 subtest의 장점

테스트 이름이 문서처럼 읽힙니다.

특히 학습 단계에서는 `지원하지 않는 통화이면 invoice를 생성할 수 없다`처럼 테스트 의도를 바로 이해할 수 있어 복습과 유지보수에 도움이 됩니다.

### 8. Ledger 전체 구현을 바로 하지 않는 이유

Ledger는 돈의 이동을 다루는 핵심 기능입니다.

공통 에러, validation, test pattern이 약한 상태에서 바로 구현하면 중복 entry, 잘못된 상태 전이, 테스트 누락 위험이 커집니다.

### 9. Ledger 구현 전 확인할 것

가장 먼저 확인할 것은 Payment와 Ledger의 연결 지점입니다.

예를 들어 어떤 Payment 상태 전이에서 Ledger transaction과 entry를 만들지, 중복 생성을 어떤 idempotency key로 막을지 정해야 합니다.

### 10. 가장 먼저 할 만한 작업

공통 에러 응답 정리가 가장 먼저 할 만합니다.

모든 API가 실패 응답을 같은 형태로 반환하면 validation, service error, repository error를 이후 기능에서도 일관되게 다룰 수 있기 때문입니다.
