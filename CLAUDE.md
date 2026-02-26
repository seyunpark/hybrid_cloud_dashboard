# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 프로젝트 개요

Hybrid Cloud Dashboard — AI 기반 하이브리드 환경 통합 모니터링 및 지능형 배포 시스템. 로컬 Docker 환경과 여러 Kubernetes 클러스터(AWS EKS, Azure AKS, On-premise)를 단일 대시보드에서 통합 모니터링하고, LLM 기반으로 Docker 컨테이너를 K8s에 지능적으로 배포한다.

**현재 상태:** 초기 단계. 설계 문서와 프로젝트 구조만 정의되어 있으며, 실제 구현 코드는 아직 작성되지 않음.

## 기술 스택

- **Backend:** Go 1.21+, Gin, docker/docker SDK, k8s.io/client-go, Gorilla WebSocket
- **Frontend:** React 18, TypeScript, Vite, TailwindCSS, React Query, Recharts
- **AI:** OpenAI GPT-4 / Claude API (Few-shot Learning, Prompt Engineering)
- **DB:** SQLite (배포 이력 저장)
- **실시간:** WebSocket (메트릭 스트리밍, 로그, 배포 상태)

## 빌드 및 실행 명령

### Docker Compose (전체 시스템)
```bash
cp configs/config.example.yaml configs/config.yaml  # 최초 1회
docker-compose up -d                                 # 실행
# Frontend: http://localhost:3000, Backend: http://localhost:8080
```

### Backend (Go)
```bash
cd backend
go mod download                    # 의존성 설치
go run cmd/server/main.go          # 실행
air                                # Hot reload 실행 (air 설치 필요)
go test ./...                      # 전체 테스트
go test ./internal/ai/...          # 특정 패키지 테스트
go test -cover ./...               # 커버리지 포함 테스트
```

### Frontend (React)
```bash
cd frontend
npm install          # 의존성 설치
npm run dev          # 개발 서버 (http://localhost:5173)
npm run build        # 프로덕션 빌드 (dist/)
npm run test         # 단위 테스트
npm run test:e2e     # E2E 테스트 (Playwright)
```

### DB 마이그레이션
```bash
cd backend
go run cmd/server/main.go migrate           # 초기화
go run cmd/server/main.go migrate:up        # 마이그레이션 실행
go run cmd/server/main.go migrate:down      # 롤백
```

### 린트
- Go: `gofmt`로 포맷팅, `golangci-lint` 통과 필수
- TypeScript: Prettier 포맷팅, ESLint 규칙 준수

## 아키텍처

### 3-tier 구조
```
Frontend (React) → REST API / WebSocket → Backend (Go) → Docker API / K8s API / LLM API
```

### Backend 핵심 모듈 (`backend/internal/`)
- `api/` — HTTP 핸들러, WebSocket 핸들러, 라우팅, 미들웨어
- `docker/` — Docker Engine API 클라이언트, 컨테이너 관리, 통계 수집, 로그 스트리밍
- `kubernetes/` — K8s API 클라이언트, 클러스터 관리, 배포 실행, 리소스 조회
- `ai/` — LLM 클라이언트, 프롬프트 빌더, Manifest 생성, 리소스 예측, Few-shot 관리
- `data/` — SQLite 기반 배포 이력 저장, 유사 배포 검색 (Few-shot용)
- `registry/` — Container Registry 이미지 푸시
- `metrics/` — Goroutine 기반 메트릭 수집 및 WebSocket 브로드캐스트

### AI Manifest Generator 플로우
1. Docker 컨테이너 정보 추출 (이미지, 포트, 환경변수, 리소스 사용량)
2. SQLite에서 유사 배포 이력 검색 (이미지/서비스 타입/언어 기준)
3. Few-shot 프롬프트 구성 (System Prompt + 유사 사례 3-5개 + 현재 요청)
4. LLM API 호출 (temperature: 0.3으로 일관된 출력)
5. 응답 파싱 → YAML 검증 → 보안 정책 검증
6. 사용자 리뷰 후 배포 실행 → 이력 저장

### API 엔드포인트 구조
- `GET /api/docker/containers` — Docker 컨테이너 관리
- `GET /api/k8s/:cluster/pods|deployments|services` — K8s 리소스 조회
- `POST /api/deploy/docker-to-k8s` — AI 기반 배포 (Manifest 생성)
- `POST /api/deploy/:deploy_id/execute` — 배포 승인 및 실행
- `WS /ws/docker/stats`, `WS /ws/k8s/:cluster/metrics` — 실시간 스트리밍
- `GET /health`, `GET /ready` — 헬스 체크

## 설정

- `configs/config.yaml` — 메인 설정 파일 (서버, AI, Docker, K8s 클러스터, Registry, DB, 로깅, 메트릭)
- `.env` — 환경변수 (API 키 등). `.env.example` 참조
- AI provider는 `openai`, `claude`, `azure-openai` 지원

## 커밋 컨벤션

Conventional Commits 형식: `<type>(<scope>): <subject>`
- Type: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`
- Scope: `api`, `ui`, `ai`, `docker`, `k8s`
- 브랜치: `feature/`, `fix/`, `docs/`, `refactor/`, `test/`, `chore/`

## 코딩 컨벤션

- Go: 공개 함수에 GoDoc 주석, table-driven 테스트 패턴 사용
- React: 함수형 컴포넌트, prop 타입을 interface로 명확히 정의
- LLM API 장애 시 템플릿 기반 fallback 패턴 적용
- 모든 K8s Manifest에 리소스 requests/limits, probe, SecurityContext 필수

## 주요 설계 결정 (ADR)

- ADR-001: 백엔드 언어로 Go 선택 (높은 동시성 처리, Docker/K8s SDK 네이티브 지원)
- ADR-003: AI Manifest 생성 시 Few-shot Learning + Chain-of-Thought 프롬프트 전략
- ADR-004: SQLite 사용 (배포 이력, 외부 의존성 최소화)
- ADR-005: WebSocket으로 실시간 메트릭/로그 스트리밍
