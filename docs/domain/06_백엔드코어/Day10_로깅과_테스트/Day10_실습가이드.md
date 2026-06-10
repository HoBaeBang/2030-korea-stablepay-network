# Day 10 실습가이드 - 로깅과 테스트 패턴

관련 Jira: [SPN-27](https://aslan0.atlassian.net/browse/SPN-27)

## 실습 흐름

![Day10 로깅과 테스트 피드백 루프](../../../confluence/diagrams/spn27-day10-logging-test-feedback.png)

오늘 실습은 테스트 파일을 읽는 데서 끝나지 않습니다. 어떤 규칙은 테스트로 고정하고, 어떤 사건은 로그로 남겨야 하는지 구분하는 것이 핵심입니다.

## 실습 목표

`Day10_실습산출물.md`에 다음 내용을 작성합니다.

1. 로그가 필요한 이벤트 후보
2. 로그에 포함할 값
3. 로그에 포함하면 안 되는 값
4. 한글 subtest 테스트 패턴
5. Ledger 구현 전 테스트 후보

## Step 1. 기존 테스트 파일 확인

확인 파일:

```text
internal/merchant/service_test.go
internal/invoice/service_test.go
internal/payment/service_test.go
```

확인할 질문:

```text
테스트 함수 이름은 어떻게 되어 있는가?
t.Run을 사용하고 있는가?
given/when/then 흐름이 보이는가?
테스트가 도메인 규칙을 설명하고 있는가?
```

## Step 2. 로그 후보 작성

다음 상황에서 어떤 로그가 필요할지 작성합니다.

```text
payment status changed
invoice created
future ledger transaction created
future indexer duplicate event ignored
future withdrawal signed
```

각 로그 후보에는 다음 값을 같이 생각합니다.

| 질문 | 예시 |
| --- | --- |
| 어떤 일이 발생했는가? | payment status changed |
| 어떤 리소스인가? | payment_id, invoice_id |
| 이전 상태와 다음 상태는 무엇인가? | old_status, new_status |
| 온체인 식별자가 있는가? | tx_hash, chain |
| 사용자가 보면 안 되는 값이 섞였는가? | private key, token |

## Step 3. 로그에 포함할 값 정리

예시:

```text
payment_id
invoice_id
merchant_id
old_status
new_status
tx_hash
chain
```

## Step 4. 로그에 포함하면 안 되는 값 정리

예시:

```text
private key
raw secret
database password
full access token
```

## Step 5. 테스트 패턴 작성

아래 형식으로 테스트 후보를 작성합니다.

```text
테스트 이름:
given:
when:
then:
```

예시:

```text
테스트 이름: ONCHAIN_DETECTED 상태로 바꿀 때 transaction_hash가 없으면 실패한다
given: PENDING 상태의 payment가 존재한다
when: transaction_hash 없이 ONCHAIN_DETECTED로 상태 변경을 요청한다
then: bad_request 계열 에러가 발생하고 상태는 변경되지 않는다
```

Day10에서는 실제 테스트 코드를 완성하지 않아도 됩니다.

하지만 Day11 이후 Ledger 구현에 들어갔을 때 바로 테스트로 옮길 수 있을 만큼 구체적인 테스트 후보를 작성해야 합니다.

## Step 6. 코드 작업 - payment 상태 전이 테스트 추가

Day10의 실제 코드 작업은 테스트를 하나 추가하는 것입니다.

현재 `payment.Service`에는 다음 규칙이 있습니다.

```text
ONCHAIN_DETECTED 상태로 변경하려면 transaction_hash가 필요하다.
```

이 규칙은 서비스 코드에 이미 있지만, 테스트로 고정해두는 것이 좋습니다.

수정할 파일:

```text
internal/payment/service_test.go
```

추가할 위치:

```text
TestService_UpdatePaymentStatus 함수 안
```

추가할 테스트:

```go
t.Run("transaction_hash 없이 ONCHAIN_DETECTED로 변경할 수 없다", func(t *testing.T) {
	store := &fakeStore{current: Payment{ID: "pay_123", Status: StatusPending}}
	service := NewService(store)
	service.now = func() time.Time { return fixedNow }

	_, err := service.UpdatePaymentStatus(context.Background(), UpdatePaymentStatusRequest{
		PaymentID:  "pay_123",
		NextStatus: StatusOnchainDetected,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if store.current.Status != StatusPending {
		t.Fatalf("status = %s, want %s", store.current.Status, StatusPending)
	}
})
```

이 테스트가 확인하는 것:

| 확인 내용 | 의미 |
| --- | --- |
| `err == nil`이면 실패 | transaction_hash 없이 성공하면 안 된다 |
| 상태가 여전히 `PENDING`인지 확인 | 실패한 요청이 상태를 바꾸면 안 된다 |
| 테스트 이름을 한글로 작성 | 실패 시 어떤 규칙이 깨졌는지 바로 읽을 수 있다 |

## Step 7. 테스트 실행

먼저 payment 패키지만 실행합니다.

```bash
go test ./internal/payment
```

문제가 없으면 전체 테스트를 실행합니다.

```bash
go test ./...
```

테스트 파일을 수정했으므로 포맷도 확인합니다.

```bash
gofmt -w internal/payment/service_test.go
```

실행 순서는 보통 아래처럼 하면 됩니다.

```bash
gofmt -w internal/payment/service_test.go
go test ./internal/payment
go test ./...
```

## Step 8. 실습산출물 작성

코드 작업이 끝나면 `Day10_실습산출물.md`에 다음 내용을 작성합니다.

```text
추가한 테스트 이름
given / when / then
이 테스트가 막아주는 버그
로그로 남겨야 할 이벤트 후보
Ledger 구현 전에 추가로 필요한 테스트 후보
```

## Step 9. 완성본 확인

Day10에서 수정하는 파일은 `internal/payment/service_test.go`입니다.

이날 추가하는 코드는 전체 파일을 새로 쓰는 것이 아니라, `TestService_UpdatePaymentStatus` 함수 안에 테스트 케이스를 하나 더 추가하는 것입니다.

아래는 테스트 추가 후 `TestService_UpdatePaymentStatus` 함수가 가져야 하는 최종 형태입니다.

```go
func TestService_UpdatePaymentStatus(t *testing.T) {
	fixedNow := time.Date(2026, 5, 21, 9, 0, 0, 0, time.UTC)

	t.Run("PENDING에서 ONCHAIN_DETECTED로 변경할 수 있다", func(t *testing.T) {
		store := &fakeStore{current: Payment{ID: "pay_123", Status: StatusPending}}
		service := NewService(store)
		service.now = func() time.Time { return fixedNow }

		txHash := "0xabc"
		got, err := service.UpdatePaymentStatus(context.Background(), UpdatePaymentStatusRequest{
			PaymentID:       "pay_123",
			NextStatus:      StatusOnchainDetected,
			TransactionHash: &txHash,
		})
		if err != nil {
			t.Fatalf("UpdatePaymentStatus returned error: %v", err)
		}
		if got.Status != StatusOnchainDetected {
			t.Fatalf("status = %s, want %s", got.Status, StatusOnchainDetected)
		}
		if got.TransactionHash == nil || *got.TransactionHash != txHash {
			t.Fatalf("transaction hash was not saved")
		}
	})

	t.Run("transaction_hash 없이 ONCHAIN_DETECTED로 변경할 수 없다", func(t *testing.T) {
		store := &fakeStore{current: Payment{ID: "pay_123", Status: StatusPending}}
		service := NewService(store)
		service.now = func() time.Time { return fixedNow }

		_, err := service.UpdatePaymentStatus(context.Background(), UpdatePaymentStatusRequest{
			PaymentID:  "pay_123",
			NextStatus: StatusOnchainDetected,
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if store.current.Status != StatusPending {
			t.Fatalf("status = %s, want %s", store.current.Status, StatusPending)
		}
	})

	t.Run("FINALIZED에서 PENDING으로 되돌릴 수 없다", func(t *testing.T) {
		store := &fakeStore{current: Payment{ID: "pay_123", Status: StatusFinalized}}
		service := NewService(store)
		service.now = func() time.Time { return fixedNow }

		_, err := service.UpdatePaymentStatus(context.Background(), UpdatePaymentStatusRequest{
			PaymentID:  "pay_123",
			NextStatus: StatusPending,
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("FINALIZED가 되면 finalized_at을 저장한다", func(t *testing.T) {
		store := &fakeStore{current: Payment{ID: "pay_123", Status: StatusOnchainDetected}}
		service := NewService(store)
		service.now = func() time.Time { return fixedNow }

		got, err := service.UpdatePaymentStatus(context.Background(), UpdatePaymentStatusRequest{
			PaymentID:  "pay_123",
			NextStatus: StatusFinalized,
		})
		if err != nil {
			t.Fatalf("UpdatePaymentStatus returned error: %v", err)
		}
		if got.FinalizedAt == nil {
			t.Fatal("finalized_at is nil")
		}
		if !got.FinalizedAt.Equal(fixedNow) {
			t.Fatalf("finalized_at = %v, want %v", got.FinalizedAt, fixedNow)
		}
	})
}
```

완성본과 비교할 때는 다음을 확인합니다.

```text
새 테스트 이름이 한글로 되어 있는가?
transaction_hash 없이 ONCHAIN_DETECTED 요청을 보내는가?
err가 nil이면 실패하도록 되어 있는가?
실패한 요청 후 상태가 PENDING으로 유지되는지 확인하는가?
```

## Step 10. 커밋 메시지

Day10 코드 실습을 완료하고 테스트까지 통과했다면 아래 커밋 메시지를 사용합니다.

```bash
git status
git add internal/payment/service_test.go docs/domain/06_백엔드코어/Day10_로깅과_테스트/Day10_실습산출물.md
git commit -m "test: 결제 상태 전이 검증 테스트 추가"
```

커밋에 포함할 파일:

```text
internal/payment/service_test.go
docs/domain/06_백엔드코어/Day10_로깅과_테스트/Day10_실습산출물.md
```

## 완료 기준

- [ ] 기존 테스트 구조를 확인했다.
- [ ] 로그가 필요한 이벤트 후보를 작성했다.
- [ ] 로그에 포함할 값과 제외할 값을 구분했다.
- [ ] 한글 subtest 후보를 작성했다.
- [ ] Ledger 구현 전 테스트 후보를 작성했다.
- [ ] `internal/payment/service_test.go`에 상태 전이 테스트를 추가했다.
- [ ] `gofmt -w internal/payment/service_test.go`를 실행했다.
- [ ] `go test ./internal/payment`를 실행했다.
- [ ] `go test ./...`를 실행했다.
- [ ] 완성본과 내 코드를 비교했다.
- [ ] 커밋 메시지를 확인했다.
