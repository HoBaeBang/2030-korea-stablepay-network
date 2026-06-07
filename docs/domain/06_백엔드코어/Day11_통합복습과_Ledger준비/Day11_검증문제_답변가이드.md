# Day 11 검증문제와 답변가이드

관련 Jira: [SPN-28](https://aslan0.atlassian.net/browse/SPN-28)

## 먼저 풀어볼 문제

1. Backend Core를 Ledger보다 먼저 정리한 이유는 무엇인가?
2. Error Response, Validation, Config, Logging, Test Pattern은 각각 어떤 역할을 하는가?
3. Ledger 구현 전 validation이 중요한 이유는 무엇인가?
4. Ledger 구현 전 logging이 중요한 이유는 무엇인가?
5. Ledger 구현 전 test pattern이 중요한 이유는 무엇인가?
6. payment finalized를 중복 처리하면 어떤 문제가 생길 수 있는가?
7. DB transaction 없이 ledger transaction과 entry를 따로 저장하면 어떤 문제가 생길 수 있는가?
8. SPN-2 에픽을 완료 처리하기 위한 기준은 무엇인가?

## 답변가이드

### 1. Backend Core를 먼저 정리한 이유

Ledger, Settlement, Event Indexer는 모두 안정적인 백엔드 공통 기반 위에서 동작해야 합니다.

공통 에러, 검증, 설정, 로그, 테스트가 없으면 이후 기능이 커질수록 수정과 장애 추적이 어려워집니다.

### 2. 각 영역의 역할

Error Response는 실패 응답 형식을 통일합니다.

Validation은 잘못된 요청과 도메인 규칙 위반을 막습니다.

Config는 실행 설정을 관리합니다.

Logging은 운영 중 발생한 일을 추적합니다.

Test Pattern은 반복 가능한 검증 방식을 만듭니다.

### 3. Ledger 전 validation이 중요한 이유

Ledger는 돈의 이동을 기록하므로 잘못된 요청이 들어오면 잘못된 돈의 이동으로 이어질 수 있습니다.

### 4. Ledger 전 logging이 중요한 이유

돈의 이동과 상태 변경을 추적하지 못하면 장애 발생 시 원인을 찾기 어렵습니다.

### 5. Ledger 전 test pattern이 중요한 이유

Ledger는 정합성이 매우 중요하므로 테스트 없이 구현하면 중복 기록, 누락 기록 같은 문제가 생기기 쉽습니다.

### 6. payment finalized 중복 처리 문제

같은 결제에 대해 ledger transaction이 두 번 생성될 수 있습니다.

그러면 실제 돈은 한 번 움직였는데 내부 원장은 두 번 반영되는 문제가 생깁니다.

### 7. DB transaction 없이 따로 저장하는 문제

ledger transaction은 저장되었는데 entry가 일부만 저장되는 식으로 데이터 정합성이 깨질 수 있습니다.

### 8. SPN-2 완료 기준

Backend Core의 주요 공통 기반이 문서와 코드 후보로 정리되고, Ledger 구현을 시작할 수 있는 체크리스트가 준비되면 완료로 볼 수 있습니다.
