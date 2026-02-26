# 프로젝트 구조

```
AI_Project/
├── README.md                       # 프로젝트 소개
├── CONTRIBUTING.md                 # 기여 가이드
├── PROJECT_STRUCTURE.md            # 이 파일
├── .gitignore                      # Git 무시 파일
├── .env.example                    # 환경변수 예시
├── docker-compose.yml              # Docker Compose 설정
│
├── backend/                        # Go 백엔드
│   ├── cmd/
│   │   └── server/
│   │       └── main.go            # 애플리케이션 진입점
│   │
│   ├── internal/                  # 내부 패키지
│   │   ├── api/                   # API 레이어
│   │   │   ├── router.go         # 라우팅 설정
│   │   │   ├── handlers.go       # HTTP 핸들러
│   │   │   ├── websocket.go      # WebSocket 핸들러
│   │   │   └── middleware.go     # 미들웨어
│   │   │
│   │   ├── docker/                # Docker 관리
│   │   │   ├── client.go         # Docker API 클라이언트
│   │   │   ├── manager.go        # 컨테이너 관리
│   │   │   ├── stats.go          # 통계 수집
│   │   │   └── logs.go           # 로그 스트리밍
│   │   │
│   │   ├── kubernetes/            # Kubernetes 관리
│   │   │   ├── client.go         # K8s API 클라이언트
│   │   │   ├── manager.go        # 클러스터 관리
│   │   │   ├── deployer.go       # 배포 실행
│   │   │   └── resources.go      # 리소스 조회
│   │   │
│   │   ├── ai/                    # AI 엔진
│   │   │   ├── client.go         # LLM API 클라이언트
│   │   │   ├── prompt_builder.go # 프롬프트 생성
│   │   │   ├── manifest_generator.go # Manifest 생성
│   │   │   ├── resource_predictor.go # 리소스 예측
│   │   │   └── few_shot.go       # Few-shot 예시 관리
│   │   │
│   │   ├── registry/              # Container Registry
│   │   │   └── pusher.go         # 이미지 푸시
│   │   │
│   │   ├── data/                  # 데이터 레이어
│   │   │   ├── deployment_store.go # 배포 이력 저장
│   │   │   ├── similarity_search.go # 유사 배포 검색
│   │   │   └── models.go         # 데이터 모델
│   │   │
│   │   ├── config/                # 설정 관리
│   │   │   └── loader.go         # 설정 로드
│   │   │
│   │   └── metrics/               # 메트릭 수집
│   │       ├── collector.go      # 메트릭 수집기
│   │       └── publisher.go      # 메트릭 발행
│   │
│   ├── pkg/                       # 공개 패키지
│   │   ├── models/               # 공통 데이터 모델
│   │   └── utils/                # 유틸리티
│   │
│   ├── go.mod                    # Go 모듈 정의
│   ├── go.sum                    # 의존성 체크섬
│   ├── Dockerfile                # Docker 이미지 빌드
│   └── .air.toml                 # Air 설정 (Hot reload)
│
├── frontend/                      # React 프론트엔드
│   ├── src/
│   │   ├── components/           # React 컴포넌트
│   │   │   ├── dashboard/       # 대시보드
│   │   │   │   ├── Dashboard.tsx
│   │   │   │   ├── ContainerCard.tsx
│   │   │   │   └── ClusterOverview.tsx
│   │   │   │
│   │   │   ├── deploy/          # 배포 관련
│   │   │   │   ├── DeployModal.tsx
│   │   │   │   ├── DeployProgress.tsx
│   │   │   │   └── ManifestPreview.tsx
│   │   │   │
│   │   │   ├── logs/            # 로그 뷰어
│   │   │   │   └── LogViewer.tsx
│   │   │   │
│   │   │   └── common/          # 공통 컴포넌트
│   │   │       ├── MetricChart.tsx
│   │   │       ├── StatusBadge.tsx
│   │   │       └── LoadingSpinner.tsx
│   │   │
│   │   ├── hooks/                # 커스텀 Hooks
│   │   │   ├── useDockerContainers.ts
│   │   │   ├── useK8sClusters.ts
│   │   │   └── useWebSocket.ts
│   │   │
│   │   ├── api/                  # API 클라이언트
│   │   │   ├── client.ts        # Axios 클라이언트
│   │   │   └── types.ts         # 타입 정의
│   │   │
│   │   ├── utils/                # 유틸리티
│   │   │   └── formatters.ts   # 포맷 함수
│   │   │
│   │   ├── App.tsx              # 루트 컴포넌트
│   │   ├── main.tsx             # 진입점
│   │   └── index.css            # 글로벌 스타일
│   │
│   ├── public/                   # 정적 파일
│   ├── package.json             # npm 패키지 정의
│   ├── tsconfig.json            # TypeScript 설정
│   ├── vite.config.ts           # Vite 설정
│   ├── tailwind.config.js       # TailwindCSS 설정
│   ├── Dockerfile               # Docker 이미지 빌드
│   └── .eslintrc.json           # ESLint 설정
│
├── docs/                          # 문서
│   ├── ARCHITECTURE.md           # 시스템 아키텍처
│   ├── API_SPEC.md               # API 명세
│   ├── AI_MANIFEST_GENERATOR.md  # AI 기능 가이드
│   ├── SETUP.md                  # 개발 환경 설정
│   │
│   └── adr/                      # Architecture Decision Records
│       ├── README.md
│       ├── ADR-001-backend-language-go.md
│       ├── ADR-002-frontend-framework-react.md
│       ├── ADR-003-ai-manifest-generation.md
│       ├── ADR-004-database-sqlite.md
│       └── ADR-005-websocket-realtime.md
│
├── configs/                       # 설정 파일
│   ├── config.example.yaml       # 설정 예시
│   └── config.yaml               # 실제 설정 (gitignore)
│
├── deployments/                   # 배포 관련
│   ├── docker-compose.yml        # Docker Compose
│   │
│   └── kubernetes/               # K8s 매니페스트
│       ├── deployment.yaml
│       ├── service.yaml
│       ├── configmap.yaml
│       └── ingress.yaml
│
├── scripts/                       # 유틸리티 스크립트
│   ├── setup.sh                  # 초기 설정
│   ├── build.sh                  # 빌드 스크립트
│   └── deploy.sh                 # 배포 스크립트
│
├── data/                          # 데이터 (gitignore)
│   └── deployments.db            # SQLite 데이터베이스
│
└── .github/                       # GitHub 설정
    ├── workflows/                # GitHub Actions
    │   ├── ci.yml               # CI 파이프라인
    │   └── release.yml          # 릴리스 자동화
    │
    ├── ISSUE_TEMPLATE/           # Issue 템플릿
    │   ├── bug_report.md
    │   └── feature_request.md
    │
    └── PULL_REQUEST_TEMPLATE.md  # PR 템플릿
```

## 주요 디렉토리 설명

### `/backend`
Go로 작성된 백엔드 애플리케이션입니다. Docker/Kubernetes API 통합, AI 기반 Manifest 생성, WebSocket 실시간 통신을 담당합니다.

### `/frontend`
React + TypeScript로 작성된 프론트엔드 애플리케이션입니다. 대시보드 UI, 실시간 모니터링, 배포 관리 화면을 제공합니다.

### `/docs`
프로젝트 문서가 위치합니다. 아키텍처, API 명세, 개발 가이드, ADR 등이 포함됩니다.

### `/configs`
애플리케이션 설정 파일입니다. YAML 형식으로 클러스터 정보, AI 설정 등을 관리합니다.

### `/deployments`
배포 관련 파일들입니다. Docker Compose와 Kubernetes Manifest가 포함됩니다.

### `/scripts`
개발 및 배포를 위한 유틸리티 스크립트입니다.

### `/data`
SQLite 데이터베이스 등 런타임 데이터가 저장됩니다. Git에서 제외됩니다.

## 파일 생성 시작 가이드

### 1. 백엔드 시작

```bash
mkdir -p backend/{cmd/server,internal/{api,docker,kubernetes,ai,registry,data,config,metrics},pkg/{models,utils}}
cd backend
go mod init github.com/your-org/hybrid-dashboard
```

### 2. 프론트엔드 시작

```bash
npm create vite@latest frontend -- --template react-ts
cd frontend
npm install
npm install -D tailwindcss postcss autoprefixer
npx tailwindcss init -p
```

### 3. 문서 확인

각 문서를 순서대로 읽으면 프로젝트를 이해할 수 있습니다:

1. README.md - 프로젝트 개요
2. docs/ARCHITECTURE.md - 시스템 구조
3. docs/SETUP.md - 개발 환경 설정
4. docs/API_SPEC.md - API 명세
5. docs/AI_MANIFEST_GENERATOR.md - AI 기능 상세

## 다음 단계

1. [개발 환경 설정](./docs/SETUP.md) 가이드를 따라 로컬 환경 구성
2. 백엔드 기본 구조 코드 작성
3. 프론트엔드 기본 UI 구현
4. Docker/Kubernetes API 통합
5. AI 엔진 구현
6. 테스트 작성
7. 문서 업데이트
