# Ledger Core 앞으로의 일정

이 문서는 Day16 이후 `2030 KOREA StablePay Network`의 Ledger, Settlement, Deposit/Withdrawal, Event Indexer 학습과 구현 흐름을 한눈에 보기 위한 일정표입니다.

현재 리듬은 아래 원칙을 따릅니다.

```text
작은 코드 작업 하나
-> 코드 흐름 설명
-> 테스트 또는 SQL 검증
-> 짧은 산출물
-> 리뷰
-> Git/Jira/Confluence 정리
```

## 현재 위치

```text
Day18: Ledger Repository 저장 구현 완료
Day19: Repository 저장 검증과 Idempotency 완료
Day20: Ledger Service와 Repository 연결 완료
Day21: Payment FINALIZED와 Ledger 연결 설계 및 Ledger Core 회고 완료
Day22: Settlement 도메인 타입과 계산 서비스 완료
Day23: Settlement DB 저장과 상태 전이 자료 준비
```

## 전체 흐름

![Ledger Core 앞으로의 일정](../../confluence/diagrams/spn-ledger-core-future-plan.png)

## 압축 일정

| Day | 주제 | 핵심 산출물 | 목적 |
| --- | --- | --- | --- |
| Day16 | Ledger DB Migration 작성 | `000002_create_ledger_core_tables.up/down.sql` | Ledger 타입을 DB 테이블 구조로 옮긴다 |
| Day17 | Ledger Repository 초안 | `internal/ledger/repository.go` | 검증된 Ledger Transaction을 DB에 저장할 준비를 한다 |
| Day18 | Ledger Repository 저장 구현 | `CreateTransaction` 메서드 초안 | 원장 거래 1건과 항목 여러 건을 하나의 DB transaction으로 저장한다 |
| Day19 | Repository 저장 검증과 Idempotency | 저장 검증과 중복 저장 방지 후보 | 저장 성공/실패 흐름과 `idempotency_key` 중복 방지를 점검한다 |
| Day20 | Ledger Service와 Repository 연결 | `ValidateTransaction -> CreateTransaction` 흐름 | Service 검증 후 Repository 저장으로 이어지게 만든다 |
| Day21 | Payment FINALIZED와 Ledger 연결 설계 및 회고 | payment -> ledger 규칙과 Ledger Core 점검 | 결제 확정을 돈의 이동 기록으로 해석하고 Ledger 구간을 마무리한다 |
| Day22 | Settlement 도메인 타입과 계산 서비스 | `internal/settlement` 타입·Calculator·테스트 | 지급 예정 Ledger Entry를 정산 묶음과 항목으로 계산한다 |
| Day23 | Settlement DB와 상태 흐름 | migration, repository, 상태 전이 | 정산 결과를 보존하고 DRAFT 이후 처리 단계를 관리한다 |
| Day24 | Reconciliation과 Settlement 검증 | DB·Ledger 비교와 종합 테스트 | 정산 누락, 중복, 불일치를 찾는다 |
| Day25 | Deposit과 Processed Event | 입금 모델과 중복 이벤트 방지 | 온체인 입금을 안전하게 Ledger에 반영할 준비를 한다 |
| Day26 | Event Indexer와 Withdrawal | polling mock과 출금 모델 | 온체인 이벤트 읽기와 출금 흐름을 연결한다 |
| Day27 | Wallet/Key Security와 포트폴리오 | Rust signer 경계, 종합 검증, README | Phase 2를 마무리하고 채용 결과물로 정리한다 |

## 구현 순서의 이유

먼저 Ledger를 만드는 이유는 단순합니다.

```text
Payment는 상태를 말한다.
Ledger는 돈의 이동을 말한다.
Settlement는 Ledger를 기반으로 지급 가능 금액을 계산한다.
Deposit/Withdrawal은 온체인 이벤트와 Ledger를 연결한다.
Event Indexer는 블록체인에서 발생한 일을 백엔드로 가져온다.
Wallet/Key Security는 출금과 서명의 안전 경계를 만든다.
```

압축 일정에서도 구현 순서는 아래처럼 유지합니다.

```text
Ledger
-> Settlement
-> Deposit/Withdrawal
-> Event Indexer
-> Wallet/Key Security
-> Portfolio packaging
```

## Day16의 위치

Day16은 Ledger를 DB에 처음으로 새기는 날입니다.

아직 repository나 API를 만들지 않습니다.

```text
Day15: 저장 전에 검증한다.
Day16: 저장할 테이블 모양을 만든다.
Day17: 테이블에 저장하는 repository를 만든다.
Day18: 실제 저장 메서드를 구현한다.
Day19: 저장 검증과 중복 방지를 점검한다.
```

## 일정 운영 기준

이 일정은 고정된 계약이 아니라 학습 진도에 맞춰 조정합니다.

다만 한 번에 너무 많은 기능을 붙이지 않기 위해 아래 기준은 유지합니다.

```text
하루에 서로 연결된 기능 흐름 하나를 완성한다.
문서는 코드 작업을 도와야 한다.
산출물은 오늘 구현한 코드와 직접 연결한다.
복습은 각 Day의 마지막 검증에 포함한다.
별도 회고일은 사용자가 특정 구간을 다시 확인하고 싶을 때만 추가한다.
```
