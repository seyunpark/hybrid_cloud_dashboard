# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 프로젝트 개요

Hybrid Cloud Dashboard — AI 기반 하이브리드 환경 통합 모니터링 및 지능형 배포 시스템. 로컬 Docker 환경과 여러 Kubernetes 클러스터(AWS EKS, Azure AKS, On-premise)를 단일 대시보드에서 통합 모니터링하고, LLM 기반으로 Docker 컨테이너를 K8s에 지능적으로 배포한다.

**현재 상태:** 백엔드(Go)와 프론트엔드(React) 핵심 기능 구현 완료. Docker/K8s 모니터링, AI 매니페스트 생성(단일+스택), 배포 라이프사이클(배포/언디플로이/리디플로이/삭제), 통합 배포 히스토리, 설정 관리, 실시간 WebSocket 스트리밍이 동작하는 상태.

## 기술 스택

- **Backend:** Go 1.21+, Gin, docker/docker SDK, k8s.io/client-go, Gorilla WebSocket
- **Frontend:** React 19, TypeScript, Vite 7, TailwindCSS v4, React Query v5, Recharts 3, React Router v7
- **AI:** OpenAI GPT-4 / Claude API / Google Gemini API (Few-shot Learning, Prompt Engineering)
- **DB:** SQLite (배포 이력, 설정, 클러스터 등록, 스택 배포 저장)
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
go run cmd/server/main.go          # 실행 (CONFIG_PATH 환경변수 필요)
air                                # Hot reload 실행 (air 설치 필요)
go build ./...                     # 전체 빌드 확인
go vet ./...                       # 정적 분석
go test ./...                      # 전체 테스트
go test ./internal/ai/...          # 특정 패키지 테스트
go test -cover ./...               # 커버리지 포함 테스트
```

### Frontend (React)
```bash
cd frontend
npm install          # 의존성 설치
npm run dev          # 개발 서버 (http://localhost:5173)
npm run build        # 프로덕션 빌드 (tsc -b && vite build → dist/)
npm run test         # 단위 테스트 (Vitest)
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
- Go: `gofmt` 포맷팅, `golangci-lint` 통과 필수
- TypeScript: ESLint (`npm run lint`), Vite 빌드 시 `tsc -b`로 타입 체크

## 아키텍처

### 3-tier 구조
```
Frontend (React) → REST API / WebSocket → Backend (Go) → Docker API / K8s API / LLM API
```

### Backend (`backend/`)

Go 모듈: `github.com/seyunpark/hybrid_cloud_dashboard`

- `cmd/server/main.go` — 진입점: 설정 로드 → 서비스 초기화 → Gin 서버 시작 + graceful shutdown
- `internal/config/` — `Config` 구조체 + YAML 로드 (`gopkg.in/yaml.v3`) + 환경변수 오버라이드
- `internal/api/` — Gin 라우터(`router.go`), 미들웨어(`middleware.go`), 핸들러(`handlers_docker.go`, `handlers_k8s.go`, `handlers_deploy.go`, `handlers_stack_deploy.go`, `handlers_config.go`), WebSocket(`websocket.go`)
- `internal/docker/` — `docker.Service` 인터페이스: ListContainers, GetContainer, RestartContainer, StopContainer, DeleteContainer
- `internal/kubernetes/` — `kubernetes.Service` 인터페이스: ListClusters, ListNamespaces, ListPods, ListDeployments, ListServices, ScaleDeployment, RestartPod
- `internal/ai/` — `ai.Service` 인터페이스: GenerateManifest, RefineManifest, GenerateStackManifest, RefineStackManifest, UpdateConfig, GetConfig, ListModels
- `internal/data/` — `data.Store` 인터페이스: Init, Close, SaveDeployment, GetDeployment, GetDeployHistory, FindSimilar, UpdateDeploymentStatus, DeleteDeploymentRecord, SaveSetting, GetSetting, GetAllSettings, DeleteSetting, SaveRegisteredCluster, DeleteRegisteredCluster, GetRegisteredClusters, SaveStackDeploy, GetStackDeploy, UpdateStackDeploy, ListStackDeploys, DeleteStackDeploy, ListUnifiedHistory, CleanupOldRecords
- `internal/registry/` — `registry.Service` 인터페이스: PushImage, TagImage
- `internal/metrics/` — `metrics.Collector`: Goroutine 기반 Start/Stop 주기적 수집
- `pkg/models/` — 전체 API 데이터 모델 (Container, Pod, Deployment, Deploy*, Stack*, UnifiedDeployItem, PaginatedResponse 등 40+ 구조체)

서비스 간 의존성 주입(DI): `api.NewServer(cfg, dockerSvc, k8sSvc, aiSvc, dataStore, registrySvc, metricsColl)`

### Frontend (`frontend/`)

- `src/api/types.ts` — 백엔드 `pkg/models/models.go` 대응 TypeScript 타입 (40+ 인터페이스)
- `src/api/client.ts` — Axios 인스턴스 + `dockerApi`, `k8sApi`, `deployApi`, `stackDeployApi`, `configApi`, `healthApi` 함수 모듈
- `src/hooks/` — `useDockerContainers`, `useK8sClusters`, `useClusterManagement`, `useStackDeploy`, `useUnifiedHistory`, `useWebSocket`
- `src/components/layout/Layout.tsx` — 사이드바 + 헤더 + `<Outlet />` 레이아웃
- `src/components/dashboard/` — `Dashboard`, `ContainerCard`, `ClusterOverview`
- `src/components/deploy/` — `DeployModal`, `StackDeployModal`, `DeployProgress`, `StackDeployProgress`, `ManifestPreview`, `StackManifestPreview`
- `src/components/common/` — `LoadingSpinner`, `StatusBadge`, `MetricChart`, `ErrorBoundary`, `Toast`
- `src/pages/` — `DeployPage`, `HistoryPage`, `ContainerDetailPage`, `ClusterDetailPage`, `StackDeployDetailPage`, `SettingsPage`
- `src/utils/formatters.ts` — `formatBytes`, `formatCpuPercent`, `formatDate`, `formatDateTime`, `formatRelativeTime`, `formatNetworkRate`
- `src/App.tsx` — React Router 라우팅 + QueryClientProvider + Toast/ErrorBoundary

라우팅:
- `/` — Dashboard (컨테이너/클러스터 개요)
- `/container/:id` — ContainerDetailPage (Docker 컨테이너 상세 + 실시간 메트릭)
- `/cluster/:name` — ClusterDetailPage (K8s 클러스터 상세 — Pod, Deployment, Service)
- `/deploy` — DeployPage (활성 스택 배포 + 최근 배포 목록)
- `/deploy/:deployId` — StackDeployDetailPage (스택 배포 상세 라이프사이클)
- `/history` — HistoryPage (통합 배포 히스토리, 페이지네이션)
- `/settings` — SettingsPage (AI 설정, 클러스터 등록/해제)

경로 별칭: `@/*` → `src/*` (tsconfig.app.json + vite.config.ts)

Vite 프록시: 개발 시 `/api` → `http://localhost:8080`, `/ws` → `ws://localhost:8080`

### API 엔드포인트 구조 (42 REST + 5 WebSocket)

**Docker** (5):
- `GET /api/docker/containers` — 컨테이너 목록
- `GET /api/docker/containers/:id` — 컨테이너 상세
- `POST /api/docker/containers/:id/restart` — 재시작
- `POST /api/docker/containers/:id/stop` — 중지
- `DELETE /api/docker/containers/:id` — 삭제

**Kubernetes** (7):
- `GET /api/k8s/clusters` — 클러스터 목록
- `GET /api/k8s/:cluster/namespaces` — 네임스페이스 목록
- `GET /api/k8s/:cluster/pods` — Pod 목록
- `GET /api/k8s/:cluster/deployments` — Deployment 목록
- `GET /api/k8s/:cluster/services` — Service 목록
- `POST /api/k8s/:cluster/deployments/:ns/:name/scale` — 스케일링
- `POST /api/k8s/:cluster/pods/:ns/:name/restart` — Pod 재시작

**단일 배포** (9):
- `POST /api/deploy/docker-to-k8s` — AI 매니페스트 생성 요청
- `POST /api/deploy/:deploy_id/execute` — 배포 실행
- `POST /api/deploy/:deploy_id/refine` — 매니페스트 수정 요청
- `POST /api/deploy/:deploy_id/undeploy` — K8s 리소스 삭제
- `POST /api/deploy/:deploy_id/redeploy` — 재배포
- `DELETE /api/deploy/:deploy_id` — 배포 기록 삭제
- `GET /api/deploy/:deploy_id/status` — 배포 상태 조회
- `GET /api/deploy/history` — 배포 이력
- `GET /api/deploy/unified-history` — 통합 배포 이력 (페이지네이션)

**스택 배포** (11):
- `GET /api/deploy/stack/` — 활성 스택 배포 목록
- `GET /api/deploy/stack/:deploy_id` — 스택 배포 상세
- `GET /api/deploy/stack/:deploy_id/status` — 스택 배포 상태
- `POST /api/deploy/stack/` — 스택 배포 생성 (AI 매니페스트 생성)
- `POST /api/deploy/stack/:deploy_id/refine` — 피드백 기반 수정
- `POST /api/deploy/stack/:deploy_id/regenerate` — 매니페스트 재생성
- `POST /api/deploy/stack/:deploy_id/reopen` — 완료된 배포 재편집
- `POST /api/deploy/stack/:deploy_id/execute` — 스택 배포 실행
- `POST /api/deploy/stack/:deploy_id/undeploy` — 스택 언디플로이
- `POST /api/deploy/stack/:deploy_id/redeploy` — 스택 재배포
- `DELETE /api/deploy/stack/:deploy_id` — 스택 배포 삭제 (soft-delete)

**설정** (7):
- `GET /api/config/clusters` — 클러스터 설정 조회
- `GET /api/config/ai` — AI 설정 조회
- `PUT /api/config/ai` — AI 설정 변경
- `GET /api/config/ai/models` — AI 모델 목록 조회
- `GET /api/config/kubecontexts` — kubeconfig 컨텍스트 목록
- `POST /api/config/clusters` — 클러스터 등록
- `DELETE /api/config/clusters/:name` — 클러스터 해제

**헬스체크** (2):
- `GET /health` — 서버 상태
- `GET /ready` — 준비 상태

**WebSocket** (5):
- `/ws/docker/stats` — Docker 컨테이너 메트릭 스트리밍
- `/ws/k8s/:cluster/metrics` — K8s 클러스터 메트릭 스트리밍
- `/ws/docker/:container_id/logs` — Docker 로그 스트리밍
- `/ws/k8s/:cluster/:namespace/:pod/logs` — K8s Pod 로그 스트리밍
- `/ws/deploy/:deploy_id/status` — 배포 상태 스트리밍

### AI Manifest Generator 플로우
1. Docker 컨테이너 정보 추출 (이미지, 포트, 환경변수, 리소스 사용량)
2. SQLite에서 유사 배포 이력 검색 (이미지/서비스 타입/언어 기준)
3. Few-shot 프롬프트 구성 (System Prompt + 유사 사례 3-5개 + 현재 요청)
4. LLM API 호출 (temperature: 0.3, 3회 재시도 + 지수 백오프)
5. 응답 파싱 (코드 펜스 제거 → JSON 추출 → 잘린 JSON 복구) → 검증
6. 사용자 리뷰 후 배포 실행 → 이력 저장
7. 스택 배포: 토폴로지 분석 → 서비스 간 연결 감지 → 배포 순서 결정 → 개별 서비스 매니페스트 생성

## 설정

- `configs/config.yaml` — 메인 설정 (서버, AI, Docker, K8s 클러스터, Registry, DB, 로깅, 메트릭, WebSocket, 보안, 기능 플래그, 제한). `configs/config.example.yaml` 참조
- `.env.example` — 환경변수 (API 키 등)
- `frontend/.env` — `VITE_API_URL`, `VITE_WS_URL`
- AI provider: `openai`, `claude`, `azure-openai`, `gemini` 지원
- 백엔드 환경변수 오버라이드: `PORT`, `OPENAI_API_KEY`, `CLAUDE_API_KEY`, `GEMINI_API_KEY`, `LOG_LEVEL`, `DATABASE_PATH`, `DOCKER_SOCKET`

## 커밋 컨벤션

Conventional Commits 형식: `<type>(<scope>): <subject>`
- Type: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`
- Scope: `api`, `ui`, `ai`, `docker`, `k8s`
- 브랜치: `feature/`, `fix/`, `docs/`, `refactor/`, `test/`, `chore/`

## 코딩 컨벤션

- Go: 공개 함수에 GoDoc 주석, table-driven 테스트 패턴, 서비스는 인터페이스로 정의하고 DI로 주입
- React: 함수형 컴포넌트, prop 타입을 interface로 정의, `@/` 경로 별칭 사용
- API 타입 동기화: 백엔드 `pkg/models/models.go` 변경 시 프론트엔드 `src/api/types.ts`도 동기화
- LLM API 장애 시 템플릿 기반 fallback 패턴 적용 (3회 재시도 + 지수 백오프 후)
- 모든 K8s Manifest에 리소스 requests/limits, probe, SecurityContext 필수

## 주요 설계 결정 (ADR)

- ADR-001: 백엔드 언어로 Go 선택 (높은 동시성 처리, Docker/K8s SDK 네이티브 지원)
- ADR-003: AI Manifest 생성 시 Few-shot Learning + Chain-of-Thought 프롬프트 전략
- ADR-004: SQLite 사용 (배포 이력, 설정, 클러스터 등록, 스택 배포 — 외부 의존성 최소화)
- ADR-005: WebSocket으로 실시간 메트릭/로그 스트리밍
