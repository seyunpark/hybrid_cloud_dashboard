# ADR-004: SQLite를 배포 이력 저장소로 선택

## 상태
승인됨

## 컨텍스트

AI 기반 Manifest 생성을 위해 과거 배포 이력을 저장하고 검색할 데이터베이스가 필요합니다.

### 요구사항
- 배포 이력 저장 (서비스명, 리소스, 성공/실패 등)
- 유사 배포 검색 (Few-shot learning용)
- 실제 리소스 사용량 저장
- 백업 및 복구 용이성
- 간단한 설정 및 운영

### 데이터 특성
- 쓰기: 배포 시마다 (일 10-50건)
- 읽기: AI 분석 시마다 (일 100-200건)
- 데이터 크기: 작음 (연간 수만 건, MB 단위)
- 복잡한 조인 불필요
- 트랜잭션 중요도: 중간 (손실되어도 치명적이지 않음)

## 결정

배포 이력 저장소로 **SQLite**를 선택합니다.

### 선택 이유

1. **간단한 설정**
   - 별도 서버 불필요
   - 단일 파일로 관리
   - 즉시 사용 가능

2. **충분한 성능**
   - 읽기 성능 우수
   - 동시 쓰기가 많지 않아 제약 없음
   - 인덱스로 빠른 검색

3. **백업 용이성**
   - 파일 복사만으로 백업
   - 버전 관리 가능
   - 복구 간단

4. **배포 간소화**
   - 단일 바이너리에 포함 가능
   - 의존성 없음
   - 컨테이너 배포 시 볼륨 하나만 필요

5. **개발 편의성**
   - SQL 표준 지원
   - Go database/sql 인터페이스
   - 테스트 용이 (In-memory 모드)

## 스키마 설계

```sql
CREATE TABLE deployment_history (
    id TEXT PRIMARY KEY,
    service_name TEXT NOT NULL,
    image_name TEXT NOT NULL,
    image_tag TEXT,
    service_type TEXT,  -- web, api, db, cache
    language TEXT,       -- node, python, go, java

    -- 배포 설정
    cpu_request TEXT,
    cpu_limit TEXT,
    memory_request TEXT,
    memory_limit TEXT,
    replicas INTEGER,

    -- 실제 사용량 (7일 평균)
    actual_cpu TEXT,
    actual_memory TEXT,

    -- 배포 정보
    target_cluster TEXT,
    namespace TEXT,
    deployed_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    -- 결과
    success BOOLEAN,
    oom_events INTEGER DEFAULT 0,
    throttle_events INTEGER DEFAULT 0,

    -- AI 관련
    ai_generated BOOLEAN DEFAULT FALSE,
    ai_confidence REAL,

    -- 검색용 인덱스
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_service_type ON deployment_history(service_type);
CREATE INDEX idx_image_name ON deployment_history(image_name);
CREATE INDEX idx_success ON deployment_history(success);
CREATE INDEX idx_deployed_at ON deployment_history(deployed_at);
```

## 사용 패턴

### 배포 이력 저장
```go
func SaveDeployment(history DeploymentHistory) error {
    _, err := db.Exec(`
        INSERT INTO deployment_history
        (id, service_name, image_name, cpu_request, ...)
        VALUES (?, ?, ?, ?, ...)
    `, history.ID, history.ServiceName, ...)
    return err
}
```

### 유사 배포 검색 (Few-shot용)
```go
func FindSimilarDeployments(imageType string, limit int) ([]DeploymentHistory, error) {
    rows, err := db.Query(`
        SELECT * FROM deployment_history
        WHERE success = 1
          AND image_name LIKE ?
          AND actual_cpu IS NOT NULL
        ORDER BY deployed_at DESC
        LIMIT ?
    `, "%"+imageType+"%", limit)
    // ...
}
```

## 결과

### 긍정적 영향
- 설정 및 운영 간소화
- 빠른 개발 시작 가능
- 백업/복구 용이
- 컨테이너 배포 간단

### 부정적 영향
- 동시 쓰기 제약 (문제 안 됨)
- 대규모 확장성 제한 (현재는 불필요)

### 마이그레이션 전략
추후 데이터가 크게 증가하면 PostgreSQL로 마이그레이션 가능:
- 동일한 SQL 인터페이스
- 마이그레이션 스크립트 작성 간단
- 애플리케이션 코드 최소 변경

## 대안

### PostgreSQL
**장점:**
- 동시 쓰기 성능 우수
- 확장성 높음
- 복잡한 쿼리 지원

**단점:**
- 별도 서버 필요
- 설정 복잡
- 오버헤드 (현재 데이터 규모에는 과함)

**선택하지 않은 이유:** 초기 단계에서는 오버엔지니어링, SQLite로 충분

### MySQL/MariaDB
**장점:**
- 성능 우수
- 널리 사용됨

**단점:**
- PostgreSQL과 유사한 단점
- 현재 필요 없음

**선택하지 않은 이유:** SQLite로 충분

### NoSQL (MongoDB 등)
**장점:**
- 스키마 유연성
- 수평 확장 용이

**단점:**
- SQL 대비 쿼리 복잡
- 검색 성능 (인덱스 설계 필요)
- 오버헤드

**선택하지 않은 이유:** 구조화된 데이터이므로 SQL이 적합

### In-memory (Redis 등)
**장점:**
- 매우 빠름

**단점:**
- 영속성 제한
- 메모리 크기 제약
- 배포 이력은 영구 저장 필요

**선택하지 않은 이유:** 영속성이 중요한 데이터

## 운영 계획

### 백업
- 일일 자동 백업 (파일 복사)
- 최근 30일 보관
- 클라우드 스토리지 업로드

### 유지보수
- 월간: 오래된 데이터 아카이빙 (1년 이상)
- 분기: 데이터베이스 최적화 (VACUUM)

### 모니터링
- 데이터베이스 파일 크기
- 쿼리 응답 시간
- 디스크 사용량

## 성능 최적화

### 인덱스
- 검색 쿼리에 사용되는 컬럼에 인덱스
- 복합 인덱스 고려 (service_type + success)

### 쿼리 최적화
- LIMIT 사용으로 결과 제한
- 필요한 컬럼만 SELECT
- EXPLAIN으로 쿼리 플랜 확인

### Connection Pool
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

## 관련 결정
- ADR-003: AI Manifest 생성 (SQLite에서 Few-shot 데이터 조회)

## 참고자료
- [SQLite Documentation](https://www.sqlite.org/docs.html)
- [SQLite Use Cases](https://www.sqlite.org/whentouse.html)
- [Go database/sql](https://pkg.go.dev/database/sql)
