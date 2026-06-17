# Day 20 실습산출물 - Ledger Service와 Repository 연결

관련 Jira: [SPN-37](https://aslan0.atlassian.net/browse/SPN-37)

Day20 산출물은 `ValidateTransaction -> CreateTransaction` 흐름을 자기 말로 설명하는 문서입니다.

오늘은 아래 문장을 이해하는 것이 가장 중요합니다.

```text
Service는 검증 후 저장 흐름을 담당하고,
Repository는 실제 DB 저장을 담당한다.
```

## 작성 전 확인

아래 파일을 먼저 확인합니다.

```bash
sed -n '1,260p' internal/ledger/service.go
sed -n '1,360p' internal/ledger/service_test.go
```

Day20 실습을 마친 뒤에는 아래 코드가 보여야 합니다.

```go
type Store interface {
	CreateTransaction(ctx context.Context, tx Transaction, entries []Entry) error
}

func (s *Service) RecordTransaction(ctx context.Context, tx Transaction, entries []Entry) error
```

## 1. Day20에서 Service와 Repository를 연결하는 이유는 무엇인가?

작성 힌트:

```text
검증된 Ledger만 저장되어야 하므로,
Service가 ValidateTransaction을 먼저 실행하고 그 다음 저장소로 넘긴다는 점을 적는다.
```

내 답변:

```text

```

## 2. Service가 `Repository` 구조체가 아니라 `Store` interface에 의존하는 이유는 무엇인가?

작성 힌트:

```text
Service는 저장 동작만 알면 되고,
실제 구현이 PostgreSQL Repository인지 테스트 fake인지 몰라도 된다는 점을 적는다.
```

내 답변:

```text

```

## 3. `fakeStore`는 왜 필요한가?

작성 힌트:

```text
실제 DB 없이 Service가 저장소를 호출했는지, 호출하지 않았는지 확인하기 위해 필요하다는 점을 적는다.
```

내 답변:

```text

```

## 4. 검증 실패 시 저장소가 호출되면 안 되는 이유는 무엇인가?

작성 힌트:

```text
debit/credit 균형이 깨진 Ledger가 DB에 저장되면 원장 자체가 신뢰할 수 없게 된다는 점을 적는다.
```

내 답변:

```text

```

## 5. Day21에서 Payment FINALIZED와 Ledger를 연결할 때 추가로 고민해야 할 것은 무엇인가?

작성 힌트:

```text
Payment에서 어떤 ledger transaction과 entries를 만들지,
idempotency_key를 어떻게 만들지,
어떤 계정에 debit/credit을 줄지 고민해야 한다는 점을 적는다.
```

내 답변:

```text

```

## 오늘 실행 결과

아래 명령 실행 결과를 짧게 기록합니다.

```bash
gofmt -w internal/ledger/service.go internal/ledger/service_test.go
go test ./internal/ledger -v
go test ./...
```

기록:

```text

```

## 아직 헷갈리는 부분

아래 후보 중 헷갈리는 것이 있으면 적습니다.

```text
interface
fakeStore
dependency
Service와 Repository 책임
RecordTransaction 실행 순서
```

메모:

```text

```

## 정답/점검 가이드

먼저 `내 답변`을 작성한 뒤 아래 내용을 펼쳐서 비교합니다.

### 1. Day20에서 Service와 Repository를 연결하는 이유는 무엇인가?

<details>
<summary>정답/점검 가이드 보기</summary>

Service와 Repository를 연결하는 이유는 검증된 Ledger만 저장하기 위해서입니다.

Service는 `ValidateTransaction`으로 debit과 credit 균형, 금액, 통화, 방향을 먼저 검증합니다.

검증을 통과한 뒤에만 Repository의 `CreateTransaction`으로 저장을 위임합니다.

</details>

### 2. Service가 `Repository` 구조체가 아니라 `Store` interface에 의존하는 이유는 무엇인가?

<details>
<summary>정답/점검 가이드 보기</summary>

Service는 PostgreSQL이라는 저장 기술을 알 필요가 없습니다.

Service가 필요한 것은 `CreateTransaction`이라는 저장 동작입니다.

그래서 `Store` interface에 의존하면 실제 운영에서는 Repository를 넣고, 테스트에서는 fakeStore를 넣을 수 있습니다.

</details>

### 3. `fakeStore`는 왜 필요한가?

<details>
<summary>정답/점검 가이드 보기</summary>

fakeStore는 실제 DB 없이 Service 흐름을 테스트하기 위해 필요합니다.

예를 들어 검증 성공 시 저장소가 1번 호출되었는지, 검증 실패 시 저장소가 0번 호출되었는지 확인할 수 있습니다.

즉 fakeStore는 Service 테스트를 빠르고 명확하게 만들어 줍니다.

</details>

### 4. 검증 실패 시 저장소가 호출되면 안 되는 이유는 무엇인가?

<details>
<summary>정답/점검 가이드 보기</summary>

검증 실패는 Ledger가 원장 규칙을 만족하지 못했다는 뜻입니다.

예를 들어 debit 10 USDC인데 credit 9 USDC만 있으면 1 USDC의 출처나 목적이 사라집니다.

이런 데이터가 저장되면 Ledger를 신뢰할 수 없으므로, 검증 실패 시 Repository는 호출되면 안 됩니다.

</details>

### 5. Day21에서 Payment FINALIZED와 Ledger를 연결할 때 추가로 고민해야 할 것은 무엇인가?

<details>
<summary>정답/점검 가이드 보기</summary>

Day21에서는 Payment가 `FINALIZED` 되었을 때 어떤 Ledger Transaction과 Entry를 만들지 정해야 합니다.

특히 아래를 고민해야 합니다.

```text
1. reference_type은 PAYMENT로 둘 것인가?
2. reference_id는 payment id를 사용할 것인가?
3. idempotency_key는 payment:{payment_id}:finalized 형태로 둘 것인가?
4. 고객 계정, 가맹점 지급 예정 계정, 플랫폼 수수료 계정에 어떤 debit/credit을 줄 것인가?
```

</details>

## 추가 보충 정리

### Codex 점검

오늘 산출물에서 가장 중요한 문장은 아래입니다.

```text
RecordTransaction은 검증이 통과된 Ledger만 저장소로 넘긴다.
```

### 코드 확인 포인트

실습이 끝난 뒤 아래 항목을 코드에서 직접 체크합니다.

```text
- Store interface가 service.go에 있는가?
- Service가 store Store 필드를 가지는가?
- RecordTransaction이 ValidateTransaction을 먼저 호출하는가?
- 검증 실패 테스트에서 fakeStore.calls가 0인가?
- 검증 성공 테스트에서 fakeStore.calls가 1인가?
```

### 다음 학습 포인트

Day21에서 특히 이어서 보면 좋은 포인트는 아래입니다.

```text
1. payment 상태 FINALIZED는 어떤 순간인가?
2. payment amount와 fee를 ledger entries로 어떻게 나눌 것인가?
3. payment idempotency_key를 어떻게 만들 것인가?
```
