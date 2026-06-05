# Ledger와 Settlement 실습 가이드

관련 Jira: [SPN-19](https://aslan0.atlassian.net/browse/SPN-19)

Confluence 문서: [Ledger와 Settlement 실습 가이드](https://aslan0.atlassian.net/wiki/spaces/SPN/pages/5013640)

이 문서는 퇴근 후 직접 진행할 Ledger와 Settlement 실습가이드입니다.

오늘의 실습은 코드를 구현하는 것이 아니라, 이후 Ledger 구현에 들어가기 전 도메인 문서를 직접 작성하는 것입니다.

## 실습 목표

`docs/domain/02_원장과_정산/Ledger와_Settlement_실습산출물.md` 파일을 직접 만들고, 다음 내용을 정리합니다.

1. Payment, Ledger, Settlement의 책임 비교
2. Payment `FINALIZED`와 Settlement `PAID`의 차이
3. 고객이 10 USDC를 결제했을 때 ledger entry 예시
4. Ledger 구현을 위한 최소 테이블 후보
5. Day 3 이후 구현으로 이어질 질문

## 작업 전 준비

로컬 public repo 위치에서 작업합니다.

```shell
cd /Users/banghobae/Documents/2030-korea-stablepay/2030-korea-stablepay-network
```

작업 전 상태를 확인합니다.

```shell
git status
```

새 파일을 만듭니다.

```shell
touch docs/domain/02_원장과_정산/Ledger와_Settlement_실습산출물.md
```

## 작성할 문서 구조

아래 목차를 그대로 사용해도 됩니다.

```markdown
# Ledger and Settlement

## 한 문장 요약

## 왜 Ledger가 필요한가

## Payment, Ledger, Settlement 책임 비교

## Payment FINALIZED와 Settlement PAID의 차이

## Double-entry 예시

## 최소 테이블 후보

## 아직 모르는 것과 다음 질문

## 검증 체크리스트
```

## 섹션별 작성 가이드

### 1. 한 문장 요약

아래 문장을 참고하되, 그대로 복사하지 말고 본인 말로 바꿔보세요.

```text
Ledger는 결제 이후 돈의 이동을 신뢰할 수 있게 기록하는 장부이고,
Settlement는 확정된 결제 금액을 가맹점에게 지급 가능한 묶음으로 계산하는 과정이다.
```

### 2. 왜 Ledger가 필요한가

다음 질문에 답하는 방식으로 작성합니다.

```text
Payment 상태만 있으면 무엇이 부족한가?
돈의 이동 기록이 없으면 어떤 문제가 생기는가?
중복 결제나 장애 복구 때 Ledger가 왜 필요한가?
```

### 3. 책임 비교표 작성

아래 표를 직접 채웁니다.

| 도메인 | 책임 | 저장해야 할 정보 | 예시 상태 |
| --- | --- | --- | --- |
| Payment |  |  |  |
| Ledger |  |  |  |
| Settlement |  |  |  |

작성 힌트:

| 도메인 | 책임 힌트 |
| --- | --- |
| Payment | 결제가 어디까지 진행됐는지 |
| Ledger | 돈이 왜 어떻게 이동했는지 |
| Settlement | 가맹점에게 얼마를 지급할지 |

### 4. FINALIZED와 PAID 차이 작성

다음 두 문장을 반드시 구분해서 작성합니다.

```text
Payment FINALIZED는 블록체인 결제가 충분히 확정됐다는 의미다.
Settlement PAID는 가맹점에게 지급할 정산 처리가 끝났다는 의미다.
```

### 5. Double-entry 예시 작성

고객이 10 USDC를 결제한 상황을 직접 적어봅니다.

예시 형식:

```text
Ledger Transaction: payment finalized for invoice_xxx

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

중요한 점:

```text
합계가 0이 되는지 확인한다.
왜 돈이 이동했는지 reason을 적는다.
payment_id 또는 invoice_id와 연결될 수 있어야 한다.
```

### 6. 최소 테이블 후보 작성

아래 표를 직접 작성합니다.

| 테이블 후보 | 필요한 이유 | 저장할 주요 값 |
| --- | --- | --- |
| ledger_accounts |  |  |
| ledger_transactions |  |  |
| ledger_entries |  |  |
| settlements |  |  |
| settlement_items |  |  |

### 7. 아직 모르는 것과 다음 질문

이 섹션은 모르는 것을 남기는 곳입니다.

예시:

```text
double-entry에서 debit/credit으로 표현할지 signed amount로 표현할지 아직 모르겠다.
settlement가 실패했을 때 ledger를 되돌려야 하는지 보상 entry를 만들어야 하는지 궁금하다.
수수료가 있는 경우 ledger entry가 몇 개가 되어야 하는지 궁금하다.
```

## 검증 체크리스트

문서 작성 후 아래 질문에 답할 수 있는지 확인합니다.

- [ ] Payment와 Ledger의 차이를 설명할 수 있다.
- [ ] Ledger와 Settlement의 차이를 설명할 수 있다.
- [ ] Payment `FINALIZED`와 Settlement `PAID`의 차이를 설명할 수 있다.
- [ ] 10 USDC 결제 예시에서 ledger entry 합계가 0이 되게 작성할 수 있다.
- [ ] 최소 테이블 후보 5개를 말할 수 있다.

## 권장 커밋 메시지

문서를 직접 작성한 뒤 커밋할 때는 아래 메시지를 추천합니다.

```shell
git add docs/domain/02_원장과_정산/Ledger와_Settlement_실습산출물.md
git commit -m "docs: SPN-19 Ledger와 Settlement 도메인 정리"
git push
```
