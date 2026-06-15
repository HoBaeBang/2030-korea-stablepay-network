# Day 16 학습 및 실습가이드 - Ledger DB Migration 작성

관련 Jira: [SPN-33](https://aslan0.atlassian.net/browse/SPN-33)

Day16은 기존의 `기초학습`, `개념학습`, `실습가이드`를 하나로 합친 통합 문서입니다.

하지만 문서가 하나로 합쳐졌다고 해서 실습 내용이 간소화되면 안 됩니다.

이 문서는 Day13 실습가이드처럼 아래 흐름을 유지합니다.

```text
사전 조건 확인
-> 오늘 만들 파일 위치 확인
-> Step별 SQL 작성
-> SQL 해석
-> migration 적용
-> 테이블/컬럼/인덱스 확인
-> 롤백 확인
-> 자주 만나는 오류
-> 실습산출물 작성
-> 커밋 메시지
```

Day16의 퇴근 후 실습은 작은 DB 작업 하나입니다.

```text
Ledger Core의 핵심 테이블 3개를 만드는 migration SQL을 작성한다.
```

## 오늘의 큰 그림

![Day16 Ledger DB Migration 작성](https://raw.githubusercontent.com/HoBaeBang/2030-korea-stablepay-network/main/docs/confluence/diagrams/spn33-day16-ledger-migration.png)

## 출퇴근 학습 목표

출퇴근 시간에는 아래 질문에 답할 수 있을 정도로 읽습니다.

```text
1. 왜 Ledger에는 DB 테이블이 필요한가?
2. Go 구조체와 DB 테이블은 무엇이 다른가?
3. Primary Key, Foreign Key, Index는 각각 무슨 역할을 하는가?
4. idempotency_key는 왜 unique해야 하는가?
5. down.sql에서 삭제 순서가 왜 중요한가?
```

## 오늘의 핵심 문장

```text
Go 구조체는 메모리 안의 모양이고,
DB 테이블은 오래 보존되는 기록의 모양이다.
```

Ledger는 돈의 이동 기록입니다.

서비스가 재시작되어도 기록이 남아야 하고, 나중에 정산과 대사에서 다시 조회할 수 있어야 합니다.

그래서 Ledger에는 DB 테이블이 필요합니다.

## 오늘 만들 테이블

Day16에서 만드는 테이블은 단순히 SQL 파일 안에 적는 이름이 아닙니다.

각 테이블은 Ledger 도메인의 서로 다른 책임을 맡습니다.

| Go 타입 | DB 테이블 | 역할 |
| --- | --- | --- |
| `Account` | `ledger_accounts` | 원장 안에서 돈이 기록되는 자리 |
| `Transaction` | `ledger_transactions` | 하나의 비즈니스 사건을 원장 거래 묶음으로 표현 |
| `Entry` | `ledger_entries` | 실제 돈의 이동 한 줄 |

이 구조는 Day12의 Go 타입과 연결됩니다.

```text
Account      -> ledger_accounts
Transaction  -> ledger_transactions
Entry        -> ledger_entries
```

## 테이블별 책임을 먼저 이해하기

Day16에서 가장 중요한 것은 SQL 문법을 외우는 것이 아닙니다.

아래 3개 테이블이 서로 다른 질문에 답한다는 점을 이해하는 것입니다.

```text
ledger_accounts
-> 돈이 어디에 기록되는가?

ledger_transactions
-> 이 돈의 이동은 어떤 비즈니스 사건에서 발생했는가?

ledger_entries
-> 실제로 얼마가 어느 계정에 debit/credit 되었는가?
```

예를 들어 고객이 10 USDC를 결제하고, 플랫폼 수수료가 0.2 USDC라고 해봅니다.

이때 Ledger는 아래처럼 생각할 수 있습니다.

```text
비즈니스 사건:
payment pay_123 finalized

원장 거래 묶음:
ledger_transactions row 1개

돈의 이동:
ledger_entries row 3개
```

구체적으로는 아래와 같습니다.

```text
ledger_transactions
  id              = led_tx_123
  reference_type  = PAYMENT
  reference_id    = pay_123
  idempotency_key = payment:pay_123:finalized

ledger_entries
  1. 고객 계정              DEBIT  10_000_000 USDC
  2. 가맹점 지급 예정 계정  CREDIT  9_800_000 USDC
  3. 플랫폼 수수료 계정      CREDIT    200_000 USDC
```

즉, `ledger_transactions`는 “왜 이 돈의 이동이 생겼는가?”를 설명하고,
`ledger_entries`는 “돈이 실제로 어떻게 나뉘어 기록되었는가?”를 설명합니다.

그리고 `ledger_accounts`는 각 Entry가 기록될 “자리”를 제공합니다.

### `ledger_accounts`는 어떤 테이블인가?

`ledger_accounts`는 실제 은행 계좌나 블록체인 지갑 주소를 저장하는 테이블이 아닙니다.

StablePay 내부 원장에서 돈의 위치와 역할을 구분하기 위한 계정 테이블입니다.

예:

```text
acct_customer_1
-> 고객이 보유한 USDC 잔액을 기록하는 자리

acct_merchant_pending_1
-> 가맹점에게 아직 지급되기 전의 USDC 금액을 기록하는 자리

acct_platform_fee_1
-> 플랫폼 수수료로 잡힌 USDC 금액을 기록하는 자리
```

그래서 `ledger_accounts`에는 아래 정보가 필요합니다.

| 컬럼 | 의미 |
| --- | --- |
| `id` | 원장 계정의 고유 ID |
| `type` | 계정의 역할, 예: `CUSTOMER`, `MERCHANT_PENDING`, `PLATFORM_FEE` |
| `owner_id` | 이 계정이 누구와 연결되는지 나타내는 ID |
| `currency` | 이 계정이 어떤 통화의 금액을 기록하는지 |
| `created_at` | 계정 생성 시각 |

### `ledger_transactions`는 어떤 테이블인가?

`ledger_transactions`는 여러 Entry를 하나로 묶는 원장 거래 테이블입니다.

여기서 Transaction은 블록체인의 transaction hash와 같은 뜻이 아닙니다.

StablePay 내부 Ledger에서 하나의 비즈니스 사건을 원장 거래로 묶은 기록입니다.

예:

```text
Payment pay_123이 FINALIZED 되었다.
-> 이 사건 때문에 Ledger Transaction led_tx_123이 생성된다.
```

그래서 `ledger_transactions`에는 아래 정보가 필요합니다.

| 컬럼 | 의미 |
| --- | --- |
| `id` | Ledger Transaction의 고유 ID |
| `reference_type` | 어떤 업무에서 발생했는지, 예: `PAYMENT`, `DEPOSIT`, `WITHDRAWAL` |
| `reference_id` | 해당 업무의 ID, 예: `pay_123` |
| `idempotency_key` | 같은 거래가 두 번 저장되지 않게 막는 키 |
| `created_at` | 원장 거래 생성 시각 |

### `ledger_entries`는 어떤 테이블인가?

`ledger_entries`는 실제 돈의 이동 한 줄을 저장합니다.

Ledger에서 가장 중요한 데이터는 이 Entry입니다.

하나의 Ledger Transaction은 여러 Entry를 가질 수 있습니다.

예:

```text
ledger_transactions: led_tx_123

ledger_entries:
  고객 계정             DEBIT  10_000_000 USDC
  가맹점 지급 예정 계정 CREDIT  9_800_000 USDC
  플랫폼 수수료 계정     CREDIT    200_000 USDC
```

그래서 `ledger_entries`에는 아래 정보가 필요합니다.

| 컬럼 | 의미 |
| --- | --- |
| `id` | Entry의 고유 ID |
| `transaction_id` | 이 Entry가 속한 Ledger Transaction ID |
| `account_id` | 이 Entry가 기록되는 Ledger Account ID |
| `direction` | `DEBIT` 또는 `CREDIT` |
| `amount` | 최소 단위 정수 금액 |
| `currency` | 통화 코드, 예: `USDC` |
| `created_at` | Entry 생성 시각 |

### 세 테이블의 관계

아래 관계를 먼저 머릿속에 넣고 SQL을 보면 훨씬 읽기 쉽습니다.

```text
ledger_accounts
  ↑
  │ account_id
  │
ledger_entries
  │
  │ transaction_id
  ↓
ledger_transactions
```

말로 풀면 아래와 같습니다.

```text
ledger_entries는 반드시 하나의 ledger_transactions에 속한다.
ledger_entries는 반드시 하나의 ledger_accounts에 기록된다.
ledger_transactions는 여러 ledger_entries를 묶는다.
ledger_accounts는 여러 ledger_entries가 기록될 수 있는 자리다.
```

## 오늘 하지 않는 것

Day16의 목표는 Go 저장 코드를 만드는 것이 아닙니다.

먼저 저장될 수 있는 DB 구조를 SQL로 정의하는 것입니다.

오늘 하지 않는 것:

```text
Repository 작성
Service와 DB 연결
HTTP API 작성
Payment FINALIZED와 Ledger 자동 연결
Settlement 계산
```

## Migration이란 무엇인가?

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

## up migration과 down migration

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

## Primary Key는 무엇인가?

Primary Key는 테이블 안에서 row 하나를 유일하게 식별하는 값입니다.

예를 들어 `ledger_accounts`의 `id`는 계정 하나를 식별합니다.

```sql
id TEXT PRIMARY KEY
```

Java 객체로 비유하면 `id` 필드와 비슷합니다.

하지만 DB에서는 단순 필드가 아니라 “중복될 수 없는 식별자”라는 제약이 붙습니다.

## Foreign Key는 무엇인가?

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

## Index는 왜 필요한가?

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

## `idempotency_key`는 왜 unique인가?

Idempotency는 같은 요청이나 같은 이벤트가 여러 번 들어와도 결과가 한 번만 반영되는 성질입니다.

Ledger에서는 같은 원장 거래가 두 번 저장되면 위험합니다.

예를 들어 같은 payment finalized 이벤트가 두 번 처리되어 Ledger Transaction이 두 번 저장되면, 가맹점 잔액이 두 번 늘어난 것처럼 보일 수 있습니다.

그래서 `idempotency_key`는 unique index로 막습니다.

```text
같은 idempotency_key를 가진 transaction은 한 번만 저장한다.
```

## `amount`는 왜 BIGINT인가?

돈을 소수점으로 저장하면 부동소수점 오차가 생길 수 있습니다.

그래서 USDC 같은 토큰은 최소 단위 정수로 저장합니다.

예:

```text
10 USDC = 10_000_000
```

큰 정수 금액을 저장하기 위해 PostgreSQL에서는 `BIGINT`를 사용합니다.

Go 코드에서는 `int64`와 대응됩니다.

## 삭제 순서가 중요한 이유

`down.sql`에서는 테이블을 만들 때와 반대 순서로 삭제해야 합니다.

```text
ledger_entries
-> ledger_transactions
-> ledger_accounts
```

이유는 `ledger_entries`가 `ledger_transactions`와 `ledger_accounts`를 참조하기 때문입니다.

참조하는 테이블을 먼저 삭제해야 외래키 관계 때문에 실패하지 않습니다.

## 사전 조건

Day16 실습 전에는 Day12~15가 완료되어 있어야 합니다.

아래 파일이 있어야 합니다.

```text
internal/ledger/ledger.go
internal/ledger/service.go
internal/ledger/service_test.go
migrations/000001_create_payment_core_tables.up.sql
migrations/000001_create_payment_core_tables.down.sql
```

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

## 오늘 만들 파일의 위치

새로 만들 파일:

```text
migrations/000002_create_ledger_core_tables.up.sql
migrations/000002_create_ledger_core_tables.down.sql
```

이미 파일이 있다면 새로 덮어쓰기 전에 내용을 먼저 확인합니다.

```bash
sed -n '1,260p' migrations/000002_create_ledger_core_tables.up.sql
sed -n '1,120p' migrations/000002_create_ledger_core_tables.down.sql
```

아직 없다면 오늘 문서의 SQL을 그대로 작성하면 됩니다.

## Step 1. `up.sql` 작성

파일:

```text
migrations/000002_create_ledger_core_tables.up.sql
```

`up.sql`은 Ledger 테이블과 조회용 인덱스를 생성하는 파일입니다.

먼저 아래 순서로 어떤 테이블이 만들어지는지 이해합니다.

```text
1. ledger_accounts 생성
   -> Entry가 기록될 원장 계정 자리를 먼저 만든다.

2. ledger_accounts 조회용 index 생성
   -> owner_id나 type으로 계정을 빠르게 찾을 수 있게 한다.

3. ledger_transactions 생성
   -> 여러 Entry를 묶는 원장 거래 묶음을 만든다.

4. ledger_transactions 조회/중복방지 index 생성
   -> reference로 거래를 찾고, idempotency_key로 중복 저장을 막는다.

5. ledger_entries 생성
   -> 실제 debit/credit 금액 이동 한 줄을 저장한다.

6. ledger_entries 조회용 index 생성
   -> transaction_id나 account_id로 Entry를 빠르게 찾을 수 있게 한다.
```

생성 순서도 중요합니다.

`ledger_entries`는 `ledger_transactions`와 `ledger_accounts`를 참조합니다.

그래서 참조당하는 테이블인 `ledger_accounts`, `ledger_transactions`를 먼저 만들고,
마지막에 참조하는 테이블인 `ledger_entries`를 만듭니다.

작성할 SQL 전체:

<details>
<summary>up.sql 최종 완성본 전체 보기</summary>

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

</details>

## Step 2. `up.sql` 코드 해석

### `ledger_accounts`

이 테이블은 “돈이 어디에 기록되는가?”에 답합니다.

원장에서 돈이 기록되는 자리입니다.

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

예를 들어 `owner_id = merchant_456`, `type = MERCHANT_PENDING`, `currency = USDC`라면
“merchant_456 가맹점에게 지급 예정인 USDC 금액을 기록하는 원장 계정”으로 이해할 수 있습니다.

### `ledger_transactions`

이 테이블은 “이 돈의 이동은 왜 발생했는가?”에 답합니다.

여러 Entry를 하나의 원장 거래로 묶습니다.

`reference_type`, `reference_id`는 이 원장 거래가 어떤 업무에서 왔는지 알려줍니다.

예:

```text
reference_type = PAYMENT
reference_id   = pay_123
```

`idempotency_key`는 같은 원장 거래가 두 번 저장되지 않게 막는 키입니다.

예를 들어 같은 Payment finalized 이벤트가 두 번 처리되더라도,
같은 `idempotency_key`를 가진 Ledger Transaction은 한 번만 저장되어야 합니다.

### `ledger_entries`

이 테이블은 “실제로 얼마가 어느 계정에 기록되었는가?”에 답합니다.

실제 돈의 이동 한 줄입니다.

```text
고객 계정 DEBIT 10_000_000 USDC
가맹점 계정 CREDIT 9_800_000 USDC
플랫폼 계정 CREDIT 200_000 USDC
```

`transaction_id`는 이 Entry가 어떤 Ledger Transaction에 속하는지 나타냅니다.

`account_id`는 이 Entry가 어떤 Ledger Account에 기록되는지 나타냅니다.

중요한 점은 `ledger_entries` 한 줄이 전체 거래 하나를 의미하지 않는다는 것입니다.

Entry 한 줄은 거래 안의 돈 이동 한 줄입니다.

하나의 결제는 보통 여러 Entry로 나뉩니다.

```text
결제 1건
-> ledger_transactions 1 row
-> ledger_entries 여러 row
```

## Step 3. `down.sql` 작성

파일:

```text
migrations/000002_create_ledger_core_tables.down.sql
```

`down.sql`은 Day16에서 만든 Ledger 테이블을 되돌리는 파일입니다.

삭제는 생성 순서의 반대로 진행합니다.

```text
ledger_entries
-> ledger_transactions
-> ledger_accounts
```

작성할 SQL 전체:

<details>
<summary>down.sql 최종 완성본 전체 보기</summary>

```sql
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS ledger_transactions;
DROP TABLE IF EXISTS ledger_accounts;
```

</details>

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
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -c "\dt ledger_*"
```

예상 결과:

```text
ledger_accounts
ledger_transactions
ledger_entries
```

## Step 7. 컬럼과 인덱스 확인

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
idempotency_key가 unique인가?
```

## Step 8. 롤백 확인

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

## Step 9. Go 테스트 실행

SQL migration만 추가했더라도 기존 Go 코드가 깨지지 않았는지 확인합니다.

```bash
go test ./...
```

이 명령은 Day16에서 작성한 SQL만 확인하는 명령이 아닙니다.

프로젝트 전체 Go 테스트를 실행하므로 Day15에서 정리한 Ledger Service 테스트도 함께 실행됩니다.

Day15 테스트는 `newTestService(t)` helper를 사용합니다.

```go
func newTestService(t *testing.T) (*Service, context.Context) {
	t.Helper()

	return NewService(), context.Background()
}
```

이 helper는 테스트마다 독립적인 `Service`와 `context`를 만들기 위한 테스트 준비 함수입니다.

나중에 `Service`가 Repository나 fake DB를 갖게 되면 아래처럼 helper만 확장하면 됩니다.

```go
func newTestService(t *testing.T) (*Service, *fakeRepository) {
	t.Helper()

	repo := &fakeRepository{}
	svc := NewService(repo)

	return svc, repo
}
```

따라서 `go test ./...`가 성공한다는 것은 단순히 컴파일만 되는 것이 아니라, 기존 Ledger 검증 규칙도 계속 안전하게 유지된다는 뜻입니다.

## Step 10. 자주 만날 수 있는 오류

### `relation "ledger_accounts" already exists`

이미 `ledger_accounts` 테이블이 있는 상태에서 up migration을 다시 실행하면 발생할 수 있습니다.

해결:

```text
1. 실습 DB인지 확인한다.
2. Day16 down migration을 적용한다.
3. 다시 up migration을 적용한다.
```

명령:

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -f migrations/000002_create_ledger_core_tables.down.sql
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -f migrations/000002_create_ledger_core_tables.up.sql
```

### `psql: command not found`

로컬에 PostgreSQL 클라이언트가 설치되어 있지 않을 수 있습니다.

이 경우 Docker 컨테이너 안의 `psql`을 사용하는 방법도 있습니다.

```bash
docker compose exec postgres psql -U stablepay -d stablepay -c "\dt"
```

프로젝트의 `docker-compose.yml` 서비스명이 `postgres`가 아니라면 아래 명령으로 서비스명을 먼저 확인합니다.

```bash
docker compose ps
```

### foreign key 때문에 down migration이 실패하는 경우

테이블 삭제 순서가 잘못되었을 가능성이 큽니다.

아래 순서여야 합니다.

```text
ledger_entries
ledger_transactions
ledger_accounts
```

`ledger_entries`가 다른 두 테이블을 참조하므로 가장 먼저 삭제해야 합니다.

## Step 11. 완성 기준

오늘 작업 후 파일 구조는 아래처럼 보여야 합니다.

```text
migrations/
  000001_create_payment_core_tables.up.sql
  000001_create_payment_core_tables.down.sql
  000002_create_ledger_core_tables.up.sql
  000002_create_ledger_core_tables.down.sql
```

아래를 모두 만족하면 Day16 완료입니다.

```text
up migration 작성 완료
down migration 작성 완료
ledger_* 테이블 생성 확인
ledger_accounts 컬럼 확인
ledger_transactions 컬럼과 unique index 확인
ledger_entries foreign key 확인
rollback 확인
go test ./... 성공
Day16 산출물 5문항 작성
```

## Step 12. 실습산출물 작성

`Day16_실습산출물.md`에는 5개 질문만 답합니다.

```text
1. ledger_accounts, ledger_transactions, ledger_entries는 각각 어떤 역할을 하는가?
2. ledger_entries가 ledger_transactions와 ledger_accounts를 참조해야 하는 이유는 무엇인가?
3. idempotency_key에 unique index가 필요한 이유는 무엇인가?
4. down.sql에서 ledger_entries를 먼저 삭제해야 하는 이유는 무엇인가?
5. 아직 헷갈리는 DB 개념 또는 Ledger 개념은 무엇인가?
```

## Step 13. 커밋 메시지

코드 작업을 완료했다면 아래 커밋 메시지를 사용합니다.

```bash
git status
git add migrations/000002_create_ledger_core_tables.up.sql migrations/000002_create_ledger_core_tables.down.sql
git commit -m "feat: Ledger 핵심 테이블 마이그레이션 추가"
```

산출물 문서를 함께 작성했다면 문서 커밋을 분리하는 것이 좋습니다.

```bash
git add docs/domain/07_Ledger_Core/Day16_Ledger_DB_Migration_작성/Day16_실습산출물.md
git commit -m "docs: Day16 Ledger 마이그레이션 산출물 정리"
```

## 다음 작업 예고

Day16이 끝나면 Day17에서는 Repository 초안으로 넘어갑니다.

```text
Day16: 테이블을 만든다.
Day17: 테이블에 저장하는 Go 코드를 만든다.
```
