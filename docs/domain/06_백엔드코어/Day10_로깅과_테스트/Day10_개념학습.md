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
