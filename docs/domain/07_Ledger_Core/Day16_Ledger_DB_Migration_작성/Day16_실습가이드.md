# Day 16 실습가이드 - Ledger DB Migration 작성

관련 Jira: SPN-33

Day16의 퇴근 후 실습은 작은 코드 작업 하나입니다.

```text
Ledger Core 테이블을 만드는 migration SQL을 작성하고,
적용과 롤백을 검증한다.
```

## 실습 흐름

![Day16 Ledger DB Migration 작성](../../../confluence/diagrams/spn33-day16-ledger-migration.png)

## 사전 조건

프로젝트 루트에서 시작합니다.

```bash
pwd
```

예상 위치:

```text
2030-korea-stablepay-network
```

현재 migration 파일을 확인합니다.

```bash
ls migrations
```

현재는 아래 파일이 있어야 합니다.

```text
000001_create_payment_core_tables.up.sql
000001_create_payment_core_tables.down.sql
```

## 오늘 만들 파일

```text
migrations/000002_create_ledger_core_tables.up.sql
migrations/000002_create_ledger_core_tables.down.sql
```

## Step 1. `up.sql` 작성

파일:

```text
migrations/000002_create_ledger_core_tables.up.sql
```

작성할 SQL:

```sql
CREATE TABLE ledger_accounts
(
    id         TEXT PRIMARY KEY,
    type       TEXT        NOT NULL,
    owner_id   TEXT        NOT NULL,
    currency   TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ledger_accounts_owner_id ON ledger_accounts (owner_id);
CREATE INDEX idx_ledger_accounts_type ON ledger_accounts (type);

CREATE TABLE ledger_transactions
(
    id              TEXT PRIMARY KEY,
    reference_type  TEXT        NOT NULL,
    reference_id    TEXT        NOT NULL,
    idempotency_key TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ledger_transactions_reference
    ON ledger_transactions (reference_type, reference_id);

CREATE UNIQUE INDEX idx_ledger_transactions_idempotency_key
    ON ledger_transactions (idempotency_key);

CREATE TABLE ledger_entries
(
    id             TEXT PRIMARY KEY,
    transaction_id TEXT        NOT NULL REFERENCES ledger_transactions (id),
    account_id     TEXT        NOT NULL REFERENCES ledger_accounts (id),
    direction      TEXT        NOT NULL,
    amount         BIGINT      NOT NULL,
    currency       TEXT        NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ledger_entries_transaction_id
    ON ledger_entries (transaction_id);

CREATE INDEX idx_ledger_entries_account_id
    ON ledger_entries (account_id);
```

## Step 2. `up.sql` 코드 해석

### `ledger_accounts`

원장에서 돈이 기록되는 주체입니다.

예:

```text
고객 계정
가맹점 지급 예정 계정
플랫폼 수수료 계정
```

`owner_id`는 이 계정이 누구와 연결되는지 나타냅니다.

예:

```text
customer_123
merchant_456
platform
```

### `ledger_transactions`

여러 Entry를 하나의 원장 거래로 묶습니다.

`reference_type`, `reference_id`는 이 원장 거래가 어떤 업무에서 왔는지 알려줍니다.

예:

```text
reference_type = PAYMENT
reference_id   = pay_123
```

`idempotency_key`는 같은 원장 거래가 두 번 저장되지 않게 막는 키입니다.

### `ledger_entries`

실제 돈의 이동 한 줄입니다.

```text
고객 계정 DEBIT 10_000_000 USDC
가맹점 계정 CREDIT 9_800_000 USDC
플랫폼 계정 CREDIT 200_000 USDC
```

`transaction_id`는 이 Entry가 어떤 Ledger Transaction에 속하는지 나타냅니다.

`account_id`는 이 Entry가 어떤 Ledger Account에 기록되는지 나타냅니다.

## Step 3. `down.sql` 작성

파일:

```text
migrations/000002_create_ledger_core_tables.down.sql
```

작성할 SQL:

```sql
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS ledger_transactions;
DROP TABLE IF EXISTS ledger_accounts;
```

삭제 순서는 중요합니다.

```text
Entry가 Transaction과 Account를 참조하므로,
참조하는 쪽인 ledger_entries를 먼저 삭제한다.
```

## Step 4. PostgreSQL 실행

로컬 PostgreSQL을 실행합니다.

```bash
docker compose up -d
docker compose ps
```

`stablepay-postgres`가 `running` 상태인지 확인합니다.

## Step 5. migration 적용

먼저 payment core 테이블을 적용합니다.

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -f migrations/000001_create_payment_core_tables.up.sql
```

그 다음 Day16 Ledger migration을 적용합니다.

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -f migrations/000002_create_ledger_core_tables.up.sql
```

이미 기존 테이블이 있어서 실패한다면, 현재 DB 상태를 확인한 뒤 로컬 실습 DB를 초기화하거나 down migration을 적용해야 합니다.

## Step 6. 테이블 생성 확인

아래 명령으로 Ledger 테이블이 만들어졌는지 확인합니다.

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -c "\\dt ledger_*"
```

예상 결과:

```text
ledger_accounts
ledger_transactions
ledger_entries
```

## Step 7. 컬럼 확인

각 테이블 구조를 확인합니다.

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -c "\\d ledger_accounts"
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -c "\\d ledger_transactions"
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -c "\\d ledger_entries"
```

확인할 것:

```text
primary key가 있는가?
foreign key가 있는가?
index가 있는가?
amount가 BIGINT인가?
created_at이 TIMESTAMPTZ인가?
```

## Step 8. 롤백 확인

Day16 migration만 롤백합니다.

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -f migrations/000002_create_ledger_core_tables.down.sql
```

다시 확인합니다.

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -c "\\dt ledger_*"
```

Ledger 테이블이 보이지 않으면 롤백이 된 것입니다.

롤백 확인 후 다시 실습을 이어가고 싶다면 `up.sql`을 한 번 더 적용합니다.

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -f migrations/000002_create_ledger_core_tables.up.sql
```

## Step 9. Go 테스트 실행

SQL migration만 추가했더라도 기존 Go 코드가 깨지지 않았는지 확인합니다.

```bash
go test ./...
```

## Step 10. 완성본 확인

오늘 작업 후 파일 구조는 아래처럼 보여야 합니다.

```text
migrations/
  000001_create_payment_core_tables.up.sql
  000001_create_payment_core_tables.down.sql
  000002_create_ledger_core_tables.up.sql
  000002_create_ledger_core_tables.down.sql
```

## Step 11. 커밋 메시지

코드 작업을 완료했다면 아래 커밋 메시지를 사용합니다.

```bash
git add migrations/000002_create_ledger_core_tables.up.sql migrations/000002_create_ledger_core_tables.down.sql
git commit -m "feat: Ledger 핵심 테이블 마이그레이션 추가"
```

산출물 문서까지 함께 정리했다면 문서 커밋은 별도로 분리합니다.

```bash
git add docs/domain/07_Ledger_Core/Day16_Ledger_DB_Migration_작성/Day16_실습산출물.md
git commit -m "docs: Day16 Ledger 마이그레이션 산출물 정리"
```

## 오늘의 완료 기준

```text
up migration 작성 완료
down migration 작성 완료
ledger_* 테이블 생성 확인
rollback 확인
go test ./... 성공
Day16 산출물 5문항 작성
```
