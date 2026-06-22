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
batches는 정산의 묶음 요약과 그 묶음의 개별 ledger entry의 내용을 구분한다.
```

Codex 점검:

```text
방향은 맞다. 두 테이블의 책임을 각각 말하면 더 정확하다.

settlement_batches는 "누구에게 어떤 통화로 총 얼마를 어떤 상태로 정산하는가?"에 답한다.
settlement_items는 "그 총액이 어떤 Ledger Entry들로 구성되었는가?"에 답한다.
Batch는 정산 묶음의 요약이고, Item은 계산 근거와 추적성을 보존한다.
```

## 2. `settlement_items.ledger_entry_id`에 UNIQUE 제약이 필요한 이유는 무엇인가?

<details>
<summary>힌트 보기</summary>

같은 Ledger Entry가 서로 다른 Batch ID와 함께 두 번 저장될 가능성을 생각합니다.

</details>

내 답변:

```text
중복된 ledger entry가 들어오게 되면 두 번 저장될 가능성이 있어서
```

Codex 점검:

```text
맞다. 조금 더 정확히는 복합 Primary Key(batch_id, ledger_entry_id)만으로는 같은
ledger_entry_id가 서로 다른 batch_id와 함께 다시 저장될 수 있다.

ledger_entry_id에 별도 UNIQUE 제약을 두면 하나의 Ledger Entry가 전체 Settlement에서
오직 한 Batch에만 포함되도록 DB가 보장한다. 즉, 동일 금액의 중복 정산을 차단한다.
```

## 3. Batch와 Items를 하나의 DB transaction으로 저장해야 하는 이유는 무엇인가?

<details>
<summary>힌트 보기</summary>

Batch 저장 후 두 번째 Item 저장에서 실패했는데 rollback이 없다면 어떤 데이터가 남는지 생각합니다.

</details>

내 답변:

```text
중간에 오류가 발생하면 모두 되돌리지 않으면 추후 중복 저장 등의 이슈가 있을 수 있기 때문에
```

Codex 점검:

```text
rollback이 필요하다는 판단은 맞다. 가장 직접적인 문제는 중복보다 "부분 저장"이다.

예를 들어 Batch 저장은 성공하고 두 번째 Item 저장이 실패하면, transaction이 없을 때는
총액과 item_count를 만족하지 못하는 불완전한 Batch가 DB에 남는다.
하나의 transaction으로 묶으면 모두 성공할 때만 commit하고, 하나라도 실패하면 Batch와
Items를 전부 rollback하여 정합성을 지킨다.
```

## 4. `UPDATE ... WHERE id = $2 AND status = $3`가 동시 상태 변경을 어떻게 막는가?

<details>
<summary>힌트 보기</summary>

두 요청이 모두 DRAFT라고 읽은 뒤 한 요청이 먼저 READY로 바꾼 상황을 생각합니다.

</details>

내 답변:

```text
아직 답변하지 못함
```

Codex 점검:

```text
두 요청이 모두 DRAFT 상태를 읽었더라도 첫 요청이 먼저 DRAFT -> READY 변경에 성공하면
DB의 현재 상태는 READY가 된다.

두 번째 요청이 같은 SQL을 실행할 때 WHERE id = $2 AND status = 'DRAFT' 조건을 만족하는
row가 없으므로 RowsAffected()가 0이 된다. Repository는 이를 오류로 반환한다.

이처럼 "내가 예상한 기존 상태가 아직 유지되는 경우에만 변경"하는 방식을 낙관적 동시성
제어라고 한다. 항상 잠금을 잡는 대신 충돌이 발생했는지를 UPDATE 결과로 확인한다.
```

## 5. Service, Calculator, Repository의 책임은 어떻게 다른가?

<details>
<summary>힌트 보기</summary>

실행 순서 조정, 비즈니스 계산, DB 접근이라는 세 책임을 나눕니다.

</details>

내 답변:

```text
Service는 각 로직의 검증 및 비즈니스 로직에 대한 책임을, Calculator는 settlement의
정산을 위한 계산을, Repository는 DB 저장에 대한 책임을 가진다.
```

Codex 점검:

```text
전체적인 구분은 맞다.

Service는 후보 조회 -> 계산 -> 저장의 실행 순서를 조정하고 상태 전이 정책을 적용한다.
Calculator는 Candidate의 계정 역할, 방향, 통화, 중복 여부를 검증하고 Batch와 Items를 계산한다.
Repository는 PostgreSQL 조회, transaction 저장, 조건부 상태 UPDATE를 담당한다.

Repository가 정산 금액을 계산하지 않고, Calculator가 SQL을 실행하지 않는 것이 핵심이다.
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
go test ./... 통과
go vet ./... 통과
실제 PostgreSQL TestRepositorySettlementFlow 통합 테스트 3개 통과

- 미정산 후보 조회와 Batch/Items 저장
- 동일 Ledger Entry 중복 정산 차단과 Batch rollback
- 예상한 현재 상태일 때만 상태 변경
```

## 오늘 구현한 파일과 메서드

아래 목록을 실제 코드와 비교해 완료한 항목을 표시합니다.

```text
[x] migrations/000003_create_settlement_tables.up.sql
[x] migrations/000003_create_settlement_tables.down.sql
[x] settlement.go 상태 상수 확장
[x] Repository.FindCandidates
[x] Repository.CreateBatch
[x] Repository.UpdateBatchStatus
[x] Service.CreateBatch
[x] Service.TransitionStatus
[x] service_test.go
[x] repository_test.go
```

## 아직 헷갈리는 부분

```text
낙관적 동시성 제어
상태 전이 규칙
```

메모:

```text
낙관적 동시성 제어
상태 전이 규칙
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

### Codex 점검 결과

```text
- settlement_batches와 settlement_items의 큰 역할은 이해했다.
- UNIQUE가 중복 정산을 막는다는 방향은 이해했으며, 복합 PK와의 차이를 보충했다.
- DB transaction과 rollback의 필요성은 이해했으며, 핵심 위험을 부분 저장으로 교정했다.
- Service, Calculator, Repository의 큰 책임 구분은 이해했다.
- 낙관적 동시성 제어와 전체 상태 전이 규칙은 추가 복습이 필요하다.
```

### 코드 검토 결과

```text
- FindCandidates의 JOIN, LEFT JOIN + IS NULL 조건은 미정산 MERCHANT_PENDING CREDIT을 올바르게 조회한다.
- CreateBatch는 Batch와 Items를 하나의 DB transaction으로 저장하고 실패 시 rollback한다.
- ledger_entry_id UNIQUE 제약으로 서로 다른 Batch 사이의 중복 정산을 차단한다.
- UpdateBatchStatus는 예상한 현재 상태를 WHERE 조건에 넣어 충돌을 감지한다.
- Service는 Store interface를 통해 Repository 구현과 분리되어 단위 테스트가 가능하다.
- 전체 테스트, 실제 PostgreSQL 통합 테스트, go vet가 모두 통과했다.
```

### 남은 설계 과제

```text
현재 CreateBatch는 Service와 Calculator가 올바른 Candidate를 전달한다는 전제를 사용한다.
DB에 저장된 Ledger Entry의 실제 recipient, amount, currency와 settlement_items의 값이
계속 일치하는지는 Day24 Reconciliation에서 다시 비교하고 검증한다.
```

### 다음 학습 포인트

Day24에서는 Ledger와 Settlement 상태를 비교해 누락, 중복, 금액 불일치를 찾는 Reconciliation을 구현합니다.
