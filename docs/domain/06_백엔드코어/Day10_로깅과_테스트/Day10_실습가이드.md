# Day 10 실습가이드 - 로깅과 테스트 패턴

관련 Jira: [SPN-27](https://aslan0.atlassian.net/browse/SPN-27)

## 실습 흐름

![Day10 로깅과 테스트 피드백 루프](../../../confluence/diagrams/spn27-day10-logging-test-feedback.png)

오늘 실습은 테스트 파일을 읽는 데서 끝나지 않습니다. 어떤 규칙은 테스트로 고정하고, 어떤 사건은 로그로 남겨야 하는지 구분하는 것이 핵심입니다.

## 실습 목표

`Day10_실습산출물.md`에 다음 내용을 작성합니다.

1. 로그가 필요한 이벤트 후보
2. 로그에 포함할 값
3. 로그에 포함하면 안 되는 값
4. 한글 subtest 테스트 패턴
5. Ledger 구현 전 테스트 후보

## Step 1. 기존 테스트 파일 확인

확인 파일:

```text
internal/merchant/service_test.go
internal/invoice/service_test.go
internal/payment/service_test.go
```

확인할 질문:

```text
테스트 함수 이름은 어떻게 되어 있는가?
t.Run을 사용하고 있는가?
given/when/then 흐름이 보이는가?
테스트가 도메인 규칙을 설명하고 있는가?
```

## Step 2. 로그 후보 작성

다음 상황에서 어떤 로그가 필요할지 작성합니다.

```text
payment status changed
invoice created
future ledger transaction created
future indexer duplicate event ignored
future withdrawal signed
```

각 로그 후보에는 다음 값을 같이 생각합니다.

| 질문 | 예시 |
| --- | --- |
| 어떤 일이 발생했는가? | payment status changed |
| 어떤 리소스인가? | payment_id, invoice_id |
| 이전 상태와 다음 상태는 무엇인가? | old_status, new_status |
| 온체인 식별자가 있는가? | tx_hash, chain |
| 사용자가 보면 안 되는 값이 섞였는가? | private key, token |

## Step 3. 로그에 포함할 값 정리

예시:

```text
payment_id
invoice_id
merchant_id
old_status
new_status
tx_hash
chain
```

## Step 4. 로그에 포함하면 안 되는 값 정리

예시:

```text
private key
raw secret
database password
full access token
```

## Step 5. 테스트 패턴 작성

아래 형식으로 테스트 후보를 작성합니다.

```text
테스트 이름:
given:
when:
then:
```

예시:

```text
테스트 이름: ONCHAIN_DETECTED 상태로 바꿀 때 transaction_hash가 없으면 실패한다
given: PENDING 상태의 payment가 존재한다
when: transaction_hash 없이 ONCHAIN_DETECTED로 상태 변경을 요청한다
then: bad_request 계열 에러가 발생하고 상태는 변경되지 않는다
```

Day10에서는 실제 테스트 코드를 완성하지 않아도 됩니다.

하지만 Day11 이후 Ledger 구현에 들어갔을 때 바로 테스트로 옮길 수 있을 만큼 구체적인 테스트 후보를 작성해야 합니다.

## 완료 기준

- [ ] 기존 테스트 구조를 확인했다.
- [ ] 로그가 필요한 이벤트 후보를 작성했다.
- [ ] 로그에 포함할 값과 제외할 값을 구분했다.
- [ ] 한글 subtest 후보를 작성했다.
- [ ] Ledger 구현 전 테스트 후보를 작성했다.
