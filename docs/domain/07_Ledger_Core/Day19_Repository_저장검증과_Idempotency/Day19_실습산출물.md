# Day 19 실습산출물 - Repository 저장 검증과 Idempotency

관련 Jira: SPN-36

Day19 산출물은 `CreateTransaction` 저장 메서드를 어떤 관점에서 검증했는지 정리하는 문서입니다.

정답처럼 길게 쓰는 것보다, 아래 세 가지를 자기 말로 설명하는 것이 중요합니다.

```text
1. 정상 저장이 왜 중요한가?
2. 중복 저장을 왜 막아야 하는가?
3. 실패했을 때 rollback을 왜 확인해야 하는가?
```

## 작성 전 확인

아래 파일을 먼저 확인합니다.

```bash
sed -n '1,280p' internal/ledger/repository.go
sed -n '1,320p' internal/ledger/repository_test.go
```

Day19 실습을 마친 뒤에는 아래 테스트 파일이 있어야 합니다.

```text
internal/ledger/repository_test.go
```

## 1. Day19에서 Repository 저장 검증을 하는 이유는 무엇인가?

작성 힌트:

```text
컴파일 성공만으로는 부족하고,
transaction row와 entry rows가 실제 DB에서 함께 저장되는지 확인해야 한다는 점을 적는다.
```

내 답변:

```text
Repository 코드는 컴파일만 성공해서는 충분하지 않다.
실제 PostgreSQL에서 Transaction과 Entries가 함께 저장되고, 실패 시 전체 rollback이 동작하는지 확인해야 돈의 이동 기록을 신뢰할 수 있다.
```

## 2. `idempotency_key`는 어떤 중복을 막기 위한 값인가?

작성 힌트:

```text
같은 payment finalized 같은 비즈니스 사건이 재시도되거나 중복 수신되어도 Ledger가 두 번 기록되지 않게 한다는 점을 적는다.
```

내 답변:

```text
같은 결제 확정처럼 동일한 비즈니스 사건이 재시도되거나 중복 수신될 때 Ledger Transaction이 두 번 저장되는 것을 막는 값이다.
단순히 모든 거래의 중복을 막는 것이 아니라, 같은 사건에 동일한 `idempotency_key`를 사용해 중복 처리를 막는다.
```

## 3. foreign key 실패 케이스를 일부러 테스트하는 이유는 무엇인가?

작성 힌트:

```text
존재하지 않는 account_id를 넣으면 entry 저장이 실패해야 하고,
그때 앞에서 넣은 transaction도 rollback되는지 확인해야 한다는 점을 적는다.
```

내 답변:

```text
존재하지 않는 `account_id`를 사용하면 Entry 저장이 foreign key 오류로 실패해야 한다.
이때 같은 DB transaction 안에서 먼저 저장했던 Ledger Transaction까지 rollback되어 일부 데이터만 남지 않는지 확인하기 위해서다.
```

## 4. Day19의 테스트는 unit test보다 integration test에 가까운 이유는 무엇인가?

작성 힌트:

```text
실제 PostgreSQL, table, unique index, foreign key 같은 DB 구성요소를 함께 확인하기 때문이라고 적는다.
```

내 답변:

```text
실제 PostgreSQL에 연결해 table, unique index, foreign key, DB transaction과 rollback을 함께 확인하기 때문이다.
Go 함수만 격리해서 확인하는 unit test가 아니라 코드와 DB 구성요소의 연결을 검증한다.
```

## 5. Day20에서 Service와 Repository를 연결할 때 주의해야 할 점은 무엇인가?

작성 힌트:

```text
Service가 먼저 Ledger 균형을 검증하고,
검증된 transaction과 entries만 Repository로 넘겨야 한다는 점을 적는다.
```

내 답변:

```text
Service가 금액, 통화, Entry 수, debit·credit 균형 같은 도메인 규칙을 먼저 검증해야 한다.
검증을 통과한 Transaction과 Entries만 Repository의 `CreateTransaction`으로 넘겨 하나의 DB transaction으로 저장해야 한다.
```

## 오늘 실행 결과

아래 명령 실행 결과를 짧게 기록합니다.

```bash
gofmt -w internal/ledger/repository_test.go
go test ./...
TEST_DATABASE_URL="postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" go test ./internal/ledger -run Repository
```

기록:

```text
`go test ./...` 실행 결과 모든 패키지가 통과했다.
`TEST_DATABASE_URL`을 지정한 통합 테스트에서는 정상 저장, idempotency 중복 방지, foreign key 실패 rollback 케이스가 모두 통과했다.

중간에 나타난 `zsh: command not found: feat:`는 Go 테스트 오류가 아니라 커밋 메시지인 `feat: ...`를 셸 명령처럼 직접 입력해서 발생한 오류다.
커밋할 때는 `git commit -m "feat: ..."` 형식으로 실행해야 한다.
```

## 아직 헷갈리는 부분

아래 후보 중 헷갈리는 것이 있으면 적습니다.

```text
idempotency
unique index
foreign key
rollback
integration test
TEST_DATABASE_URL
```

메모:

```text
Day19 테스트는 Repository가 실제 DB 제약조건과 함께 안전하게 동작하는지 확인하는 integration test다.
`TEST_DATABASE_URL`이 없으면 DB 테스트는 skip되고, 값을 설정하면 Docker PostgreSQL을 대상으로 실행된다.
```

## 정답/점검 가이드

먼저 `내 답변`을 작성한 뒤 아래 내용을 펼쳐서 비교합니다.

### 1. Day19에서 Repository 저장 검증을 하는 이유는 무엇인가?

<details>
<summary>정답/점검 가이드 보기</summary>

Repository 저장 검증은 Day18에서 만든 `CreateTransaction`이 실제 DB에서도 안전하게 동작하는지 확인하기 위해 필요합니다.

특히 Ledger는 돈의 이동 기록이므로, 단순히 코드가 컴파일되는 것만으로는 부족합니다.

`ledger_transactions` row 1개와 `ledger_entries` row 여러 개가 함께 저장되는지 확인해야 합니다.

</details>

### 2. `idempotency_key`는 어떤 중복을 막기 위한 값인가?

<details>
<summary>정답/점검 가이드 보기</summary>

`idempotency_key`는 같은 비즈니스 사건이 두 번 처리되는 것을 막기 위한 값입니다.

예를 들어 `payment:pay_123:finalized`는 `pay_123` 결제가 확정되었다는 사건을 나타냅니다.

같은 사건이 재시도되거나 중복 수신되어도 이 키가 같으면 DB의 unique index가 두 번째 저장을 막습니다.

</details>

### 3. foreign key 실패 케이스를 일부러 테스트하는 이유는 무엇인가?

<details>
<summary>정답/점검 가이드 보기</summary>

`ledger_entries.account_id`는 `ledger_accounts.id`를 참조합니다.

존재하지 않는 account id를 넣으면 entry 저장은 실패해야 합니다.

그리고 이 실패가 발생했을 때 앞에서 저장했던 `ledger_transactions` row도 rollback되어야 합니다.

그래야 transaction만 남고 entry가 없는 깨진 원장을 막을 수 있습니다.

</details>

### 4. Day19의 테스트는 unit test보다 integration test에 가까운 이유는 무엇인가?

<details>
<summary>정답/점검 가이드 보기</summary>

Day19 테스트는 실제 PostgreSQL을 연결해서 실행하는 테스트입니다.

따라서 Go 함수 하나만 검증하는 것이 아니라 DB table, unique index, foreign key, transaction rollback까지 함께 확인합니다.

그래서 unit test보다 integration test에 가깝습니다.

</details>

### 5. Day20에서 Service와 Repository를 연결할 때 주의해야 할 점은 무엇인가?

<details>
<summary>정답/점검 가이드 보기</summary>

Service는 먼저 Ledger entry의 도메인 규칙을 검증해야 합니다.

예를 들어 debit과 credit 합계가 맞는지, 금액이 0보다 큰지, 통화가 비어 있지 않은지 확인합니다.

그 다음 검증된 transaction과 entries만 Repository로 넘겨 저장해야 합니다.

즉 Day20의 핵심 흐름은 아래입니다.

```text
ValidateTransaction
-> CreateTransaction
```

</details>

## 추가 보충 정리

### Codex 점검

오늘 산출물에서 가장 중요한 문장은 아래입니다.

```text
Ledger Repository 테스트는 정상 저장뿐 아니라 중복 방지와 rollback까지 확인해야 한다.
```

작성한 핵심 방향은 맞았다. 다만 `idempotency_key`는 막연한 거래 중복이 아니라 **같은 비즈니스 사건의 중복 처리**를 막는 키라는 점이 중요하다.

Day20 연결 순서는 아래처럼 구체적으로 기억한다.

```text
Service 도메인 검증
-> 검증 실패 시 Repository 호출 금지
-> 검증 성공 시 Repository.CreateTransaction 호출
-> Repository가 PostgreSQL transaction으로 원자적 저장
```

### 코드 검토 결과

기존 `CreateTransaction`은 아래처럼 검증 오류를 버리고 있었다.

```go
_ = validateTransaction(tx)
_ = validateEntries(tx, entries)
```

이를 검증 오류를 즉시 반환하는 형태로 수정했다.

```go
if err := validateTransaction(tx); err != nil {
	return err
}

if err := validateEntries(tx, entries); err != nil {
	return err
}
```

Entry가 하나도 없는 입력도 DB 저장 전에 실패하도록 검증과 테스트를 추가했다.

### 코드 확인 포인트

실습이 끝난 뒤 아래 항목을 코드에서 직접 체크합니다.

```text
- TEST_DATABASE_URL이 없으면 테스트가 skip되는가?
- 테스트용 account seed가 있는가?
- 정상 저장 후 transaction count와 entry count를 확인하는가?
- 같은 idempotency_key로 두 번째 저장할 때 실패를 기대하는가?
- foreign key 실패 후 transaction row가 0개인지 확인하는가?
```

### 다음 학습 포인트

Day20에서 특히 이어서 보면 좋은 포인트는 아래입니다.

```text
1. Service가 Repository를 직접 가지게 할 것인가?
2. Service 테스트에서 fake repository를 쓸 것인가?
3. ValidateTransaction과 CreateTransaction을 하나의 use case로 묶을 것인가?
```
