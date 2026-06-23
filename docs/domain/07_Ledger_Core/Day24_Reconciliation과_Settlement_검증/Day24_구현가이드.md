# Day 24 구현가이드 - Reconciliation과 Settlement 검증

관련 Jira: [SPN-41](https://aslan0.atlassian.net/browse/SPN-41)

Day23에서는 Settlement Batch와 Items를 PostgreSQL에 저장하고 상태를 안전하게 변경했습니다. Day24에서는 이미 저장된 Settlement 결과를 원본 Ledger와 다시 비교하여 누락, 중복, 금액·통화·수취인 불일치를 발견합니다.

이 문서는 출퇴근 학습과 퇴근 후 실습을 하나로 합친 자료입니다. 위에서 배경을 읽고, `실습 순서`부터 실제 코드를 작성합니다.

## 오늘 만들 것

```text
Reconciliation Snapshot
= 비교에 필요한 Settlement와 Ledger 데이터를 한 번에 담은 조회 결과

Reconciler
= Snapshot을 비교해 Issue를 수집하는 읽기 전용 서비스

Reconciliation Report
= 정상 여부와 발견한 불일치 목록을 담은 결과

Repository JOIN 조회
= Batch, Item, Ledger Entry, Ledger Account를 연결해 Snapshot 생성
```

오늘 새로 작성할 파일:

```text
internal/settlement/reconciliation.go
internal/settlement/reconciliation_repository.go
internal/settlement/reconciliation_test.go
internal/settlement/reconciliation_repository_test.go
```

오늘은 migration, HTTP API, 자동 수정, 정산 상태 변경을 추가하지 않습니다.

## 왜 Reconciliation이 필요한가

Day23의 `CreateBatch`는 저장 시점에 Batch와 Items가 서로 맞는지 검증합니다. 하지만 운영 중에는 다음 문제가 발생할 수 있습니다.

```text
코드 버그로 잘못된 금액 저장
운영자 SQL로 데이터가 수동 변경됨
과거 버전의 계산 규칙과 현재 규칙이 달라짐
외부 지급 시스템 결과와 내부 상태가 달라짐
일부 처리 재시도 과정에서 데이터가 어긋남
```

저장할 때 검증했다고 해서 데이터가 영원히 맞는다고 가정할 수 없습니다. 따라서 일정한 시점에 원본과 결과를 다시 비교해야 합니다.

```text
Validation
= 지금 처리하려는 입력이 규칙에 맞는지 확인

Reconciliation
= 이미 저장되고 처리된 서로 다른 데이터 집합이 지금도 일치하는지 다시 확인
```

우리 프로젝트에서 비교하는 관계:

```text
Ledger
= 돈 이동의 원본 기록, Source of Truth

Settlement
= Ledger 중 지급 가능한 항목을 선택해 만든 정산 결과

Reconciliation
= Settlement 결과가 Ledger 원본과 여전히 맞는지 검사
```

## 전체 흐름

![Day24 Reconciliation 흐름](https://raw.githubusercontent.com/HoBaeBang/2030-korea-stablepay-network/main/docs/confluence/diagrams/spn41-day24-reconciliation-flow.png)

```text
운영자·Scheduler·테스트
-> Reconciler.Check(batchID)
-> Repository.FindReconciliationSnapshot(batchID)
-> Settlement와 Ledger JOIN 조회
-> 금액, 통화, 수취인, 개수, 계정 역할, 방향 비교
-> Healthy Report 또는 Mismatch Report 반환
```

중요한 원칙:

```text
Reconciliation은 발견한다.
자동으로 고치지 않는다.
```

금액 불일치를 발견했다고 어느 쪽을 진실로 보고 수정할지는 별도의 운영 정책과 승인 과정이 필요합니다. 감지 코드가 자동으로 돈 데이터를 수정하면 오히려 원인을 숨기거나 피해를 키울 수 있습니다.

## 핵심 용어

| 영어 | 한글 의미 | 프로젝트에서의 뜻 |
| --- | --- | --- |
| Reconciliation | 대사, 장부 맞추기 | Ledger와 Settlement가 일치하는지 다시 검사 |
| Source of Truth | 신뢰 원본 | 돈 이동의 기준인 Ledger |
| Snapshot | 특정 시점의 묶음 | 한 번의 비교에 사용할 Batch·Item·Ledger 조회 결과 |
| Mismatch | 불일치 | 기대값과 실제값이 다른 상태 |
| Issue | 발견 문제 | 불일치 종류와 기대값·실제값을 담은 결과 |
| Read-only | 읽기 전용 | 조회와 비교만 하고 DB를 변경하지 않음 |

## 어떤 항목을 비교하는가

| 비교 항목 | 기대 기준 | 발견하려는 문제 |
| --- | --- | --- |
| Item 개수 | `batch.item_count == len(items)` | 정산 항목 누락 또는 추가 |
| Batch 총액 | `batch.total_amount == sum(item.amount)` | 요약 총액과 근거 합계 불일치 |
| Ledger 존재 | Item의 `ledger_entry_id`가 Ledger에 존재 | 원본 누락 또는 잘못된 참조 |
| 금액 | `settlement_item.amount == ledger_entry.amount` | 원본과 정산 금액 불일치 |
| 통화 | Batch, Item, Ledger 통화가 같음 | USDC와 다른 통화 혼합 |
| 수취인 | Batch 수취인과 Ledger Account 소유자가 같음 | 다른 사람의 Entry 포함 |
| Account Type | `MERCHANT_PENDING` | 지급 예정이 아닌 계정 포함 |
| Direction | `CREDIT` | 증가 기록이 아닌 Entry 포함 |

## 패키지 책임

```text
Repository
= DB JOIN을 실행하고 비교 가능한 Snapshot을 만든다.

Reconciler
= Snapshot을 비교하고 Issue를 수집한다.

ReconciliationReport
= 검사 결과를 호출자에게 전달한다.
```

Repository에서 `if amount != ...` 같은 업무 비교를 하지 않습니다. Reconciler가 SQL을 직접 실행하지도 않습니다.

## 실습 전 확인

프로젝트 루트에서 실행합니다.

```bash
git status --short
docker compose ps
go test ./internal/settlement
```

PostgreSQL과 기존 Settlement 테이블 확인:

```bash
docker compose exec -T postgres psql -U stablepay -d stablepay -c "\dt"
```

## 실습 순서

## Step 1. Reconciliation 타입과 비교 서비스를 작성한다

파일 위치:

```text
internal/settlement/reconciliation.go
```

먼저 아래 전체 코드를 작성합니다.

<details>
<summary>reconciliation.go 최종 완성본 전체 보기</summary>

```go
package settlement

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"
)

// MismatchType은 Reconciliation에서 발견한 불일치 종류를 나타낸다.
type MismatchType string

const (
	MismatchItemCount      MismatchType = "ITEM_COUNT_MISMATCH"
	MismatchBatchTotal     MismatchType = "BATCH_TOTAL_MISMATCH"
	MismatchLedgerMissing  MismatchType = "LEDGER_ENTRY_MISSING"
	MismatchAmount         MismatchType = "AMOUNT_MISMATCH"
	MismatchCurrency       MismatchType = "CURRENCY_MISMATCH"
	MismatchRecipient      MismatchType = "RECIPIENT_MISMATCH"
	MismatchAccountType    MismatchType = "ACCOUNT_TYPE_MISMATCH"
	MismatchEntryDirection MismatchType = "ENTRY_DIRECTION_MISMATCH"
)

// ReconciliationItem은 Settlement Item과 원본 Ledger 정보를 함께 담은 비교 단위다.
type ReconciliationItem struct {
	LedgerEntryID      string
	SettlementAmount   int64
	SettlementCurrency string
	LedgerEntryFound   bool
	LedgerAmount       int64
	LedgerCurrency     string
	LedgerRecipientID  string
	LedgerAccountType  ledger.AccountType
	LedgerDirection    ledger.EntryDirection
}

// ReconciliationSnapshot은 한 시점에 조회한 Batch와 비교 대상 Item 목록이다.
type ReconciliationSnapshot struct {
	Batch Batch
	Items []ReconciliationItem
}

// ReconciliationIssue는 기대값과 실제값이 다른 한 가지 문제를 나타낸다.
type ReconciliationIssue struct {
	Type          MismatchType
	BatchID       string
	LedgerEntryID string
	Expected      string
	Actual        string
}

// ReconciliationReport는 한 Batch를 점검한 결과다.
type ReconciliationReport struct {
	BatchID          string
	CheckedItemCount int
	Issues           []ReconciliationIssue
}

// IsHealthy는 발견된 불일치가 없는지 알려준다.
func (r ReconciliationReport) IsHealthy() bool {
	return len(r.Issues) == 0
}

// ReconciliationStore는 Reconciler가 비교에 필요한 Snapshot을 조회하는 경계다.
type ReconciliationStore interface {
	FindReconciliationSnapshot(ctx context.Context, batchID string) (*ReconciliationSnapshot, error)
}

// Reconciler는 Settlement와 Ledger 사이의 불일치를 찾는다.
type Reconciler struct {
	store ReconciliationStore
}

// NewReconciler는 Reconciliation 점검기를 만든다.
func NewReconciler(store ReconciliationStore) *Reconciler {
	return &Reconciler{store: store}
}

// Check는 한 Settlement Batch와 원본 Ledger 데이터를 비교해 리포트를 만든다.
func (r *Reconciler) Check(ctx context.Context, batchID string) (*ReconciliationReport, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}

	batchID = strings.TrimSpace(batchID)
	if batchID == "" {
		return nil, fmt.Errorf("settlement batch id가 필요합니다")
	}
	if r.store == nil {
		return nil, fmt.Errorf("reconciliation store가 필요합니다")
	}

	snapshot, err := r.store.FindReconciliationSnapshot(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("reconciliation snapshot 조회 실패: %w", err)
	}
	if snapshot == nil {
		return nil, fmt.Errorf("reconciliation snapshot이 필요합니다")
	}

	report := &ReconciliationReport{
		BatchID:          snapshot.Batch.ID,
		CheckedItemCount: len(snapshot.Items),
		Issues:           make([]ReconciliationIssue, 0),
	}

	if snapshot.Batch.ItemCount != len(snapshot.Items) {
		report.addIssue(
			MismatchItemCount,
			"",
			strconv.Itoa(snapshot.Batch.ItemCount),
			strconv.Itoa(len(snapshot.Items)),
		)
	}

	var itemTotal int64
	for _, item := range snapshot.Items {
		itemTotal += item.SettlementAmount
		r.compareItem(report, snapshot.Batch, item)
	}

	if snapshot.Batch.TotalAmount != itemTotal {
		report.addIssue(
			MismatchBatchTotal,
			"",
			strconv.FormatInt(snapshot.Batch.TotalAmount, 10),
			strconv.FormatInt(itemTotal, 10),
		)
	}

	return report, nil
}

func (r *Reconciler) compareItem(
	report *ReconciliationReport,
	batch Batch,
	item ReconciliationItem,
) {
	if !item.LedgerEntryFound {
		report.addIssue(MismatchLedgerMissing, item.LedgerEntryID, "존재", "없음")
		return
	}

	if item.SettlementAmount != item.LedgerAmount {
		report.addIssue(
			MismatchAmount,
			item.LedgerEntryID,
			strconv.FormatInt(item.LedgerAmount, 10),
			strconv.FormatInt(item.SettlementAmount, 10),
		)
	}

	if item.SettlementCurrency != item.LedgerCurrency || batch.Currency != item.SettlementCurrency {
		report.addIssue(
			MismatchCurrency,
			item.LedgerEntryID,
			batch.Currency+"/"+item.LedgerCurrency,
			item.SettlementCurrency,
		)
	}

	if batch.RecipientID != item.LedgerRecipientID {
		report.addIssue(
			MismatchRecipient,
			item.LedgerEntryID,
			batch.RecipientID,
			item.LedgerRecipientID,
		)
	}

	if item.LedgerAccountType != ledger.AccountTypeMerchantPending {
		report.addIssue(
			MismatchAccountType,
			item.LedgerEntryID,
			string(ledger.AccountTypeMerchantPending),
			string(item.LedgerAccountType),
		)
	}

	if item.LedgerDirection != ledger.EntryDirectionCredit {
		report.addIssue(
			MismatchEntryDirection,
			item.LedgerEntryID,
			string(ledger.EntryDirectionCredit),
			string(item.LedgerDirection),
		)
	}
}

func (r *ReconciliationReport) addIssue(
	mismatchType MismatchType,
	ledgerEntryID string,
	expected string,
	actual string,
) {
	r.Issues = append(r.Issues, ReconciliationIssue{
		Type:          mismatchType,
		BatchID:       r.BatchID,
		LedgerEntryID: ledgerEntryID,
		Expected:      expected,
		Actual:        actual,
	})
}

```

</details>

### 코드 해설

`MismatchType`은 문자열을 아무 곳에서나 직접 작성하지 않도록 불일치 종류를 타입과 상수로 고정합니다.

```text
ITEM_COUNT_MISMATCH
= Batch가 기억하는 개수와 실제 Item 개수가 다름

BATCH_TOTAL_MISMATCH
= Batch 총액과 Item 합계가 다름

LEDGER_ENTRY_MISSING
= Settlement Item이 참조한 Ledger Entry를 찾지 못함

AMOUNT/CURRENCY/RECIPIENT_MISMATCH
= Settlement 결과와 Ledger 원본의 핵심 값이 다름
```

`ReconciliationSnapshot`은 Repository가 조회한 비교 재료이고 `ReconciliationReport`는 Reconciler가 비교한 결과입니다. 입력과 출력을 분리하면 Repository와 비교 규칙을 각각 테스트하기 쉽습니다.

```go
func (r ReconciliationReport) IsHealthy() bool {
    return len(r.Issues) == 0
}
```

Issue가 하나도 없을 때만 정상입니다. 불일치는 프로그램 실행 실패가 아니므로 `error`가 아니라 `Issues`에 담습니다.

```text
error
= DB 연결 실패, context 만료처럼 검사를 수행할 수 없음

Issue
= 검사는 성공했지만 데이터가 서로 다름
```

`Check`는 모든 Item을 검사하면서 Issue를 계속 모읍니다. 첫 번째 불일치에서 즉시 종료하지 않기 때문에 운영자는 한 번의 실행으로 전체 문제를 볼 수 있습니다.

`strconv.FormatInt`는 `int64` 금액을 Report의 `Expected`, `Actual` 문자열에 넣기 위해 사용합니다.

## Step 2. Snapshot 조회 Repository를 작성한다

파일 위치:

```text
internal/settlement/reconciliation_repository.go
```

먼저 아래 전체 코드를 작성합니다.

<details>
<summary>reconciliation_repository.go 최종 완성본 전체 보기</summary>

```go
package settlement

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"
)

// FindReconciliationSnapshot은 Settlement Batch, Items, 원본 Ledger 정보를 함께 조회한다.
func (r *Repository) FindReconciliationSnapshot(
	ctx context.Context,
	batchID string,
) (*ReconciliationSnapshot, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}

	batchID = strings.TrimSpace(batchID)
	if batchID == "" {
		return nil, fmt.Errorf("settlement batch id가 필요합니다")
	}
	if r.db == nil {
		return nil, fmt.Errorf("settlement repository db가 필요합니다")
	}

	const query = `
		SELECT
			sb.id,
			sb.recipient_id,
			sb.currency,
			sb.total_amount,
			sb.item_count,
			sb.status,
			si.ledger_entry_id,
			si.amount,
			si.currency,
			le.id,
			le.amount,
			le.currency,
			la.owner_id,
			la.type,
			le.direction
		FROM settlement_batches sb
		JOIN settlement_items si ON si.batch_id = sb.id
		LEFT JOIN ledger_entries le ON le.id = si.ledger_entry_id
		LEFT JOIN ledger_accounts la ON la.id = le.account_id
		WHERE sb.id = $1
		ORDER BY si.ledger_entry_id
	`

	rows, err := r.db.QueryContext(ctx, query, batchID)
	if err != nil {
		return nil, fmt.Errorf("reconciliation snapshot 조회 실패: %w", err)
	}
	defer rows.Close()

	snapshot := &ReconciliationSnapshot{
		Items: make([]ReconciliationItem, 0),
	}
	initialized := false

	for rows.Next() {
		var batch Batch
		var item ReconciliationItem
		var ledgerEntryID sql.NullString
		var ledgerAmount sql.NullInt64
		var ledgerCurrency sql.NullString
		var ledgerRecipientID sql.NullString
		var ledgerAccountType sql.NullString
		var ledgerDirection sql.NullString

		if err := rows.Scan(
			&batch.ID,
			&batch.RecipientID,
			&batch.Currency,
			&batch.TotalAmount,
			&batch.ItemCount,
			&batch.Status,
			&item.LedgerEntryID,
			&item.SettlementAmount,
			&item.SettlementCurrency,
			&ledgerEntryID,
			&ledgerAmount,
			&ledgerCurrency,
			&ledgerRecipientID,
			&ledgerAccountType,
			&ledgerDirection,
		); err != nil {
			return nil, fmt.Errorf("reconciliation snapshot 변환 실패: %w", err)
		}

		if !initialized {
			snapshot.Batch = batch
			initialized = true
		}

		item.LedgerEntryFound = ledgerEntryID.Valid
		if ledgerAmount.Valid {
			item.LedgerAmount = ledgerAmount.Int64
		}
		if ledgerCurrency.Valid {
			item.LedgerCurrency = ledgerCurrency.String
		}
		if ledgerRecipientID.Valid {
			item.LedgerRecipientID = ledgerRecipientID.String
		}
		if ledgerAccountType.Valid {
			item.LedgerAccountType = ledger.AccountType(ledgerAccountType.String)
		}
		if ledgerDirection.Valid {
			item.LedgerDirection = ledger.EntryDirection(ledgerDirection.String)
		}

		snapshot.Items = append(snapshot.Items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reconciliation snapshot 순회 실패: %w", err)
	}
	if !initialized {
		return nil, fmt.Errorf("settlement batch를 찾을 수 없습니다: %s", batchID)
	}

	return snapshot, nil
}

```

</details>

### SQL이 연결하는 데이터

```sql
FROM settlement_batches sb
JOIN settlement_items si ON si.batch_id = sb.id
LEFT JOIN ledger_entries le ON le.id = si.ledger_entry_id
LEFT JOIN ledger_accounts la ON la.id = le.account_id
```

처리 순서:

```text
1. settlement_batches에서 검사할 Batch를 찾는다.
2. settlement_items로 총액을 구성한 근거를 찾는다.
3. ledger_entries로 각 Item의 원본 돈 이동을 찾는다.
4. ledger_accounts로 원본 Entry의 소유자와 계정 역할을 찾는다.
```

Ledger에는 `LEFT JOIN`을 사용합니다. 정상 DB에서는 Foreign Key 때문에 Ledger Entry 누락이 쉽게 발생하지 않지만, Reconciliation은 데이터 이상을 탐지하는 기능이므로 원본이 없어도 Settlement Item을 조회 결과에서 제거하지 않아야 합니다.

### 왜 `sql.NullString`, `sql.NullInt64`를 사용하는가

`LEFT JOIN` 오른쪽에서 일치하는 Ledger 데이터가 없으면 `le.id`, `le.amount` 등이 SQL `NULL`이 됩니다.

```go
var ledgerEntryID sql.NullString
var ledgerAmount sql.NullInt64
```

일반 `string`, `int64`는 SQL `NULL`을 직접 받을 수 없습니다. `sql.NullString.Valid`가 `false`이면 값이 없다는 뜻입니다.

```go
item.LedgerEntryFound = ledgerEntryID.Valid
```

이 값으로 Reconciler가 `LEDGER_ENTRY_MISSING` Issue를 만듭니다.

### SELECT와 Scan 순서는 반드시 같아야 한다

첫 번째 SELECT 컬럼 `sb.id`는 첫 번째 Scan 대상 `&batch.ID`로 들어갑니다. 컬럼을 추가하거나 순서를 바꾸면 Scan 쪽도 같은 순서로 바꿔야 합니다.

`initialized`는 첫 번째 row를 읽었는지 확인합니다.

```text
false로 끝남
= Batch가 없거나 Item이 없어 JOIN 결과가 한 행도 없음

true
= Batch와 최소 한 개의 Item을 Snapshot으로 구성함
```

## Step 3. Reconciler 단위 테스트를 작성한다

파일 위치:

```text
internal/settlement/reconciliation_test.go
```

먼저 아래 전체 코드를 작성합니다.

<details>
<summary>reconciliation_test.go 최종 완성본 전체 보기</summary>

```go
package settlement

import (
	"context"
	"errors"
	"testing"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger"
)

type fakeReconciliationStore struct {
	snapshot *ReconciliationSnapshot
	err      error
	calls    int
	batchID  string
}

func (f *fakeReconciliationStore) FindReconciliationSnapshot(
	ctx context.Context,
	batchID string,
) (*ReconciliationSnapshot, error) {
	f.calls++
	f.batchID = batchID
	return f.snapshot, f.err
}

func healthyReconciliationSnapshot() *ReconciliationSnapshot {
	return &ReconciliationSnapshot{
		Batch: Batch{
			ID:          "stl_batch_reconciliation_1",
			RecipientID: "merchant_1",
			Currency:    "USDC",
			TotalAmount: 14_800_000,
			ItemCount:   2,
			Status:      StatusReady,
		},
		Items: []ReconciliationItem{
			{
				LedgerEntryID:      "led_entry_1",
				SettlementAmount:   9_800_000,
				SettlementCurrency: "USDC",
				LedgerEntryFound:   true,
				LedgerAmount:       9_800_000,
				LedgerCurrency:     "USDC",
				LedgerRecipientID:  "merchant_1",
				LedgerAccountType:  ledger.AccountTypeMerchantPending,
				LedgerDirection:    ledger.EntryDirectionCredit,
			},
			{
				LedgerEntryID:      "led_entry_2",
				SettlementAmount:   5_000_000,
				SettlementCurrency: "USDC",
				LedgerEntryFound:   true,
				LedgerAmount:       5_000_000,
				LedgerCurrency:     "USDC",
				LedgerRecipientID:  "merchant_1",
				LedgerAccountType:  ledger.AccountTypeMerchantPending,
				LedgerDirection:    ledger.EntryDirectionCredit,
			},
		},
	}
}

func TestReconcilerCheck(t *testing.T) {
	t.Run("Ledger와 Settlement가 일치하면 정상 리포트를 반환한다", func(t *testing.T) {
		store := &fakeReconciliationStore{snapshot: healthyReconciliationSnapshot()}
		reconciler := NewReconciler(store)

		report, err := reconciler.Check(context.Background(), "stl_batch_reconciliation_1")
		if err != nil {
			t.Fatalf("reconciliation이 성공해야 합니다: %v", err)
		}
		if !report.IsHealthy() {
			t.Fatalf("정상 데이터에는 issue가 없어야 합니다: %+v", report.Issues)
		}
		if report.CheckedItemCount != 2 {
			t.Fatalf("검사한 Item은 2개여야 합니다: %d", report.CheckedItemCount)
		}
		if store.calls != 1 || store.batchID != "stl_batch_reconciliation_1" {
			t.Fatal("요청한 Batch ID로 Snapshot을 한 번 조회해야 합니다")
		}
	})

	t.Run("요약과 Ledger 정보가 다르면 모든 불일치를 수집한다", func(t *testing.T) {
		snapshot := healthyReconciliationSnapshot()
		snapshot.Batch.ItemCount = 3
		snapshot.Batch.TotalAmount = 20_000_000
		snapshot.Items[0].SettlementAmount = 9_700_000
		snapshot.Items[0].SettlementCurrency = "KRW"
		snapshot.Items[0].LedgerRecipientID = "merchant_other"
		snapshot.Items[0].LedgerAccountType = ledger.AccountTypePlatformFee
		snapshot.Items[0].LedgerDirection = ledger.EntryDirectionDebit

		store := &fakeReconciliationStore{snapshot: snapshot}
		report, err := NewReconciler(store).Check(context.Background(), snapshot.Batch.ID)
		if err != nil {
			t.Fatalf("불일치는 오류가 아니라 리포트로 반환해야 합니다: %v", err)
		}

		expectedTypes := []MismatchType{
			MismatchItemCount,
			MismatchAmount,
			MismatchCurrency,
			MismatchRecipient,
			MismatchAccountType,
			MismatchEntryDirection,
			MismatchBatchTotal,
		}
		for _, mismatchType := range expectedTypes {
			if !reportHasIssue(report, mismatchType) {
				t.Fatalf("%s issue가 있어야 합니다: %+v", mismatchType, report.Issues)
			}
		}
	})

	t.Run("원본 Ledger Entry가 없으면 누락 issue를 반환한다", func(t *testing.T) {
		snapshot := healthyReconciliationSnapshot()
		snapshot.Items[0].LedgerEntryFound = false

		report, err := NewReconciler(&fakeReconciliationStore{snapshot: snapshot}).Check(
			context.Background(),
			snapshot.Batch.ID,
		)
		if err != nil {
			t.Fatalf("누락은 리포트로 반환해야 합니다: %v", err)
		}
		if !reportHasIssue(report, MismatchLedgerMissing) {
			t.Fatalf("Ledger 누락 issue가 있어야 합니다: %+v", report.Issues)
		}
	})

	t.Run("Snapshot 조회 실패를 전달한다", func(t *testing.T) {
		store := &fakeReconciliationStore{err: errors.New("DB 조회 실패")}
		if _, err := NewReconciler(store).Check(context.Background(), "stl_batch_error"); err == nil {
			t.Fatal("Snapshot 조회 실패는 error를 반환해야 합니다")
		}
	})
}

func reportHasIssue(report *ReconciliationReport, mismatchType MismatchType) bool {
	for _, issue := range report.Issues {
		if issue.Type == mismatchType {
			return true
		}
	}
	return false
}

```

</details>

### 테스트가 확인하는 것

| 테스트 | 검증 내용 |
| --- | --- |
| 정상 Snapshot | Issue가 없고 `IsHealthy()`가 true |
| 여러 불일치 | 첫 오류에서 멈추지 않고 Issue 종류를 모두 수집 |
| Ledger 누락 | 실행 error가 아니라 `LEDGER_ENTRY_MISSING` Issue |
| Store 실패 | Snapshot을 만들 수 없으므로 error 반환 |

`fakeReconciliationStore`는 DB 없이 Reconciler의 비교 규칙만 검사합니다. Repository의 SQL 테스트와 Reconciler의 정책 테스트를 분리하기 위한 구조입니다.

테스트에서 정상 Snapshot을 함수로 만든 뒤 필요한 필드만 바꾸는 이유:

```text
정상 기준 데이터를 한 곳에 둔다.
각 테스트는 어떤 값을 깨뜨렸는지 명확하게 보여준다.
테스트 간에는 매번 새 Snapshot을 받아 상태를 공유하지 않는다.
```

## Step 4. 실제 PostgreSQL 통합 테스트를 작성한다

파일 위치:

```text
internal/settlement/reconciliation_repository_test.go
```

먼저 아래 전체 코드를 작성합니다.

<details>
<summary>reconciliation_repository_test.go 최종 완성본 전체 보기</summary>

```go
package settlement

import "testing"

func TestRepositoryReconciliationFlow(t *testing.T) {
	repository, db, ctx := newTestRepository(t)

	accountID := "acct_reconciliation_1"
	transactionID := "led_tx_reconciliation_1"
	entryID := "led_entry_reconciliation_1"
	batchID := "stl_batch_reconciliation_repository_1"

	seedSettlementCandidate(
		t,
		ctx,
		db,
		accountID,
		"merchant_reconciliation_1",
		transactionID,
		entryID,
		9_800_000,
	)
	t.Cleanup(func() {
		cleanupSettlementFixture(
			t,
			ctx,
			db,
			[]string{batchID},
			[]string{entryID},
			[]string{transactionID},
			[]string{accountID},
		)
	})

	candidates, err := repository.FindCandidates(ctx, "merchant_reconciliation_1", "USDC")
	if err != nil {
		t.Fatalf("정산 후보 조회가 성공해야 합니다: %v", err)
	}
	batch, items, err := NewCalculator().BuildBatch(
		ctx,
		batchID,
		"merchant_reconciliation_1",
		"USDC",
		candidates,
	)
	if err != nil {
		t.Fatalf("정산 계산이 성공해야 합니다: %v", err)
	}
	if err := repository.CreateBatch(ctx, *batch, items); err != nil {
		t.Fatalf("정산 저장이 성공해야 합니다: %v", err)
	}

	t.Run("저장된 Settlement와 Ledger가 일치한다", func(t *testing.T) {
		report, err := NewReconciler(repository).Check(ctx, batchID)
		if err != nil {
			t.Fatalf("reconciliation이 성공해야 합니다: %v", err)
		}
		if !report.IsHealthy() {
			t.Fatalf("정상 데이터에는 issue가 없어야 합니다: %+v", report.Issues)
		}
	})

	t.Run("Settlement Item 금액이 바뀌면 불일치를 찾는다", func(t *testing.T) {
		if _, err := db.ExecContext(
			ctx,
			"UPDATE settlement_items SET amount = $1 WHERE batch_id = $2 AND ledger_entry_id = $3",
			9_700_000,
			batchID,
			entryID,
		); err != nil {
			t.Fatalf("불일치 fixture 생성 실패: %v", err)
		}

		report, err := NewReconciler(repository).Check(ctx, batchID)
		if err != nil {
			t.Fatalf("불일치 점검이 성공해야 합니다: %v", err)
		}
		if !reportHasIssue(report, MismatchAmount) {
			t.Fatalf("금액 불일치를 찾아야 합니다: %+v", report.Issues)
		}
		if !reportHasIssue(report, MismatchBatchTotal) {
			t.Fatalf("Batch 총액 불일치를 찾아야 합니다: %+v", report.Issues)
		}
	})
}

```

</details>

### 통합 테스트 흐름

```text
Ledger Account/Transaction/Entry fixture 생성
-> Candidate 조회
-> Calculator로 Batch/Items 계산
-> Repository로 실제 DB 저장
-> Reconciler 실행: Healthy 확인
-> settlement_items.amount를 테스트에서 일부러 변경
-> Reconciler 재실행: Amount와 Batch Total 불일치 확인
```

실제 서비스 코드가 데이터를 잘못 바꾸는 것이 아닙니다. 테스트가 Reconciliation의 감지 능력을 확인하기 위해 의도적으로 불일치 fixture를 만드는 것입니다.

기존 `repository_test.go`의 helper를 재사용할 수 있는 이유는 두 테스트 파일이 모두 `package settlement`로 작성되어 같은 패키지 안에 있기 때문입니다.

## Step 5. 포맷과 테스트를 실행한다

패키지 전체 포맷:

```bash
go fmt ./internal/settlement
```

Reconciler 단위 테스트:

```bash
go test ./internal/settlement -run TestReconcilerCheck -v
```

실제 PostgreSQL 통합 테스트:

```bash
TEST_DATABASE_URL="postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" \
go test ./internal/settlement -run TestRepositoryReconciliationFlow -v
```

전체 테스트와 정적 검사:

```bash
go test ./...
go vet ./...
```

예상 핵심 결과:

```text
--- PASS: TestReconcilerCheck
--- PASS: TestRepositoryReconciliationFlow
PASS
```

## 자주 만나는 오류

### `TEST_DATABASE_URL이 없어서 ... 건너뜁니다`

단위 테스트만 실행됐고 PostgreSQL 통합 테스트는 실행되지 않은 것입니다. 위의 환경변수가 포함된 명령으로 다시 실행합니다.

### `relation "settlement_batches" does not exist`

Day23 migration이 적용되지 않았습니다.

```bash
docker compose exec -T postgres psql -U stablepay -d stablepay < migrations/000003_create_settlement_tables.up.sql
```

이미 테이블이 있다면 migration을 다시 적용하지 않습니다.

### `sql: Scan error ... converting NULL`

`LEFT JOIN`으로 NULL이 될 수 있는 컬럼을 일반 `string`, `int64`로 Scan했는지 확인합니다. `sql.NullString`, `sql.NullInt64`를 사용해야 합니다.

### 정상 데이터인데 `BATCH_TOTAL_MISMATCH`

`batch.TotalAmount`와 `SettlementAmount` 합계를 비교했는지 확인합니다. LedgerAmount 합계가 아니라 Settlement Item 합계가 Batch 요약의 직접 근거입니다.

### 불일치를 `error`로 반환했다

데이터가 다른 것은 검사가 성공해서 발견한 결과입니다. `ReconciliationIssue`로 반환합니다. DB 조회 실패처럼 검사 자체를 완료하지 못한 경우만 `error`입니다.

## 이번 구현의 한계

```text
한 Batch를 요청받아 즉시 검사한다.
리포트를 DB에 보존하지 않는다.
Scheduler를 아직 연결하지 않는다.
Issue 알림을 아직 보내지 않는다.
불일치를 자동 수정하지 않는다.
```

이 제한은 의도적입니다. 먼저 비교 규칙을 확실하게 만든 뒤, 나중에 Scheduler, Report 저장, 운영 알림을 붙입니다.

## 완성 기준

```text
[ ] reconciliation.go를 작성했다.
[ ] reconciliation_repository.go를 작성했다.
[ ] reconciliation_test.go를 작성했다.
[ ] reconciliation_repository_test.go를 작성했다.
[ ] 정상 Snapshot에서 Healthy Report가 나온다.
[ ] 금액 불일치에서 Amount와 Batch Total Issue가 나온다.
[ ] 실제 PostgreSQL 통합 테스트가 통과한다.
[ ] go test ./...와 go vet ./...가 통과한다.
[ ] Day24 실습산출물을 자기 말로 작성했다.
```

## 커밋 메시지

```bash
git add internal/settlement
git commit -m "feat: settlement reconciliation 검증 추가"
```

## 빠른 복습

### Reconciliation은 왜 자동 수정하지 않는가?

<details>
<summary>답변 보기</summary>

불일치만 보고 어느 데이터가 잘못됐는지 확정할 수 없고 돈 데이터를 자동 수정하면 원인을 숨기거나 피해를 확대할 수 있기 때문입니다. Reconciliation은 발견과 리포트까지만 담당합니다.

</details>

### 데이터 불일치는 왜 `error`가 아니라 `Issue`인가?

<details>
<summary>답변 보기</summary>

검사는 정상적으로 수행됐고 그 결과로 데이터 차이를 발견했기 때문입니다. 조회 실패나 context 만료처럼 검사 자체를 수행하지 못한 경우가 `error`입니다.

</details>

### 다음 작업

Day25에서는 온체인 Deposit 모델과 같은 블록체인 이벤트가 여러 번 읽혀도 한 번만 반영되게 하는 Processed Event 멱등성 기반을 구현합니다.
