# Day 9 개념학습 - 설정 로딩과 애플리케이션 시작 구조

관련 Jira: [SPN-26](https://aslan0.atlassian.net/browse/SPN-26)

## config란?

config는 애플리케이션 실행에 필요한 설정값입니다.

예시:

```text
서버가 몇 번 포트에서 뜰 것인가?
어떤 DB에 연결할 것인가?
어떤 블록체인 RPC를 바라볼 것인가?
로그 레벨은 무엇인가?
```

## 환경 변수란?

환경 변수는 운영체제나 실행 환경에서 프로그램에 전달하는 값입니다.

Go 코드에서는 보통 `os.Getenv`로 읽습니다.

예시:

```go
port := os.Getenv("PORT")
```

## 왜 코드에 직접 쓰면 안 되는가?

예를 들어 DB 주소를 코드에 직접 쓰면 로컬, 테스트, 운영 환경을 바꿀 때마다 코드를 수정해야 합니다.

```text
로컬 DB
테스트 DB
운영 DB
```

이 값은 코드가 아니라 실행 환경에서 주입하는 것이 좋습니다.

## main.go의 역할

`main.go`는 애플리케이션을 조립하는 시작점입니다.

주요 책임:

1. 설정 읽기
2. DB 연결
3. repository 생성
4. service 생성
5. handler 등록
6. HTTP server 시작

## 애플리케이션 시작 구조

```text
main()
  |
  +--> load config
  |
  +--> open database
  |
  +--> create repositories
  |
  +--> create services
  |
  +--> register HTTP routes
  |
  +--> listen and serve
```

## Phase 2 설정 후보

| 설정 | 의미 |
| --- | --- |
| PORT | HTTP 서버 포트 |
| DATABASE_URL | PostgreSQL 연결 문자열 |
| LOG_LEVEL | 로그 출력 수준 |
| BLOCKCHAIN_RPC_URL | 블록체인 RPC 주소 |
| INDEXER_POLL_INTERVAL | 이벤트 인덱서 조회 간격 |
| FINALITY_CONFIRMATIONS | 입금 확정으로 볼 confirmation 수 |
| SIGNER_BASE_URL | Rust signer 서비스 주소 |

## 설정 누락 처리

필수 설정이 없으면 서버가 시작되지 않는 것이 좋습니다.

예를 들어 `DATABASE_URL`이 없는데 서버가 떠버리면, 실제 요청을 받을 때 뒤늦게 장애가 발생합니다.

처음 시작할 때 명확히 실패시키는 편이 안전합니다.
