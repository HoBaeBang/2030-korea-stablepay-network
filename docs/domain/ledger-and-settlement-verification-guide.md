# Ledger와 Settlement 검증문제와 답변가이드

관련 Jira: [SPN-19](https://aslan0.atlassian.net/browse/SPN-19)

Confluence 문서: [Ledger와 Settlement 검증문제와 답변가이드](https://aslan0.atlassian.net/wiki/spaces/SPN/pages/4948076)

이 문서는 Ledger와 Settlement 학습 후 스스로 확인할 검증문제와 답변가이드입니다.

먼저 문제를 풀고, 그 다음 답변가이드를 확인하세요.

## 검증문제

### 문제 1

Payment와 Ledger의 차이를 한 문장씩 설명하세요.

### 문제 2

Payment `FINALIZED`와 Settlement `PAID`는 왜 같은 의미가 아닌가요?

### 문제 3

고객이 10 USDC를 결제했을 때, double-entry 방식으로 ledger entry를 2개 작성해보세요.

### 문제 4

Ledger entry의 합계가 0이 되어야 한다는 말은 무슨 뜻인가요?

### 문제 5

`ledger_accounts`, `ledger_transactions`, `ledger_entries`는 각각 어떤 역할을 하나요?

### 문제 6

`settlements`와 `settlement_items`는 왜 분리하는 것이 좋을까요?

### 문제 7

Payment 상태만으로 정산 기능을 만들면 어떤 문제가 생길 수 있나요?

### 문제 8

Ledger가 있으면 장애 복구나 중복 처리에서 어떤 점이 좋아지나요?

## 답변가이드

### 답변 1

Payment는 결제가 어디까지 진행됐는지를 나타내는 상태 모델입니다.

Ledger는 돈이 어느 계정에서 어느 계정으로 왜 이동했는지를 기록하는 장부입니다.

핵심 차이:

```text
Payment = 상태
Ledger = 돈의 이동 기록
```

### 답변 2

Payment `FINALIZED`는 블록체인 결제가 충분히 확정되었다는 의미입니다.

Settlement `PAID`는 가맹점에게 지급할 정산 처리가 끝났다는 의미입니다.

즉 결제 확정과 가맹점 지급 완료는 서로 다른 단계입니다.

### 답변 3

예시 답변:

```text
Ledger Transaction: payment finalized

Entry 1:
Account: Customer Account
Amount: -10 USDC
Reason: customer paid invoice

Entry 2:
Account: Merchant Pending Account
Amount: +10 USDC
Reason: merchant settlement target increased

Check:
-10 + 10 = 0
```

정답의 핵심은 계정 이름이 완전히 같지 않아도 됩니다.

다만 다음 조건을 만족해야 합니다.

1. 한쪽은 감소해야 합니다.
2. 다른 한쪽은 증가해야 합니다.
3. 합계가 0이어야 합니다.
4. 왜 이동했는지 이유가 있어야 합니다.

### 답변 4

합계가 0이라는 말은 돈이 시스템 안에서 갑자기 생기거나 사라지지 않았다는 뜻입니다.

```text
Customer Account     -10
Merchant Pending     +10
Total                  0
```

이렇게 해야 원장 데이터가 신뢰 가능합니다.

### 답변 5

| 테이블 | 역할 |
| --- | --- |
| `ledger_accounts` | 돈이 들어가거나 나가는 계정 정보를 저장합니다 |
| `ledger_transactions` | 하나의 돈 이동 사건을 묶음으로 저장합니다 |
| `ledger_entries` | 각 계정의 증가/감소 기록을 저장합니다 |

예를 들어 하나의 `ledger_transaction` 안에 고객 계정 감소 entry와 가맹점 pending 계정 증가 entry가 함께 들어갈 수 있습니다.

### 답변 6

`settlements`는 정산 묶음 자체를 나타냅니다.

`settlement_items`는 그 정산 묶음에 어떤 payment들이 포함됐는지를 나타냅니다.

분리하면 다음이 쉬워집니다.

```text
하나의 정산에 여러 payment 포함
정산 금액 계산 근거 추적
특정 payment가 어떤 정산에 포함됐는지 확인
정산 실패 또는 재계산 시 영향 범위 확인
```

### 답변 7

Payment 상태만 있으면 돈의 이동 이유와 정산 근거가 부족합니다.

예를 들어 `FINALIZED` 상태인 payment가 100개 있어도 다음 질문에 답하기 어렵습니다.

```text
어떤 가맹점에게 얼마를 지급해야 하는가?
수수료는 반영됐는가?
중복으로 반영된 payment는 없는가?
이미 정산된 payment인지 어떻게 아는가?
```

그래서 Ledger와 Settlement가 필요합니다.

### 답변 8

Ledger가 있으면 각 돈 이동이 entry로 남기 때문에 장애가 나도 어디까지 처리됐는지 추적할 수 있습니다.

또한 `payment_id`, `transaction_hash`, `ledger_transaction_id` 같은 기준을 두면 같은 결제가 여러 번 들어와도 중복 entry를 막을 수 있습니다.

핵심:

```text
Ledger는 돈의 이동 이력이다.
이력이 있어야 중복 처리와 장애 복구를 판단할 수 있다.
```

## 통과 기준

아래 기준을 만족하면 학습은 충분히 진행된 것입니다.

- [ ] Payment, Ledger, Settlement를 각각 한 문장으로 설명할 수 있다.
- [ ] FINALIZED와 PAID의 차이를 설명할 수 있다.
- [ ] 10 USDC 결제 예시를 double-entry로 작성할 수 있다.
- [ ] 최소 테이블 후보 5개를 말할 수 있다.
- [ ] 아직 모르는 질문을 2개 이상 남길 수 있다.
