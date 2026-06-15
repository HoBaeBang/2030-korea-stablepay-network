# Day 17 학습 및 실습가이드 - Ledger Repository 초안 작성

관련 Jira: [SPN-34](https://aslan0.atlassian.net/browse/SPN-34)

Day17은 Day16에서 만든 Ledger DB 테이블을 Go 코드에서 다루기 위한 Repository 초안을 만드는 날입니다.

오늘의 퇴근 후 실습은 작은 코드 작업 하나입니다.

```text
internal/ledger/repository.go 파일을 만들고,
Ledger Repository가 DB 접근 경계라는 사실을 코드로 표현한다.
```

## 오늘의 큰 그림

![Day17 Ledger Repository 초안 작성](https://raw.githubusercontent.com/HoBaeBang/2030-korea-stablepay-network/main/docs/confluence/diagrams/spn34-day17-ledger-repository.png)

## 출퇴근 학습 목표

출퇴근 시간에는 아래 질문에 답할 수 있을 정도로 읽습니다.

```text
1. Repository는 왜 필요한가?
2. Service와 Repository는 무엇이 다른가?
3. Repository가 *sql.DB를 가진다는 말은 무슨 뜻인가?
4. 왜 Day17에서는 INSERT 저장 로직까지 만들지 않는가?
5. Day18에서는 무엇을 검증하게 되는가?
```

## 오늘의 핵심 문장

```text
Service는 규칙을 검증하고,
Repository는 검증된 데이터를 DB에 저장하거나 DB에서 읽어오는 경계다.
```

Day15에서는 Ledger Service가 debit/credit 균형을 검증했습니다.

Day16에서는 Ledger 데이터를 저장할 DB 테이블을 만들었습니다.

Day17에서는 이 둘 사이를 연결하기 위한 첫 번째 코드 경계인 Repository를 만듭니다.

```text
Day15: 저장 전에 검증한다.
Day16: 저장할 테이블을 만든다.
Day17: 저장을 담당할 Go 객체의 자리를 만든다.
Day18: 실제 저장 테스트로 검증한다.
```

## Repository란 무엇인가?

Repository는 DB 접근을 담당하는 객체입니다.

Java 백엔드에서 익숙한 표현으로 보면 Repository 또는 DAO와 비슷한 역할입니다.

다만 Go에서는 Java처럼 class, annotation, framework magic을 크게 사용하지 않습니다.

Go에서는 보통 아래처럼 명시적으로 구조체를 만들고, 그 구조체가 DB 연결을 필드로 가집니다.

```go
type Repository struct {
    db *sql.DB
}
```

이 코드는 아래 의미입니다.

| 코드 | 의미 |
| --- | --- |
| `type Repository struct` | Repository라는 구조체 타입을 만든다 |
| `db *sql.DB` | Repository가 DB 연결 풀을 가리키는 포인터를 가진다 |
| `*sql.DB` | Go 표준 라이브러리 `database/sql`이 제공하는 DB 연결 풀 타입 |

여기서 `*sql.DB`는 단일 DB 연결 하나라기보다 DB 연결 풀에 가깝습니다.

즉, Repository가 `*sql.DB`를 가진다는 말은 아래와 같습니다.

```text
Repository가 DB에 query/insert/update/delete를 실행할 수 있는 통로를 가진다.
```

## Service와 Repository의 차이

둘을 섞으면 코드가 빠르게 지저분해집니다.

그래서 역할을 분리합니다.

| 구분 | 책임 | 예시 |
| --- | --- | --- |
| Service | 도메인 규칙을 검증하고 업무 흐름을 결정한다 | debit 합계와 credit 합계가 같은지 확인 |
| Repository | DB에 저장하거나 DB에서 읽어온다 | `ledger_transactions`, `ledger_entries`에 INSERT |

Day17 기준으로는 아래처럼 생각하면 됩니다.

```text
Service
-> 이 Ledger Transaction을 저장해도 되는지 판단한다.

Repository
-> 판단이 끝난 데이터를 PostgreSQL에 저장한다.
```

## 왜 바로 INSERT까지 만들지 않는가?

Repository를 만들면 바로 INSERT SQL을 작성하고 싶을 수 있습니다.

하지만 Day17에서는 일부러 거기까지 가지 않습니다.

이유는 아래와 같습니다.

| 이유 | 설명 |
| --- | --- |
| 학습 단위를 작게 유지 | Repository의 책임과 `*sql.DB` 구조를 먼저 이해한다 |
| 테스트 흐름 분리 | 실제 INSERT 검증은 DB가 필요하므로 Day18에서 별도로 다룬다 |
| 실패 지점 분리 | 오늘은 컴파일과 구조 확인, 내일은 저장 로직과 DB 검증에 집중한다 |
| 설계 안정성 | 저장 메서드의 입력 타입과 transaction 처리 방식을 다음 단계에서 더 신중히 정한다 |

## 오늘 확인할 기존 파일

오늘 수정하지 않고 확인만 하는 파일입니다.

### `internal/ledger/ledger.go`

확인할 타입:

```go
type Account struct {
	ID        string
	Type      AccountType
	OwnerID   string
	Currency  string
	CreatedAt time.Time
}

type Transaction struct {
	ID             string
	ReferenceType  string
	ReferenceID    string
	IdempotencyKey string
	CreatedAt      time.Time
}

type Entry struct {
	ID            string
	TransactionID string
	AccountID     string
	Direction     EntryDirection
	Amount        int64
	Currency      string
	CreatedAt     time.Time
}
```

확인 포인트:

```text
Go 타입은 메모리 안의 데이터 모양이다.
Day16 migration은 이 타입들이 DB에 저장될 테이블 모양이다.
Day17 repository는 이 둘 사이의 저장 경계다.
```

### `internal/ledger/service.go`

확인할 메서드:

```go
func (s *Service) ValidateTransaction(ctx context.Context, entries []Entry) error
```

확인 포인트:

```text
Service는 DB에 저장하지 않는다.
Service는 entries가 Ledger 규칙을 만족하는지 검증한다.
```

## 오늘 만들 파일

새로 만들 파일:

```text
internal/ledger/repository.go
```

오늘 작성할 코드의 목표:

```text
1. Repository 구조체를 만든다.
2. Repository가 *sql.DB를 필드로 가지게 한다.
3. NewRepository 생성자를 만든다.
4. 아직 저장 메서드는 만들지 않는다.
```

## Step 1. 파일 생성

아래 파일을 생성합니다.

```text
internal/ledger/repository.go
```

파일이 이미 있는지 먼저 확인합니다.

```bash
sed -n '1,200p' internal/ledger/repository.go
```

파일이 없으면 에러가 나도 괜찮습니다.

```text
No such file or directory
```

## Step 2. package 선언

`repository.go`는 `internal/ledger` 폴더 안에 있으므로 package 이름은 `ledger`입니다.

```go
package ledger
```

Go에서는 같은 폴더 안의 파일들이 같은 package 이름을 가집니다.

그래서 `ledger.go`, `service.go`, `repository.go`는 모두 `package ledger`입니다.

## Step 3. `database/sql` import

Repository는 DB 연결 풀 타입인 `*sql.DB`를 사용합니다.

그래서 표준 라이브러리 `database/sql`을 import합니다.

```go
import "database/sql"
```

`database/sql`은 Go 표준 라이브러리입니다.

즉, 별도 외부 라이브러리를 설치하는 것이 아니라 Go가 기본으로 제공하는 DB 추상화 패키지입니다.

## Step 4. Repository 구조체 작성

```go
type Repository struct {
	db *sql.DB
}
```

이 구조체는 아래 의미입니다.

```text
Repository는 DB 작업을 담당한다.
DB 작업을 하려면 *sql.DB가 필요하다.
그래서 Repository가 db 필드로 *sql.DB를 가진다.
```

여기서 `db`가 소문자인 이유도 중요합니다.

Go에서 소문자로 시작하는 필드나 함수는 package 밖으로 공개되지 않습니다.

즉, `db` 필드는 `ledger` package 내부에서만 직접 접근할 수 있습니다.

외부에서는 `NewRepository` 생성자를 통해 Repository를 만들게 됩니다.

## Step 5. 생성자 작성

```go
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}
```

이 코드를 천천히 풀어보면 아래와 같습니다.

| 코드 | 의미 |
| --- | --- |
| `func NewRepository` | Repository를 만드는 일반 함수 |
| `(db *sql.DB)` | DB 연결 풀 포인터를 파라미터로 받는다 |
| `*Repository` | Repository 포인터를 반환한다 |
| `&Repository{db: db}` | Repository 구조체 값을 만들고 그 주소를 반환한다 |

Java에 비유하면 아래와 비슷합니다.

```java
public Repository(DataSource db) {
    this.db = db;
}
```

다만 Go에는 생성자 문법이 따로 없어서, `NewRepository` 같은 일반 함수를 생성자처럼 사용합니다.

## `repository.go` 최종 완성본 전체

<details>
<summary>repository.go 최종 완성본 전체 보기</summary>

```go
package ledger

import "database/sql"

// Repository는 Ledger 데이터를 DB에 저장하고 조회하는 경계이다.
type Repository struct {
	db *sql.DB
}

// NewRepository는 Ledger Repository 인스턴스를 만든다.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}
```

</details>

## Step 6. gofmt 실행

파일 작성 후 Go 포맷을 맞춥니다.

```bash
gofmt -w internal/ledger/repository.go
```

여러 Go 파일을 한 번에 포맷하고 싶다면 아래처럼 실행할 수 있습니다.

```bash
gofmt -w internal/ledger
```

`gofmt`는 Go 코드의 들여쓰기, import 정렬, 공백 스타일을 Go 표준 스타일로 맞춰주는 도구입니다.

## Step 7. 테스트 실행

오늘은 새 repository 파일을 만들지만 아직 DB 저장 메서드를 작성하지 않습니다.

그래도 새 파일이 컴파일을 깨지 않는지 확인해야 합니다.

```bash
go test ./internal/ledger -v
go test ./...
```

예상 결과:

```text
internal/ledger 테스트 통과
전체 프로젝트 테스트 통과
```

## 자주 만날 수 있는 오류

### `undefined: sql`

`database/sql` import가 빠졌을 가능성이 큽니다.

확인:

```go
import "database/sql"
```

### `imported and not used`

Go는 import한 패키지를 사용하지 않으면 컴파일 에러를 냅니다.

Day17 코드에서는 `*sql.DB`를 사용하므로 `database/sql` import가 사용됩니다.

### `cannot use db as *sql.DB`

`NewRepository`에 넘기는 값의 타입이 `*sql.DB`가 아닐 때 발생할 수 있습니다.

Day17에서는 실제 연결 코드를 작성하지 않으므로 이 오류를 만날 가능성은 낮습니다.

## 완성 기준

아래를 모두 만족하면 Day17 실습은 완료입니다.

```text
internal/ledger/repository.go 파일이 있다.
Repository 구조체가 있다.
Repository가 db *sql.DB 필드를 가진다.
NewRepository(db *sql.DB) *Repository 생성자가 있다.
go test ./internal/ledger -v 가 성공한다.
go test ./... 가 성공한다.
Day17 실습산출물 5문항을 작성한다.
```

## 실습산출물 작성

`Day17_실습산출물.md`에는 5개 질문만 답합니다.

```text
1. Repository는 Service와 무엇이 다른가?
2. Repository가 *sql.DB를 가지는 이유는 무엇인가?
3. NewRepository(db *sql.DB) *Repository는 어떤 흐름으로 Repository를 만드는가?
4. 오늘 INSERT SQL을 만들지 않은 이유는 무엇인가?
5. Day18에서 어떤 테스트가 필요할 것 같은가?
```

## 커밋 메시지

코드 작업을 완료했다면 아래 커밋 메시지를 사용합니다.

```bash
git status
git add internal/ledger/repository.go
git commit -m "feat: Ledger Repository 초안 추가"
```

산출물 문서를 함께 작성했다면 문서 커밋을 분리하는 것이 좋습니다.

```bash
git add docs/domain/07_Ledger_Core/Day17_Ledger_Repository_초안/Day17_실습산출물.md
git commit -m "docs: Day17 Ledger Repository 산출물 정리"
```

## 다음 작업 예고

Day18에서는 Repository 저장 테스트로 넘어갑니다.

```text
Day17: Repository의 자리를 만든다.
Day18: Repository가 실제로 ledger_transactions와 ledger_entries를 저장하는지 테스트한다.
```

