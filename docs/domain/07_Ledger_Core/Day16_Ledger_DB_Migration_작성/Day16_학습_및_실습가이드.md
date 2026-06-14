# Day 16 학습 및 실습가이드 - Ledger DB Migration 작성

관련 Jira: SPN-33

Day16은 Ledger Core를 DB에 처음으로 새기는 날입니다.

Day12에서는 Ledger 도메인 타입을 만들었습니다.

```text
Account
Transaction
Entry
```

Day15에서는 Ledger Transaction이 저장되기 전에 안전한지 검증하는 Service를 준비했습니다.

Day16에서는 그 타입들이 실제 PostgreSQL 테이블로 어떻게 바뀌는지 학습하고, migration SQL을 작성합니다.

## 오늘의 큰 그림

![Day16 Ledger DB Migration 작성](../../../confluence/diagrams/spn33-day16-ledger-migration.png)

## 오늘의 핵심 문장

```text
Go 구조체는 메모리 안의 모양이고,
DB 테이블은 오래 보존되는 기록의 모양이다.
```

Ledger는 돈의 이동 기록입니다.

서비스가 재시작되어도 기록이 남아야 하고, 나중에 정산과 대사에서 다시 조회할 수 있어야 합니다.

그래서 Ledger에는 DB 테이블이 필요합니다.

## 오늘 읽고 작업할 순서

이 문서는 기존의 `기초학습`, `개념학습`, `실습가이드`를 하나로 합친 문서입니다.

아래 순서대로 위에서 아래로 읽으면 됩니다.

```text
1. 왜 Ledger에 DB 테이블이 필요한지 이해한다.
2. Migration, up/down, primary key, foreign key, index를 이해한다.
3. Account, Transaction, Entry가 어떤 테이블로 바뀌는지 연결한다.
4. up.sql과 down.sql을 작성한다.
5. PostgreSQL에서 적용과 롤백을 확인한다.
6. Go 테스트를 실행한다.
7. 실습산출물을 작성한다.
```

## 1. 오늘 만들 테이블

| Go 타입 | DB 테이블 | 역할 |
| --- | --- | --- |
| `Account` | `ledger_accounts` | 원장에서 돈이 기록되는 주체 |
| `Transaction` | `ledger_transactions` | 여러 Entry를 하나로 묶는 원장 거래 |
| `Entry` | `ledger_entries` | 실제 돈의 이동 한 줄 |

이 구조는 Day12의 Go 타입과 연결됩니다.

```text
Account      -> ledger_accounts
Transaction  -> ledger_transactions
Entry        -> ledger_entries
```

## 2. 오늘 만들거나 보강하는 범위

오늘 작성할 파일:

```text
migrations/000002_create_ledger_core_tables.up.sql
migrations/000002_create_ledger_core_tables.down.sql
```

오늘 하지 않는 것:

```text
Repository 작성
Service와 DB 연결
HTTP API 작성
Payment FINALIZED와 Ledger 자동 연결
Settlement 계산
```

Day16의 목표는 Go 저장 코드를 만드는 것이 아닙니다.

먼저 저장될 수 있는 DB 구조를 SQL로 정의하는 것입니다.

## 3. Migration이란 무엇인가?

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

## 4. up migration과 down migration

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

## 5. Primary Key는 무엇인가?

Primary Key는 테이블 안에서 row 하나를 유일하게 식별하는 값입니다.

예를 들어 `ledger_accounts`의 `id`는 계정 하나를 식별합니다.

```sql
id TEXT PRIMARY KEY
```

Java 객체로 비유하면 `id` 필드와 비슷합니다.

하지만 DB에서는 단순 필드가 아니라 “중복될 수 없는 식별자”라는 제약이 붙습니다.

## 6. Foreign Key는 무엇인가?

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

## 7. Index는 왜 필요한가?

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

## 8. `idempotency_key`는 왜 unique인가?

Idempotency는 같은 요청이나 같은 이벤트가 여러 번 들어와도 결과가 한 번만 반영되는 성질입니다.

Ledger에서는 같은 원장 거래가 두 번 저장되면 위험합니다.

예를 들어 같은 payment finalized 이벤트가 두 번 처리되어 Ledger Transaction이 두 번 저장되면, 가맹점 잔액이 두 번 늘어난 것처럼 보일 수 있습니다.

그래서 `idempotency_key`는 unique index로 막습니다.

```text
같은 idempotency_key를 가진 transaction은 한 번만 저장한다.
```

## 9. `amount`는 왜 BIGINT인가?

돈을 소수점으로 저장하면 부동소수점 오차가 생길 수 있습니다.

그래서 USDC 같은 토큰은 최소 단위 정수로 저장합니다.

예:

```text
10 USDC = 10_000_000
```

큰 정수 금액을 저장하기 위해 PostgreSQL에서는 `BIGINT`를 사용합니다.

Go 코드에서는 `int64`와 대응됩니다.

## 10. 삭제 순서가 중요한 이유

`down.sql`에서는 테이블을 만들 때와 반대 순서로 삭제해야 합니다.

```text
ledger_entries
-> ledger_transactions
-> ledger_accounts
```

이유는 `ledger_entries`가 `ledger_transactions`와 `ledger_accounts`를 참조하기 때문입니다.

참조하는 테이블을 먼저 삭제해야 외래키 관계 때문에 실패하지 않습니다.

## 11. 실습 전 현재 파일 확인

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

Day16을 진행하기 전에는 `000002` 파일이 없어도 정상입니다.

오늘 새로 만들 파일입니다.

## 12. 실습 SQL 작성 원칙

실습할 때는 SQL 조각만 보고 맞히는 방식으로 진행하지 않습니다.

아래 `최종 완성본 전체`를 기준으로 새 파일을 만듭니다.

```text
1. migrations 폴더에 000002 파일 2개가 있는지 확인한다.
2. 없다면 새로 만든다.
3. 아래 최종 완성본 전체를 기준으로 up.sql과 down.sql을 작성한다.
4. PostgreSQL에 적용한다.
5. 테이블 생성, 컬럼, 인덱스, 롤백을 확인한다.
```

Day16에서 직접 작성할 파일은 2개입니다.

```text
migrations/000002_create_ledger_core_tables.up.sql
migrations/000002_create_ledger_core_tables.down.sql
```

## 13. `up.sql` 최종 완성본 전체

파일:

```text
migrations/000002_create_ledger_core_tables.up.sql
```

최종 파일 전체:

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

위 내용을 `migrations/000002_create_ledger_core_tables.up.sql` 파일 전체 내용으로 작성합니다.

## 14. `up.sql` 코드 해석

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

## 15. `down.sql` 최종 완성본 전체

파일:

```text
migrations/000002_create_ledger_core_tables.down.sql
```

최종 파일 전체:

```sql
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS ledger_transactions;
DROP TABLE IF EXISTS ledger_accounts;
```

위 내용을 `migrations/000002_create_ledger_core_tables.down.sql` 파일 전체 내용으로 작성합니다.

삭제 순서는 중요합니다.

```text
Entry가 Transaction과 Account를 참조하므로,
참조하는 쪽인 ledger_entries를 먼저 삭제한다.
```

## 16. PostgreSQL 실행

로컬 PostgreSQL을 실행합니다.

```bash
docker compose up -d
docker compose ps
```

`stablepay-postgres`가 `running` 상태인지 확인합니다.

## 17. migration 적용

먼저 payment core 테이블을 적용합니다.

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -f migrations/000001_create_payment_core_tables.up.sql
```

그 다음 Day16 Ledger migration을 적용합니다.

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -f migrations/000002_create_ledger_core_tables.up.sql
```

이미 기존 테이블이 있어서 실패한다면, 현재 DB 상태를 확인한 뒤 로컬 실습 DB를 초기화하거나 down migration을 적용해야 합니다.

## 18. 테이블 생성 확인

아래 명령으로 Ledger 테이블이 만들어졌는지 확인합니다.

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -c "\dt ledger_*"
```

예상 결과:

```text
ledger_accounts
ledger_transactions
ledger_entries
```

## 19. 컬럼 확인

각 테이블 구조를 확인합니다.

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -c "\d ledger_accounts"
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -c "\d ledger_transactions"
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -c "\d ledger_entries"
```

확인할 것:

```text
primary key가 있는가?
foreign key가 있는가?
index가 있는가?
amount가 BIGINT인가?
created_at이 TIMESTAMPTZ인가?
```

## 20. 롤백 확인

Day16 migration만 롤백합니다.

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -f migrations/000002_create_ledger_core_tables.down.sql
```

다시 확인합니다.

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -c "\dt ledger_*"
```

Ledger 테이블이 보이지 않으면 롤백이 된 것입니다.

롤백 확인 후 다시 실습을 이어가고 싶다면 `up.sql`을 한 번 더 적용합니다.

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -f migrations/000002_create_ledger_core_tables.up.sql
```

## 21. Go 테스트 실행

SQL migration만 추가했더라도 기존 Go 코드가 깨지지 않았는지 확인합니다.

```bash
go test ./...
```

## 22. 완성본 확인

오늘 작업 후 파일 구조는 아래처럼 보여야 합니다.

```text
migrations/
  000001_create_payment_core_tables.up.sql
  000001_create_payment_core_tables.down.sql
  000002_create_ledger_core_tables.up.sql
  000002_create_ledger_core_tables.down.sql
```

두 파일의 최종 내용은 반드시 이 문서의 `최종 완성본 전체`와 비교합니다.

```text
up.sql은 테이블과 인덱스를 만든다.
down.sql은 만든 테이블을 참조 관계의 반대 순서로 지운다.
```

## 23. 완료 기준

아래를 모두 만족하면 Day16 완료입니다.

```text
up migration 작성 완료
down migration 작성 완료
ledger_* 테이블 생성 확인
rollback 확인
go test ./... 성공
Day16 산출물 5문항 작성
```

## 24. 커밋 메시지

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

## 25. 다음 작업 예고

Day16이 끝나면 Day17에서는 Repository 초안으로 넘어갑니다.

```text
Day16: 테이블을 만든다.
Day17: 테이블에 저장하는 Go 코드를 만든다.
```
