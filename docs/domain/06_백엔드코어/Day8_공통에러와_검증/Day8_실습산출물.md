# Day 8 실습산출물 - 공통 에러 응답과 요청 검증

관련 Jira: [SPN-25](https://aslan0.atlassian.net/browse/SPN-25)

## 1. 현재 API 에러 응답 위치

작성할 내용:

```text
merchant API:
invoice API:
payment API:
```

## 2. 공통 에러 응답 초안

작성할 내용:

```json
{
  "error": {
    "code": "",
    "message": "",
    "field": ""
  }
}
```

## 3. error code 후보

Day7 실습산출물에서 정리한 후보를 기준으로 작성한다.

| HTTP status | code | 의미 | 사용하는 상황 | 초기 구현 여부 |
| --- | --- | --- | --- | --- |
| 400 | `bad_request` | 요청 형식이나 필드 값이 잘못됨 |  | 우선 사용 |
| 401 | `unauthorized` | 인증되지 않은 요청 |  | 인증 도입 후 |
| 402 | `payment_required` | 결제가 필요한 요청 |  | 보류 |
| 403 | `forbidden` | 인증은 되었지만 권한이 없음 |  | 권한 도입 후 |
| 404 | `not_found` | 대상 리소스를 찾을 수 없음 |  | 우선 사용 |
| 405 | `method_not_allowed` | 지원하지 않는 HTTP method |  | 필요 시 |
| 406 | `not_acceptable` | 요청한 응답 형식을 제공할 수 없음 |  | 보류 |
| 407 | `proxy_authentication_required` | proxy 인증 필요 |  | 보류 |
| 408 | `request_timeout` | 요청 시간이 초과됨 |  | 필요 시 |
| 409 | `conflict` | 현재 상태와 요청이 충돌함 |  | 우선 사용 |
| 415 | `unsupported_media_type` | 지원하지 않는 요청 본문 형식 |  | 우선 사용 |
| 422 | `unprocessable_entity` | 형식은 맞지만 도메인 규칙을 만족하지 못함 |  | 필요 시 |
| 500 | `internal_server_error` | 서버 내부 오류 |  | 우선 사용 |

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
| Create merchant |  |  |  |
| Create invoice |  |  |  |
| Detect payment |  |  |  |
| Finalize payment |  |  |  |

## 5. 다음 구현 후보

작성할 내용:

```text
1.
2.
3.
```

## 6. 오늘 헷갈린 개념

작성할 내용:

```text
-
-
-
```

## 7. 오늘의 결론

작성할 내용:

```text
Day 8을 통해 내가 이해한 Backend Core의 역할은 ...
```
