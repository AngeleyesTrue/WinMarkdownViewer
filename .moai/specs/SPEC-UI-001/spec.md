---
id: SPEC-UI-001
title: "Markdown Viewer MVP - File Open and WebView2 Rendering"
version: 1.0.0
status: draft
created: 2026-03-06
updated: 2026-03-06
author: "Claud Archive"
priority: P1
lifecycle: spec-first
tags: [webview2, goldmark, markdown, mvp]
---

# SPEC-UI-001: Markdown Viewer MVP - 파일 열기 및 WebView2 렌더링

## HISTORY

| 버전 | 날짜 | 작성자 | 변경 내용 |
|------|------|--------|-----------|
| 1.0.0 | 2026-03-06 | Claud Archive | 최초 작성 |

---

## 1. Environment (환경)

### 1.1 대상 플랫폼
- Windows 10 21H2 이상, Windows 11
- Microsoft Edge WebView2 Runtime 설치 필수 (Evergreen 배포)

### 1.2 기술 스택
- Go 1.26.0 (CGO 불필요)
- github.com/jchv/go-webview2 - pure Go WebView2 바인딩
- github.com/yuin/goldmark - GFM Markdown 파서
- github.com/yuin/goldmark-highlighting - 코드 구문 강조 (chroma 기반)
- go:embed - 정적 리소스 임베딩 (HTML 템플릿, CSS)

### 1.3 프로젝트 구조
```
WinMarkdownViewer/
  cmd/winmdview/main.go              # 진입점
  internal/markdown/renderer.go      # goldmark 렌더링 엔진
  internal/viewer/viewer.go          # WebView2 윈도우 관리
  web/
    templates/viewer.html            # HTML 템플릿
    css/github-markdown.css          # GitHub 스타일 CSS
    embed.go                         # go:embed 선언
  go.mod
  go.sum
```

---

## 2. Assumptions (가정)

- A1: 사용자의 Windows 시스템에 Microsoft Edge WebView2 Runtime이 설치되어 있다.
- A2: 입력 파일은 UTF-8 인코딩된 Markdown(.md) 파일이다.
- A3: 사용자는 명령줄에서 파일 경로를 인자로 전달하여 프로그램을 실행한다.
- A4: 단일 파일만 열며, 동시에 여러 파일을 여는 기능은 이 SPEC 범위 밖이다.
- A5: 파일 크기는 일반적인 Markdown 문서 범위(10MB 이하)를 가정한다.

---

## 3. Requirements (요구사항)

### 3.1 Ubiquitous (항상 활성)

- **REQ-U-001**: 시스템은 **항상** 렌더링된 HTML에 GitHub 스타일 CSS를 적용해야 한다.
- **REQ-U-002**: 시스템은 **항상** GFM(GitHub Flavored Markdown) 확장 문법을 지원해야 한다 (테이블, 취소선, 자동 링크, 태스크 리스트).
- **REQ-U-003**: 시스템은 **항상** 코드 블록에 구문 강조(syntax highlighting)를 적용해야 한다.

### 3.2 Event-Driven (이벤트 기반)

- **REQ-E-001**: **WHEN** 사용자가 명령줄 인자로 .md 파일 경로를 전달하면, **THEN** 시스템은 해당 파일을 읽어 HTML로 변환하고 WebView2 윈도우에 표시해야 한다.
- **REQ-E-002**: **WHEN** 사용자가 WebView2 윈도우를 닫으면, **THEN** 시스템은 모든 리소스를 정리하고 프로세스를 종료해야 한다.
- **REQ-E-003**: **WHEN** 사용자가 인자 없이 프로그램을 실행하면, **THEN** 시스템은 사용법 안내 메시지를 표시하고 종료해야 한다.
  - 참고: 릴리스 빌드(`-H windowsgui`)에서는 Windows MessageBox API로 에러 표시, 개발 빌드에서는 stderr 출력.

### 3.3 Unwanted (금지 행위)

- **REQ-N-001**: 시스템은 외부 네트워크 요청을 **하지 않아야 한다** (모든 리소스는 go:embed로 임베딩).
- **REQ-N-002**: 시스템은 **자체적으로** 임시 파일을 파일시스템에 생성하지 않아야 한다 (WebView2 Runtime이 자체적으로 생성하는 사용자 데이터 폴더는 제외).
- **REQ-N-003**: 시스템은 사용자에게 관리자 권한을 **요구하지 않아야 한다**.

### 3.4 State-Driven (상태 기반)

- **REQ-S-001**: **IF** 지정된 파일이 존재하지 않으면, **THEN** 시스템은 에러 메시지를 표시하고 비정상 종료 코드(exit code 1)로 종료해야 한다.
- **REQ-S-002**: **IF** 파일 읽기 중 권한 오류가 발생하면, **THEN** 시스템은 구체적인 에러 메시지를 표시하고 종료해야 한다.
- **REQ-S-003**: **IF** WebView2 Runtime이 설치되지 않은 환경이면, **THEN** 시스템은 사용자에게 WebView2 설치 안내 메시지를 표시해야 한다.

### 3.5 Optional (선택 사항)

- **REQ-O-001**: **가능하면** WebView2 윈도우의 타이틀바에 열린 파일명을 표시해야 한다.
- **REQ-O-002**: **가능하면** 빈 Markdown 파일에 대해 "내용이 없습니다" 안내 메시지를 표시해야 한다.

---

## 4. Specifications (상세 명세)

### 4.1 Go 모듈 초기화

- 모듈 경로: `github.com/AngeleyesTrue/WinMarkdownViewer` (또는 프로젝트에 맞는 경로)
- `go mod init` 실행 후 의존성 추가:
  - `github.com/jchv/go-webview2`
  - `github.com/yuin/goldmark`
  - `github.com/yuin/goldmark-highlighting`

### 4.2 Markdown 렌더링 엔진 (`internal/markdown/`)

- goldmark 인스턴스를 GFM 확장 및 highlighting 확장과 함께 구성
- `Render(content []byte) (string, error)` 함수 제공
- 입력: Markdown 바이트 슬라이스
- 출력: HTML 문자열
- HTML 템플릿과 결합하여 완전한 HTML 문서 생성

### 4.3 WebView2 뷰어 (`internal/viewer/`)

- `go-webview2` 라이브러리를 사용한 네이티브 윈도우 생성
- 윈도우 크기: 1024x768 (기본값)
- 윈도우 타이틀: 파일명 포함 (예: "filename.md - WinMarkdownViewer")
- `SetHtml()` 또는 `Navigate()` 메서드로 렌더링된 HTML 표시

### 4.4 HTML 템플릿 및 CSS (`web/`)

- `templates/viewer.html`: 기본 HTML 구조, `{{.Content}}` 플레이스홀더
- `css/github-markdown.css`: GitHub Markdown 스타일 CSS (github-markdown-css 기반)
- `embed.go`: `//go:embed` 디렉티브로 templates/viewer.html, css/github-markdown.css 임베딩

### 4.5 진입점 (`cmd/winmdview/main.go`)

- `os.Args` 파싱으로 파일 경로 획득
- 파일 존재 여부 및 읽기 권한 검증
- Markdown 렌더링 파이프라인 실행
- WebView2 윈도우 생성 및 표시
- 윈도우 종료 시 클린업 처리

---

## 5. Constraints (제약사항)

- CGO 사용 금지 (pure Go 빌드)
- 외부 CDN/네트워크 리소스 사용 금지
- Windows 전용 (Linux/macOS 지원 불필요)
- 이 SPEC에서 파일 감시(watcher), WebSocket, 시스템 트레이, 컨텍스트 메뉴, 인스톨러는 범위 밖
- KaTeX/Mermaid 렌더링은 이 SPEC 범위에 포함되지 않음 (후속 SPEC에서 구현)
- MVP는 라이트 테마만 지원. `prefers-color-scheme` 기반 다크 모드는 후속 SPEC에서 구현

---

## 6. Traceability (추적성)

| 요구사항 ID | 구현 파일 | 테스트 시나리오 |
|-------------|-----------|-----------------|
| REQ-U-001 | web/css/github-markdown.css, web/templates/viewer.html | ACC-001 |
| REQ-U-002 | internal/markdown/renderer.go | ACC-002 |
| REQ-U-003 | internal/markdown/renderer.go | ACC-002 |
| REQ-E-001 | cmd/winmdview/main.go, internal/viewer/ | ACC-001 |
| REQ-E-002 | internal/viewer/viewer.go | ACC-003 |
| REQ-E-003 | cmd/winmdview/main.go | ACC-004 |
| REQ-N-001 | web/embed.go | ACC-005 |
| REQ-N-002 | cmd/winmdview/main.go | ACC-005 |
| REQ-N-003 | cmd/winmdview/main.go | ACC-012 |
| REQ-O-001 | internal/viewer/viewer.go | ACC-001 |
| REQ-O-002 | internal/viewer/viewer.go, internal/markdown/renderer.go | ACC-008 |
| REQ-S-001 | cmd/winmdview/main.go | ACC-006 |
| REQ-S-002 | cmd/winmdview/main.go | ACC-006-B |
| REQ-S-003 | internal/viewer/viewer.go | ACC-007 |
