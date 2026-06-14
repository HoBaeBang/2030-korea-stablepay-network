# Day 16 검증문제와 답변가이드

관련 Jira: SPN-33

먼저 문제를 풀어보고, 필요할 때만 답변을 펼쳐서 확인합니다.

## 먼저 풀어볼 문제

1. Migration은 무엇이고 왜 필요한가?
2. `up.sql`과 `down.sql`의 차이는 무엇인가?
3. `ledger_accounts`, `ledger_transactions`, `ledger_entries`는 각각 무엇을 저장하는가?
4. `ledger_entries.transaction_id`가 foreign key인 이유는 무엇인가?
5. `idempotency_key`에 unique index를 두는 이유는 무엇인가?
6. `amount`를 `BIGINT`로 저장하는 이유는 무엇인가?
7. `down.sql`에서 테이블 삭제 순서가 중요한 이유는 무엇인가?

## 내 답변 작성 공간

아래 공간에 먼저 내 생각을 적어봅니다.

정답을 바로 펼치지 말고, 최소 한 문장이라도 먼저 작성한 뒤 답변가이드를 확인합니다.

### 1. Migration은 무엇이고 왜 필요한가?

```text
내 답변:
```

### 2. `up.sql`과 `down.sql`의 차이는 무엇인가?

```text
내 답변:
```

### 3. `ledger_accounts`, `ledger_transactions`, `ledger_entries`는 각각 무엇을 저장하는가?

```text
내 답변:
```

### 4. `ledger_entries.transaction_id`가 foreign key인 이유는 무엇인가?

```text
내 답변:
```

### 5. `idempotency_key`에 unique index를 두는 이유는 무엇인가?

```text
내 답변:
```

### 6. `amount`를 `BIGINT`로 저장하는 이유는 무엇인가?

```text
내 답변:
```

### 7. `down.sql`에서 테이블 삭제 순서가 중요한 이유는 무엇인가?

```text
내 답변:
```

## 답변가이드

### 1. Migration은 무엇이고 왜 필요한가?

<details>
<summary>답변 보기</summary>

Migration은 DB 구조 변경을 파일로 관리하는 방식입니다.

테이블 생성, 컬럼 추가, 인덱스 추가 같은 변경을 SQL 파일로 남기면 어떤 DB 변경이 언제 들어갔는지 추적할 수 있습니다.

팀 프로젝트나 운영 환경에서는 DB 구조가 코드만큼 중요하므로 migration으로 관리합니다.

</details>

### 2. `up.sql`과 `down.sql`의 차이는 무엇인가?

<details>
<summary>답변 보기</summary>

`up.sql`은 변경을 적용하는 파일입니다.

예:

```text
테이블 생성
인덱스 생성
컬럼 추가
```

`down.sql`은 변경을 되돌리는 파일입니다.

예:

```text
테이블 삭제
인덱스 삭제
컬럼 삭제
```

Day16에서는 `up.sql`로 Ledger 테이블을 만들고, `down.sql`로 Ledger 테이블을 삭제합니다.

</details>

### 3. Ledger 테이블 3개는 각각 무엇을 저장하는가?

<details>
<summary>답변 보기</summary>

`ledger_accounts`는 돈이 기록되는 원장 계정을 저장합니다.

`ledger_transactions`는 여러 Entry를 하나로 묶는 거래 단위를 저장합니다.

`ledger_entries`는 실제 돈의 이동 한 줄을 저장합니다.

```text
Account      -> ledger_accounts
Transaction  -> ledger_transactions
Entry        -> ledger_entries
```

</details>

### 4. `ledger_entries.transaction_id`가 foreign key인 이유는 무엇인가?

<details>
<summary>답변 보기</summary>

Entry는 반드시 어떤 Ledger Transaction에 속해야 합니다.

그래서 `ledger_entries.transaction_id`는 `ledger_transactions.id`를 참조합니다.

이렇게 하면 존재하지 않는 transaction에 연결된 entry가 저장되는 것을 막을 수 있습니다.

</details>

### 5. `idempotency_key`에 unique index를 두는 이유는 무엇인가?

<details>
<summary>답변 보기</summary>

같은 Ledger Transaction이 두 번 저장되면 돈이 두 번 움직인 것처럼 보일 수 있습니다.

예를 들어 같은 payment finalized 이벤트가 두 번 처리되면 가맹점 지급 예정 금액이 두 번 늘어날 수 있습니다.

`idempotency_key`에 unique index를 두면 같은 key를 가진 Ledger Transaction이 중복 저장되는 것을 DB 레벨에서도 막을 수 있습니다.

</details>

### 6. `amount`를 `BIGINT`로 저장하는 이유는 무엇인가?

<details>
<summary>답변 보기</summary>

돈을 소수점으로 저장하면 부동소수점 오차가 생길 수 있습니다.

그래서 USDC 같은 토큰은 최소 단위 정수로 저장합니다.

예:

```text
10 USDC = 10_000_000
```

큰 정수 금액을 저장하기 위해 PostgreSQL에서는 `BIGINT`를 사용합니다.

Go 코드에서는 `int64`와 대응됩니다.

</details>

### 7. `down.sql`에서 테이블 삭제 순서가 중요한 이유는 무엇인가?

<details>
<summary>답변 보기</summary>

`ledger_entries`가 `ledger_transactions`와 `ledger_accounts`를 참조합니다.

따라서 참조당하는 테이블을 먼저 삭제하면 외래키 관계 때문에 실패할 수 있습니다.

그래서 참조하는 쪽부터 삭제합니다.

```text
ledger_entries
-> ledger_transactions
-> ledger_accounts
```

</details>
