# Day 22 실습산출물 - Settlement 도메인 타입과 계산 서비스

관련 Jira: [SPN-39](https://aslan0.atlassian.net/browse/SPN-39)

이 문서는 Day22 구현을 마친 뒤 Settlement 계산 흐름을 자기 말로 설명하는 산출물입니다.

힌트는 기본적으로 접혀 있습니다. 먼저 답변을 작성하고 막힐 때만 펼쳐봅니다.

## 1. Ledger와 Settlement의 책임은 어떻게 다른가?

<details>
<summary>힌트 보기</summary>

Ledger는 돈의 이동 사실을 기록하고, Settlement는 그 기록 중 지급 가능한 후보를 묶고 계산한다는 점을 떠올립니다.

</details>

내 답변:

```text

```

## 2. `MerchantPending CREDIT`만 정산 후보로 받는 이유는 무엇인가?

<details>
<summary>힌트 보기</summary>

Customer DEBIT, MerchantPending CREDIT, PlatformFee CREDIT 중 실제 가맹점에게 귀속된 미지급 금액이 무엇인지 생각합니다.

</details>

내 답변:

```text

```

## 3. 아래 후보로 만들어지는 Batch와 Items를 계산해보자.

```text
recipient = merchant_1
currency = USDC

candidate 1 = 9_800_000 USDC
candidate 2 = 5_000_000 USDC
```

<details>
<summary>힌트 보기</summary>

Batch에는 수취인, 통화, 총액, 항목 수, 상태가 들어갑니다. Item은 각 Ledger Entry와 금액을 연결합니다.

</details>

내 답변:

```text
Batch total =
Batch item count =
Batch status =

Item 1 =
Item 2 =
```

## 4. Batch와 Item을 별도 타입으로 관리하는 이유는 무엇인가?

<details>
<summary>힌트 보기</summary>

총액만 있을 때 “어떤 Ledger Entry 때문에 이 금액이 나왔는가?”라는 질문에 답할 수 있는지 생각합니다.

</details>

내 답변:

```text

```

## 5. 같은 `LedgerEntryID`가 두 번 들어오면 왜 실패해야 하는가?

<details>
<summary>힌트 보기</summary>

하나의 돈 이동 기록을 두 번 합산하면 정산 금액과 실제 지급액에 어떤 문제가 생기는지 생각합니다.

</details>

내 답변:

```text

```

## 오늘 실행 결과

```bash
gofmt -w ./internal/settlement
go test ./internal/settlement -v
go test ./...
```

기록:

```text

```

## 아직 헷갈리는 부분

```text
Candidate
Recipient
Batch와 Item
MerchantPending CREDIT
seenEntryIDs
Settlement DRAFT
```

메모:

```text

```

## 정답/점검 가이드

먼저 답변을 작성한 뒤 필요한 항목만 펼쳐서 비교합니다.

### 1. Ledger와 Settlement의 책임은 어떻게 다른가?

<details>
<summary>답변 보기</summary>

Ledger는 결제, 수수료 같은 돈의 이동을 변경하지 않는 기록으로 남깁니다.

Settlement는 Ledger 기록 중 정산 정책을 만족하는 후보를 수취인과 통화 기준으로 묶고 지급 가능한 금액을 계산합니다.

</details>

### 2. `MerchantPending CREDIT`만 정산 후보로 받는 이유는 무엇인가?

<details>
<summary>답변 보기</summary>

`MerchantPending CREDIT`은 가맹점에게 귀속되었지만 아직 지급되지 않은 금액을 뜻합니다.

Customer DEBIT은 고객 측 금액 감소이고 PlatformFee CREDIT은 플랫폼 몫이므로 가맹점 정산 금액에 포함하면 안 됩니다.

</details>

### 3. 아래 후보로 만들어지는 Batch와 Items를 계산해보자.

<details>
<summary>답변 보기</summary>

```text
Batch total = 9_800_000 + 5_000_000 = 14_800_000 USDC
Batch item count = 2
Batch status = DRAFT

Item 1 = 첫 번째 Ledger Entry / 9_800_000 USDC
Item 2 = 두 번째 Ledger Entry / 5_000_000 USDC
```

</details>

### 4. Batch와 Item을 별도 타입으로 관리하는 이유는 무엇인가?

<details>
<summary>답변 보기</summary>

Batch는 수취인, 통화, 총액, 상태 같은 정산 묶음 전체 정보를 나타냅니다.

Item은 Batch 금액의 근거가 된 개별 Ledger Entry를 연결합니다. 둘을 분리해야 총액 계산 근거, 중복 정산 여부, 실패 시 영향 범위를 추적할 수 있습니다.

</details>

### 5. 같은 `LedgerEntryID`가 두 번 들어오면 왜 실패해야 하는가?

<details>
<summary>답변 보기</summary>

같은 Ledger Entry를 두 번 합산하면 가맹점에게 지급할 금액이 실제보다 커집니다.

메모리 계산에서는 `seenEntryIDs` map으로 막고, DB 저장 단계에서는 `settlement_items.ledger_entry_id`에 unique 제약을 두어 다시 방어해야 합니다.

</details>

## 추가 보충 정리

### Codex 점검

```text
Settlement는 Ledger의 합계를 무조건 더하는 기능이 아니다.
정산 가능한 계정 역할, 방향, 수취인, 통화, 금액, 중복 여부를 검증한 뒤 묶는 기능이다.
```

### 코드 확인 포인트

```text
- Candidate가 Ledger Entry와 Recipient 정보를 함께 가지는가?
- BuildBatch가 Context와 필수 값을 검증하는가?
- MerchantPending CREDIT만 허용하는가?
- 다른 수취인이나 통화가 섞이면 실패하는가?
- 중복 LedgerEntryID를 차단하는가?
- Batch 총액과 Item 개수가 정확한가?
```

### 다음 학습 포인트

새 Day23에서는 Settlement Batch와 Items를 DB에 저장하고, DRAFT 이후 상태 흐름을 구현합니다.

