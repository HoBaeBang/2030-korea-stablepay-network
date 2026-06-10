# Day 9 실습가이드 - 설정 로딩과 애플리케이션 시작 구조

관련 Jira: [SPN-26](https://aslan0.atlassian.net/browse/SPN-26)

## 실습 흐름

![Day9 설정 로딩과 애플리케이션 시작 구조](../../../confluence/diagrams/spn26-day9-config-startup-flow.png)

오늘 실습은 `main.go`를 “서버 실행 파일”로만 보는 것이 아니라, 설정과 의존성이 어떤 순서로 조립되는지 읽는 연습입니다.

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

읽을 때는 단순히 위에서 아래로 따라가지 말고, 아래 질문을 같이 표시합니다.

| 관찰 지점 | 메모할 내용 |
| --- | --- |
| 설정값 | 코드에 직접 박힌 값인가, 환경 변수에서 온 값인가 |
| DB 연결 | 실패하면 어디서 에러가 처리되는가 |
| route 등록 | domain별 등록 함수가 있는가 |
| 의존성 생성 | handler가 service를 알고, service가 repository를 아는 방향인가 |
| server 시작 | 주소와 포트가 어디서 결정되는가 |

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

추가로 다음 질문도 적어봅니다.

```text
config 패키지는 internal/platform/config가 좋을까?
아니면 internal/config가 좋을까?
database 연결 설정도 config에 둘까, database 패키지에 둘까?
```

## Step 5. 설정 누락 기준 작성

다음 질문에 답합니다.

```text
PORT가 없으면 기본값을 써도 되는가?
DATABASE_URL이 없으면 서버를 시작해도 되는가?
BLOCKCHAIN_RPC_URL이 없으면 모든 기능이 막히는가, 일부 기능만 막히는가?
```

Phase 2에서는 모든 기능이 한 번에 켜지지 않을 수 있습니다.

예를 들어 결제 백엔드 API만 실행할 때는 `BLOCKCHAIN_RPC_URL`이 없어도 될 수 있지만, Event Indexer를 실행하는 프로세스라면 필수입니다.

그래서 설정 누락 기준은 “전체 서버 기준”이 아니라 “그 프로세스가 맡은 역할 기준”으로 생각해야 합니다.

## Step 6. 코드 작업 - config 패키지 만들기

Day9의 실제 코드 작업은 `main.go`에 직접 들어있는 설정 로딩 코드를 별도 패키지로 옮기는 것입니다.

목표:

```text
cmd/api/main.go가 환경 변수 이름을 직접 많이 알지 않도록 한다.
설정 읽기, 기본값 적용, 필수값 검증을 config 패키지로 모은다.
```

### 6-1. 새 파일 만들기

생성할 파일:

```text
internal/platform/config/config.go
```

작성할 코드:

```go
package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable"
	}

	return Config{
		Port:        port,
		DatabaseURL: databaseURL,
	}
}
```

여기서 `Config`는 설정값을 담는 구조체입니다.

`Load()`는 환경 변수를 읽고, 비어 있으면 기본값을 채운 뒤 `Config` 값을 반환합니다.

### 6-2. main.go 수정하기

수정할 파일:

```text
cmd/api/main.go
```

기존에는 `main.go` 안에서 `os.Getenv`를 직접 호출합니다.

수정 후에는 아래처럼 `config.Load()`를 먼저 호출합니다.

```go
cfg := config.Load()
```

그리고 기존 코드에서 사용하던 값을 다음처럼 바꿉니다.

```go
db, err := database.Open(ctx, cfg.DatabaseURL)
```

```go
Addr: ":" + cfg.Port,
```

```go
log.Printf("stablepay api listening on :%s", cfg.Port)
```

import도 함께 정리합니다.

```go
import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/httpapi"
	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/platform/config"
	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/platform/database"
)
```

`os`는 더 이상 `main.go`에서 직접 쓰지 않으므로 제거합니다.

### 6-3. 포인터 관점으로 읽기

이번 코드에서 `Load()`는 `Config` 값을 반환합니다.

```go
func Load() Config
```

즉 `*Config` 포인터가 아니라 `Config` 값입니다.

현재 설정값은 작고 복사 비용이 크지 않으므로 값으로 반환해도 충분합니다. 나중에 설정이 커지거나 공유 변경이 필요해지면 `*Config`를 고려할 수 있습니다.

### 6-4. 검증 명령 실행

코드 작성 후 실행합니다.

```bash
gofmt -w internal/platform/config/config.go cmd/api/main.go
go test ./...
```

서버 실행까지 확인하고 싶다면 DB가 켜진 상태에서 실행합니다.

```bash
go run ./cmd/api
```

확인할 로그:

```text
database connection ok
stablepay api listening on :8080
```

## Step 7. 실습산출물 작성

코드 작업이 끝나면 `Day9_실습산출물.md`에 다음 내용을 작성합니다.

```text
main.go 실행 순서
config 패키지를 만든 이유
Config 구조체에 넣은 값
기본값을 둔 설정과 필수로 볼 설정
이번 코드 작업에서 헷갈린 점
```

## Step 8. 완성본 확인

코드 작성이 끝나면 아래 완성본과 본인이 작성한 파일을 비교합니다.

완전히 외워서 따라 쓰는 것이 목적은 아닙니다. 다만 다음을 확인해야 합니다.

```text
config.go가 설정 읽기 책임을 가진다.
main.go가 os.Getenv를 직접 호출하지 않는다.
main.go는 config.Load() 결과를 사용한다.
go fmt ./...와 go test ./...가 성공한다.
```

### internal/platform/config/config.go 완성본

```go
package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable"
	}

	return Config{
		Port:        port,
		DatabaseURL: databaseURL,
	}
}
```

### cmd/api/main.go 완성본

```go
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/httpapi"
	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/platform/config"
	"github.com/HoBaeBang/2030-korea-stablepay-network/internal/platform/database"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	log.Println("database connection ok")

	// mux는 multiplexer의 줄임말이다. 여러 HTTP 요청 중 경로와 method에 맞는 handler로 분배한다.
	// http.NewServeMux()는 *http.ServeMux, 즉 ServeMux 구조체의 포인터를 반환한다.
	mux := http.NewServeMux()
	httpapi.RegisterHealthRoutes(mux)
	httpapi.RegisterMerchantRoutes(mux, db)
	httpapi.RegisterInvoiceRoutes(mux, db)
	httpapi.RegisterPaymentRoutes(mux, db)

	// &http.Server{...}는 Server 구조체 값을 만들고, 그 값의 메모리 주소를 포인터로 가져온다.
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("stablepay api listening on :%s", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
```

## Step 9. 커밋 메시지

Day9 코드 실습을 완료하고 테스트까지 통과했다면 아래 커밋 메시지를 사용합니다.

```bash
git status
git add internal/platform/config/config.go cmd/api/main.go docs/domain/06_백엔드코어/Day9_설정과_시작구조/Day9_실습산출물.md
git commit -m "feat: 설정 로딩 config 패키지 추가"
```

커밋에 포함할 파일:

```text
internal/platform/config/config.go
cmd/api/main.go
docs/domain/06_백엔드코어/Day9_설정과_시작구조/Day9_실습산출물.md
```

## 완료 기준

- [ ] main.go 실행 순서를 작성했다.
- [ ] 현재 설정 후보를 작성했다.
- [ ] Phase 2 설정 후보를 작성했다.
- [ ] config 구현 후보를 작성했다.
- [ ] 설정 누락 처리 기준을 작성했다.
- [ ] `internal/platform/config/config.go`를 작성했다.
- [ ] `cmd/api/main.go`가 `config.Load()`를 사용하도록 수정했다.
- [ ] `gofmt -w internal/platform/config/config.go cmd/api/main.go`를 실행했다.
- [ ] `go test ./...`를 실행했다.
- [ ] 완성본과 내 코드를 비교했다.
- [ ] 커밋 메시지를 확인했다.
