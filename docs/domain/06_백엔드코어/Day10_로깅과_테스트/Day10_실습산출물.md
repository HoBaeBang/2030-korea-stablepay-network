# Day 10 실습산출물 - 로깅과 테스트 패턴

관련 Jira: [SPN-27](https://aslan0.atlassian.net/browse/SPN-27)

## 1. 기존 테스트 구조 관찰

실제 테스트 파일을 보고 테스트가 어떤 도메인 규칙을 설명하는지 적는다.

```text
merchant service test:
invoice service test:
payment service test:
```

## 2. 로그가 필요한 이벤트 후보

| 이벤트 | 로그가 필요한 이유 | 포함할 값 |
| --- | --- | --- |
| payment status changed | 결제 상태 흐름을 추적해야 한다 | payment_id, old_status, new_status, tx_hash |
| invoice created | 결제 요청서 생성 시점을 추적해야 한다 | invoice_id, merchant_id, amount, currency |
| ledger transaction created | 돈의 이동 기록 생성 여부를 추적해야 한다 | ledger_transaction_id, payment_id, amount |
| duplicate event ignored | 중복 이벤트 방어가 실제로 동작했는지 확인해야 한다 | chain, tx_hash, log_index |
| withdrawal signed | 출금 transaction 서명이 완료됐는지 추적해야 한다 | withdrawal_id, signer_id, tx_id |

## 3. 로그에 포함하면 안 되는 값

```text
- private key
- seed phrase
- full access token
- database password
- raw webhook secret
```

email, wallet address, request body 일부는 정책에 따라 마스킹이 필요할 수 있다.

## 4. 한글 subtest 후보

```text
1. 금액이 0이면 invoice를 생성할 수 없다
2. 지원하지 않는 통화이면 payment를 생성할 수 없다
3. transaction_hash 없이 ONCHAIN_DETECTED로 변경할 수 없다
4. 이미 FINALIZED인 payment는 다시 ONCHAIN_DETECTED로 변경할 수 없다
5. 중복 이벤트는 ledger에 두 번 반영되지 않는다
```

## 5. given / when / then 테스트 패턴

```text
테스트 이름: transaction_hash 없이 ONCHAIN_DETECTED로 변경할 수 없다
given: PENDING 상태의 payment가 존재한다
when: transaction_hash 없이 ONCHAIN_DETECTED 상태 변경을 요청한다
then: 에러가 발생하고 payment 상태는 변경되지 않는다
```

## 6. Ledger 구현 전 필요한 테스트 후보

```text
- debit/credit 합계가 0이 아니면 ledger transaction을 생성할 수 없다
- 같은 payment_id로 ledger transaction이 두 번 생성되지 않는다
- payment finalized 처리와 ledger transaction 생성이 하나의 DB transaction으로 묶인다
```

## 7. 오늘의 결론

작성할 내용:

```text
Day 10을 통해 내가 이해한 로그와 테스트의 역할은 ...
```
