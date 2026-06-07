# Backend Core 개념 학습

관련 Jira: [SPN-24](https://aslan0.atlassian.net/browse/SPN-24)

이 문서는 출퇴근 시간에 읽는 Day 7 개념 학습자료입니다.

Day 7에서는 블록체인 도메인 자체보다, 그 도메인을 구현하기 위한 Go 백엔드 기반을 정리합니다.

## 1. Backend Core란 무엇인가

`Backend Core`는 특정 도메인 하나를 의미하지 않습니다.

우리 프로젝트에서 Backend Core는 다음 기능들이 공통으로 기대는 기반입니다.

```text
Merchant
Invoice
Payment
Ledger
Settlement
Deposit
Withdrawal
Event Indexer
```

이 도메인들은 서로 다르지만, 서버 입장에서 공통으로 필요한 것들이 있습니다.

```text
요청을 받는다.
요청값을 검증한다.
실패하면 일관된 에러를 반환한다.
설정을 읽는다.
DB transaction을 다룬다.
중요한 상태 변경을 로그로 남긴다.
테스트로 검증한다.
```

이 공통 기반을 정리하는 것이 Day 7의 Backend Core입니다.

## 2. 왜 Backend Core를 먼저 정리하는가

Ledger나 Indexer를 먼저 만들 수도 있습니다.

하지만 공통 기반이 정리되지 않은 상태에서 도메인을 계속 추가하면 다음 문제가 생깁니다.

| 문제 | 예시 |
| --- | --- |
| 에러 응답이 제각각이다 | 어떤 API는 plain text, 어떤 API는 JSON error를 반환한다. |
| validation 위치가 흔들린다 | handler, service, repository 여기저기에 검증 로직이 흩어진다. |
| 설정이 흩어진다 | DB URL, RPC URL, 포트 설정이 여러 파일에 섞인다. |
| 로그가 부족하다 | payment 상태 변경이나 ledger 생성 실패 원인을 추적하기 어렵다. |
| 테스트 패턴이 반복되지 않는다 | 새 기능을 붙일 때마다 테스트 구조를 새로 고민한다. |

Phase 2는 돈과 관련된 기능을 다룹니다.

그래서 작은 실수도 중복 결제, 잘못된 ledger entry, 정산 불일치로 이어질 수 있습니다.

```text
도메인 구현 속도보다
일관된 실패 처리와 검증 구조가 먼저다.
```

## 3. 공통 에러 응답

현재 API에서 실패가 발생하면 사용자는 다음을 알고 싶어 합니다.

```text
무엇이 실패했는가?
왜 실패했는가?
내가 다시 요청해도 되는가?
어떤 값을 고쳐야 하는가?
```

그래서 에러 응답은 일정한 형태를 가지는 것이 좋습니다.

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

우리 프로젝트에서는 처음부터 복잡한 에러 시스템을 만들 필요는 없습니다.

하지만 최소한 다음 정도는 정리하면 좋습니다.

| 항목 | 의미 |
| --- | --- |
| `code` | 에러 종류를 구분하는 짧은 값 |
| `message` | 사람이 읽을 수 있는 설명 |
| `field` | 특정 요청 필드가 잘못된 경우 필드명 |

## 4. Validation

`Validation`은 요청값이 유효한지 확인하는 작업입니다.

예를 들어 invoice 생성 요청에는 다음 검증이 필요할 수 있습니다.

```text
amount > 0
currency is supported
merchant_id is not empty
```

중요한 것은 검증 위치입니다.

```text
HTTP Handler:
  JSON body를 읽고 기본 형식을 검증한다.

Service:
  도메인 규칙을 검증한다.

Repository:
  DB 저장과 조회에 집중한다.
```

Repository는 가능하면 요청 검증의 중심이 되지 않는 것이 좋습니다.

Repository까지 잘못된 값이 내려오면, 도메인 규칙과 DB 제약이 뒤섞이기 쉽습니다.

## 5. Config

`Config`는 서버가 실행될 때 필요한 설정입니다.

예시:

```text
PORT
DATABASE_URL
LOG_LEVEL
BLOCKCHAIN_RPC_URL
FINALITY_CONFIRMATIONS
```

Phase 1에서는 설정이 많지 않습니다.

하지만 Phase 2로 가면 RPC URL, indexer polling interval, confirmation 기준, signer URL 같은 설정이 늘어납니다.

그래서 설정을 구조체로 모아두는 것이 좋습니다.

```go
type Config struct {
    Port        string
    DatabaseURL string
    LogLevel    string
}
```

## 6. Logging

`Logging`은 서버가 어떤 일을 했는지 기록하는 것입니다.

블록체인 결제 백엔드에서는 특히 다음 로그가 중요합니다.

```text
payment 상태 변경
ledger transaction 생성
settlement batch 생성
indexer checkpoint 변경
withdrawal signing 요청
broadcast 결과
```

로그가 없으면 장애가 났을 때 이런 질문에 답하기 어렵습니다.

```text
이 payment는 왜 FINALIZED가 되었는가?
ledger entry는 생성됐는가?
같은 event가 두 번 처리됐는가?
어느 block까지 indexer가 처리했는가?
```

## 7. Test Pattern

Go 테스트에서는 보통 `_test.go` 파일을 만들고 `go test`로 실행합니다.

우리 프로젝트에서는 한글 subtest를 사용하면 학습과 가독성에 도움이 됩니다.

```go
func TestCreatePayment(t *testing.T) {
    t.Run("금액이 0 이하이면 결제를 생성할 수 없다", func(t *testing.T) {
        // given
        // when
        // then
    })
}
```

테스트는 기능이 맞는지 확인하는 도구이기도 하지만, 문서 역할도 합니다.

나중에 채용 담당자나 동료가 코드를 볼 때, 테스트 이름만 보고도 의도를 이해할 수 있습니다.

## 8. Day 7에서 바로 구현하지 않는 것

Day 7은 아직 Ledger 전체 구현일이 아닙니다.

다음은 오늘의 직접 범위가 아닙니다.

```text
ledger_entries 전체 구현
settlement batch 생성
blockchain RPC 연결
Rust signer 구현
실제 withdrawal broadcast
```

오늘은 위 기능들이 들어오기 전에 필요한 기반을 정리합니다.

## 9. 오늘 기억할 요약

```text
Backend Core는 특정 기능 하나가 아니라,
모든 기능이 공통으로 기대는 실패 처리, 검증, 설정, 로그, 테스트의 기반이다.
```

Day 7을 잘 정리하면 Sprint 2에서 Ledger와 Settlement를 구현할 때 코드가 훨씬 덜 흔들립니다.
