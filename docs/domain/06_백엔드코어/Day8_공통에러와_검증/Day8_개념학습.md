# Day 8 개념학습 - 공통 에러 응답과 요청 검증

관련 Jira: [SPN-25](https://aslan0.atlassian.net/browse/SPN-25)

## 공통 에러 응답이란?

공통 에러 응답은 API가 실패했을 때 항상 같은 구조로 실패 이유를 알려주는 방식입니다.

예시:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "amount must be greater than zero",
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

예를 들어 `amount must be greater than zero`라는 문장은 나중에 `amount should be positive`로 바뀔 수 있습니다.

반면 `invalid_request` 같은 code는 클라이언트가 안정적으로 분기할 수 있습니다.

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

```text
Handler
  - HTTP 요청 형식 확인
  - JSON 파싱
  - path/query/body 기본 검증

Service
  - 도메인 규칙 확인
  - 상태 전이 확인
  - 비즈니스 정책 판단

Repository
  - SQL 실행
  - DB constraint 결과 처리
```

## 우리 프로젝트에서 먼저 쓸 error code 후보

| code | 의미 |
| --- | --- |
| invalid_request | 요청 형식이나 값이 잘못됨 |
| not_found | 대상 리소스를 찾을 수 없음 |
| conflict | 이미 처리되었거나 상태가 맞지 않음 |
| internal_error | 서버 내부 오류 |

## 오늘의 주의점

처음부터 너무 많은 error code를 만들 필요는 없습니다.

중요한 것은 모든 API가 같은 구조의 실패 응답을 사용하도록 만드는 것입니다.
