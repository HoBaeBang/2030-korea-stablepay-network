# Backend Core 실습산출물

관련 Jira: [SPN-24](https://aslan0.atlassian.net/browse/SPN-24)

이 문서는 Day 7 퇴근 후 직접 작성하는 실습산출물입니다.

## 한 문장 요약

> Backend Core는 Phase 2의 Ledger, Settlement, Indexer, 입출금 기능이 안정적으로 붙을 수 있도록 공통 에러, 검증, 설정, 로그, 테스트 패턴을 정리하는 기반이다.

## 현재 코드 구조 관찰

| 경로 | 내가 이해한 역할 | Backend Core 관점에서 확인할 점 |
| --- | --- | --- |
| `cmd/api` |  | config를 읽고 의존성을 조립하는 위치로 적절한가? |
| `internal/httpapi` |  | 공통 error response와 route 등록을 정리할 수 있는가? |
| `internal/merchant` |  | service/repository/test 패턴을 참고할 수 있는가? |
| `internal/invoice` |  | request validation 패턴을 참고할 수 있는가? |
| `internal/payment` |  | 상태 전이와 에러 처리를 참고할 수 있는가? |
| `internal/platform/database` |  | DB 연결과 transaction boundary를 확장할 수 있는가? |
| `migrations` |  | Phase 2 schema 변경을 어떤 순서로 관리할 것인가? |

## 공통 에러 응답 초안

### 응답 형태 후보

```json
{
  "error": {
    "code": "invalid_request",
    "message": "amount must be greater than zero",
    "field": "amount"
  }
}
```

### error code 후보

| code | 의미 | 사용 예시 |
| --- | --- | --- |
| `invalid_request` | 요청 형식이나 필드 값이 잘못됨 | amount가 0 이하, currency 누락 |
| `not_found` | 대상 리소스를 찾을 수 없음 | merchant, invoice, payment 없음 |
| `conflict` | 현재 상태와 요청이 충돌함 | 이미 finalized된 payment를 다시 변경 |
| `internal_error` | 서버 내부 오류 | DB 오류, 예상하지 못한 실패 |

## Validation 위치 정리

| 대상 | Handler에서 검증 | Service에서 검증 | Repository에서 맡지 않을 것 |
| --- | --- | --- | --- |
| Invoice 생성 | JSON 파싱, amount/currency 필수 여부 | 지원 통화인지, 금액 정책을 만족하는지 | 요청 필드 의미 검증 |
| Payment 상태 변경 | path variable, body 형식 | 허용된 상태 전이인지 | 상태 전이 정책 판단 |
| Ledger entry 생성 | 요청 형식 또는 내부 command 형식 | debit/credit 합계가 0인지, 중복 key가 없는지 | 복식부기 정책 판단 |

## Config 후보

| 설정 | 필요한 이유 | 당장 구현 여부 |
| --- | --- | --- |
| `PORT` | API 서버 포트 | 필요 |
| `DATABASE_URL` | PostgreSQL 연결 | 필요 |
| `LOG_LEVEL` | 로그 수준 제어 | 후보 |
| `BLOCKCHAIN_RPC_URL` | Indexer/RPC 연결 | 나중 |
| `INDEXER_POLL_INTERVAL` | polling 주기 | 나중 |
| `FINALITY_CONFIRMATIONS` | finality 판단 기준 | 나중 |
| `SIGNER_BASE_URL` | Rust signer 호출 | 나중 |

## Logging 후보

| 이벤트 | 로그가 필요한 이유 |
| --- | --- |
| payment status changed | payment가 왜 변경됐는지 추적하기 위해 |
| ledger transaction created | 돈 이동 기록 생성 시점을 추적하기 위해 |
| ledger duplicate prevented | 멱등성으로 중복 처리를 막았는지 확인하기 위해 |
| settlement batch created | 정산 묶음 생성 근거를 추적하기 위해 |
| indexer checkpoint advanced | 어디 block까지 처리했는지 확인하기 위해 |
| withdrawal signed | signer 요청과 결과를 추적하기 위해 |
| withdrawal broadcasted | tx hash와 네트워크 전송 결과를 추적하기 위해 |

## Test Pattern 후보

앞으로 테스트는 가능하면 다음 구조를 반복한다.

```text
given: 어떤 데이터와 상태가 주어졌는가
when: 어떤 동작을 실행했는가
then: 어떤 결과를 기대하는가
```

예시 subtest:

```text
지원하지 않는 통화이면 invoice를 생성할 수 없다
허용되지 않은 상태 전이는 실패한다
같은 idempotency key로 ledger entry가 두 번 생성되지 않는다
```

## Sprint 2 첫 구현 후보

| 후보 작업 | 우선순위 | 이유 | 먼저 확인할 질문 |
| --- | --- | --- | --- |
| 공통 에러 응답 정리 | 1 | 모든 API와 이후 Ledger/Settlement 기능에서 바로 사용된다. | 현재 handler들이 실패를 어떻게 반환하고 있는가? |
| 요청 validation 정리 | 2 | 잘못된 요청이 service/repository로 내려가는 것을 막는다. | handler와 service의 검증 책임을 어디까지 나눌 것인가? |
| config 구조 정리 | 3 | Phase 2에서 RPC, signer, finality 설정이 늘어난다. | 현재 main에서 어떤 설정을 직접 읽고 있는가? |
| 테스트 패턴 정리 | 4 | 이후 기능 구현 때 반복 가능한 검증 구조가 필요하다. | 기존 service test의 좋은 패턴과 부족한 점은 무엇인가? |
| logging 정리 | 5 | 상태 변경과 장애 추적에 필요하다. | 표준 라이브러리 log로 시작할지, 구조화 로그를 도입할지? |

## 오늘의 판단

선택:

- [ ] 구현 시작 가능
- [ ] 부분 보강 후 시작
- [ ] 하루 더 설계 필요

판단 이유:

```text
여기에 판단 이유를 작성한다.
```

## 오늘의 회고

### 오늘 가장 잘 이해된 개념

### 아직 가장 약한 개념

### 다음 구현에서 조심할 점
