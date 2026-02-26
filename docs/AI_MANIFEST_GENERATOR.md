# AI Manifest Generator 가이드

## 개요

AI Manifest Generator는 Docker 컨테이너 정보를 분석하여 Kubernetes Deployment, Service 등의 Manifest를 자동으로 생성하는 핵심 기능입니다.

## 작동 원리

### 1. 데이터 수집 단계

Docker 컨테이너에서 다음 정보를 추출합니다:

```go
type ContainerInfo struct {
    // 기본 정보
    Name        string
    Image       string
    ImageTag    string

    // 설정
    EnvVars     map[string]string
    Ports       []int
    Volumes     []VolumeMount
    Command     []string
    WorkingDir  string

    // 리소스 사용량
    CPUUsage    string  // 현재 사용 중인 CPU
    MemoryUsage string  // 현재 사용 중인 Memory

    // 네트워크
    NetworkMode string
}
```

### 2. 유사 배포 검색

과거 배포 이력에서 유사한 서비스를 검색합니다:

**검색 기준:**
- 이미지 이름 유사도
- 서비스 타입 (web, api, db, cache)
- 언어/런타임 (node, python, go, java)

**검색 쿼리:**
```sql
SELECT * FROM deployment_history
WHERE success = true
  AND (
    image_name LIKE '%nginx%' OR
    service_type = 'web'
  )
  AND actual_cpu IS NOT NULL
ORDER BY deployed_at DESC
LIMIT 5
```

### 3. 프롬프트 생성 (Few-shot Learning)

LLM에게 전달할 프롬프트를 구성합니다:

```
[System Prompt]
당신은 Kubernetes 전문가입니다. Docker 컨테이너 정보를 분석하여
프로덕션 레벨의 Kubernetes Manifest를 생성합니다.

[사내 정책]
- 모든 Deployment에 리소스 requests/limits 필수
- readiness/liveness probe 필수
- SecurityContext 설정 (non-root 실행)
- NetworkPolicy 기본 deny-all
- 모든 리소스에 적절한 labels 추가

[Few-shot Examples]
예시 1: nginx 웹서버
입력:
  - 이미지: nginx:1.21
  - 포트: 80
  - 환경변수: PORT=80
  - 현재 리소스: CPU 150m, Memory 200Mi

출력:
  - CPU: 500m (request), 1000m (limit)
  - Memory: 512Mi (request), 1Gi (limit)
  - Replicas: 3
  - HPA: 활성화 (CPU 70% 임계값)
  - 근거: nginx는 일반적으로 낮은 리소스 사용. 트래픽 변동 대비 HPA 권장.

예시 2: Node.js API
...

[현재 요청]
컨테이너명: my-api
이미지: node:18-alpine
포트: 3000
환경변수:
  - PORT=3000
  - NODE_ENV=production
  - DB_HOST=postgres.default.svc.cluster.local
현재 리소스: CPU 200m, Memory 350Mi

요구사항:
- 고가용성 필요
- Auto scaling 필요

[출력 형식]
1. 서비스 타입 분석
2. 리소스 추천 및 근거
3. Deployment YAML
4. Service YAML
5. HPA YAML (필요 시)
6. ConfigMap YAML (필요 시)
```

### 4. LLM 호출

```go
func GenerateManifest(ctx context.Context, containerInfo ContainerInfo, similarDeployments []DeploymentHistory) (*ManifestResult, error) {
    // 프롬프트 구성
    prompt := buildPrompt(containerInfo, similarDeployments)

    // OpenAI API 호출
    response, err := openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model: "gpt-4-turbo-preview",
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    "system",
                Content: systemPrompt,
            },
            {
                Role:    "user",
                Content: prompt,
            },
        },
        Temperature: 0.3,  // 일관된 출력을 위해 낮은 값
        MaxTokens:   2000,
    })

    // 응답 파싱
    return parseResponse(response.Choices[0].Message.Content)
}
```

### 5. 응답 파싱 및 검증

LLM 응답을 파싱하고 검증합니다:

```go
func parseResponse(content string) (*ManifestResult, error) {
    var result ManifestResult

    // YAML 블록 추출
    result.Deployment = extractYAML(content, "Deployment")
    result.Service = extractYAML(content, "Service")
    result.HPA = extractYAML(content, "HorizontalPodAutoscaler")

    // YAML 문법 검증
    if err := validateYAML(result.Deployment); err != nil {
        return nil, fmt.Errorf("invalid deployment YAML: %w", err)
    }

    // 보안 정책 검증
    if err := validateSecurityPolicy(result.Deployment); err != nil {
        return nil, fmt.Errorf("security policy violation: %w", err)
    }

    // 추천 근거 추출
    result.Reasoning = extractReasoning(content)

    return &result, nil
}
```

## 프롬프트 최적화 기법

### 1. Few-shot Learning

**목적:** AI가 사내 패턴을 학습하도록 우수 사례 제공

**예시 선별 기준:**
- 안정적으로 운영 중 (success = true, oom_events = 0)
- 리소스 효율적 (과할당 < 20%)
- 다양한 타입 커버 (web, api, db, cache)

**예시 개수:**
- 최소 3개, 최대 5개
- 너무 많으면 토큰 낭비, 너무 적으면 학습 부족

### 2. Chain-of-Thought

AI가 단계별로 추론하도록 유도:

```
다음 단계로 분석하세요:

1. 서비스 타입 분석
   - 이미지 이름에서 추론
   - 환경변수에서 추론
   - 포트 번호에서 추론

2. 리소스 요구사항 추정
   - 현재 사용량 분석
   - 유사 서비스 평균 계산
   - 피크 트래픽 대비 여유 추가

3. 고가용성 설계
   - Replica 수 결정
   - HPA 필요성 판단
   - Health check 설정

4. 보안 설정
   - SecurityContext 적용
   - NetworkPolicy 필요성
   - Secret 관리 방안
```

### 3. Temperature 조정

```go
Temperature: 0.3  // 낮은 값 = 일관된 출력, 높은 값 = 창의적 출력
```

**Manifest 생성은 일관성이 중요**하므로 낮은 Temperature 사용

### 4. 출력 형식 강제

JSON Schema 또는 명확한 구분자 사용:

```
출력 형식:

## 분석 결과
[서비스 타입, 추천 근거]

## Deployment YAML
```yaml
[YAML 내용]
```

## Service YAML
```yaml
[YAML 내용]
```
```

## 지속적 개선

### 1. 배포 피드백 루프

```
배포 실행
    ↓
7일간 모니터링
    ↓
실제 리소스 사용량 수집
    ↓
AI 추천 vs 실제 비교
    ↓
차이 분석
    ↓
데이터베이스 저장
    ↓
다음 배포 시 Few-shot 예시로 활용
```

**저장 데이터:**
```go
type DeploymentFeedback struct {
    DeploymentID string

    // AI 추천
    RecommendedCPU    string
    RecommendedMemory string

    // 실제 사용량 (7일 평균)
    ActualCPU    string
    ActualMemory string

    // 차이
    CPUDiff    float64  // (추천 - 실제) / 실제
    MemoryDiff float64

    // 이벤트
    OOMEvents      int
    ThrottleEvents int

    // 평가
    Optimal bool  // 차이가 10-30% 범위면 최적
}
```

### 2. 프롬프트 A/B 테스트

```go
type PromptVersion struct {
    ID          string
    Version     string  // "v1.0", "v1.1"
    SystemPrompt string
    Template    string

    // 성과 지표
    UsageCount      int
    SuccessRate     float64
    OptimalRate     float64
    AvgConfidence   float64
    UserSatisfaction float64
}
```

**테스트 프로세스:**
- 새 프롬프트 버전 작성
- 50% 트래픽에 적용
- 1주일간 성과 측정
- 우수한 버전을 기본으로 채택

### 3. Few-shot 예시 업데이트

**주간:**
- 최근 1주일 배포 중 우수 사례 추가
- 실패 사례 분석 및 제외

**월간:**
- 전체 예시 재평가
- 오래된 예시 교체
- 새로운 패턴 추가

## 성능 최적화

### 1. 응답 캐싱

동일한 이미지에 대한 요청은 캐싱:

```go
type CacheKey struct {
    ImageName string
    ImageTag  string
    Options   DeployOptions
}

var manifestCache = make(map[CacheKey]*ManifestResult)
var cacheTTL = 1 * time.Hour
```

### 2. 프롬프트 길이 최적화

- 불필요한 설명 제거
- 예시는 핵심만 포함
- 목표: 2000 토큰 이내

### 3. 병렬 처리

여러 컨테이너 동시 배포 시:

```go
func GenerateManifests(containers []ContainerInfo) []*ManifestResult {
    results := make([]*ManifestResult, len(containers))
    var wg sync.WaitGroup

    for i, container := range containers {
        wg.Add(1)
        go func(index int, c ContainerInfo) {
            defer wg.Done()
            results[index], _ = GenerateManifest(c)
        }(i, container)
    }

    wg.Wait()
    return results
}
```

## 에러 처리

### 1. LLM API 장애

```go
func GenerateManifestWithFallback(containerInfo ContainerInfo) (*ManifestResult, error) {
    // 1차: AI 생성 시도
    result, err := GenerateManifest(containerInfo)
    if err == nil {
        return result, nil
    }

    log.Warn("AI generation failed, using template fallback", "error", err)

    // 2차: 템플릿 기반 생성
    result, err = GenerateFromTemplate(containerInfo)
    if err == nil {
        result.AIGenerated = false
        return result, nil
    }

    return nil, fmt.Errorf("both AI and template generation failed: %w", err)
}
```

### 2. 잘못된 YAML 생성

```go
func validateAndFix(yamlContent string) (string, error) {
    // YAML 파싱 시도
    var obj map[string]interface{}
    if err := yaml.Unmarshal([]byte(yamlContent), &obj); err != nil {
        // 자동 수정 시도
        fixed := autoFixYAML(yamlContent)
        if err := yaml.Unmarshal([]byte(fixed), &obj); err != nil {
            return "", fmt.Errorf("invalid YAML: %w", err)
        }
        return fixed, nil
    }

    return yamlContent, nil
}
```

### 3. 보안 정책 위반

```go
func validateSecurityPolicy(deployment string) error {
    violations := []string{}

    // privileged 모드 금지
    if strings.Contains(deployment, "privileged: true") {
        violations = append(violations, "privileged mode not allowed")
    }

    // hostNetwork 금지
    if strings.Contains(deployment, "hostNetwork: true") {
        violations = append(violations, "hostNetwork not allowed")
    }

    // 리소스 limits 필수
    if !strings.Contains(deployment, "limits:") {
        violations = append(violations, "resource limits required")
    }

    if len(violations) > 0 {
        return fmt.Errorf("security violations: %v", violations)
    }

    return nil
}
```

## 비용 관리

### LLM API 비용

**GPT-4 Turbo 가격 (2024년 1월 기준):**
- Input: $0.01 / 1K tokens
- Output: $0.03 / 1K tokens

**예상 비용:**
```
배포 1회당:
- 프롬프트: 약 1500 tokens ($0.015)
- 응답: 약 500 tokens ($0.015)
- 총: $0.03

월 1000회 배포 시: $30
```

**절감 방안:**
1. 캐싱 (중복 요청 방지)
2. 프롬프트 최적화 (토큰 수 감소)
3. 필요 시 Claude (더 저렴) 사용

## 모니터링 지표

### 성능 지표

```go
type AIMetrics struct {
    // 사용량
    TotalRequests     int64
    CachedResponses   int64
    CacheHitRate      float64

    // 성능
    AvgLatency        time.Duration
    P95Latency        time.Duration
    P99Latency        time.Duration

    // 품질
    SuccessRate       float64  // YAML 파싱 성공률
    UserApprovalRate  float64  // 사용자가 수정 없이 승인한 비율
    OptimalRate       float64  // 리소스 할당이 최적 범위인 비율

    // 비용
    TotalTokens       int64
    TotalCost         float64
}
```

### 알림 조건

- LLM API 에러율 > 10%
- 응답 시간 P95 > 10초
- 사용자 승인률 < 50%
- 일일 비용 > $100

## 테스트

### 단위 테스트

```go
func TestGenerateManifest(t *testing.T) {
    containerInfo := ContainerInfo{
        Name:  "test-nginx",
        Image: "nginx:1.21",
        Ports: []int{80},
    }

    result, err := GenerateManifest(context.Background(), containerInfo, nil)

    assert.NoError(t, err)
    assert.NotEmpty(t, result.Deployment)
    assert.Contains(t, result.Deployment, "kind: Deployment")

    // YAML 유효성 검증
    var deployment appsv1.Deployment
    err = yaml.Unmarshal([]byte(result.Deployment), &deployment)
    assert.NoError(t, err)
}
```

### 통합 테스트

실제 LLM API 호출 (staging 환경):

```go
func TestAIGenerationE2E(t *testing.T) {
    // 다양한 이미지 타입 테스트
    testCases := []string{
        "nginx:1.21",
        "postgres:14",
        "redis:alpine",
        "node:18-alpine",
    }

    for _, image := range testCases {
        result, err := GenerateManifest(context.Background(), ContainerInfo{Image: image}, nil)
        assert.NoError(t, err)
        assert.NotEmpty(t, result.Reasoning)
    }
}
```

## 참고자료

- [OpenAI API Documentation](https://platform.openai.com/docs)
- [Kubernetes API Conventions](https://kubernetes.io/docs/reference/using-api/api-concepts/)
- [Prompt Engineering Guide](https://www.promptingguide.ai/)
