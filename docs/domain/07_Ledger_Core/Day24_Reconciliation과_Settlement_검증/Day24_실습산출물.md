# Day 24 실습산출물 - Reconciliation과 Settlement 검증

관련 Jira: [SPN-41](https://aslan0.atlassian.net/browse/SPN-41)

이 문서는 Day24 구현을 마친 뒤 Reconciliation의 목적과 코드 흐름을 자기 말로 정리하는 산출물입니다.

먼저 구현과 테스트를 끝낸 뒤 답변합니다. 힌트와 정답은 막혔을 때만 펼쳐봅니다.

## 1. 저장할 때 검증했는데도 Reconciliation이 다시 필요한 이유는 무엇인가?

<details>
<summary>힌트 보기</summary>

저장 이후 코드 버그, 운영자 수동 변경, 재처리, 외부 시스템 결과 차이로 데이터가 달라질 수 있다는 점을 생각합니다.

</details>

내 답변:

```text

```

## 2. 불일치를 `error`가 아니라 `ReconciliationIssue`로 반환하는 이유는 무엇인가?

<details>
<summary>힌트 보기</summary>

검사를 실행하지 못한 것과 검사는 성공했지만 데이터 차이를 발견한 것을 구분합니다.

</details>

내 답변:

```text

```

## 3. Snapshot 조회 SQL에서 Ledger 테이블을 `LEFT JOIN`하고 `sql.NullString`을 사용하는 이유는 무엇인가?

<details>
<summary>힌트 보기</summary>

원본 Ledger Entry가 없더라도 Settlement Item을 결과에서 유지해야 하는 이유와 SQL NULL을 Go 값으로 받는 방법을 생각합니다.

</details>

내 답변:

```text

```

## 4. `Reconciler.Check`는 Batch와 Ledger 사이에서 어떤 항목들을 비교하는가?

<details>
<summary>힌트 보기</summary>

개수, 총액, 원본 존재 여부, 금액, 통화, 수취인, Account Type, Entry Direction을 확인합니다.

</details>

내 답변:

```text

```

## 5. Reconciliation이 불일치를 발견해도 자동 수정하지 않는 이유는 무엇인가?

<details>
<summary>힌트 보기</summary>

Settlement와 Ledger 중 어느 쪽이 잘못됐는지, 돈 데이터를 고칠 권한과 승인 절차가 필요한지 생각합니다.

</details>

내 답변:

```text

```

## 오늘 실행 결과

실행 명령:

```bash
go fmt ./internal/settlement
go test ./internal/settlement -run TestReconcilerCheck -v
TEST_DATABASE_URL="postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" \
go test ./internal/settlement -run TestRepositoryReconciliationFlow -v
go test ./...
go vet ./...
```

기록:

```text

```

## 오늘 구현한 파일과 동작

실제 코드와 비교해 완료한 항목을 표시합니다.

```text
[ ] internal/settlement/reconciliation.go
[ ] internal/settlement/reconciliation_repository.go
[ ] internal/settlement/reconciliation_test.go
[ ] internal/settlement/reconciliation_repository_test.go
[ ] 정상 Snapshot에서 Healthy Report 확인
[ ] 금액 불일치에서 AMOUNT_MISMATCH 확인
[ ] Item 합계 불일치에서 BATCH_TOTAL_MISMATCH 확인
[ ] 실제 PostgreSQL 통합 테스트 통과
```

## 코드 흐름을 한 문장씩 정리하기

```text
Repository.FindReconciliationSnapshot:

Reconciler.Check:

ReconciliationReport.IsHealthy:

fakeReconciliationStore:

TestRepositoryReconciliationFlow:
```

## 아직 헷갈리는 부분

아래 후보 중 실제로 헷갈리는 내용만 남기거나 직접 추가합니다.

```text
Validation과 Reconciliation 차이
Snapshot이라는 이름의 의미
LEFT JOIN과 sql.NullString
Issue와 error 차이
Repository와 Reconciler 책임
단위 테스트와 통합 테스트 차이
```

메모:

```text

```

## 정답/점검 가이드

답변을 먼저 작성한 뒤 비교합니다.

### 1. Reconciliation이 필요한 이유

<details>
<summary>답변 보기</summary>

저장 시점의 validation은 그 순간의 입력과 규칙만 확인합니다. 저장 이후 코드 버그, 운영자 변경, 재처리, 외부 지급 결과 차이로 데이터가 어긋날 수 있으므로 Ledger 원본과 Settlement 결과를 주기적으로 다시 비교해야 합니다.

</details>

### 2. Issue와 error의 차이

<details>
<summary>답변 보기</summary>

DB 장애나 context 만료는 검사를 끝내지 못한 실행 실패이므로 `error`입니다. 금액이나 통화 불일치는 검사는 성공했고 그 결과로 발견한 업무 데이터 문제이므로 `ReconciliationIssue`에 담습니다.

</details>

### 3. LEFT JOIN과 nullable 타입

<details>
<summary>답변 보기</summary>

원본 Ledger가 없어도 Settlement Item을 조회 결과에 남겨 누락을 발견해야 하므로 Ledger에 `LEFT JOIN`합니다. 오른쪽 데이터가 없으면 SQL NULL이 나오므로 Go에서는 `sql.NullString`, `sql.NullInt64`로 Scan하고 `Valid`를 확인합니다.

</details>

### 4. 비교 항목

<details>
<summary>답변 보기</summary>

Batch의 `item_count`와 실제 Item 개수, Batch 총액과 Item 합계, Ledger Entry 존재 여부, Item과 Ledger의 금액·통화, Batch 수취인과 Ledger Account 소유자, `MERCHANT_PENDING` Account Type, `CREDIT` Direction을 비교합니다.

</details>

### 5. 자동 수정하지 않는 이유

<details>
<summary>답변 보기</summary>

불일치만으로는 Settlement와 Ledger 중 어느 쪽이 잘못됐는지 확정할 수 없습니다. 돈 데이터를 자동 수정하면 원인을 숨기거나 피해를 키울 수 있으므로 Reconciliation은 발견과 리포트까지만 담당하고 수정은 별도 승인·복구 정책으로 처리합니다.

</details>

## 추가 보충 정리

Day24 완료 후 Codex가 아래 내용을 실제 코드와 답변을 기준으로 채웁니다.

### Codex 점검 예정 항목

```text
- Reconciliation을 단순 validation이라고 설명하지 않았는가?
- Ledger를 Source of Truth로 이해했는가?
- 데이터 불일치와 실행 error를 구분했는가?
- Repository가 비교 정책을 담당한다고 설명하지 않았는가?
- 자동 수정이 위험한 이유를 설명했는가?
- 실제 PostgreSQL 통합 테스트가 실행됐는가?
```

### 다음 학습 포인트

Day25에서는 온체인 Deposit과 Processed Event를 구현해 같은 blockchain event를 여러 번 읽어도 Ledger에 한 번만 반영되게 만듭니다.
