# Day 10 기초학습 - 로깅과 테스트 패턴

관련 Jira: [SPN-27](https://aslan0.atlassian.net/browse/SPN-27)

Day 10의 목표는 로그와 테스트 패턴을 정리하는 것입니다.

Phase 2에서는 돈의 이동, 상태 변경, 온체인 이벤트 처리 같은 작업이 많아집니다.

이때 로그와 테스트가 없으면 문제가 생겼을 때 어디서 왜 틀어졌는지 알기 어렵습니다.

## 오늘의 큰 그림

```text
Request / Job
      |
      v
Domain Action
      |
      +--> Log important event
      |
      v
State Change
      |
      +--> Test expected behavior
```

## 오늘의 목표

1. 어떤 상황에서 로그가 필요한지 설명할 수 있다.
2. 로그와 에러 응답의 차이를 이해한다.
3. Go 테스트의 기본 구조를 다시 복습한다.
4. 한글 subtest로 테스트 케이스를 읽기 좋게 작성하는 방식을 정리한다.
5. Backend Core 이후 Ledger 테스트 패턴을 준비한다.

## 오늘 꼭 잡아야 하는 문장

```text
로그는 운영 중인 시스템의 기억이고,
테스트는 코드가 지켜야 할 약속이다.
```

## 완료 기준

- [ ] 로그가 필요한 이벤트 후보를 작성했다.
- [ ] 테스트 패턴을 given/when/then으로 정리했다.
- [ ] 한글 subtest 예시를 작성했다.
- [ ] 실습산출물을 작성했다.
- [ ] 검증문제를 풀었다.
