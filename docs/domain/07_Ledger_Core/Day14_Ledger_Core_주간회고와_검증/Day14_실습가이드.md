# Day 14 실습가이드 - Ledger Core 주간 회고와 검증

관련 Jira: [SPN-31](https://aslan0.atlassian.net/browse/SPN-31)

Day14는 새 기능을 크게 추가하지 않습니다.

오늘은 Day12와 Day13에서 만든 코드가 무엇을 의미하는지 다시 읽고, 테스트를 실행하고, 다음 구현으로 넘어갈 준비가 되었는지 확인합니다.

## 실습 흐름

![Day14 Ledger Core 회고 흐름](../../../confluence/diagrams/spn31-day14-ledger-review.png)

## 먼저 확인할 것

Day14는 Day13 실습이 완료된 뒤 진행합니다.

프로젝트 루트에서 아래 명령을 실행합니다.

```bash
ls internal/ledger
```

아래 3개 파일이 모두 보여야 합니다.

```text
ledger.go
service.go
service_test.go
```

만약 `service.go`, `service_test.go`가 없다면 아직 Day13이 끝나지 않은 상태입니다.

그 경우 Day14를 진행하지 말고 먼저 아래 문서를 진행합니다.

```text
docs/domain/07_Ledger_Core/Day13_Ledger_균형검증_테스트/Day13_실습가이드.md
```

## 오늘 확인할 파일

Day12 파일:

```text
internal/ledger/ledger.go
```

Day13 파일:

```text
internal/ledger/service.go
internal/ledger/service_test.go
```

## Step 1. 현재 파일 목록 확인

프로젝트 루트에서 실행합니다.

```bash
ls internal/ledger
```

예상 파일:

```text
ledger.go
service.go
service_test.go
```

이 단계의 목적은 “오늘 회고할 코드가 실제로 있는지” 확인하는 것입니다.

파일이 없다면 산출물에 억지로 답을 쓰지 않습니다.

## Step 2. 타입 정의 다시 읽기

아래 명령으로 핵심 타입 위치를 확인합니다.

```bash
rg -n "type Account|type Transaction|type Entry|type EntryDirection" internal/ledger
```

확인할 것:

```text
Account는 돈이 기록되는 주체다.
Transaction은 여러 Entry를 묶는다.
Entry는 돈의 이동 한 줄이다.
EntryDirection은 DEBIT/CREDIT 방향을 표현한다.
```

읽을 때 아래 질문에 답해봅니다.

```text
Account는 실제 은행 계좌인가, 원장 기록을 위한 계정인가?
Transaction은 블록체인 transaction hash와 같은 개념인가?
Entry는 왜 여러 줄이 필요한가?
```

## Step 3. 균형 검증 로직 다시 읽기

아래 명령으로 검증 메서드를 찾습니다.

```bash
rg -n "ValidateTransaction|totals|EntryDirectionDebit|EntryDirectionCredit" internal/ledger
```

확인할 것:

```text
entries 개수를 확인한다.
amount가 0보다 큰지 확인한다.
currency가 비어 있지 않은지 확인한다.
DEBIT은 더한다.
CREDIT은 뺀다.
최종 합계가 0인지 확인한다.
```

읽을 때 아래 코드 흐름을 손으로 따라가 봅니다.

```text
고객 DEBIT 10_000_000
가맹점 CREDIT 9_800_000
플랫폼 CREDIT 200_000

totals["USDC"] = 10_000_000 - 9_800_000 - 200_000
totals["USDC"] = 0
```

이 결과가 0이면 균형이 맞는 거래입니다.

## Step 4. Ledger 테스트 실행

Ledger 패키지만 테스트합니다.

```bash
go test ./internal/ledger -v
```

테스트 이름을 읽으면서 아래를 확인합니다.

```text
어떤 케이스가 성공해야 하는가?
어떤 케이스가 실패해야 하는가?
실패해야 하는 케이스가 실제로 error를 반환하는가?
```

실행 결과 예시는 아래와 비슷합니다.

```text
=== RUN   TestServiceValidateTransaction
=== RUN   TestServiceValidateTransaction/debit과_credit_합계가_같으면_성공한다
=== RUN   TestServiceValidateTransaction/credit_합계가_부족하면_실패한다
=== RUN   TestServiceValidateTransaction/금액이_0이면_실패한다
=== RUN   TestServiceValidateTransaction/알_수_없는_방향이면_실패한다
--- PASS: TestServiceValidateTransaction
PASS
```

한글 테스트 이름이 보이면 `t.Run`으로 만든 하위 테스트가 실행된 것입니다.

## Step 5. 전체 테스트 실행

프로젝트 전체 테스트를 실행합니다.

```bash
go test ./...
```

전체 테스트가 성공해야 Day12~13의 변경이 다른 기능을 깨뜨리지 않은 것입니다.

만약 전체 테스트가 실패하면 먼저 실패한 패키지를 확인합니다.

```text
internal/ledger에서 실패했는가?
다른 패키지에서 실패했는가?
```

Day14의 목적은 새 기능 개발이 아니므로, 실패 원인이 Day14 회고 범위를 벗어나면 기록만 남기고 별도 작업으로 분리합니다.

## Step 6. 다음 DB 테이블 후보 상상하기

아직 DB migration을 작성하지 않습니다.

다만 다음 구현을 준비하기 위해 타입과 테이블이 어떻게 연결될지 생각해봅니다.

```text
Account      -> ledger_accounts
Transaction  -> ledger_transactions
Entry        -> ledger_entries
```

Day14 산출물에 이 연결을 짧게 적습니다.

아직 컬럼을 완벽히 설계할 필요는 없습니다.

오늘은 아래 정도만 연결해도 충분합니다.

```text
Account.ID       -> ledger_accounts.id
Transaction.ID   -> ledger_transactions.id
Entry.AccountID  -> ledger_entries.account_id
Entry.Amount     -> ledger_entries.amount
```

## Step 7. 실습산출물 작성

`Day14_실습산출물.md`에는 5개 질문만 답합니다.

```text
1. Day12에서 만든 타입 3개는 각각 무엇인가?
2. Day13에서 만든 ValidateTransaction은 어떤 흐름으로 동작하는가?
3. Day13 테스트 4개 중 가장 중요한 테스트는 무엇이라고 생각하는가?
4. 다음 DB 테이블 후보 3개는 각각 어떤 타입과 연결되는가?
5. 다음 구현으로 넘어가기 전에 아직 헷갈리는 부분은 무엇인가?
```

## Step 8. 커밋 메시지

Day14는 보통 코드 커밋이 없습니다.

산출물 문서를 작성했다면 아래 메시지를 사용합니다.

```bash
git add docs/domain/07_Ledger_Core/Day14_Ledger_Core_주간회고와_검증/Day14_실습산출물.md
git commit -m "docs: Day14 Ledger Core 회고 산출물 정리"
```

만약 Day14 중 Day13 코드 오타를 고쳤다면 코드 커밋은 별도로 분리합니다.

```bash
git commit -m "fix: Ledger 균형 검증 코드 오타 수정"
```
