# Vineyard

AI 기반 하이브리드 환경 통합 모니터링 및 지능형 배포 시스템

---

## 실행 가이드

### 사전 요구사항

| 구분 | Docker Compose 실행 | 로컬 개발 실행 |
|------|---------------------|---------------|
| **공통** | Docker Desktop | Go 1.25+, Node.js 22+ |
| **K8s 연동** | kubeconfig 파일 | kubeconfig 파일 |
| **AI 기능** | OpenAI / Claude / Gemini API Key (선택) | 동일 |

### 방법 1. Docker Compose (권장)

OS에 관계없이 Docker Desktop만 설치하면 실행 가능합니다.

```bash
# 1. 저장소 클론
git clone <repository-url>
cd AI_Project

# 2. 설정 파일 생성
cp configs/config.example.yaml configs/config.yaml

# 3. 빌드 및 실행
# macOS / Linux
./build.sh

# Windows (PowerShell)
docker compose build --no-cache
docker compose up -d
```

실행 후 접속:
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080

종료: `docker compose down`

> **참고:** AI 기능을 사용하려면 Settings 페이지에서 API Key를 등록하거나, 환경변수로 전달하세요:
> ```bash
> GEMINI_API_KEY=your-key docker compose up -d
> ```

### 방법 2. 로컬 개발 실행

#### macOS / Linux

```bash
# 1. 설정 파일 생성
cp configs/config.example.yaml configs/config.yaml

# 2. 의존성 설치
cd backend && go mod download && cd ..
cd frontend && npm install && cd ..

# 3. 실행
./run.sh
```

#### Windows (Git Bash 또는 WSL2)

```bash
# 1. 설정 파일 생성
cp configs/config.example.yaml configs/config.yaml

# 2. 의존성 설치
cd backend && go mod download && cd ..
cd frontend && npm install && cd ..

# 3. 실행
./run_win.sh
```

실행 후 접속:
- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8888 (config.yaml에서 변경 가능)

### kubeconfig 파일 위치

| OS | 경로 |
|----|------|
| macOS / Linux | `~/.kube/config` |
| Windows | `%USERPROFILE%\.kube\config` (예: `C:\Users\<username>\.kube\config`) |

Docker Compose 실행 시 자동으로 마운트됩니다. 로컬 실행 시에는 기본 경로에 있으면 자동 인식됩니다.

---

## 개요

로컬 Docker 환경과 여러 Kubernetes 클러스터(AWS EKS, Azure AKS, GKE, On-premise)를 단일 대시보드에서 통합 모니터링하고, AI 기반으로 Docker 컨테이너를 K8s에 지능적으로 배포하는 시스템입니다.

## 주요 기능

### 1. 통합 모니터링
- 로컬 Docker 컨테이너 실시간 모니터링
- 여러 Kubernetes 클러스터 통합 모니터링
- 실시간 리소스 사용량 (CPU, Memory, Network)
- Pod/Deployment 상태 추적

### 2. AI 기반 지능형 배포
- LLM을 활용한 Kubernetes Manifest 자동 생성 (단일 컨테이너 + 멀티 컨테이너 스택)
- 과거 배포 패턴 학습 및 최적 설정 추천 (Few-shot Learning)
- 리소스 할당 자동 최적화
- 보안 설정 자동 적용
- 서비스 간 토폴로지 분석 및 배포 순서 결정

### 3. 배포 라이프사이클
- Docker → Kubernetes 자동 배포
- 배포 상태 실시간 모니터링 (WebSocket)
- 배포/언디플로이/리디플로이/삭제
- 통합 배포 히스토리 (단일+스택, 페이지네이션)

## 기술 스택

### Backend
- Go 1.25, Gin, docker/docker SDK, k8s.io/client-go
- OpenAI / Claude / Gemini API
- SQLite, Gorilla WebSocket

### Frontend
- React 19, TypeScript, Vite 7, TailwindCSS v4
- React Query v5, React Router v7, Recharts 3

## 프로젝트 구조

```
.
├── backend/                # Go 백엔드 (~9,100 LOC)
│   ├── cmd/server/         # 진입점
│   ├── internal/
│   │   ├── api/            # REST (42) + WebSocket (5) 핸들러
│   │   ├── ai/             # AI 엔진 (OpenAI/Claude/Gemini)
│   │   ├── docker/         # Docker 관리
│   │   ├── kubernetes/     # K8s 관리
│   │   ├── data/           # SQLite 데이터 레이어
│   │   └── config/         # 설정 관리
│   └── pkg/models/         # 공유 데이터 모델
├── frontend/               # React 프론트엔드 (~5,800 LOC)
│   └── src/
│       ├── components/     # UI 컴포넌트
│       ├── hooks/          # React Query/WebSocket 훅
│       ├── pages/          # 페이지 컴포넌트
│       └── api/            # API 클라이언트 + 타입
├── configs/                # 설정 파일
├── docs/                   # 문서
├── docker-compose.yml      # Docker Compose 실행
├── build.sh                # 빌드 + 실행 스크립트 (macOS/Linux)
├── run.sh                  # 로컬 개발 실행 (macOS/Linux)
└── run_win.sh              # 로컬 개발 실행 (Windows Git Bash/WSL2)
```

## 문서

- [시스템 아키텍처](./docs/ARCHITECTURE.md)
- [API 명세](./docs/API_SPEC.md)
- [AI Manifest Generator 가이드](./docs/AI_MANIFEST_GENERATOR.md)
- [개발 환경 설정](./docs/SETUP.md)
- [수행 과제 결과 보고서](./docs/RESULT_REPORT.md)

## 라이선스

MIT License
