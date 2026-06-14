# Day 15 실습산출물 - Ledger Service 균형 검증 구현

관련 Jira: SPN-32

Day15 산출물은 5개 질문만 작성합니다.

오늘 산출물은 “내가 Ledger 전체를 다 이해했는지”를 묻는 문서가 아닙니다.

오늘 만든 코드 하나를 정확히 이해했는지 확인하는 문서입니다.

## 작성 전 확인

아래 파일을 먼저 확인합니다.

```bash
ls internal/ledger
```

아래 파일이 보여야 합니다.

```text
ledger.go
service.go
service_test.go
```

## 1. `ValidateTransaction`은 어떤 책임을 가지는가?

작성할 때 볼 파일:

```text
internal/ledger/service.go
```

작성 힌트:

```text
DB 저장을 하는 함수인지,
저장 전에 규칙을 검증하는 함수인지 구분해서 적는다.
```

내 답변:

```text

```

## 2. debit과 credit 합계가 맞지 않으면 왜 실패해야 하는가?

작성할 때 참고할 흐름:

```text
고객 DEBIT 10_000_000
가맹점 CREDIT 9_800_000
플랫폼 CREDIT 200_000
```

작성 힌트:

```text
돈이 생기거나 사라진 것처럼 보이는 문제,
정산과 장애 복구가 틀어지는 문제를 연결해서 적는다.
```

내 답변:

```text

```

## 3. `map[string]int64`는 오늘 코드에서 어떤 역할을 하는가?

작성할 때 볼 코드:

```go
totals := make(map[string]int64)
```

작성 힌트:

```text
string key가 무엇인지,
int64 value가 무엇인지,
왜 통화별로 나누어 계산하는지 적는다.
```

내 답변:

```text

```

## 4. 오늘 테스트 중 가장 중요하다고 생각하는 실패 케이스는 무엇인가?

작성할 때 볼 파일:

```text
internal/ledger/service_test.go
```

작성 힌트:

```text
성공 케이스보다 실패 케이스가 왜 중요한지,
그 테스트가 어떤 버그를 막는지 적는다.
```

내 답변:

```text

```

## 5. 다음 단계에서 Repository와 DB가 필요한 이유는 무엇인가?

작성 힌트:

```text
오늘 Service는 검증만 한다.
검증된 Ledger Transaction을 영구 보관하려면 무엇이 필요한지 적는다.
```

내 답변:

```text

```

## 오늘 실행 결과

아래 명령 실행 결과를 짧게 기록합니다.

```bash
go test ./internal/ledger -v
go test ./...
```

기록:

```text

```

## 아직 헷갈리는 부분

아래 후보 중 헷갈리는 것이 있으면 적습니다.

```text
receiver
context
map
int64
debit/credit
t.Run
err == nil
```

메모:

```text

```

## 정답/점검 가이드

먼저 `내 답변`을 작성한 뒤 아래 내용을 펼쳐서 비교합니다.

### 1. `ValidateTransaction`은 어떤 책임을 가지는가?

<details>
<summary>정답/점검 가이드 보기</summary>

`ValidateTransaction`은 Ledger Transaction을 DB에 저장하기 전에 비즈니스 규칙을 검증하는 함수입니다.

이 함수의 핵심 책임은 다음과 같습니다.

- 거래 안에 Entry가 존재하는지 확인한다.
- 각 Entry의 금액이 0보다 큰지 확인한다.
- Entry의 방향이 `DEBIT` 또는 `CREDIT`인지 확인한다.
- 통화별로 debit 합계와 credit 합계가 같은지 확인한다.

중요한 점은 이 함수가 DB 저장을 직접 하지 않는다는 것입니다.

즉, `ValidateTransaction`은 “원장에 기록해도 되는 거래인가?”를 판단하는 검증 계층의 함수입니다.

</details>

### 2. debit과 credit 합계가 맞지 않으면 왜 실패해야 하는가?

<details>
<summary>정답/점검 가이드 보기</summary>

Ledger는 돈의 이동을 기록하는 시스템이므로, 하나의 거래 안에서 빠져나간 돈과 들어간 돈의 합이 같아야 합니다.

예를 들어 고객에게서 10 USDC가 빠져나갔다면, 가맹점 지급 예정 금액과 플랫폼 수수료 등을 합쳐서 반드시 10 USDC가 되어야 합니다.

합계가 맞지 않으면 다음 문제가 생깁니다.

- 시스템 안에서 돈이 새로 생긴 것처럼 보일 수 있다.
- 시스템 안에서 돈이 사라진 것처럼 보일 수 있다.
- 정산 금액이 틀어진다.
- 장애 복구나 Reconciliation 때 어떤 금액이 맞는지 판단하기 어렵다.

그래서 Ledger Service는 저장 전에 debit 합계와 credit 합계가 같은지 반드시 검증해야 합니다.

</details>

### 3. `map[string]int64`는 오늘 코드에서 어떤 역할을 하는가?

<details>
<summary>정답/점검 가이드 보기</summary>

`map[string]int64`는 통화별 잔액 차이를 계산하기 위해 사용합니다.

```go
totals := make(map[string]int64)
```

여기서 `string` key는 통화 코드입니다.

예:

```text
USDC
KRW
```

`int64` value는 해당 통화의 debit과 credit 차이를 누적한 값입니다.

통화별로 따로 계산하는 이유는 서로 다른 통화를 섞어서 균형을 맞추면 안 되기 때문입니다.

예를 들어 `10 USDC debit`과 `10 KRW credit`은 숫자는 같아 보여도 같은 가치의 이동이 아니므로 균형이 맞는 거래가 아닙니다.

</details>

### 4. 오늘 테스트 중 가장 중요하다고 생각하는 실패 케이스는 무엇인가?

<details>
<summary>정답/점검 가이드 보기</summary>

가장 중요한 실패 케이스 후보는 “debit과 credit 합계가 맞지 않는 거래는 실패해야 한다”입니다.

이 테스트가 중요한 이유는 Ledger의 가장 핵심 규칙을 지키기 때문입니다.

성공 케이스는 정상적인 거래가 통과되는지 확인합니다.

하지만 실패 케이스는 잘못된 돈의 이동이 시스템에 저장되는 것을 막습니다.

특히 Ledger는 결제, 정산, 장애 복구, 온체인 이벤트 반영의 기준 데이터가 되므로 잘못된 거래가 한 번 저장되면 뒤의 모든 계산이 틀어질 수 있습니다.

</details>

### 5. 다음 단계에서 Repository와 DB가 필요한 이유는 무엇인가?

<details>
<summary>정답/점검 가이드 보기</summary>

Day15의 Service는 Ledger Transaction이 올바른지 검증하는 역할까지만 담당합니다.

하지만 검증을 통과한 거래는 사라지면 안 됩니다.

결제 이후 정산, 환불, 장애 복구, Reconciliation에서 다시 확인할 수 있어야 하므로 DB에 영구 보관해야 합니다.

그래서 다음 단계에서는 Repository와 DB가 필요합니다.

- Service: 거래 규칙을 검증한다.
- Repository: 검증된 거래를 DB에 저장하고 조회한다.
- DB: Ledger Account, Transaction, Entry를 영구 보관한다.

</details>
