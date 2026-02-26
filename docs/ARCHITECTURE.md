# 시스템 아키텍처

## 전체 시스템 구조

```
┌─────────────────────────────────────────────────┐
│                   사용자                         │
│              (웹 브라우저)                        │
└────────────────────┬────────────────────────────┘
                     │ HTTPS
                     ↓
┌─────────────────────────────────────────────────┐
│              Frontend (React)                   │
│  - TypeScript                                   │
│  - TailwindCSS                                  │
│  - React Query (상태 관리)                       │
│  - Recharts (차트)                              │
│  - WebSocket (실시간 통신)                       │
└────────────────────┬────────────────────────────┘
                     │ REST API / WebSocket
                     ↓
┌─────────────────────────────────────────────────┐
│             Backend (Go)                        │
│                                                 │
│  API Server (Gin Framework)                     │
│  ├─ REST API Handler                           │
│  ├─ WebSocket Handler                          │
│  └─ Configuration Loader                       │
│                                                 │
│  Core Services                                  │
│  ├─ Docker Manager                             │
│  ├─ Kubernetes Manager                         │
│  ├─ Registry Manager                           │
│  └─ Deployment Controller                      │
│                                                 │
│  AI Engine                                      │
│  ├─ LLM Client (OpenAI/Claude API)             │
│  ├─ Prompt Builder (Few-shot)                  │
│  ├─ Manifest Generator                         │
│  ├─ Resource Predictor                         │
│  └─ Historical Data Manager                    │
│                                                 │
│  Real-time Engine                               │
│  ├─ Metrics Collector (Goroutines)             │
│  ├─ Log Streamer                               │
│  └─ Event Publisher                            │
│                                                 │
│  Data Layer                                     │
│  ├─ Deployment History DB (SQLite)             │
│  └─ Cache (In-memory)                          │
└──────┬──────────┬──────────┬────────────────────┘
       │          │          │
       ↓          ↓          ↓
  ┌─────────┐ ┌──────────┐ ┌──────────┐
  │  Docker │ │ K8s API  │ │  LLM API │
  │   API   │ │ Servers  │ │ (OpenAI) │
  │ (Local) │ │ (Remote) │ │          │
  └─────────┘ └──────────┘ └──────────┘
```

## 주요 컴포넌트

### 1. Frontend Layer

#### 책임
- 사용자 인터페이스 제공
- 실시간 데이터 시각화
- 사용자 입력 처리
- WebSocket을 통한 실시간 업데이트

#### 기술 스택
- **React 18**: UI 라이브러리
- **TypeScript**: 타입 안정성
- **React Query**: 서버 상태 관리 및 캐싱
- **Recharts**: 메트릭 차트 렌더링
- **TailwindCSS**: 스타일링
- **WebSocket API**: 실시간 통신

#### 주요 모듈
- `components/`: UI 컴포넌트
  - `dashboard/`: 대시보드 화면
  - `deploy/`: 배포 관련 UI
  - `logs/`: 로그 뷰어
  - `common/`: 공통 컴포넌트
- `hooks/`: 커스텀 React Hooks
- `api/`: API 클라이언트

### 2. Backend Layer

#### 책임
- RESTful API 제공
- Docker/Kubernetes 통합
- AI 기반 Manifest 생성
- 실시간 메트릭 수집 및 스트리밍
- 배포 이력 관리

#### 기술 스택
- **Go 1.21+**: 높은 성능, 동시성 처리
- **Gin**: HTTP 웹 프레임워크
- **Gorilla WebSocket**: WebSocket 서버
- **docker/docker**: Docker Engine API
- **k8s.io/client-go**: Kubernetes API
- **OpenAI Go SDK**: LLM API 클라이언트

#### 주요 모듈

##### API Layer (`internal/api/`)
- `router.go`: API 라우팅 설정
- `handlers.go`: HTTP 요청 핸들러
- `websocket.go`: WebSocket 핸들러
- `middleware.go`: 인증, 로깅 등

##### Docker Manager (`internal/docker/`)
- `client.go`: Docker API 클라이언트
- `manager.go`: 컨테이너 관리
- `stats.go`: 컨테이너 통계 수집
- `logs.go`: 로그 스트리밍

##### Kubernetes Manager (`internal/kubernetes/`)
- `client.go`: K8s API 클라이언트
- `manager.go`: 클러스터 관리
- `deployer.go`: 배포 실행
- `resources.go`: 리소스 조회

##### AI Engine (`internal/ai/`)
- `client.go`: LLM API 클라이언트
- `prompt_builder.go`: 프롬프트 생성
- `manifest_generator.go`: Manifest 생성
- `resource_predictor.go`: 리소스 예측
- `few_shot.go`: Few-shot 예시 관리

##### Data Layer (`internal/data/`)
- `deployment_store.go`: 배포 이력 저장
- `similarity_search.go`: 유사 배포 검색
- `models.go`: 데이터 모델

### 3. AI Engine

#### 책임
- Docker 컨테이너 정보 분석
- Kubernetes Manifest 자동 생성
- 리소스 할당 최적화
- 과거 배포 패턴 학습

#### 주요 기능

##### 3.1 LLM 기반 Manifest 생성
```go
// 입력
type ContainerInfo struct {
    Name        string
    Image       string
    Ports       []int
    EnvVars     map[string]string
    CPUUsage    string
    MemoryUsage string
}

// 출력
type ManifestResult struct {
    Deployment string
    Service    string
    ConfigMap  string
    HPA        string
    Reasoning  string
    Confidence float64
}
```

##### 3.2 Few-shot Learning
- 사내 우수 배포 사례 10-15개 큐레이션
- 유사 서비스 검색 (이미지 타입, 언어, 용도 기반)
- 프롬프트에 유사 사례 3-5개 포함

##### 3.3 프롬프트 구조
```
System Prompt:
- 역할 정의 (K8s 전문가)
- 사내 정책 및 표준
- 출력 형식 지정

Few-shot Examples:
- 과거 유사 배포 사례 3-5개

User Input:
- 현재 컨테이너 정보
- 배포 요구사항
```

### 4. Data Flow

#### 4.1 모니터링 데이터 플로우

```
Docker/K8s API
      ↓
Metrics Collector (Goroutine)
      ↓
Process & Aggregate
      ↓
WebSocket Broadcast
      ↓
Frontend Update
```

#### 4.2 배포 데이터 플로우

```
User Input (Frontend)
      ↓
Deployment Request (API)
      ↓
Container Info Extraction
      ↓
Historical Data Search
      ↓
AI Analysis (LLM)
      ↓
Manifest Generation
      ↓
User Review (Frontend)
      ↓
Deployment Execution
      ↓
Status Monitoring
      ↓
History Storage
```

## 데이터 모델

### Deployment History

```go
type DeploymentHistory struct {
    ID              string
    ServiceName     string
    ImageName       string
    ImageTag        string
    ServiceType     string  // web, api, db, cache
    Language        string  // node, python, go, java

    // 설정
    CPURequest      string
    CPULimit        string
    MemoryRequest   string
    MemoryLimit     string
    Replicas        int

    // 실제 사용량 (7일 평균)
    ActualCPU       string
    ActualMemory    string

    // 배포 정보
    TargetCluster   string
    Namespace       string
    DeployedAt      time.Time

    // 결과
    Success         bool
    OOMEvents       int
    ThrottleEvents  int

    // AI 관련
    AIGenerated     bool
    AIConfidence    float64
}
```

### Cluster Configuration

```yaml
clusters:
  - name: aws-eks-seoul
    type: kubernetes
    kubeconfig: /path/to/kubeconfig
    context: arn:aws:eks:...
    registry: 123456789.dkr.ecr.ap-northeast-2.amazonaws.com

  - name: azure-aks-korea
    type: kubernetes
    kubeconfig: /path/to/kubeconfig
    context: azure-aks
    registry: myregistry.azurecr.io

docker:
  local:
    socket: unix:///var/run/docker.sock

ai:
  provider: openai
  model: gpt-4-turbo-preview
  api_key: ${OPENAI_API_KEY}
  temperature: 0.3
  max_tokens: 2000
```

## 보안 고려사항

### 1. 인증 및 권한
- API 키 기반 인증
- Kubernetes RBAC 연동
- Registry 자격증명 안전한 저장

### 2. 데이터 보호
- 민감 정보 (API Key, Password) 환경변수 관리
- Secret 데이터 암호화
- HTTPS 통신 강제

### 3. API 보안
- Rate Limiting
- CORS 설정
- Input Validation

## 성능 최적화

### 1. 백엔드
- Goroutine을 활용한 병렬 처리
- 메트릭 수집 주기 최적화 (2-5초)
- API 응답 캐싱 (Redis 또는 In-memory)

### 2. 프론트엔드
- React Query 캐싱
- 컴포넌트 lazy loading
- Virtual scrolling (긴 목록)

### 3. AI
- LLM 응답 캐싱 (동일 컨테이너)
- 프롬프트 길이 최적화
- Batch 요청 (가능한 경우)

## 확장성

### 수평 확장
- Stateless 설계
- 클러스터 설정 파일 기반 (코드 변경 없이 추가)
- WebSocket 연결 관리

### 기능 확장
- 플러그인 아키텍처 고려
- AI 프로바이더 추상화 (OpenAI, Claude, Azure OpenAI)
- Registry 타입 확장 (ECR, ACR, GCR, Harbor)

## 모니터링 및 로깅

### 애플리케이션 로그
- 구조화된 로깅 (JSON)
- 로그 레벨: DEBUG, INFO, WARN, ERROR
- 컨텍스트 정보 포함 (request ID, user ID)

### 메트릭
- API 응답 시간
- Docker/K8s API 호출 횟수
- AI API 호출 횟수 및 비용
- 배포 성공/실패율
- WebSocket 연결 수

### 헬스 체크
- `/health`: 기본 헬스 체크
- `/ready`: 의존성 체크 (Docker, K8s, AI API)

## 배포 아키텍처

### 로컬 개발
```
docker-compose.yml
├─ backend (Go)
├─ frontend (React)
└─ sqlite (데이터)
```

### 프로덕션 (Kubernetes)
```
Deployment
├─ backend pods (2 replicas)
└─ frontend pods (2 replicas)

Services
├─ backend-svc (ClusterIP)
└─ frontend-svc (LoadBalancer)

PersistentVolume
└─ deployment-history-pv (SQLite 데이터)
```

## 장애 처리

### Docker/K8s API 장애
- Retry 로직 (exponential backoff)
- Circuit Breaker 패턴
- 에러 알림 (사용자에게 표시)

### AI API 장애
- Fallback: 기본 템플릿 사용
- 에러 메시지 명확히 표시
- 수동 Manifest 작성 옵션 제공

### WebSocket 연결 끊김
- 자동 재연결 (최대 5회)
- Polling 모드로 전환 (fallback)
