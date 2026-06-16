package ledger

import "database/sql"

// Repository는 Ledger 데이터를 DB에 저장하고 조회하는 경계이다.
type Repository struct {
	db *sql.DB
}

// NewRepository는 Ledger Repository 인스턴스를 만든다.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}
