# StablePay API 실행 가이드

이 문서는 `2030 KOREA StablePay Network` Phase 1 API를 로컬에서 실행하고 검증하는 방법을 정리한다.

## 전제 조건

PostgreSQL과 API 서버가 실행되어 있어야 한다.

```bash
docker compose up -d
go run ./cmd/api
```

서버 기본 주소:

```text
http://localhost:8080
```

## 실행 파일

HTTP client 예시는 아래 파일에 있다.

```text
api/stablepay.http
```

JetBrains IDE나 VS Code REST Client 같은 도구에서 열어 각 요청을 순서대로 실행한다.

## 전체 흐름

```text
1. GET /health
   서버 상태 확인

2. POST /merchants
   가맹점 생성

3. POST /merchants/{merchantId}/invoices
   가맹점의 결제 요청서 생성

4. POST /invoices/{invoiceId}/payments
   결제 추적을 위한 payment 생성

5. PATCH /payments/{paymentId}/status
   ONCHAIN_DETECTED 상태로 변경

6. PATCH /payments/{paymentId}/status
   FINALIZED 상태로 변경
```

## 변수 사용법

`api/stablepay.http` 상단에는 다음 변수가 있다.

```http
@baseUrl = http://localhost:8080
@merchantId = mer_응답에서_복사
@invoiceId = inv_응답에서_복사
@paymentId = pay_응답에서_복사
```

각 생성 API를 실행한 뒤 응답 JSON의 `id` 값을 다음 요청의 변수에 복사한다.

예를 들어 merchant 생성 응답이 아래와 같다면:

```json
{
  "id": "mer_1779283381135248000",
  "name": "Cafe Korea",
  "email": "owner@cafe.example",
  "created_at": "2026-05-30T12:00:00Z"
}
```

HTTP 파일 상단을 이렇게 바꾼다.

```http
@merchantId = mer_1779283381135248000
```

## 주요 API

### Health check

```http
GET /health
```

예상 응답:

```json
{
  "status": "ok",
  "service": "stablepay-api"
}
```

### Merchant 생성

```http
POST /merchants
```

요청:

```json
{
  "name": "Cafe Korea",
  "email": "owner@cafe.example"
}
```

성공 시 `201 Created`를 반환한다.

실패 예시:

```json
{
  "error": "email is invalid"
}
```

### Invoice 생성

```http
POST /merchants/{merchantId}/invoices
```

요청:

```json
{
  "amount": 10000,
  "currency": "USDC"
}
```

`expires_at`은 선택 값이다. 넣을 경우 RFC3339 형식을 사용한다.

```json
{
  "amount": 25000,
  "currency": "USDC",
  "expires_at": "2026-06-30T15:00:00Z"
}
```

### Payment 생성

```http
POST /invoices/{invoiceId}/payments
```

요청:

```json
{
  "amount": 10000,
  "currency": "USDC"
}
```

생성 직후 payment는 `PENDING` 상태다.

### Payment 상태 변경

```http
PATCH /payments/{paymentId}/status
```

온체인 transaction 감지:

```json
{
  "status": "ONCHAIN_DETECTED",
  "transaction_hash": "0xabc123"
}
```

결제 확정:

```json
{
  "status": "FINALIZED"
}
```

## 상태 전이 규칙

정상 흐름:

```text
PENDING -> ONCHAIN_DETECTED -> FINALIZED -> SETTLED
```

현재 API로 직접 테스트하는 주요 흐름:

```text
PENDING -> ONCHAIN_DETECTED -> FINALIZED
```

차단되는 예시:

```text
FINALIZED -> PENDING
PENDING -> FINALIZED
SETTLED -> FINALIZED
```

## 현재 한계

현재 Phase 1에서는 실제 블록체인 이벤트를 읽지 않는다.

```text
현재:
사용자가 PATCH API를 호출해서 payment 상태를 변경한다.

미래:
Blockchain Event Indexer가 transaction hash와 finality를 확인한 뒤 자동으로 상태를 변경한다.
```
