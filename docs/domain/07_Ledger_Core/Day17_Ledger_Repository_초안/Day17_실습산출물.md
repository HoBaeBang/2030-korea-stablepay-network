# Day 17 실습산출물 - Ledger Repository 초안 작성

관련 Jira: SPN-34

Day17 산출물은 5개 질문만 작성합니다.

오늘 산출물은 Repository 코드를 많이 외웠는지 확인하는 문서가 아닙니다.

오늘 만든 `Repository`가 왜 필요한지, 그리고 아직 무엇을 하지 않았는지 확인하는 문서입니다.

## 작성 전 확인

아래 파일을 먼저 확인합니다.

```bash
sed -n '1,200p' internal/ledger/repository.go
```

아래 구조가 보여야 합니다.

```go
type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}
```

## 1. Repository는 Service와 무엇이 다른가?

작성 힌트:

```text
Service는 규칙 검증,
Repository는 DB 저장/조회 경계라는 점을 비교해서 적는다.
```

내 답변:

```text

```

## 2. Repository가 `*sql.DB`를 가지는 이유는 무엇인가?

작성 힌트:

```text
*sql.DB가 단일 연결 하나라기보다 DB 연결 풀에 가깝다는 점을 생각한다.
Repository가 DB 작업을 하려면 무엇이 필요한지 적는다.
```

내 답변:

```text

```

## 3. `NewRepository(db *sql.DB) *Repository`는 어떤 흐름으로 Repository를 만드는가?

작성 힌트:

```text
파라미터로 db를 받고,
Repository 구조체를 만들고,
그 주소를 반환한다는 흐름으로 적는다.
```

내 답변:

```text

```

## 4. 오늘 INSERT SQL을 만들지 않은 이유는 무엇인가?

작성 힌트:

```text
Day17은 저장 경계 초안,
Day18은 실제 저장 테스트라는 분리를 생각한다.
```

내 답변:

```text

```

## 5. Day18에서 어떤 테스트가 필요할 것 같은가?

작성 힌트:

```text
ledger_transactions와 ledger_entries가 실제 DB에 저장되는지,
idempotency_key 중복이 막히는지,
transaction_id/account_id 관계가 맞는지 생각한다.
```

내 답변:

```text

```

## 오늘 실행 결과

아래 명령 실행 결과를 짧게 기록합니다.

```bash
gofmt -w internal/ledger/repository.go
go test ./internal/ledger -v
go test ./...
```

기록:

```text

```

## 아직 헷갈리는 부분

아래 후보 중 헷갈리는 것이 있으면 적습니다.

```text
Repository
Service
DAO
*sql.DB
생성자 함수
포인터 반환
구조체 리터럴
DB 연결 풀
```

메모:

```text

```

## 정답/점검 가이드

먼저 `내 답변`을 작성한 뒤 아래 내용을 펼쳐서 비교합니다.

### 1. Repository는 Service와 무엇이 다른가?

<details>
<summary>정답/점검 가이드 보기</summary>

Service는 도메인 규칙을 검증하고 업무 흐름을 결정합니다.

Repository는 DB에 저장하거나 DB에서 조회하는 경계입니다.

Ledger 기준으로 보면 Service는 debit과 credit 합계가 맞는지 검증하고, Repository는 검증된 Transaction과 Entry를 DB 테이블에 저장합니다.

</details>

### 2. Repository가 `*sql.DB`를 가지는 이유는 무엇인가?

<details>
<summary>정답/점검 가이드 보기</summary>

Repository가 DB에 query, insert, update, delete를 실행하려면 DB에 접근할 통로가 필요합니다.

Go 표준 라이브러리에서는 그 통로로 `*sql.DB`를 주로 사용합니다.

`*sql.DB`는 단일 연결 하나가 아니라 DB 연결 풀에 가까운 객체입니다.

그래서 Repository가 `db *sql.DB` 필드를 가지면, 나중에 `r.db.QueryContext`, `r.db.ExecContext` 같은 메서드로 DB 작업을 실행할 수 있습니다.

</details>

### 3. `NewRepository(db *sql.DB) *Repository`는 어떤 흐름으로 Repository를 만드는가?

<details>
<summary>정답/점검 가이드 보기</summary>

`NewRepository`는 Repository를 만드는 일반 함수입니다.

파라미터로 `db *sql.DB`를 받고, `Repository{db: db}` 구조체 값을 만든 뒤, `&`를 붙여 그 주소를 반환합니다.

즉 아래 흐름입니다.

```text
DB 연결 풀을 받는다
-> Repository 구조체에 넣는다
-> Repository의 주소를 반환한다
```

</details>

### 4. 오늘 INSERT SQL을 만들지 않은 이유는 무엇인가?

<details>
<summary>정답/점검 가이드 보기</summary>

Day17의 목표는 저장 로직 전체가 아니라 Repository의 책임과 경계를 만드는 것입니다.

INSERT SQL을 만들면 DB transaction 처리, 실패 시 rollback, idempotency_key 중복 처리, integration test까지 함께 고민해야 합니다.

그래서 Day17에서는 Repository 초안만 만들고, 실제 저장 검증은 Day18에서 작게 이어갑니다.

</details>

### 5. Day18에서 어떤 테스트가 필요할 것 같은가?

<details>
<summary>정답/점검 가이드 보기</summary>

Day18에서는 실제 DB 저장 흐름을 검증해야 합니다.

예를 들어 아래 테스트가 필요합니다.

```text
1. Ledger Transaction이 ledger_transactions에 저장되는가?
2. Ledger Entry 여러 개가 ledger_entries에 저장되는가?
3. Entry의 transaction_id가 저장된 transaction을 가리키는가?
4. Entry의 account_id가 실제 account를 가리키는가?
5. 같은 idempotency_key로 중복 저장하려고 할 때 막히는가?
```

</details>

