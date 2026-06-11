# Day 12 기초학습 - Ledger 도메인 타입 초안 작성

관련 Jira: [SPN-29](https://aslan0.atlassian.net/browse/SPN-29)

Day12는 Ledger Core의 첫 코드 진입일입니다.

Day11까지는 Payment 상태 전이 테스트를 통해 “상태를 안전하게 바꾸는 법”을 확인했습니다.

Day12부터는 Payment만으로 부족한 “돈의 이동 기록”을 코드로 표현하기 시작합니다.

## 오늘의 큰 그림

![Day12 Ledger 도메인 타입 흐름](../../../confluence/diagrams/spn29-day12-ledger-types.png)

## 오늘의 핵심 문장

```text
Payment는 결제의 상태를 말하고,
Ledger는 돈의 이동 기록을 말한다.
```

Day12에서 만드는 것은 전체 Ledger 기능이 아닙니다.

오늘은 `internal/ledger/ledger.go`에 Ledger를 구성하는 가장 기본적인 타입 초안을 만듭니다.

```text
Account
Transaction
Entry
```

## 오늘 읽을 순서

| 순서 | 문서 | 목적 |
| --- | --- | --- |
| 1 | [Day12_기초학습.md](Day12_기초학습.md) | 오늘의 목표와 진행 순서를 확인한다 |
| 2 | [Day12_개념학습.md](Day12_개념학습.md) | Ledger가 왜 필요하고 어떤 타입으로 나뉘는지 이해한다 |
| 3 | [Day12_실습가이드.md](Day12_실습가이드.md) | `internal/ledger/ledger.go` 타입 초안을 작성한다 |
| 4 | [Day12_실습산출물.md](Day12_실습산출물.md) | 핵심 5문항으로 이해를 정리한다 |
| 5 | [Day12_검증문제_답변가이드.md](Day12_검증문제_답변가이드.md) | 문제를 먼저 풀고 답변가이드와 비교한다 |

## 출퇴근 학습에서 잡을 것

출퇴근 시간에는 아래 질문을 계속 떠올리면 됩니다.

```text
왜 Payment만으로는 부족한가?
Ledger Account는 무엇인가?
Ledger Transaction은 무엇인가?
Ledger Entry는 무엇인가?
debit과 credit은 왜 처음에 헷갈리는가?
```

## 퇴근 후 코드 작업

퇴근 후에는 아래 하나만 합니다.

```text
internal/ledger/ledger.go 파일을 만들고,
Account, Transaction, Entry 타입 초안을 작성한다.
```

오늘은 아래 작업을 하지 않습니다.

```text
DB migration 작성
repository 작성
service 구현
API 연결
Payment FINALIZED와 Ledger 연결
```

## 완료 기준

- [ ] Payment와 Ledger의 차이를 설명할 수 있다.
- [ ] Account, Transaction, Entry의 역할을 설명할 수 있다.
- [ ] `internal/ledger/ledger.go` 타입 초안을 작성할 준비가 되었다.
- [ ] `go test ./...` 실행 방법을 확인했다.
- [ ] Day12 실습산출물 5문항을 작성할 수 있다.

## 다음 작업 예고

Day12에서 타입 초안을 만든 뒤, 다음 작업에서는 Ledger service 테스트를 먼저 작성합니다.

```text
하나의 Ledger Transaction 안에서 debit과 credit 합계가 맞아야 한다.
```

이 규칙을 테스트로 고정하는 것이 Day13의 자연스러운 후보입니다.
