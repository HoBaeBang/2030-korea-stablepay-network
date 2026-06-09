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

## 설정 로딩 흐름

![Day9 설정 로딩과 애플리케이션 시작 구조](../../../confluence/diagrams/spn26-day9-config-startup-flow.png)

설정은 단순히 `os.Getenv`를 여러 곳에서 호출하는 문제가 아닙니다.

중요한 것은 “언제 읽고, 어디에 모으고, 누가 사용하게 할 것인가”입니다.

나쁜 예시는 handler, service, repository 곳곳에서 직접 환경 변수를 읽는 방식입니다.

```go
func (s *Service) CreateInvoice(...) {
    rpcURL := os.Getenv("BLOCKCHAIN_RPC_URL")
    // ...
}
```

이렇게 되면 테스트할 때도 환경 변수를 준비해야 하고, 어떤 설정이 필요한지 한눈에 알기 어렵습니다.

좋은 방향은 시작 시점에 설정을 한 번 읽고 `Config` 구조체로 모은 뒤 필요한 계층에 명시적으로 전달하는 것입니다.

```go
type Config struct {
    Port        string
    DatabaseURL string
    LogLevel    string
}
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

`main.go`는 비즈니스 규칙을 많이 담는 파일이 아닙니다.

`main.go`에 도메인 규칙이 많아지면 테스트하기 어렵고, 애플리케이션 시작 코드와 비즈니스 코드가 섞입니다.

좋은 시작 구조는 다음과 같습니다.

| 단계 | main.go가 하는 일 | 다른 패키지로 뺄 수 있는 일 |
| --- | --- | --- |
| Config 로딩 | `config.Load()` 호출 | 환경 변수 파싱, 기본값 적용, 필수값 검증 |
| DB 연결 | `database.Open(ctx, cfg.DatabaseURL)` 호출 | ping, pool 설정 |
| Router 생성 | `http.NewServeMux()` 호출 | route 등록 함수 |
| 의존성 조립 | repository/service/handler 생성 | app factory로 분리 가능 |
| Server 시작 | address 결정 후 listen | graceful shutdown은 추후 추가 |

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

## 필수 설정과 선택 설정

모든 설정을 같은 강도로 다룰 필요는 없습니다.

| 설정 | 권장 처리 |
| --- | --- |
| `PORT` | 없으면 `8080` 기본값 사용 가능 |
| `DATABASE_URL` | 없으면 서버 시작 실패 |
| `LOG_LEVEL` | 없으면 `info` 기본값 사용 가능 |
| `BLOCKCHAIN_RPC_URL` | Phase 2 온체인 기능이 켜져 있다면 필수 |
| `SIGNER_BASE_URL` | Withdrawal 기능이 켜져 있다면 필수 |
| `FINALITY_CONFIRMATIONS` | 없으면 체인별 안전 기본값을 정해야 함 |

이 판단이 중요한 이유는 설정 누락이 단순 개발 편의 문제가 아니라 결제 안정성과 연결되기 때문입니다.
