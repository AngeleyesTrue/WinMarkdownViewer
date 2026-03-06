---
spec_id: SPEC-UI-001
type: implementation-plan
version: 1.0.0
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
---

# SPEC-UI-001 구현 계획

## 1. 구현 전략

### 1.1 개발 방법론

TDD (Test-Driven Development) - RED-GREEN-REFACTOR 사이클에 따라 구현합니다.

### 1.2 전체 접근 방식

바텀업(Bottom-Up) 방식으로 핵심 모듈부터 구현합니다:

1. 프로젝트 스캐폴딩 (Go 모듈, 디렉토리 구조)
2. 정적 리소스 (HTML 템플릿, CSS, go:embed)
3. Markdown 렌더링 엔진 (goldmark)
4. WebView2 뷰어 (윈도우 생성 및 HTML 표시)
5. CLI 진입점 (인자 파싱, 에러 처리, 파이프라인 연결)

---

## 2. 마일스톤

### Primary Goal: 프로젝트 초기화 및 정적 리소스

**Task 1: Go 모듈 초기화**
- 파일: `go.mod`
- 작업: `go mod init`, 의존성 추가
- 의존성: 없음

**Task 2: 디렉토리 구조 생성**
- 파일: 디렉토리만 생성
- 작업: `cmd/winmdview/`, `internal/markdown/`, `internal/viewer/`, `web/` 디렉토리 생성
- 의존성: Task 1

**Task 3: HTML 템플릿 작성**
- 파일: `web/template.html`
- 작업: 기본 HTML5 구조, `{{.Content}}` 플레이스홀더, CSS 인라인 삽입 구조
- 의존성: Task 2

**Task 4: GitHub 스타일 CSS 작성**
- 파일: `web/style.css`
- 작업: GitHub Markdown 렌더링 스타일 CSS 작성 (github-markdown-css 기반)
- 코드 블록 구문 강조용 chroma CSS 포함
- 의존성: Task 2

**Task 5: go:embed 설정**
- 파일: `web/embed.go`
- 작업: `//go:embed template.html style.css` 디렉티브 선언, 패키지 레벨 변수 노출
- 의존성: Task 3, Task 4

### Secondary Goal: Markdown 렌더링 엔진

**Task 6: Markdown 렌더러 테스트 작성 (RED)**
- 파일: `internal/markdown/renderer_test.go`
- 작업: 기본 Markdown 변환, GFM 확장, 코드 블록 구문 강조 테스트
- 의존성: Task 1

**Task 7: Markdown 렌더러 구현 (GREEN)**
- 파일: `internal/markdown/renderer.go`
- 작업:
  - goldmark 인스턴스 구성 (GFM extension, highlighting extension)
  - `Render(content []byte) (string, error)` 함수
  - HTML 템플릿과 결합하는 `RenderPage(content []byte, fileName string) (string, error)` 함수
- 의존성: Task 5, Task 6

**Task 8: 렌더러 리팩토링 (REFACTOR)**
- 파일: `internal/markdown/renderer.go`
- 작업: goldmark 옵션 분리, 에러 처리 개선
- 의존성: Task 7

### Final Goal: WebView2 뷰어 및 CLI 통합

**Task 9: WebView2 뷰어 구현**
- 파일: `internal/viewer/viewer.go`
- 작업:
  - `New(title string, width, height int) (*Viewer, error)` 생성자
  - `SetContent(html string) error` HTML 설정
  - `Run() error` 윈도우 이벤트 루프 실행
  - WebView2 Runtime 미설치 시 에러 처리
- 의존성: Task 1
- 참고: WebView2는 GUI 컴포넌트로 단위 테스트 대신 빌드 검증 및 수동 테스트

**Task 10: CLI 진입점 테스트 작성 (RED)**
- 파일: `cmd/winmdview/main_test.go`
- 작업: 인자 파싱 로직, 파일 검증 로직 테스트 (WebView2 의존 부분 제외)
- 의존성: Task 6

**Task 11: CLI 진입점 구현 (GREEN)**
- 파일: `cmd/winmdview/main.go`
- 작업:
  - `os.Args` 파싱
  - 파일 존재 여부 및 읽기 권한 검증
  - Markdown 읽기 -> 렌더링 -> WebView2 표시 파이프라인
  - 에러 시 `os.Exit(1)` 및 stderr 메시지 출력
  - 사용법 안내 메시지
- 의존성: Task 7, Task 9, Task 10

**Task 12: 빌드 및 통합 검증**
- 작업: `go build ./cmd/winmdview/` 성공 확인, 실제 .md 파일로 실행 검증
- 의존성: Task 11

---

## 3. 파일 영향 분석

| 파일 | 작업 유형 | 복잡도 | 관련 Task |
|------|-----------|--------|-----------|
| `go.mod` | 신규 생성 | 낮음 | Task 1 |
| `web/template.html` | 신규 생성 | 낮음 | Task 3 |
| `web/style.css` | 신규 생성 | 중간 | Task 4 |
| `web/embed.go` | 신규 생성 | 낮음 | Task 5 |
| `internal/markdown/renderer.go` | 신규 생성 | 중간 | Task 7, 8 |
| `internal/markdown/renderer_test.go` | 신규 생성 | 중간 | Task 6 |
| `internal/viewer/viewer.go` | 신규 생성 | 중간 | Task 9 |
| `cmd/winmdview/main.go` | 신규 생성 | 중간 | Task 11 |
| `cmd/winmdview/main_test.go` | 신규 생성 | 중간 | Task 10 |

**총 파일 수**: 9개 (신규 생성)
**전체 복잡도**: 중간

---

## 4. 기술적 접근

### 4.1 Markdown 렌더링 아키텍처

```
[.md 파일] -> os.ReadFile() -> []byte
                                  |
                                  v
                        goldmark.Convert()
                    (GFM + Highlighting 확장)
                                  |
                                  v
                          HTML 문자열
                                  |
                                  v
                    template.html + style.css
                    (go:embed, text/template)
                                  |
                                  v
                        완전한 HTML 문서
                                  |
                                  v
                    WebView2.SetHtml() 호출
```

### 4.2 goldmark 구성

```
goldmark.New(
    goldmark.WithExtensions(
        extension.GFM,           // 테이블, 취소선, 자동 링크, 태스크 리스트
        highlighting.NewHighlighting(
            highlighting.WithStyle("github"),
        ),
    ),
    goldmark.WithRendererOptions(
        html.WithUnsafe(),       // 인라인 HTML 허용
    ),
)
```

### 4.3 에러 처리 전략

| 에러 상황 | 처리 방식 | exit code |
|-----------|-----------|-----------|
| 인자 없음 | stderr에 사용법 출력 | 1 |
| 파일 미존재 | stderr에 에러 메시지 | 1 |
| 읽기 권한 없음 | stderr에 에러 메시지 | 1 |
| Markdown 변환 실패 | stderr에 에러 메시지 | 1 |
| WebView2 미설치 | stderr에 설치 안내 | 1 |

---

## 5. 리스크 및 대응

| 리스크 | 영향도 | 대응 방안 |
|--------|--------|-----------|
| go-webview2가 최신 Go 1.26과 호환되지 않을 수 있음 | 높음 | go.mod에서 호환 버전 확인, 필요 시 fork 검토 |
| WebView2 Runtime 미설치 환경에서 panic 발생 가능 | 중간 | Runtime 존재 여부를 사전 확인하는 로직 추가 |
| 대용량 Markdown 파일(>5MB) 렌더링 시 성능 저하 | 낮음 | MVP 범위에서는 10MB 이하 가정, 추후 스트리밍 검토 |
| GitHub 스타일 CSS의 라이선스 호환성 | 낮음 | MIT 라이선스인 github-markdown-css 기반으로 작성 |

---

## 6. 범위 외 (Out of Scope)

다음 기능은 후속 SPEC에서 다룹니다:

| 기능 | 예정 SPEC |
|------|-----------|
| 파일 감시 / 실시간 미리보기 | SPEC-002 |
| WebSocket 서버 | SPEC-002 |
| 컨텍스트 메뉴 등록 | SPEC-003 |
| 시스템 트레이 | SPEC-003 |
| 단일 인스턴스 관리 | SPEC-003 |
| MSI 인스톨러 | SPEC-004 |
| 사용자 설정/환경설정 | SPEC-005 |
| KaTeX, Mermaid 지원 | 후속 SPEC |

---

## 7. 다음 단계

1. `/moai:2-run SPEC-UI-001` 실행하여 TDD 사이클로 구현 시작
2. 구현 완료 후 `/moai:3-sync SPEC-UI-001` 실행하여 문서화
