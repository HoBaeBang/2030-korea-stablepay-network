# Day 16 개념학습 - Ledger 테이블과 Migration

관련 Jira: SPN-33

## 1. Migration이란 무엇인가?

Migration은 DB 구조 변경을 코드처럼 관리하는 파일입니다.

예를 들어 새로운 테이블을 만들거나, 컬럼을 추가하거나, 인덱스를 추가할 때 migration 파일을 작성합니다.

이 프로젝트에서는 아래 폴더에 SQL 파일을 둡니다.

```text
migrations/
```

현재 이미 payment core 테이블을 만드는 migration이 있습니다.

```text
000001_create_payment_core_tables.up.sql
000001_create_payment_core_tables.down.sql
```

Day16에서는 두 번째 migration을 추가합니다.

```text
000002_create_ledger_core_tables.up.sql
000002_create_ledger_core_tables.down.sql
```

## 2. up migration과 down migration

`up.sql`은 앞으로 가는 변경입니다.

```text
테이블 생성
컬럼 추가
인덱스 추가
```

`down.sql`은 되돌리는 변경입니다.

```text
인덱스 삭제
테이블 삭제
```

Day16에서는 아래처럼 생각하면 됩니다.

```text
up   = Ledger 테이블을 만든다.
down = Ledger 테이블을 지운다.
```

## 3. Ledger 테이블 3개가 필요한 이유

Ledger는 하나의 테이블에 전부 넣기보다 역할별로 나눕니다.

| 테이블 | 역할 |
| --- | --- |
| `ledger_accounts` | 돈이 기록되는 원장 계정 |
| `ledger_transactions` | 여러 Entry를 묶는 거래 단위 |
| `ledger_entries` | 실제 돈의 이동 한 줄 |

이 구조는 Day12의 Go 타입과 연결됩니다.

```text
Account      -> ledger_accounts
Transaction  -> ledger_transactions
Entry        -> ledger_entries
```

## 4. Primary Key는 무엇인가?

Primary Key는 테이블 안에서 row 하나를 유일하게 식별하는 값입니다.

예를 들어 `ledger_accounts`의 `id`는 계정 하나를 식별합니다.

```sql
id TEXT PRIMARY KEY
```

Java 객체로 비유하면 `id` 필드와 비슷합니다.

하지만 DB에서는 단순 필드가 아니라 “중복될 수 없는 식별자”라는 제약이 붙습니다.

## 5. Foreign Key는 무엇인가?

Foreign Key는 다른 테이블의 row를 참조하는 값입니다.

예를 들어 `ledger_entries.transaction_id`는 `ledger_transactions.id`를 참조합니다.

```sql
transaction_id TEXT NOT NULL REFERENCES ledger_transactions (id)
```

뜻:

```text
Entry는 반드시 어떤 Ledger Transaction에 속해야 한다.
존재하지 않는 Transaction을 가리키는 Entry는 저장할 수 없다.
```

`ledger_entries.account_id`도 `ledger_accounts.id`를 참조합니다.

```sql
account_id TEXT NOT NULL REFERENCES ledger_accounts (id)
```

뜻:

```text
Entry는 반드시 어떤 원장 계정에 기록되어야 한다.
존재하지 않는 Account에 대한 Entry는 저장할 수 없다.
```

## 6. Index는 왜 필요한가?

Index는 조회를 빠르게 하기 위한 보조 구조입니다.

Ledger에서는 나중에 아래 조회가 자주 필요합니다.

```text
특정 transaction의 entry 목록 조회
특정 account의 entry 목록 조회
특정 reference와 연결된 ledger transaction 조회
idempotency_key로 중복 처리 여부 확인
```

그래서 아래 index 후보가 필요합니다.

```sql
CREATE INDEX idx_ledger_entries_transaction_id ON ledger_entries (transaction_id);
CREATE INDEX idx_ledger_entries_account_id ON ledger_entries (account_id);
CREATE INDEX idx_ledger_transactions_reference ON ledger_transactions (reference_type, reference_id);
CREATE UNIQUE INDEX idx_ledger_transactions_idempotency_key
    ON ledger_transactions (idempotency_key);
```

## 7. `idempotency_key`는 왜 unique인가?

Idempotency는 같은 요청이나 같은 이벤트가 여러 번 들어와도 결과가 한 번만 반영되는 성질입니다.

Ledger에서는 같은 원장 거래가 두 번 저장되면 위험합니다.

예를 들어 같은 payment finalized 이벤트가 두 번 처리되어 Ledger Transaction이 두 번 저장되면, 가맹점 잔액이 두 번 늘어난 것처럼 보일 수 있습니다.

그래서 `idempotency_key`는 unique index로 막습니다.

```text
같은 idempotency_key를 가진 transaction은 한 번만 저장한다.
```

## 8. 삭제 순서가 중요한 이유

`down.sql`에서는 테이블을 만들 때와 반대 순서로 삭제해야 합니다.

```text
ledger_entries
-> ledger_transactions
-> ledger_accounts
```

이유는 `ledger_entries`가 `ledger_transactions`와 `ledger_accounts`를 참조하기 때문입니다.

참조하는 테이블을 먼저 지우면 외래키 관계가 깨집니다.

## 9. 오늘의 결론

```text
Day16의 핵심은 Ledger를 DB에 저장하는 코드를 만드는 것이 아니다.

먼저 저장될 수 있는 구조를 SQL로 정의하고,
나중에 Repository가 믿고 사용할 수 있는 테이블을 준비하는 것이다.
```
