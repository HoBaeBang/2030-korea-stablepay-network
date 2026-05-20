package invoice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrMerchantNotFound = errors.New("merchant not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, inv Invoice) (*Invoice, error) {
	const query = `
INSERT INTO invoices (id, merchant_id, amount, currency, status, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, merchant_id, amount, currency, status, expires_at, created_at
`
	var saved Invoice
	var expiresAt sql.NullTime

	err := r.db.QueryRowContext(
		ctx,
		query,
		inv.ID,
		inv.MerchantID,
		inv.Amount,
		inv.Currency,
		inv.Status,
		inv.ExpiresAt,
	).Scan(
		&saved.ID,
		&saved.MerchantID,
		&saved.Amount,
		&saved.Currency,
		&saved.Status,
		&expiresAt,
		&saved.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return nil, ErrMerchantNotFound
		}
		return nil, fmt.Errorf("insert invoice: %w", err)
	}

	if expiresAt.Valid {
		saved.ExpiresAt = &expiresAt.Time
	}

	return &saved, nil
}
