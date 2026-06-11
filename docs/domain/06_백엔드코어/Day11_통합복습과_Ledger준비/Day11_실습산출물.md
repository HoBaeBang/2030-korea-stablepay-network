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

```

## 2. 테스트와 로그의 차이는 무엇인가?

작성 예시:

```text
테스트는 개발 시점에 규칙이 깨지지 않도록 막는 장치다.
로그는 운영 시점에 이미 발생한 일을 추적하기 위한 기록이다.
```

내 답변:

```text

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

```

## 4. Ledger 첫 구현 후보는 무엇인가?

작성 예시:

```text
가장 먼저 `internal/ledger/ledger.go`에 Account, Transaction, Entry 타입 초안을 만들고,
그 다음 `internal/ledger/service_test.go`에서 debit/credit 합계 검증 테스트를 작성하는 것이 좋다.
```

내 답변:

```text

```

## 5. 지금 가장 헷갈리는 개념은 무엇인가?

작성 예시:

```text
debit과 credit의 방향이 아직 헷갈린다.
특히 사용자가 돈을 냈을 때 우리 시스템 기준으로 어느 계정에 debit이 생기고 어느 계정에 credit이 생기는지 더 예시가 필요하다.
```

내 답변:

```text

```

## 실행 결과

실행한 명령:

```bash
go test ./internal/payment
go test ./...
```

결과:

```text

```

## 오늘의 결론

```text
Day11에서 확인한 결론:

다음 구현으로 넘어가기 전에 남은 질문:
```
