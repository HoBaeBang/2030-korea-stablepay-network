# Wallet, Deposit, Withdrawal

이 문서는 Day 3 퇴근 후 실습에서 직접 채워 넣는 학습 산출물입니다.

관련 가이드: [Wallet, Deposit, Withdrawal 실습 가이드](wallet-deposit-withdrawal-practice-guide.md)

## 한 문장 요약

TODO: Deposit, Withdrawal, Wallet, Key Security를 본인 말로 한 문장으로 정리합니다.

## Deposit과 Withdrawal의 차이

| 구분 | Deposit | Withdrawal |
| --- | --- | --- |
| 자산 방향 | TODO | TODO |
| 온체인 역할 | TODO | TODO |
| 주요 위험 | TODO | TODO |
| Ledger 연결 | TODO | TODO |
| 보안 포인트 | TODO | TODO |

## Deposit 상태 흐름

TODO: 입금 상태 흐름을 작성합니다.

```text
DETECTED
-> CONFIRMING
-> CONFIRMED
-> CREDITED
```

## Withdrawal 상태 흐름

TODO: 출금 상태 흐름을 작성합니다.

```text
REQUESTED
-> APPROVED
-> SIGNED
-> BROADCASTED
-> CONFIRMED
```

## Wallet과 Key Security

TODO: address, private key, signing의 차이를 정리합니다.

## Rust Signer가 필요한 이유

TODO: Rust signer가 왜 필요한지 한 문단으로 정리합니다.

## 최소 테이블 후보

| 테이블 후보 | 필요한 이유 | 저장할 주요 값 |
| --- | --- | --- |
| wallets | TODO | TODO |
| deposit_addresses | TODO | TODO |
| deposits | TODO | TODO |
| withdrawals | TODO | TODO |
| withdrawal_approvals | TODO | TODO |

## 아직 모르는 것과 다음 질문

- TODO
- TODO

## 검증 체크리스트

- [ ] Deposit과 Withdrawal의 차이를 설명할 수 있다.
- [ ] Deposit 상태 흐름을 설명할 수 있다.
- [ ] Withdrawal 상태 흐름을 설명할 수 있다.
- [ ] Wallet address와 private key의 차이를 설명할 수 있다.
- [ ] Rust signer가 왜 필요한지 한 문단으로 설명할 수 있다.
