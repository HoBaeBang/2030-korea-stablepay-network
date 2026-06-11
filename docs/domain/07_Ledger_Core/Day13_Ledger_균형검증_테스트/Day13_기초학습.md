# Day 13 기초학습 - Ledger 균형 검증 테스트 작성

관련 Jira: [SPN-30](https://aslan0.atlassian.net/browse/SPN-30)

Day13은 Ledger Core에서 가장 중요한 규칙 하나를 테스트로 고정하는 날입니다.

```text
Ledger Transaction 안의 debit 합계와 credit 합계는 같아야 한다.
```

Day12에서 `Account`, `Transaction`, `Entry` 타입 초안을 만들었다면, Day13에서는 그 타입들을 이용해 “균형이 맞는 거래만 통과시킨다”는 규칙을 코드로 표현합니다.

## 오늘의 큰 그림

![Day13 Ledger 균형 검증 흐름](../../../confluence/diagrams/spn30-day13-ledger-balance-test.png)

## 오늘의 핵심 문장

```text
Payment는 상태를 바꾸고,
Ledger는 돈의 이동이 맞는지 검증한 뒤 기록해야 한다.
```

Ledger에서 가장 먼저 지켜야 하는 규칙은 “돈이 갑자기 생기거나 사라지지 않아야 한다”는 것입니다.

예를 들어 고객이 10 USDC를 결제했다면, 원장에는 아래처럼 기록될 수 있습니다.

```text
고객 계정        DEBIT   10 USDC
가맹점 지급예정  CREDIT   9.8 USDC
플랫폼 수수료    CREDIT   0.2 USDC
```

이 경우 debit 합계는 10 USDC이고, credit 합계도 10 USDC입니다.

그래서 Ledger Transaction은 균형이 맞습니다.

## 오늘 읽을 순서

| 순서 | 문서 | 목적 |
| --- | --- | --- |
| 1 | [Day13_기초학습.md](Day13_기초학습.md) | 오늘의 목표와 흐름을 확인한다 |
| 2 | [Day13_개념학습.md](Day13_개념학습.md) | debit/credit 균형 검증과 Go map 사용 이유를 이해한다 |
| 3 | [Day13_실습가이드.md](Day13_실습가이드.md) | `service.go`, `service_test.go`를 작성한다 |
| 4 | [Day13_실습산출물.md](Day13_실습산출물.md) | 오늘 작성한 코드와 테스트를 5문항으로 정리한다 |
| 5 | [Day13_검증문제_답변가이드.md](Day13_검증문제_답변가이드.md) | 문제를 먼저 풀고 답변가이드와 비교한다 |

## 출퇴근 학습에서 잡을 것

출퇴근 시간에는 아래 질문을 중심으로 보면 됩니다.

```text
왜 debit과 credit 합계가 같아야 하는가?
왜 이 규칙을 먼저 테스트로 고정하는가?
Go의 map은 합계 계산에 왜 적합한가?
테스트 이름을 한글로 적으면 어떤 장점이 있는가?
context.Context는 오늘 코드에서 어떤 역할을 하는가?
```

## 퇴근 후 코드 작업

퇴근 후에는 큰 기능을 만들지 않습니다.

오늘 할 일은 아래 두 파일만 작성하는 것입니다.

```text
internal/ledger/service.go
internal/ledger/service_test.go
```

오늘은 아래 작업을 하지 않습니다.

```text
DB migration 작성
repository 작성
API 연결
Payment FINALIZED와 Ledger 자동 연결
정산 Settlement 구현
```

## 완료 기준

- [ ] debit과 credit 합계가 왜 같아야 하는지 설명할 수 있다.
- [ ] `map[string]int64`로 통화별 합계를 계산하는 이유를 설명할 수 있다.
- [ ] 한글 테스트 케이스 이름을 읽고 어떤 상황을 검증하는지 알 수 있다.
- [ ] `go test ./internal/ledger -v`가 성공한다.
- [ ] Day13 실습산출물 5문항을 작성할 수 있다.

## 다음 작업 예고

Day14는 새 기능을 크게 추가하지 않고, Day12와 Day13에서 만든 Ledger 타입과 균형 검증 테스트를 복습합니다.

Day14를 통과하면 다음 후보는 DB migration과 repository 초안입니다.
