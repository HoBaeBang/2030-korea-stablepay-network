# Day 22 보충학습 - Settlement 계산에서 다시 확인할 개념

관련 Jira: [SPN-39](https://aslan0.atlassian.net/browse/SPN-39)

이 문서는 Day22 실습산출물 검토에서 실제로 헷갈렸던 부분만 다시 학습하기 위한 자료입니다.

구현가이드 전체를 처음부터 다시 읽기보다 아래 다섯 항목을 우선 확인합니다.

```text
1. MERCHANT_PENDING, CREDIT, DRAFT의 차이
2. Batch와 Item의 일대다 관계
3. Candidate와 Recipient의 의미
4. Service, Calculator, Repository의 책임
5. map과 comma-ok를 이용한 중복 검사
```

## 1. 역할, 방향, 상태를 구분한다

Day22에서 가장 먼저 보충해야 할 부분은 다음 세 값이 서로 다른 대상을 설명한다는 점입니다.

| 값 | 분류 | 무엇을 설명하는가 |
| --- | --- | --- |
| `MERCHANT_PENDING` | `AccountType` | Ledger Account가 어떤 역할의 돈을 보관하는가 |
| `CREDIT` | `EntryDirection` | Ledger Entry가 해당 계정의 금액을 증가시키는가 |
| `DRAFT` | `Settlement Status` | Settlement Batch가 어느 처리 단계에 있는가 |

```text
MERCHANT_PENDING CREDIT
= 가맹점에게 귀속됐지만 아직 지급되지 않은 금액이 지급 예정 계정에 증가했다.

Settlement DRAFT
= 그 지급 예정 금액들을 모아 정산 묶음을 계산했지만 아직 승인·지급하지 않았다.
```

즉 `MerchantPending`은 Batch 상태가 아닙니다. 새 Batch의 상태는 `DRAFT`입니다.

## 2. Batch는 요약이고 Item은 근거다

![Settlement Batch와 Items](https://raw.githubusercontent.com/HoBaeBang/2030-korea-stablepay-network/main/docs/confluence/diagrams/spn39-day22-settlement-calculation.png)

Batch 하나는 정산 묶음 전체를 요약합니다.

```text
Batch
- recipient: merchant_1
- currency: USDC
- total: 14.8 USDC
- item_count: 2
- status: DRAFT
```

Item은 그 총액을 구성한 개별 Ledger Entry를 가리킵니다.

```text
Item 1 -> led_entry_payment_1 -> 9.8 USDC
Item 2 -> led_entry_payment_2 -> 5.0 USDC
```

관계는 다음과 같습니다.

```text
Batch 1개
└── Item 여러 개
    └── Item 하나는 Ledger Entry 하나를 참조
```

따라서 Batch가 Ledger Entry 하나를 의미하는 것이 아닙니다.

## 3. Candidate와 Recipient를 구분한다

`Candidate`는 아직 정산에 포함되기 전의 **정산 후보 데이터**입니다.

```go
type Candidate struct {
	LedgerEntryID string
	RecipientID   string
	AccountType   ledger.AccountType
	Direction     ledger.EntryDirection
	Amount        int64
	Currency      string
}
```

`Recipient`는 정산금을 받을 주체입니다. 현재 프로젝트 예시에서는 가맹점 ID인 `merchant_1`입니다.

```text
Candidate 1: merchant_1에게 지급할 9.8 USDC
Candidate 2: merchant_1에게 지급할 5.0 USDC

같은 Recipient + 같은 Currency
-> Batch 하나로 계산 가능
```

수취인이나 통화가 다르면 같은 Batch에 합치지 않습니다.

## 4. 계산과 DB 저장의 책임을 분리한다

Day22에서는 테스트가 Candidate를 직접 만들기 때문에 DB 조회가 아직 보이지 않습니다.

Day23 이후 실제 흐름은 다음처럼 확장됩니다.

```text
Settlement Service
-> Repository에 정산 후보 조회 요청
-> Repository가 DB에서 미정산 MERCHANT_PENDING CREDIT 조회
-> Calculator가 후보 검증, 총액 합산, Batch와 Items 생성
-> Repository가 Batch와 Items를 DB에 저장
```

| 구성 요소 | 책임 |
| --- | --- |
| `Service` | 조회, 계산, 저장 순서를 조정한다 |
| `Repository` | DB에서 후보를 조회하고 결과를 저장한다 |
| `Calculator` | 후보를 검증하고 Batch와 Items를 계산한다 |

Repository가 비즈니스 계산까지 담당하게 만들지 않습니다.

## 5. `seenEntryIDs`는 Set 역할을 한다

```go
seenEntryIDs := make(map[string]struct{}, len(candidates))
```

Go에는 기본 Set 자료형이 없기 때문에 `map[string]struct{}`를 Set처럼 사용합니다.

```go
if _, exists := seenEntryIDs[candidate.LedgerEntryID]; exists {
	return nil, nil, fmt.Errorf("중복된 ledger entry입니다: %s", candidate.LedgerEntryID)
}

seenEntryIDs[candidate.LedgerEntryID] = struct{}{}
```

Map 조회 결과는 Key가 아니라 `Value, 존재 여부`입니다.

```text
_, exists := map[key]

_      = Value를 사용하지 않으므로 버린다.
exists = 해당 Key가 이미 있으면 true다.
```

처음 본 Entry ID는 Map에 저장하고, 같은 ID가 다시 나오면 중복 정산과 과지급을 막기 위해 실패합니다.

## 보충 확인 문제

### 1. `MERCHANT_PENDING`, `CREDIT`, `DRAFT`는 각각 무엇을 설명하는가?

<details>
<summary>답변 보기</summary>

`MERCHANT_PENDING`은 Ledger Account 역할, `CREDIT`은 Ledger Entry 방향, `DRAFT`는 Settlement Batch 상태입니다.

</details>

### 2. Batch 하나와 Item 하나는 각각 무엇을 의미하는가?

<details>
<summary>답변 보기</summary>

Batch는 같은 수취인과 통화의 정산 묶음 요약이고, Item은 그 묶음에 포함된 개별 Ledger Entry의 근거입니다.

</details>

### 3. 수취인이나 통화가 다른 Candidate를 같은 Batch에 넣으면 안 되는 이유는 무엇인가?

<details>
<summary>답변 보기</summary>

Batch는 한 수취인에게 한 통화로 지급할 금액을 나타내므로 수취인이나 통화가 다르면 지급 대상과 금액 단위가 섞이게 됩니다.

</details>

### 4. Repository가 아니라 Calculator가 합계를 계산하는 이유는 무엇인가?

<details>
<summary>답변 보기</summary>

Repository는 데이터 조회·저장을 담당하고 Calculator는 정산 정책 검증과 계산을 담당해야 각 책임이 분리되고 단위 테스트하기 쉬워집니다.

</details>

### 5. `exists`가 `true`이면 무엇을 의미하는가?

<details>
<summary>답변 보기</summary>

현재 Ledger Entry ID가 이미 `seenEntryIDs` Map에 저장되어 있어 같은 계산 안에서 두 번째로 등장했다는 뜻입니다.

</details>

## 보충 완료 기준

아래 문장을 자료를 보지 않고 설명할 수 있으면 Day22 보충 학습이 완료된 것입니다.

```text
MERCHANT_PENDING은 계정 역할이고 CREDIT은 Entry 방향이며 DRAFT는 정산 상태다.
Batch는 여러 Item의 요약이고 Item은 개별 Ledger Entry를 가리킨다.
Repository는 조회·저장하고 Calculator는 검증·계산하며 Service는 실행 순서를 조정한다.
seenEntryIDs는 같은 Ledger Entry가 두 번 정산되는 것을 막는 Set 역할의 Map이다.
```
