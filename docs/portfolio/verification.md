# Verification Guide

이 문서는 `2030 KOREA StablePay Network`를 로컬에서 실행하고 검증하는 절차를 정리한다.

목표는 세 가지다.

1. 테스트 명령과 기대 결과를 명확히 남긴다.
2. PostgreSQL과 API 서버 실행 절차를 재현 가능하게 만든다.
3. 문제가 생겼을 때 먼저 확인할 항목을 정리한다.

## 현재 검증 범위

현재 검증 범위는 Phase 1 MVP 기준이다.

검증 가능한 것:

```text
Go unit test
PostgreSQL container 실행
API server 실행
Health check
Merchant 생성
Invoice 생성
Payment 생성
Payment 상태 변경
실패 케이스 일부
```

아직 검증하지 않는 것:

```text
실제 블록체인 RPC 연결
실제 온체인 transaction 조회
실제 wallet 결제
실제 입금/출금
Rust signer
Rust chain prototype
```

## 빠른 검증 순서

가장 짧은 검증 순서는 다음과 같다.

```bash
go test ./...
docker compose up -d
docker compose exec -T postgres psql -U stablepay -d stablepay < migrations/000001_create_payment_core_tables.up.sql
go run ./cmd/api
curl http://localhost:8080/health
```

기대 결과:

```text
go test ./... 통과
PostgreSQL container healthy
DB schema 적용 완료
API server listening on :8080
GET /health 응답 200 OK
```

## 1. Go 테스트 실행

명령:

```bash
go test ./...
```

현재 기대 결과:

```text
cmd/api                                        no test files
internal/httpapi                              no test files
internal/invoice                              ok
internal/merchant                             ok
internal/payment                              ok
internal/platform/database                    no test files
```

테스트가 검증하는 범위:

| 패키지 | 테스트 내용 |
| --- | --- |
| `internal/merchant` | 정상 merchant 생성, name 누락 실패 |
| `internal/invoice` | 정상 invoice 생성, amount 0 실패 |
| `internal/payment` | payment 생성, payment 상태 변경, 잘못된 상태 전이 차단 |

현재 테스트 케이스 예시:

```text
정상 입력하면 이름과 이메일을 정리하고 merchant를 생성한다
name이 비어 있으면 에러를 반환한다
정상 입력이면 통화를 정규화하고 PENDING invoice를 생성한다
amount가 0이면 에러를 반환한다
정상 요청이면 PENDING 상태의 payment를 생성한다
PENDING에서 ONCHAIN_DETECTED로 변경할 수 있다
FINALIZED에서 PENDING으로 되돌릴 수 없다
FINALIZED가 되면 finalized_at을 저장한다
```

## 2. PostgreSQL 실행

명령:

```bash
docker compose up -d
```

상태 확인:

```bash
docker compose ps
```

현재 `docker-compose.yml` 기준:

```text
container name: stablepay-postgres
image: postgres:16-alpine
database: stablepay
user: stablepay
password: stablepay
port: 5432
```

기본 DB URL:

```text
postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable
```

## 3. DB schema 적용

현재 프로젝트는 애플리케이션이 migration을 자동 실행하지 않는다.

따라서 깨끗한 PostgreSQL volume에서 처음 실행한다면 API 서버를 띄우기 전에 schema를 먼저 적용해야 한다.

명령:

```bash
docker compose exec -T postgres psql -U stablepay -d stablepay < migrations/000001_create_payment_core_tables.up.sql
```

기대 결과:

```text
CREATE TABLE
CREATE INDEX
CREATE TABLE
CREATE INDEX
CREATE INDEX
CREATE TABLE
CREATE INDEX
CREATE INDEX
CREATE INDEX
```

이미 테이블이 있는 상태에서 다시 실행하면 `relation already exists` 오류가 날 수 있다.

그 경우에는 기존 DB volume을 쓰고 있다는 의미일 수 있으므로, 지금 단계에서는 중복 실행하지 않는다.

현재 생성되는 테이블:

```text
merchants
invoices
payments
```

## 4. API 서버 실행

명령:

```bash
go run ./cmd/api
```

기대 로그:

```text
database connection ok
stablepay api listening on :8080
```

다른 포트로 실행:

```bash
PORT=8081 go run ./cmd/api
```

다른 DB URL로 실행:

```bash
DATABASE_URL="postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" go run ./cmd/api
```

## 5. Health check

명령:

```bash
curl http://localhost:8080/health
```

기대 응답:

```json
{"status":"ok","service":"stablepay-api"}
```

이 응답이 나오면 API 서버는 요청을 받을 수 있는 상태다.

## 6. API 흐름 검증

자세한 실행 예시는 아래 파일을 사용한다.

```text
api/stablepay.http
```

자세한 설명은 아래 문서를 본다.

```text
docs/api/README.md
```

검증 흐름:

```mermaid
flowchart TD
    Health["GET /health"]
    Merchant["POST /merchants"]
    Invoice["POST /merchants/{merchantId}/invoices"]
    Payment["POST /invoices/{invoiceId}/payments"]
    Detected["PATCH /payments/{paymentId}/status<br/>ONCHAIN_DETECTED"]
    Finalized["PATCH /payments/{paymentId}/status<br/>FINALIZED"]

    Health --> Merchant --> Invoice --> Payment --> Detected --> Finalized
```

### 6.1 Merchant 생성

요청:

```http
POST /merchants
Content-Type: application/json

{
  "name": "Cafe Korea",
  "email": "owner@cafe.example"
}
```

기대 결과:

```text
201 Created
응답 JSON에 id, name, email, created_at 포함
```

응답의 `id`를 `merchantId`로 사용한다.

### 6.2 Invoice 생성

요청:

```http
POST /merchants/{merchantId}/invoices
Content-Type: application/json

{
  "amount": 10000,
  "currency": "USDC"
}
```

기대 결과:

```text
201 Created
응답 JSON에 id, merchant_id, amount, currency, status 포함
status는 PENDING
```

응답의 `id`를 `invoiceId`로 사용한다.

### 6.3 Payment 생성

요청:

```http
POST /invoices/{invoiceId}/payments
Content-Type: application/json

{
  "amount": 10000,
  "currency": "USDC"
}
```

기대 결과:

```text
201 Created
응답 JSON에 id, invoice_id, amount, currency, status 포함
status는 PENDING
```

응답의 `id`를 `paymentId`로 사용한다.

### 6.4 Payment 상태 변경

온체인 감지 상태로 변경:

```http
PATCH /payments/{paymentId}/status
Content-Type: application/json

{
  "status": "ONCHAIN_DETECTED",
  "transaction_hash": "0xabc123"
}
```

기대 결과:

```text
200 OK
status는 ONCHAIN_DETECTED
transaction_hash가 저장됨
```

결제 확정 상태로 변경:

```http
PATCH /payments/{paymentId}/status
Content-Type: application/json

{
  "status": "FINALIZED"
}
```

기대 결과:

```text
200 OK
status는 FINALIZED
finalized_at이 채워짐
```

## 7. 실패 케이스 검증

### 잘못된 이메일

요청:

```http
POST /merchants
Content-Type: application/json

{
  "name": "Invalid Merchant",
  "email": "invalid-email"
}
```

기대 결과:

```text
4xx 응답
email 형식 오류
```

### 지원하지 않는 통화

요청:

```http
POST /invoices/{invoiceId}/payments
Content-Type: application/json

{
  "amount": 10000,
  "currency": "KRW"
}
```

기대 결과:

```text
4xx 응답
지원하지 않는 currency 오류
```

### 허용되지 않는 상태 전이

예:

```text
FINALIZED -> PENDING
PENDING -> FINALIZED
SETTLED -> FINALIZED
```

기대 결과:

```text
4xx 응답
invalid payment status transition
```

## 8. 문제가 생겼을 때 확인할 것

### `connection refused`

의미:

```text
API 서버 또는 PostgreSQL이 실행 중이 아닐 수 있다.
```

확인:

```bash
docker compose ps
go run ./cmd/api
```

### `database connection failed`

의미:

```text
DATABASE_URL이 잘못됐거나 PostgreSQL이 아직 준비되지 않았을 수 있다.
```

확인:

```bash
docker compose ps
docker compose logs postgres
```

### `relation "merchants" does not exist`

의미:

```text
PostgreSQL은 실행 중이지만 migration SQL이 아직 적용되지 않았다.
```

해결:

```bash
docker compose exec -T postgres psql -U stablepay -d stablepay < migrations/000001_create_payment_core_tables.up.sql
```

### `address already in use`

의미:

```text
이미 8080 포트를 사용하는 프로세스가 있다.
```

해결:

```bash
PORT=8081 go run ./cmd/api
```

### HTTP 파일 변수 오류

의미:

```text
merchantId, invoiceId, paymentId 변수에 실제 응답 id를 복사하지 않았을 수 있다.
```

확인:

```text
api/stablepay.http 상단의 변수 값
```

## 9. 검증 체크리스트

작업을 마치기 전에 아래를 확인한다.

```text
[ ] go test ./... 통과
[ ] docker compose up -d 실행
[ ] DB schema 적용
[ ] API 서버 실행
[ ] GET /health 성공
[ ] Merchant 생성 성공
[ ] Invoice 생성 성공
[ ] Payment 생성 성공
[ ] ONCHAIN_DETECTED 상태 변경 성공
[ ] FINALIZED 상태 변경 성공
[ ] 잘못된 상태 전이 실패 확인
[ ] git status clean
```

## 10. 현재 검증의 한계

현재 검증은 Phase 1 MVP 기준이다.

한계:

```text
실제 블록체인 transaction을 검증하지 않는다.
실제 USDC를 전송하지 않는다.
실제 confirmation/finality를 계산하지 않는다.
실제 wallet private key를 다루지 않는다.
부하 테스트나 보안 테스트는 포함하지 않는다.
```

이 한계는 정상이다.

Phase 2에서 Blockchain Event Indexer, Ledger, Settlement, Deposit/Withdrawal이 추가되면 검증 문서도 함께 확장한다.
