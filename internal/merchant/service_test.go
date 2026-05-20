package merchant

import (
	"context"
	"testing"
)

type fakeStore struct {
	created Merchant
}

func (f *fakeStore) Create(ctx context.Context, m Merchant) (*Merchant, error) {
	f.created = m
	return &m, nil
}

func TestServiceCreateMerchant(t *testing.T) {
	t.Run("정상 입력하면 이름과 이메일을 정리하고 merchant를 생성한다", func(t *testing.T) {

		store := &fakeStore{}
		service := NewService(store)

		got, err := service.CreateMerchant(context.Background(), CreateMerchantRequest{
			Name:  " Cafe Korea ",
			Email: " OWNER@CAFE.EXAMPLE ",
		})
		if err != nil {
			t.Fatalf("CreateMerchant returned error: %v", err)
		}

		if got.Name != "Cafe Korea" {
			t.Fatalf("expected trimmed name, got %q", got.Name)
		}
		if got.Email != "owner@cafe.example" {
			t.Fatalf("expected normalized email, got %q", got.Email)
		}
		if got.ID == "" {
			t.Fatal("expected merchant id to be generated")
		}
	})

	t.Run("name이 비어 있으면 에러를 반환한다", func(t *testing.T) {
		service := NewService(&fakeStore{})

		_, err := service.CreateMerchant(context.Background(), CreateMerchantRequest{
			Name:  " ",
			Email: "owner@cafe.example",
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
