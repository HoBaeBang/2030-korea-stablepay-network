# 2030 KOREA StablePay Network

`2030 KOREA StablePay Network`는 스테이블코인 결제 흐름을 학습하고 포트폴리오로 보여주기 위한 Go 기반 블록체인 결제 백엔드 프로젝트입니다.

Phase 1의 목표는 자체 코인이나 체인을 바로 만드는 것이 아니라, 가맹점이 스테이블코인 결제를 받을 때 필요한 결제 백엔드의 기본 흐름을 구현하는 것입니다.

```text
Merchant -> Invoice -> Payment -> Payment Status
```

## 현재 구현 범위

현재 public repo는 Phase 1 MVP 단계입니다.

- Go HTTP API server
- PostgreSQL 연결
- SQL migration
- Merchant 생성 API
- Invoice 생성 API
- Payment 생성 API
- Payment 상태 변경 API
- Payment 상태 전이 규칙
- Service layer unit test
- HTTP client 실행 예시

아직 실제 블록체인 RPC, 지갑, 온체인 이벤트 인덱서, 입출금, 정산, Rust 체인 프로토타입은 구현되어 있지 않습니다. 이 부분은 Phase 2 이후 확장 범위입니다.

## 도메인 흐름

이 프로젝트의 핵심 도메인은 세 가지입니다.

```text
Merchant
= StablePay를 사용하는 가맹점

Invoice
= 가맹점이 고객에게 발행하는 결제 요청서

Payment
= invoice에 대해 실제 결제가 어디까지 진행됐는지 추적하는 상태 기록
```

예시 흐름:

```text
1. Cafe Korea라는 가맹점을 생성한다.
2. Cafe Korea가 10,000 USDC 결제 요청서(invoice)를 만든다.
3. 해당 invoice에 대한 payment를 생성한다.
4. 블록체인에서 transaction이 감지됐다고 가정하고 상태를 ONCHAIN_DETECTED로 변경한다.
5. 충분히 확정됐다고 가정하고 상태를 FINALIZED로 변경한다.
```

## Payment 상태

Payment는 단순히 성공/실패만 가지지 않습니다. 블록체인 결제는 감지와 확정 사이에 시간이 있기 때문에 상태를 나누어 관리합니다.

```text
PENDING
= payment가 생성됐지만 아직 온체인 transaction이 감지되지 않은 상태

ONCHAIN_DETECTED
= 블록체인에서 transaction hash가 감지된 상태

FINALIZED
= finality가 충분히 확보되어 결제가 확정된 상태

SETTLED
= 가맹점 정산까지 완료된 상태

FAILED
= 결제가 실패했거나 더 이상 진행할 수 없는 상태
```

정상 흐름:

```text
PENDING -> ONCHAIN_DETECTED -> FINALIZED -> SETTLED
```

차단해야 하는 흐름:

```text
FINALIZED -> PENDING
SETTLED -> FINALIZED
PENDING -> FINALIZED
```

## 프로젝트 구조

```text
api/
  stablepay.http                         HTTP client 실행 예시

docs/
  api/                                   API 실행 문서
  architecture/                          시스템 아키텍처 문서
  roadmap/                               Phase 2 확장 로드맵
  portfolio/                             포트폴리오 범위와 검증 문서
  domain/                                Phase 2 도메인 학습 문서

cmd/
  api/
    main.go                              API server entrypoint

internal/
  httpapi/                               HTTP handler와 route 등록
  merchant/                              Merchant domain, service, repository
  invoice/                               Invoice domain, service, repository
  payment/                               Payment domain, status machine, service, repository
  platform/database/                     PostgreSQL connection helper

migrations/                              SQL migration files
docker-compose.yml                       Local PostgreSQL environment
go.mod                                   Go module definition
go.sum                                   Go dependency checksum file
```

## 실행 방법

### 1. PostgreSQL 실행

```bash
docker compose up -d
docker compose ps
```

### 2. API 서버 실행

기본 DB URL은 `postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable` 입니다.

```bash
go run ./cmd/api
```

다른 포트나 DB URL을 사용하려면 환경변수를 지정합니다.

```bash
PORT=8081 DATABASE_URL="postgres://stablepay:stablepay@localhost:5432/stablepay?sslmode=disable" go run ./cmd/api
```

### 3. Health check

```bash
curl http://localhost:8080/health
```

예상 응답:

```json
{"status":"ok","service":"stablepay-api"}
```

## API 요약

| Method | Path | 설명 |
| --- | --- | --- |
| `GET` | `/health` | API 서버 상태 확인 |
| `POST` | `/merchants` | 가맹점 생성 |
| `POST` | `/merchants/{merchantId}/invoices` | 가맹점의 결제 요청서 생성 |
| `POST` | `/invoices/{invoiceId}/payments` | invoice에 대한 payment 생성 |
| `PATCH` | `/payments/{paymentId}/status` | payment 상태 변경 |

전체 문서 목록과 Confluence 매핑은 [docs/README.md](docs/README.md)에서 확인합니다.

자세한 실행 순서는 [docs/api/README.md](docs/api/README.md)를 보고, 직접 호출할 때는 [api/stablepay.http](api/stablepay.http)를 사용합니다.

시스템이 어떤 구조로 확장될지는 [docs/architecture/target-architecture.md](docs/architecture/target-architecture.md)에서 확인할 수 있습니다.

Phase 2 구현 순서와 확장 전략은 [docs/roadmap/phase-2-roadmap.md](docs/roadmap/phase-2-roadmap.md)에서 확인할 수 있습니다.

포트폴리오에서 보여주려는 범위는 [docs/portfolio/project-scope.md](docs/portfolio/project-scope.md), 로컬 검증 절차는 [docs/portfolio/verification.md](docs/portfolio/verification.md)에 정리되어 있습니다.

Phase 2 도메인 전체 지도는 [docs/domain/phase-2-domain-map.md](docs/domain/phase-2-domain-map.md)에 정리되어 있습니다.

Ledger와 Settlement 학습 자료는 [docs/domain/ledger-and-settlement-basic-learning.md](docs/domain/ledger-and-settlement-basic-learning.md)에서 시작합니다.

## 테스트

```bash
go test ./...
```

테스트는 각 도메인의 service layer를 중심으로 작성되어 있습니다.

- Merchant 생성 검증
- Invoice 생성 검증
- Payment 생성 검증
- Payment 상태 전이 검증

자세한 로컬 검증 절차는 [docs/portfolio/verification.md](docs/portfolio/verification.md)를 참고합니다.

## 현재 한계

현재 Phase 1 MVP는 실제 블록체인 네트워크와 연결되어 있지 않습니다.

현재 상태 변경은 사람이 API로 직접 호출합니다.

```text
현재:
PATCH /payments/{paymentId}/status

미래:
Blockchain Event Indexer가 온체인 이벤트를 읽고 payment 상태를 자동 변경
```

아직 구현하지 않은 영역:

- 실제 wallet 결제
- 실제 transaction hash 조회
- 블록 confirmation/finality 자동 판정
- ledger와 settlement
- deposit/withdrawal
- wallet/key security
- Rust signer lab
- Rust chain prototype

## Phase 2 방향

Phase 1이 결제 백엔드 MVP라면, Phase 2는 거래소/월렛/블록체인 금융 백엔드에 더 가까운 기능을 추가하는 단계입니다.

예정된 확장 방향:

```text
1. Blockchain Backend Core
2. Ledger and Settlement
3. Blockchain Event Indexer
4. Deposit and Withdrawal
5. Wallet and Key Security
6. Rust Signer Lab
7. Rust Chain Prototype
8. Devnet and Operations
```

구체적인 순서와 이유는 [Phase 2 Roadmap](docs/roadmap/phase-2-roadmap.md)에 정리되어 있습니다.

## 포트폴리오 관점

이 프로젝트는 단순 CRUD API가 아니라, 결제 도메인에서 중요한 상태 관리와 데이터 정합성을 학습하기 위한 프로젝트입니다.

현재 Phase 1에서 보여주려는 역량:

- Go backend project structure
- HTTP API design
- PostgreSQL persistence
- Domain-oriented package structure
- Payment status machine
- Service/repository separation
- Unit testing
- Blockchain payment domain understanding

상세한 포트폴리오 범위와 면접 설명 포인트는 [docs/portfolio/project-scope.md](docs/portfolio/project-scope.md)에 정리되어 있습니다.

