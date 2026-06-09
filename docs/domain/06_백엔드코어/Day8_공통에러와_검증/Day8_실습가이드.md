# Day 8 실습가이드 - 공통 에러 응답과 요청 검증

관련 Jira: [SPN-25](https://aslan0.atlassian.net/browse/SPN-25)

이 문서는 Day 8 퇴근 후 실습을 위한 가이드입니다.

## 실습 흐름

![Day8 공통 에러 응답과 검증 흐름](../../../confluence/diagrams/spn25-day8-error-validation-flow.png)

오늘 실습은 코드를 바로 고치는 것보다, 기존 코드에서 에러가 어디서 만들어지고 어떤 공통 구조로 바뀌어야 하는지 관찰하는 것이 우선입니다.

## 실습 목표

`Day8_실습산출물.md`에 다음 내용을 작성합니다.

1. 현재 API의 에러 응답 위치
2. 공통 에러 응답 JSON 초안
3. error code 후보
4. merchant, invoice, payment validation 후보
5. handler/service/repository 책임 분리

## Step 1. 현재 handler 코드 확인

다음 파일을 확인합니다.

```text
internal/httpapi/merchant.go
internal/httpapi/invoice.go
internal/httpapi/payment.go
```

확인할 질문:

```text
요청 body는 어디에서 읽는가?
JSON 파싱 실패는 어떻게 처리되는가?
service에서 error가 오면 handler는 어떻게 응답하는가?
각 API의 실패 응답 형식은 일관적인가?
```

## Step 2. 현재 service 코드 확인

다음 파일을 확인합니다.

```text
internal/merchant/service.go
internal/invoice/service.go
internal/payment/service.go
```

확인할 질문:

```text
비어 있는 name은 어디에서 막는가?
지원하지 않는 currency는 어디에서 막는가?
잘못된 payment status 전이는 어디에서 막는가?
```

## Step 3. 공통 에러 응답 초안 작성

다음 형식을 기준으로 우리 프로젝트에 맞는 초안을 작성합니다.

```json
{
  "error": {
    "code": "bad_request",
    "message": "결제 금액은 0보다 커야 합니다.",
    "field": "amount"
  }
}
```

고민할 점:

```text
field가 항상 필요한가?
여러 필드가 동시에 틀렸을 때는 어떻게 할 것인가?
처음에는 단일 field만 지원해도 되는가?
Day7에서 정리한 error code 후보 중 무엇을 먼저 구현할 것인가?
```

## Step 4. validation 후보 작성

아래 표를 실습산출물에 채웁니다.

| API | Handler validation | Service validation |
| --- | --- | --- |
| Create merchant |  |  |
| Create invoice |  |  |
| Mark payment on-chain detected |  |  |
| Finalize payment |  |  |

## Step 5. 구현 후보 정리

오늘 꼭 코드를 수정하지 않아도 됩니다.

다만 다음 구현 후보를 구체화합니다.

```text
internal/httpapi/response.go 추가
writeError(w, status, code, message, field) 함수 후보
도메인 error type 추가 여부
service error를 handler error response로 변환하는 방식
```

## 완료 기준

- [ ] 현재 API의 에러 응답 위치를 찾았다.
- [ ] 공통 에러 응답 JSON 초안을 작성했다.
- [ ] error code 후보를 작성했다.
- [ ] validation 책임을 나눴다.
- [ ] 다음 코드 수정 후보를 정리했다.
