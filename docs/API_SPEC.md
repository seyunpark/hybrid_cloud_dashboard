# API 명세

## 개요

Base URL: `http://localhost:8080/api`

모든 API는 JSON 형식으로 데이터를 주고받습니다.

## 인증

현재 버전에서는 인증을 구현하지 않습니다. 향후 버전에서 API Key 또는 JWT 기반 인증을 추가할 예정입니다.

## 에러 응답 형식

```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Container not found",
    "details": {}
  }
}
```

---

## Docker API

### 컨테이너 목록 조회

```
GET /api/docker/containers
```

**Query Parameters:**
- `all` (boolean, optional): 중지된 컨테이너 포함 여부 (default: false)

**Response:**
```json
{
  "containers": [
    {
      "id": "abc123def456",
      "name": "nginx-app",
      "image": "nginx:1.21",
      "status": "running",
      "state": "running",
      "created_at": "2024-01-15T10:00:00Z",
      "ports": [
        {
          "private_port": 80,
          "public_port": 8080,
          "type": "tcp"
        }
      ],
      "stats": {
        "cpu_percent": 5.2,
        "memory_usage": 134217728,
        "memory_limit": 536870912,
        "memory_percent": 25.0,
        "network_rx": 1024000,
        "network_tx": 512000
      }
    }
  ]
}
```

### 컨테이너 상세 조회

```
GET /api/docker/containers/:id
```

**Response:**
```json
{
  "id": "abc123def456",
  "name": "nginx-app",
  "image": "nginx:1.21",
  "status": "running",
  "created_at": "2024-01-15T10:00:00Z",
  "config": {
    "env": ["PATH=/usr/local/sbin:...", "PORT=80"],
    "cmd": ["nginx", "-g", "daemon off;"],
    "working_dir": "/app",
    "exposed_ports": ["80/tcp"]
  },
  "mounts": [
    {
      "type": "bind",
      "source": "/host/path",
      "destination": "/container/path"
    }
  ],
  "network": {
    "ip_address": "172.17.0.2",
    "gateway": "172.17.0.1",
    "mac_address": "02:42:ac:11:00:02"
  }
}
```

### 컨테이너 재시작

```
POST /api/docker/containers/:id/restart
```

**Response:**
```json
{
  "success": true,
  "message": "Container restarted successfully"
}
```

### 컨테이너 중지

```
POST /api/docker/containers/:id/stop
```

**Response:**
```json
{
  "success": true,
  "message": "Container stopped successfully"
}
```

### 컨테이너 삭제

```
DELETE /api/docker/containers/:id
```

**Query Parameters:**
- `force` (boolean, optional): 강제 삭제 여부 (default: false)

**Response:**
```json
{
  "success": true,
  "message": "Container deleted successfully"
}
```

---

## Kubernetes API

### 클러스터 목록 조회

```
GET /api/k8s/clusters
```

**Response:**
```json
{
  "clusters": [
    {
      "name": "aws-eks-seoul",
      "type": "kubernetes",
      "context": "arn:aws:eks:ap-northeast-2:...",
      "status": "connected",
      "info": {
        "nodes": 3,
        "pods": 24,
        "namespaces": 5,
        "version": "1.28"
      }
    }
  ]
}
```

### 네임스페이스 목록 조회

```
GET /api/k8s/:cluster/namespaces
```

**Response:**
```json
{
  "namespaces": ["default", "kube-system", "kube-public", "my-app"]
}
```

### Pod 목록 조회

```
GET /api/k8s/:cluster/pods
```

**Query Parameters:**
- `namespace` (string, optional): 네임스페이스 필터 (default: "default")
- `label` (string, optional): 라벨 셀렉터 (예: "app=nginx")

**Response:**
```json
{
  "pods": [
    {
      "name": "nginx-deployment-7d64d9f5b4-xk8tz",
      "namespace": "default",
      "status": "Running",
      "phase": "Running",
      "node": "ip-10-0-1-100",
      "ip": "10.244.1.5",
      "created_at": "2024-01-15T10:00:00Z",
      "containers": [
        {
          "name": "nginx",
          "image": "nginx:1.21",
          "ready": true,
          "restart_count": 0,
          "state": "running"
        }
      ],
      "resources": {
        "cpu_request": "500m",
        "cpu_limit": "1000m",
        "memory_request": "512Mi",
        "memory_limit": "1Gi"
      },
      "conditions": [
        {
          "type": "Ready",
          "status": "True",
          "reason": "",
          "message": ""
        }
      ]
    }
  ]
}
```

### Deployment 목록 조회

```
GET /api/k8s/:cluster/deployments
```

**Query Parameters:**
- `namespace` (string, optional): 네임스페이스 필터 (default: "default")

**Response:**
```json
{
  "deployments": [
    {
      "name": "nginx-deployment",
      "namespace": "default",
      "replicas": 3,
      "ready_replicas": 3,
      "available_replicas": 3,
      "updated_replicas": 3,
      "image": "nginx:1.21",
      "created_at": "2024-01-15T10:00:00Z",
      "conditions": [...],
      "selector": { "app": "nginx" }
    }
  ]
}
```

### Service 목록 조회

```
GET /api/k8s/:cluster/services
```

**Query Parameters:**
- `namespace` (string, optional): 네임스페이스 필터 (default: "default")

**Response:**
```json
{
  "services": [
    {
      "name": "nginx-service",
      "namespace": "default",
      "type": "ClusterIP",
      "cluster_ip": "10.96.100.50",
      "ports": [
        { "name": "http", "port": 80, "target_port": 80, "protocol": "TCP" }
      ],
      "selector": { "app": "nginx" }
    }
  ]
}
```

### Deployment 스케일 조정

```
POST /api/k8s/:cluster/deployments/:namespace/:name/scale
```

**Request Body:**
```json
{
  "replicas": 5
}
```

**Response:**
```json
{
  "success": true,
  "message": "Deployment scaled to 5 replicas"
}
```

### Pod 재시작

```
POST /api/k8s/:cluster/pods/:namespace/:name/restart
```

**Response:**
```json
{
  "success": true,
  "message": "Pod deleted for restart"
}
```

---

## 단일 컨테이너 배포 API

### Docker -> K8s 배포 (AI 매니페스트 생성)

```
POST /api/deploy/docker-to-k8s
```

**Request Body:**
```json
{
  "container_id": "abc123def456",
  "cluster_name": "aws-eks-seoul",
  "namespace": "default",
  "options": {
    "high_availability": true,
    "enable_hpa": true
  }
}
```

**Response:**
```json
{
  "deploy_id": "deploy-xyz789",
  "status": "analyzing",
  "ai_analysis": {
    "service_type": "web-server",
    "detected_language": "none",
    "similar_deployments": 3
  },
  "recommendations": {
    "cpu_request": "500m",
    "cpu_limit": "1000m",
    "memory_request": "512Mi",
    "memory_limit": "1Gi",
    "replicas": 2,
    "enable_hpa": true,
    "reasoning": "Based on nginx web server pattern..."
  },
  "manifests": {
    "deployment": "apiVersion: apps/v1\nkind: Deployment\n...",
    "service": "apiVersion: v1\nkind: Service\n...",
    "hpa": "apiVersion: autoscaling/v2\n..."
  },
  "estimated_cost": {
    "monthly_usd": 45.50,
    "breakdown": "CPU: $30, Memory: $15.50"
  }
}
```

### 배포 실행

```
POST /api/deploy/:deploy_id/execute
```

**Request Body:**
```json
{
  "approved": true,
  "modifications": {
    "replicas": 3
  }
}
```

### 매니페스트 수정 요청

```
POST /api/deploy/:deploy_id/refine
```

**Request Body:**
```json
{
  "feedback": "CPU limit을 2000m으로 높여주세요"
}
```

### 배포 상태 조회

```
GET /api/deploy/:deploy_id/status
```

**Response:**
```json
{
  "deploy_id": "deploy-xyz789",
  "status": "completed",
  "started_at": "2024-01-15T11:00:00Z",
  "completed_at": "2024-01-15T11:02:30Z",
  "steps": [
    {
      "step": "push_image",
      "status": "completed",
      "message": "Image pushed successfully",
      "completed_at": "2024-01-15T11:01:00Z"
    }
  ],
  "result": {
    "deployment_name": "nginx-app",
    "namespace": "default",
    "replicas": 2,
    "pods_ready": "2/2",
    "service_url": "http://nginx-app.default.svc.cluster.local"
  }
}
```

### 배포 이력 조회

```
GET /api/deploy/history
```

**Query Parameters:**
- `limit` (integer, optional): 결과 개수 제한 (default: 50)

### 통합 배포 이력 조회 (페이지네이션)

```
GET /api/deploy/unified-history
```

**Query Parameters:**
- `page` (integer, optional): 페이지 번호 (default: 1)
- `limit` (integer, optional): 페이지 당 항목 수 (default: 20, max: 100)

**Response:**
```json
{
  "items": [
    {
      "id": "deploy-xyz789",
      "type": "single",
      "name": "nginx-app",
      "image_summary": "nginx:1.21",
      "cluster": "aws-eks-seoul",
      "namespace": "default",
      "status": "deployed",
      "ai_generated": true,
      "confidence": 0.92,
      "deployed_at": "2024-01-15T11:00:00Z",
      "single_detail": {
        "image_name": "nginx",
        "image_tag": "1.21",
        "replicas": 2
      }
    },
    {
      "id": "stack-abc123",
      "type": "stack",
      "name": "my-web-stack",
      "image_summary": "3 services",
      "cluster": "local-k8s",
      "namespace": "production",
      "status": "deployed",
      "ai_generated": true,
      "confidence": 0.87,
      "deployed_at": "2024-01-15T10:00:00Z",
      "stack_detail": {
        "service_count": 3,
        "services": ["frontend", "backend", "postgres"],
        "deploy_order": ["postgres", "backend", "frontend"]
      }
    }
  ],
  "total": 150,
  "page": 1,
  "limit": 20,
  "total_pages": 8
}
```

### 언디플로이 (K8s 리소스 삭제)

```
POST /api/deploy/:deploy_id/undeploy
```

배포된 K8s 리소스(Deployment, Service 등)를 삭제하고 배포 상태를 "deleted"로 변경합니다.

### 재배포

```
POST /api/deploy/:deploy_id/redeploy
```

저장된 매니페스트를 사용하여 재배포합니다.

### 배포 기록 삭제

```
DELETE /api/deploy/:deploy_id
```

DB에서 배포 기록을 삭제합니다.

---

## 스택 배포 API

여러 컨테이너를 하나의 연결된 스택으로 K8s에 배포합니다. AI가 서비스 간 토폴로지를 분석하고 배포 순서를 결정합니다.

### 활성 스택 배포 목록

```
GET /api/deploy/stack/
```

**Response:**
```json
[
  {
    "deploy_id": "stack-abc123",
    "stack_name": "my-web-stack",
    "status": "pending",
    "topology": { ... },
    "manifests": { ... }
  }
]
```

### 스택 배포 상세 조회

```
GET /api/deploy/stack/:deploy_id
```

**Response:**
```json
{
  "deploy_id": "stack-abc123",
  "status": "pending",
  "stack_name": "my-web-stack",
  "topology": {
    "services": [
      {
        "container_id": "abc123",
        "service_name": "frontend",
        "service_type": "web-application",
        "image": "react-app:latest"
      },
      {
        "container_id": "def456",
        "service_name": "backend",
        "service_type": "api-server",
        "image": "node-api:latest"
      }
    ],
    "connections": [
      {
        "from": "frontend",
        "to": "backend",
        "port": 3000,
        "env_var": "API_URL"
      }
    ],
    "deploy_order": ["backend", "frontend"]
  },
  "manifests": {
    "Deployment": {
      "frontend": "apiVersion: apps/v1\n...",
      "backend": "apiVersion: apps/v1\n..."
    },
    "Service": {
      "frontend": "apiVersion: v1\n...",
      "backend": "apiVersion: v1\n..."
    },
    "ConfigMap": {
      "frontend-config": "apiVersion: v1\n..."
    }
  },
  "reasoning": "Detected web frontend + API backend pattern...",
  "confidence": 0.87
}
```

### 스택 배포 상태 조회

```
GET /api/deploy/stack/:deploy_id/status
```

**Response:**
```json
{
  "deploy_id": "stack-abc123",
  "status": "deploying",
  "stack_name": "my-web-stack",
  "started_at": "2024-01-15T11:00:00Z",
  "services": {
    "backend": {
      "service_name": "backend",
      "status": "completed",
      "steps": [
        { "step": "push_image", "status": "completed" },
        { "step": "apply_manifest", "status": "completed" }
      ]
    },
    "frontend": {
      "service_name": "frontend",
      "status": "in_progress",
      "steps": [
        { "step": "push_image", "status": "completed" },
        { "step": "apply_manifest", "status": "in_progress" }
      ]
    }
  },
  "deploy_order": ["backend", "frontend"]
}
```

### 스택 배포 생성

```
POST /api/deploy/stack/
```

**Request Body:**
```json
{
  "container_ids": ["abc123", "def456", "ghi789"],
  "cluster_name": "local-k8s",
  "namespace": "production",
  "stack_name": "my-web-stack",
  "create_namespace": true,
  "prompt": "frontend와 backend 사이에 nginx reverse proxy를 추가해주세요",
  "options": {
    "high_availability": true,
    "enable_hpa": false
  }
}
```

**Response:**
```json
{
  "deploy_id": "stack-abc123",
  "status": "generating",
  "stack_name": "my-web-stack"
}
```

AI 매니페스트 생성은 비동기로 진행됩니다. 상태는 `GET /api/deploy/stack/:deploy_id`로 폴링하거나 WebSocket으로 수신합니다.

### 스택 매니페스트 수정 (피드백)

```
POST /api/deploy/stack/:deploy_id/refine
```

**Request Body:**
```json
{
  "feedback": "backend의 메모리를 2Gi로 늘려주세요"
}
```

### 스택 매니페스트 재생성

```
POST /api/deploy/stack/:deploy_id/regenerate
```

현재 컨테이너 정보와 프롬프트를 기반으로 매니페스트를 완전히 재생성합니다.

### 스택 배포 재편집 (Reopen)

```
POST /api/deploy/stack/:deploy_id/reopen
```

완료/실패된 스택 배포를 "pending" 상태로 되돌려 매니페스트를 수정할 수 있게 합니다.

### 스택 배포 실행

```
POST /api/deploy/stack/:deploy_id/execute
```

**Request Body:**
```json
{
  "approved": true,
  "cluster_name": "local-k8s",
  "namespace": "production",
  "create_namespace": true
}
```

### 스택 언디플로이

```
POST /api/deploy/stack/:deploy_id/undeploy
```

배포된 스택의 모든 K8s 리소스를 삭제합니다. 배포 순서의 역순으로 삭제됩니다.

### 스택 재배포

```
POST /api/deploy/stack/:deploy_id/redeploy
```

**Request Body (optional):**
```json
{
  "cluster_name": "different-cluster",
  "namespace": "staging",
  "create_namespace": true
}
```

### 스택 배포 삭제

```
DELETE /api/deploy/stack/:deploy_id
```

Soft-delete: 배포 상태를 "deleted"로 변경합니다. 통합 히스토리에 "deleted" 상태로 남습니다.

---

## 설정 API

### 클러스터 설정 조회

```
GET /api/config/clusters
```

**Response:**
```json
{
  "clusters": [
    {
      "name": "local-k8s",
      "type": "kubernetes",
      "kubeconfig": "~/.kube/config",
      "context": "docker-desktop",
      "registry": ""
    }
  ]
}
```

### kubeconfig 컨텍스트 목록

```
GET /api/config/kubecontexts
```

**Query Parameters:**
- `kubeconfig` (string, optional): kubeconfig 파일 경로

**Response:**
```json
{
  "contexts": [
    {
      "name": "docker-desktop",
      "cluster": "docker-desktop",
      "user": "docker-desktop",
      "namespace": "",
      "is_active": true
    }
  ]
}
```

### 클러스터 등록

```
POST /api/config/clusters
```

**Request Body:**
```json
{
  "name": "my-cluster",
  "context": "my-context",
  "type": "kubernetes",
  "kubeconfig": "~/.kube/config",
  "registry": "my-registry.com"
}
```

### 클러스터 등록 해제

```
DELETE /api/config/clusters/:name
```

### AI 설정 조회

```
GET /api/config/ai
```

**Response:**
```json
{
  "provider": "gemini",
  "model": "gemini-2.0-flash",
  "temperature": 0.3,
  "configured": true
}
```

### AI 설정 변경

```
PUT /api/config/ai
```

**Request Body:**
```json
{
  "provider": "gemini",
  "api_key": "your-api-key",
  "model": "gemini-2.0-flash"
}
```

### AI 모델 목록 조회

```
GET /api/config/ai/models
```

**Query Parameters:**
- `provider` (string, required): AI 프로바이더 (openai, claude, gemini)
- `api_key` (string, optional): API 키 (설정에 저장된 키 사용 가능)

**Response:**
```json
{
  "models": ["gemini-2.0-flash", "gemini-2.5-flash-preview-05-20", "gemini-2.5-pro-preview-05-06"]
}
```

---

## WebSocket API

### Docker 통계 스트리밍

```
WS /ws/docker/stats
```

2초 간격으로 모든 컨테이너의 CPU, 메모리, 네트워크 메트릭을 전송합니다.

### K8s 메트릭 스트리밍

```
WS /ws/k8s/:cluster/metrics
```

5초 간격으로 Pod, Deployment 정보를 전송합니다.

### Docker 로그 스트리밍

```
WS /ws/docker/:container_id/logs
```

Docker 컨테이너 로그를 실시간으로 스트리밍합니다.

### K8s Pod 로그 스트리밍

```
WS /ws/k8s/:cluster/:namespace/:pod/logs
```

### 배포 상태 스트리밍

```
WS /ws/deploy/:deploy_id/status
```

1초 간격으로 배포 상태를 전송합니다. 배포 완료/실패 시 연결이 종료됩니다.

---

## 헬스 체크

### 기본 헬스 체크

```
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T11:00:00Z"
}
```

### 준비 상태 체크

```
GET /ready
```

**Response:**
```json
{
  "status": "ready",
  "checks": {
    "docker": "ok",
    "kubernetes": "ok",
    "ai_api": "ok",
    "database": "ok"
  },
  "timestamp": "2024-01-15T11:00:00Z"
}
```

---

## 엔드포인트 요약

| 그룹 | 메서드 | 경로 | 설명 |
|------|--------|------|------|
| Docker | GET | `/api/docker/containers` | 컨테이너 목록 |
| Docker | GET | `/api/docker/containers/:id` | 컨테이너 상세 |
| Docker | POST | `/api/docker/containers/:id/restart` | 재시작 |
| Docker | POST | `/api/docker/containers/:id/stop` | 중지 |
| Docker | DELETE | `/api/docker/containers/:id` | 삭제 |
| K8s | GET | `/api/k8s/clusters` | 클러스터 목록 |
| K8s | GET | `/api/k8s/:cluster/namespaces` | 네임스페이스 목록 |
| K8s | GET | `/api/k8s/:cluster/pods` | Pod 목록 |
| K8s | GET | `/api/k8s/:cluster/deployments` | Deployment 목록 |
| K8s | GET | `/api/k8s/:cluster/services` | Service 목록 |
| K8s | POST | `/api/k8s/:cluster/deployments/:ns/:name/scale` | 스케일링 |
| K8s | POST | `/api/k8s/:cluster/pods/:ns/:name/restart` | Pod 재시작 |
| Deploy | POST | `/api/deploy/docker-to-k8s` | AI 매니페스트 생성 |
| Deploy | POST | `/api/deploy/:id/execute` | 배포 실행 |
| Deploy | POST | `/api/deploy/:id/refine` | 매니페스트 수정 |
| Deploy | POST | `/api/deploy/:id/undeploy` | 언디플로이 |
| Deploy | POST | `/api/deploy/:id/redeploy` | 재배포 |
| Deploy | DELETE | `/api/deploy/:id` | 기록 삭제 |
| Deploy | GET | `/api/deploy/:id/status` | 상태 조회 |
| Deploy | GET | `/api/deploy/history` | 이력 조회 |
| Deploy | GET | `/api/deploy/unified-history` | 통합 이력 (페이지네이션) |
| Stack | GET | `/api/deploy/stack/` | 활성 스택 목록 |
| Stack | GET | `/api/deploy/stack/:id` | 스택 상세 |
| Stack | GET | `/api/deploy/stack/:id/status` | 스택 상태 |
| Stack | POST | `/api/deploy/stack/` | 스택 생성 |
| Stack | POST | `/api/deploy/stack/:id/refine` | 스택 수정 |
| Stack | POST | `/api/deploy/stack/:id/regenerate` | 스택 재생성 |
| Stack | POST | `/api/deploy/stack/:id/reopen` | 스택 재편집 |
| Stack | POST | `/api/deploy/stack/:id/execute` | 스택 실행 |
| Stack | POST | `/api/deploy/stack/:id/undeploy` | 스택 언디플로이 |
| Stack | POST | `/api/deploy/stack/:id/redeploy` | 스택 재배포 |
| Stack | DELETE | `/api/deploy/stack/:id` | 스택 삭제 |
| Config | GET | `/api/config/clusters` | 클러스터 설정 |
| Config | GET | `/api/config/kubecontexts` | kubeconfig 컨텍스트 |
| Config | POST | `/api/config/clusters` | 클러스터 등록 |
| Config | DELETE | `/api/config/clusters/:name` | 클러스터 해제 |
| Config | GET | `/api/config/ai` | AI 설정 조회 |
| Config | PUT | `/api/config/ai` | AI 설정 변경 |
| Config | GET | `/api/config/ai/models` | AI 모델 목록 |
| Health | GET | `/health` | 헬스 체크 |
| Health | GET | `/ready` | 준비 상태 |
| WS | GET | `/ws/docker/stats` | Docker 메트릭 |
| WS | GET | `/ws/k8s/:cluster/metrics` | K8s 메트릭 |
| WS | GET | `/ws/docker/:id/logs` | Docker 로그 |
| WS | GET | `/ws/k8s/:cluster/:ns/:pod/logs` | K8s 로그 |
| WS | GET | `/ws/deploy/:id/status` | 배포 상태 |

**총 42 REST + 5 WebSocket = 47 엔드포인트**
