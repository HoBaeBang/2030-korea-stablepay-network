# Day 11 실습가이드 - Backend Core 통합 복습과 Ledger 구현 준비

관련 Jira: [SPN-28](https://aslan0.atlassian.net/browse/SPN-28)

## 실습 흐름

![Day11 Backend Core에서 Ledger로 넘어가는 준비도](../../../confluence/diagrams/spn28-day11-ledger-readiness.png)

오늘 실습은 “복습했다”로 끝나면 안 됩니다. Day8~10에서 정리한 공통 기반이 Ledger 구현의 어떤 위험을 줄여주는지 연결해서 설명해야 합니다.

## 실습 목표

`Day11_실습산출물.md`에 다음 내용을 작성합니다.

1. Backend Core 통합 체크리스트
2. Day 8~10에서 가장 약한 개념
3. Ledger 구현 전 위험 요소
4. 다음 구현 티켓 후보
5. SPN-2 에픽 완료 판단

## Step 1. Day 8~10 산출물 다시 보기

확인할 문서:

```text
Day8_실습산출물.md
Day9_실습산출물.md
Day10_실습산출물.md
```

## Step 2. Backend Core 체크리스트 작성

아래 질문에 답합니다.

```text
공통 에러 응답 형식은 정해졌는가?
validation 위치는 정해졌는가?
config 후보는 정해졌는가?
로그 후보는 정해졌는가?
테스트 패턴은 정해졌는가?
```

체크리스트를 작성할 때는 단순히 “했다/안 했다”가 아니라, Ledger 구현과 어떻게 연결되는지를 같이 씁니다.

| 항목 | Ledger와 연결되는 이유 |
| --- | --- |
| 공통 에러 응답 | Ledger 실패를 API에서 일관되게 표현해야 한다 |
| validation | 잘못된 금액과 중복 요청을 원장에 반영하지 않아야 한다 |
| config | DB와 향후 RPC/signer 설정 누락을 시작 시점에 잡아야 한다 |
| logging | 돈의 이동 기록 생성 시점을 추적해야 한다 |
| test pattern | 복식부기 불변식과 중복 방어를 자동 검증해야 한다 |

## Step 3. Ledger 구현 전 위험 요소 작성

예시:

```text
ledger entry의 debit/credit 방향을 헷갈릴 수 있다.
payment finalized를 중복 처리하면 ledger가 중복 생성될 수 있다.
DB transaction 없이 ledger transaction과 entry를 따로 저장하면 정합성이 깨질 수 있다.
```

## Step 4. 다음 구현 티켓 후보 작성

예시:

```text
Ledger migration 작성
Ledger repository 작성
Ledger service 작성
Payment finalized와 Ledger 연결
Ledger 테스트 작성
```

티켓 후보는 너무 추상적으로 쓰지 않습니다.

좋은 후보:

```text
Ledger account/transaction/entry migration 작성
Ledger transaction 생성 service와 debit/credit 합계 검증 구현
Payment finalized 시 ledger transaction 생성 연결
동일 payment에 대한 ledger 중복 생성 방지 테스트 작성
```

아직 피해야 하는 후보:

```text
Ledger 전체 구현
정산 전체 구현
블록체인 연동 전체 구현
```

너무 큰 티켓은 하루나 이틀 단위로 검증하기 어렵습니다.

## Step 5. SPN-2 완료 판단

다음 기준으로 판단합니다.

```text
Backend Core를 문서로 설명할 수 있는가?
공통 기반 구현 후보가 구체적인가?
Ledger 구현을 시작할 때 필요한 준비가 되어 있는가?
```

판단 결과는 다음 세 가지 중 하나로 적습니다.

| 판단 | 의미 |
| --- | --- |
| 완료 가능 | Ledger 첫 구현으로 넘어갈 수 있다 |
| 일부 보강 후 완료 | 특정 문서나 구현 후보만 보강하면 된다 |
| 완료 보류 | 공통 기반이 아직 너무 약해 Ledger로 넘어가면 위험하다 |

## Step 6. 코드 점검 실습 - Ledger 구현 전 마지막 확인

Day11은 새 기능을 크게 추가하는 날이 아닙니다.

대신 Day9~10에서 했어야 하는 코드 작업이 실제로 반영되었는지 확인하고, Ledger 구현 전에 깨진 테스트가 없는지 점검합니다.

확인할 파일:

```text
cmd/api/main.go
internal/platform/config/config.go
internal/payment/service.go
internal/payment/service_test.go
internal/merchant/service_test.go
internal/invoice/service_test.go
```

## Step 7. 코드 기준 체크리스트

아래 질문을 코드에서 직접 확인합니다.

| 확인 항목 | 확인 파일 | 확인 결과 |
| --- | --- | --- |
| `main.go`가 설정 로딩을 직접 많이 하지 않는가 | `cmd/api/main.go` |  |
| config 패키지가 설정 기본값을 관리하는가 | `internal/platform/config/config.go` |  |
| payment 상태 전이 규칙이 service에 있는가 | `internal/payment/service.go` |  |
| 상태 전이 규칙이 테스트로 고정되어 있는가 | `internal/payment/service_test.go` |  |
| 한글 subtest로 실패 원인을 읽을 수 있는가 | `*_test.go` |  |

## Step 8. 검증 명령 실행

아래 명령을 실행합니다.

```bash
go test ./...
```

테스트가 실패하면 바로 넘어가지 말고 다음 순서로 확인합니다.

```text
1. 어떤 패키지가 실패했는가?
2. 어떤 테스트 이름이 실패했는가?
3. 에러 메시지는 무엇인가?
4. 코드 문제인가, 테스트 기대값 문제인가?
5. 수정 후 다시 go test ./...를 실행했는가?
```

## Step 9. Ledger 첫 구현 전에 필요한 코드 작업 후보 정리

Day11에서 실제 Ledger 코드를 바로 작성하지 않는 이유는, Ledger는 migration, repository, service, 테스트가 함께 들어가는 큰 작업이기 때문입니다.

대신 다음 구현을 작게 나누어 준비합니다.

| 다음 코드 작업 후보 | 만들 파일 후보 | 먼저 확인할 것 |
| --- | --- | --- |
| Ledger 모델 정의 | `internal/ledger/ledger.go` | account, transaction, entry 용어 이해 |
| Ledger migration 작성 | `migrations/000002_create_ledger_tables.up.sql` | debit/credit 구조 |
| Ledger repository 작성 | `internal/ledger/repository.go` | DB transaction 필요 여부 |
| Ledger service 작성 | `internal/ledger/service.go` | 합계 0 검증, 중복 방어 |
| Ledger 테스트 작성 | `internal/ledger/service_test.go` | given/when/then 테스트 패턴 |

## Step 10. 실습산출물 작성

코드 점검이 끝나면 `Day11_실습산출물.md`에 다음 내용을 작성합니다.

```text
go test ./... 실행 결과
Day9 config 코드가 이해되는지
Day10 payment 테스트가 이해되는지
Ledger 구현 전 가장 위험해 보이는 부분
다음 구현 티켓 후보
```

## Step 11. 완성본 확인

Day11은 새 코드를 크게 작성하는 날이 아니라, Day9~10 코드 작업이 제대로 반영되었는지 확인하는 날입니다.

따라서 Day11의 완성본은 특정 Go 파일 하나가 아니라 아래 상태입니다.

| 확인 대상 | 완성 기준 |
| --- | --- |
| `internal/platform/config/config.go` | `Config` 구조체와 `Load()` 함수가 있다 |
| `cmd/api/main.go` | `config.Load()`를 사용하고 `os.Getenv`를 직접 호출하지 않는다 |
| `internal/payment/service_test.go` | `transaction_hash 없이 ONCHAIN_DETECTED로 변경할 수 없다` 테스트가 있다 |
| 전체 테스트 | `go test ./...`가 성공한다 |
| 실습산출물 | Day8~10 기반, Ledger 위험 요소, 다음 티켓 후보가 작성되어 있다 |

Day11에서 확인할 최종 명령:

```bash
go fmt ./...
go test ./...
git status
```

`git status`에서 Day9 또는 Day10 실습 파일이 남아 있다면 먼저 해당 날짜의 커밋을 완료한 뒤 Day11 정리 커밋을 진행합니다.

## Step 12. 커밋 메시지

Day11은 코드 점검과 산출물 정리 중심이므로, 보통 산출물 문서만 커밋합니다.

```bash
git status
git add docs/domain/06_백엔드코어/Day11_통합복습과_Ledger준비/Day11_실습산출물.md
git commit -m "docs: Backend Core 통합 점검 산출물 정리"
```

만약 Day11 과정에서 Day9~10 코드 누락을 함께 수정했다면 커밋을 섞지 않습니다.

이 경우 아래처럼 나눕니다.

```text
1. Day9 config 코드 수정 커밋
2. Day10 payment 테스트 수정 커밋
3. Day11 산출물 정리 커밋
```

커밋을 나누는 이유:

```text
설정 코드 변경, 테스트 코드 변경, 학습 산출물 정리는 변경 목적이 다르기 때문이다.
나중에 Git 기록을 볼 때 어떤 작업이 어떤 이유로 들어갔는지 추적하기 쉬워진다.
```

## 완료 기준

- [ ] 통합 체크리스트를 작성했다.
- [ ] 약한 개념을 정리했다.
- [ ] Ledger 구현 전 위험 요소를 작성했다.
- [ ] 다음 구현 티켓 후보를 작성했다.
- [ ] SPN-2 완료 가능 여부를 판단했다.
- [ ] Day9 config 코드 작업 반영 여부를 확인했다.
- [ ] Day10 payment 테스트 추가 여부를 확인했다.
- [ ] `go test ./...`를 실행했다.
- [ ] Ledger 첫 구현 후보를 파일 단위로 나누어 작성했다.
- [ ] 완성 기준 표와 현재 프로젝트 상태를 비교했다.
- [ ] 커밋 메시지를 확인했다.
