package merchant

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

// ErrDuplicateEmail은 같은 email을 가진 merchant가 이미 있을 때 반환한다.
var ErrDuplicateEmail = errors.New("merchant email already exists")

// Repository는 merchant 데이터를 저장하고 조회하는 DB 접근 계층이다.
type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create는 merchants 테이블에 merchant를 저장한다.
func (r *Repository) Create(ctx context.Context, m Merchant) (*Merchant, error) {
	const query = `
		INSERT INTO merchants (id, name, email)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, created_at
		`

	var saved Merchant
	err := r.db.QueryRowContext(ctx, query, m.ID, m.Name, m.Email).Scan(
		&saved.ID,
		&saved.Name,
		&saved.Email,
		&saved.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateEmail
		}
		return nil, fmt.Errorf("insert merchant: %w", err)
	}
	return &saved, nil
}
