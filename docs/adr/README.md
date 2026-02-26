# Architecture Decision Records (ADR)

이 디렉토리는 프로젝트의 주요 아키텍처 결정사항을 기록합니다.

## ADR 목록

- [ADR-001: Go를 백엔드 언어로 선택](./ADR-001-backend-language-go.md)
- [ADR-002: React를 프론트엔드 프레임워크로 선택](./ADR-002-frontend-framework-react.md)
- [ADR-003: LLM 기반 Manifest 생성 방식](./ADR-003-ai-manifest-generation.md)
- [ADR-004: SQLite를 배포 이력 저장소로 선택](./ADR-004-database-sqlite.md)
- [ADR-005: WebSocket 실시간 통신 사용](./ADR-005-websocket-realtime.md)

## ADR 템플릿

새로운 ADR을 작성할 때는 다음 구조를 따릅니다:

```markdown
# ADR-XXX: [제목]

## 상태
[제안됨 | 승인됨 | 거부됨 | 대체됨 | 폐기됨]

## 컨텍스트
[결정이 필요한 배경과 문제 상황]

## 결정
[내린 결정]

## 결과
[이 결정으로 인한 긍정적/부정적 결과]

## 대안
[고려했던 다른 옵션들과 선택하지 않은 이유]
```
