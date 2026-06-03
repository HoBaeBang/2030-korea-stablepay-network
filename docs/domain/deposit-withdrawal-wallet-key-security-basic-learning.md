# Deposit, Withdrawal, Wallet, Key Security 기초 학습

관련 Jira: [SPN-20](https://aslan0.atlassian.net/browse/SPN-20)

이 문서는 Phase 2 Day 3 학습 허브입니다.

Day 3의 목표는 입금, 출금, 지갑, 키 보안이 왜 일반 CRUD와 다르게 다뤄져야 하는지 이해하고, 퇴근 후 직접 `docs/domain/wallet-deposit-withdrawal.md`를 작성할 준비를 하는 것입니다.

## 오늘의 큰 그림

![SPN-20 Day 3 학습 흐름](../confluence/diagrams/spn20-day3-learning-flow.png)

이 그림은 Day 3에서 먼저 개념을 잡고, 돈의 방향과 보안 경계를 구분한 뒤, 퇴근 후 직접 도메인 문서로 정리하는 흐름을 보여줍니다.

## 오늘의 목표

1. Deposit과 Withdrawal의 차이를 설명할 수 있다.
2. Wallet address, private key, signing의 의미를 설명할 수 있다.
3. 출금이 요청, 승인, 서명, 전송, 확정 단계로 나뉘는 이유를 이해한다.
4. 개인키를 일반 DB 필드처럼 다루면 안 되는 이유를 설명한다.
5. Rust signer가 왜 별도 컴포넌트 후보인지 한 문단으로 정리한다.

## 읽기 순서

| 순서 | 문서 | 목적 |
| --- | --- | --- |
| 1 | [Deposit, Withdrawal, Wallet, Key Security 개념 학습](deposit-withdrawal-wallet-key-security-concepts.md) | 출퇴근 시간에 읽을 핵심 개념 자료 |
| 2 | [Wallet, Deposit, Withdrawal 실습 가이드](wallet-deposit-withdrawal-practice-guide.md) | 퇴근 후 직접 작성할 문서 가이드 |
| 3 | [Deposit, Withdrawal, Wallet, Key Security 검증문제와 답변가이드](deposit-withdrawal-wallet-key-security-verification-guide.md) | 학습 후 스스로 확인할 문제와 답변 기준 |

## 오늘 꼭 잡아야 하는 문장

```text
입금은 온체인에서 들어온 자산을 감지하고 내부 장부에 반영하는 흐름이고,
출금은 내부 장부에서 출금 가능 여부를 확인한 뒤 안전하게 서명하고 온체인으로 전송하는 흐름이다.

지갑 주소는 공개되어도 되지만, 개인키는 절대 일반 데이터처럼 다루면 안 된다.
```

## 퇴근 후 작업의 원칙

퇴근 후 작업은 사용자가 직접 진행합니다.

Codex가 대신 완성 문서를 작성하는 것이 아니라, 사용자가 아래 작업을 직접 수행합니다.

1. GitHub repo에서 `docs/domain/wallet-deposit-withdrawal.md` 파일을 만든다.
2. 실습가이드의 목차를 따라 직접 내용을 작성한다.
3. Deposit과 Withdrawal 상태 흐름을 직접 정리한다.
4. Wallet/Key Security 원칙을 직접 적어본다.
5. Rust signer가 왜 필요한지 한 문단으로 정리한다.
6. 검증문제를 풀고 답변가이드와 비교한다.

## 관련 GitHub 원본 문서

- [Phase 2 Domain Map](phase-2-domain-map.md)
- [Phase 2 Roadmap](../roadmap/phase-2-roadmap.md)
- [Target Architecture](../architecture/target-architecture.md)
- [Ledger and Settlement](ledger-and-settlement.md)

## 완료 기준

- [ ] Deposit과 Withdrawal의 차이를 설명할 수 있다.
- [ ] 출금 상태 machine 초안을 문서화한다.
- [ ] Wallet과 private key의 차이를 설명할 수 있다.
- [ ] 개인키를 DB에 평문 저장하면 안 되는 이유를 설명할 수 있다.
- [ ] Rust signer가 왜 필요한지 한 문단으로 정리한다.
