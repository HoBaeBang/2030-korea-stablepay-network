# Day 15 검증문제와 답변가이드

관련 Jira: SPN-32

먼저 문제를 풀어보고, 필요할 때만 답변을 펼쳐서 확인합니다.

## 먼저 풀어볼 문제

1. `Service`는 오늘 코드에서 어떤 역할을 하는가?
2. `ValidateTransaction(ctx, entries)`가 `error`만 반환하는 이유는 무엇인가?
3. `len(entries) < 2`를 실패로 보는 이유는 무엇인가?
4. `DEBIT`은 더하고 `CREDIT`은 빼는 방식으로 합계를 계산하는 이유는 무엇인가?
5. `map[string]int64`에서 `string`과 `int64`는 각각 무엇을 의미하는가?
6. 실패해야 하는 테스트에서 `err == nil`이면 왜 테스트 실패인가?
7. Day15 이후 Repository와 DB migration이 필요한 이유는 무엇인가?

## 내 답변 작성 공간

아래 공간에 먼저 내 생각을 적어봅니다.

정답을 바로 펼치지 말고, 최소 한 문장이라도 먼저 작성한 뒤 답변가이드를 확인합니다.

### 1. `Service`는 오늘 코드에서 어떤 역할을 하는가?

```text
내 답변:
```

### 2. `ValidateTransaction(ctx, entries)`가 `error`만 반환하는 이유는 무엇인가?

```text
내 답변:
```

### 3. `len(entries) < 2`를 실패로 보는 이유는 무엇인가?

```text
내 답변:
```

### 4. `DEBIT`은 더하고 `CREDIT`은 빼는 방식으로 합계를 계산하는 이유는 무엇인가?

```text
내 답변:
```

### 5. `map[string]int64`에서 `string`과 `int64`는 각각 무엇을 의미하는가?

```text
내 답변:
```

### 6. 실패해야 하는 테스트에서 `err == nil`이면 왜 테스트 실패인가?

```text
내 답변:
```

### 7. Day15 이후 Repository와 DB migration이 필요한 이유는 무엇인가?

```text
내 답변:
```

## 답변가이드

### 1. `Service`는 오늘 코드에서 어떤 역할을 하는가?

<details>
<summary>답변 보기</summary>

`Service`는 Ledger 도메인 규칙을 검증하는 영역입니다.

오늘 만든 `Service`는 DB에 저장하지 않습니다.

대신 Ledger Transaction을 저장해도 되는지 먼저 확인합니다.

```text
Service = 규칙 검증
Repository = 저장
DB = 기록 보존
```

</details>

### 2. `ValidateTransaction(ctx, entries)`가 `error`만 반환하는 이유는 무엇인가?

<details>
<summary>답변 보기</summary>

오늘 함수는 새로운 데이터를 만들어 반환하는 함수가 아닙니다.

입력으로 받은 `entries`가 유효한지만 판단합니다.

그래서 성공하면 `nil` error를 반환하고, 실패하면 실패 이유가 담긴 error를 반환합니다.

```text
nil  -> 검증 성공
error -> 검증 실패
```

</details>

### 3. `len(entries) < 2`를 실패로 보는 이유는 무엇인가?

<details>
<summary>답변 보기</summary>

Ledger Transaction은 돈의 이동을 기록합니다.

돈의 이동은 최소한 한쪽에서 빠지고, 다른 쪽에 들어가는 구조여야 합니다.

Entry가 1개뿐이면 상대 항목이 없어서 균형을 만들 수 없습니다.

그래서 최소 2개 이상이어야 합니다.

</details>

### 4. `DEBIT`은 더하고 `CREDIT`은 빼는 방식으로 합계를 계산하는 이유는 무엇인가?

<details>
<summary>답변 보기</summary>

오늘 코드는 균형 여부를 계산하기 위해 한쪽 방향을 양수, 다른 방향을 음수처럼 다룹니다.

```text
DEBIT  10_000_000  -> +10_000_000
CREDIT 9_800_000   -> -9_800_000
CREDIT 200_000     -> -200_000
```

최종 합계가 0이면 debit과 credit이 정확히 맞는 것입니다.

</details>

### 5. `map[string]int64`에서 `string`과 `int64`는 각각 무엇을 의미하는가?

<details>
<summary>답변 보기</summary>

`string`은 통화 코드입니다.

예:

```text
USDC
KRW
USDT
```

`int64`는 해당 통화의 debit/credit 계산 결과입니다.

통화별로 따로 계산해야 하기 때문에 map을 사용합니다.

```text
totals["USDC"] = 0
totals["KRW"] = 0
```

</details>

### 6. 실패해야 하는 테스트에서 `err == nil`이면 왜 테스트 실패인가?

<details>
<summary>답변 보기</summary>

실패 케이스는 `ValidateTransaction`이 error를 반환해야 정상입니다.

예를 들어 균형이 맞지 않는 Entry가 통과하면 안 됩니다.

따라서 실패해야 하는 입력을 넣었는데 `err == nil`이면 “잘못된 데이터가 통과했다”는 뜻입니다.

그래서 테스트를 실패시켜야 합니다.

```go
if err := svc.ValidateTransaction(context.Background(), entries); err == nil {
	t.Fatal("에러가 발생해야 하는데 nil이 반환되었습니다")
}
```

</details>

### 7. Day15 이후 Repository와 DB migration이 필요한 이유는 무엇인가?

<details>
<summary>답변 보기</summary>

Day15의 Service는 검증만 합니다.

하지만 실제 서비스에서는 검증된 Ledger Transaction을 나중에 다시 조회할 수 있어야 합니다.

정산, 대사, 장애 복구, 중복 처리 확인을 하려면 기록이 DB에 남아야 합니다.

그래서 다음 단계에서는 아래 구조가 필요합니다.

```text
ledger_accounts
ledger_transactions
ledger_entries
```

그리고 이 테이블에 저장하고 조회하는 Repository가 필요합니다.

</details>
