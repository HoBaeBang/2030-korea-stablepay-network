# Day 8 검증문제와 답변가이드

관련 Jira: [SPN-25](https://aslan0.atlassian.net/browse/SPN-25)

## 먼저 풀어볼 문제

1. 공통 에러 응답이 필요한 이유는 무엇인가?
2. `message`와 `code`는 각각 어떤 역할을 하는가?
3. handler validation과 service validation의 차이는 무엇인가?
4. repository에 validation을 많이 넣으면 어떤 문제가 생길 수 있는가?
5. `bad_request`와 `conflict`는 어떤 차이가 있는가?
6. field 값은 언제 필요하고 언제 없어도 되는가?
7. payment 상태 전이 검증은 handler와 service 중 어디가 더 적절한가?
8. 공통 에러 응답을 만들 때 처음부터 많은 error code가 필요하지 않은 이유는 무엇인가?

## 답변가이드

### 1. 공통 에러 응답이 필요한 이유

API마다 실패 응답 형식이 다르면 클라이언트가 매번 다른 방식으로 에러를 처리해야 합니다.

공통 에러 응답을 사용하면 실패를 예측 가능한 구조로 표현할 수 있습니다.

### 2. message와 code의 역할

`message`는 사람이 읽는 설명입니다.

`code`는 프로그램이 안정적으로 분기하기 위한 값입니다.

### 3. handler validation과 service validation

handler validation은 HTTP 요청 형식에 가깝습니다.

service validation은 도메인 규칙에 가깝습니다.

### 4. repository에 validation을 많이 넣을 때의 문제

repository는 DB 접근 책임을 가져야 합니다.

검증 규칙이 repository에 섞이면 service의 도메인 정책이 흐려지고 테스트도 어려워집니다.

### 5. bad_request와 conflict의 차이

`bad_request`는 요청 형식이나 요청값 자체가 잘못된 경우입니다.

`conflict`는 요청값 형식은 맞지만 현재 상태와 충돌하는 경우입니다.

예시:

```text
amount = -1            -> bad_request
이미 finalized인 payment를 다시 finalized 처리 -> conflict
```

### 6. field 값이 필요한 경우

특정 요청 필드가 잘못되었을 때 사용합니다.

서버 내부 오류처럼 특정 필드와 관련 없는 경우에는 생략할 수 있습니다.

### 7. payment 상태 전이 검증 위치

service가 더 적절합니다.

상태 전이는 HTTP 형식이 아니라 도메인 규칙이기 때문입니다.

### 8. error code를 처음부터 많이 만들 필요가 없는 이유

초기에 너무 많은 code를 만들면 관리가 어려워집니다.

먼저 `bad_request`, `not_found`, `conflict`, `unsupported_media_type`, `internal_server_error` 정도로 시작하고, 실제 필요가 생기면 늘리는 것이 좋습니다.
