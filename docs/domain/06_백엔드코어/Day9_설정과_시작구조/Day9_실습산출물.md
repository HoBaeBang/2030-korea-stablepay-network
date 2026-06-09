# Day 9 실습산출물 - 설정 로딩과 애플리케이션 시작 구조

관련 Jira: [SPN-26](https://aslan0.atlassian.net/browse/SPN-26)

## 실습 흐름

![Day9 설정 로딩과 애플리케이션 시작 구조](../../../confluence/diagrams/spn26-day9-config-startup-flow.png)

## 1. main.go 실행 순서

아래 순서를 실제 `cmd/api/main.go`를 읽고 프로젝트 기준으로 채운다.

```text
1. 설정값을 읽는다.
2. DB 연결을 만든다.
3. HTTP mux를 생성한다.
4. merchant/invoice/payment route를 등록한다.
5. HTTP server를 시작한다.
```

## 2. 현재 필요한 설정값

| 설정 | 필수 여부 | 기본값 가능 여부 | 이유 |
| --- | --- | --- | --- |
| PORT | 선택에 가까움 | 가능, 예: 8080 | 로컬 실행에서는 기본 포트를 둘 수 있다 |
| DATABASE_URL | 필수 | 불가능 | DB 연결이 없으면 결제/원장 데이터를 저장할 수 없다 |

## 3. Phase 2에서 추가될 설정값

| 설정 | 필요한 기능 | 설명 |
| --- | --- | --- |
| LOG_LEVEL | Logging | 운영 중 필요한 로그 수준을 조절한다 |
| BLOCKCHAIN_RPC_URL | Event Indexer, Withdrawal | 블록체인 네트워크와 통신할 RPC 주소 |
| INDEXER_POLL_INTERVAL | Event Indexer | 온체인 이벤트를 몇 초마다 조회할지 결정 |
| FINALITY_CONFIRMATIONS | Deposit 확정 | 입금을 확정으로 볼 confirmation 수 |
| SIGNER_BASE_URL | Withdrawal Signer | Rust signer 서비스 호출 주소 |

## 4. config 구현 후보

```text
패키지 위치: internal/platform/config 또는 internal/config
구조체 이름: Config
함수 후보: Load() (*Config, error)
```

고민할 점:

```text
config가 database 패키지를 알아야 하는가?
config는 값을 읽고 검증만 할 것인가?
기본값 적용은 어디에서 할 것인가?
```

## 5. 설정 누락 처리 기준

```text
서버 시작 시 반드시 실패해야 하는 설정: DATABASE_URL
기본값을 사용해도 되는 설정: PORT, LOG_LEVEL
기능별로 나중에 검증해도 되는 설정: BLOCKCHAIN_RPC_URL, SIGNER_BASE_URL
```

단, Event Indexer 프로세스를 실행하는 날에는 `BLOCKCHAIN_RPC_URL`이 필수가 될 수 있다.

Withdrawal 전송 기능을 실행하는 날에는 `SIGNER_BASE_URL`이 필수가 될 수 있다.

## 6. 오늘 헷갈린 개념

작성할 내용:

```text
-
-
-
```

## 7. 오늘의 결론

작성할 내용:

```text
Day 9를 통해 내가 이해한 config의 역할은 ...
```
