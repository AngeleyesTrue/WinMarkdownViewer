# WinMarkdownViewer - 기술 스택

## 언어 및 런타임

| 항목 | 선택 | 버전 | 사유 |
|------|------|------|------|
| 언어 | Go | 1.23+ | 빠른 컴파일, 단일 바이너리, Windows API 호출 용이 |
| 빌드 | Go toolchain | - | CGO 불필요 (순수 Go WebView2 바인딩) |

## 핵심 라이브러리

| 라이브러리 | 용도 | 비고 |
|-----------|------|------|
| `github.com/jchv/go-webview2` | WebView2 바인딩 | Windows 내장 Edge 엔진 활용 |
| `github.com/yuin/goldmark` | 마크다운 파싱/렌더링 | GFM 확장 지원 |
| `github.com/yuin/goldmark-highlighting` | 코드 구문 강조 | chroma 기반 |
| `github.com/fsnotify/fsnotify` | 파일 변경 감시 | 크로스 플랫폼 파일 시스템 감시 |
| `github.com/gorilla/websocket` | WebSocket 통신 | 실시간 새로고침용 |
| `golang.org/x/sys/windows` | Windows API | 레지스트리, 시스템 트레이 |
| `github.com/getlantern/systray` | 시스템 트레이 | Windows 시스템 트레이 아이콘 |

## 프론트엔드 (WebView2 내 HTML/CSS/JS)

| 항목 | 선택 | 사유 |
|------|------|------|
| CSS | GitHub Markdown CSS | GitHub 스타일 마크다운 렌더링 |
| 코드 하이라이트 | highlight.js (임베디드) | 구문 강조 테마 |
| 다이어그램 | mermaid.js (임베디드) | 다이어그램 렌더링 |
| 수식 | KaTeX (임베디드) | 수학 수식 렌더링 |
| 실시간 연결 | WebSocket (네이티브) | 파일 변경 시 자동 새로고침 |

## 빌드 및 배포

| 항목 | 선택 | 사유 |
|------|------|------|
| 빌드 도구 | Go build + ldflags | 버전 정보 임베딩 |
| 리소스 임베딩 | Go embed | HTML/CSS/JS를 바이너리에 포함 |
| 아이콘 | rsrc 또는 goversioninfo | Windows 실행 파일 아이콘/매니페스트 |
| MSI 빌드 | WiX Toolset v4 | Windows Installer 패키지 |
| CI/CD | GitHub Actions | 자동 빌드/릴리스 |

## 아키텍처 결정 사항

### WebView2 선택 이유
- Windows 10/11에 Edge WebView2 Runtime 기본 포함 (또는 자동 설치)
- 별도 브라우저 없이 네이티브 웹 렌더링
- 마크다운 HTML을 그대로 렌더링 가능
- Go 바인딩(go-webview2)이 안정적

### 실시간 미리보기 아키텍처
```
[파일 시스템] --fsnotify--> [Go 서버] --WebSocket--> [WebView2]
     |                         |                        |
  .md 파일 변경           goldmark 렌더링         HTML 업데이트
```

1. fsnotify로 .md 파일 변경 감지
2. goldmark으로 마크다운을 HTML로 변환
3. 내장 HTTP 서버를 통해 WebSocket으로 변경된 HTML 전송
4. WebView2가 DOM을 업데이트 (스크롤 위치 유지)

### 단일 인스턴스 패턴
- Named Mutex로 이미 실행 중인 인스턴스 감지
- Named Pipe 또는 TCP로 새 파일 경로를 기존 인스턴스에 전달
- 기존 인스턴스가 새 파일을 열어 표시

### 순수 Go 빌드
- go-webview2는 CGO 불필요 (Windows COM API 직접 호출)
- GCC/MinGW 설치 불필요
- `go build`만으로 빌드 가능

## 개발 환경 요구사항

- Go 1.23 이상
- WiX Toolset v4 - MSI 빌드용 (선택)
- Windows 10/11 개발 머신
