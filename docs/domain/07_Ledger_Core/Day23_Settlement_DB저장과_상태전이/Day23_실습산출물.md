# Day 23 실습산출물 - Settlement DB 저장과 상태 전이

관련 Jira: [SPN-40](https://aslan0.atlassian.net/browse/SPN-40)

이 문서는 Day23 구현을 마친 뒤 DB 저장과 상태 전이 흐름을 자기 말로 설명하는 산출물입니다.

힌트는 기본적으로 접혀 있습니다. 먼저 답변을 작성하고 막힐 때만 펼쳐봅니다.

## 1. `settlement_batches`와 `settlement_items`는 각각 어떤 질문에 답하는가?

<details>
<summary>힌트 보기</summary>

정산 묶음 전체의 요약과 그 총액을 구성한 개별 Ledger Entry 근거를 구분합니다.

</details>

내 답변:

```text

```

## 2. `settlement_items.ledger_entry_id`에 UNIQUE 제약이 필요한 이유는 무엇인가?

<details>
<summary>힌트 보기</summary>

같은 Ledger Entry가 서로 다른 Batch ID와 함께 두 번 저장될 가능성을 생각합니다.

</details>

내 답변:

```text

```

## 3. Batch와 Items를 하나의 DB transaction으로 저장해야 하는 이유는 무엇인가?

<details>
<summary>힌트 보기</summary>

Batch 저장 후 두 번째 Item 저장에서 실패했는데 rollback이 없다면 어떤 데이터가 남는지 생각합니다.

</details>

내 답변:

```text

```

## 4. `UPDATE ... WHERE id = $2 AND status = $3`가 동시 상태 변경을 어떻게 막는가?

<details>
<summary>힌트 보기</summary>

두 요청이 모두 DRAFT라고 읽은 뒤 한 요청이 먼저 READY로 바꾼 상황을 생각합니다.

</details>

내 답변:

```text

```

## 5. Service, Calculator, Repository의 책임은 어떻게 다른가?

<details>
<summary>힌트 보기</summary>

실행 순서 조정, 비즈니스 계산, DB 접근이라는 세 책임을 나눕니다.

</details>

내 답변:

```text

```

## 오늘 실행 결과

```bash
gofmt -w ./internal/settlement
go test ./internal/settlement -run 'TestCalculator|TestService|Example_statusFlow' -v
TEST_DATABASE_URL="postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" go test ./internal/settlement -run TestRepositorySettlementFlow -v
go test ./...
```

기록:

```text

```

## 오늘 구현한 파일과 메서드

아래 목록을 실제 코드와 비교해 완료한 항목을 표시합니다.

```text
[ ] migrations/000003_create_settlement_tables.up.sql
[ ] migrations/000003_create_settlement_tables.down.sql
[ ] settlement.go 상태 상수 확장
[ ] Repository.FindCandidates
[ ] Repository.CreateBatch
[ ] Repository.UpdateBatchStatus
[ ] Service.CreateBatch
[ ] Service.TransitionStatus
[ ] service_test.go
[ ] repository_test.go
```

## 아직 헷갈리는 부분

```text
LEFT JOIN과 IS NULL
UNIQUE와 PRIMARY KEY 차이
DB transaction과 rollback
낙관적 동시성 제어
상태 전이 규칙
Store interface와 Repository 구현 관계
```

메모:

```text

```

## 정답/점검 가이드

먼저 답변을 작성한 뒤 필요한 항목만 펼쳐서 비교합니다.

### 1. 두 Settlement 테이블의 책임

<details>
<summary>답변 보기</summary>

`settlement_batches`는 수취인, 통화, 총액, 항목 수, 상태를 가진 정산 묶음의 요약입니다.

`settlement_items`는 그 Batch에 포함된 개별 Ledger Entry와 금액을 연결해 총액의 계산 근거를 보존합니다.

</details>

### 2. `ledger_entry_id` UNIQUE

<details>
<summary>답변 보기</summary>

복합 Primary Key만으로는 동일한 Ledger Entry가 다른 Batch ID와 함께 다시 저장될 수 있습니다. `ledger_entry_id` UNIQUE는 Ledger Entry 하나가 전체 Settlement 시스템에서 한 번만 정산되도록 DB에서 보장합니다.

</details>

### 3. 하나의 DB transaction

<details>
<summary>답변 보기</summary>

Batch와 일부 Item만 저장된 불완전한 정산을 막기 위해서입니다. 모든 INSERT가 성공하면 commit하고 하나라도 실패하면 Batch와 Items를 모두 rollback합니다.

</details>

### 4. 현재 상태 조건부 UPDATE

<details>
<summary>답변 보기</summary>

첫 요청이 DRAFT를 READY로 변경하면 두 번째 요청의 `WHERE status = 'DRAFT'` 조건은 더 이상 맞지 않아 변경 row가 0개가 됩니다. 그래서 앞선 상태 변경을 조용히 덮어쓰지 못합니다.

</details>

### 5. 세 구성 요소의 책임

<details>
<summary>답변 보기</summary>

Service는 후보 조회, 계산, 저장 순서와 상태 전이 정책을 조정합니다. Calculator는 후보 검증과 금액 계산을 담당합니다. Repository는 PostgreSQL 조회와 저장만 담당합니다.

</details>

## 추가 보충 정리

### Codex 점검 예정 항목

```text
- 테이블 관계와 unique 제약을 자기 말로 설명했는가?
- 상태와 AccountType, EntryDirection을 혼동하지 않았는가?
- Repository가 계산한다고 설명하지 않았는가?
- transaction과 상태 조건부 UPDATE의 목적을 설명했는가?
- 실제 코드와 통합 테스트 결과가 산출물 답변과 일치하는가?
```

### 다음 학습 포인트

Day24에서는 Ledger와 Settlement 상태를 비교해 누락, 중복, 금액 불일치를 찾는 Reconciliation을 구현합니다.
