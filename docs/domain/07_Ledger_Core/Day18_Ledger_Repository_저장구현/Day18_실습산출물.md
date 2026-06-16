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

```

## 2. 왜 `sql.Tx`가 필요한가?

작성 힌트:

```text
중간에 하나라도 실패하면 전체를 rollback해야 한다는 점을 적는다.
```

내 답변:

```text

```

## 3. 왜 `ledger_transactions`를 먼저 저장하고 `ledger_entries`를 나중에 저장하는가?

작성 힌트:

```text
entry의 transaction_id가 transaction을 참조하는 foreign key라는 점을 적는다.
```

내 답변:

```text

```

## 4. 왜 오늘은 `ledger_accounts` INSERT를 같이 만들지 않았는가?

작성 힌트:

```text
오늘은 거래 저장 경계에 집중하고,
account는 이미 존재한다고 가정한다는 점을 적는다.
```

내 답변:

```text

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

```

## 오늘 실행 결과

아래 명령 실행 결과를 짧게 기록합니다.

```bash
gofmt -w internal/ledger/repository.go
go test ./...
```

기록:

```text

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
