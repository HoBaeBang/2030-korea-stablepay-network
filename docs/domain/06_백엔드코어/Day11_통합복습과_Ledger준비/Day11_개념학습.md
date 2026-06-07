# Day 11 개념학습 - Backend Core 통합 복습과 Ledger 구현 준비

관련 Jira: [SPN-28](https://aslan0.atlassian.net/browse/SPN-28)

## Backend Core에서 정리한 것

| 영역 | 역할 |
| --- | --- |
| Error Response | API 실패 응답을 일관되게 만든다. |
| Validation | 잘못된 요청과 도메인 규칙 위반을 막는다. |
| Config | 실행 환경 설정을 한 곳에서 관리한다. |
| Logging | 운영 중 발생한 일을 추적한다. |
| Test Pattern | 반복 가능한 검증 방식을 만든다. |

## 왜 Ledger 전에 필요한가?

Ledger는 돈의 이동을 기록합니다.

따라서 Ledger 구현 전에 다음 질문에 답할 수 있어야 합니다.

```text
잘못된 요청은 어디에서 막을 것인가?
실패 응답은 어떤 모양으로 줄 것인가?
설정 누락은 언제 발견할 것인가?
중요한 상태 변경은 어떤 로그로 남길 것인가?
돈의 이동은 어떤 테스트로 검증할 것인가?
```

## Ledger 구현으로 넘어가기 전 체크포인트

```text
Error Response 준비
        |
Validation 준비
        |
Config 준비
        |
Logging 준비
        |
Test Pattern 준비
        |
Ledger 구현 시작
```

## Ledger 첫 구현 후보

Backend Core 이후에는 다음 순서가 적절합니다.

1. Ledger Account 모델 설계
2. Ledger Transaction 모델 설계
3. Ledger Entry 모델 설계
4. migration 작성
5. repository 작성
6. service 작성
7. payment finalized 시 ledger transaction 생성

## 오늘의 핵심 판단

Day 11에서는 “모든 것을 완벽히 아는가?”를 확인하는 것이 아닙니다.

다음 구현을 시작했을 때 막힐 위험이 큰 부분을 미리 발견하는 것이 목적입니다.
