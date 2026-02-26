# 개발 환경 설정 가이드

## 시스템 요구사항

### 필수
- Docker 20.10 이상
- Docker Compose 2.0 이상
- Git

### 개발 시 필요
- Go 1.21 이상
- Node.js 18 이상
- npm 또는 yarn

### 선택사항
- Kubernetes 클러스터 (로컬 또는 원격)
- kubectl CLI
- OpenAI 또는 Claude API Key

## 빠른 시작 (Docker Compose)

### 1. 저장소 클론

```bash
git clone <repository-url>
cd AI_Project
```

### 2. 설정 파일 생성

```bash
# 예시 설정 파일 복사
cp configs/config.example.yaml configs/config.yaml
```

`configs/config.yaml` 편집:

```yaml
# AI 설정
ai:
  provider: openai  # openai 또는 claude
  api_key: your-api-key-here
  model: gpt-4-turbo-preview
  temperature: 0.3

# Docker 설정
docker:
  local:
    socket: unix:///var/run/docker.sock

# Kubernetes 클러스터 (선택사항)
clusters:
  - name: local-k8s
    type: kubernetes
    kubeconfig: ~/.kube/config
    context: docker-desktop

# Registry 설정 (선택사항)
registry:
  url: your-registry.io
  username: your-username
  password: your-password
```

### 3. 실행

```bash
docker-compose up -d
```

### 4. 접속

브라우저에서 http://localhost:3000 접속

## 로컬 개발 환경 설정

### 백엔드 (Go)

#### 1. Go 설치

```bash
# macOS (Homebrew)
brew install go

# Linux
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Windows
# https://go.dev/dl/ 에서 설치 프로그램 다운로드
```

#### 2. 의존성 설치

```bash
cd backend
go mod download
```

#### 3. 환경변수 설정

`.env` 파일 생성:

```bash
# backend/.env
OPENAI_API_KEY=your-api-key
CONFIG_PATH=../configs/config.yaml
PORT=8080
ENV=development
```

#### 4. 실행

```bash
# 백엔드 디렉토리에서
go run cmd/server/main.go
```

또는 Air를 사용한 Hot Reload:

```bash
# Air 설치
go install github.com/cosmtrek/air@latest

# 실행
air
```

#### 5. 테스트

```bash
# 전체 테스트
go test ./...

# 커버리지와 함께
go test -cover ./...

# 특정 패키지
go test ./internal/ai/...
```

### 프론트엔드 (React)

#### 1. Node.js 설치

```bash
# macOS (Homebrew)
brew install node

# Linux (nvm 사용 권장)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash
nvm install 18
nvm use 18

# Windows
# https://nodejs.org/ 에서 설치 프로그램 다운로드
```

#### 2. 의존성 설치

```bash
cd frontend
npm install
```

#### 3. 환경변수 설정

`.env` 파일 생성:

```bash
# frontend/.env
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

#### 4. 실행

```bash
npm run dev
```

브라우저에서 http://localhost:5173 접속

#### 5. 빌드

```bash
npm run build
```

빌드 결과물은 `dist/` 폴더에 생성됩니다.

#### 6. 테스트

```bash
# 단위 테스트
npm run test

# E2E 테스트 (Playwright)
npm run test:e2e
```

## 데이터베이스 설정

SQLite를 사용하므로 별도 설정이 필요 없습니다.

데이터베이스 파일 위치: `data/deployments.db`

### 초기화

```bash
# 백엔드 디렉토리에서
go run cmd/server/main.go migrate
```

### 마이그레이션

```bash
# 새 마이그레이션 생성
go run cmd/server/main.go migrate:create <migration_name>

# 마이그레이션 실행
go run cmd/server/main.go migrate:up

# 롤백
go run cmd/server/main.go migrate:down
```

## Kubernetes 클러스터 설정

### 로컬 클러스터 (Docker Desktop)

#### 1. Docker Desktop 설치

https://www.docker.com/products/docker-desktop

#### 2. Kubernetes 활성화

Docker Desktop 설정 → Kubernetes → Enable Kubernetes

#### 3. kubectl 설치

```bash
# macOS
brew install kubectl

# Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Windows
choco install kubernetes-cli
```

#### 4. 컨텍스트 확인

```bash
kubectl config get-contexts
kubectl config use-context docker-desktop
```

### 원격 클러스터 (AWS EKS, Azure AKS 등)

#### AWS EKS

```bash
# AWS CLI 설치
pip install awscli

# kubeconfig 업데이트
aws eks update-kubeconfig --region ap-northeast-2 --name my-cluster

# 확인
kubectl get nodes
```

#### Azure AKS

```bash
# Azure CLI 설치
curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash

# 로그인
az login

# kubeconfig 업데이트
az aks get-credentials --resource-group myResourceGroup --name myAKSCluster

# 확인
kubectl get nodes
```

## Container Registry 설정

### AWS ECR

```bash
# 로그인
aws ecr get-login-password --region ap-northeast-2 | docker login --username AWS --password-stdin 123456789.dkr.ecr.ap-northeast-2.amazonaws.com

# config.yaml에 추가
registry:
  url: 123456789.dkr.ecr.ap-northeast-2.amazonaws.com
```

### Azure ACR

```bash
# 로그인
az acr login --name myregistry

# config.yaml에 추가
registry:
  url: myregistry.azurecr.io
  username: <username>
  password: <password>
```

### Docker Hub

```bash
# 로그인
docker login

# config.yaml에 추가
registry:
  url: docker.io
  username: <username>
  password: <password>
```

## 개발 도구

### VS Code 확장

**Go 개발:**
- Go (golang.go)
- Go Test Explorer

**React 개발:**
- ES7+ React/Redux/React-Native snippets
- Prettier - Code formatter
- ESLint

**공통:**
- Docker
- Kubernetes
- YAML
- GitLens

### VS Code 설정

`.vscode/settings.json`:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "[typescript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "[typescriptreact]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  }
}
```

### Git Hooks (pre-commit)

```bash
# pre-commit 설치
pip install pre-commit

# 훅 설치
pre-commit install

# 수동 실행
pre-commit run --all-files
```

`.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.4.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files

  - repo: https://github.com/golangci/golangci-lint
    rev: v1.54.2
    hooks:
      - id: golangci-lint

  - repo: https://github.com/pre-commit/mirrors-prettier
    rev: v3.0.3
    hooks:
      - id: prettier
        types_or: [javascript, jsx, ts, tsx, json, yaml]
```

## 문제 해결

### Docker 소켓 권한 에러

```bash
# Linux
sudo usermod -aG docker $USER
newgrp docker

# 재로그인 필요
```

### Go 모듈 에러

```bash
# 캐시 정리
go clean -modcache

# 의존성 재설치
go mod download
go mod tidy
```

### 포트 충돌

다른 프로세스가 포트를 사용 중인 경우:

```bash
# 사용 중인 프로세스 확인
lsof -i :8080  # 백엔드
lsof -i :3000  # 프론트엔드

# 프로세스 종료
kill -9 <PID>
```

### Kubernetes 연결 에러

```bash
# 컨텍스트 확인
kubectl config get-contexts

# 올바른 컨텍스트로 전환
kubectl config use-context <context-name>

# 연결 테스트
kubectl cluster-info
```

## 성능 최적화 (개발 환경)

### Go 빌드 캐시

```bash
# 빌드 캐시 위치 확인
go env GOCACHE

# 캐시 정리 (필요 시)
go clean -cache
```

### Node.js 메모리 증가

```bash
# package.json scripts 수정
"dev": "NODE_OPTIONS=--max-old-space-size=4096 vite"
```

### Docker 빌드 속도 향상

`.dockerignore` 파일 생성:

```
node_modules
dist
.git
.env
*.log
```

## 다음 단계

- [기여 가이드](../CONTRIBUTING.md) 읽기
- [아키텍처 문서](./ARCHITECTURE.md) 이해하기
- [API 명세](./API_SPEC.md) 확인하기
- 첫 Issue 선택하여 기여하기
