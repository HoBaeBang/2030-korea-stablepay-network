# Day 10 실습산출물 - 로깅과 테스트 패턴

관련 Jira: [SPN-27](https://aslan0.atlassian.net/browse/SPN-27)

이 문서는 Day10 실습을 마친 뒤 남겨두는 기준 산출물입니다.

Day10을 진행하면서 확인한 결론은 다음입니다.

```text
문서 빈칸을 많이 채우는 방식보다,
작은 코드 작업 하나를 정확히 이해하고 테스트로 고정하는 방식이 더 효과적이다.
```

따라서 이 산출물은 빈칸형 워크시트가 아니라, Day10에서 이해해야 하는 내용을 완성 예시 형태로 정리합니다.

## 실습 흐름

![Day10 로깅과 테스트 피드백 루프](../../../confluence/diagrams/spn27-day10-logging-test-feedback.png)

## 1. 기존 테스트 구조 관찰

오늘 집중해서 본 파일:

```text
internal/payment/service_test.go
```

오늘 집중해서 본 함수:

```text
TestService_UpdatePaymentStatus
```

| 테스트 파일 | 테스트하는 대상 | 확인한 규칙 | 내가 이해한 의미 |
| --- | --- | --- | --- |
| `internal/payment/service_test.go` | `Service.UpdatePaymentStatus` | `PENDING`에서 `ONCHAIN_DETECTED`로 변경할 수 있다 | 블록체인에서 관련 transaction이 감지되면 payment 상태를 다음 단계로 진행할 수 있다 |
| `internal/payment/service_test.go` | `Service.UpdatePaymentStatus` | `FINALIZED`에서 `PENDING`으로 되돌릴 수 없다 | 최종 확정된 결제는 이전 상태로 되돌리면 안 된다 |
| `internal/payment/service_test.go` | `Service.UpdatePaymentStatus` | `FINALIZED`가 되면 `finalized_at`을 저장한다 | 최종 확정 시각은 나중에 정산, 장애 추적, 운영 분석에 필요하다 |
| `internal/payment/service_test.go` | `Service.UpdatePaymentStatus` | `transaction_hash 없이 ONCHAIN_DETECTED로 변경할 수 없다` | 온체인 감지 상태는 실제 블록체인 거래 식별자가 있어야 신뢰할 수 있다 |

관찰 메모:

```text
테스트 이름이 한글이면 테스트 실패 시 어떤 도메인 규칙이 깨졌는지 바로 읽기 쉽다.
Go 테스트는 Java의 테스트 클래스처럼 별도 클래스를 만들지 않고, 같은 패키지의 `_test.go` 파일에 작성한다.
`t.Run`은 하나의 테스트 함수 안에서 세부 테스트 케이스를 나누는 방식이다.
```

## 2. 오늘 추가한 테스트 정리

오늘 추가한 테스트 이름:

```text
transaction_hash 없이 ONCHAIN_DETECTED로 변경할 수 없다
```

이 테스트가 확인하는 규칙:

```text
ONCHAIN_DETECTED는 블록체인에서 transaction을 감지했다는 의미다.
따라서 어떤 transaction을 감지했는지 추적할 수 있는 transaction_hash가 필요하다.
```

내가 실제로 작성한 코드 위치:

```text
파일: internal/payment/service_test.go
함수: TestService_UpdatePaymentStatus
테스트 이름: transaction_hash 없이 ONCHAIN_DETECTED로 변경할 수 없다
```

이 테스트가 막아주는 버그:

```text
transaction_hash 없이 ONCHAIN_DETECTED 상태가 저장되면,
나중에 어떤 온체인 거래를 근거로 결제 상태가 바뀌었는지 추적할 수 없다.

이 경우 운영 중 장애가 발생했을 때 DB의 payment 상태와 실제 블록체인 transaction을 연결하기 어려워진다.
그래서 서비스 계층에서 transaction_hash 없는 ONCHAIN_DETECTED 요청을 막아야 한다.
```

## 3. 로그가 필요한 이벤트 후보

Day10에서 바로 연결되는 로그 후보:

| 이벤트 | 로그가 필요한 이유 | 포함할 값 |
| --- | --- | --- |
| `payment status transition rejected` | transaction_hash 없이 ONCHAIN_DETECTED로 바꾸려는 잘못된 요청을 추적하기 위해 | `payment_id`, `current_status`, `requested_status`, `reason` |
| `payment status changed` | 결제 상태가 실제로 바뀐 이력을 추적하기 위해 | `payment_id`, `old_status`, `new_status`, `transaction_hash` |
| `payment finalized` | 결제가 최종 확정된 시점과 근거를 추적하기 위해 | `payment_id`, `finalized_at`, `transaction_hash` |
| `invalid payment transition requested` | 허용되지 않는 상태 전이 요청을 운영 중 확인하기 위해 | `payment_id`, `from_status`, `to_status`, `reason` |

미래 기능에서 필요할 수 있는 로그 후보:

| 미래 기능 | 로그 후보 | 왜 필요할까 |
| --- | --- | --- |
| Ledger | `ledger transaction created` | 돈의 이동 기록이 생성된 시점을 추적하기 위해 |
| Indexer | `duplicate blockchain event ignored` | 같은 온체인 이벤트를 두 번 반영하지 않았음을 확인하기 위해 |
| Withdrawal | `withdrawal transaction signed` | 출금 transaction 서명 시점을 추적하기 위해 |
| Settlement | `settlement batch created` | 정산 묶음이 언제 어떤 기준으로 만들어졌는지 추적하기 위해 |

오늘 기준으로 아직 더 학습이 필요한 로그 후보:

```text
실제 운영 로그를 어느 레이어에서 남길지 아직 더 학습이 필요하다.
예를 들어 service에서 남길지, handler에서 남길지, middleware에서 남길지 기준이 필요하다.

다만 돈의 상태가 바뀌거나, 온체인 transaction과 연결되거나, 중복 처리를 막는 순간에는 반드시 추적 가능한 로그가 필요하다.
```

## 4. 로그에 포함하면 안 되는 값

| 값 이름 | 포함하면 안 되는 이유 |
| --- | --- |
| `private_key` | 노출되면 지갑 자산을 직접 이동시킬 수 있다 |
| `seed_phrase` | 지갑 전체를 복구할 수 있는 비밀값이다 |
| `full_access_token` | 사용자의 인증 권한이 탈취될 수 있다 |
| `database_password` | DB 접근 권한이 노출될 수 있다 |
| `raw_request_body` | 카드정보, 토큰, 개인정보가 섞여 있을 수 있다 |

마스킹이 필요할 수 있는 값:

```text
email
wallet_address
phone_number
authorization header
external customer id
```

정리:

```text
로그는 장애 추적을 위해 필요하지만, 비밀값을 저장하는 장소가 아니다.
특히 wallet/key/security와 관련된 값은 절대 원문으로 남기면 안 된다.
```

## 5. 한글 subtest 후보

Day10에서 사용할 수 있는 한글 subtest 후보:

```text
1. PENDING에서 ONCHAIN_DETECTED로 변경할 수 있다
2. transaction_hash 없이 ONCHAIN_DETECTED로 변경할 수 없다
3. FINALIZED에서 PENDING으로 되돌릴 수 없다
4. FINALIZED가 되면 finalized_at을 저장한다
5. 지원하지 않는 상태 전이는 실패한다
```

한글 subtest를 쓰는 이유:

```text
테스트 이름이 곧 도메인 규칙 문서가 된다.
실패한 테스트 이름만 봐도 어떤 비즈니스 규칙이 깨졌는지 알 수 있다.
학습 단계에서는 영어 변수명보다 한글 테스트명이 이해를 더 빠르게 만든다.
```

## 6. given / when / then 테스트 패턴

오늘 추가한 테스트를 given/when/then으로 풀면 다음과 같습니다.

```text
테스트 이름:
transaction_hash 없이 ONCHAIN_DETECTED로 변경할 수 없다

given:
PENDING 상태의 payment가 존재한다.

when:
transaction_hash 없이 ONCHAIN_DETECTED 상태 변경을 요청한다.

then:
에러가 발생한다.
payment 상태는 PENDING으로 유지된다.
```

왜 상태 유지까지 확인하는가?

```text
에러가 발생하더라도 내부 상태가 이미 변경되었다면 더 위험하다.
실패한 요청은 아무 것도 바꾸지 않아야 한다.
특히 돈과 결제 상태를 다루는 시스템에서는 실패 시 상태가 보존되는지 확인하는 테스트가 중요하다.
```

## 7. Ledger 구현 전 필요한 테스트 후보

Ledger는 payment 상태만으로 부족한 "돈의 이동 기록"을 남기는 영역입니다.

예를 들어 payment가 `FINALIZED`라고만 되어 있으면 아래 질문에 충분히 답하기 어렵습니다.

```text
누가 누구에게 얼마를 보냈는가?
어떤 통화 단위인가?
수수료는 얼마인가?
같은 온체인 이벤트가 두 번 반영되지는 않았는가?
장애 후 재처리해도 돈이 중복 기록되지 않는가?
```

Ledger 구현 전 테스트 후보:

| 막고 싶은 위험 | 테스트 이름 후보 |
| --- | --- |
| 같은 온체인 이벤트를 두 번 반영함 | 같은 tx_hash와 log_index는 한 번만 Ledger에 반영된다 |
| 금액이 0 이하인데 Ledger entry가 생성됨 | 금액이 0 이하이면 Ledger entry를 생성할 수 없다 |
| debit과 credit 합계가 맞지 않음 | 하나의 Ledger transaction은 debit과 credit 합계가 같아야 한다 |
| payment는 실패했는데 Ledger만 저장됨 | payment 처리 실패 시 Ledger 기록도 생성되지 않는다 |
| finalized 처리를 재시도했을 때 Ledger가 중복 생성됨 | 같은 payment_id는 하나의 Ledger transaction만 생성한다 |

## 8. 오늘 실행한 명령

Day10 코드 실습 기준으로 실행해야 하는 명령:

```bash
gofmt -w internal/payment/service_test.go
go test ./internal/payment
go test ./...
```

실행 결과 기록:

```text
gofmt 또는 go fmt 실행 여부:
Day10 테스트 파일 작성 후 실행 필요

go test ./internal/payment 결과:
payment 패키지 테스트가 성공해야 한다.

go test ./... 결과:
전체 패키지 테스트가 성공해야 한다.
```

## 9. 오늘의 결론

```text
Day 10을 통해 로그와 테스트는 역할이 다르다는 점을 이해했다.

테스트는 개발 시점에 도메인 규칙을 고정하는 장치다.
로그는 운영 시점에 어떤 일이 있었는지 추적하는 장치다.

payment 상태 변경 테스트에서 가장 중요한 점은,
ONCHAIN_DETECTED처럼 온체인 근거가 필요한 상태에는 transaction_hash가 반드시 필요하다는 것이다.

또한 실패한 요청은 에러만 반환하는 것으로 충분하지 않고,
기존 상태를 바꾸지 않았는지도 테스트로 확인해야 한다.
```
