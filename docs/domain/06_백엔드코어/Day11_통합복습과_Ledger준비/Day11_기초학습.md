# Day 11 기초학습 - Backend Core 통합 복습과 Ledger 구현 준비

관련 Jira: [SPN-28](https://aslan0.atlassian.net/browse/SPN-28)

Day 11은 Backend Core의 마지막 점검일입니다.

Day 7부터 Day 10까지 정리한 공통 기반을 다시 확인하고, 다음 에픽인 Ledger 구현으로 넘어갈 준비가 되었는지 판단합니다.

## 오늘의 큰 그림

![Day11 Backend Core에서 Ledger로 넘어가는 준비도](../../../confluence/diagrams/spn28-day11-ledger-readiness.png)

Day11은 새로운 기능을 배우기보다, 지금까지 정리한 Backend Core가 실제로 Ledger 구현을 시작할 만큼 준비되었는지 판단하는 날입니다.

Ledger는 돈의 이동 기록입니다. 따라서 Ledger 구현이 시작되면 단순 CRUD보다 더 강한 기준이 필요합니다.

| 준비 영역 | 준비되지 않았을 때 생기는 문제 |
| --- | --- |
| Error Response | 잘못된 원장 요청이 어떤 이유로 실패했는지 API 사용자가 알기 어렵다 |
| Validation | 잘못된 금액, 잘못된 상태 전이, 중복 요청이 원장 기록으로 이어질 수 있다 |
| Config | DB, RPC, signer 같은 실행 환경 누락을 늦게 발견한다 |
| Logging | 돈의 이동이 언제 왜 발생했는지 추적하기 어렵다 |
| Test Pattern | 중복 원장, 불균형 debit/credit 같은 치명적인 버그를 반복 검증하기 어렵다 |

## 오늘의 목표

1. Backend Core가 왜 먼저 필요했는지 설명한다.
2. Day 8~10에서 정리한 내용을 하나의 체크리스트로 묶는다.
3. Ledger 구현 전 필요한 코드/문서 준비 상태를 확인한다.
4. 다음 에픽의 첫 작업 후보를 정리한다.
5. 모르는 개념을 질문 목록으로 남긴다.

## 완료 기준

- [ ] Backend Core 통합 체크리스트를 작성했다.
- [ ] Ledger 구현 전 위험 요소를 정리했다.
- [ ] 다음 작업 후보를 작성했다.
- [ ] 검증문제를 풀었다.
- [ ] SPN-2 에픽 완료 여부를 판단했다.

## 오늘 꼭 잡아야 하는 문장

```text
Ledger는 새로운 도메인이지만,
안전한 Ledger 구현은 Day8~10에서 만든 공통 기반 위에서 시작된다.
```
