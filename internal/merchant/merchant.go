package merchant

import "time"

// Merchant는 StablePay에서 돈을 받을 가맹점 정보를 나타낸다.
type Merchant struct {
	ID        string
	Name      string
	Email     string
	CreatedAt time.Time
}

// CreateMerchantRequest는 merchant 생성을 위해 service가 받는 입력 값이다.
type CreateMerchantRequest struct {
	Name  string
	Email string
}
