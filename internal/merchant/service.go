package merchant

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Store는 Service가 필요로 하는 저장소 동작을 정의한다.
type Store interface {
	Create(ctx context.Context, m Merchant) (*Merchant, error)
}

// Service는 merchant 생성 비즈니스 흐름을 담당한다.
type Service struct {
	store Store
	now   func() time.Time
}

// NewService는 merchant Service를 생성한다.
func NewService(store Store) *Service {
	return &Service{
		store: store,
		now:   time.Now,
	}
}

// CreateMerchant는 merchant 입력 값을 정리하고 저장소에 생성을 요청한다.
func (s *Service) CreateMerchant(ctx context.Context, req CreateMerchantRequest) (*Merchant, error) {
	name := strings.TrimSpace(req.Name)
	email := strings.ToLower(strings.TrimSpace(req.Email))

	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	merchant := Merchant{
		ID:    newMerchantID(s.now()),
		Name:  name,
		Email: email,
	}

	saved, err := s.store.Create(ctx, merchant)
	if err != nil {
		return nil, fmt.Errorf("create merchant: %w", err)
	}

	return saved, nil
}

func newMerchantID(now time.Time) string {
	return fmt.Sprintf("m_%d", now.UnixNano())
}
