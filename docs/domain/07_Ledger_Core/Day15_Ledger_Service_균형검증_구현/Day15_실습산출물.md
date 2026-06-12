# Day 15 실습산출물 - Ledger Service 균형 검증 구현

관련 Jira: SPN-32

Day15 산출물은 5개 질문만 작성합니다.

오늘 산출물은 “내가 Ledger 전체를 다 이해했는지”를 묻는 문서가 아닙니다.

오늘 만든 코드 하나를 정확히 이해했는지 확인하는 문서입니다.

## 작성 전 확인

아래 파일을 먼저 확인합니다.

```bash
ls internal/ledger
```

아래 파일이 보여야 합니다.

```text
ledger.go
service.go
service_test.go
```

## 1. `ValidateTransaction`은 어떤 책임을 가지는가?

작성할 때 볼 파일:

```text
internal/ledger/service.go
```

작성 힌트:

```text
DB 저장을 하는 함수인지,
저장 전에 규칙을 검증하는 함수인지 구분해서 적는다.
```

내 답변:

```text

```

## 2. debit과 credit 합계가 맞지 않으면 왜 실패해야 하는가?

작성할 때 참고할 흐름:

```text
고객 DEBIT 10_000_000
가맹점 CREDIT 9_800_000
플랫폼 CREDIT 200_000
```

작성 힌트:

```text
돈이 생기거나 사라진 것처럼 보이는 문제,
정산과 장애 복구가 틀어지는 문제를 연결해서 적는다.
```

내 답변:

```text

```

## 3. `map[string]int64`는 오늘 코드에서 어떤 역할을 하는가?

작성할 때 볼 코드:

```go
totals := make(map[string]int64)
```

작성 힌트:

```text
string key가 무엇인지,
int64 value가 무엇인지,
왜 통화별로 나누어 계산하는지 적는다.
```

내 답변:

```text

```

## 4. 오늘 테스트 중 가장 중요하다고 생각하는 실패 케이스는 무엇인가?

작성할 때 볼 파일:

```text
internal/ledger/service_test.go
```

작성 힌트:

```text
성공 케이스보다 실패 케이스가 왜 중요한지,
그 테스트가 어떤 버그를 막는지 적는다.
```

내 답변:

```text

```

## 5. 다음 단계에서 Repository와 DB가 필요한 이유는 무엇인가?

작성 힌트:

```text
오늘 Service는 검증만 한다.
검증된 Ledger Transaction을 영구 보관하려면 무엇이 필요한지 적는다.
```

내 답변:

```text

```

## 오늘 실행 결과

아래 명령 실행 결과를 짧게 기록합니다.

```bash
go test ./internal/ledger -v
go test ./...
```

기록:

```text

```

## 아직 헷갈리는 부분

아래 후보 중 헷갈리는 것이 있으면 적습니다.

```text
receiver
context
map
int64
debit/credit
t.Run
err == nil
```

메모:

```text

```
