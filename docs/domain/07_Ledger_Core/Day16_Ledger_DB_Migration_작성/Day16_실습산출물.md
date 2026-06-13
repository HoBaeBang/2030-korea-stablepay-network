# Day 16 실습산출물 - Ledger DB Migration 작성

관련 Jira: SPN-33

Day16 산출물은 5개 질문만 작성합니다.

오늘 산출물은 SQL을 외웠는지 확인하는 문서가 아닙니다.

오늘 만든 Ledger 테이블이 어떤 의미를 가지는지 확인하는 문서입니다.

## 작성 전 확인

아래 파일을 먼저 확인합니다.

```bash
ls migrations
```

아래 파일이 보여야 합니다.

```text
000002_create_ledger_core_tables.up.sql
000002_create_ledger_core_tables.down.sql
```

## 1. `ledger_accounts`는 무엇을 저장하는 테이블인가?

작성할 때 볼 파일:

```text
migrations/000002_create_ledger_core_tables.up.sql
```

작성 힌트:

```text
실제 은행 계좌인지,
원장 안에서 돈을 기록하기 위한 계정인지 구분해서 적는다.
```

내 답변:

```text

```

## 2. `ledger_transactions`와 `ledger_entries`는 왜 분리되는가?

작성 힌트:

```text
Transaction은 묶음이고 Entry는 돈의 이동 한 줄이라는 점을 연결해서 적는다.
```

내 답변:

```text

```

## 3. `ledger_entries.transaction_id`와 `ledger_entries.account_id`는 왜 foreign key인가?

작성 힌트:

```text
Entry가 어떤 Transaction에 속하는지,
어떤 Account에 기록되는지 설명한다.
```

내 답변:

```text

```

## 4. `idempotency_key`에 unique index를 둔 이유는 무엇인가?

작성 힌트:

```text
같은 payment finalized 이벤트나 같은 요청이 두 번 처리될 때 어떤 위험이 있는지 적는다.
```

내 답변:

```text

```

## 5. `down.sql`에서 `ledger_entries`를 먼저 삭제하는 이유는 무엇인가?

작성 힌트:

```text
참조하는 테이블과 참조당하는 테이블의 관계를 생각해서 적는다.
```

내 답변:

```text

```

## 오늘 실행 결과

아래 명령 실행 결과를 짧게 기록합니다.

```bash
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -f migrations/000002_create_ledger_core_tables.up.sql
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -c "\\dt ledger_*"
psql "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" -f migrations/000002_create_ledger_core_tables.down.sql
go test ./...
```

기록:

```text

```

## 아직 헷갈리는 부분

아래 후보 중 헷갈리는 것이 있으면 적습니다.

```text
primary key
foreign key
index
unique index
up migration
down migration
BIGINT
TIMESTAMPTZ
```

메모:

```text

```
