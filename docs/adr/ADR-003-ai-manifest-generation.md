# ADR-003: LLM 기반 Manifest 생성 방식

## 상태
승인됨

## 컨텍스트

Docker 컨테이너를 Kubernetes에 배포할 때 Deployment, Service 등의 Manifest를 생성해야 합니다. 수동 작성은 시간이 오래 걸리고, 리소스 할당 및 보안 설정 등에서 실수가 발생할 수 있습니다.

### 요구사항
- 컨테이너 정보 기반 자동 Manifest 생성
- 적절한 리소스 할당 추천
- 사내 표준 정책 자동 적용
- 보안 설정 포함
- 사용자가 이해할 수 있는 설명 제공
- 지속적인 학습 및 개선

## 결정

**LLM (Large Language Model) 기반 Few-shot Learning 방식**으로 Kubernetes Manifest를 생성합니다.

### 주요 접근 방식

1. **LLM 활용**
   - GPT-4 또는 Claude 3.5 Sonnet 사용
   - 컨테이너 정보를 분석하여 Manifest 생성
   - 리소스 할당 추천 및 근거 제시

2. **Few-shot Learning**
   - 사내 우수 배포 사례 10-15개 큐레이션
   - 유사 서비스 검색하여 3-5개를 프롬프트에 포함
   - 과거 패턴을 학습하여 추천 정확도 향상

3. **프롬프트 엔지니어링**
   - 시스템 프롬프트에 사내 정책 명시
   - 구조화된 입력/출력 형식
   - Chain-of-Thought로 추론 과정 명시

4. **지속적 개선**
   - 배포 이력 저장 (성공/실패, 실제 리소스 사용량)
   - 피드백 루프로 Few-shot 예시 업데이트
   - A/B 테스트로 프롬프트 최적화

## 결정 근거

### 1. LLM이 적합한 이유

**복잡한 규칙 처리:**
- 수십 가지 설정 옵션과 정책 규칙
- 단순 템플릿으로는 모든 케이스 커버 어려움
- LLM은 컨텍스트 이해하고 유연하게 생성

**자연어 추론:**
- 이미지 이름, 환경변수에서 서비스 타입 추론
- "nginx" → 웹서버 → 적절한 리소스 할당
- "postgres" → DB → StatefulSet 권장

**설명 가능성:**
- 추천 근거를 자연어로 제공
- 사용자 신뢰도 향상
- 학습 효과 (사용자가 K8s 학습)

### 2. Few-shot Learning 선택

**Fine-tuning 대비 장점:**
- 빠른 시작 (데이터 100개로 충분)
- 프롬프트만 수정하여 즉시 개선 가능
- 모델 학습 인프라 불필요
- 비용 효율적

**유사 사례 활용:**
- "Node.js API"는 과거 Node.js API 배포 참고
- 실제 사용된 설정이므로 현실적
- 트래픽 패턴, 리소스 사용량 반영

### 3. 데이터 수집 및 학습

**초기 데이터:**
```
사내 K8s 클러스터에서 수집:
- 기존 Deployment 100-200개
- 실제 리소스 사용량 (Prometheus)
- 배포 성공/실패 이력
- OOM, CPU throttling 이벤트
```

**데이터 구조:**
```json
{
  "service_name": "user-api",
  "image": "node:16-alpine",
  "type": "api",
  "env_count": 15,
  "ports": [3000],
  "allocated_cpu": "1000m",
  "allocated_memory": "1Gi",
  "actual_cpu_avg": "450m",
  "actual_memory_avg": "800Mi",
  "success": true,
  "oom_events": 0
}
```

**Few-shot 예시 선별:**
- 안정적으로 운영 중인 서비스
- 리소스 효율적인 서비스
- 다양한 타입 (웹, API, DB, 캐시 등)

## 구현 아키텍처

```
User Request
    ↓
Extract Container Info
    ↓
Search Similar Deployments
    ↓
Build Prompt
├─ System: 사내 정책
├─ Few-shot: 유사 사례 3-5개
└─ User: 현재 컨테이너 정보
    ↓
LLM API Call
    ↓
Parse Response
├─ Manifest YAML
├─ Resource Recommendations
└─ Reasoning
    ↓
User Review & Edit
    ↓
Deploy
    ↓
Store History
    ↓
Update Few-shot Examples
```

## 프롬프트 구조

```
System Prompt:
- 역할: Kubernetes 전문가
- 사내 정책:
  * 모든 Pod에 리소스 requests/limits 필수
  * readiness/liveness probe 필수
  * SecurityContext 설정 (non-root)
  * NetworkPolicy 적용
- 출력 형식: YAML + 추천 근거

Few-shot Examples:
예시 1: nginx 웹서버
- 이미지: nginx:1.21
- 리소스: CPU 500m, Memory 512Mi
- Replicas: 3
- 실제 사용: CPU 300m, Memory 380Mi
- 결과: 성공, 안정적 운영

예시 2: Node.js API
...

User Input:
컨테이너명: my-api
이미지: node:18-alpine
환경변수: PORT=3000, DB_HOST=...
현재 리소스: CPU 200m, Memory 350Mi
요구사항: 고가용성 필요
```

## 결과

### 긍정적 영향
- **개발 생산성 향상**: Manifest 작성 시간 30분 → 2분
- **리소스 최적화**: 과할당 평균 20% 감소
- **배포 안정성**: 정책 위반 자동 방지
- **학습 효과**: 사용자가 K8s 모범 사례 학습
- **L5 AI 역량**: 프롬프트 엔지니어링, Few-shot learning 경험

### 부정적 영향
- **API 비용**: LLM API 호출 비용 발생
- **응답 시간**: 5-10초 소요 (실시간은 아님)
- **정확도**: 100% 완벽하지 않음 (사용자 리뷰 필요)

### 완화 전략
- **비용**: 응답 캐싱, 동일 이미지 재사용
- **응답 시간**: 비동기 처리, 로딩 UI
- **정확도**: 사용자 승인 단계, 수정 가능

## 대안

### 1. 규칙 기반 템플릿
**방식:**
- if-else 로직으로 템플릿 선택
- 간단한 변수 치환

**장점:**
- 빠른 응답 (즉시)
- 비용 없음
- 예측 가능

**단점:**
- 복잡한 케이스 처리 어려움
- 새로운 패턴 추가 시 코드 수정 필요
- 리소스 최적화 제한적

**선택하지 않은 이유:** 유연성 부족, AI 역량 활용 불가

### 2. Fine-tuned Model
**방식:**
- CodeLLaMA 등 오픈소스 모델 파인튜닝
- 자체 호스팅

**장점:**
- API 비용 없음
- 데이터 외부 유출 없음
- 도메인 특화 가능

**단점:**
- 학습 데이터 많이 필요 (1000개+)
- GPU 인프라 필요
- 모델 유지보수 부담
- 초기 구축 시간 오래 걸림

**선택하지 않은 이유:** 프로젝트 일정 제약, Few-shot으로도 충분한 성능

### 3. Retrieval-Augmented Generation (RAG)
**방식:**
- 벡터 DB에 과거 Manifest 저장
- 유사도 검색 후 LLM에 전달

**장점:**
- 더 많은 컨텍스트 제공 가능
- 검색 정확도 높음

**단점:**
- 벡터 DB 인프라 필요
- 복잡도 증가
- Few-shot과 큰 차이 없을 수 있음

**선택하지 않은 이유:** Few-shot으로 충분, 초기 단계에서는 오버엔지니어링

## 성능 지표

### 평가 기준
1. **생성 정확도**: 사용자 수정 없이 배포 가능한 비율 (목표 80%)
2. **리소스 최적화**: 과할당 비율 (목표 20% 감소)
3. **정책 준수**: 보안 정책 위반 감소 (목표 95% 이상 준수)
4. **배포 성공률**: OOM, 에러 감소 (목표 10% → 3%)
5. **사용자 만족도**: 피드백 점수 (목표 4.5/5.0)

### 개선 프로세스
- 주간: 사용자 피드백 분석
- 월간: 프롬프트 A/B 테스트
- 분기: Few-shot 예시 업데이트

## 보안 고려사항

### API Key 관리
- 환경변수로 관리
- 로그에 노출 방지

### 데이터 프라이버시
- 민감 정보 (비밀번호 등) 제거 후 LLM 전송
- API 호출 로그 보관 정책

### 출력 검증
- 생성된 YAML 문법 검증
- 보안 정책 자동 체크
- 위험한 설정 차단 (privileged mode 등)

## 관련 결정
- ADR-004: SQLite 배포 이력 저장 (Few-shot 데이터 소스)

## 참고자료
- [OpenAI API Best Practices](https://platform.openai.com/docs/guides/prompt-engineering)
- [Few-shot Learning in LLMs](https://arxiv.org/abs/2005.14165)
- [Kubernetes Best Practices](https://kubernetes.io/docs/concepts/configuration/overview/)
