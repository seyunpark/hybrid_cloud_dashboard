# Hybrid Cloud Dashboard

AI 기반 하이브리드 환경 통합 모니터링 및 지능형 배포 시스템

## 개요

로컬 Docker 환경과 여러 Kubernetes 클러스터(AWS EKS, Azure AKS, On-premise)를 단일 대시보드에서 통합 모니터링하고, AI 기반으로 Docker 컨테이너를 K8s에 지능적으로 배포하는 시스템입니다.

## 주요 기능

### 1. 통합 모니터링
- 로컬 Docker 컨테이너 실시간 모니터링
- 여러 Kubernetes 클러스터 통합 모니터링
- 실시간 리소스 사용량 (CPU, Memory, Network)
- Pod/Deployment 상태 추적

### 2. AI 기반 지능형 배포
- LLM을 활용한 Kubernetes Manifest 자동 생성
- 과거 배포 패턴 학습 및 최적 설정 추천
- 리소스 할당 자동 최적화
- 보안 설정 자동 적용

### 3. 원클릭 배포
- Docker → Kubernetes 자동 배포
- Container Registry 푸시 자동화
- 배포 상태 실시간 모니터링

## 기술 스택

### Backend
- Go 1.21+
- Gin (Web Framework)
- docker/docker (Docker SDK)
- k8s.io/client-go (Kubernetes SDK)
- OpenAI/Claude API (AI 기능)

### Frontend
- React 18
- TypeScript
- TailwindCSS
- React Query
- Recharts

### AI
- GPT-4 / Claude 3.5 Sonnet
- Few-shot Learning
- Prompt Engineering

## 프로젝트 구조

```
.
├── backend/                # Go 백엔드
│   ├── cmd/
│   │   └── server/
│   ├── internal/
│   │   ├── api/           # API 핸들러
│   │   ├── docker/        # Docker 관리
│   │   ├── kubernetes/    # K8s 관리
│   │   ├── ai/            # AI 엔진
│   │   └── config/        # 설정 관리
│   └── pkg/
├── frontend/              # React 프론트엔드
│   ├── src/
│   │   ├── components/
│   │   ├── hooks/
│   │   ├── api/
│   │   └── utils/
│   └── public/
├── docs/                  # 문서
│   ├── adr/              # Architecture Decision Records
│   ├── api/              # API 명세
│   └── guides/           # 가이드 문서
├── deployments/          # 배포 설정
│   ├── docker-compose.yml
│   └── kubernetes/
├── configs/              # 설정 파일
└── scripts/              # 유틸리티 스크립트
```

## 빠른 시작

### 사전 요구사항

- Docker 20.10+
- Docker Compose 2.0+
- Go 1.21+ (개발 시)
- Node.js 18+ (개발 시)
- Kubernetes 클러스터 접근 권한
- OpenAI 또는 Claude API Key

### 설치 및 실행

1. 저장소 클론
```bash
git clone <repository-url>
cd AI_Project
```

2. 설정 파일 생성
```bash
cp configs/config.example.yaml configs/config.yaml
# config.yaml 파일 편집 (API Key, 클러스터 정보 등)
```

3. Docker Compose로 실행
```bash
docker-compose up -d
```

4. 브라우저에서 접속
```
http://localhost:3000
```

## 개발 가이드

자세한 개발 가이드는 [CONTRIBUTING.md](./CONTRIBUTING.md)를 참조하세요.

## 문서

- [시스템 아키텍처](./docs/ARCHITECTURE.md)
- [API 명세](./docs/API_SPEC.md)
- [AI Manifest Generator 가이드](./docs/AI_MANIFEST_GENERATOR.md)
- [개발 환경 설정](./docs/SETUP.md)
- [Architecture Decision Records](./docs/adr/)

## 라이선스

MIT License

## 기여

기여는 언제나 환영합니다! [CONTRIBUTING.md](./CONTRIBUTING.md)를 참조해주세요.

## 문의

프로젝트 관련 문의사항은 Issue를 통해 남겨주세요.
