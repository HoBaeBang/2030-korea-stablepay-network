# Day 11 실습가이드 - Backend Core 통합 복습과 Ledger 구현 준비

관련 Jira: [SPN-28](https://aslan0.atlassian.net/browse/SPN-28)

## 실습 흐름

![Day11 Backend Core에서 Ledger로 넘어가는 준비도](../../../confluence/diagrams/spn28-day11-ledger-readiness.png)

오늘 실습은 “복습했다”로 끝나면 안 됩니다. Day8~10에서 정리한 공통 기반이 Ledger 구현의 어떤 위험을 줄여주는지 연결해서 설명해야 합니다.

## 실습 목표

`Day11_실습산출물.md`에 다음 내용을 작성합니다.

1. Backend Core 통합 체크리스트
2. Day 8~10에서 가장 약한 개념
3. Ledger 구현 전 위험 요소
4. 다음 구현 티켓 후보
5. SPN-2 에픽 완료 판단

## Step 1. Day 8~10 산출물 다시 보기

확인할 문서:

```text
Day8_실습산출물.md
Day9_실습산출물.md
Day10_실습산출물.md
```

## Step 2. Backend Core 체크리스트 작성

아래 질문에 답합니다.

```text
공통 에러 응답 형식은 정해졌는가?
validation 위치는 정해졌는가?
config 후보는 정해졌는가?
로그 후보는 정해졌는가?
테스트 패턴은 정해졌는가?
```

체크리스트를 작성할 때는 단순히 “했다/안 했다”가 아니라, Ledger 구현과 어떻게 연결되는지를 같이 씁니다.

| 항목 | Ledger와 연결되는 이유 |
| --- | --- |
| 공통 에러 응답 | Ledger 실패를 API에서 일관되게 표현해야 한다 |
| validation | 잘못된 금액과 중복 요청을 원장에 반영하지 않아야 한다 |
| config | DB와 향후 RPC/signer 설정 누락을 시작 시점에 잡아야 한다 |
| logging | 돈의 이동 기록 생성 시점을 추적해야 한다 |
| test pattern | 복식부기 불변식과 중복 방어를 자동 검증해야 한다 |

## Step 3. Ledger 구현 전 위험 요소 작성

예시:

```text
ledger entry의 debit/credit 방향을 헷갈릴 수 있다.
payment finalized를 중복 처리하면 ledger가 중복 생성될 수 있다.
DB transaction 없이 ledger transaction과 entry를 따로 저장하면 정합성이 깨질 수 있다.
```

## Step 4. 다음 구현 티켓 후보 작성

예시:

```text
Ledger migration 작성
Ledger repository 작성
Ledger service 작성
Payment finalized와 Ledger 연결
Ledger 테스트 작성
```

티켓 후보는 너무 추상적으로 쓰지 않습니다.

좋은 후보:

```text
Ledger account/transaction/entry migration 작성
Ledger transaction 생성 service와 debit/credit 합계 검증 구현
Payment finalized 시 ledger transaction 생성 연결
동일 payment에 대한 ledger 중복 생성 방지 테스트 작성
```

아직 피해야 하는 후보:

```text
Ledger 전체 구현
정산 전체 구현
블록체인 연동 전체 구현
```

너무 큰 티켓은 하루나 이틀 단위로 검증하기 어렵습니다.

## Step 5. SPN-2 완료 판단

다음 기준으로 판단합니다.

```text
Backend Core를 문서로 설명할 수 있는가?
공통 기반 구현 후보가 구체적인가?
Ledger 구현을 시작할 때 필요한 준비가 되어 있는가?
```

판단 결과는 다음 세 가지 중 하나로 적습니다.

| 판단 | 의미 |
| --- | --- |
| 완료 가능 | Ledger 첫 구현으로 넘어갈 수 있다 |
| 일부 보강 후 완료 | 특정 문서나 구현 후보만 보강하면 된다 |
| 완료 보류 | 공통 기반이 아직 너무 약해 Ledger로 넘어가면 위험하다 |

## 완료 기준

- [ ] 통합 체크리스트를 작성했다.
- [ ] 약한 개념을 정리했다.
- [ ] Ledger 구현 전 위험 요소를 작성했다.
- [ ] 다음 구현 티켓 후보를 작성했다.
- [ ] SPN-2 완료 가능 여부를 판단했다.
