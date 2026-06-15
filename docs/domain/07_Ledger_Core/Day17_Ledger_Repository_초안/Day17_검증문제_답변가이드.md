# Day 17 검증문제 답변가이드 - Ledger Repository 초안 작성

관련 Jira: SPN-34

이 문서는 Day17 학습 후 스스로 확인할 검증문제와 답변가이드입니다.

먼저 `내 답변 작성 공간`에 답을 적고, 그 다음 아래 답변가이드를 펼쳐서 비교합니다.

## 검증문제

1. Repository는 왜 필요한가?
2. Service에 DB 저장 코드를 직접 넣으면 어떤 문제가 생길 수 있는가?
3. `*sql.DB`에서 `*`는 무엇을 의미하는가?
4. `Repository{db: db}`에서 `db: db`는 무엇을 의미하는가?
5. `&Repository{db: db}`에서 `&`는 무엇을 의미하는가?
6. Go에서 생성자 문법이 따로 없는데 `NewRepository`를 쓰는 이유는 무엇인가?
7. Day17에서 INSERT SQL을 만들지 않는 이유는 무엇인가?
8. Day18에서 Repository를 검증하려면 어떤 테스트가 필요할까?

## 내 답변 작성 공간

### 1. Repository는 왜 필요한가?

```text

```

### 2. Service에 DB 저장 코드를 직접 넣으면 어떤 문제가 생길 수 있는가?

```text

```

### 3. `*sql.DB`에서 `*`는 무엇을 의미하는가?

```text

```

### 4. `Repository{db: db}`에서 `db: db`는 무엇을 의미하는가?

```text

```

### 5. `&Repository{db: db}`에서 `&`는 무엇을 의미하는가?

```text

```

### 6. Go에서 생성자 문법이 따로 없는데 `NewRepository`를 쓰는 이유는 무엇인가?

```text

```

### 7. Day17에서 INSERT SQL을 만들지 않는 이유는 무엇인가?

```text

```

### 8. Day18에서 Repository를 검증하려면 어떤 테스트가 필요할까?

```text

```

## 답변가이드

### 1. Repository는 왜 필요한가?

<details>
<summary>답변 보기</summary>

Repository는 DB 접근을 Service에서 분리하기 위해 필요합니다.

Service는 도메인 규칙을 검증하고, Repository는 DB에 저장하거나 DB에서 읽어오는 일을 담당합니다.

이렇게 분리하면 Service 테스트는 DB 없이도 가능하고, DB 저장 로직은 Repository 테스트에서 따로 검증할 수 있습니다.

</details>

### 2. Service에 DB 저장 코드를 직접 넣으면 어떤 문제가 생길 수 있는가?

<details>
<summary>답변 보기</summary>

Service에 DB 저장 코드가 직접 들어가면 도메인 규칙과 저장 기술이 섞입니다.

그러면 테스트가 어려워지고, 나중에 저장 방식이 바뀔 때 Service까지 크게 흔들릴 수 있습니다.

Ledger처럼 돈의 이동을 다루는 코드는 규칙 검증과 저장 경계를 분리하는 편이 안전합니다.

</details>

### 3. `*sql.DB`에서 `*`는 무엇을 의미하는가?

<details>
<summary>답변 보기</summary>

`*sql.DB`는 `sql.DB` 값 자체가 아니라 `sql.DB`를 가리키는 포인터 타입입니다.

포인터를 사용하면 큰 구조체를 복사하지 않고 같은 DB 연결 풀 객체를 여러 곳에서 참조할 수 있습니다.

여기서 `sql.DB`는 Go 표준 라이브러리 `database/sql`에 있는 DB 연결 풀 타입입니다.

</details>

### 4. `Repository{db: db}`에서 `db: db`는 무엇을 의미하는가?

<details>
<summary>답변 보기</summary>

왼쪽 `db`는 `Repository` 구조체의 필드 이름입니다.

오른쪽 `db`는 `NewRepository` 함수의 파라미터 이름입니다.

즉, 함수로 받은 `db` 값을 Repository 구조체의 `db` 필드에 넣겠다는 뜻입니다.

</details>

### 5. `&Repository{db: db}`에서 `&`는 무엇을 의미하는가?

<details>
<summary>답변 보기</summary>

`&`는 값의 주소를 가져오는 연산자입니다.

`Repository{db: db}`는 Repository 구조체 값을 만듭니다.

`&Repository{db: db}`는 그 구조체 값의 주소를 반환합니다.

그래서 반환 타입이 `*Repository`가 됩니다.

</details>

### 6. Go에서 생성자 문법이 따로 없는데 `NewRepository`를 쓰는 이유는 무엇인가?

<details>
<summary>답변 보기</summary>

Go에는 Java의 constructor 같은 전용 생성자 문법이 없습니다.

대신 `NewTypeName` 형태의 일반 함수를 만들어 객체 초기화 흐름을 명확히 표현하는 패턴을 자주 사용합니다.

`NewRepository(db)`를 사용하면 외부 코드가 Repository를 만들 때 어떤 의존성이 필요한지 알 수 있습니다.

</details>

### 7. Day17에서 INSERT SQL을 만들지 않는 이유는 무엇인가?

<details>
<summary>답변 보기</summary>

Day17의 목표는 Repository의 경계를 이해하고 초안 파일을 만드는 것입니다.

INSERT SQL까지 만들면 DB transaction, rollback, idempotency, integration test까지 함께 들어가서 학습 단위가 커집니다.

그래서 Day17은 Repository 구조만 만들고, Day18에서 저장 테스트와 함께 INSERT 흐름을 검증합니다.

</details>

### 8. Day18에서 Repository를 검증하려면 어떤 테스트가 필요할까?

<details>
<summary>답변 보기</summary>

Day18에서는 실제 DB에 저장되는지 확인하는 테스트가 필요합니다.

예를 들어 `ledger_transactions` row가 저장되는지, `ledger_entries` 여러 row가 저장되는지, foreign key 관계가 맞는지, 같은 `idempotency_key` 중복 저장이 막히는지 확인해야 합니다.

이 테스트는 단순 unit test보다 DB가 필요한 integration test에 가깝습니다.

</details>

