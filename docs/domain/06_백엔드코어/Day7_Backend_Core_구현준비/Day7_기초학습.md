# Backend Core 기초 학습

관련 Jira: [SPN-24](https://aslan0.atlassian.net/browse/SPN-24)

이 문서는 Phase 2 Day 7 학습 허브입니다.

Day 7의 목표는 Ledger, Settlement, Indexer 같은 큰 기능을 바로 구현하는 것이 아니라, 그 기능들이 안정적으로 붙을 수 있도록 Go 백엔드의 공통 기반을 먼저 정리하는 것입니다.

## 오늘의 큰 그림

```text
Phase 2 도메인 이해
        |
        v
Backend Core 정리
        |
        v
Ledger / Settlement / Indexer 구현
```

Day 6까지는 왜 Ledger, Settlement, Deposit, Withdrawal, Event Indexer가 필요한지 이해하는 데 집중했습니다.

Day 7부터는 다음 질문으로 넘어갑니다.

```text
이제 실제 구현을 시작한다면,
모든 기능이 공통으로 사용할 백엔드 기반은 어떻게 정리해야 하는가?
```

## 오늘의 목표

1. Backend Core가 무엇인지 설명할 수 있다.
2. 공통 에러 응답이 왜 필요한지 이해한다.
3. 요청 validation을 어디에서 처리할지 판단할 수 있다.
4. config, logging, test pattern이 왜 Phase 2 구현 전에 필요한지 설명할 수 있다.
5. Sprint 2 첫 구현 후보를 실제 작업 단위로 나눌 수 있다.

## 오늘의 권장 진행 순서

| 순서 | 문서 | 언제 보는가 | 목적 |
| --- | --- | --- | --- |
| 1 | [Day7 개념 학습](Day7_개념학습.md) | 출퇴근 학습 | Backend Core의 역할과 구현 후보를 이해한다. |
| 2 | [Day7 실습 가이드](Day7_실습가이드.md) | 퇴근 후 작업 전 | 실습산출물에 무엇을 작성해야 하는지 확인한다. |
| 3 | [Day7 실습산출물](Day7_실습산출물.md) | 퇴근 후 작업 | Sprint 2 구현 전 백엔드 코어 설계 초안을 작성한다. |
| 4 | [Day7 검증문제와 답변가이드](Day7_검증문제_답변가이드.md) | 산출물 작성 후 | 이해한 내용을 문제로 점검한다. |

## Day 7에서 다루는 Backend Core 범위

| 영역 | 의미 | 왜 먼저 필요한가 |
| --- | --- | --- |
| Error Response | API 실패 응답 형식 | 기능마다 에러 응답 모양이 달라지지 않게 한다. |
| Validation | 요청값 검증 | 잘못된 요청을 service/repository까지 보내지 않는다. |
| Config | 설정 구조 | DB, 포트, 외부 RPC 설정을 코드에 흩뿌리지 않는다. |
| Logging | 로그 | 상태 변경과 장애 원인을 추적할 수 있게 한다. |
| Test Pattern | 테스트 패턴 | 새 도메인을 붙일 때 검증 방식을 반복 가능하게 만든다. |

## 오늘 꼭 잡아야 하는 문장

```text
Backend Core는 화려한 기능이 아니라,
화려한 기능이 흔들리지 않게 붙을 수 있는 바닥이다.
```

## 퇴근 후 작업의 원칙

퇴근 후에는 코드를 바로 많이 고치기보다, Sprint 2에서 실제로 고칠 백엔드 기반을 먼저 정리합니다.

1. 현재 코드에서 공통화할 수 있는 부분을 찾는다.
2. error response, validation, config, logging, test pattern을 각각 어떻게 정리할지 작성한다.
3. 바로 구현할 것과 나중에 할 것을 나눈다.
4. 다음 구현 티켓 후보를 정리한다.

## 완료 기준

- [ ] Backend Core가 필요한 이유를 설명했다.
- [ ] 공통 에러 응답 후보를 작성했다.
- [ ] 요청 validation 위치를 정리했다.
- [ ] config/logging/test pattern의 필요성을 정리했다.
- [ ] Sprint 2 첫 구현 작업 후보를 구체화했다.
