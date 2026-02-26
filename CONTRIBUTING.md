# 기여 가이드

프로젝트에 관심을 가져주셔서 감사합니다! 이 문서는 프로젝트에 기여하는 방법을 안내합니다.

## 행동 강령

- 존중하고 포용적인 태도 유지
- 건설적인 피드백 제공
- 다양한 의견과 경험 존중
- 프로젝트와 커뮤니티에 집중

## 시작하기

### 1. 개발 환경 설정

[개발 환경 설정 가이드](./docs/SETUP.md)를 따라 로컬 환경을 구성하세요.

### 2. 저장소 포크 및 클론

```bash
# 포크 (GitHub UI에서)
# 클론
git clone https://github.com/your-username/AI_Project.git
cd AI_Project

# upstream 추가
git remote add upstream https://github.com/original-repo/AI_Project.git
```

### 3. 브랜치 생성

```bash
git checkout -b feature/your-feature-name
```

**브랜치 명명 규칙:**
- `feature/` - 새로운 기능
- `fix/` - 버그 수정
- `docs/` - 문서 변경
- `refactor/` - 리팩토링
- `test/` - 테스트 추가/수정
- `chore/` - 기타 작업

예시:
```
feature/add-redis-cache
fix/deployment-error-handling
docs/update-api-spec
```

## 기여 유형

### 버그 리포트

버그를 발견하셨나요? Issue를 생성해주세요.

**Issue 템플릿:**

```markdown
## 버그 설명
[명확하고 간결한 버그 설명]

## 재현 방법
1. 어디로 가서
2. 무엇을 클릭하고
3. 무엇을 보면
4. 에러 발생

## 예상 동작
[예상했던 동작]

## 실제 동작
[실제 발생한 동작]

## 스크린샷
[가능하면 스크린샷 첨부]

## 환경
- OS: [예: macOS 13.0]
- 브라우저: [예: Chrome 120]
- Docker 버전: [예: 24.0.0]
- Kubernetes 버전: [예: 1.28]

## 추가 정보
[기타 관련 정보]
```

### 기능 제안

새로운 기능을 제안하고 싶으신가요?

**Feature Request 템플릿:**

```markdown
## 기능 설명
[기능에 대한 명확한 설명]

## 해결하려는 문제
[이 기능이 해결할 문제]

## 제안하는 해결책
[어떻게 구현할 수 있을지]

## 대안
[고려한 다른 방법들]

## 추가 정보
[참고 자료, 관련 링크 등]
```

### 코드 기여

#### 1. Issue 선택

- "good first issue" 라벨이 붙은 Issue부터 시작하는 것을 추천합니다
- Issue에 댓글을 남겨 작업 중임을 알려주세요

#### 2. 코드 작성

**코딩 스타일:**

**Go:**
- `gofmt`로 포맷팅
- `golangci-lint` 통과
- 명확한 변수명 사용
- 공개 함수에 주석 추가

예시:
```go
// GenerateManifest generates Kubernetes manifest from Docker container info
// using AI-based analysis and few-shot learning.
func GenerateManifest(ctx context.Context, info ContainerInfo) (*ManifestResult, error) {
    // 구현
}
```

**TypeScript/React:**
- Prettier로 포맷팅
- ESLint 규칙 준수
- 함수형 컴포넌트 사용
- 명확한 prop 타입 정의

예시:
```typescript
interface ContainerCardProps {
  container: Container;
  onDeploy: (id: string) => void;
}

export function ContainerCard({ container, onDeploy }: ContainerCardProps) {
  // 구현
}
```

#### 3. 테스트 작성

**모든 새로운 코드는 테스트가 필요합니다.**

**Go 테스트:**
```go
func TestGenerateManifest(t *testing.T) {
    tests := []struct {
        name    string
        input   ContainerInfo
        want    *ManifestResult
        wantErr bool
    }{
        {
            name: "nginx container",
            input: ContainerInfo{
                Image: "nginx:1.21",
                Ports: []int{80},
            },
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := GenerateManifest(context.Background(), tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("GenerateManifest() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            // 추가 assertion
        })
    }
}
```

**React 테스트:**
```typescript
import { render, screen } from '@testing-library/react';
import { ContainerCard } from './ContainerCard';

describe('ContainerCard', () => {
  it('renders container name', () => {
    const container = {
      id: '123',
      name: 'nginx',
      image: 'nginx:1.21',
      status: 'running',
    };

    render(<ContainerCard container={container} onDeploy={() => {}} />);

    expect(screen.getByText('nginx')).toBeInTheDocument();
  });
});
```

#### 4. 커밋

**커밋 메시지 규칙:**

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type:**
- `feat`: 새로운 기능
- `fix`: 버그 수정
- `docs`: 문서 변경
- `style`: 코드 스타일 (포맷팅, 세미콜론 등)
- `refactor`: 리팩토링
- `test`: 테스트 추가/수정
- `chore`: 빌드, 설정 변경

**Scope (선택사항):**
- `api`: API 관련
- `ui`: UI 관련
- `ai`: AI 엔진 관련
- `docker`: Docker 관련
- `k8s`: Kubernetes 관련

**예시:**
```
feat(ai): add few-shot learning for manifest generation

- Implement similarity search for past deployments
- Build prompt with 3-5 similar examples
- Add caching for repeated requests

Closes #123
```

```
fix(api): handle null response from Docker API

Docker API can return null for stopped containers.
Add null check and return empty array instead.

Fixes #456
```

#### 5. Pull Request

**PR 생성 전 체크리스트:**
- [ ] 코드가 정상 동작함
- [ ] 테스트 작성 및 통과
- [ ] 린트 에러 없음
- [ ] 문서 업데이트 (필요 시)
- [ ] 커밋 메시지 규칙 준수
- [ ] 브랜치가 최신 main과 동기화됨

**PR 템플릿:**

```markdown
## 변경 사항
[변경한 내용 설명]

## 관련 Issue
Closes #123

## 변경 유형
- [ ] 버그 수정
- [ ] 새로운 기능
- [ ] 리팩토링
- [ ] 문서 업데이트
- [ ] 기타 (설명: )

## 테스트
- [ ] 단위 테스트 추가
- [ ] 통합 테스트 추가
- [ ] 수동 테스트 완료

## 체크리스트
- [ ] 코드가 코딩 스타일을 따름
- [ ] 자체 코드 리뷰 완료
- [ ] 주석 추가 (복잡한 로직)
- [ ] 문서 업데이트
- [ ] 경고 없음
- [ ] 테스트 통과

## 스크린샷 (UI 변경 시)
[스크린샷 첨부]

## 추가 정보
[기타 리뷰어가 알아야 할 정보]
```

**PR 크기:**
- 작은 PR이 좋은 PR입니다 (300줄 이하 권장)
- 큰 변경은 여러 PR로 나누세요
- 리팩토링과 기능 추가를 분리하세요

#### 6. 코드 리뷰

- 리뷰어의 피드백에 정중하게 응답
- 요청된 변경사항 반영
- 논의가 필요한 경우 건설적으로 토론

#### 7. 병합

- 모든 리뷰 승인 필요
- CI/CD 통과 필요
- Squash merge 사용 (깔끔한 히스토리)

## 문서 기여

문서 개선도 큰 기여입니다!

**문서 유형:**
- README.md: 프로젝트 소개
- docs/ARCHITECTURE.md: 시스템 아키텍처
- docs/API_SPEC.md: API 명세
- docs/adr/: 아키텍처 결정 기록
- docs/SETUP.md: 개발 환경 설정

**문서 작성 시:**
- 명확하고 간결하게
- 예시 코드 포함
- 스크린샷 활용 (UI 관련)
- 최신 상태 유지

## 릴리스 프로세스

메인테이너가 담당합니다:

1. 버전 결정 (Semantic Versioning)
2. CHANGELOG.md 업데이트
3. Git 태그 생성
4. GitHub Release 생성
5. Docker 이미지 빌드 및 푸시

## 커뮤니티

### 소통 채널

- GitHub Issues: 버그 리포트, 기능 제안
- GitHub Discussions: 질문, 아이디어 토론
- Pull Requests: 코드 리뷰, 기술 토론

### 도움 요청

막힌 부분이 있나요?

1. 문서 확인 (docs/)
2. 기존 Issue 검색
3. 새 Issue 생성 (질문 라벨)

## 기여자 인정

모든 기여자는 README.md의 Contributors 섹션에 추가됩니다.

## 라이선스

프로젝트에 기여하면 MIT License에 동의하는 것으로 간주됩니다.

## 추가 리소스

- [Git 브랜치 전략](https://nvie.com/posts/a-successful-git-branching-model/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [Effective Go](https://go.dev/doc/effective_go)
- [React TypeScript Cheatsheet](https://react-typescript-cheatsheet.netlify.app/)

## 감사합니다!

여러분의 기여가 이 프로젝트를 더 좋게 만듭니다. 🎉
