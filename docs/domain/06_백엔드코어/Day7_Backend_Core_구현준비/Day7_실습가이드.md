# Backend Core 실습 가이드

관련 Jira: [SPN-24](https://aslan0.atlassian.net/browse/SPN-24)

이 문서는 Day 7 퇴근 후 작업을 위한 실습 가이드입니다.

오늘의 실습은 코드를 바로 크게 고치는 것이 아니라, Sprint 2에서 구현할 Backend Core의 작업 범위를 문서로 구체화하는 것입니다.

## 실습 목표

`docs/domain/06_백엔드코어/Day7_Backend_Core_구현준비/Day7_실습산출물.md`를 채우면서 다음 내용을 정리합니다.

1. 현재 코드에서 공통화가 필요한 부분
2. 공통 에러 응답 초안
3. validation 책임 위치
4. config/logging/test pattern 정리 방향
5. Sprint 2 첫 구현 티켓 후보

## Step 1. 현재 코드 구조 다시 보기

아래 폴더를 확인합니다.

```text
cmd/api
internal/httpapi
internal/merchant
internal/invoice
internal/payment
internal/platform/database
migrations
api
```

확인할 질문:

```text
HTTP handler는 어디에 있는가?
service는 어디에 있는가?
repository는 어디에 있는가?
에러 응답은 어디에서 만들어지는가?
request validation은 어디에서 처리되는가?
```

## Step 2. 공통 에러 응답 후보 작성

현재 API에서 실패가 발생했을 때 어떤 JSON을 반환하면 좋을지 초안을 작성합니다.

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

실습산출물에는 다음을 적습니다.

```text
우리 프로젝트에서 우선 사용할 error code 후보
각 code가 어떤 상황에서 쓰이는지
handler에서 어떻게 응답할지
```

## Step 3. Validation 위치 정리

다음 기준으로 validation 위치를 나눠봅니다.

| 위치 | 책임 |
| --- | --- |
| Handler | JSON 파싱, path variable, 필수 필드 같은 요청 형식 검증 |
| Service | 도메인 규칙 검증 |
| Repository | DB 저장/조회, DB constraint에 집중 |

실습산출물에는 `Invoice 생성`, `Payment 상태 변경`, `미래 Ledger entry 생성` 기준으로 각각 어떤 검증이 어디에 있어야 하는지 적습니다.

## Step 4. Config 후보 작성

Phase 2에서 필요할 설정 후보를 적습니다.

예시:

```text
PORT
DATABASE_URL
LOG_LEVEL
BLOCKCHAIN_RPC_URL
INDEXER_POLL_INTERVAL
FINALITY_CONFIRMATIONS
SIGNER_BASE_URL
```

오늘은 모든 설정을 구현하지 않아도 됩니다.

다만 어떤 설정이 나중에 필요해질지 미리 정리합니다.

## Step 5. Logging 후보 작성

나중에 장애가 났을 때 추적하고 싶은 이벤트를 적습니다.

예시:

```text
payment status changed
ledger transaction created
ledger entry duplicated
settlement batch created
indexer checkpoint advanced
withdrawal signed
withdrawal broadcasted
```

## Step 6. Test Pattern 정리

테스트 이름은 가능하면 한글 subtest로 작성합니다.

예시:

```go
t.Run("지원하지 않는 통화이면 invoice를 생성할 수 없다", func(t *testing.T) {
    // ...
})
```

실습산출물에는 앞으로 반복할 테스트 패턴을 적습니다.

```text
given: 어떤 데이터가 주어졌는가
when: 어떤 동작을 실행했는가
then: 어떤 결과를 기대하는가
```

## Step 7. Sprint 2 첫 구현 후보 정리

Day 6에서 정리한 다음 작업 후보를 더 구체화합니다.

| 후보 | Day 7에서 해야 할 판단 |
| --- | --- |
| Backend Core migration 설계 | 공통 에러/validation/config/logging/test 중 무엇부터 구현할지 결정 |
| Ledger account/entry 모델 설계 | Backend Core 이후 바로 들어갈 수 있게 최소 모델 후보 작성 |
| Payment finalized 이후 ledger 연결 | 어떤 상태 전이에서 ledger를 만들지 후보 작성 |
| Settlement skeleton 작성 | merchant settlement로 시작할지, 더 일반적인 account settlement로 시작할지 판단 |

## 완료 기준

- [ ] 실습산출물에 공통 에러 응답 초안을 작성했다.
- [ ] validation 위치를 handler/service/repository로 나눴다.
- [ ] config 후보를 작성했다.
- [ ] logging 후보를 작성했다.
- [ ] 테스트 패턴을 작성했다.
- [ ] Sprint 2 첫 구현 후보를 정리했다.

## 권장 커밋 메시지

```bash
git commit -m "docs: Day7 Backend Core 학습자료와 실습산출물 정리"
```
