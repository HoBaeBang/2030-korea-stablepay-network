# Day 12 실습산출물 - Ledger 도메인 타입 초안

관련 Jira: [SPN-29](https://aslan0.atlassian.net/browse/SPN-29)

Day12 산출물은 5개 질문만 작성합니다.

## 1. 오늘 만든 타입은 무엇인가?

작성 예시:

```text
오늘은 `internal/ledger/ledger.go`에 Account, Transaction, Entry 타입 초안을 만들었다.
이 타입들은 Ledger를 구현하기 전에 돈의 이동 기록을 코드로 표현하기 위한 기본 언어다.
```

내 답변:

```text
오늘 만든 타입들은 Ledger를 구성하는 항목들에 대한 타입이다.
구체적으로 Account, Transaction, Entry를 만들었고,
AccountType과 EntryDirection으로 계정의 역할과 원장 항목의 방향을 표현했다.

중요한 점은 Day12에서 Ledger 전체 기능을 만든 것이 아니라,
Ledger를 만들기 위한 기본 도메인 타입 초안을 작성했다는 것이다.
```

## 2. Account, Transaction, Entry는 각각 무엇인가?

작성 예시:

```text
Account는 돈이 기록되는 주체다.
Transaction은 여러 Entry를 하나로 묶는 원장 거래다.
Entry는 실제 돈의 이동 한 줄이다.
```

내 답변:

```text
Account는 돈이 기록되는 주체다.
계좌보다 조금 더 넓은 의미이며, 고객 보유 계정, 가맹점 지급 예정 계정, 플랫폼 수수료 계정처럼 원장에서 돈을 분리해서 기록하기 위한 단위다.

Transaction은 여러 Entry를 하나로 묶어주는 원장 거래다.
블록체인의 transaction hash와는 다른 개념이고, 우리 서비스 내부 원장에서 여러 줄의 돈 이동을 하나의 의미 있는 거래로 묶는 역할을 한다.

Entry는 실제 돈의 이동 한 줄이다.
예를 들어 고객 계정에서 10 USDC가 빠지는 줄, 가맹점 지급 예정 계정에 9.8 USDC가 들어가는 줄, 플랫폼 수수료 계정에 0.2 USDC가 들어가는 줄이 각각 Entry가 될 수 있다.
```

## 3. Amount를 int64로 둔 이유는 무엇인가?

작성 예시:

```text
돈을 float으로 다루면 소수점 오차가 생길 수 있다.
그래서 USDC의 최소 단위 같은 정수 단위로 금액을 저장하기 위해 int64를 사용한다.
```

내 답변:

```text
float을 사용하면 소수점 오차가 발생할 수 있기 때문이다.
돈을 다룰 때는 10.1 같은 소수 금액을 그대로 float으로 저장하기보다, USDC의 최소 단위 같은 정수 단위로 바꿔 저장하는 편이 안전하다.

예를 들어 USDC가 소수점 6자리라면 10 USDC는 10_000_000으로 저장할 수 있다.
그래서 Amount는 int64로 둔다.
```

## 4. 이 타입들이 다음 구현에서 어디로 이어지는가?

작성 예시:

```text
Account, Transaction, Entry 타입은 다음에 Ledger service 테스트와 DB migration으로 이어진다.
특히 Entry의 Direction과 Amount는 debit/credit 합계 검증 테스트에서 사용될 수 있다.
```

내 답변:

```text
Ledger에 대해 만들어둔 타입들은 다음 단계에서 validation 관련 기능과 테스트 코드로 이어진다.

특히 Entry의 Direction과 Amount를 이용해서 하나의 Ledger Transaction 안에서 debit 합계와 credit 합계가 같은지 검증할 수 있다.
그 다음에는 DB migration, repository, Payment FINALIZED와 Ledger 연결로 확장될 수 있다.
```

## 5. 아직 헷갈리는 개념은 무엇인가?

작성 예시:

```text
debit과 credit의 방향이 아직 헷갈린다.
또한 고객 결제, 가맹점 지급 예정, 플랫폼 수수료가 각각 어떤 account에 기록되는지 예시가 더 필요하다.
```

내 답변:

```text
Day12에서 작업한 내용은 Ledger 전체 기능을 만든 것이 아니라, Ledger를 구성하는 항목들에 대한 타입을 구성해놓은 단계라는 점이 중요하다.

아직 헷갈릴 수 있는 부분은 debit/credit의 방향과 Account가 실제 계좌라기보다 원장 기록을 위한 주체라는 점이다.
또한 정산 완료 계정 같은 표현은 실제 구현에서 별도 계정으로 둘지, settlement 상태와 ledger entry 조합으로 표현할지 나중에 다시 설계해야 한다.
```

## 실행 결과

실행한 명령:

```bash
gofmt -w internal/ledger/ledger.go
go test ./...
```

결과:

```text
?       github.com/HoBaeBang/2030-korea-stablepay-network/cmd/api                         [no test files]
?       github.com/HoBaeBang/2030-korea-stablepay-network/internal/httpapi                 [no test files]
ok      github.com/HoBaeBang/2030-korea-stablepay-network/internal/invoice                 (cached)
?       github.com/HoBaeBang/2030-korea-stablepay-network/internal/ledger                  [no test files]
ok      github.com/HoBaeBang/2030-korea-stablepay-network/internal/merchant                (cached)
ok      github.com/HoBaeBang/2030-korea-stablepay-network/internal/payment                 (cached)
?       github.com/HoBaeBang/2030-korea-stablepay-network/internal/platform/config         [no test files]
?       github.com/HoBaeBang/2030-korea-stablepay-network/internal/platform/database       [no test files]
```

## 오늘의 결론

```text
Day12에서 확인한 결론:
Ledger의 구성요소는 Account, Transaction, Entry이며,
AccountType과 EntryDirection은 각각 계정 역할과 원장 항목 방향을 표현한다.

Day12는 Ledger 자체를 완성한 날이 아니라 Ledger를 만들기 위한 타입 초안을 만든 날이다.
다음 구현에서는 이 타입을 기반으로 debit/credit 균형 검증 테스트를 작성한다.

다음 구현으로 넘어가기 전에 남은 질문:
debit/credit 방향과 실제 계정 설계는 Day13 이후 예시를 보며 계속 보강한다.
```
