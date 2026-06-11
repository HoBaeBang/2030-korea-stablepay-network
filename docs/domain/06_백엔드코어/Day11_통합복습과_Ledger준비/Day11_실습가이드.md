# Day 11 실습가이드 - 작은 코드 이해 후 Ledger 준비

관련 Jira: [SPN-28](https://aslan0.atlassian.net/browse/SPN-28)

Day11은 긴 산출물을 채우는 날이 아닙니다.

오늘은 Day10에서 만든 payment 테스트를 기준으로, 다음 Ledger 구현으로 넘어가기 전 최소한의 준비만 합니다.

## 실습 흐름

![Day11 Backend Core에서 Ledger로 넘어가는 준비도](../../../confluence/diagrams/spn28-day11-ledger-readiness.png)

## 오늘 실습 목표

오늘의 목표는 딱 5개입니다.

```text
1. Day10 테스트가 어떤 버그를 막는지 설명한다.
2. payment 상태 전이 흐름을 다시 확인한다.
3. 테스트와 로그의 차이를 정리한다.
4. Ledger 첫 구현 후보를 파일 단위로 쪼갠다.
5. go test ./... 로 현재 프로젝트 상태를 확인한다.
```

## Step 1. Day10 테스트 다시 읽기

열 파일:

```text
internal/payment/service_test.go
```

찾을 함수:

```text
TestService_UpdatePaymentStatus
```

찾을 테스트:

```text
transaction_hash 없이 ONCHAIN_DETECTED로 변경할 수 없다
```

읽을 때 볼 것:

| 볼 코드 | 확인할 질문 |
| --- | --- |
| `StatusPending` | 시작 상태가 무엇인가 |
| `StatusOnchainDetected` | 어떤 상태로 바꾸려 하는가 |
| `TransactionHash`가 없음 | 왜 실패해야 하는가 |
| `err == nil` 확인 | 에러가 반드시 발생해야 하는가 |
| `store.current.Status` 확인 | 실패한 요청이 상태를 바꾸지 않았는가 |

## Step 2. payment 상태 흐름을 그림으로 확인

오늘 이해할 상태 흐름:

```text
PENDING
  |
  | transaction_hash 필요
  v
ONCHAIN_DETECTED
  |
  | finalization 처리
  v
FINALIZED
```

중요한 점:

```text
ONCHAIN_DETECTED는 블록체인 transaction을 감지했다는 뜻이다.
그래서 transaction_hash 없이 이 상태로 바뀌면 추적이 불가능해진다.
```

## Step 3. 테스트 실행

먼저 payment 패키지만 실행합니다.

```bash
go test ./internal/payment
```

그 다음 전체 테스트를 실행합니다.

```bash
go test ./...
```

코드 포맷이 필요하면 아래 명령을 실행합니다.

```bash
go fmt ./...
```

## Step 4. Day11 실습산출물 작성

`Day11_실습산출물.md`에는 5개 질문만 답합니다.

```text
1. Day10 테스트가 막는 버그는 무엇인가?
2. 테스트와 로그의 차이는 무엇인가?
3. Payment와 Ledger의 차이는 무엇인가?
4. Ledger 첫 구현 후보는 무엇인가?
5. 지금 가장 헷갈리는 개념은 무엇인가?
```

길게 쓰지 않아도 됩니다.

한 질문당 2~4문장 정도면 충분합니다.

## Step 5. Ledger 첫 구현 후보 정리

다음 구현은 크게 시작하지 않습니다.

아래처럼 파일 단위로 작게 시작합니다.

| 순서 | 후보 | 예상 파일 |
| --- | --- | --- |
| 1 | Ledger 도메인 타입 초안 | `internal/ledger/ledger.go` |
| 2 | Ledger service 테스트 초안 | `internal/ledger/service_test.go` |
| 3 | Ledger service 생성 | `internal/ledger/service.go` |
| 4 | Ledger migration | `migrations/000002_create_ledger_tables.up.sql` |
| 5 | Payment finalized와 Ledger 연결 | `internal/payment/service.go` 또는 별도 application service |

Day11에서는 위 파일을 만들지 않습니다.

오늘은 다음 구현 범위를 이해하는 데서 멈춥니다.

## Step 6. 커밋 기준

Day11에서 문서만 정리했다면:

```bash
git add docs/domain/06_백엔드코어/Day11_통합복습과_Ledger준비/Day11_실습산출물.md
git commit -m "docs: Day11 Backend Core와 Ledger 준비 정리"
```

Day10 테스트 코드가 아직 커밋되지 않았다면 먼저 Day10 코드 커밋을 분리합니다.

```bash
git add internal/payment/service_test.go
git commit -m "test: 결제 상태 전이 검증 테스트 추가"
```

커밋을 분리하는 이유:

```text
테스트 코드 변경과 문서 정리는 목적이 다르기 때문이다.
나중에 Git 기록을 볼 때 어떤 작업이 어떤 이유로 들어갔는지 더 쉽게 추적할 수 있다.
```

## 완료 기준

- [ ] `internal/payment/service_test.go`의 Day10 테스트를 다시 읽었다.
- [ ] `go test ./internal/payment`를 실행했다.
- [ ] `go test ./...`를 실행했다.
- [ ] Day11 실습산출물의 5개 질문에 답했다.
- [ ] Ledger 첫 구현 후보를 파일 단위로 이해했다.
- [ ] Day10 코드 커밋과 Day11 문서 커밋을 섞지 않는 기준을 이해했다.
