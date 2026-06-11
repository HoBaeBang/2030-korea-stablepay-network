# Day 11 개념학습 - Payment 테스트에서 Ledger 구현으로 넘어가기

관련 Jira: [SPN-28](https://aslan0.atlassian.net/browse/SPN-28)

Day11의 개념학습은 Day8~10 전체를 넓게 외우는 자료가 아닙니다.

오늘은 아래 하나의 흐름만 잡습니다.

```text
Payment 상태 전이 테스트
  -> 상태 변경 규칙을 고정한다
  -> 운영 로그 후보를 떠올린다
  -> 다음 Ledger 구현에서 필요한 테스트 감각으로 이어진다
```

## Payment와 Ledger의 차이

| 구분 | Payment | Ledger |
| --- | --- | --- |
| 핵심 질문 | 결제가 지금 어떤 상태인가? | 돈이 누구에게서 누구에게 얼마만큼 이동했는가? |
| 예시 | `PENDING`, `ONCHAIN_DETECTED`, `FINALIZED` | `debit`, `credit`, `transaction`, `entry` |
| 주요 위험 | 잘못된 상태 전이 | 중복 기록, 금액 불일치, 부분 저장 |
| 테스트 관점 | 상태가 올바르게 바뀌는가 | 돈의 이동 기록이 균형 있게 남는가 |

Payment는 결제의 “상태”를 관리합니다.

Ledger는 돈의 “이동 기록”을 관리합니다.

## Day10 테스트가 중요한 이유

Day10에서 추가한 테스트는 아래 규칙을 고정합니다.

```text
transaction_hash 없이 ONCHAIN_DETECTED로 변경할 수 없다.
```

이 테스트는 단순히 에러를 확인하는 테스트가 아닙니다.

아래 위험을 막습니다.

```text
블록체인 거래 식별자 없이 결제 상태만 온체인 감지로 바뀌는 문제
```

만약 이 상태가 저장되면 나중에 이런 질문에 답하기 어렵습니다.

```text
어떤 transaction 때문에 상태가 바뀌었는가?
실제 블록체인에 그 transaction이 존재하는가?
같은 transaction이 두 번 처리되지는 않았는가?
나중에 Ledger 기록과 연결할 수 있는가?
```

## 테스트와 로그의 차이

| 구분 | 테스트 | 로그 |
| --- | --- | --- |
| 시점 | 개발/CI 시점 | 운영 시점 |
| 목적 | 규칙이 깨지지 않게 막는다 | 발생한 일을 추적한다 |
| 예시 | transaction_hash 없으면 실패해야 한다 | payment status transition rejected |
| 보는 사람 | 개발자 | 개발자, 운영자 |

테스트는 “앞으로도 이 규칙을 깨지 말라”는 장치입니다.

로그는 “이미 발생한 일을 나중에 찾을 수 있게 남기는 기록”입니다.

둘은 서로 대체하지 않습니다.

## Backend Core가 Ledger에 연결되는 지점

![Day11 Backend Core에서 Ledger로 넘어가는 준비도](../../../confluence/diagrams/spn28-day11-ledger-readiness.png)

| Backend Core 기반 | Ledger에서 필요한 이유 |
| --- | --- |
| Error Response | 잘못된 Ledger 요청을 일관된 실패 응답으로 표현한다 |
| Validation | 금액, 통화, debit/credit 합계 규칙을 막는다 |
| Config | DB, 향후 RPC, signer 설정 누락을 시작 시점에 잡는다 |
| Logging | 돈의 이동 기록 생성/중복 무시/실패 지점을 추적한다 |
| Test Pattern | 중복 Ledger, 불균형 entry, 부분 저장을 반복 검증한다 |

## Ledger 첫 구현에서 가장 먼저 조심할 것

Ledger는 “테이블 만들고 저장하면 끝”인 CRUD가 아닙니다.

Ledger에서 특히 위험한 것은 아래입니다.

| 위험 | 왜 위험한가 | 막는 방법 |
| --- | --- | --- |
| 중복 원장 생성 | 같은 payment가 두 번 반영되면 내부 돈이 두 번 움직인 것처럼 보인다 | unique key, idempotency test |
| 불균형 entry | debit과 credit 합계가 맞지 않으면 돈이 생기거나 사라진 것처럼 보인다 | service validation, test |
| 부분 저장 | transaction은 저장되고 entry 일부가 실패하면 장부가 깨진다 | DB transaction |
| 추적 불가 | 어떤 payment에서 Ledger가 생겼는지 알 수 없다 | payment_id, ledger_transaction_id 로그 |

## Day11의 핵심 결론

```text
Day11은 Ledger를 구현하는 날이 아니다.

Day11은 Payment 상태 전이 테스트를 이해하고,
다음 Ledger 구현을 작게 시작할 수 있게 준비하는 날이다.
```
