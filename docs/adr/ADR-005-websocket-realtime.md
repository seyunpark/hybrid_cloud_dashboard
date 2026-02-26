# ADR-005: WebSocket 실시간 통신 사용

## 상태
승인됨

## 컨텍스트

대시보드에서 Docker 컨테이너와 Kubernetes Pod의 리소스 사용량 (CPU, Memory)을 실시간으로 모니터링해야 합니다.

### 요구사항
- 실시간 메트릭 업데이트 (2-5초 간격)
- 로그 실시간 스트리밍
- 여러 클라이언트 동시 지원
- 낮은 지연 시간
- 효율적인 네트워크 사용

### 데이터 특성
- 업데이트 빈도: 높음 (초당 여러 번)
- 방향: 주로 서버 → 클라이언트 (단방향)
- 데이터 크기: 작음 (JSON, 수 KB)
- 지속 시간: 사용자가 대시보드를 보는 동안 계속

## 결정

실시간 데이터 전송을 위해 **WebSocket**을 사용합니다.

### 선택 이유

1. **양방향 실시간 통신**
   - Full-duplex 통신
   - 낮은 지연 시간
   - 서버 푸시 가능

2. **효율적인 네트워크 사용**
   - HTTP 폴링 대비 오버헤드 감소
   - 연결 유지로 handshake 반복 불필요
   - 헤더 오버헤드 최소화

3. **브라우저 네이티브 지원**
   - 모든 현대 브라우저 지원
   - JavaScript WebSocket API
   - 별도 플러그인 불필요

4. **Go 생태계 지원**
   - gorilla/websocket 라이브러리
   - Goroutine과 자연스러운 통합
   - 높은 동시성 처리

## 구현 설계

### 백엔드 (Go)

```go
// WebSocket 연결 관리
type Hub struct {
    clients    map[*Client]bool
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}

// 메트릭 수집 및 브로드캐스트
func (h *Hub) Run() {
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // Docker/K8s 메트릭 수집
            metrics := collectMetrics()

            // 모든 클라이언트에게 전송
            h.broadcast <- metrics

        case client := <-h.register:
            h.clients[client] = true

        case client := <-h.unregister:
            delete(h.clients, client)
            close(client.send)
        }
    }
}
```

### 프론트엔드 (React)

```typescript
// WebSocket 훅
function useWebSocket(url: string) {
  const [data, setData] = useState(null);
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    const ws = new WebSocket(url);

    ws.onopen = () => setIsConnected(true);
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      setData(data);
    };
    ws.onerror = () => setIsConnected(false);
    ws.onclose = () => setIsConnected(false);

    return () => ws.close();
  }, [url]);

  return { data, isConnected };
}
```

## WebSocket 엔드포인트

### 메트릭 스트리밍
- `/ws/docker/stats`: 로컬 Docker 컨테이너 통계
- `/ws/k8s/{cluster}/metrics`: 특정 클러스터 메트릭

### 로그 스트리밍
- `/ws/docker/{container_id}/logs`: Docker 로그
- `/ws/k8s/{cluster}/{namespace}/{pod}/logs`: K8s Pod 로그

### 배포 상태
- `/ws/deploy/{deploy_id}/status`: 배포 진행 상황

## 메시지 형식

```json
{
  "type": "docker_stats",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "containers": [
      {
        "id": "abc123",
        "name": "nginx",
        "cpu_percent": 5.2,
        "memory_usage": 134217728,
        "memory_limit": 536870912,
        "network_rx": 1024000,
        "network_tx": 512000
      }
    ]
  }
}
```

## 결과

### 긍정적 영향
- **사용자 경험**: 실시간 업데이트로 즉각적 피드백
- **네트워크 효율**: 폴링 대비 트래픽 80% 감소
- **서버 부하**: 불필요한 HTTP 요청 제거
- **확장성**: Goroutine으로 수천 개 연결 처리

### 부정적 영향
- **연결 관리 복잡도**: 재연결 로직 필요
- **방화벽/프록시**: 일부 환경에서 제약
- **디버깅**: HTTP보다 어려울 수 있음

### 완화 전략
- **재연결**: 자동 재연결 로직 (exponential backoff)
- **Fallback**: WebSocket 실패 시 HTTP 폴링으로 전환
- **모니터링**: 연결 상태, 메시지 전송 실패 추적

## 대안

### 1. HTTP Polling
**방식:**
- 클라이언트가 주기적으로 GET 요청
- 서버가 최신 데이터 응답

**장점:**
- 구현 간단
- 방화벽 문제 없음
- 디버깅 쉬움

**단점:**
- 네트워크 오버헤드 높음
- 서버 부하 높음 (불필요한 요청 많음)
- 실시간성 떨어짐 (폴링 간격 제약)

**선택하지 않은 이유:** 비효율적, 실시간 요구사항 충족 어려움

### 2. Server-Sent Events (SSE)
**방식:**
- 서버 → 클라이언트 단방향 스트리밍
- HTTP 기반

**장점:**
- 구현 간단
- 자동 재연결
- 방화벽 문제 적음

**단점:**
- 단방향만 지원 (클라이언트 → 서버 불가)
- 바이너리 데이터 전송 제한
- 브라우저 연결 제한 (도메인당 6개)

**선택하지 않은 이유:** 양방향 통신 필요 (향후 확장), 브라우저 제약

### 3. gRPC Streaming
**방식:**
- gRPC 양방향 스트리밍
- HTTP/2 기반

**장점:**
- 성능 우수
- 타입 안정성
- 양방향 스트리밍

**단점:**
- 브라우저 지원 제한적 (gRPC-Web 필요)
- 복잡한 설정
- 오버헤드

**선택하지 않은 이유:** 웹 브라우저 지원 복잡, WebSocket으로 충분

## 보안 고려사항

### 인증
- WebSocket 연결 시 토큰 전달
- 초기 handshake에서 검증
- 만료 시 연결 종료

### 권한 관리
- 클러스터별 접근 권한 확인
- 민감한 로그 필터링

### Rate Limiting
- 클라이언트당 메시지 전송 제한
- 비정상 트래픽 차단

## 성능 최적화

### 백엔드
```go
// 연결 풀 크기 제한
const maxClients = 1000

// 메시지 버퍼 크기
const bufferSize = 256

// 메트릭 수집 주기 최적화
const collectInterval = 2 * time.Second
```

### 프론트엔드
- 불필요한 재렌더링 방지 (React.memo, useMemo)
- 데이터 throttling (업데이트 빈도 제한)
- 화면에 보이는 항목만 업데이트

### 네트워크
- 메시지 압축 (gzip)
- Delta 업데이트 (변경된 부분만 전송)

## 장애 처리

### 연결 끊김
```typescript
function useResilientWebSocket(url: string) {
  const [reconnectCount, setReconnectCount] = useState(0);
  const maxReconnects = 5;

  useEffect(() => {
    let ws: WebSocket;

    function connect() {
      ws = new WebSocket(url);

      ws.onclose = () => {
        if (reconnectCount < maxReconnects) {
          const delay = Math.min(1000 * 2 ** reconnectCount, 30000);
          setTimeout(() => {
            setReconnectCount(prev => prev + 1);
            connect();
          }, delay);
        } else {
          // Fallback to polling
          switchToPollingMode();
        }
      };
    }

    connect();
    return () => ws?.close();
  }, [url, reconnectCount]);
}
```

### Heartbeat
- 주기적 ping/pong 메시지
- 연결 유지 확인
- Dead connection 감지

## 모니터링

### 메트릭
- 활성 WebSocket 연결 수
- 메시지 전송 성공/실패율
- 평균 메시지 크기
- 연결 지속 시간

### 알림
- 동시 연결 수 임계값 초과
- 메시지 전송 실패율 높음
- 재연결 시도 반복

## 관련 결정
- ADR-001: Go 백엔드 (Goroutine으로 WebSocket 효율적 처리)
- ADR-002: React 프론트엔드 (WebSocket API 지원)

## 참고자료
- [WebSocket RFC 6455](https://tools.ietf.org/html/rfc6455)
- [gorilla/websocket](https://github.com/gorilla/websocket)
- [MDN WebSocket API](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)
