# Day 11 개념학습 - Backend Core 통합 복습과 Ledger 구현 준비

관련 Jira: [SPN-28](https://aslan0.atlassian.net/browse/SPN-28)

## Backend Core에서 정리한 것

| 영역 | 역할 |
| --- | --- |
| Error Response | API 실패 응답을 일관되게 만든다. |
| Validation | 잘못된 요청과 도메인 규칙 위반을 막는다. |
| Config | 실행 환경 설정을 한 곳에서 관리한다. |
| Logging | 운영 중 발생한 일을 추적한다. |
| Test Pattern | 반복 가능한 검증 방식을 만든다. |

## Backend Core와 Ledger의 연결

![Day11 Backend Core에서 Ledger로 넘어가는 준비도](../../../confluence/diagrams/spn28-day11-ledger-readiness.png)

Ledger는 `account`, `transaction`, `entry` 같은 새로운 모델을 추가하는 작업입니다.

하지만 모델만 추가한다고 Ledger가 안전해지는 것은 아닙니다.

Ledger가 안전하려면 다음 기반이 필요합니다.

| Backend Core 기반 | Ledger에서 연결되는 지점 |
| --- | --- |
| Error Response | 원장 생성 실패, 중복 요청, 잘못된 금액을 일관된 응답으로 표현 |
| Validation | debit/credit 합계 0, amount > 0, 중복 key 방어 |
| Config | DB 연결, 향후 chain/RPC 설정, indexer 설정 |
| Logging | ledger transaction 생성, 중복 요청 무시, 정산 묶음 생성 기록 |
| Test Pattern | 복식부기 불변식, 중복 방어, DB transaction 정합성 검증 |

## 왜 Ledger 전에 필요한가?

Ledger는 돈의 이동을 기록합니다.

따라서 Ledger 구현 전에 다음 질문에 답할 수 있어야 합니다.

```text
잘못된 요청은 어디에서 막을 것인가?
실패 응답은 어떤 모양으로 줄 것인가?
설정 누락은 언제 발견할 것인가?
중요한 상태 변경은 어떤 로그로 남길 것인가?
돈의 이동은 어떤 테스트로 검증할 것인가?
```

## Ledger 구현으로 넘어가기 전 체크포인트

Ledger 구현으로 넘어가기 전에는 다음 질문에 답할 수 있어야 합니다.

| 체크포인트 | 질문 |
| --- | --- |
| Error Response | Ledger 생성 실패를 어떤 `code`로 표현할 것인가 |
| Validation | debit/credit 합계가 0이 아닌 요청을 어디에서 막을 것인가 |
| Config | Ledger 구현에 필요한 DB 설정은 시작 시점에 검증되는가 |
| Logging | Ledger transaction이 생성될 때 어떤 식별자를 로그에 남길 것인가 |
| Test Pattern | 중복 finalized 처리 시 Ledger가 두 번 생성되지 않는 테스트가 있는가 |

## Ledger 첫 구현 후보

Backend Core 이후에는 다음 순서가 적절합니다.

1. Ledger Account 모델 설계
2. Ledger Transaction 모델 설계
3. Ledger Entry 모델 설계
4. migration 작성
5. repository 작성
6. service 작성
7. payment finalized 시 ledger transaction 생성

## Ledger에서 특히 조심할 위험

| 위험 | 설명 | 방어 방향 |
| --- | --- | --- |
| 중복 원장 생성 | 같은 payment finalized 이벤트가 두 번 처리되어 ledger transaction이 두 번 생김 | idempotency key, unique constraint, service test |
| 불균형 entry | debit 합계와 credit 합계가 맞지 않음 | service validation, DB transaction, 테스트 |
| 부분 저장 | transaction은 저장됐지만 entry 일부가 저장되지 않음 | DB transaction 사용 |
| 추적 불가 | 장애가 났지만 어떤 payment에서 원장이 생겼는지 모름 | payment_id, ledger_transaction_id 로그 |
| 응답 불일치 | API마다 실패 응답이 달라 클라이언트가 분기하기 어려움 | Day8 공통 error response 적용 |

## 오늘의 핵심 판단

Day 11에서는 “모든 것을 완벽히 아는가?”를 확인하는 것이 아닙니다.

다음 구현을 시작했을 때 막힐 위험이 큰 부분을 미리 발견하는 것이 목적입니다.
