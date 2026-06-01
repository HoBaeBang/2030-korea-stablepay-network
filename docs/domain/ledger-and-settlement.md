# Ledger and Settlement

이 문서는 Day 2 퇴근 후 실습에서 직접 채워 넣는 학습 산출물입니다.

관련 가이드: [Ledger와 Settlement 실습 가이드](ledger-and-settlement-practice-guide.md)

## 한 문장 요약

TODO: Ledger와 Settlement를 본인 말로 한 문장으로 정리합니다.

## 왜 Ledger가 필요한가

TODO: Payment 상태만으로 부족한 점과 Ledger가 필요한 이유를 정리합니다.

## Payment, Ledger, Settlement 책임 비교

| 도메인 | 책임 | 저장해야 할 정보 | 예시 상태 |
| --- | --- | --- | --- |
| Payment | TODO | TODO | TODO |
| Ledger | TODO | TODO | TODO |
| Settlement | TODO | TODO | TODO |

## Payment FINALIZED와 Settlement PAID의 차이

TODO: 결제 확정과 정산 완료가 왜 다른 단계인지 설명합니다.

## Double-entry 예시

TODO: 고객이 10 USDC를 결제했을 때 원장 entry 2개를 작성합니다.

```text
Ledger Transaction:

Entry 1:
Account:
Amount:
Reason:

Entry 2:
Account:
Amount:
Reason:

Check:
```

## 최소 테이블 후보

| 테이블 후보 | 필요한 이유 | 저장할 주요 값 |
| --- | --- | --- |
| ledger_accounts | TODO | TODO |
| ledger_transactions | TODO | TODO |
| ledger_entries | TODO | TODO |
| settlements | TODO | TODO |
| settlement_items | TODO | TODO |

## 아직 모르는 것과 다음 질문

- TODO
- TODO

## 검증 체크리스트

- [ ] Payment와 Ledger의 차이를 설명할 수 있다.
- [ ] Ledger와 Settlement의 차이를 설명할 수 있다.
- [ ] Payment `FINALIZED`와 Settlement `PAID`의 차이를 설명할 수 있다.
- [ ] 10 USDC 결제 예시에서 ledger entry 합계가 0이 되게 작성할 수 있다.
- [ ] 최소 테이블 후보 5개를 말할 수 있다.
