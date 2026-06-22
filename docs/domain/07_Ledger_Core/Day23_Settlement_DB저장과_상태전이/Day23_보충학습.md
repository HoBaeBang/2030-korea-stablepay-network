# Day 23 보충학습 - Settlement 저장 정합성과 상태 충돌

관련 Jira: [SPN-40](https://aslan0.atlassian.net/browse/SPN-40)

이 문서는 Day23 산출물 검토에서 추가 확인이 필요했던 내용만 정리합니다.

## 이번에 보충할 내용

```text
1. Batch와 Items를 transaction으로 묶는 진짜 이유
2. 복합 Primary Key와 ledger_entry_id UNIQUE의 차이
3. 조건부 UPDATE를 이용한 낙관적 동시성 제어
4. Settlement 상태 전이 규칙
```

![Settlement DB 저장과 상태 전이](https://raw.githubusercontent.com/HoBaeBang/2030-korea-stablepay-network/main/docs/confluence/diagrams/spn40-day23-settlement-db-status.png)

## 1. DB transaction은 부분 저장을 막는다

정산 묶음은 Batch 하나와 Item 여러 개가 함께 있어야 완전합니다.

```text
settlement_batches
  stl_batch_1, total_amount=14.8 USDC, item_count=2

settlement_items
  stl_batch_1 + entry_1 + 9.8 USDC
  stl_batch_1 + entry_2 + 5.0 USDC
```

transaction이 없는데 두 번째 Item INSERT가 실패하면 다음처럼 불완전한 데이터가 남을 수 있습니다.

```text
Batch: item_count=2라고 기록됨
Item:  1개만 저장됨

결과: Batch 요약과 계산 근거가 서로 맞지 않음
```

`CreateBatch`는 이를 다음 순서로 방지합니다.

```text
BeginTx
-> Batch INSERT
-> 모든 Item INSERT
-> 전부 성공하면 Commit
-> 하나라도 실패하면 Rollback
```

따라서 transaction의 첫 번째 목적은 중복 방지가 아니라 **원자성**입니다.

```text
원자성(Atomicity)
= 관련된 작업이 전부 반영되거나 전부 반영되지 않는 성질
```

중복 정산은 별도의 `ledger_entry_id UNIQUE` 제약이 담당합니다.

## 2. 복합 Primary Key와 UNIQUE는 막는 중복이 다르다

현재 `settlement_items`의 Primary Key는 다음 두 컬럼의 조합입니다.

```sql
PRIMARY KEY (batch_id, ledger_entry_id)
```

이 제약만 있으면 아래 두 row는 서로 다른 조합이므로 모두 저장될 수 있습니다.

| batch_id | ledger_entry_id | 복합 PK 기준 |
| --- | --- | --- |
| `batch_1` | `entry_1` | 고유함 |
| `batch_2` | `entry_1` | 고유함 |

하지만 두 row 모두 같은 `entry_1`을 정산하므로 실제 돈은 중복 지급될 수 있습니다.

```sql
CREATE UNIQUE INDEX idx_settlement_items_ledger_entry_id
    ON settlement_items (ledger_entry_id);
```

별도 UNIQUE 제약을 적용하면 `entry_1`은 전체 테이블에서 한 번만 등장할 수 있습니다.

```text
복합 Primary Key
= 한 Batch 안에서 같은 Ledger Entry가 중복되는 것을 막음

ledger_entry_id UNIQUE
= 서로 다른 Batch 사이에서도 같은 Ledger Entry가 다시 정산되는 것을 막음
```

## 3. 낙관적 동시성 제어는 충돌을 결과로 감지한다

두 요청이 거의 동시에 같은 Batch를 읽었다고 가정합니다.

```text
요청 A가 읽은 상태: DRAFT
요청 B가 읽은 상태: DRAFT
```

두 요청은 다음 조건부 UPDATE를 실행합니다.

```sql
UPDATE settlement_batches
SET status = 'READY', updated_at = now()
WHERE id = 'batch_1'
  AND status = 'DRAFT';
```

실제 처리 결과:

```text
1. 요청 A가 먼저 성공한다.
   DB 상태: DRAFT -> READY
   RowsAffected: 1

2. 요청 B가 실행된다.
   WHERE status = 'DRAFT'를 만족하는 row가 없다.
   RowsAffected: 0

3. Repository가 충돌을 오류로 반환한다.
```

`WHERE id = $2`만 사용했다면 요청 B도 앞선 변경을 덮어쓸 수 있습니다. `status = $3`까지 조건에 넣어 **내가 읽었던 상태가 아직 유효한지** 함께 검사합니다.

```text
낙관적 동시성 제어(Optimistic Concurrency Control)
= 충돌이 자주 발생하지 않는다고 보고 먼저 작업을 시도한 뒤,
  변경 조건이나 version을 이용해 다른 요청이 먼저 수정했는지 확인하는 방식
```

## 4. Settlement 상태 전이 규칙

정상적인 지급 흐름:

```text
DRAFT -> READY -> APPROVED -> PROCESSING -> PAID
```

상태 의미:

| 상태 | 의미 | 허용되는 다음 상태 |
| --- | --- | --- |
| `DRAFT` | 후보를 계산해 묶음을 생성함 | `READY`, `CANCELED` |
| `READY` | 항목 검증이 끝나 승인 가능 | `APPROVED`, `CANCELED` |
| `APPROVED` | 내부 지급 승인이 완료됨 | `PROCESSING`, `CANCELED` |
| `PROCESSING` | 실제 지급을 처리 중 | `PAID`, `FAILED` |
| `FAILED` | 지급 시도가 실패함 | `PROCESSING`, `CANCELED` |
| `PAID` | 지급 완료 | 없음 |
| `CANCELED` | 지급 전 취소 | 없음 |

Service와 Repository가 각각 다른 안전장치를 담당합니다.

```text
Service.canTransition
= 업무 정책상 허용된 상태 전이인지 확인

Repository.UpdateBatchStatus
= DB 상태가 Service가 예상한 현재 상태와 같은지 확인
```

둘 중 하나만으로는 부족합니다.

```text
Service만 있으면
-> 동시에 들어온 요청이 오래된 상태를 기준으로 판단할 수 있음

Repository 조건부 UPDATE만 있으면
-> DRAFT -> PAID 같은 업무상 금지된 전이를 Repository 직접 호출로 시도할 수 있음
```

## 확인 문제

### 1. Batch INSERT 성공 후 Item INSERT가 실패했을 때 transaction이 없다면 무엇이 문제인가?

<details>
<summary>답변 보기</summary>

Batch 요약만 남거나 일부 Item만 남아 `total_amount`, `item_count`, 실제 계산 근거가 서로 맞지 않는 부분 저장 상태가 됩니다. transaction은 모두 commit하거나 모두 rollback하여 원자성을 지킵니다.

</details>

### 2. `(batch_id, ledger_entry_id)` Primary Key만으로 중복 정산을 완전히 막을 수 없는 이유는 무엇인가?

<details>
<summary>답변 보기</summary>

같은 `ledger_entry_id`라도 `batch_id`가 다르면 복합키 조합은 서로 다릅니다. 따라서 별도의 `ledger_entry_id UNIQUE`가 있어야 서로 다른 Batch 사이의 중복 정산을 막을 수 있습니다.

</details>

### 3. 조건부 UPDATE에서 `RowsAffected() == 0`은 무엇을 의미하는가?

<details>
<summary>답변 보기</summary>

대상 Batch가 없거나 DB의 현재 상태가 요청에서 예상한 상태와 달라 UPDATE 조건을 만족하지 못했다는 뜻입니다. 다른 요청이 먼저 상태를 바꿨을 가능성이 있으므로 충돌로 처리합니다.

</details>

## 보충 완료 기준

```text
[ ] transaction의 주목적을 "부분 저장 방지와 원자성"으로 설명할 수 있다.
[ ] 복합 PK와 ledger_entry_id UNIQUE가 막는 중복의 차이를 설명할 수 있다.
[ ] 조건부 UPDATE와 RowsAffected가 충돌을 감지하는 과정을 설명할 수 있다.
[ ] Service 정책 검증과 Repository 동시성 검증의 차이를 설명할 수 있다.
```
