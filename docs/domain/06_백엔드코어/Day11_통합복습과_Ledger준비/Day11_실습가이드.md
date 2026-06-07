# Day 11 실습가이드 - Backend Core 통합 복습과 Ledger 구현 준비

관련 Jira: [SPN-28](https://aslan0.atlassian.net/browse/SPN-28)

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

## Step 5. SPN-2 완료 판단

다음 기준으로 판단합니다.

```text
Backend Core를 문서로 설명할 수 있는가?
공통 기반 구현 후보가 구체적인가?
Ledger 구현을 시작할 때 필요한 준비가 되어 있는가?
```

## 완료 기준

- [ ] 통합 체크리스트를 작성했다.
- [ ] 약한 개념을 정리했다.
- [ ] Ledger 구현 전 위험 요소를 작성했다.
- [ ] 다음 구현 티켓 후보를 작성했다.
- [ ] SPN-2 완료 가능 여부를 판단했다.
