# Day 9 실습가이드 - 설정 로딩과 애플리케이션 시작 구조

관련 Jira: [SPN-26](https://aslan0.atlassian.net/browse/SPN-26)

## 실습 목표

`Day9_실습산출물.md`에 다음 내용을 작성합니다.

1. 현재 `main.go` 실행 순서
2. 현재 필요한 설정값
3. Phase 2에서 추가될 설정값
4. 설정 누락 시 처리 기준
5. config 패키지 구현 후보

## Step 1. main.go 읽기

확인 파일:

```text
cmd/api/main.go
```

확인할 질문:

```text
서버 포트는 어디에서 정해지는가?
DB 연결은 어디에서 만들어지는가?
repository와 service는 어떤 순서로 만들어지는가?
route 등록은 어디에서 일어나는가?
```

## Step 2. 현재 설정 후보 정리

현재 프로젝트 기준으로 반드시 필요한 설정을 적습니다.

```text
PORT
DATABASE_URL
```

## Step 3. Phase 2 설정 후보 정리

다음 기능이 들어올 때 필요한 설정을 상상해봅니다.

| 기능 | 필요한 설정 |
| --- | --- |
| Event Indexer |  |
| Withdrawal Signer |  |
| Blockchain RPC |  |
| Logging |  |

## Step 4. config 구현 후보 작성

예시 후보:

```text
internal/platform/config/config.go
```

구조체 예시:

```go
type Config struct {
    Port string
    DatabaseURL string
    LogLevel string
}
```

오늘은 정확한 코드보다, 어떤 구조가 좋을지 정리하는 것이 우선입니다.

## Step 5. 설정 누락 기준 작성

다음 질문에 답합니다.

```text
PORT가 없으면 기본값을 써도 되는가?
DATABASE_URL이 없으면 서버를 시작해도 되는가?
BLOCKCHAIN_RPC_URL이 없으면 모든 기능이 막히는가, 일부 기능만 막히는가?
```

## 완료 기준

- [ ] main.go 실행 순서를 작성했다.
- [ ] 현재 설정 후보를 작성했다.
- [ ] Phase 2 설정 후보를 작성했다.
- [ ] config 구현 후보를 작성했다.
- [ ] 설정 누락 처리 기준을 작성했다.
