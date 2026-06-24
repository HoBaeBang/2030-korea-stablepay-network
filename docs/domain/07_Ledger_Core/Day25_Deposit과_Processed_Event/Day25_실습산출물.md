# Day 25 실습산출물 - Deposit과 Processed Event

관련 Jira: [SPN-42](https://aslan0.atlassian.net/browse/SPN-42)  
관련 Confluence:
[Day25 메인](https://aslan0.atlassian.net/wiki/spaces/SPN/pages/14876674/Deposit+Processed+Event),
[구현가이드](https://aslan0.atlassian.net/wiki/spaces/SPN/pages/14909442),
[실습산출물](https://aslan0.atlassian.net/wiki/spaces/SPN/pages/14942209)

이 문서는 Day25 구현을 마친 뒤, 온체인 입금 event를 내부 Ledger에 한 번만 반영하는 흐름을 자기 말로 정리하는 산출물입니다.

먼저 구현과 테스트를 끝낸 뒤 답변합니다. 힌트와 정답은 막혔을 때만 펼쳐봅니다.

## 1. Day25의 대표 기능 흐름을 한 문장으로 설명하라

무엇을 보고 답해야 하는가:

```text
internal/deposit/service.go의 Processor.ProcessEvent 흐름
Day25 구현가이드의 "대표 진입점 함수와 전체 호출 흐름"
```

<details>
<summary>힌트 보기</summary>

누가 event를 넘겨주는지, 어떤 중복 위험이 있는지, 최종적으로 어떤 DB 상태가 만들어지는지 연결합니다.

</details>

내 답변:

```text

```

## 2. 같은 blockchain event가 여러 번 들어올 수 있는 이유는 무엇인가?

무엇을 보고 답해야 하는가:

```text
RPC 재시도
Indexer 재시작
block range 중복 polling
장애 복구 재처리
```

<details>
<summary>힌트 보기</summary>

중복 event가 항상 버그는 아닙니다. 안전한 indexer는 일부러 과거 구간을 다시 읽을 수도 있습니다.

</details>

내 답변:

```text

```

## 3. `processed_events` 테이블과 `deposits` 테이블의 역할 차이는 무엇인가?

무엇을 보고 답해야 하는가:

```text
migrations/000004_create_deposit_tables.up.sql
internal/deposit/repository.go의 SaveDepositCredit
```

<details>
<summary>힌트 보기</summary>

하나는 event 처리 여부를 판단하는 장치이고, 하나는 입금이라는 업무 기록입니다.

</details>

내 답변:

```text

```

## 4. Event Key를 `chain + tx_hash + log_index`로 만드는 이유는 무엇인가?

무엇을 보고 답해야 하는가:

```text
internal/deposit/deposit.go의 ChainTransferEvent.EventKey
processed_events의 unique 기준
```

<details>
<summary>힌트 보기</summary>

하나의 transaction 안에도 여러 event log가 있을 수 있습니다. `tx_hash`만으로는 부족할 수 있습니다.

</details>

내 답변:

```text

```

## 5. Deposit을 Ledger에 반영할 때 왜 entry가 2개 필요한가?

무엇을 보고 답해야 하는가:

```text
internal/deposit/service.go에서 생성하는 ledger.Entry 목록
internal/ledger/service.go의 ValidateTransaction 규칙
```

<details>
<summary>힌트 보기</summary>

Ledger는 돈이 증가한 한쪽만 기록하지 않습니다. debit과 credit의 합계가 맞아야 합니다.

</details>

내 답변:

```text

```

## 6. 중복 event를 `error`가 아니라 `Duplicate result`로 반환하는 이유는 무엇인가?

무엇을 보고 답해야 하는가:

```text
internal/deposit/service.go의 ProcessEvent
internal/deposit/repository.go의 SaveDepositCredit 반환값
```

<details>
<summary>힌트 보기</summary>

중복 event 확인은 처리 실패가 아니라 멱등 처리의 정상 결과입니다.

</details>

내 답변:

```text

```

## 오늘 실행 결과

실행 명령:

```bash
go fmt ./internal/ledger ./internal/deposit
go test ./internal/deposit -run TestProcessorProcessEvent -v
TEST_DATABASE_URL="postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" \
go test ./internal/deposit -run TestRepositorySaveDepositCredit -v
go test ./...
go vet ./...
```

기록:

```text

```

## 오늘 구현한 파일과 동작

실제 코드와 비교해 완료한 항목을 표시합니다.

```text
[ ] migrations/000004_create_deposit_tables.up.sql
[ ] migrations/000004_create_deposit_tables.down.sql
[ ] internal/ledger/ledger.go에 DEPOSIT_CLEARING 추가
[ ] internal/deposit/deposit.go
[ ] internal/deposit/service.go
[ ] internal/deposit/repository.go
[ ] internal/deposit/service_test.go
[ ] internal/deposit/repository_test.go
[ ] 새 event는 Deposit과 Ledger로 저장됨
[ ] 같은 event 재처리는 Duplicate result로 반환됨
[ ] 실제 PostgreSQL 통합 테스트 통과
```

## 코드 흐름을 한 문장씩 정리하기

```text
DepositProcessor.ProcessEvent:

ChainTransferEvent.EventKey:

Repository.SaveDepositCredit:

processed_events:

deposits:

ledger_transactions / ledger_entries:

fakeStore:
```

## 아직 헷갈리는 부분

아래 후보 중 실제로 헷갈리는 내용만 남기거나 직접 추가합니다.

```text
Flow-first 방식으로 service를 먼저 보는 이유
processed_events와 deposits 차이
event_key를 만드는 기준
idempotency와 duplicate result 차이
입금인데 debit entry가 필요한 이유
Repository가 Ledger 테이블까지 같은 transaction에서 저장하는 이유
interface에 fake store와 repository를 모두 넣을 수 있는 이유
slice append와 entries 복사 패턴
```

메모:

```text

```

## 정답/점검 가이드

답변을 먼저 작성한 뒤 비교합니다.

### 1. 대표 기능 흐름

<details>
<summary>답변 보기</summary>

Indexer 또는 테스트가 넘긴 온체인 transfer event를 `DepositProcessor.ProcessEvent`가 검증하고, `chain + tx_hash + log_index`로 중복 여부를 확인한 뒤, 처음 보는 event라면 `deposits`와 `ledger_transactions`, `ledger_entries`를 한 번만 저장하는 흐름입니다.

</details>

### 2. 같은 event가 여러 번 들어오는 이유

<details>
<summary>답변 보기</summary>

RPC 호출 실패 후 재시도, worker 재시작, 안전한 장애 복구를 위한 과거 block range 재조회, 여러 worker의 중복 polling 때문에 같은 event가 여러 번 들어올 수 있습니다. 따라서 중복 event 자체를 비정상으로 보면 안 되고, 같은 event가 들어와도 결과가 한 번만 반영되도록 만들어야 합니다.

</details>

### 3. `processed_events`와 `deposits` 차이

<details>
<summary>답변 보기</summary>

`processed_events`는 blockchain event를 이미 처리했는지 판단하기 위한 멱등성 테이블입니다. `deposits`는 그 event가 입금으로 해석되어 사용자, 금액, 계정, ledger transaction과 연결된 업무 기록입니다. 즉 processed event는 중복 방지 장치이고, deposit은 입금 도메인 기록입니다.

</details>

### 4. Event Key 기준

<details>
<summary>답변 보기</summary>

같은 chain 안에서 transaction hash는 transaction을 식별하지만, 하나의 transaction 안에 여러 event log가 있을 수 있습니다. 그래서 `tx_hash`만 쓰면 서로 다른 log를 같은 event로 착각할 수 있습니다. `chain + tx_hash + log_index`를 함께 쓰면 어떤 체인의 어떤 transaction 안의 몇 번째 log인지 식별할 수 있습니다.

</details>

### 5. Ledger entry가 2개 필요한 이유

<details>
<summary>답변 보기</summary>

Ledger는 돈의 증가 한 줄만 저장하지 않고 debit과 credit 합계가 맞는 거래를 저장합니다. 온체인 입금으로 고객 내부 잔액을 늘리려면 고객 계정에 `CREDIT`을 만들고, 반대편에는 `DEPOSIT_CLEARING` 계정의 `DEBIT`을 만들어 같은 금액과 통화로 균형을 맞춰야 합니다.

</details>

### 6. Duplicate result와 error 차이

<details>
<summary>답변 보기</summary>

DB 장애, context 취소, 필수값 누락은 처리를 수행할 수 없는 실패이므로 `error`입니다. 반면 이미 처리한 event를 다시 발견한 것은 멱등 처리에서 예상 가능한 정상 결과입니다. 따라서 중복 event는 `error`가 아니라 `Duplicate result`로 반환해서 호출자가 안전하게 무시하거나 로그만 남길 수 있게 합니다.

</details>

## 추가 보충 정리

Day25 완료 후 Codex가 아래 내용을 실제 코드와 답변을 기준으로 채웁니다.

### Codex 점검 예정 항목

```text
- DepositProcessor.ProcessEvent 흐름을 먼저 설명했는가?
- processed_events를 단순 로그 테이블이라고 오해하지 않았는가?
- deposits와 processed_events의 책임을 분리했는가?
- tx_hash만으로 event idempotency를 잡는다고 설명하지 않았는가?
- 입금 credit만 만들면 Ledger 균형이 깨진다는 점을 이해했는가?
- duplicate event와 error를 구분했는가?
- 실제 PostgreSQL 통합 테스트가 실행됐는가?
```

### 다음 학습 포인트

Day26에서는 mock Event Indexer를 만들어 block range에서 event를 읽고, Day25의 `DepositProcessor.ProcessEvent`로 넘기는 흐름을 구현합니다. 이어서 Withdrawal 상태 모델도 함께 잡아 입금과 출금의 차이를 비교합니다.
