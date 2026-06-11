# Day 10 실습가이드 - 로깅과 테스트 패턴

관련 Jira: [SPN-27](https://aslan0.atlassian.net/browse/SPN-27)

## 오늘의 위치

Day10은 서비스 전체를 전부 이해했다고 가정하고 진행하는 날이 아닙니다.

오늘은 `payment.Service.UpdatePaymentStatus` 테스트 하나를 추가하면서, 다음 감각을 익히는 날입니다.

```text
1. 결제 상태 변경에는 규칙이 있다.
2. 그 규칙은 테스트로 고정해야 한다.
3. 운영 중 추적해야 하는 사건은 로그 후보로 정리해야 한다.
4. 아직 구현하지 않은 Ledger/Indexer/Withdrawal은 "미래에 테스트가 필요할 후보" 정도로만 정리한다.
```

## 실습 흐름

![Day10 로깅과 테스트 피드백 루프](../../../confluence/diagrams/spn27-day10-logging-test-feedback.png)

오늘 코드 실습에서 따라갈 실제 흐름은 아래 하나입니다.

```text
테스트 코드
  -> payment.Service.UpdatePaymentStatus 호출
  -> 기존 payment 조회
  -> 상태 전이 가능 여부 확인
  -> ONCHAIN_DETECTED에는 transaction_hash가 있는지 확인
  -> 성공하면 상태 저장, 실패하면 상태를 바꾸지 않음
```

즉, 오늘은 "결제 상태를 바꿀 때 transaction_hash가 필요한 경우가 있다"는 규칙을 테스트로 고정합니다.

## 오늘 실습과 산출물의 연결

아래 표가 Day10에서 가장 중요합니다. 실습가이드와 산출물이 따로 노는 느낌이 들면 이 표를 기준으로 작성하면 됩니다.

| 실습에서 하는 일 | 보는 파일/함수 | 산출물에 적는 위치 | 적어야 하는 내용 |
| --- | --- | --- | --- |
| 기존 테스트 구조 읽기 | `internal/payment/service_test.go` | `1. 기존 테스트 구조 관찰` | 어떤 테스트가 어떤 규칙을 검증하는지 |
| 테스트 하나 추가 | `TestService_UpdatePaymentStatus` | `4. 한글 subtest 후보`, `5. given/when/then` | 추가한 테스트 이름과 조건/행동/결과 |
| 실패 시 상태가 바뀌지 않는지 확인 | `store.current.Status != StatusPending` | `5. given/when/then` | 실패한 요청은 기존 상태를 유지해야 함 |
| 운영 로그 후보 생각 | 오늘 추가한 실패 케이스 기준 | `2. 로그가 필요한 이벤트 후보` | 상태 변경 거절, 상태 변경 성공 같은 추적 이벤트 |
| 비밀값 구분 | 로그 후보 작성 중 | `3. 로그에 포함하면 안 되는 값` | private key, access token, DB password 등 |
| 미래 Ledger 테스트 생각 | 아직 구현 전 | `6. Ledger 구현 전 필요한 테스트 후보` | 중복 반영, 금액 불일치 같은 미래 위험 |

## 실습 목표

Day10을 마치면 아래 질문에 답할 수 있어야 합니다.

```text
왜 ONCHAIN_DETECTED 상태에는 transaction_hash가 필요한가?
왜 실패한 상태 변경 요청이 payment 상태를 바꾸면 안 되는가?
왜 테스트 이름을 한글로 써도 되는가?
로그는 사용자 응답과 무엇이 다른가?
Ledger 구현 전에는 어떤 위험을 테스트 후보로 잡아야 하는가?
```

## Step 1. 기존 테스트 파일 확인

확인 파일:

```text
internal/merchant/service_test.go
internal/invoice/service_test.go
internal/payment/service_test.go
```

오늘 가장 집중해서 볼 파일은 아래입니다.

```text
internal/payment/service_test.go
```

확인할 질문:

```text
테스트 함수 이름은 어떻게 되어 있는가?
t.Run을 사용하고 있는가?
테스트 이름이 도메인 규칙을 설명하고 있는가?
실패해야 하는 케이스에서 err를 확인하는가?
실패 후 상태가 바뀌지 않았는지 확인하는가?
```

여기서 `t.Run`은 Java의 `@DisplayName`이 붙은 세부 테스트 케이스처럼 생각하면 됩니다.

```go
t.Run("FINALIZED에서 PENDING으로 되돌릴 수 없다", func(t *testing.T) {
    // 이 이름만 읽어도 어떤 규칙을 검증하는지 알 수 있다.
})
```

## Step 2. 오늘 코드에서 이해해야 하는 결제 상태 흐름

현재 프로젝트의 payment 상태는 대략 아래 흐름을 가집니다.

```text
PENDING
  -> ONCHAIN_DETECTED
  -> FINALIZED
```

각 상태의 의미:

| 상태 | 의미 |
| --- | --- |
| `PENDING` | 결제는 만들어졌지만 아직 온체인 거래가 감지되지 않았다 |
| `ONCHAIN_DETECTED` | 블록체인에서 관련 transaction을 감지했다 |
| `FINALIZED` | 결제가 최종 확정되었다 |

오늘 추가하는 규칙:

```text
PENDING -> ONCHAIN_DETECTED 로 바꾸려면 transaction_hash가 필요하다.
```

왜 필요할까요?

`ONCHAIN_DETECTED`는 "블록체인에서 거래를 봤다"는 뜻입니다. 그런데 어떤 거래를 봤는지 나타내는 `transaction_hash`가 없으면, 나중에 장애가 났을 때 실제 온체인 거래를 추적할 수 없습니다.

## Step 3. 코드 작업 - payment 상태 전이 테스트 추가

수정할 파일:

```text
internal/payment/service_test.go
```

추가할 위치:

```text
TestService_UpdatePaymentStatus 함수 안
```

권장 위치:

```text
"PENDING에서 ONCHAIN_DETECTED로 변경할 수 있다" 테스트 바로 아래
```

이유는 성공 케이스와 실패 케이스를 붙여두면 읽기 쉽기 때문입니다.

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

## Step 4. 추가한 테스트 해석

| 코드 | 의미 |
| --- | --- |
| `store := &fakeStore{...}` | 실제 DB 대신 테스트용 저장소를 만든다 |
| `Status: StatusPending` | 시작 상태는 `PENDING`이다 |
| `NextStatus: StatusOnchainDetected` | `ONCHAIN_DETECTED`로 바꾸려고 한다 |
| `TransactionHash`를 넣지 않음 | 일부러 transaction_hash가 없는 실패 케이스를 만든다 |
| `if err == nil` | 에러가 없으면 규칙이 깨진 것이므로 테스트 실패 |
| `store.current.Status != StatusPending` | 실패한 요청이 상태를 바꿨다면 위험하므로 테스트 실패 |

여기서 포인터 관점으로도 하나 짚고 갑니다.

`TransactionHash`는 `*string` 타입입니다. 즉, 값이 있을 수도 있고 없을 수도 있습니다.

```go
TransactionHash: &txHash // 값이 있음
TransactionHash: nil     // 값이 없음
```

오늘 실패 테스트에서는 `TransactionHash`를 아예 넣지 않기 때문에 기본값 `nil`이 됩니다.

## Step 5. 테스트 실행

먼저 payment 패키지만 실행합니다.

```bash
go test ./internal/payment
```

문제가 없으면 전체 테스트를 실행합니다.

```bash
go test ./...
```

테스트 파일을 수정했으므로 포맷도 맞춥니다.

```bash
gofmt -w internal/payment/service_test.go
```

실행 순서는 보통 아래처럼 하면 됩니다.

```bash
gofmt -w internal/payment/service_test.go
go test ./internal/payment
go test ./...
```

참고로 여러 Go 파일을 한 번에 포맷하려면 아래처럼 전체 패키지를 대상으로 실행할 수도 있습니다.

```bash
go fmt ./...
```

## Step 6. 실습산출물 작성 방법

코드 작업이 끝나면 `Day10_실습산출물.md`를 작성합니다.

오늘 산출물은 아래 순서대로 채우면 됩니다.

```text
1. 기존 테스트 구조 관찰
   -> internal/payment/service_test.go에서 TestService_UpdatePaymentStatus를 보고 작성한다.

2. 로그가 필요한 이벤트 후보
   -> 오늘 추가한 실패 케이스를 기준으로 "상태 변경 거절" 로그 후보를 작성한다.

3. 로그에 포함하면 안 되는 값
   -> private key, access token처럼 로그에 남기면 안 되는 값을 작성한다.

4. 한글 subtest 후보
   -> 오늘 추가한 테스트 이름을 그대로 적고, 비슷한 후보를 1~2개 더 생각한다.

5. given / when / then
   -> 오늘 추가한 테스트를 given/when/then으로 풀어 쓴다.

6. Ledger 구현 전 필요한 테스트 후보
   -> 지금 구현하지 않는다. 미래에 돈 기록을 만들 때 막아야 할 위험만 적는다.
```

## Step 7. 산출물 작성 예시

아래 예시는 정답을 외우라는 뜻이 아니라, 어떤 방향으로 쓰면 되는지 보여주는 기준입니다.

```text
테스트 이름: transaction_hash 없이 ONCHAIN_DETECTED로 변경할 수 없다

given: PENDING 상태의 payment가 있다.
when: transaction_hash 없이 ONCHAIN_DETECTED 상태 변경을 요청한다.
then: 에러가 발생하고 payment 상태는 PENDING으로 유지된다.

이 테스트가 막아주는 버그:
블록체인 거래 식별자 없이 온체인 감지 상태로 바뀌어 나중에 어떤 거래를 근거로 상태가 바뀌었는지 추적하지 못하는 문제를 막는다.
```

로그 후보 예시:

```text
이벤트: payment status transition rejected
이유: 잘못된 상태 변경 요청이 들어왔을 때 어떤 payment에서 어떤 규칙 때문에 거절되었는지 추적하기 위해 필요하다.
포함할 값: payment_id, current_status, requested_status, reason
포함하면 안 되는 값: private key, full access token, database password
```

## Step 8. 완성본 확인

Day10에서 수정하는 코드는 `internal/payment/service_test.go`의 `TestService_UpdatePaymentStatus` 함수 안에 테스트 케이스를 하나 추가하는 것입니다.

완성본과 비교할 때는 다음을 확인합니다.

```text
새 테스트 이름이 한글로 되어 있는가?
transaction_hash 없이 ONCHAIN_DETECTED 요청을 보내는가?
err가 nil이면 실패하도록 되어 있는가?
실패한 요청 후 상태가 PENDING으로 유지되는지 확인하는가?
가능하면 성공 케이스 바로 아래에 실패 케이스가 배치되어 있는가?
```

테스트 이름 끝의 마침표는 기능상 문제는 없지만, 기존 테스트 이름과 맞추려면 보통 빼는 편이 깔끔합니다.

```text
권장: transaction_hash 없이 ONCHAIN_DETECTED로 변경할 수 없다
비권장: transaction_hash 없이 ONCHAIN_DETECTED로 변경할 수 없다.
```

## Step 9. 커밋 메시지

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
- [ ] `internal/payment/service_test.go`에 상태 전이 테스트를 추가했다.
- [ ] 추가한 테스트를 given/when/then으로 설명했다.
- [ ] 실패한 상태 변경 요청이 기존 상태를 바꾸면 안 되는 이유를 설명했다.
- [ ] 로그가 필요한 이벤트 후보를 오늘 테스트 기준으로 작성했다.
- [ ] 로그에 포함할 값과 제외할 값을 구분했다.
- [ ] Ledger 구현 전 테스트 후보를 "미래 위험" 기준으로 작성했다.
- [ ] `gofmt -w internal/payment/service_test.go` 또는 `go fmt ./...`를 실행했다.
- [ ] `go test ./internal/payment`를 실행했다.
- [ ] `go test ./...`를 실행했다.
- [ ] 커밋 메시지를 확인했다.
