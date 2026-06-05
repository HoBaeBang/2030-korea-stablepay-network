# Deposit, Withdrawal, Wallet, Key Security 검증문제와 답변가이드

관련 Jira: [SPN-20](https://aslan0.atlassian.net/browse/SPN-20)

이 문서는 Day 3 학습 후 스스로 확인할 검증문제와 답변가이드입니다.

먼저 문제를 풀고, 그 다음 답변가이드를 확인하세요.

## 검증문제

### 문제 1

Deposit과 Withdrawal의 차이를 한 문장씩 설명하세요.

### 문제 2

입금은 사용자가 돈을 보냈다고 말하는 순간 바로 내부 잔액에 반영해도 될까요? 왜 그런가요?

### 문제 3

Withdrawal을 `REQUESTED -> APPROVED -> SIGNED -> BROADCASTED -> CONFIRMED`로 나누는 이유를 설명하세요.

### 문제 4

Wallet address와 private key의 차이를 설명하세요.

### 문제 5

private key를 PostgreSQL에 평문으로 저장하면 어떤 문제가 생길 수 있나요?

### 문제 6

Rust signer는 왜 전체 백엔드가 아니라 작은 별도 컴포넌트로 두는 것이 좋을까요?

### 문제 7

출금 요청이 같은 이유로 두 번 처리되면 어떤 문제가 생기고, 어떻게 막아야 할까요?

### 문제 8

Deposit과 Withdrawal은 Ledger와 각각 어떻게 연결되나요?

## 답변가이드

### 답변 1

Deposit은 외부 지갑에서 우리 시스템으로 자산이 들어오는 흐름입니다.

Withdrawal은 우리 시스템에서 외부 지갑으로 자산이 나가는 흐름입니다.

핵심 차이:

```text
Deposit = 들어오는 자산 감지
Withdrawal = 나가는 자산 검증, 승인, 서명, 전송
```

### 답변 2

바로 반영하면 위험합니다.

transaction hash가 실제로 존재하는지, 받는 주소가 우리 시스템 주소인지, 토큰과 금액이 맞는지, finality가 충분한지, 이미 처리한 transaction은 아닌지 확인해야 합니다.

따라서 입금은 보통 감지, 확인 중, 확정, 내부 원장 반영 같은 단계로 나누어 처리합니다.

### 답변 3

출금은 되돌리기 어려운 온체인 전송이기 때문에 단계를 나누어야 합니다.

각 단계의 책임은 다음과 같습니다.

| 상태 | 의미 |
| --- | --- |
| REQUESTED | 사용자가 출금을 요청함 |
| APPROVED | 정책상 출금 가능하다고 승인됨 |
| SIGNED | transaction에 서명함 |
| BROADCASTED | 네트워크에 전송함 |
| CONFIRMED | 온체인에서 확정됨 |

### 답변 4

Wallet address는 외부에 공개 가능한 주소입니다.

Private key는 해당 주소의 자산을 움직일 수 있는 비밀 값입니다.

주소는 계좌번호에 가깝고, private key는 그 계좌의 돈을 실제로 움직일 수 있는 권한에 가깝습니다.

### 답변 5

private key를 평문으로 저장하면 DB가 유출되었을 때 자산도 함께 탈취될 수 있습니다.

또한 애플리케이션 로그, 백업, 운영자 접근, SQL 조회 등 여러 경로로 키가 노출될 수 있습니다.

따라서 개인키는 암호화, HSM/KMS, signer service 분리 같은 별도 보안 경계 안에서 다뤄야 합니다.

### 답변 6

Rust signer는 개인키와 transaction 서명이라는 위험한 책임을 작은 컴포넌트로 격리하기 위해 둡니다.

Go API 서버가 모든 것을 직접 처리하면 API, DB, 비즈니스 로직, 개인키가 같은 경계에 섞일 수 있습니다.

Signer를 분리하면 개인키 접근 범위를 줄이고, 서명 요청 검증과 감사 로그를 더 명확히 만들 수 있습니다.

### 답변 7

같은 출금 요청이 두 번 처리되면 사용자의 돈이 두 번 빠져나갈 수 있습니다.

이를 막기 위해 idempotency key, withdrawal status machine, unique constraint, 이미 처리된 transaction hash 확인, ledger entry 중복 방지 같은 장치가 필요합니다.

### 답변 8

Deposit은 확정된 입금을 내부 사용자 계정의 credit entry로 연결합니다.

Withdrawal은 출금 요청 또는 확정 시점에 사용자 계정의 debit entry로 연결합니다.

핵심은 입출금도 결국 돈의 이동이므로 Ledger에 기록되어야 한다는 점입니다.

## 통과 기준

- [ ] Deposit과 Withdrawal을 각각 한 문장으로 설명할 수 있다.
- [ ] 출금 상태 흐름을 설명할 수 있다.
- [ ] address와 private key의 차이를 설명할 수 있다.
- [ ] private key를 평문 저장하면 안 되는 이유를 설명할 수 있다.
- [ ] Rust signer가 필요한 이유를 한 문단으로 설명할 수 있다.
