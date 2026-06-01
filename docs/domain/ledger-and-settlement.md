# Ledger and Settlement

이 문서는 Day 2 퇴근 후 실습에서 직접 채워 넣는 학습 산출물입니다.

관련 가이드: [Ledger와 Settlement 실습 가이드](ledger-and-settlement-practice-guide.md)

## 한 문장 요약

Ledger는 결제 과정에서 돈의 이동을 기록하는 장부이고,

Settlement는 확정된 결제 금액 중 가맹점에게 실제로 지급 가능한 금액을 묶음으로 계산하고 처리하는 과정이다.

## 왜 Ledger가 필요한가

Payment는 결제가 현재 어떤 상태인지 보여주지만, 돈이 어느 계정에서 어느 계정으로 이동했는지, 왜 이동했는지, 그 이동이 중복 처리된 것은 아닌지까지 설명하지 못한다.

따라서 결제 확정 이후에는 Ledger를 통해 돈의 이동을 entry로 남겨야 한다. Ledger가 있으면 고객 계정 감소, 가맹점 pending 계정 증가, 수수료 차감 같은 흐름을 추적할 수 있고, 각 entry의 합계가 맞는지 확인하면서 데이터 정합성을 검증할 수 있다.

또한 장애가 발생하거나 같은 온체인 이벤트가 여러 번 들어왔을 때, 이미 처리된 payment인지, 어떤 ledger transaction까지 생성됐는지 확인할 수 있으므로 중복 처리 방지와 복구 기준으로 사용할 수 있다.

## Payment, Ledger, Settlement 책임 비교

| 도메인 | 책임 | 저장해야 할 정보 | 예시 상태 |
| --- | --- | --- | --- |
| Payment | 결제가 현재 어디까지 진행됐는지 추적한다. 결제 요청서(invoice)에 대해 온체인 감지, 확정, 실패, 정산 완료 여부를 상태로 관리한다. | payment ID, invoice ID, amount, currency, status, transaction hash, finalized time, 실패 사유 | PENDING, ONCHAIN_DETECTED, FINALIZED, SETTLED, FAILED |
| Ledger | 돈이 어느 계정에서 어느 계정으로 왜 이동했는지 기록한다. 결제, 수수료, 정산, 환불 같은 돈의 이동을 entry로 남겨 데이터 정합성을 검증한다. | ledger transaction ID, account ID, entry amount, currency, reason, reference type, reference ID, created time | 상태보다 기록 중심이다. 예: 고객 계정 -10 USDC, 가맹점 pending 계정 +10 USDC |
| Settlement | 확정된 결제들을 모아 수수료, 보류, 환불, 실패 건을 고려한 뒤 가맹점에게 지급 가능한 금액을 계산하고 지급 처리 흐름을 관리한다. | settlement ID, merchant ID, settlement period, gross amount, fee amount, hold amount, net amount, settlement item 목록, paid time | PENDING, APPROVED, PAID, FAILED, CANCELED |

## Payment FINALIZED와 Settlement PAID의 차이

Payment `FINALIZED`는 고객의 결제가 블록체인상에서 충분히 확정되었다는 의미이다. 즉, 해당 invoice에 대한 결제가 성공했고 되돌아갈 가능성이 낮다고 판단할 수 있는 상태다.

하지만 Payment가 `FINALIZED` 되었다고 해서 가맹점에게 정산금이 바로 지급된 것은 아니다. 정산 전에 수수료, 환불 가능성, 보류 금액, 정산 주기, 가맹점 지급 정보, 내부 대사 결과 등을 확인해야 할 수 있다.

Settlement `PAID`는 정산 대상 결제들을 모아 지급 가능한 금액을 계산하고, 실제 지급 처리까지 완료했다는 의미이다.

따라서 두 상태는 다음처럼 구분해야 한다.

```text
Payment FINALIZED
= 고객 결제가 확정된 상태

Settlement PAID
= 가맹점에게 지급할 정산 처리가 완료된 상태
```

## Double-entry 예시

한글로는 복식부기라고 하고, 간단하게 말하면 돈의 이동을 한 줄로만 기록하지 않고 어느 계정에서 얼마가 빠지고 어느 계정으로 얼마만큼 들어갔는지를 함께 기록하는 방식이다.

```text
Ledger Transaction: payment finalized for invoice_123

Entry 1:
Account: Customer Account
Amount: -10 USDC
Reason: customer paid invoice

Entry 2:
Account: Merchant Pending Account
Amount: +10 USDC
Reason: merchant will receive settlement later

Check:
-10 + 10 = 0
```

여기서 `Customer Account`와 `Merchant Pending Account`는 실제 사람 이름이라기보다 원장에서 돈의 출처와 도착지를 표현하는 계정이다.

핵심은 두 entry의 합계가 0이 되는 것이다. 그래야 시스템 안에서 돈이 갑자기 생기거나 사라지지 않았다고 검증할 수 있다.

## 최소 테이블 후보

| 테이블 후보 | 필요한 이유 | 저장할 주요 값 |
| --- | --- | --- |
| ledger_accounts | 돈이 들어가거나 나가는 주체를 정의하기 위해 필요하다. 고객, 가맹점 pending, 가맹점 payable, 수수료, 시스템 계정처럼 돈의 위치를 구분한다. | id, owner_type, owner_id, account_type, currency, created_at |
| ledger_transactions | 하나의 돈 이동 사건을 묶기 위해 필요하다. 결제 확정, 수수료 차감, 정산 생성, 환불 같은 사건 단위를 표현한다. | id, transaction_type, reference_type, reference_id, description, created_at |
| ledger_entries | 각 계정의 증가/감소를 실제로 기록하기 위해 필요하다. 하나의 ledger transaction에는 최소 2개 이상의 entry가 포함된다. | id, ledger_transaction_id, account_id, amount, currency, direction 또는 signed_amount, created_at |
| settlements | 가맹점별 정산 묶음과 상태를 관리하기 위해 필요하다. 특정 기간의 결제들을 모아 지급 가능한 금액을 계산한다. | id, merchant_id, status, period_start, period_end, gross_amount, fee_amount, hold_amount, net_amount, paid_at |
| settlement_items | 어떤 payment가 어떤 settlement에 포함됐는지 추적하기 위해 필요하다. 정산 금액의 근거를 남기고 중복 정산을 막는다. | id, settlement_id, payment_id, amount, fee_amount, net_amount, created_at |

## 아직 모르는 것과 다음 질문

- Ledger에서 금액을 `+10`, `-10` 같은 signed amount로 저장할지, debit/credit 컬럼으로 나눠 저장할지 궁금하다.
- Settlement가 실패했을 때 기존 ledger entry를 되돌리는지, 아니면 보상 entry를 새로 만드는지 궁금하다.
- 수수료, 환불, 보류 금액이 함께 있을 때 ledger entry가 몇 개로 나뉘어야 하는지 더 학습해야 한다.
- 실제 서비스에서는 settlement를 하루 단위로 만드는지, 가맹점 요청 시점에 만드는지 궁금하다.

## 다시 확인할 학습 포인트

- Payment는 상태 모델이고, Ledger는 돈의 이동 기록이라는 차이를 다시 말로 설명해본다.
- Ledger Transaction은 하나의 사건이고, Ledger Entry는 그 사건으로 인해 각 account에 생긴 증가/감소 기록이라는 점을 다시 확인한다.
- Settlement는 `확정된 결제 금액 전체`가 아니라, 그중 정책상 지급 가능한 금액을 계산하는 과정이라는 점을 다시 확인한다.
- `FINALIZED`와 `PAID`를 같은 의미로 쓰지 않도록 주의한다.

## 검증 체크리스트

- [x] Payment와 Ledger의 차이를 설명할 수 있다.
- [x] Ledger와 Settlement의 차이를 설명할 수 있다.
- [x] Payment `FINALIZED`와 Settlement `PAID`의 차이를 설명할 수 있다.
- [x] 10 USDC 결제 예시에서 ledger entry 합계가 0이 되게 작성할 수 있다.
- [x] 최소 테이블 후보 5개를 말할 수 있다.
