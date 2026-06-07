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

| code | 사용하는 상황 |
| --- | --- |
| invalid_request |  |
| not_found |  |
| conflict |  |
| internal_error |  |

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
