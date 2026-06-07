# Backend Core 일정과 티켓 관리

관련 Jira Epic: [SPN-2 Blockchain Backend Core](https://aslan0.atlassian.net/browse/SPN-2)

이 문서는 `SPN-2 Blockchain Backend Core` 에픽을 Day 단위로 어떻게 진행할지 정리한 운영 문서입니다.

## 결론

Backend Core 에픽은 Day 7부터 Day 11까지, 총 5일 일정으로 진행합니다.

이 에픽의 목표는 Ledger, Settlement, Event Indexer를 바로 구현하는 것이 아니라, 그 기능들이 안정적으로 붙을 수 있는 Go 백엔드 공통 기반을 먼저 만드는 것입니다.

```text
Day 7  Backend Core 구현 준비
Day 8  공통 에러 응답과 요청 검증
Day 9  설정 로딩과 애플리케이션 시작 구조
Day 10 로깅과 테스트 패턴
Day 11 Backend Core 통합 복습과 Ledger 구현 준비
```

## 왜 5일로 잡는가?

너무 짧게 잡으면 공통 기반을 제대로 이해하지 못한 채 Ledger 구현으로 넘어가게 됩니다.

너무 길게 잡으면 실제 도메인 구현으로 들어가는 속도가 늦어집니다.

현재 프로젝트는 이미 merchant, invoice, payment의 기본 구조가 있으므로 Backend Core는 새 기능을 크게 만드는 단계가 아니라 기존 구조를 정리하고 반복 가능한 패턴을 만드는 단계입니다.

그래서 5일이 적당합니다.

## 전체 흐름

```text
Phase 2 Domain Foundation
        |
        v
Backend Core
        |
        +--> Error Response
        +--> Validation
        +--> Config
        +--> Logging
        +--> Test Pattern
        |
        v
Ledger Implementation
```

## Day별 목표

| Day | Jira | 주제 | 완료 기준 |
| --- | --- | --- | --- |
| Day 7 | SPN-24 | Backend Core 구현 준비와 Sprint 2 작업 분해 | 공통 기반 구현 후보를 설명할 수 있다. |
| Day 8 | [SPN-25](https://aslan0.atlassian.net/browse/SPN-25) | 공통 에러 응답과 요청 검증 | API 실패 응답과 validation 위치를 코드로 정리한다. |
| Day 9 | [SPN-26](https://aslan0.atlassian.net/browse/SPN-26) | 설정 로딩과 애플리케이션 시작 구조 | `PORT`, `DATABASE_URL` 같은 설정을 한 곳에서 읽는다. |
| Day 10 | [SPN-27](https://aslan0.atlassian.net/browse/SPN-27) | 로깅과 테스트 패턴 | 상태 변경 로그와 한글 subtest 테스트 패턴을 정리한다. |
| Day 11 | [SPN-28](https://aslan0.atlassian.net/browse/SPN-28) | Backend Core 통합 복습과 Ledger 준비 | Ledger 구현 전 체크리스트를 완성한다. |

## Jira 상태 운영 기준

| 상태 | 의미 |
| --- | --- |
| 백로그 | 아직 진행하지 않을 티켓 |
| 진행 예정 | 자료가 준비되어 있고 다음에 시작할 수 있는 티켓 |
| 진행 중 | 사용자가 학습/실습을 시작한 티켓 |
| 검토 중 | 산출물 작성 후 Codex 검토가 필요한 티켓 |
| 완료 | 산출물, 검증문제, 필요한 코드/문서 반영이 끝난 티켓 |

## Sprint 관리 기준

현재 Rovo 도구에서는 Jira의 실제 Sprint 객체를 직접 생성하거나 이슈를 Sprint에 배정하는 기능이 노출되어 있지 않습니다.

따라서 우선은 다음 기준으로 운영합니다.

```text
Sprint 1 = Phase 2 Domain Foundation
Sprint 2 = Blockchain Backend Core
Sprint 3 = Ledger / Settlement Implementation
```

Jira 화면에서 별도의 Sprint 보드를 만들고 싶다면 Jira UI에서 Sprint를 생성하고, 해당 티켓들을 Sprint 2로 이동하면 됩니다.

우리는 우선 에픽과 상태를 기준으로 관리합니다.

## Backend Core 이후 다음 흐름

Backend Core가 끝나면 다음 에픽은 Ledger 구현으로 넘어갑니다.

예상 흐름:

```text
Backend Core 완료
        |
        v
Ledger Account / Transaction / Entry 모델 구현
        |
        v
Payment finalized -> Ledger transaction 생성
        |
        v
Settlement skeleton 구현
```

## 학습 포인트

Backend Core는 포트폴리오에서 화려해 보이는 기능은 아닙니다.

하지만 실무에서는 다음 역량을 보여주는 핵심 구간입니다.

1. API 에러 응답을 일관되게 설계할 수 있다.
2. 요청 검증과 도메인 검증을 분리할 수 있다.
3. 설정을 코드에 흩뿌리지 않고 관리할 수 있다.
4. 로그로 장애 원인을 추적할 수 있다.
5. 테스트 패턴을 반복 가능하게 만들 수 있다.
