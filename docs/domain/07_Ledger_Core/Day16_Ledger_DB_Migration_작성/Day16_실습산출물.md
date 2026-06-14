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

## 정답/점검 가이드

먼저 `내 답변`을 작성한 뒤 아래 내용을 펼쳐서 비교합니다.

### 1. `ledger_accounts`는 무엇을 저장하는 테이블인가?

<details>
<summary>정답/점검 가이드 보기</summary>

`ledger_accounts`는 실제 은행 계좌를 저장하는 테이블이 아닙니다.

StablePay 내부 원장에서 돈의 위치와 역할을 구분하기 위한 계정을 저장하는 테이블입니다.

예를 들어 다음과 같은 계정이 있을 수 있습니다.

- 고객 보유 계정
- 가맹점 지급 예정 계정
- 플랫폼 수수료 계정
- 정산 완료 계정

Ledger Entry는 반드시 어떤 Account에 기록되어야 합니다.

그래야 “이 돈이 누구의 어떤 성격의 잔액으로 기록되었는가?”를 추적할 수 있습니다.

</details>

### 2. `ledger_transactions`와 `ledger_entries`는 왜 분리되는가?

<details>
<summary>정답/점검 가이드 보기</summary>

`ledger_transactions`는 하나의 원장 거래 묶음을 의미합니다.

`ledger_entries`는 그 거래 안에서 실제로 기록되는 돈의 이동 한 줄을 의미합니다.

예를 들어 고객이 10 USDC를 결제하면 하나의 Ledger Transaction 안에 여러 Entry가 생길 수 있습니다.

```text
Transaction: payment pay_123 finalized

Entry 1: 고객 계정 DEBIT 10 USDC
Entry 2: 가맹점 지급 예정 계정 CREDIT 9.8 USDC
Entry 3: 플랫폼 수수료 계정 CREDIT 0.2 USDC
```

거래의 의미는 Transaction에 남기고, 금액의 이동은 Entry에 남기는 구조입니다.

이렇게 분리해야 하나의 비즈니스 사건이 여러 돈의 이동으로 나뉘어도 정확히 기록할 수 있습니다.

</details>

### 3. `ledger_entries.transaction_id`와 `ledger_entries.account_id`는 왜 foreign key인가?

<details>
<summary>정답/점검 가이드 보기</summary>

`ledger_entries.transaction_id`는 이 Entry가 어떤 Ledger Transaction에 속하는지 나타냅니다.

`ledger_entries.account_id`는 이 Entry가 어떤 Ledger Account에 기록되는지 나타냅니다.

둘 다 반드시 실제로 존재하는 대상이어야 하므로 foreign key로 관리합니다.

foreign key가 없으면 다음 문제가 생길 수 있습니다.

- 존재하지 않는 Transaction에 Entry가 연결될 수 있다.
- 존재하지 않는 Account에 금액이 기록될 수 있다.
- 원장 데이터의 신뢰도가 떨어진다.

Ledger는 돈의 근거 데이터이므로 관계가 깨진 데이터를 허용하면 안 됩니다.

</details>

### 4. `idempotency_key`에 unique index를 둔 이유는 무엇인가?

<details>
<summary>정답/점검 가이드 보기</summary>

`idempotency_key`는 같은 요청이나 같은 이벤트가 여러 번 들어와도 한 번만 처리되게 만드는 키입니다.

예를 들어 같은 `payment finalized` 이벤트가 네트워크 재시도나 인덱서 재처리 때문에 두 번 들어올 수 있습니다.

이때 unique index가 없으면 같은 결제가 Ledger에 두 번 기록될 수 있습니다.

그 결과 다음 문제가 생깁니다.

- 가맹점에게 지급 예정 금액이 두 번 잡힌다.
- 플랫폼 수수료가 두 번 기록된다.
- 정산 금액이 틀어진다.

그래서 `idempotency_key`에는 unique index를 두어 같은 원장 거래가 중복 저장되지 않도록 막습니다.

</details>

### 5. `down.sql`에서 `ledger_entries`를 먼저 삭제하는 이유는 무엇인가?

<details>
<summary>정답/점검 가이드 보기</summary>

`ledger_entries`는 `ledger_transactions`와 `ledger_accounts`를 참조합니다.

즉, Entry가 부모 테이블을 바라보고 있는 자식 테이블입니다.

참조 관계가 있는 상태에서 부모 테이블을 먼저 삭제하면 DB는 데이터 무결성이 깨질 수 있다고 판단해 삭제를 막을 수 있습니다.

그래서 rollback에서는 참조하는 쪽인 `ledger_entries`를 먼저 삭제해야 합니다.

삭제 순서는 보통 생성 순서의 반대입니다.

```text
생성 순서:
ledger_accounts
ledger_transactions
ledger_entries

삭제 순서:
ledger_entries
ledger_transactions
ledger_accounts
```

</details>
