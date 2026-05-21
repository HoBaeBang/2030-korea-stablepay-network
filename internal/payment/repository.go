package payment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	// ErrInvoiceNotFound는 존재하지 않는 invoice에 payment를 만들려고 할 때 반환한다.
	ErrInvoiceNotFound = errors.New("invoice not found")

	// ErrPaymentNotFound는 존재하지 않는 payment를 조회하거나 수정하려고 할 때 반환한다.
	ErrPaymentNotFound = errors.New("payment not found")
)

// Repository는 payment 데이터를 PostgreSQL에 저장하고 조회하는 DB 접근 계층이다.
type Repository struct {
	// db는 database/sql의 DB connection pool이다.
	// *sql.DB는 단일 연결 하나가 아니라 여러 DB 연결을 관리하는 객체라고 이해하면 된다.
	db *sql.DB
}

// NewRepository는 Repository를 생성한다.
// 반환 타입이 *Repository이므로 Repository 값의 주소를 반환한다.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create는 새 payment를 DB에 저장한다.
func (r *Repository) Create(ctx context.Context, p Payment) (*Payment, error) {
	const query = `
INSERT INTO payments (id, invoice_id, amount, currency, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, invoice_id, amount, currency, status, transaction_hash, finalized_at, created_at
`

	// QueryRowContext는 SQL을 실행하고 결과 row 하나를 반환한다.
	// scanPayment는 그 row를 Payment 구조체로 변환한다.
	saved, err := scanPayment(r.db.QueryRowContext(
		ctx,
		query,
		p.ID,
		p.InvoiceID,
		p.Amount,
		p.Currency,
		p.Status,
	))
	if err != nil {
		var pgErr *pgconn.PgError

		// PostgreSQL error code 23503은 foreign key violation이다.
		// 즉 payments.invoice_id가 invoices.id를 참조하지 못했다는 뜻이다.
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, ErrInvoiceNotFound
		}

		return nil, fmt.Errorf("insert payment: %w", err)
	}

	return saved, nil
}

// FindByID는 payment id로 payment 한 건을 조회한다.
func (r *Repository) FindByID(ctx context.Context, id string) (*Payment, error) {
	const query = `
SELECT id, invoice_id, amount, currency, status, transaction_hash, finalized_at, created_at
FROM payments
WHERE id = $1
`

	p, err := scanPayment(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		// sql.ErrNoRows는 SELECT 결과가 한 건도 없을 때 발생한다.
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}

		return nil, fmt.Errorf("select payment: %w", err)
	}

	return p, nil
}

// StatusUpdate는 payment 상태 변경에 필요한 DB update 값을 모아둔 타입이다.
type StatusUpdate struct {
	PaymentID  string
	NextStatus Status

	// nil이면 기존 transaction_hash를 유지한다.
	TransactionHash *string

	// nil이면 기존 finalized_at을 유지한다.
	FinalizedAt *time.Time
}

// UpdateStatus는 payment의 상태와 상태 변경에 필요한 부가 정보를 수정한다.
func (r *Repository) UpdateStatus(ctx context.Context, update StatusUpdate) (*Payment, error) {
	const query = `
UPDATE payments
SET status = $2,
    transaction_hash = COALESCE($3, transaction_hash),
    finalized_at = COALESCE($4, finalized_at)
WHERE id = $1
RETURNING id, invoice_id, amount, currency, status, transaction_hash, finalized_at, created_at
`

	// COALESCE($3, transaction_hash)는 $3이 NULL이면 기존 transaction_hash를 유지한다는 뜻이다.
	// Go에서 update.TransactionHash가 nil이면 DB에는 NULL처럼 전달된다.
	p, err := scanPayment(r.db.QueryRowContext(
		ctx,
		query,
		update.PaymentID,
		update.NextStatus,
		update.TransactionHash,
		update.FinalizedAt,
	))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}

		return nil, fmt.Errorf("update payment status: %w", err)
	}

	return p, nil
}

// rowScanner는 Scan method만 가진 작은 interface다.
// *sql.Row와 *sql.Rows 모두 Scan method를 가지고 있어서 scanPayment를 재사용할 수 있다.
type rowScanner interface {
	Scan(dest ...any) error
}

// scanPayment는 DB row를 Payment 구조체로 변환한다.
// Create, FindByID, UpdateStatus에서 같은 Scan 코드를 반복하지 않기 위해 분리했다.
func scanPayment(row rowScanner) (*Payment, error) {
	var p Payment

	// transaction_hash와 finalized_at은 DB에서 NULL일 수 있다.
	// sql.NullString/sql.NullTime은 DB NULL 여부를 Valid 필드로 알려준다.
	var transactionHash sql.NullString
	var finalizedAt sql.NullTime

	err := row.Scan(
		&p.ID,
		&p.InvoiceID,
		&p.Amount,
		&p.Currency,
		&p.Status,
		&transactionHash,
		&finalizedAt,
		&p.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Valid가 true일 때만 실제 값이 있다는 뜻이다.
	// &transactionHash.String은 String 필드의 주소를 p.TransactionHash에 넣는 것이다.
	if transactionHash.Valid {
		p.TransactionHash = &transactionHash.String
	}
	if finalizedAt.Valid {
		p.FinalizedAt = &finalizedAt.Time
	}

	return &p, nil
}
