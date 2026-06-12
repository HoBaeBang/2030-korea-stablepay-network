# Day 14 개념학습 - Ledger Core 주간 회고

관련 Jira: [SPN-31](https://aslan0.atlassian.net/browse/SPN-31)

## 1. Day12와 Day13은 무엇을 만든 흐름인가?

Day12와 Day13은 Ledger Core의 뼈대를 잡는 단계입니다.

아직 결제 완료 시 자동으로 Ledger를 쓰거나, DB에 저장하거나, 정산을 계산하지 않습니다.

Day14는 이 뼈대가 제대로 이해되었는지 확인하는 날입니다.

즉 Day14의 핵심은 “새로운 코드를 많이 작성하는 것”이 아니라 아래 질문에 답할 수 있는지 확인하는 것입니다.

```text
내가 만든 타입이 무엇인지 설명할 수 있는가?
내가 만든 테스트가 어떤 위험을 막는지 설명할 수 있는가?
다음 DB 저장 단계가 왜 필요한지 설명할 수 있는가?
```

대신 아래 두 가지를 먼저 잡습니다.

```text
1. 원장에서 사용할 도메인 언어
2. 원장이 반드시 지켜야 할 첫 번째 안전 규칙
```

## 2. Day12: 도메인 언어 만들기

Day12에서 만든 핵심 타입은 3개입니다.

| 타입 | 한글 의미 | 역할 |
| --- | --- | --- |
| Account | 원장 계정 | 돈이 기록되는 주체 |
| Transaction | 원장 거래 묶음 | 여러 Entry를 하나의 거래로 묶는다 |
| Entry | 원장 항목 | 실제 돈의 이동 한 줄 |

이 타입들은 나중에 DB 테이블과 거의 직접 연결됩니다.

```text
Account      -> ledger_accounts
Transaction  -> ledger_transactions
Entry        -> ledger_entries
```

## 3. Day13: 안전 규칙 테스트로 고정하기

Day13에서 만든 핵심 규칙은 아래입니다.

```text
하나의 Ledger Transaction 안에서 debit 합계와 credit 합계는 같아야 한다.
```

이 규칙은 Ledger의 가장 기본적인 안전장치입니다.

균형이 맞지 않는 거래가 저장되면 아래 문제가 생깁니다.

```text
돈이 실제보다 더 생긴 것처럼 보일 수 있다.
돈이 실제보다 사라진 것처럼 보일 수 있다.
정산 결과가 틀릴 수 있다.
장애 복구 때 어떤 기록이 맞는지 판단하기 어려워진다.
```

## 4. Payment와 Ledger는 왜 분리되는가?

Payment는 결제 상태를 다룹니다.

예를 들어:

```text
PENDING
ONCHAIN_DETECTED
FINALIZED
FAILED
```

Ledger는 돈의 이동을 다룹니다.

예를 들어:

```text
고객 계정에서 10 USDC 차감
가맹점 지급 예정 계정에 9.8 USDC 반영
플랫폼 수수료 계정에 0.2 USDC 반영
```

Payment와 Ledger를 하나로 섞으면 처음에는 편해 보일 수 있습니다.

하지만 서비스가 커지면 아래 문제가 생깁니다.

```text
결제 상태와 돈의 이동 기록이 섞인다.
장애 복구가 어려워진다.
중복 처리 여부를 판단하기 어려워진다.
정산과 회계성 기록을 분리해서 보기 어렵다.
```

그래서 두 영역을 분리합니다.

## 5. 테스트 우선 흐름이 왜 중요한가?

Day13에서는 DB 저장보다 테스트를 먼저 작성했습니다.

이유는 단순합니다.

```text
저장하기 전에 무엇이 올바른 데이터인지 먼저 정해야 한다.
```

DB에 저장하는 코드를 먼저 만들면 잘못된 Ledger Transaction도 저장될 수 있습니다.

반대로 균형 검증 테스트를 먼저 만들어두면 다음 작업에서 아래 흐름이 됩니다.

```text
요청 수신
-> Ledger Entry 생성
-> 균형 검증
-> DB 저장
```

## 6. 다음 단계는 무엇인가?

Day14 이후 자연스러운 다음 단계는 DB migration입니다.

후보 테이블은 아래와 같습니다.

| 테이블 후보 | 역할 |
| --- | --- |
| ledger_accounts | 원장 계정 저장 |
| ledger_transactions | 원장 거래 묶음 저장 |
| ledger_entries | 원장 항목 저장 |

그리고 그 다음 단계는 repository입니다.

```text
Service가 검증한다.
Repository가 저장한다.
DB가 기록을 보존한다.
```

이 흐름은 Java 백엔드의 레이어드 구조와 비교하면 이렇게 볼 수 있습니다.

| Java식 감각 | 현재 Go 프로젝트에서의 위치 |
| --- | --- |
| Domain Model | `internal/ledger/ledger.go` |
| Service | `internal/ledger/service.go` |
| Repository | 다음 단계에서 만들 후보 |
| Migration/DDL | 다음 단계에서 만들 후보 |
| Controller/API | 아직 만들지 않음 |

Go에서는 패키지와 파일 이름이 Java처럼 항상 `controller`, `service`, `repository` 폴더로 고정되지는 않습니다.

하지만 지금 Ledger Core에서는 학습을 위해 역할을 명확히 나누고 있습니다.

```text
ledger.go       -> 도메인 타입
service.go      -> 도메인 규칙
service_test.go -> 규칙 검증
```

## 7. 아직 구현하지 않은 것

아래 내용은 아직 구현하지 않았습니다.

```text
Payment FINALIZED 시 Ledger 자동 생성
DB migration
Ledger repository
idempotency key 중복 방지
Settlement와 연결
Deposit/Withdrawal과 연결
```

이것들은 이후 단계에서 하나씩 붙입니다.

Day14에서 중요한 것은 “아직 안 만든 것을 아는 것”입니다.

## 8. Day14에서 스스로 확인할 질문

아래 질문에 답이 막히면 Day13 자료로 돌아가서 다시 봅니다.

```text
왜 Payment 상태만으로 돈의 이동을 설명할 수 없는가?
왜 Entry는 최소 2개 이상이어야 하는가?
왜 금액은 float이 아니라 int64인가?
왜 debit과 credit 합계가 다르면 저장하면 안 되는가?
왜 service 테스트를 DB migration보다 먼저 작성했는가?
```

## 9. 오늘의 결론

```text
Ledger Core는 Payment 이후에 붙는 부가 기능이 아니라,
돈의 이동을 안전하게 기록하기 위한 별도 도메인이다.

Day12는 언어를 만들었고,
Day13은 첫 번째 안전 규칙을 테스트로 고정했다.
Day14는 이 흐름을 복습하고 다음 구현 준비 상태를 점검한다.
```
