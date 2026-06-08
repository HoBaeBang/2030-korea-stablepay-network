# Day7 실습산출물 - Backend Core 구현 준비

관련 Jira: [SPN-24](https://aslan0.atlassian.net/browse/SPN-24)

이 문서는 Day 7 퇴근 후 직접 작성하는 실습산출물입니다.

## 한 문장 요약

> Backend Core는 Phase 2의 Ledger, Settlement, Indexer, 입출금 기능이 안정적으로 붙을 수 있도록 공통 에러, 검증, 설정, 로그, 테스트 패턴을 정리하는 기반이다.

## 현재 코드 구조 관찰

| 경로 | 내가 이해한 역할                                                                                                | Backend Core 관점에서 확인할 점 |
| --- |----------------------------------------------------------------------------------------------------------| --- |
| `cmd/api` | 백엔드 애플리케이션의 시작점이다. DB 연결, service/repository 생성, HTTP route 등록처럼 서버 실행에 필요한 의존성을 조립한다. | config를 읽고 의존성을 조립하는 위치로 적절한가? |
| `internal/httpapi` | API request/response 처리와 route 진입점을 담당한다. 외부 HTTP 요청을 내부 service 호출로 연결한다. | 공통 error response와 route 등록을 정리할 수 있는가? |
| `internal/merchant` | merchant 도메인에 관련된 service, repository, test를 포함한다. 가맹점 정보를 생성하고 조회하는 기능을 지원한다. | service/repository/test 패턴을 참고할 수 있는가? |
| `internal/invoice` | invoice 도메인에 관련된 service, repository, test를 포함한다. 가맹점이 결제 요청서를 생성하는 기능을 제공한다. | request validation 패턴을 참고할 수 있는가? |
| `internal/payment` | invoice에 대한 결제 진행 상태와 상태 전이 정보를 기록한다. 결제 상태 변경 규칙을 확인할 수 있다. | 상태 전이와 에러 처리를 참고할 수 있는가? |
| `internal/platform/database` | 데이터베이스 연결을 만들고 유지하는 인프라 기능을 지원한다. | DB 연결과 transaction boundary를 확장할 수 있는가? |
| `migrations` | DB 테이블 구조와 변경 이력을 SQL로 정의한다. | Phase 2 schema 변경을 어떤 순서로 관리할 것인가? |

## 공통 에러 응답 초안

### 응답 형태 후보

```json
{
  "error": {
    "code": "invalid_request",
    "message": "결제 금액은 0보다 커야 합니다.",
    "field": "amount"
  }
}
```

### 응답 필드 의미

| 필드 | 의미 | 작성 기준 |
| --- | --- | --- |
| `code` | 프로그램이 에러 종류를 구분하기 위한 값 | 영어 snake_case로 작성한다. |
| `message` | 사용자가 읽는 설명 | 우리 서비스에서는 한글로 작성한다. |
| `field` | 문제가 발생한 요청 필드 | 특정 필드 문제일 때만 사용한다. |

`code`는 클라이언트나 테스트가 안정적으로 분기하기 위한 값이므로 영어로 유지한다.

반면 `message`는 사람이 읽는 문장이므로 한글로 작성하는 것이 더 좋다.

### error code 후보

우선 400번대 REST Client Error를 중심으로 정리한다.

| HTTP status | code | 의미 | 사용 예시 | message 예시 | 초기 구현 여부 |
| --- | --- | --- | --- | --- | --- |
| 400 | `bad_request` | 요청 형식이나 필드 값이 잘못됨 | JSON 파싱 실패, amount가 0 이하, currency 누락 | 요청 형식이 올바르지 않습니다. | 우선 사용 |
| 401 | `unauthorized` | 인증되지 않은 요청 | API key 또는 token이 없거나 유효하지 않음 | 인증 정보가 필요합니다. | 인증 도입 후 |
| 402 | `payment_required` | 결제가 필요한 요청 | 유료 기능, 수수료 결제 필요 정책이 생겼을 때 | 결제가 필요한 요청입니다. | 보류 |
| 403 | `forbidden` | 인증은 되었지만 권한이 없음 | 다른 merchant의 invoice에 접근 | 해당 리소스에 접근할 권한이 없습니다. | 권한 도입 후 |
| 404 | `not_found` | 대상 리소스를 찾을 수 없음 | merchant, invoice, payment 없음 | 요청한 리소스를 찾을 수 없습니다. | 우선 사용 |
| 405 | `method_not_allowed` | 지원하지 않는 HTTP method | `GET /invoices`만 지원하는데 `DELETE` 요청 | 지원하지 않는 HTTP 메서드입니다. | 필요 시 |
| 406 | `not_acceptable` | 요청한 응답 형식을 제공할 수 없음 | 지원하지 않는 `Accept` header | 요청한 응답 형식을 제공할 수 없습니다. | 보류 |
| 407 | `proxy_authentication_required` | proxy 인증 필요 | 프록시 서버 인증이 필요한 특수 환경 | 프록시 인증이 필요합니다. | 보류 |
| 408 | `request_timeout` | 요청 시간이 초과됨 | 클라이언트 요청 body 수신 지연 | 요청 시간이 초과되었습니다. | 필요 시 |
| 409 | `conflict` | 현재 상태와 요청이 충돌함 | 이미 finalized된 payment를 다시 finalized 처리 | 현재 상태에서는 요청을 처리할 수 없습니다. | 우선 사용 |
| 415 | `unsupported_media_type` | 지원하지 않는 요청 본문 형식 | `Content-Type: text/plain`으로 JSON API 호출 | 지원하지 않는 요청 형식입니다. | 우선 사용 |
| 422 | `unprocessable_entity` | 형식은 맞지만 도메인 규칙을 만족하지 못함 | currency는 문자열이지만 지원하지 않는 통화 | 요청값이 처리 가능한 조건을 만족하지 않습니다. | 필요 시 |

500번대 서버 오류는 400번대와 분리해서 관리한다.

| HTTP status | code | 의미 | 사용 예시 | message 예시 |
| --- | --- | --- | --- | --- |
| 500 | `internal_server_error` | 서버 내부 오류 | DB 오류, 예상하지 못한 실패 | 서버 내부 오류가 발생했습니다. |

초기 구현에서는 너무 많은 에러 코드를 한 번에 구현하지 않는다.

먼저 `bad_request`, `not_found`, `conflict`, `unsupported_media_type`, `internal_server_error` 정도로 시작하고, 인증/권한 기능이 들어올 때 `unauthorized`, `forbidden`을 추가하는 것이 적절하다.

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

- [x] 구현 시작 가능
- [ ] 부분 보강 후 시작
- [ ] 하루 더 설계 필요

판단 이유:

```text
공통 에러 응답의 방향, validation 책임 분리, config/logging/test pattern 후보가 정리되었으므로 Day8부터 구현 준비를 시작할 수 있다.
다만 error code는 한 번에 모두 구현하지 않고 REST 400번대 중 현재 API에 바로 필요한 code부터 작게 시작한다.
```

## 오늘의 회고

### 오늘 가장 잘 이해된 개념

Backend Core는 특정 도메인 기능이 아니라 이후 Ledger, Settlement, Indexer가 공통으로 사용할 실패 처리, 검증, 설정, 로그, 테스트 기반이라는 점을 이해했다.

### 아직 가장 약한 개념

HTTP status code와 우리 서비스의 error code를 어떻게 매핑할지, 그리고 어떤 에러를 handler에서 처리하고 어떤 에러를 service에서 처리할지 더 익숙해질 필요가 있다.

### 다음 구현에서 조심할 점

에러 코드를 너무 많이 만들기보다 실제 API에서 바로 필요한 code부터 구현한다.

`message`는 한글로 작성하되, `code`는 테스트와 클라이언트 분기를 위해 안정적인 영어 snake_case로 유지한다.
