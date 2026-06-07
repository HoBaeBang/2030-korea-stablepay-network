# Phase 2 Day 1~5 통합 복습과 구현 전 점검

관련 Jira: [SPN-23](https://aslan0.atlassian.net/browse/SPN-23)

이 문서는 Phase 2 Day 6 학습 허브입니다.

Day 6의 목표는 새로운 기능을 더 배우는 것이 아니라, Day 1~5에서 배운 내용을 한 번에 다시 연결하고 Sprint 2 Backend Core 구현을 시작할 준비가 되었는지 점검하는 것입니다.

## 오늘의 큰 그림

![SPN-23 Day 6 통합 복습 흐름](../../confluence/diagrams/spn23-day6-review-flow.png)

Day 6는 잠깐 멈춰서 보는 날입니다. 지금까지 많이 따라왔지만, 실제 구현에 들어가려면 다음 질문에 답할 수 있어야 합니다.

```text
Payment 상태만 바꾸던 Phase 1 백엔드가
왜 Ledger, Settlement, Indexer, Deposit, Withdrawal, Wallet, Key Security로 확장되어야 하는가?
```

## 오늘의 목표

1. Day 1~5의 핵심 내용을 본인 말로 다시 설명할 수 있다.
2. Payment, Ledger, Settlement, Indexer, Deposit, Withdrawal, Wallet의 역할을 구분할 수 있다.
3. On-chain과 Off-chain의 차이를 설명할 수 있다.
4. Sprint 2 첫 구현 범위가 왜 Backend Core vertical slice인지 이해한다.
5. 아직 약한 개념을 숨기지 않고 질문 목록으로 정리한다.
6. Sprint 2 구현을 바로 시작할 수 있는지 판단한다.

## 오늘의 권장 진행 순서

Day 6는 새 개념을 많이 추가하는 날이 아니라, Day 1~5에서 배운 내용을 한 번에 연결하고 스스로 이해도를 점검하는 날입니다.

아래 순서대로 보면 됩니다.

| 순서 | 문서 | 언제 보는가 | 목적 |
| --- | --- | --- |
| 1 | [Phase 2 Day 1~5 종합 교재](Phase_2_Day1-5_종합_교재.md) | 출퇴근 또는 시작 전 | Day 1~5 전체 흐름을 한 번에 다시 잡는다. |
| 2 | [Phase 2 통합 복습 기초 학습](Phase_2_통합복습_기초학습.md) | 현재 문서 | Day 6의 목표, 진행 순서, 완료 기준을 확인한다. |
| 3 | [Phase 2 통합 복습 개념 학습](Phase_2_통합복습_개념학습.md) | 출퇴근 학습 | Payment, Ledger, Settlement, Deposit, Withdrawal, Indexer를 다시 연결한다. |
| 4 | [Phase 2 통합 복습 실습 가이드](Phase_2_통합복습_실습가이드.md) | 퇴근 후 작업 전 | 실습산출물에 무엇을 작성해야 하는지 확인한다. |
| 5 | [Phase 2 Review Checkpoint 실습산출물](Phase_2_Review_Checkpoint_실습산출물.md) | 퇴근 후 작업 | 본인이 이해한 내용을 직접 작성한다. |
| 6 | [Phase 2 통합 복습 검증문제와 답변가이드](Phase_2_통합복습_검증문제_답변가이드.md) | 산출물 작성 후 | 문제를 먼저 풀고, 그 다음 답변가이드와 비교한다. |

## 검증문제 활용법

검증문제는 답을 외우는 용도가 아니라, 지금 어떤 개념이 아직 약한지 찾기 위한 도구입니다.

1. 답변가이드를 먼저 보지 않는다.
2. 검증문제 12개를 본인 말로 먼저 작성한다.
3. 답변가이드와 비교한다.
4. 틀린 문제보다 애매하게 맞힌 문제를 더 중요하게 표시한다.
5. 약한 개념을 `Phase_2_Review_Checkpoint_실습산출물.md`의 질문 목록에 옮긴다.

예를 들어 `Ledger와 Payment 차이`를 설명할 수는 있지만 `Ledger Entry가 왜 필요한지`가 애매하다면, 그것은 구현 전에 반드시 다시 확인해야 할 약한 개념입니다.

## 오늘 꼭 잡아야 하는 문장

```text
Day 6의 목적은 많이 아는 척하는 것이 아니라,
구현 전에 무엇을 알고 있고 무엇을 아직 모르는지 분명하게 나누는 것이다.
```

## 퇴근 후 작업의 원칙

퇴근 후 작업은 사용자가 직접 진행합니다.

1. `docs/domain/00_종합/Phase_2_Review_Checkpoint_실습산출물.md` 파일을 채운다.
2. Day 1~5 핵심 개념을 한 문장씩 요약한다.
3. 아직 약한 개념을 따로 표시한다.
4. Sprint 2 구현 전 준비도 체크리스트를 작성한다.
5. 다음 구현 티켓으로 만들 수 있는 후보를 정리한다.

## 완료 기준

- [ ] Day 1~5 핵심 개념을 본인 말로 요약했다.
- [ ] 아직 약한 개념과 질문을 따로 정리했다.
- [ ] Backend Core 구현을 시작할 수 있는지 판단했다.
- [ ] Sprint 2 구현 전 체크리스트를 작성했다.
