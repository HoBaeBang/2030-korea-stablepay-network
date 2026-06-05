# Wallet, Deposit, Withdrawal

이 문서는 Day 3 퇴근 후 실습에서 직접 채워 넣는 학습 산출물입니다.

관련 가이드: [Wallet, Deposit, Withdrawal 실습 가이드](Wallet_Deposit_Withdrawal_실습가이드.md)

## 한 문장 요약

Deposit은 외부 지갑에서 우리 서비스가 관리하는 입금 주소로 들어온 온체인 transaction을 감지하고, 검증한 뒤 내부 Ledger에 credit으로 반영하는 흐름입니다.

Withdrawal은 우리 서비스의 내부 잔액과 정책을 검증한 뒤, 외부 지갑으로 보낼 transaction을 만들고 private key로 서명해서 블록체인 네트워크에 전송하는 흐름입니다.

Wallet은 블록체인 주소와 키를 중심으로 자산 이동 권한을 표현하는 개념입니다. 다만 실제 잔액과 정산 상태는 지갑 자체만으로 판단하지 않고, 우리 서비스의 내부 Ledger와 온체인 상태를 함께 봐야 합니다.

Key Security는 private key와 transaction signing 권한을 안전하게 보호하고, 일반 API/DB 영역과 분리하는 설계입니다.

## Deposit과 Withdrawal의 차이

| 구분 | Deposit | Withdrawal |
| --- | --- | --- |
| 자산 방향 | 외부 지갑에서 우리 서비스가 관리하는 입금 주소로 들어오는 흐름 | 우리 서비스가 관리하는 지갑에서 외부 지갑으로 나가는 흐름 |
| 온체인 역할 | 이미 발생한 transaction/event를 감지하고 검증한다 | 새 transaction을 만들고 서명한 뒤 네트워크에 전송한다 |
| 주요 위험 | 잘못된 입금 인정, 지원하지 않는 토큰/체인 오인, 중복 반영 | 잘못된 주소 전송, private key 노출, 중복 출금, 승인되지 않은 출금 |
| Ledger 연결 | 입금이 확정되면 사용자 계정에 credit entry를 기록한다 | 출금 요청/승인/확정 단계에 따라 hold, debit, release 같은 entry를 기록한다 |
| 보안 포인트 | 입금 주소 검증, transaction hash 중복 방지, confirmation/finality 확인 | 주소 검증, 출금 승인 정책, signing 경계 분리, nonce/idempotency 관리 |

## Deposit 상태 흐름

Deposit은 우리 서버가 돈을 직접 받는 행위가 아니라, 블록체인에 이미 기록된 transaction/event를 감지해서 내부 장부에 반영하는 흐름입니다.

Event Indexer는 블록체인 안에서 실행되는 컴포넌트가 아니라, 우리 백엔드의 off-chain worker/indexer layer에서 실행되는 별도 프로세스입니다. 처음에는 일정 주기로 블록체인 RPC를 조회하는 polling 방식으로 구현하고, 이후 WebSocket subscription과 재처리/backfill을 섞은 hybrid 방식으로 확장할 수 있습니다.

![Deposit 실행 시퀀스](../../confluence/diagrams/spn20-deposit-sequence.png)

```text
DETECTED
-> CONFIRMING
-> CONFIRMED
-> CREDITED
```

| 상태 | 의미 |
| --- | --- |
| DETECTED | Event Indexer가 블록체인에서 우리 deposit address로 들어온 입금 후보를 발견한 상태 |
| CONFIRMING | transaction은 발견됐지만 confirmation/finality 기준을 기다리는 상태 |
| CONFIRMED | 주소, token, amount, tx hash, finality 검증을 통과해 입금으로 인정 가능한 상태 |
| CREDITED | 내부 Ledger에 사용자 잔액 증가 credit entry까지 반영된 상태 |

Deposit에서 조심해야 할 점은 "사용자가 보냈다고 주장하는 것"과 "우리 시스템이 입금으로 인정해도 되는 것"이 다르다는 점입니다. 그래서 transaction hash가 실제로 존재하는지, 받는 주소가 우리 deposit address인지, 지원하는 token인지, 이미 처리한 transaction은 아닌지 확인해야 합니다.

## Withdrawal 상태 흐름

Withdrawal은 내부 정책으로 출금 가능 여부를 판단한 뒤, 블록체인에 제출할 transaction을 만들고 서명해서 전송하는 흐름입니다.

![Withdrawal 실행 시퀀스](../../confluence/diagrams/spn20-withdrawal-sequence.png)

```text
REQUESTED
-> APPROVED
-> SIGNED
-> BROADCASTED
-> CONFIRMED
```

| 상태 | 의미 |
| --- | --- |
| REQUESTED | 사용자가 외부 주소로 출금을 요청한 상태 |
| APPROVED | 잔액, 한도, 주소, 위험 정책, 중복 요청 여부를 검증하고 출금 진행이 승인된 상태 |
| SIGNED | 블록체인에 제출할 transaction을 private key로 서명한 상태 |
| BROADCASTED | signed transaction을 블록체인 네트워크에 전송한 상태 |
| CONFIRMED | 온체인에서 transaction 성공과 confirmation/finality 기준을 만족한 상태 |
| FAILED | 서명, 전송, 온체인 실행, finality 확인 중 실패한 상태 |
| CANCELED | 승인 전 사용자가 취소했거나 내부 정책상 중단한 상태 |

여기서 SIGNED는 내부 Ledger에 서명했다는 뜻이 아닙니다. SIGNED는 블록체인 네트워크가 검증할 수 있도록 transaction에 private key로 암호학적 서명을 했다는 뜻입니다.

## Wallet과 Key Security

| 개념 | 의미 | 우리 프로젝트에서의 역할 |
| --- | --- | --- |
| Address | 외부에 공개 가능한 블록체인 주소 | 입금 수신 주소 또는 출금 목적지 주소로 사용한다 |
| Public Key | private key에서 파생되는 공개 키 | 서명을 검증할 때 사용되는 공개 정보의 기반이다 |
| Private Key | 해당 주소의 자산을 움직일 수 있는 비밀 키 | transaction signing 권한이므로 절대 평문 DB에 저장하면 안 된다 |
| Signing | private key로 transaction에 서명하는 행위 | 출금 transaction이 해당 지갑의 권한으로 만들어졌음을 증명한다 |

Address는 공개되어도 되지만 private key는 노출되면 안 됩니다. Private key가 노출되면 공격자가 우리 출금 지갑의 자산을 임의로 움직일 수 있기 때문입니다.

따라서 API 서버가 private key를 직접 읽거나 일반 DB 컬럼에 평문으로 저장하는 구조는 피해야 합니다. 출금 검증과 상태 관리는 Go 백엔드가 담당하되, private key와 signing은 더 좁은 signer boundary 안에 격리하는 방향이 좋습니다.

![Wallet과 Key Security 경계](../../confluence/diagrams/spn20-wallet-key-security-boundary.png)

## Rust Signer가 필요한 이유

Rust signer는 전체 백엔드를 Rust로 바꾸기 위한 것이 아니라, private key와 transaction signing이라는 위험한 책임을 작은 경계로 분리하기 위한 컴포넌트입니다.

Go 백엔드는 출금 요청 검증, 내부 정책 판단, Ledger 상태 저장, API 응답을 담당합니다. Rust signer는 승인된 출금 요청에 대해서만 unsigned transaction을 받아 private key로 서명하고 signed transaction을 반환합니다. 이렇게 분리하면 private key가 노출될 수 있는 범위를 줄이고, signing 권한을 더 엄격하게 통제할 수 있습니다.

## 최소 테이블 후보

| 테이블 후보 | 필요한 이유 | 저장할 주요 값 |
| --- | --- | --- |
| wallets | 우리 서비스가 관리하는 지갑 또는 사용자/가맹점과 연결된 지갑 정보를 관리하기 위해 필요 | id, owner_type, owner_id, chain, address, status, created_at |
| deposit_addresses | 사용자나 가맹점에게 발급한 입금 주소를 추적하기 위해 필요 | id, wallet_id, owner_id, chain, address, token_symbol, status, created_at |
| deposits | 온체인에서 감지한 입금 transaction과 처리 상태를 기록하기 위해 필요 | id, deposit_address_id, tx_hash, chain, token_symbol, amount, status, detected_at, confirmed_at |
| withdrawals | 사용자의 출금 요청과 온체인 전송 상태를 추적하기 위해 필요 | id, wallet_id, to_address, chain, token_symbol, amount, status, tx_hash, requested_at, confirmed_at |
| withdrawal_approvals | 출금 승인 정책, 승인자, 승인 시점, 승인 사유를 감사 가능하게 남기기 위해 필요 | id, withdrawal_id, approver_id, decision, reason, approved_at |

추가로 실제 구현 단계에서는 `ledger_accounts`, `ledger_transactions`, `ledger_entries`와 연결해야 합니다. Deposit은 최종적으로 credit entry를 만들고, Withdrawal은 승인/확정 단계에 따라 hold 또는 debit entry를 만들 수 있습니다.

## 아직 모르는 것과 다음 질문

- Deposit address는 사용자별로 하나씩 발급할지, merchant별로 발급할지, invoice별로 발급할지 결정해야 합니다.
- Ethereum, Cosmos, Solana처럼 체인별 address 검증 규칙과 finality 기준이 다르므로 어떤 체인을 먼저 지원할지 정해야 합니다.
- Withdrawal 승인 정책에서 최소/최대 출금 금액, 일일 한도, 고액 출금 추가 승인 기준을 정해야 합니다.
- Rust signer와 Go API 서버가 어떤 프로토콜로 통신할지 정해야 합니다. 예: HTTP, gRPC, message queue.
- private key를 signer 내부에서 어떻게 보관할지 정해야 합니다. 예: KMS, HSM, encrypted key store, local dev용 mock key.
- 출금 실패 시 재시도할지, 수동 검토로 넘길지, Ledger hold를 언제 해제할지 정책이 필요합니다.

## 검증 체크리스트

- [x] Deposit과 Withdrawal의 차이를 설명할 수 있다.
- [x] Deposit 상태 흐름을 설명할 수 있다.
- [x] Withdrawal 상태 흐름을 설명할 수 있다.
- [x] Wallet address와 private key의 차이를 설명할 수 있다.
- [x] Rust signer가 왜 필요한지 한 문단으로 설명할 수 있다.
