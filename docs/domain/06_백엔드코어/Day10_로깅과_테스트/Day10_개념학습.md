# Day 10 개념학습 - 로깅과 테스트 패턴

관련 Jira: [SPN-27](https://aslan0.atlassian.net/browse/SPN-27)

## 로그란?

로그는 애플리케이션이 실행되는 동안 발생한 중요한 일을 기록하는 것입니다.

예시:

```text
payment status changed
ledger transaction created
indexer checkpoint advanced
withdrawal broadcasted
```

## 로그와 테스트의 전체 흐름

![Day10 로깅과 테스트 피드백 루프](../../../confluence/diagrams/spn27-day10-logging-test-feedback.png)

Day10에서 로그와 테스트를 같이 보는 이유는 둘 다 “신뢰할 수 있는 백엔드”를 만드는 도구이기 때문입니다.

로그는 운영 중인 시스템에서 일어난 사실을 남깁니다.

테스트는 코드가 지켜야 할 규칙을 자동으로 검증합니다.

예를 들어 payment가 `PENDING -> ONCHAIN_DETECTED -> FINALIZED`로 이동하는 규칙은 테스트로 검증해야 하고, 실제 운영 중 어떤 payment가 언제 어떤 상태로 바뀌었는지는 로그로 남겨야 합니다.

## 로그와 에러 응답의 차이

| 구분 | 대상 | 목적 |
| --- | --- | --- |
| 에러 응답 | API 사용자 | 요청이 왜 실패했는지 알려준다. |
| 로그 | 개발자/운영자 | 서버 내부에서 무슨 일이 있었는지 추적한다. |

사용자에게 모든 내부 정보를 보여주면 안 됩니다.

예를 들어 DB 에러의 상세 SQL을 API 응답으로 그대로 주면 보안상 위험할 수 있습니다.

하지만 로그에는 장애 분석을 위해 더 자세한 정보를 남길 수 있습니다.

## 어떤 로그가 중요한가?

돈과 상태가 바뀌는 순간이 중요합니다.

| 이벤트 | 로그가 필요한 이유 |
| --- | --- |
| payment status changed | 결제 상태 흐름 추적 |
| ledger transaction created | 돈의 이동 기록 생성 추적 |
| duplicate event ignored | 중복 이벤트 방어 확인 |
| settlement batch created | 정산 묶음 생성 추적 |
| withdrawal signed | 서명 완료 추적 |
| withdrawal broadcasted | 온체인 전송 추적 |

## 로그에 남길 값과 남기면 안 되는 값

로그는 많이 남긴다고 좋은 것이 아닙니다.

장애를 추적할 수 있는 식별자는 남기고, 자산 탈취나 개인정보 유출로 이어질 수 있는 값은 절대 남기지 않아야 합니다.

| 구분 | 예시 | 이유 |
| --- | --- | --- |
| 남길 수 있는 값 | `payment_id`, `invoice_id`, `merchant_id`, `tx_hash`, `old_status`, `new_status` | 장애 추적과 상태 흐름 파악에 필요 |
| 주의해서 남길 값 | email, wallet address, request body 일부 | 개인정보/민감정보 정책에 따라 마스킹 필요 |
| 남기면 안 되는 값 | private key, seed phrase, access token, database password | 노출 시 자산이나 시스템 접근 권한이 탈취될 수 있음 |

좋은 로그는 사람이 읽기 좋은 문장과 프로그램이 필터링할 수 있는 필드를 같이 가집니다.

```text
payment status changed payment_id=pay_123 old_status=PENDING new_status=ONCHAIN_DETECTED tx_hash=0xabc
```

반대로 아래처럼 내부 에러를 그대로 사용자 응답으로 내보내거나 로그에 비밀값을 섞는 것은 피해야 합니다.

```text
database password=...
private_key=...
full_access_token=...
```

## Go 테스트 구조

Go 테스트 파일은 보통 `_test.go`로 끝납니다.

예시:

```go
func TestCreateInvoice(t *testing.T) {
    t.Run("금액이 0이면 invoice를 생성할 수 없다", func(t *testing.T) {
        // given
        // when
        // then
    })
}
```

## 한글 subtest를 쓰는 이유

테스트 이름이 한글이면 테스트가 문서처럼 읽힙니다.

특히 지금처럼 학습과 포트폴리오를 동시에 진행할 때는 장점이 큽니다.

```text
지원하지 않는 통화이면 invoice를 생성할 수 없다
이미 완료된 payment는 다시 finalized 처리할 수 없다
중복 이벤트는 ledger에 두 번 반영되지 않는다
```

## given / when / then

| 단계 | 의미 |
| --- | --- |
| given | 어떤 상황이 주어졌는가 |
| when | 어떤 행동을 실행하는가 |
| then | 어떤 결과를 기대하는가 |

이 구조는 테스트를 읽는 사람이 의도를 빠르게 이해하게 해줍니다.

## 어떤 테스트를 먼저 작성할까?

Phase 2 Backend Core에서는 다음 테스트 후보가 중요합니다.

| 영역 | 테스트 후보 |
| --- | --- |
| Error Response | 잘못된 JSON 요청이 `400 bad_request`를 반환한다 |
| Validation | amount가 0 이하이면 invoice/payment를 생성할 수 없다 |
| Status Transition | `FINALIZED` 상태 payment는 다시 `ONCHAIN_DETECTED`로 돌아갈 수 없다 |
| Config | `DATABASE_URL`이 없으면 config 로딩이 실패한다 |
| Ledger 준비 | debit/credit 합계가 0이 아니면 ledger transaction을 만들 수 없다 |

처음부터 모든 계층 테스트를 완벽하게 만들 필요는 없습니다.

우선 service test로 도메인 규칙을 고정하고, handler test로 HTTP 응답 형식을 확인하는 방향이 현실적입니다.
