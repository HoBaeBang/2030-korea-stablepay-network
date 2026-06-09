# Day 8 개념학습 - 공통 에러 응답과 요청 검증

관련 Jira: [SPN-25](https://aslan0.atlassian.net/browse/SPN-25)

## 공통 에러 응답이란?

공통 에러 응답은 API가 실패했을 때 항상 같은 구조로 실패 이유를 알려주는 방식입니다.

예시:

```json
{
  "error": {
    "code": "bad_request",
    "message": "결제 금액은 0보다 커야 합니다.",
    "field": "amount"
  }
}
```

여기서 각 값의 의미는 다음과 같습니다.

| 항목 | 의미 |
| --- | --- |
| code | 프로그램이 구분하기 쉬운 에러 종류 |
| message | 사람이 읽을 수 있는 설명 |
| field | 문제가 발생한 요청 필드 |

## 왜 message만으로 부족한가?

`message`는 사람이 읽기 좋지만, 프로그램이 안정적으로 판단하기 어렵습니다.

예를 들어 `결제 금액은 0보다 커야 합니다.`라는 문장은 나중에 `결제 금액을 다시 확인해주세요.`로 바뀔 수 있습니다.

반면 `bad_request` 같은 code는 클라이언트가 안정적으로 분기할 수 있습니다.

## validation이란?

validation은 요청값이 처리 가능한 값인지 확인하는 과정입니다.

예시:

```text
merchant name이 비어 있지 않은가?
invoice amount가 0보다 큰가?
currency가 지원하는 통화인가?
payment status 전이가 허용되는가?
```

## handler validation

handler는 HTTP 요청의 형식을 가장 먼저 만나는 곳입니다.

handler에서 확인하기 좋은 것:

| 검증 | 예시 |
| --- | --- |
| JSON 파싱 | 요청 body가 JSON 형식인가? |
| path variable | merchant id가 비어 있지 않은가? |
| 필수 필드 | amount가 요청에 포함되어 있는가? |
| 기본 타입 | amount가 숫자인가? |

## service validation

service는 도메인 규칙을 판단하는 곳입니다.

service에서 확인하기 좋은 것:

| 검증 | 예시 |
| --- | --- |
| 도메인 규칙 | invoice amount는 0보다 커야 한다. |
| 상태 전이 | pending payment만 on_chain_detected로 바꿀 수 있다. |
| 정책 | 지원 통화는 USDC만 허용한다. |

## repository validation

repository는 검증을 많이 넣는 곳이 아닙니다.

repository의 주 책임은 DB 저장과 조회입니다.

물론 DB constraint 때문에 에러가 발생할 수 있지만, 요청 검증을 repository에 몰아넣으면 service가 어떤 규칙을 지키는지 알기 어려워집니다.

## 책임 분리 그림

![Day8 공통 에러 응답과 검증 흐름](../../../confluence/diagrams/spn25-day8-error-validation-flow.png)

이 그림을 볼 때는 화살표보다 “책임 경계”를 먼저 봐야 합니다.

| 계층 | 봐야 하는 것 | 보면 안 되는 것 |
| --- | --- | --- |
| Handler | HTTP 요청 형식, JSON 파싱, path/query/body 기본 검증 | DB table 구조, 복잡한 도메인 정책 |
| Service | 도메인 규칙, 상태 전이, 비즈니스 정책 | HTTP status code 세부 선택, JSON 응답 모양 |
| Repository | SQL 실행, DB constraint 결과 전달 | 요청 field 검증, payment 상태 전이 정책 판단 |

실제 구현에서는 Service에서 발생한 error를 Handler가 받아서 HTTP status와 error code로 바꾸게 됩니다.

예를 들어 `payment.ErrPaymentNotFound`가 Service에서 올라오면 Handler는 이것을 `404 not_found`로 바꾸고, `invalid payment status transition`은 `409 conflict` 또는 정책에 따라 `400 bad_request`로 바꿀 수 있습니다.

## Day7에서 정리한 error code 후보

Day7 실습산출물에서는 400번대 REST Client Error를 중심으로 후보를 정리했습니다.

Day8에서는 이 후보를 바탕으로 실제 구현에 먼저 사용할 코드를 고릅니다.

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

500번대 서버 오류는 400번대와 분리해서 관리합니다.

| HTTP status | code | 의미 | 사용 예시 | message 예시 | 초기 구현 여부 |
| --- | --- | --- | --- | --- | --- |
| 500 | `internal_server_error` | 서버 내부 오류 | DB 오류, 예상하지 못한 실패 | 서버 내부 오류가 발생했습니다. | 우선 사용 |

## 우리 프로젝트에서 먼저 구현할 error code

처음 구현에서는 모든 후보를 한 번에 넣지 않습니다.

먼저 아래 코드만 구현하고, 인증/권한 기능이 들어올 때 `unauthorized`, `forbidden`을 추가합니다.

| code | 사용하는 상황 |
| --- | --- |
| `bad_request` | JSON 파싱 실패, 필수 필드 누락, amount가 0 이하인 경우 |
| `not_found` | merchant, invoice, payment를 찾을 수 없는 경우 |
| `conflict` | 이미 처리된 payment를 다시 처리하는 것처럼 현재 상태와 충돌하는 경우 |
| `unsupported_media_type` | JSON API인데 `Content-Type`이 JSON이 아닌 경우 |
| `internal_server_error` | DB 오류처럼 클라이언트가 고칠 수 없는 서버 내부 오류 |

## 오늘의 주의점

처음부터 너무 많은 error code를 만들 필요는 없습니다.

중요한 것은 모든 API가 같은 구조의 실패 응답을 사용하도록 만드는 것입니다.
