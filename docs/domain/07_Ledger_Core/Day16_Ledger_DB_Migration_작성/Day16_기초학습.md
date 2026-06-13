# Day 16 기초학습 - Ledger DB Migration 작성

관련 Jira: SPN-33

Day16은 Ledger Core를 DB에 처음으로 새기는 날입니다.

Day12에서는 Ledger의 도메인 타입을 만들었습니다.

```text
Account
Transaction
Entry
```

Day15에서는 Ledger Transaction이 저장되기 전에 안전한지 검증하는 Service를 준비했습니다.

Day16에서는 그 타입들이 실제 PostgreSQL 테이블로 어떻게 바뀌는지 학습하고, migration SQL을 작성합니다.

## 오늘의 큰 그림

![Day16 Ledger DB Migration 작성](../../../confluence/diagrams/spn33-day16-ledger-migration.png)

## 오늘의 핵심 문장

```text
Go 구조체는 메모리 안의 모양이고,
DB 테이블은 오래 보존되는 기록의 모양이다.
```

Ledger는 돈의 이동 기록입니다.

서비스가 재시작되어도 기록이 남아야 하고, 나중에 정산과 대사에서 다시 조회할 수 있어야 합니다.

그래서 Ledger에는 DB 테이블이 필요합니다.

## 오늘 만들 테이블

| Go 타입 | DB 테이블 | 역할 |
| --- | --- | --- |
| `Account` | `ledger_accounts` | 원장에서 돈이 기록되는 주체 |
| `Transaction` | `ledger_transactions` | 여러 Entry를 하나로 묶는 원장 거래 |
| `Entry` | `ledger_entries` | 실제 돈의 이동 한 줄 |

## 오늘 읽을 순서

| 순서 | 문서 | 목적 |
| --- | --- | --- |
| 1 | [Day16_기초학습.md](Day16_기초학습.md) | 오늘 만들 DB 구조의 큰 그림을 잡는다 |
| 2 | [Day16_개념학습.md](Day16_개념학습.md) | migration, primary key, foreign key, index를 이해한다 |
| 3 | [Day16_실습가이드.md](Day16_실습가이드.md) | migration SQL을 작성하고 적용/롤백을 검증한다 |
| 4 | [Day16_실습산출물.md](Day16_실습산출물.md) | 오늘 만든 테이블과 제약조건을 5문항으로 정리한다 |
| 5 | [Day16_검증문제_답변가이드.md](Day16_검증문제_답변가이드.md) | 먼저 문제를 풀고 답변가이드와 비교한다 |

## 출퇴근 학습에서 잡을 것

출퇴근 시간에는 SQL을 외우기보다 아래 개념을 잡습니다.

```text
테이블은 무엇을 보존하는가?
primary key는 왜 필요한가?
foreign key는 왜 필요한가?
index는 왜 필요한가?
up migration과 down migration은 무엇이 다른가?
```

영어 용어도 같이 잡습니다.

| 용어 | 한글 감각 | 오늘 코드에서의 의미 |
| --- | --- | --- |
| Migration | DB 변경 이력 | 테이블 생성/삭제 SQL |
| Primary Key | 기본키 | 한 row를 유일하게 식별하는 값 |
| Foreign Key | 외래키 | 다른 테이블의 row를 참조하는 값 |
| Index | 색인 | 조회를 빠르게 하기 위한 보조 구조 |
| Rollback | 되돌리기 | 적용한 DB 변경을 취소하는 것 |

## 퇴근 후 작업

오늘 작성할 파일:

```text
migrations/000002_create_ledger_core_tables.up.sql
migrations/000002_create_ledger_core_tables.down.sql
```

오늘 하지 않는 것:

```text
Repository 작성
Service와 DB 연결
HTTP API 작성
Payment FINALIZED와 Ledger 자동 연결
Settlement 계산
```

## 완료 기준

- [ ] Ledger 테이블 3개가 왜 필요한지 설명할 수 있다.
- [ ] `up.sql`과 `down.sql`의 차이를 설명할 수 있다.
- [ ] `ledger_entries.transaction_id`가 왜 외래키인지 설명할 수 있다.
- [ ] migration SQL을 작성했다.
- [ ] 적용과 롤백을 한 번씩 검증했다.
- [ ] Day16 실습산출물 5문항을 작성할 수 있다.

## 다음 작업 예고

Day16이 끝나면 Day17에서는 Repository 초안으로 넘어갑니다.

```text
Day16: 테이블을 만든다.
Day17: 테이블에 저장하는 Go 코드를 만든다.
```
