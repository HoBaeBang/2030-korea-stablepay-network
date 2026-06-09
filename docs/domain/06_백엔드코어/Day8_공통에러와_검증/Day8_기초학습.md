# Day 8 기초학습 - 공통 에러 응답과 요청 검증

관련 Jira: [SPN-25](https://aslan0.atlassian.net/browse/SPN-25)

Day 8의 목표는 API 실패 응답과 요청 검증을 정리하는 것입니다.

Phase 2에서는 Ledger, Deposit, Withdrawal, Settlement처럼 돈과 상태가 연결된 기능이 많아집니다.

이때 API마다 에러 응답 형식이 다르면 클라이언트도 혼란스럽고, 나중에 장애를 추적하기도 어렵습니다.

## 오늘의 큰 그림

```text
Client Request
      |
      v
HTTP Handler
      |
      +--> JSON 파싱
      +--> path variable 확인
      +--> 기본 요청값 검증
      |
      v
Service
      |
      +--> 도메인 규칙 검증
      +--> 상태 변경 판단
      |
      v
Repository
      |
      +--> DB 저장/조회
```

## 오늘의 목표

1. 공통 에러 응답이 왜 필요한지 설명할 수 있다.
2. `bad_request`, `not_found`, `conflict`, `internal_server_error` 같은 error code 후보를 정리한다.
3. handler와 service의 validation 책임을 구분한다.
4. 기존 merchant, invoice, payment API에 어떤 에러 응답이 필요한지 찾는다.
5. 실습산출물에 우리 프로젝트의 에러 응답 정책 초안을 작성한다.

## 출퇴근 학습

출퇴근 시간에는 코드를 외우려고 하지 말고, 다음 질문에 답할 수 있게 읽습니다.

```text
왜 에러 메시지만 반환하면 부족할까?
왜 error code가 필요할까?
handler validation과 service validation은 어떻게 다를까?
왜 repository가 요청 검증까지 담당하면 안 될까?
```

## 퇴근 후 작업

퇴근 후에는 기존 코드를 보면서 다음을 정리합니다.

1. 현재 API에서 에러 응답이 만들어지는 위치를 찾는다.
2. 공통 에러 응답 JSON 초안을 작성한다.
3. merchant, invoice, payment 요청별 validation 후보를 적는다.
4. handler/service/repository 책임을 나눈다.

## 오늘 꼭 잡아야 하는 문장

```text
Validation은 사용자의 요청을 믿지 않기 위한 장치이고,
Error Response는 실패를 예측 가능한 언어로 표현하기 위한 장치다.
```

## 완료 기준

- [ ] 공통 에러 응답 형식을 설명했다.
- [ ] error code 후보를 작성했다.
- [ ] handler validation과 service validation을 구분했다.
- [ ] 실습산출물을 작성했다.
- [ ] 검증문제를 풀고 답변가이드와 비교했다.
