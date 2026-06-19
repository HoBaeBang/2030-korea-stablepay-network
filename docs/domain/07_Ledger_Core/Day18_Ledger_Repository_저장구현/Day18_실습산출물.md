# Day 18 실습산출물 - Ledger Repository 저장 구현

관련 Jira: SPN-35

Day18 산출물은 오늘 만든 저장 메서드의 책임과 저장 흐름을 자기 말로 설명하는 문서입니다.

정답처럼 완벽하게 쓰는 것보다,
오늘 코드에서 무엇을 이해했고 무엇이 아직 헷갈리는지 분명하게 적는 것이 더 중요합니다.

## 작성 전 확인

아래 파일을 먼저 확인합니다.

```bash
sed -n '1,240p' internal/ledger/repository.go
```

아래 메서드가 보여야 합니다.

```go
func (r *Repository) CreateTransaction(ctx context.Context, tx Transaction, entries []Entry) error
```

## 1. 오늘 만든 `CreateTransaction` 메서드의 책임은 무엇인가?

작성 힌트:

```text
transaction 1건과 entry 여러 건을
하나의 DB transaction으로 저장한다는 점을 적는다.
```

내 답변:

```text
`CreateTransaction`은 Ledger Transaction 1건과 그 거래에 속하는 Entry 여러 건을 하나의 DB transaction 안에서 저장한다.
모든 INSERT가 성공하면 commit하고, 하나라도 실패하면 전체를 rollback해 원장이 일부만 저장되는 것을 막는다.
```

## 2. 왜 `sql.Tx`가 필요한가?

작성 힌트:

```text
중간에 하나라도 실패하면 전체를 rollback해야 한다는 점을 적는다.
```

내 답변:

```text
여러 INSERT를 하나의 작업 단위로 묶기 위해 필요하다.
Transaction 저장이나 Entry 저장 중 하나라도 실패하면 앞에서 실행한 INSERT까지 모두 rollback해야 깨진 원장이 남지 않는다.
```

## 3. 왜 `ledger_transactions`를 먼저 저장하고 `ledger_entries`를 나중에 저장하는가?

작성 힌트:

```text
entry의 transaction_id가 transaction을 참조하는 foreign key라는 점을 적는다.
```

내 답변:

```text
`ledger_entries.transaction_id`가 `ledger_transactions.id`를 foreign key로 참조하기 때문이다.
부모인 Transaction row를 먼저 저장해야 자식인 Entry row를 저장할 수 있다.
```

## 4. 왜 오늘은 `ledger_accounts` INSERT를 같이 만들지 않았는가?

작성 힌트:

```text
오늘은 거래 저장 경계에 집중하고,
account는 이미 존재한다고 가정한다는 점을 적는다.
```

내 답변:

```text
Day18은 이미 존재하는 Ledger Account 사이의 거래를 저장하는 경계에 집중하기 때문이다.
Account는 있을 수도 있고 없을 수도 있는 값이 아니라, Entry를 저장하기 전에 반드시 존재해야 한다. 존재하지 않으면 foreign key 오류가 발생한다.
```

## 5. Day19에서 무엇을 더 검증해야 할 것 같은가?

작성 힌트:

```text
idempotency_key 중복,
저장 성공/실패,
foreign key 오류,
실제 DB 저장 검증을 떠올린다.
```

내 답변:

```text
정상적으로 Transaction과 Entries가 함께 저장되는지, 같은 `idempotency_key`가 중복 저장을 막는지 확인해야 한다.
또한 존재하지 않는 Account를 참조해 Entry 저장이 실패하면 앞서 저장한 Transaction도 rollback되는지 실제 PostgreSQL에서 검증해야 한다.
```

## 오늘 실행 결과

아래 명령 실행 결과를 짧게 기록합니다.

```bash
gofmt -w internal/ledger/repository.go
go test ./...
```

기록:

```text
`go test ./...` 실행 결과 모든 패키지가 통과했다.
테스트 파일이 없는 패키지는 `[no test files]`로 표시되었고, Ledger를 포함한 테스트 패키지는 `ok`로 완료되었다.
```

## 아직 헷갈리는 부분

아래 후보 중 헷갈리는 것이 있으면 적습니다.

```text
sql.Tx
commit
rollback
foreign key
idempotency_key
Repository와 Service의 경계
```

메모:

```text
처음에는 BeginTx 직후 committed가 false라서 쿼리 실행 전에 rollback하는 것으로 이해했다.
하지만 defer에 등록된 함수는 그 자리에서 실행되는 것이 아니라 CreateTransaction이 반환될 때 실행된다.

BeginTx로 DB transaction 경계를 먼저 열고, 이후 쿼리를 r.db가 아니라 sqlTx.ExecContext로 실행하면 모든 쿼리가 그 transaction에 포함된다.
중간에 return하면 committed가 false이므로 rollback하고, Commit이 성공한 뒤 committed를 true로 바꾸면 defer는 rollback하지 않는다.
```

## 정답/점검 가이드

먼저 `내 답변`을 작성한 뒤 아래 내용을 펼쳐서 비교합니다.

### 1. 오늘 만든 `CreateTransaction` 메서드의 책임은 무엇인가?

<details>
<summary>정답/점검 가이드 보기</summary>

`CreateTransaction` 메서드는 Ledger Transaction 1건과 Ledger Entry 여러 건을 하나의 DB transaction으로 저장하는 책임을 가집니다.

즉, 원장 거래의 비즈니스 사건 설명 row와 실제 돈의 이동 row들을 함께 저장하는 메서드입니다.

</details>

### 2. 왜 `sql.Tx`가 필요한가?

<details>
<summary>정답/점검 가이드 보기</summary>

`sql.Tx`가 필요한 이유는 여러 INSERT가 하나의 저장 묶음으로 움직여야 하기 때문입니다.

중간에 하나라도 실패하면 앞에서 저장한 row도 함께 취소되어야 원장 기록이 깨지지 않습니다.

그래서 BeginTx -> 여러 INSERT -> Commit, 실패 시 Rollback 흐름이 필요합니다.

</details>

### 3. 왜 `ledger_transactions`를 먼저 저장하고 `ledger_entries`를 나중에 저장하는가?

<details>
<summary>정답/점검 가이드 보기</summary>

`ledger_entries.transaction_id`는 `ledger_transactions.id`를 참조하는 foreign key입니다.

그래서 transaction row가 먼저 있어야 entry row가 그 id를 참조할 수 있습니다.

즉 저장 순서는 transaction 먼저, entry 나중입니다.

</details>

### 4. 왜 오늘은 `ledger_accounts` INSERT를 같이 만들지 않았는가?

<details>
<summary>정답/점검 가이드 보기</summary>

Day18의 목표는 "거래 저장 경계"를 만드는 것입니다.

`ledger_accounts`는 계정 마스터 데이터에 가깝고, 오늘의 핵심 저장 단위는 transaction과 entries입니다.

그래서 오늘은 entry가 참조하는 account row가 이미 존재한다고 가정하고, 거래 저장 메서드 구현에만 집중합니다.

</details>

### 5. Day19에서 무엇을 더 검증해야 할 것 같은가?

<details>
<summary>정답/점검 가이드 보기</summary>

Day19에서는 아래를 더 검증해야 합니다.

```text
1. 실제 DB에 transaction과 entries가 저장되는가?
2. 중간 실패 시 rollback이 잘 되는가?
3. 같은 idempotency_key를 다시 저장하려고 하면 어떻게 되는가?
4. 존재하지 않는 account_id를 넣으면 어떤 foreign key 오류가 나는가?
5. 저장 테스트를 어떤 형태로 붙이는 것이 좋은가?
```

</details>

## 추가 보충 정리

### Codex 점검

오늘 산출물에서 가장 중요한 것은 아래 두 문장을 자기 말로 설명하는 것입니다.

```text
1. Ledger 저장은 transaction row 1개와 entry row 여러 개를 함께 저장하는 일이다.
2. 그래서 sql.Tx가 필요하다.
```

작성한 개념은 전반적으로 맞았다. 다만 Account는 선택 사항이 아니라 Entry가 참조하기 전에 반드시 존재해야 한다.

또한 코드 점검 과정에서 `validateTransaction`과 `validateEntries`의 반환 오류를 `_ =`로 버리고 있던 문제를 발견했다. 검증 오류를 즉시 반환하도록 수정하고, DB 접근 전에 검증이 끝나는 회귀 테스트를 추가했다.

### `BeginTx`, `defer`, `committed` 실행 순서

```text
BeginTx로 transaction 시작
-> sqlTx로 Transaction INSERT
-> sqlTx로 Entries INSERT
-> Commit 성공
-> committed = true
-> 함수 종료 시 defer 실행
-> committed가 true이므로 Rollback하지 않음
```

중간 INSERT 또는 Commit 전에 오류가 반환되면 `committed`는 계속 false이므로, 함수 종료 시 defer가 Rollback을 시도한다.

### 코드 확인 포인트

실습이 끝난 뒤 아래 항목을 코드에서 직접 체크합니다.

```text
- BeginTx가 있는가?
- rollback defer가 있는가?
- ledger_transactions INSERT가 먼저 실행되는가?
- entries loop INSERT가 있는가?
- 마지막에 Commit이 있는가?
```

### 다음 학습 포인트

Day19에서 특히 이어서 보면 좋은 포인트는 아래입니다.

```text
1. idempotency_key unique index가 실제로 어떤 중복을 막는가?
2. foreign key 오류는 저장 흐름에서 어떻게 드러나는가?
3. DB 테스트를 붙일 때 seed account를 어떻게 준비해야 하는가?
```
