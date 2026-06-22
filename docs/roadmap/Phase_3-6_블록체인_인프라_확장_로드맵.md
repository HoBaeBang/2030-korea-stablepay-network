# Phase 3~6 블록체인 인프라 확장 로드맵

이 문서는 `2030 KOREA StablePay Network`의 Phase 2 이후 확장 방향을 정리한다.

StablePay를 새 프로젝트로 버리고 다시 만드는 것이 아니라, 지금 구현 중인 결제·원장·정산 백엔드를 실제 퍼블릭 체인과 연결하고 대용량 블록체인 데이터 인프라로 확장한다.

## 확장 목표

최종 포트폴리오는 다음 질문에 코드와 실행 결과로 답할 수 있어야 한다.

1. 퍼블릭 체인에서 블록, 트랜잭션, 이벤트를 어떻게 읽는가?
2. 같은 이벤트를 여러 번 읽어도 왜 돈이 중복 반영되지 않는가?
3. 체인 재구성, RPC 장애, 프로세스 재시작 후 어떻게 복구하는가?
4. 이벤트 처리량이 증가할 때 어떻게 분산하고 병목을 찾는가?
5. Ethereum과 Non-EVM 체인의 차이를 어떻게 흡수하는가?
6. 개인키와 서명 권한을 어떻게 백엔드에서 분리하는가?

## 전체 단계

| Phase | 중심 목표 | 핵심 기술 | 포트폴리오 증거 |
| --- | --- | --- | --- |
| Phase 2 | 금융 백엔드 코어 완성 | Go, PostgreSQL, Ledger, Settlement | 멱등성, DB transaction, 정합성 테스트 |
| Phase 3 | 실제 Ethereum 체인 연결 | JSON-RPC, Block, Transaction, Log | 재시작 가능한 Ethereum Indexer |
| Phase 4 | 대용량 데이터 처리와 RPC 안정화 | Kafka, Redis, Elasticsearch, RPC Proxy | 장애 복구 결과와 성능 지표 |
| Phase 5 | Non-EVM과 운영 환경 확장 | Solana, Docker, Kubernetes, AWS, ArgoCD | 멀티체인 처리와 배포·복구 시나리오 |
| Phase 6 | 키 보안과 체인 코어 심화 | Rust Signer, Smart Contract, OSS 분석 | 독립 서명기와 오픈소스 분석 기록 |

## Phase 3: 실제 Ethereum 체인 연결

### 목표

Mock 이벤트가 아니라 실제 Ethereum 호환 RPC에서 데이터를 읽어 StablePay의 Deposit과 Payment에 연결한다.

### 구현 범위

| 작업 | 설명 |
| --- | --- |
| JSON-RPC Client | 블록 번호, 블록, transaction receipt, log를 조회한다 |
| Ethereum Indexer | 블록 범위를 순서대로 읽고 필요한 이벤트를 파싱한다 |
| Checkpoint | 마지막으로 안전하게 처리한 block height를 저장한다 |
| Event Idempotency | `chain + tx_hash + log_index`로 중복 처리를 막는다 |
| Confirmation/Finality | 정책에서 정한 확정 기준 이후에 자산을 반영한다 |
| Reorg 대응 | 이전 블록 해시 불일치를 감지하고 rollback 또는 replay한다 |
| 장애 복구 | RPC·DB 오류와 프로세스 재시작 후 checkpoint부터 다시 처리한다 |

### 완료 기준

```text
실제 RPC 또는 로컬 Ethereum devnet에서 블록과 ERC-20 Transfer log를 읽는다.
Indexer를 재시작해도 이벤트가 유실되거나 중복 반영되지 않는다.
RPC 실패와 DB 실패 테스트가 있다.
Reorg를 모사한 테스트 또는 재처리 시나리오가 있다.
```

## Phase 4: 대용량 데이터 처리와 RPC 안정화

### 기술 도입 원칙

Kafka, Redis, Elasticsearch는 포트폴리오의 기술 개수를 늘리기 위해 넣지 않는다. 아래 문제가 실제로 나타나는 시점에 도입하고, 도입 전후의 차이를 기록한다.

| 기술 | 해결할 문제 | 적용 후보 |
| --- | --- | --- |
| Kafka | 수집과 처리를 분리하고 이벤트를 재처리해야 한다 | Indexer가 읽은 이벤트를 Processor에 전달 |
| Redis | 짧은 수명의 공유 상태와 빠른 제한이 필요하다 | Rate Limit, RPC 응답 캐시, 분산 작업 보조 |
| Elasticsearch | 대량의 transaction·event를 여러 조건으로 검색해야 한다 | 운영 조회 API와 탐색 화면 |
| RPC Proxy | 여러 RPC 노드의 장애와 트래픽 편중을 줄여야 한다 | Health Check, 분산, 재시도, Circuit Breaker |

### RPC Proxy 구현 범위

```text
복수 upstream RPC 등록
주기적 health check
weighted round-robin 또는 least-connection 분산
timeout, retry, circuit breaker
요청별 rate limit
upstream별 latency와 error rate 측정
```

### 성능 검증

평균 응답시간만 기록하지 않는다.

```text
처리량: events/sec, requests/sec
지연시간: p50, p95, p99
안정성: error rate, retry count, consumer lag
자원: CPU, memory, DB connection, goroutine 수
복구: 장애 발생 후 정상화까지 걸린 시간
```

### 완료 기준

```text
Docker Compose 한 번으로 데이터 파이프라인을 실행할 수 있다.
Kafka consumer를 중단했다가 다시 실행해도 처리를 이어간다.
RPC 노드 하나가 실패해도 다른 노드로 요청이 전달된다.
부하 테스트 결과와 병목 개선 전후 수치가 README에 기록된다.
```

## Phase 5: Non-EVM과 운영 환경 확장

### Non-EVM 체인

첫 Non-EVM 대상으로 Solana를 사용한다. Ethereum 코드를 그대로 복사하지 않고 공통 처리 흐름과 체인별 구현을 분리한다.

```text
공통 경계
= checkpoint, raw event 저장, idempotency, retry, 관측 지표

체인별 구현
= RPC 요청, block/transaction 구조, event 파싱, finality 판정
```

### 운영 환경

| 작업 | 목적 |
| --- | --- |
| Docker | 로컬 재현 가능한 실행 환경 |
| Kubernetes | API, Indexer, Processor를 독립적으로 배포·확장 |
| AWS | 외부에서 접근 가능한 검증 환경 |
| ArgoCD | Git 변경을 기준으로 반복 가능한 배포 |
| Observability | 로그, metric, trace로 장애 원인을 찾기 |

### 완료 기준

```text
Ethereum과 Solana 데이터를 동일한 조회 API에서 구분해 볼 수 있다.
한 체인의 장애가 다른 체인의 처리를 막지 않는다.
Kubernetes에서 worker 수를 늘려 처리량 변화를 확인한다.
배포, rollback, 장애 복구 절차가 문서화된다.
```

## Phase 6: Rust 보안 경계와 체인 코어 심화

Rust는 Go 백엔드를 대체하지 않는다. 개인키와 서명처럼 메모리 안전성과 권한 분리가 중요한 작은 경계부터 적용한다.

### 구현 범위

```text
Go Backend -> 서명 요청 생성
Rust Signer -> 정책 확인과 transaction 서명
Go Broadcaster -> signed transaction 전송
Indexer -> transaction 결과 추적
```

추가 심화 후보:

```text
간단한 ERC-20 결제 Smart Contract
Geth 또는 Solana 오픈소스 코드 흐름 분석
문서·테스트·작은 버그 수정 형태의 오픈소스 기여
작은 자체 chain prototype과 local devnet
```

## 실제 서비스처럼 보이게 만드는 기준

기능 구현만으로 완료하지 않는다. 각 Phase는 아래 증거를 남긴다.

```text
한 명령으로 실행 가능한 로컬 환경
단위 테스트와 실제 DB/RPC 통합 테스트
중복, 재시작, 네트워크 실패, 부분 실패 시나리오
부하 테스트와 p95/p99 결과
설계 결정과 trade-off를 기록한 ADR
운영 dashboard 또는 metric 예시
README의 아키텍처, 실행 방법, 현재 한계
```

## 예상 기간

퇴근 후와 주말을 사용하는 현재 학습 속도 기준의 범위다.

| 구간 | 예상 기간 |
| --- | --- |
| Phase 2 완료 | 2~3주 |
| Phase 3 Ethereum Indexer | 4~6주 |
| Phase 4 데이터 파이프라인과 RPC Proxy | 6~8주 |
| Phase 5 Solana와 Kubernetes 운영 | 6~8주 |
| Phase 6 Rust·보안·포트폴리오 심화 | 4~8주 |

지원 시점은 두 단계로 본다.

```text
3~4개월:
블록체인 결제 백엔드 또는 주니어 블록체인 백엔드 직무에 지원을 시작한다.

6~9개월:
멀티체인 인덱싱, RPC Proxy, 성능·운영 경험을 요구하는 블록체인 코어 백엔드 직무에 도전한다.
```

경력 연차 조건은 포트폴리오만으로 대체할 수 없으므로, 기술 준비와 별개로 도전 지원과 인접 직무 지원을 병행한다.

## 진행 원칙

```text
Phase 2를 먼저 코드와 테스트로 끝낸다.
다음 Phase의 상세 Day 문서는 직전 Phase가 끝난 뒤 만든다.
한 번에 모든 인프라를 붙이지 않는다.
문제 재현 -> 최소 구현 -> 측정 -> 기술 도입 -> 재측정 순서를 지킨다.
각 기술은 왜 필요한지, 실패하면 어떻게 복구하는지 설명할 수 있어야 한다.
```
