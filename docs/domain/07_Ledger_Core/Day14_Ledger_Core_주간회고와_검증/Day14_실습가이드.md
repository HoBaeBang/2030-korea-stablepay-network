# Day 14 실습가이드 - Ledger Core 주간 회고와 검증

관련 Jira: [SPN-31](https://aslan0.atlassian.net/browse/SPN-31)

Day14는 새 기능을 크게 추가하지 않습니다.

오늘은 Day12와 Day13에서 만든 코드가 무엇을 의미하는지 다시 읽고, 테스트를 실행하고, 다음 구현으로 넘어갈 준비가 되었는지 확인합니다.

## 실습 흐름

![Day14 Ledger Core 회고 흐름](../../../confluence/diagrams/spn31-day14-ledger-review.png)

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

## Step 5. 전체 테스트 실행

프로젝트 전체 테스트를 실행합니다.

```bash
go test ./...
```

전체 테스트가 성공해야 Day12~13의 변경이 다른 기능을 깨뜨리지 않은 것입니다.

## Step 6. 다음 DB 테이블 후보 상상하기

아직 DB migration을 작성하지 않습니다.

다만 다음 구현을 준비하기 위해 타입과 테이블이 어떻게 연결될지 생각해봅니다.

```text
Account      -> ledger_accounts
Transaction  -> ledger_transactions
Entry        -> ledger_entries
```

Day14 산출물에 이 연결을 짧게 적습니다.

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
