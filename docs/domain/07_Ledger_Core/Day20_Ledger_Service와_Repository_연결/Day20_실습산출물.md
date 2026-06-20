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
검증에 성공한 Ledger만 DB에 영구적으로 저장되도록 연결하기 위해서다.
Service는 ValidateTransaction으로 원장 규칙을 먼저 확인하고, 성공한 Transaction과 Entries만 Store를 통해 Repository에 전달한다.
```

## 2. Service가 `Repository` 구조체가 아니라 `Store` interface에 의존하는 이유는 무엇인가?

작성 힌트:

```text
Service는 저장 동작만 알면 되고,
실제 구현이 PostgreSQL Repository인지 테스트 fake인지 몰라도 된다는 점을 적는다.
```

내 답변:

```text
Service의 책임은 DB에 직접 저장하는 것이 아니라 검증 후 저장을 요청하는 업무 흐름을 실행하는 것이다.
Store interface에 의존하면 PostgreSQL Repository와 결합되지 않아 구현을 교체하기 쉽고, 테스트에서는 fakeStore를 주입할 수 있어 유지보수성·확장성·테스트 가능성이 좋아진다.
```

## 3. `fakeStore`는 왜 필요한가?

작성 힌트:

```text
실제 DB 없이 Service가 저장소를 호출했는지, 호출하지 않았는지 확인하기 위해 필요하다는 점을 적는다.
```

내 답변:

```text
실제 DB 없이 Service가 저장소를 호출했는지 확인하기 위해 필요하다.
fakeStore는 호출 횟수와 전달받은 Transaction·Entries를 기록하고, 원하는 저장 오류도 반환할 수 있어 Service 흐름만 독립적으로 검증할 수 있다.
```

## 4. 검증 실패 시 저장소가 호출되면 안 되는 이유는 무엇인가?

작성 힌트:

```text
debit/credit 균형이 깨진 Ledger가 DB에 저장되면 원장 자체가 신뢰할 수 없게 된다는 점을 적는다.
```

내 답변:

```text
검증에 실패한 데이터가 저장되면 debit·credit 균형이 깨진 원장이 남아 돈의 이동 기록을 신뢰할 수 없기 때문이다.
따라서 RecordTransaction은 ValidateTransaction이 실패하면 Store를 호출하지 않고 즉시 오류를 반환해야 한다.
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
Payment FINALIZED 사건에서 어떤 Ledger Transaction과 Entries를 생성할지 결정해야 한다.
reference_type과 reference_id, 같은 사건의 중복 기록을 막을 idempotency_key, 고객·가맹점 지급 예정·플랫폼 수수료 계정의 debit/credit 배분을 함께 설계해야 한다.
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
`go test ./internal/ledger -v` 실행 결과 ValidateTransaction과 RecordTransaction 테스트가 모두 통과했다.
Repository 통합 테스트는 TEST_DATABASE_URL이 없어 skip되었고, DB 없이 실행되는 Repository 검증 테스트는 통과했다.
`go test ./...` 실행 결과 전체 패키지가 통과했다.
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
newTestService(t)는 NewService(nil)을 반환하므로 저장소가 필요 없는 ValidateTransaction 테스트에서 사용한다.
newTestServiceWithStore(t, store)는 전달받은 fakeStore를 Service에 넣으므로 RecordTransaction 테스트에서 사용한다.

RecordTransaction 테스트에서는 store := &fakeStore{}를 만든 뒤 newTestServiceWithStore(t, store)를 호출한다.
그 결과 svc.store와 테스트의 store 변수가 같은 fakeStore를 가리키며, 호출 뒤 store.calls와 전달값을 확인할 수 있다.
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

작성한 답변의 핵심 방향은 맞았다. 아래 두 가지를 더 구체적으로 기억한다.

```text
1. Service는 DB에 직접 저장하지 않고 Store interface를 통해 저장을 위임한다.
2. interface의 핵심 장점은 단순한 "추상화"가 아니라 구현 교체와 독립적인 Service 테스트다.
```

### 코드 검토 결과

```text
- Store interface와 Repository의 메서드 시그니처가 일치한다.
- RecordTransaction은 ValidateTransaction을 먼저 실행한다.
- 검증 실패 시 fakeStore.calls가 0임을 확인한다.
- 검증 성공 시 Transaction과 Entries가 Store에 전달되는지 확인한다.
- Store 오류를 Service가 반환하는지 확인한다.
- 전체 Go 테스트가 통과한다.
```

필수 코드 수정 사항은 발견되지 않았다. `store == nil` 오류 분기를 별도 테스트로 추가하면 방어 로직의 테스트 범위를 더 넓힐 수 있지만, Day20 완료 기준에는 영향을 주지 않는다.

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
