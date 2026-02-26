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
    "env": [
      "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
      "PORT=80"
    ],
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
      "context": "arn:aws:eks:ap-northeast-2:123456789012:cluster/my-cluster",
      "status": "connected",
      "info": {
        "nodes": 3,
        "pods": 24,
        "namespaces": 5,
        "version": "1.28"
      }
    },
    {
      "name": "azure-aks-korea",
      "type": "kubernetes",
      "context": "azure-aks",
      "status": "connected",
      "info": {
        "nodes": 2,
        "pods": 12,
        "namespaces": 3,
        "version": "1.27"
      }
    }
  ]
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
      "node": "ip-10-0-1-100.ap-northeast-2.compute.internal",
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
      "conditions": [
        {
          "type": "Available",
          "status": "True",
          "reason": "MinimumReplicasAvailable"
        }
      ],
      "selector": {
        "app": "nginx"
      }
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
        {
          "name": "http",
          "port": 80,
          "target_port": 80,
          "protocol": "TCP"
        }
      ],
      "selector": {
        "app": "nginx"
      }
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
  "message": "Deployment scaled to 5 replicas",
  "deployment": {
    "name": "nginx-deployment",
    "namespace": "default",
    "replicas": 5
  }
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

## AI 기반 배포 API

### Docker → K8s 배포 (AI 기반 Manifest 생성)

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
    "reasoning": "Based on nginx web server pattern. Similar services average 300m CPU and 380Mi memory. Recommending 2 replicas for high availability with HPA enabled."
  },
  "manifests": {
    "deployment": "apiVersion: apps/v1\nkind: Deployment\n...",
    "service": "apiVersion: v1\nkind: Service\n...",
    "hpa": "apiVersion: autoscaling/v2\nkind: HorizontalPodAutoscaler\n..."
  },
  "estimated_cost": {
    "monthly_usd": 45.50,
    "breakdown": "CPU: $30, Memory: $15.50"
  }
}
```

### 배포 승인 및 실행

```
POST /api/deploy/:deploy_id/execute
```

**Request Body:**
```json
{
  "approved": true,
  "modifications": {
    "replicas": 3,
    "cpu_request": "700m"
  }
}
```

**Response:**
```json
{
  "deploy_id": "deploy-xyz789",
  "status": "deploying",
  "steps": [
    {
      "step": "push_image",
      "status": "in_progress",
      "message": "Pushing image to registry..."
    },
    {
      "step": "create_configmap",
      "status": "pending"
    },
    {
      "step": "create_deployment",
      "status": "pending"
    },
    {
      "step": "create_service",
      "status": "pending"
    }
  ]
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
    },
    {
      "step": "create_deployment",
      "status": "completed",
      "message": "Deployment created",
      "completed_at": "2024-01-15T11:02:00Z"
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
- `cluster` (string, optional): 클러스터 필터
- `success` (boolean, optional): 성공 여부 필터

**Response:**
```json
{
  "deployments": [
    {
      "id": "deploy-xyz789",
      "service_name": "nginx-app",
      "image": "nginx:1.21",
      "cluster": "aws-eks-seoul",
      "namespace": "default",
      "deployed_at": "2024-01-15T11:00:00Z",
      "success": true,
      "ai_generated": true,
      "ai_confidence": 0.92,
      "resources": {
        "cpu_request": "500m",
        "memory_request": "512Mi"
      }
    }
  ],
  "total": 150,
  "page": 1,
  "page_size": 50
}
```

## WebSocket API

### Docker 통계 스트리밍

```
WS /ws/docker/stats
```

**메시지 형식:**
```json
{
  "type": "docker_stats",
  "timestamp": "2024-01-15T11:00:00Z",
  "containers": [
    {
      "id": "abc123",
      "name": "nginx",
      "cpu_percent": 5.2,
      "memory_usage": 134217728,
      "memory_limit": 536870912,
      "network_rx_bytes": 1024000,
      "network_tx_bytes": 512000
    }
  ]
}
```

### K8s 메트릭 스트리밍

```
WS /ws/k8s/:cluster/metrics
```

**Query Parameters:**
- `namespace` (string, optional): 네임스페이스 필터

**메시지 형식:**
```json
{
  "type": "k8s_metrics",
  "timestamp": "2024-01-15T11:00:00Z",
  "cluster": "aws-eks-seoul",
  "pods": [
    {
      "name": "nginx-deployment-7d64d9f5b4-xk8tz",
      "namespace": "default",
      "cpu_usage": "250m",
      "memory_usage": "400Mi",
      "status": "Running"
    }
  ]
}
```

### 로그 스트리밍

```
WS /ws/docker/:container_id/logs
WS /ws/k8s/:cluster/:namespace/:pod/logs
```

**Query Parameters:**
- `follow` (boolean, optional): 실시간 팔로우 (default: true)
- `tail` (integer, optional): 마지막 N줄 (default: 100)
- `container` (string, optional): 특정 컨테이너 (K8s Pod용)

**메시지 형식:**
```json
{
  "type": "log",
  "timestamp": "2024-01-15T11:00:00Z",
  "log": "2024-01-15 11:00:00 [INFO] Server started on port 3000\n"
}
```

### 배포 상태 스트리밍

```
WS /ws/deploy/:deploy_id/status
```

**메시지 형식:**
```json
{
  "type": "deploy_status",
  "deploy_id": "deploy-xyz789",
  "timestamp": "2024-01-15T11:00:00Z",
  "step": "push_image",
  "status": "in_progress",
  "progress": 45,
  "message": "Pushing layer 3/7..."
}
```

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
      "name": "aws-eks-seoul",
      "type": "kubernetes",
      "kubeconfig_path": "/path/to/kubeconfig",
      "context": "arn:aws:eks:...",
      "registry": "123456789.dkr.ecr.ap-northeast-2.amazonaws.com"
    }
  ]
}
```

### AI 설정 조회

```
GET /api/config/ai
```

**Response:**
```json
{
  "provider": "openai",
  "model": "gpt-4-turbo-preview",
  "temperature": 0.3,
  "few_shot_enabled": true,
  "few_shot_examples": 5
}
```

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

의존성 (Docker, K8s, AI API) 확인 포함

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

## Rate Limiting

현재 버전에서는 Rate Limiting을 구현하지 않습니다. 향후 버전에서 추가 예정:
- API: 100 requests/minute per IP
- WebSocket: 1000 messages/minute per connection

## CORS

개발 환경에서는 모든 Origin 허용
프로덕션 환경에서는 허용된 도메인만 설정

## 버전 관리

API 버전은 URL 경로에 포함: `/api/v1/...`
현재는 버전 없이 시작하며, 향후 호환성 문제 발생 시 버전 추가 예정
