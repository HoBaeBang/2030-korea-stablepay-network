# Day 9 기초학습 - 설정 로딩과 애플리케이션 시작 구조

관련 Jira: [SPN-26](https://aslan0.atlassian.net/browse/SPN-26)

Day 9의 목표는 서버 실행에 필요한 설정을 한 곳에서 읽고, 애플리케이션 시작 구조를 이해하는 것입니다.

## 오늘의 큰 그림

```text
Environment Variables
        |
        v
Config
        |
        v
main.go
        |
        +--> DB 연결
        +--> Service 생성
        +--> Handler 등록
        +--> HTTP Server 시작
```

## 왜 설정 구조가 필요한가?

처음에는 `PORT`, `DATABASE_URL` 정도만 필요합니다.

하지만 Phase 2가 진행되면 다음 설정이 추가됩니다.

```text
BLOCKCHAIN_RPC_URL
INDEXER_POLL_INTERVAL
FINALITY_CONFIRMATIONS
SIGNER_BASE_URL
LOG_LEVEL
```

이 값들이 코드 여러 곳에 흩어지면 변경하기 어렵고, 테스트도 어려워집니다.

## 오늘의 목표

1. config가 무엇인지 설명할 수 있다.
2. 환경 변수와 코드 설정의 차이를 이해한다.
3. `main.go`가 어떤 순서로 서버를 시작하는지 설명할 수 있다.
4. Phase 2에 필요한 설정 후보를 정리한다.
5. 설정 누락 시 어떤 에러를 내야 하는지 판단한다.

## 오늘 꼭 잡아야 하는 문장

```text
Config는 실행 환경과 애플리케이션 코드를 연결하는 경계다.
```

## 완료 기준

- [ ] 설정 후보를 작성했다.
- [ ] main.go의 실행 순서를 설명했다.
- [ ] 환경 변수 누락 시 처리 방식을 정리했다.
- [ ] 실습산출물을 작성했다.
- [ ] 검증문제를 풀었다.
