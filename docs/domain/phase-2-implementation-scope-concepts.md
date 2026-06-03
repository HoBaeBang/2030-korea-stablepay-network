# Phase 2 첫 구현 범위 결정 개념 학습

관련 Jira: [SPN-22](https://aslan0.atlassian.net/browse/SPN-22)

이 문서는 출퇴근 시간에 읽는 Day 5 개념 학습자료입니다.

## 1. 왜 Day 5에서 구현 범위를 결정하는가

Day 1~4에서 우리는 Phase 2의 핵심 도메인을 학습했습니다.

| Day | 학습 주제 | 핵심 질문 |
| --- | --- | --- |
| Day 1 | Phase 2 전체 도메인 지도 | Phase 1 MVP가 어떤 금융 백엔드로 확장되는가 |
| Day 2 | Ledger와 Settlement | 돈의 이동과 정산은 왜 Payment 상태만으로 부족한가 |
| Day 3 | Deposit, Withdrawal, Wallet, Key Security | 입출금과 서명 경계는 왜 일반 CRUD가 아닌가 |
| Day 4 | Blockchain Event Indexer | 온체인 이벤트를 어떻게 읽고 중복 없이 반영할 것인가 |

Day 5에서는 이 지식을 바탕으로 첫 구현 범위를 결정합니다.

## 2. 첫 구현 후보 비교

![Phase 2 First Implementation Decision Map](../confluence/diagrams/spn22-scope-decision-map.png)

첫 구현 후보는 크게 세 가지입니다.

| 후보 | 장점 | 위험 | 판단 |
| --- | --- | --- | --- |
| Backend Core | 이후 모든 기능의 공통 패턴을 만든다 | 블록체인 기능처럼 화려해 보이지 않을 수 있다 | 가장 먼저 추천 |
| Ledger Core | 돈의 이동 기록이라는 핵심에 바로 들어간다 | 공통 에러/검증/테스트 구조가 약하면 구현이 흔들릴 수 있다 | Sprint 3 추천 |
| Indexer Skeleton | 블록체인 연결을 빨리 보여줄 수 있다 | Ledger/idempotency/상태 전이 기반이 약하면 위험하다 | Sprint 5 추천 |

## 3. 왜 Backend Core를 먼저 추천하는가

Backend Core는 블록체인 도메인 자체는 아니지만, 이후 모든 블록체인 기능이 의존하는 기반입니다.

예를 들어 Ledger, Settlement, Indexer, Withdrawal을 구현할 때 모두 다음이 필요합니다.

```text
공통 에러 응답
요청 validation
config 구조
logging
service/repository 테스트 패턴
상태 변경 실패 처리
```

이 기반 없이 바로 Ledger나 Indexer로 들어가면, 기능은 생겨도 코드의 일관성이 떨어지고 테스트하기 어려워질 수 있습니다.

## 4. Vertical Slice로 작게 구현하기

![Vertical Slice Scope](../confluence/diagrams/spn22-vertical-slice-scope.png)

`Vertical slice`는 한 기능을 DB, domain, repository, service, API/test까지 얇게 끝까지 연결하는 방식입니다.

Day 5에서 결정할 것은 Phase 2 전체가 아니라, 다음 Sprint에서 끝낼 수 있는 첫 조각입니다.

## 5. Sprint 2 백로그 후보

![Sprint 2 Backlog Flow](../confluence/diagrams/spn22-sprint2-backlog-flow.png)

Sprint 2의 목표는 `Backend Core 정리`입니다.

추천 백로그 후보:

| 후보 | 설명 | 완료 기준 |
| --- | --- | --- |
| 공통 에러 응답 | API 실패 응답 형식을 통일한다 | 모든 handler가 같은 error envelope를 반환한다 |
| 요청 validation 정리 | request body, path variable 검증을 일관되게 만든다 | 잘못된 요청에 대해 명확한 400 응답을 준다 |
| 설정 구조 정리 | PORT, DATABASE_URL 같은 환경변수를 config 구조로 모은다 | main에서 config를 읽고 의존성에 전달한다 |
| logging 정리 | 요청, 실패, 상태 변경 로그를 남긴다 | 주요 상태 변경 시 로그가 남는다 |
| API boundary 정리 | public API와 internal API 방향을 나눈다 | 외부 공개 API와 내부 처리 책임이 문서화된다 |
| 테스트 패턴 정리 | handler/service 테스트 전략을 정리한다 | 새 기능 추가 시 따라갈 테스트 예시가 생긴다 |

## 6. 오늘 기억할 요약

```text
Day 5의 결정은 "가장 멋진 기능"을 고르는 것이 아니다.
다음 Sprint에서 작게 끝까지 구현할 수 있고,
이후 모든 Phase 2 기능의 기반이 되는 첫 조각을 고르는 것이다.

추천 첫 구현 범위는 Backend Core다.
```
