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
ledger의 정보를 저장하기 전 해당 ledger의 유효성을 확인하고 검증한다.
컨텍스트가 비어있지는 않은지, 에러가 발생하지 않았는지, entry의 개수가 2개이상인지
entry의 합계가 0이되는지 확인한다.
```

Codex 점검:

```text
방향은 좋습니다. 핵심은 ValidateTransaction이 DB에 저장하는 함수가 아니라,
저장 전에 원장 거래가 규칙을 만족하는지 검증하는 함수라는 점입니다.

조금 더 정확히 쓰면 "entry의 합계가 0"이 아니라
"통화별 debit 합계와 credit 합계가 서로 같아서 최종 차이가 0인지 확인한다"입니다.
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
돈이 중복 출금 되거나 중복 입금된 상황으로 돈의 절대값은 같아야 하지만 그 절댓값이 달라지게 되기때문에
credit, debit의 합계가 0으로 항상동일해야 한다.
```

Codex 점검:

```text
핵심을 잘 잡았습니다. debit과 credit 합계가 맞지 않으면 시스템 안에서 돈이 새로 생기거나 사라진 것처럼 보입니다.

다만 "중복 출금/중복 입금"은 원인이 될 수 있는 여러 문제 중 하나입니다.
Ledger 균형 검증의 더 일반적인 목적은 모든 원장 거래가 돈의 보존 법칙을 지키도록 막는 것입니다.
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
각 통화에 해당하는 entry의 debit, credit을 정산한다.
key 는 통화의 종류, value는 합계이다.
```

Codex 점검:

```text
맞습니다. key는 "USDC" 같은 통화 코드이고, value는 해당 통화에서 debit과 credit을 반영한 누적 합계입니다.

여기서 "정산"이라는 표현보다는 "균형 검증을 위한 통화별 합계 계산"이라고 쓰면 더 정확합니다.
정산(settlement)은 나중에 가맹점이나 수취인에게 지급 가능한 금액을 계산하는 별도 단계입니다.
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
newTestService(t) helper가 테스트 중복을 어떻게 줄이는지도 함께 확인한다.
```

내 답변:

```text
통화가 비어있거나, entry가 하나뿐일경우를 검증하는것으로
통화가 비어있으면 결제 통화를 정할수 없기에 불가능한 거래가 되게되고,
entry가 하나뿐일경우 단일 인출, 입금의 이벤트로 거래의 정보가 명확하지 않기에 검증이 중요하다
```

Codex 점검:

```text
좋은 관점입니다. 통화가 비어 있으면 어떤 자산의 이동인지 알 수 없고,
entry가 하나뿐이면 돈이 어디서 빠져나와 어디로 들어갔는지 짝이 맞지 않습니다.

추가로 가장 핵심적인 실패 케이스는 "debit과 credit 합계가 맞지 않는 경우"입니다.
이 테스트가 Ledger의 기본 규칙인 돈의 보존을 직접 지키기 때문입니다.

newTestService(t) helper는 각 테스트가 독립적인 Service와 context를 갖게 하면서,
반복되는 테스트 준비 코드를 줄이는 역할을 합니다.
```

## 5. 다음 단계에서 Repository와 DB가 필요한 이유는 무엇인가?

작성 힌트:

```text
오늘 Service는 검증만 한다.
검증된 Ledger Transaction을 영구 보관하려면 무엇이 필요한지 적는다.
```

내 답변:

```text
검증에 성공한 ledger는 db에 기록해서 서비스에 반영 영구 보관하면 된다.
```

Codex 점검:

```text
맞습니다. Service는 "저장해도 되는지"를 검증하고,
Repository와 DB는 "검증된 원장 거래를 사라지지 않게 보관"합니다.

조금 더 확장하면 DB에 저장된 Ledger는 나중에 정산, 환불, 장애 복구,
Reconciliation에서 기준 데이터로 사용됩니다.
```

## 오늘 실행 결과

아래 명령 실행 결과를 짧게 기록합니다.

```bash
go test ./internal/ledger -v
go test ./...
```

기록:

```text
go test ./internal/ledger -v: PASS
go test ./...: PASS

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
Day15 기준으로는 debit/credit 균형 검증, map[string]int64의 역할,
newTestService(t) helper의 목적을 계속 반복해서 보면 좋다.

```

## 추가 보충 정리

```text
1. ValidateTransaction은 저장 함수가 아니라 검증 함수다.
2. debit과 credit 합계가 맞아야 원장 안에서 돈이 생기거나 사라지지 않는다.
3. map[string]int64는 통화별 균형을 따로 계산하기 위한 자료구조다.
4. 실패 테스트는 잘못된 원장 거래가 시스템에 저장되는 것을 막는 안전장치다.
5. newTestService(t)는 전역 변수 공유 없이 테스트 준비 중복을 줄이는 helper다.
6. Repository와 DB는 검증된 Ledger Transaction을 영구 보관하기 위해 필요하다.
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

테스트 코드에서는 `newTestService(t)` helper를 사용해 각 테스트 케이스가 독립적인 `Service`와 `context`를 받도록 정리하는 것이 좋습니다.

전역 변수로 `Service`를 공유하면 지금은 편해 보여도, 나중에 Repository나 fake DB가 붙었을 때 테스트끼리 상태가 섞일 수 있습니다.

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
