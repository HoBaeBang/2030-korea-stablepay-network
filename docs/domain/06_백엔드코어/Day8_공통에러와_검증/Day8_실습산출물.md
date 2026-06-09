# Day 8 실습산출물 - 공통 에러 응답과 요청 검증

관련 Jira: [SPN-25](https://aslan0.atlassian.net/browse/SPN-25)

## 1. 현재 API 에러 응답 위치

```text
merchant API:
- 파일: internal/httpapi/merchant.go
- 위치: handlerCreateMerchant
- 현재 방식: errorResponse{Error: "..."} 형태로 문자열 에러를 반환한다.
- 주요 에러: invalid json body, name/email 필수값 누락, email 형식 오류, 중복 email, internal server error

invoice API:
- 파일: internal/httpapi/invoice.go
- 위치: handleCreateInvoice
- 현재 방식: errorResponse{Error: "..."} 형태로 문자열 에러를 반환한다.
- 주요 에러: merchant id 누락, invalid json body, expires_at 형식 오류, merchant not found, amount/currency 검증 오류, internal server error

payment API:
- 파일: internal/httpapi/payment.go
- 위치: handleCreatePayment, handleUpdatePaymentStatus
- 현재 방식: errorResponse{Error: "..."} 형태로 문자열 에러를 반환한다.
- 주요 에러: invoice/payment id 누락, invalid json body, invoice/payment not found, 잘못된 payment status, 잘못된 상태 전이, transaction_hash 누락, internal server error
```

## 2. 공통 에러 응답 초안

```json
{
  "error": {
    "code": "bad_request",
    "message": "결제 금액은 0보다 커야 합니다.",
    "field": "amount"
  }
}
```

`code`는 클라이언트와 테스트가 안정적으로 분기하기 위한 값이므로 영어 snake_case로 유지한다.

`message`는 사람이 읽는 설명이므로 한글로 작성한다.

`field`는 특정 요청 필드가 잘못되었을 때만 사용하고, 서버 내부 오류처럼 특정 필드와 직접 연결되지 않는 경우에는 생략할 수 있다.

## 3. error code 후보

Day7 실습산출물에서 정리한 후보를 기준으로 작성한다.

| HTTP status | code | 의미 | 사용하는 상황 | 초기 구현 여부 |
| --- | --- | --- | --- | --- |
| 400 | `bad_request` | 요청 형식이나 필드 값이 잘못됨 | JSON 파싱 실패, 필수 필드 누락, amount가 0 이하, expires_at 형식 오류 | 우선 사용 |
| 401 | `unauthorized` | 인증되지 않은 요청 | API key 또는 token이 없거나 유효하지 않은 경우 | 인증 도입 후 |
| 402 | `payment_required` | 결제가 필요한 요청 | 유료 기능, 수수료 결제 필요 정책이 생겼을 때 | 보류 |
| 403 | `forbidden` | 인증은 되었지만 권한이 없음 | 다른 merchant의 invoice/payment에 접근하려는 경우 | 권한 도입 후 |
| 404 | `not_found` | 대상 리소스를 찾을 수 없음 | merchant, invoice, payment가 존재하지 않는 경우 | 우선 사용 |
| 405 | `method_not_allowed` | 지원하지 않는 HTTP method | `POST`만 지원하는 endpoint에 `GET` 또는 `DELETE`로 요청한 경우 | 필요 시 |
| 406 | `not_acceptable` | 요청한 응답 형식을 제공할 수 없음 | 지원하지 않는 `Accept` header를 보낸 경우 | 보류 |
| 407 | `proxy_authentication_required` | proxy 인증 필요 | 프록시 서버 인증이 필요한 특수 환경 | 보류 |
| 408 | `request_timeout` | 요청 시간이 초과됨 | 클라이언트 요청 body 수신이 너무 오래 걸린 경우 | 필요 시 |
| 409 | `conflict` | 현재 상태와 요청이 충돌함 | 이미 finalized된 payment를 다시 on_chain_detected로 바꾸려는 경우 | 우선 사용 |
| 415 | `unsupported_media_type` | 지원하지 않는 요청 본문 형식 | JSON API인데 `Content-Type: text/plain`으로 요청한 경우 | 우선 사용 |
| 422 | `unprocessable_entity` | 형식은 맞지만 도메인 규칙을 만족하지 못함 | currency 형식은 문자열이지만 지원하지 않는 통화인 경우 | 필요 시 |
| 500 | `internal_server_error` | 서버 내부 오류 | DB 오류, 예상하지 못한 서버 내부 실패 | 우선 사용 |

초기 구현에서 먼저 사용할 코드:

```text
bad_request
not_found
conflict
unsupported_media_type
internal_server_error
```

## 4. validation 책임 분리

| API | Handler validation | Service validation | Repository 책임 |
| --- | --- | --- | --- |
| Create merchant | JSON 파싱, name/email 필수값, email 기본 형식 확인 | name/email 정규화, 중복 email 같은 도메인 오류를 처리 가능한 에러로 전달 | merchant 저장, unique constraint 결과 전달 |
| Create invoice | merchantId path variable 확인, JSON 파싱, expires_at RFC3339 형식 확인 | merchant 존재 여부, amount > 0, currency 지원 여부, invoice 초기 상태 결정 | invoice 저장, foreign key/DB 오류 전달 |
| Detect payment | paymentId path variable 확인, JSON 파싱, status/transaction_hash 입력 형식 확인 | payment 존재 여부, 상태 전이 가능 여부, ONCHAIN_DETECTED일 때 transaction_hash 필수 여부 | payment 조회와 상태 업데이트 |
| Finalize payment | paymentId path variable 확인, JSON 파싱, status 입력 형식 확인 | payment 존재 여부, ONCHAIN_DETECTED -> FINALIZED 상태 전이 가능 여부, finalized_at 기록 여부 판단 | payment 조회와 상태 업데이트 |

## 5. 다음 구현 후보

```text
1. internal/httpapi/response.go 추가
   - 공통 성공 응답과 에러 응답을 한 파일에서 관리한다.

2. writeError(w, status, code, message, field) 함수 추가
   - 기존 errorResponse{Error: "..."} 대신 {"error": {"code": "...", "message": "...", "field": "..."}} 구조를 사용한다.

3. service error를 handler에서 공통 error code로 변환하는 규칙 정리
   - 예: ErrMerchantNotFound -> 404 not_found
   - 예: amount must be greater than zero -> 400 bad_request
   - 예: invalid payment status transition -> 409 conflict 또는 400 bad_request 중 정책 결정
```

## 6. 오늘 헷갈린 개념

```text
- 404 not_found는 API 경로가 없다는 뜻이 아니라, 보통 요청한 merchant/invoice/payment 같은 리소스가 없다는 뜻이다.
- bad_request와 unprocessable_entity는 비슷해 보이지만, 초기 구현에서는 bad_request로 작게 시작하는 것이 좋다.
- Repository는 검증을 전혀 하지 않는 곳이라기보다, 도메인 검증 책임을 갖지 않고 DB 저장/조회와 DB constraint 결과 전달을 맡는 곳이다.
```

## 7. 오늘의 결론

```text
Day 8을 통해 내가 이해한 Backend Core의 역할은 API 실패를 일관된 구조로 표현하고,
요청 형식 검증과 도메인 규칙 검증의 책임을 나누어 이후 Ledger, Deposit, Withdrawal,
Settlement 같은 기능이 추가되어도 안정적으로 확장할 수 있는 기반을 만드는 것이다.
```
