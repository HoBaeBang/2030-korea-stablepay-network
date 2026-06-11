# Day 12 개념학습 - Ledger Account, Transaction, Entry

관련 Jira: [SPN-29](https://aslan0.atlassian.net/browse/SPN-29)

Day12의 출퇴근 자료는 조금 넉넉하게 읽어도 괜찮습니다.

이 문서의 목적은 “오늘 코드로 만들 타입이 왜 필요한지”를 이해하는 것입니다.

## 1. 왜 Ledger가 필요한가?

Payment는 결제의 상태를 관리합니다.

예를 들어 현재 프로젝트에는 이런 상태가 있습니다.

```text
PENDING
ONCHAIN_DETECTED
FINALIZED
```

이 상태만 보면 결제가 어디까지 진행되었는지는 알 수 있습니다.

하지만 아래 질문에는 충분히 답하기 어렵습니다.

```text
누가 돈을 냈는가?
누가 돈을 받을 예정인가?
플랫폼 수수료는 얼마인가?
정산 가능한 금액은 얼마인가?
같은 온체인 transaction이 두 번 반영되지는 않았는가?
장애가 난 뒤 다시 처리해도 돈이 중복 기록되지 않는가?
```

이 질문에 답하기 위해 필요한 것이 Ledger, 즉 원장입니다.

## 2. Ledger를 한 문장으로 말하면

```text
Ledger는 돈의 이동을 추적 가능한 기록으로 남기는 장부다.
```

여기서 중요한 말은 “상태”가 아니라 “이동”입니다.

Payment가 이렇게 말한다면:

```text
이 결제는 FINALIZED 상태다.
```

Ledger는 이렇게 말해야 합니다.

```text
이 결제로 인해 고객 계정에서 10 USDC가 나갔고,
가맹점 지급 예정 계정에 9.8 USDC가 들어갔고,
플랫폼 수수료 계정에 0.2 USDC가 들어갔다.
```

## 3. 세 가지 핵심 타입

Day12에서는 Ledger를 세 가지 타입으로 나눠 생각합니다.

| 타입 | 한글 의미 | 역할 |
| --- | --- | --- |
| Account | 계정 | 돈이 들어오거나 나가는 주체 |
| Transaction | 거래 묶음 | 하나의 원장 기록 묶음 |
| Entry | 원장 항목 | 실제 돈의 증가/감소 한 줄 |

## 4. Account란 무엇인가?

Account는 돈이 기록되는 장소입니다.

은행 앱의 “계좌”와 비슷하게 생각해도 되지만, Ledger의 account는 실제 은행 계좌보다 더 넓은 개념입니다.

예를 들어 StablePay에는 이런 account가 있을 수 있습니다.

```text
고객 보유 계정
가맹점 지급 예정 계정
플랫폼 수수료 계정
정산 완료 계정
```

Day12 타입 초안에서는 account를 이렇게 생각합니다.

```text
Account = 돈의 이동을 기록할 대상
```

## 5. Transaction이란 무엇인가?

Transaction은 여러 Entry를 하나로 묶는 단위입니다.

여기서 말하는 Transaction은 블록체인의 transaction hash와 다릅니다.

| 용어 | 의미 |
| --- | --- |
| Blockchain Transaction | 온체인에서 실행된 거래 |
| Ledger Transaction | 우리 내부 원장에서 하나로 묶은 돈의 이동 기록 |

예를 들어 고객이 10 USDC를 결제하면 내부 원장에서는 하나의 Ledger Transaction이 생길 수 있습니다.

```text
Ledger Transaction: payment finalized for pay_123
```

이 거래 묶음 안에 여러 Entry가 들어갑니다.

## 6. Entry란 무엇인가?

Entry는 실제 돈의 변화 한 줄입니다.

예를 들어 10 USDC 결제를 단순화하면 다음과 같습니다.

```text
Entry 1: 고객 계정에서 10 USDC 감소
Entry 2: 가맹점 지급 예정 계정에 9.8 USDC 증가
Entry 3: 플랫폼 수수료 계정에 0.2 USDC 증가
```

Entry는 “돈이 어떻게 움직였는지”를 가장 구체적으로 남기는 줄입니다.

## 7. debit과 credit은 왜 어렵나?

debit과 credit은 회계 용어입니다.

처음부터 회계 기준을 완벽히 외우려 하면 어렵습니다.

지금은 이렇게 시작하면 됩니다.

```text
Entry에는 방향이 필요하다.
어떤 계정에서는 돈이 빠져나가고,
어떤 계정에서는 돈이 들어온다.
```

Day12에서는 방향을 `debit`, `credit`으로 타입화해 둡니다.

중요한 것은 오늘 당장 모든 회계 규칙을 외우는 것이 아닙니다.

중요한 것은 “돈의 이동 한 줄에는 방향이 있어야 한다”는 점입니다.

## 8. 왜 amount는 int64로 다루는가?

돈을 `float64`로 다루면 소수점 오차가 생길 수 있습니다.

예를 들어 0.1 + 0.2 같은 계산이 정확히 0.3으로 표현되지 않을 수 있습니다.

그래서 금융 시스템에서는 보통 최소 단위 정수로 저장합니다.

예를 들어 USDC가 소수점 6자리라면:

```text
1 USDC = 1_000_000
10 USDC = 10_000_000
0.2 USDC = 200_000
```

Day12 타입 초안에서는 amount를 `int64`로 둡니다.

```go
Amount int64
```

## 9. Day12 코드에서 만들 타입 후보

오늘 만들 파일:

```text
internal/ledger/ledger.go
```

오늘 만들 타입:

```text
Account
Transaction
Entry
AccountType
EntryDirection
```

각 타입의 목적:

| 타입 | 목적 |
| --- | --- |
| `Account` | 돈이 기록될 계정 |
| `Transaction` | 하나의 원장 거래 묶음 |
| `Entry` | 원장 거래 안의 한 줄 |
| `AccountType` | 계정 종류를 제한 |
| `EntryDirection` | debit/credit 방향을 제한 |

## 10. 오늘의 핵심 결론

```text
Ledger는 Payment의 다음 단계가 아니라, Payment와 다른 책임을 가진 도메인이다.

Payment는 결제 상태를 관리하고,
Ledger는 돈의 이동 기록을 남긴다.

Day12에서는 전체 Ledger 기능을 구현하지 않고,
Ledger를 표현할 최소 타입 언어를 Go 코드로 만든다.
```
