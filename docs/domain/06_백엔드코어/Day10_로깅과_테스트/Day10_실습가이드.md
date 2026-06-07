# Day 10 실습가이드 - 로깅과 테스트 패턴

관련 Jira: [SPN-27](https://aslan0.atlassian.net/browse/SPN-27)

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

## 완료 기준

- [ ] 기존 테스트 구조를 확인했다.
- [ ] 로그가 필요한 이벤트 후보를 작성했다.
- [ ] 로그에 포함할 값과 제외할 값을 구분했다.
- [ ] 한글 subtest 후보를 작성했다.
- [ ] Ledger 구현 전 테스트 후보를 작성했다.
