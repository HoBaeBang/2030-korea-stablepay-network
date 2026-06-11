# Day 11 실습산출물 - Backend Core에서 Ledger로 넘어가기

관련 Jira: [SPN-28](https://aslan0.atlassian.net/browse/SPN-28)

이 문서는 긴 빈칸형 산출물이 아닙니다.

Day11에서는 아래 5개 질문에만 답합니다.

## 1. Day10 테스트가 막는 버그는 무엇인가?

작성 예시:

```text
transaction_hash 없이 ONCHAIN_DETECTED 상태가 저장되면,
나중에 어떤 블록체인 transaction을 근거로 결제 상태가 바뀌었는지 추적할 수 없다.
이 테스트는 온체인 근거 없는 상태 변경을 막는다.
```

내 답변:

```text
transaction_hash 없이 상태만 ONCHAIN_DETECTED로 변경되면,
나중에 어떤 블록체인 transaction을 근거로 결제 상태가 바뀌었는지 추적하기 어렵다.

그 결과 같은 transaction을 중복 처리했는지 확인하기 어려워지고,
장애가 생겼을 때 DB의 payment 상태와 실제 온체인 거래를 비교하거나 복구하기 어려워진다.

따라서 이 테스트는 온체인 근거가 없는 상태 변경을 막고,
나중에 Ledger나 Reconciliation으로 이어질 수 있는 추적 기준을 지키게 해준다.
```

리뷰 메모:

```text
처음 답변에서 중복 처리 위험을 떠올린 것은 좋은 방향이다.
다만 "로그를 남길 수 없다"보다는 "로그를 남겨도 핵심 식별자인 transaction_hash가 없어 추적력이 떨어진다"가 더 정확하다.
"롤백이 불가능하다"도 너무 강한 표현이고, 이 경우에는 "대사와 복구가 어려워진다"가 더 정확하다.
```

## 2. 테스트와 로그의 차이는 무엇인가?

작성 예시:

```text
테스트는 개발 시점에 규칙이 깨지지 않도록 막는 장치다.
로그는 운영 시점에 이미 발생한 일을 추적하기 위한 기록이다.
```

내 답변:

```text
테스트는 개발 시점에 작성하며, 꼭 지켜져야 하는 도메인 규칙이 깨지지 않도록 검증하는 장치다.

예를 들어 transaction_hash 없이 ONCHAIN_DETECTED로 변경할 수 없다는 규칙은 테스트로 고정할 수 있다.

로그는 운영 중 발생한 일을 기록해서, 나중에 문제가 생겼을 때 원인 파악과 장애 대응에 활용하는 기록이다.
로그는 rollback 자체를 대신하지는 않지만, 어떤 요청과 상태 변경이 있었는지 추적하는 근거가 된다.
```

리뷰 메모:

```text
테스트와 로그의 차이를 잘 잡았다.
다만 로그는 "롤백을 하기 위한 기록"이라기보다 "문제 원인을 찾고 복구 판단을 하기 위한 추적 기록"에 가깝다.
```

## 3. Payment와 Ledger의 차이는 무엇인가?

작성 예시:

```text
Payment는 결제가 현재 어떤 상태인지 관리한다.
Ledger는 돈이 누구에게서 누구에게 얼마만큼 이동했는지 기록한다.
Payment만 있으면 상태는 알 수 있지만 돈의 이동 내역을 충분히 설명하기 어렵다.
```

내 답변:

```text
Payment는 결제의 상태를 관리한다.

예를 들어 PENDING, ONCHAIN_DETECTED, FINALIZED처럼 결제가 어느 단계에 있는지를 표현한다.

Ledger는 돈이 누구에게서 누구에게, 어떤 통화로, 얼마만큼 이동했는지 기록하는 장부다.

Payment는 결제 상태를 설명하고, Ledger는 돈의 이동 내역을 설명한다.
따라서 Payment만 있으면 결제가 완료되었는지는 알 수 있지만,
돈이 어떻게 이동했고 수수료나 정산 금액이 어떻게 나뉘었는지는 충분히 설명하기 어렵다.
```

리뷰 메모:

```text
"Payment는 상태, Ledger는 기록"이라는 핵심 구분은 잘 잡았다.
조금 더 정확히는 Ledger를 "결제의 기록"이라고만 하기보다 "돈의 이동 기록"이라고 표현하는 것이 좋다.
```

## 4. Ledger 첫 구현 후보는 무엇인가?

작성 예시:

```text
가장 먼저 `internal/ledger/ledger.go`에 Account, Transaction, Entry 타입 초안을 만들고,
그 다음 `internal/ledger/service_test.go`에서 debit/credit 합계 검증 테스트를 작성하는 것이 좋다.
```

내 답변:

```text
Ledger 첫 구현은 바로 DB나 API부터 만들기보다, 먼저 도메인 타입을 작게 정의하는 것이 좋다.

첫 번째 후보는 `internal/ledger/ledger.go`를 만들고 Account, Transaction, Entry 타입 초안을 작성하는 것이다.

두 번째 후보는 `internal/ledger/service_test.go`에서 하나의 Ledger Transaction 안에 debit과 credit 합계가 맞아야 한다는 테스트를 먼저 작성하는 것이다.

그 다음에 service 구현, migration, repository, Payment FINALIZED와 Ledger 연결 순서로 확장하는 것이 좋다.
```

리뷰 메모:

```text
이 부분이 헷갈린 것은 자연스럽다.
Ledger는 개념상 Account, Transaction, Entry가 한 번에 나오기 때문에 처음에는 전체가 크게 느껴진다.

다음 Day12에서는 전체 Ledger를 구현하지 않고,
`internal/ledger/ledger.go`에 타입 초안을 만드는 것부터 시작하면 된다.
```

## 5. 지금 가장 헷갈리는 개념은 무엇인가?

작성 예시:

```text
debit과 credit의 방향이 아직 헷갈린다.
특히 사용자가 돈을 냈을 때 우리 시스템 기준으로 어느 계정에 debit이 생기고 어느 계정에 credit이 생기는지 더 예시가 필요하다.
```

내 답변:

```text
Ledger를 구성하는 정보와 좋은 Ledger 설계 기준이 아직 헷갈린다.

특히 Account, Transaction, Entry가 각각 무엇을 책임지는지,
그리고 debit과 credit을 어떤 방향으로 기록해야 하는지 더 학습이 필요하다.

또한 Ledger가 Payment, Settlement, Reconciliation과 어떻게 연결되는지도 다음 단계에서 더 확인해야 한다.
```

리뷰 메모:

```text
현재 남은 질문은 아주 적절하다.
Day12에서는 Ledger 구성 요소를 코드 타입으로 만들면서 이 질문을 줄여가는 것이 좋다.
```

## 실행 결과

실행한 명령:

```bash
go test ./internal/payment
go test ./...
```

결과:

```text
go test ./internal/payment: 통과
go test ./...: 통과
```

확인된 주요 패키지:

```text
ok github.com/HoBaeBang/2030-korea-stablepay-network/internal/invoice
ok github.com/HoBaeBang/2030-korea-stablepay-network/internal/merchant
ok github.com/HoBaeBang/2030-korea-stablepay-network/internal/payment
```

## 오늘의 결론

```text
Day11에서 확인한 결론:
Payment와 Ledger의 차이를 정리했고,
Backend Core 테스트를 확인한 이유가 다음 Ledger 구현의 안전한 출발점과 연결된다는 점을 이해했다.

Payment는 상태를 설명하고, Ledger는 돈의 이동 기록을 설명한다.
Day10의 payment 상태 전이 테스트는 나중에 Ledger와 온체인 transaction을 안전하게 연결하기 위한 기본 규칙을 지킨다.

다음 구현으로 넘어가기 전에 남은 질문:
Ledger의 구성 요소인 Account, Transaction, Entry를 어떤 필드로 설계해야 하는가?
debit과 credit은 실제 결제 예시에서 어느 방향으로 기록되는가?
좋은 Ledger를 만들기 위해 service에서 어떤 검증을 먼저 해야 하는가?
```

## 다음 작업 후보

```text
Day12에서는 `internal/ledger/ledger.go`에 Ledger 도메인 타입 초안을 만든다.

목표는 전체 Ledger 구현이 아니라,
Account, Transaction, Entry가 각각 무엇을 표현하는지 코드로 확인하는 것이다.
```
