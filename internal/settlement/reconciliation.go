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
