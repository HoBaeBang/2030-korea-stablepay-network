# Phase 2 Review Checkpoint

관련 Jira: [SPN-23](https://aslan0.atlassian.net/browse/SPN-23)

이 문서는 Day 6 퇴근 후 직접 작성하는 통합 복습 산출물입니다.

## 한 문장 요약

> 여기에 Day 1~5를 통합해서 이해한 내용을 한 문장으로 작성한다.

## Day 1~5 핵심 요약

| Day | 주제 | 내가 이해한 한 문장 |
| --- | --- | --- |
| Day 1 | Phase 2 Domain Map |  |
| Day 2 | Ledger & Settlement |  |
| Day 3 | Deposit / Withdrawal / Wallet / Key Security |  |
| Day 4 | Blockchain Event Indexer |  |
| Day 5 | First Implementation Scope |  |

## 도메인별 이해 상태

| 도메인 | 이해도(상/중/하) | 이유 |
| --- | --- | --- |
| Payment |  |  |
| Ledger |  |  |
| Settlement |  |  |
| Deposit |  |  |
| Withdrawal |  |  |
| Wallet |  |  |
| Key Security |  |  |
| Event Indexer |  |  |
| Idempotency |  |  |
| Reconciliation |  |  |
| Finality |  |  |

## 아직 약한 개념과 질문

### 약한 개념 1

- 현재 이해:
- 헷갈리는 지점:
- 구현 전에 확인할 질문:

### 약한 개념 2

- 현재 이해:
- 헷갈리는 지점:
- 구현 전에 확인할 질문:

### 약한 개념 3

- 현재 이해:
- 헷갈리는 지점:
- 구현 전에 확인할 질문:

## Sprint 2 구현 전 체크리스트

| 체크 항목 | 결과 | 메모 |
| --- | --- | --- |
| Payment와 Ledger의 책임 차이를 설명할 수 있다 |  |  |
| Ledger entry에 `+`, `-` 금액이 왜 필요한지 설명할 수 있다 |  |  |
| Settlement가 단순 합계가 아닌 이유를 설명할 수 있다 |  |  |
| Deposit과 Withdrawal의 위험 차이를 설명할 수 있다 |  |  |
| Event Indexer가 off-chain worker라는 점을 설명할 수 있다 |  |  |
| Idempotency key 후보를 말할 수 있다 |  |  |
| Reconciliation이 무엇을 비교하는지 설명할 수 있다 |  |  |
| Backend Core vertical slice를 먼저 구현해야 하는 이유를 설명할 수 있다 |  |  |

## Sprint 2 진입 판단

선택:

- [ ] 구현 시작 가능
- [ ] 부분 보강 후 시작
- [ ] 하루 더 복습 필요

판단 이유:

```text
여기에 판단 이유를 작성한다.
```

## 다음 작업 후보

| 후보 작업 | 해야 하는 이유 | 예상 난이도 | 먼저 확인할 질문 |
| --- | --- | --- | --- |
| Backend Core migration 설계 |  |  |  |
| Ledger account/entry 모델 설계 |  |  |  |
| Payment finalized 이후 ledger 연결 |  |  |  |
| Settlement skeleton 작성 |  |  |  |

## 오늘의 회고

### 오늘 가장 잘 이해된 개념

### 아직 가장 약한 개념

### 다음 구현에서 조심할 점
